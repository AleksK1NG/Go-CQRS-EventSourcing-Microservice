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

type bankAccountMongoSubscription struct {
	log        logger.Logger
	cfg        *config.Config
	bs         *service.BankAccountService
	mp         es.Projection
	serializer es.Serializer
	mr         domain.MongoRepository
	as         es.AggregateStore
}

func NewBankAccountMongoSubscription(
	log logger.Logger,
	cfg *config.Config,
	bs *service.BankAccountService,
	mp es.Projection,
	serializer es.Serializer,
	mr domain.MongoRepository,
	as es.AggregateStore,
) *bankAccountMongoSubscription {
	return &bankAccountMongoSubscription{log: log, cfg: cfg, bs: bs, mp: mp, serializer: serializer, mr: mr, as: as}
}

func (s *bankAccountMongoSubscription) ProcessMessagesErrGroup(ctx context.Context, r *kafka.Reader, workerID int) error {

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		m, err := r.FetchMessage(ctx)
		if err != nil {
			s.log.Warnf("(bankAccountMongoSubscription) workerID: %v, err: %v", workerID, err)
			continue
		}

		s.logProcessMessage(m, workerID)

		switch m.Topic {
		case es.GetTopicName(s.cfg.KafkaPublisherConfig.TopicPrefix, string(domain.BankAccountAggregateType)):
			s.handleBankAccountEvents(ctx, r, m)
		}
	}
}

func (s *bankAccountMongoSubscription) handleBankAccountEvents(ctx context.Context, r *kafka.Reader, m kafka.Message) {
	ctx, span := tracing.StartKafkaConsumerTracerSpan(ctx, m.Headers, "bankAccountMongoSubscription.handleBankAccountEvents")
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

func (s *bankAccountMongoSubscription) handle(ctx context.Context, r *kafka.Reader, m kafka.Message, event es.Event) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "bankAccountMongoSubscription.handle")
	defer span.Finish()

	err := s.mp.When(ctx, event)
	if err != nil {
		s.log.Errorf("MongoSubscription When err: %v", err)

		recreateErr := s.recreateProjection(ctx, event)
		if err != nil {
			return tracing.TraceWithErr(span, errors.Wrapf(recreateErr, "recreateProjection err: %v", err))
		}

		s.log.Infof("[Mongo projection recreated] commit aggregateID: %s", event.GetAggregateID())
		s.commitErrMessage(ctx, r, m)
		return tracing.TraceWithErr(span, errors.Wrapf(err, "When type: %s, aggregateID: %s", event.GetEventType(), event.GetAggregateID()))
	}

	s.log.Infof("Mongo projection handle event: %s", event.String())
	return nil
}

func (s *bankAccountMongoSubscription) recreateProjection(ctx context.Context, event es.Event) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "bankAccountMongoSubscription.recreateProjection")
	defer span.Finish()

	err := s.mr.DeleteByAggregateID(ctx, event.GetAggregateID())
	if err != nil {
		s.log.Errorf("MongoSubscription DeleteByAggregateID err: %v", err)
		return tracing.TraceWithErr(span, errors.Wrapf(err, "When DeleteByAggregateID type: %s, aggregateID: %s", event.GetEventType(), event.GetAggregateID()))
	}

	bankAccountAggregate := domain.NewBankAccountAggregate(event.GetAggregateID())
	err = s.as.Load(ctx, bankAccountAggregate)
	if err != nil {
		s.log.Errorf("MongoSubscription as.Load err: %v", err)
		return tracing.TraceWithErr(span, errors.Wrapf(err, "When as.Load type: %s, aggregateID: %s", event.GetEventType(), event.GetAggregateID()))
	}

	err = s.mr.Insert(ctx, mappers.BankAccountToMongoProjection(bankAccountAggregate))
	if err != nil {
		s.log.Errorf("MongoSubscription mr.Insert err: %v", err)
		return tracing.TraceWithErr(span, errors.Wrapf(err, "When mr.Insert type: %s, aggregateID: %s", event.GetEventType(), event.GetAggregateID()))
	}

	return nil
}
