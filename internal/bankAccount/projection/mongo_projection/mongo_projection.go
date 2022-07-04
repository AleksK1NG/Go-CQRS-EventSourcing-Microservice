package mongo_projection

import (
	"context"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/config"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/internal/bankAccount/domain"
	bankAccountErrors "github.com/AleksK1NG/go-cqrs-eventsourcing/internal/bankAccount/errors"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/internal/bankAccount/events"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/es"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/logger"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/tracing"
	"github.com/Rhymond/go-money"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
	"time"
)

type bankAccountMongoProjection struct {
	log             logger.Logger
	cfg             *config.Config
	serializer      es.Serializer
	mongoRepository domain.MongoRepository
}

func NewBankAccountMongoProjection(
	log logger.Logger,
	cfg *config.Config,
	serializer es.Serializer,
	mongoRepository domain.MongoRepository,
) *bankAccountMongoProjection {
	return &bankAccountMongoProjection{log: log, cfg: cfg, serializer: serializer, mongoRepository: mongoRepository}
}

func (b *bankAccountMongoProjection) When(ctx context.Context, esEvent es.Event) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "bankAccountMongoProjection.When")
	defer span.Finish()

	deserializedEvent, err := b.serializer.DeserializeEvent(esEvent)
	if err != nil {
		return errors.Wrapf(err, "serializer.DeserializeEvent aggregateID: %s, type: %s", esEvent.GetAggregateID(), esEvent.GetEventType())
	}

	switch event := deserializedEvent.(type) {

	case *events.BankAccountCreatedEventV1:
		return b.onBankAccountCreated(ctx, esEvent, event)

	case *events.BalanceDepositedEventV1:
		return b.onBalanceDeposited(ctx, esEvent, event)

	case *events.BalanceWithdrawnEventV1:
		return b.onBalanceWithdrawn(ctx, esEvent, event)

	case *events.EmailChangedEventV1:
		return b.onEmailChanged(ctx, esEvent, event)

	default:
		return errors.Wrapf(bankAccountErrors.ErrUnknownEventType, "esEvent: %s", esEvent.String())
	}
}

func (b *bankAccountMongoProjection) onBankAccountCreated(ctx context.Context, esEvent es.Event, event *events.BankAccountCreatedEventV1) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "bankAccountMongoProjection.onBankAccountCreated")
	defer span.Finish()
	span.LogFields(log.String("aggregateID", esEvent.GetAggregateID()))

	if esEvent.GetVersion() != 1 {
		return errors.Wrapf(es.ErrInvalidEventVersion, "type: %s, version: %d", esEvent.GetEventType(), esEvent.GetVersion())
	}

	projection := &domain.BankAccountMongoProjection{
		AggregateID: esEvent.GetAggregateID(),
		Version:     esEvent.GetVersion(),
		Email:       event.Email,
		Address:     event.Address,
		FirstName:   event.FirstName,
		LastName:    event.LastName,
		Balance: domain.Balance{
			Amount:   event.Balance.AsMajorUnits(),
			Currency: event.Balance.Currency().Code,
		},
		Status:    event.Status,
		UpdatedAt: time.Now().UTC(),
		CreatedAt: time.Now().UTC(),
	}

	err := b.mongoRepository.Insert(ctx, projection)
	if err != nil {
		return tracing.TraceWithErr(span, errors.Wrapf(err, "[onBankAccountCreated] mongoRepository.Insert aggregateID: %s", esEvent.GetAggregateID()))
	}

	b.log.Infof("[onBankAccountCreated] projection: %#v", projection)
	return nil
}

func (b *bankAccountMongoProjection) onBalanceDeposited(ctx context.Context, esEvent es.Event, event *events.BalanceDepositedEventV1) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "bankAccountMongoProjection.onBalanceDeposited")
	defer span.Finish()
	span.LogFields(log.String("aggregateID", esEvent.GetAggregateID()))

	if err := b.mongoRepository.UpdateConcurrently(ctx, esEvent.GetAggregateID(), func(projection *domain.BankAccountMongoProjection) *domain.BankAccountMongoProjection {
		projection.Balance.Amount += money.New(event.Amount, money.USD).AsMajorUnits()
		projection.Version = esEvent.GetVersion()
		return projection
	}, esEvent.GetVersion()-1); err != nil {
		return tracing.TraceWithErr(span, errors.Wrapf(err, "[onBalanceDeposited] mongoRepository.UpdateConcurrently aggregateID: %s", esEvent.GetAggregateID()))
	}

	b.log.Infof("[onBalanceDeposited] aggregateID: %s, eventType: %s, version: %d", esEvent.GetAggregateID(), esEvent.GetEventType(), esEvent.GetVersion())
	return nil
}

func (b *bankAccountMongoProjection) onBalanceWithdrawn(ctx context.Context, esEvent es.Event, event *events.BalanceWithdrawnEventV1) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "bankAccountMongoProjection.onBalanceWithdrawn")
	defer span.Finish()
	span.LogFields(log.String("aggregateID", esEvent.GetAggregateID()))

	if err := b.mongoRepository.UpdateConcurrently(ctx, esEvent.GetAggregateID(), func(projection *domain.BankAccountMongoProjection) *domain.BankAccountMongoProjection {
		projection.Balance.Amount -= money.New(event.Amount, money.USD).AsMajorUnits()
		projection.Version = esEvent.GetVersion()

		return projection
	}, esEvent.GetVersion()-1); err != nil {
		return tracing.TraceWithErr(span, errors.Wrapf(err, "[onBalanceWithdrawn] mongoRepository.UpdateConcurrently aggregateID: %s", esEvent.GetAggregateID()))
	}

	b.log.Infof("[onBalanceWithdrawn] aggregateID: %s, eventType: %s, version: %d", esEvent.GetAggregateID(), esEvent.GetEventType(), esEvent.GetVersion())
	return nil
}

func (b *bankAccountMongoProjection) onEmailChanged(ctx context.Context, esEvent es.Event, event *events.EmailChangedEventV1) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "bankAccountMongoProjection.onEmailChanged")
	defer span.Finish()
	span.LogFields(log.String("aggregateID", esEvent.GetAggregateID()))

	if err := b.mongoRepository.UpdateConcurrently(ctx, esEvent.GetAggregateID(), func(projection *domain.BankAccountMongoProjection) *domain.BankAccountMongoProjection {
		projection.Email = event.Email
		projection.Version = esEvent.GetVersion()
		return projection
	}, esEvent.GetVersion()-1); err != nil {
		return tracing.TraceWithErr(span, errors.Wrapf(err, "[onEmailChanged] mongoRepository.UpdateConcurrently aggregateID: %s", esEvent.GetAggregateID()))
	}

	b.log.Infof("[onEmailChanged] aggregateID: %s, eventType: %s, version: %d", esEvent.GetAggregateID(), esEvent.GetEventType(), esEvent.GetVersion())
	return nil
}
