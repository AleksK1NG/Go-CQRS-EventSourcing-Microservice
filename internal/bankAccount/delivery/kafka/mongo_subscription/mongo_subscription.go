package mongo_subscription

import (
	"context"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/config"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/internal/bankAccount/domain"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/internal/bankAccount/service"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/internal/mappers"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/es"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/es/serializer"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/logger"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/tracing"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
)

type mongoSubscription struct {
	log                logger.Logger
	cfg                *config.Config
	bankAccountService *service.BankAccountService
	projection         es.Projection
	serializer         es.Serializer
	mongoRepository    domain.MongoRepository
	aggregateStore     es.AggregateStore
	eventBus           es.EventsBus
}

func NewBankAccountMongoSubscription(
	log logger.Logger,
	cfg *config.Config,
	bankAccountService *service.BankAccountService,
	projection es.Projection,
	serializer es.Serializer,
	mongoRepository domain.MongoRepository,
	aggregateStore es.AggregateStore,
	eventBus es.EventsBus,
) *mongoSubscription {
	return &mongoSubscription{
		log:                log,
		cfg:                cfg,
		bankAccountService: bankAccountService,
		projection:         projection,
		serializer:         serializer,
		mongoRepository:    mongoRepository,
		aggregateStore:     aggregateStore,
		eventBus:           eventBus,
	}
}

func (s *mongoSubscription) ProcessMessagesErrGroup(ctx context.Context, r *kafka.Reader, workerID int) error {

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		m, err := r.FetchMessage(ctx)
		if err != nil {
			s.log.Warnf("(mongoSubscription) workerID: %v, err: %v", workerID, err)
			continue
		}

		s.logProcessMessage(m, workerID)

		switch m.Topic {
		case es.GetTopicName(s.cfg.KafkaPublisherConfig.TopicPrefix, string(domain.BankAccountAggregateType)):
			s.handleBankAccountEvents(ctx, r, m)
		}
	}
}

func (s *mongoSubscription) handleBankAccountEvents(ctx context.Context, r *kafka.Reader, m kafka.Message) {
	ctx, span := tracing.StartKafkaConsumerTracerSpan(ctx, m.Headers, "mongoSubscription.handleBankAccountEvents")
	defer span.Finish()

	var events []es.Event
	if err := serializer.Unmarshal(m.Value, &events); err != nil {
		s.log.Errorf("serializer.Unmarshal: %v", tracing.TraceWithErr(span, err))
		s.commitErrMessage(ctx, r, m)
		return
	}

	for _, event := range events {
		if err := s.handle(ctx, r, m, event); err != nil {
			s.log.Errorf("handleBankAccountEvents handle err: %v", err)
			return
		}
	}
	s.commitMessage(ctx, r, m)
}

func (s *mongoSubscription) handle(ctx context.Context, r *kafka.Reader, m kafka.Message, event es.Event) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "mongoSubscription.handle")
	defer span.Finish()

	err := s.projection.When(ctx, event)
	if err != nil {
		s.log.Errorf("MongoSubscription When err: %v", err)

		s.log.Warnf("Mongo recreating projection >>>>> aggregateID: %s, version: %d, type: %s", event.GetAggregateID(), event.GetVersion(), event.GetEventType())
		recreateErr := s.recreateProjection(ctx, event)
		if recreateErr != nil {
			s.log.Errorf("Mongo recreate projection >>>>> error err: %v", recreateErr)

			//err := s.eventBus.ProcessEvents(ctx, []es.Event{event})
			//if err != nil {
			//	return tracing.TraceWithErr(span, errors.Wrapf(err, "[eventBus] republish event err: %v", err))
			//}

			s.log.Warnf("Mongo RECREATE PROJECTION REPUBLISHED EVENT >>>>>>>> aggregateID: %s, version: %d, type: %s", event.GetAggregateID(), event.GetVersion(), event.GetEventType())
			return tracing.TraceWithErr(span, errors.Wrapf(recreateErr, "recreateProjection err: %v", err))
		}

		s.commitErrMessage(ctx, r, m)
		return tracing.TraceWithErr(span, errors.Wrapf(err, "When type: %s, aggregateID: %s", event.GetEventType(), event.GetAggregateID()))
	}

	s.log.Infof("Mongo projection handle event: %s", event.String())
	return nil
}

func (s *mongoSubscription) recreateProjection(ctx context.Context, event es.Event) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "mongoSubscription.recreateProjection")
	defer span.Finish()

	err := s.mongoRepository.DeleteByAggregateID(ctx, event.GetAggregateID())
	if err != nil {
		s.log.Errorf("MongoSubscription DeleteByAggregateID err: %v", err)
		return tracing.TraceWithErr(span, errors.Wrapf(err, "When DeleteByAggregateID type: %s, aggregateID: %s", event.GetEventType(), event.GetAggregateID()))
	}

	bankAccountAggregate := domain.NewBankAccountAggregate(event.GetAggregateID())
	err = s.aggregateStore.Load(ctx, bankAccountAggregate)
	if err != nil {
		s.log.Errorf("MongoSubscription aggregateStore.Load err: %v", err)
		return tracing.TraceWithErr(span, errors.Wrapf(err, "When aggregateStore.Load type: %s, aggregateID: %s", event.GetEventType(), event.GetAggregateID()))
	}

	err = s.mongoRepository.Insert(ctx, mappers.BankAccountToMongoProjection(bankAccountAggregate))
	if err != nil {
		s.log.Errorf("MongoSubscription mongoRepository.Insert err: %v", err)
		return tracing.TraceWithErr(span, errors.Wrapf(err, "When mongoRepository.Insert type: %s, aggregateID: %s", event.GetEventType(), event.GetAggregateID()))
	}

	s.log.Warnf("[Mongo Subscription] ***projection recreated commit*** aggregateID: %s, version: %d", event.GetAggregateID(), bankAccountAggregate.GetVersion())
	return nil
}
