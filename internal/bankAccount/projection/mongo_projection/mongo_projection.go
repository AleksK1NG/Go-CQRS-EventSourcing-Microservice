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
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
	"time"
)

type bankAccountMongoProjection struct {
	log        logger.Logger
	cfg        *config.Config
	serializer es.Serializer
	mr         domain.MongoRepository
}

func NewBankAccountMongoProjection(
	log logger.Logger,
	cfg *config.Config,
	serializer es.Serializer,
	mr domain.MongoRepository,
) *bankAccountMongoProjection {
	return &bankAccountMongoProjection{log: log, cfg: cfg, serializer: serializer, mr: mr}
}

func (b *bankAccountMongoProjection) When(ctx context.Context, esEvent es.Event) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "bankAccountMongoProjection.bankAccountMongoProjection")
	defer span.Finish()

	deserializedEvent, err := b.serializer.DeserializeEvent(esEvent)
	if err != nil {
		return errors.Wrapf(err, "serializer.DeserializeEvent aggregateID: %s, type: %s", esEvent.GetAggregateID(), esEvent.GetEventType())
	}

	switch event := deserializedEvent.(type) {

	case *events.BankAccountCreatedEventV1:
		return b.onBankAccountCreated(ctx, esEvent.GetAggregateID(), event)

	case *events.BalanceDepositedEventV1:
		return b.onBalanceDeposited(ctx, esEvent.GetAggregateID(), event)

	case *events.EmailChangedEventV1:
		return b.onEmailChanged(ctx, esEvent.GetAggregateID(), event)

	default:
		return errors.Wrapf(bankAccountErrors.ErrUnknownEventType, "esEvent: %s", esEvent.String())
	}
}

func (b *bankAccountMongoProjection) onBankAccountCreated(ctx context.Context, aggregateID string, event *events.BankAccountCreatedEventV1) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "bankAccountMongoProjection.onBankAccountCreated")
	defer span.Finish()
	span.LogFields(log.String("aggregateID", aggregateID))

	projection := &domain.BankAccountMongoProjection{
		AggregateID: aggregateID,
		Email:       event.Email,
		Address:     event.Address,
		FirstName:   event.FirstName,
		LastName:    event.LastName,
		Balance:     event.Balance,
		Status:      event.Status,
		UpdatedAt:   time.Now().UTC(),
		CreatedAt:   time.Now().UTC(),
	}

	err := b.mr.Insert(ctx, projection)
	if err != nil {
		return tracing.TraceWithErr(span, errors.Wrapf(err, "[onBankAccountCreated] mr.Insert aggregateID: %s", aggregateID))
	}

	b.log.Infof("[onBankAccountCreated] projection: %#v", projection)
	return nil
}

func (b *bankAccountMongoProjection) onBalanceDeposited(ctx context.Context, aggregateID string, event *events.BalanceDepositedEventV1) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "bankAccountMongoProjection.onBalanceDeposited")
	defer span.Finish()
	span.LogFields(log.String("aggregateID", aggregateID))

	projection, err := b.mr.GetByAggregateID(ctx, aggregateID)
	if err != nil {
		return tracing.TraceWithErr(span, errors.Wrapf(err, "[onBalanceDeposited] mr.GetByAggregateID aggregateID: %s", aggregateID))
	}

	projection.Balance += event.Amount

	err = b.mr.Update(ctx, projection)
	if err != nil {
		return tracing.TraceWithErr(span, errors.Wrapf(err, "[onBalanceDeposited] mr.Update aggregateID: %s", aggregateID))
	}

	b.log.Infof("[onBalanceDeposited] projection: %#v", projection)
	return nil
}

func (b *bankAccountMongoProjection) onEmailChanged(ctx context.Context, aggregateID string, event *events.EmailChangedEventV1) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "bankAccountMongoProjection.onEmailChanged")
	defer span.Finish()
	span.LogFields(log.String("aggregateID", aggregateID))

	projection, err := b.mr.GetByAggregateID(ctx, aggregateID)
	if err != nil {
		return tracing.TraceWithErr(span, errors.Wrapf(err, "[onEmailChanged] mr.GetByAggregateID aggregateID: %s", aggregateID))
	}

	projection.Email = event.Email

	err = b.mr.Update(ctx, projection)
	if err != nil {
		return tracing.TraceWithErr(span, errors.Wrapf(err, "[onEmailChanged] mr.Update aggregateID: %s", aggregateID))
	}

	b.log.Infof("[onEmailChanged] projection: %#v", projection)
	return nil
}
