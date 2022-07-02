package elasticsearch_subscription

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

type elasticSearchSubscription struct {
	log                logger.Logger
	cfg                *config.Config
	bankAccountService *service.BankAccountService
	projection         es.Projection
	serializer         es.Serializer
	elasticSearchRepo  domain.ElasticSearchRepository
	aggregateStore     es.AggregateStore
	eventBus           es.EventsBus
}

func NewElasticSearchSubscription(
	log logger.Logger,
	cfg *config.Config,
	bankAccountService *service.BankAccountService,
	projection es.Projection,
	serializer es.Serializer,
	elasticSearchRepo domain.ElasticSearchRepository,
	aggregateStore es.AggregateStore,
	eventBus es.EventsBus,
) *elasticSearchSubscription {
	return &elasticSearchSubscription{
		log:                log,
		cfg:                cfg,
		bankAccountService: bankAccountService,
		projection:         projection,
		serializer:         serializer,
		elasticSearchRepo:  elasticSearchRepo,
		aggregateStore:     aggregateStore,
		eventBus:           eventBus,
	}
}

func (s *elasticSearchSubscription) ProcessMessagesErrGroup(ctx context.Context, r *kafka.Reader, workerID int) error {

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		m, err := r.FetchMessage(ctx)
		if err != nil {
			s.log.Warnf("(elasticSearchSubscription) workerID: %v, err: %v", workerID, err)
			continue
		}

		s.logProcessMessage(m, workerID)

		switch m.Topic {
		case es.GetTopicName(s.cfg.KafkaPublisherConfig.TopicPrefix, string(domain.BankAccountAggregateType)):
			s.handleBankAccountEvents(ctx, r, m)
		}
	}
}

func (s *elasticSearchSubscription) handleBankAccountEvents(ctx context.Context, r *kafka.Reader, m kafka.Message) {
	ctx, span := tracing.StartKafkaConsumerTracerSpan(ctx, m.Headers, "elasticSearchSubscription.handleBankAccountEvents")
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

func (s *elasticSearchSubscription) handle(ctx context.Context, r *kafka.Reader, m kafka.Message, event es.Event) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "bankAccountMongoSubscription.handle")
	defer span.Finish()

	err := s.projection.When(ctx, event)
	if err != nil {
		s.log.Errorf("ElasticSearch Subscription When err: %v", err)

		s.log.Warnf("ElasticSearch Subscription recreating projection >>>>> aggregateID: %s, version: %d, type: %s", event.GetAggregateID(), event.GetVersion(), event.GetEventType())

		recreateErr := s.recreateProjection(ctx, event)
		if recreateErr != nil {
			s.log.Errorf("ElasticSearch Subscription recreate projection >>>>> error err: %v", recreateErr)

			//err := s.eventBus.ProcessEvents(ctx, []es.Event{event})
			//if err != nil {
			//	return tracing.TraceWithErr(span, errors.Wrapf(err, "ElasticSearchSubscription [eventBus] republish event err: %v", err))
			//}

			s.log.Warnf("ElasticSearch Subscription RECREATE PROJECTION REPUBLISHED EVENT >>>>>>>> aggregateID: %s, version: %d, type: %s", event.GetAggregateID(), event.GetVersion(), event.GetEventType())
			return tracing.TraceWithErr(span, errors.Wrapf(recreateErr, "recreateProjection err: %v", err))
		}

		s.commitErrMessage(ctx, r, m)
		return tracing.TraceWithErr(span, errors.Wrapf(err, "When type: %s, aggregateID: %s", event.GetEventType(), event.GetAggregateID()))
	}

	s.log.Infof("ElasticSearchSubscription projection handle event: %s", event.String())
	return nil
}

func (s *elasticSearchSubscription) recreateProjection(ctx context.Context, event es.Event) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "elasticSearchSubscription.recreateProjection")
	defer span.Finish()

	//_, err := s.elasticSearchRepo.GetByAggregateID(ctx, event.GetAggregateID())
	//if err != nil && strings.Contains(strings.ToLower(err.Error()), "404") {
	//	s.log.Warnf("RECREATE ELASTIC NOT FOUND >>>>>>>>>>>>>>> err: %v", err)
	//	return nil
	//}

	err := s.elasticSearchRepo.DeleteByAggregateID(ctx, event.GetAggregateID())
	if err != nil {
		s.log.Errorf("ElasticSearchSubscription DeleteByAggregateID err: %v", err)
		return tracing.TraceWithErr(span, errors.Wrapf(err, "When DeleteByAggregateID type: %s, aggregateID: %s", event.GetEventType(), event.GetAggregateID()))
	}

	bankAccountAggregate := domain.NewBankAccountAggregate(event.GetAggregateID())
	err = s.aggregateStore.Load(ctx, bankAccountAggregate)
	if err != nil {
		s.log.Errorf("ElasticSearchSubscription as.Load err: %v", err)
		return tracing.TraceWithErr(span, errors.Wrapf(err, "When as.Load type: %s, aggregateID: %s", event.GetEventType(), event.GetAggregateID()))
	}

	err = s.elasticSearchRepo.Index(ctx, mappers.BankAccountToElasticProjection(bankAccountAggregate))
	if err != nil {
		s.log.Errorf("ElasticSearchSubscription Index err: %v", err)
		return tracing.TraceWithErr(span, errors.Wrapf(err, "When Index type: %s, aggregateID: %s", event.GetEventType(), event.GetAggregateID()))
	}

	s.log.Infof("[ElasticSearchSubscription]  ***projection recreated commit*** aggregateID: %s, version: %d", event.GetAggregateID(), bankAccountAggregate.GetVersion())
	return nil
}
