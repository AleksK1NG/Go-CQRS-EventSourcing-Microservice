package elasticsearch_projection

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

type elasticProjection struct {
	log               logger.Logger
	cfg               *config.Config
	serializer        es.Serializer
	elasticSearchRepo domain.ElasticSearchRepository
}

func NewElasticProjection(log logger.Logger, cfg *config.Config, serializer es.Serializer, elasticSearchRepo domain.ElasticSearchRepository) *elasticProjection {
	return &elasticProjection{log: log, cfg: cfg, serializer: serializer, elasticSearchRepo: elasticSearchRepo}
}

func (e *elasticProjection) When(ctx context.Context, esEvent es.Event) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "elasticProjection.When")
	defer span.Finish()

	deserializedEvent, err := e.serializer.DeserializeEvent(esEvent)
	if err != nil {
		return errors.Wrapf(err, "serializer.DeserializeEvent aggregateID: %s, type: %s", esEvent.GetAggregateID(), esEvent.GetEventType())
	}

	switch event := deserializedEvent.(type) {

	case *events.BankAccountCreatedEventV1:
		return e.onBankAccountCreated(ctx, esEvent, event)

	case *events.BalanceDepositedEventV1:
		return e.onBalanceDeposited(ctx, esEvent, event)

	case *events.BalanceWithdrawnEventV1:
		return e.onBalanceWithdrawn(ctx, esEvent, event)

	case *events.EmailChangedEventV1:
		return e.onEmailChanged(ctx, esEvent, event)

	default:
		return errors.Wrapf(bankAccountErrors.ErrUnknownEventType, "esEvent: %s", esEvent.String())
	}
}

func (e *elasticProjection) onBankAccountCreated(ctx context.Context, esEvent es.Event, event *events.BankAccountCreatedEventV1) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "elasticProjection.onBankAccountCreated")
	defer span.Finish()
	span.LogFields(log.String("aggregateID", esEvent.GetAggregateID()))

	if esEvent.GetVersion() != 1 {
		return errors.Wrapf(es.ErrInvalidEventVersion, "type: %s, version: %d", esEvent.GetEventType(), esEvent.GetVersion())
	}

	projection := &domain.ElasticSearchProjection{
		ID:          esEvent.GetAggregateID(),
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

	err := e.elasticSearchRepo.Index(ctx, projection)
	if err != nil {
		return tracing.TraceWithErr(span, errors.Wrapf(err, "[onBalanceDeposited] elasticSearchRepo.Index aggregateID: %s", esEvent.GetAggregateID()))
	}

	e.log.Infof("ElasticSearch when [onBankAccountCreated] projection: %s", projection)
	return nil
}

func (e *elasticProjection) onBalanceDeposited(ctx context.Context, esEvent es.Event, event *events.BalanceDepositedEventV1) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "elasticProjection.onBalanceDeposited")
	defer span.Finish()
	span.LogFields(log.String("aggregateID", esEvent.GetAggregateID()))

	projection, err := e.elasticSearchRepo.GetByAggregateID(ctx, esEvent.GetAggregateID())
	if err != nil {
		return tracing.TraceWithErr(span, errors.Wrapf(err, "[onBalanceDeposited] elasticSearchRepo.GetByAggregateID aggregateID: %s", esEvent.GetAggregateID()))
	}

	if err := e.validateEventVersion(projection.Version, esEvent); err != nil {
		return tracing.TraceWithErr(span, err)
	}

	projection.Balance.Amount += money.New(event.Amount, money.USD).AsMajorUnits()
	projection.Version = esEvent.GetVersion()

	if err := e.elasticSearchRepo.Update(ctx, projection); err != nil {
		return tracing.TraceWithErr(span, errors.Wrapf(err, "[onBalanceWithdrawn] elasticSearchRepo.Update aggregateID: %s", esEvent.GetAggregateID()))
	}

	e.log.Infof("ElasticSearch when [onBalanceDeposited] projection: %s", projection)
	return nil
}

func (e *elasticProjection) onBalanceWithdrawn(ctx context.Context, esEvent es.Event, event *events.BalanceWithdrawnEventV1) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "elasticProjection.onBalanceWithdrawn")
	defer span.Finish()
	span.LogFields(log.String("aggregateID", esEvent.GetAggregateID()))

	projection, err := e.elasticSearchRepo.GetByAggregateID(ctx, esEvent.GetAggregateID())
	if err != nil {
		return tracing.TraceWithErr(span, errors.Wrapf(err, "[onBalanceWithdrawn] elasticSearchRepo.GetByAggregateID aggregateID: %s", esEvent.GetAggregateID()))
	}

	if err := e.validateEventVersion(projection.Version, esEvent); err != nil {
		return tracing.TraceWithErr(span, err)
	}

	projection.Balance.Amount -= money.New(event.Amount, money.USD).AsMajorUnits()
	projection.Version = esEvent.GetVersion()

	if err := e.elasticSearchRepo.Update(ctx, projection); err != nil {
		return tracing.TraceWithErr(span, errors.Wrapf(err, "[onBalanceWithdrawn] elasticSearchRepo.Update aggregateID: %s", esEvent.GetAggregateID()))
	}

	e.log.Infof("ElasticSearch when [onBalanceWithdrawn] projection: %s", projection)
	return nil
}

func (e *elasticProjection) onEmailChanged(ctx context.Context, esEvent es.Event, event *events.EmailChangedEventV1) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "elasticProjection.onEmailChanged")
	defer span.Finish()
	span.LogFields(log.String("aggregateID", esEvent.GetAggregateID()))

	projection, err := e.elasticSearchRepo.GetByAggregateID(ctx, esEvent.GetAggregateID())
	if err != nil {
		return tracing.TraceWithErr(span, errors.Wrapf(err, "[onEmailChanged] elasticSearchRepo.GetByAggregateID aggregateID: %s", esEvent.GetAggregateID()))
	}

	if err := e.validateEventVersion(projection.Version, esEvent); err != nil {
		return tracing.TraceWithErr(span, err)
	}

	projection.Email = event.Email
	projection.Version = esEvent.GetVersion()

	if err := e.elasticSearchRepo.Update(ctx, projection); err != nil {
		return tracing.TraceWithErr(span, errors.Wrapf(err, "[onEmailChanged] elasticSearchRepo.Update aggregateID: %s", esEvent.GetAggregateID()))
	}

	e.log.Infof("ElasticSearch when [onEmailChanged] projection: %s", projection)
	return nil
}

func (e *elasticProjection) validateEventVersion(version uint64, esEvent es.Event) error {
	if version != esEvent.GetVersion()-1 {
		return errors.Wrapf(es.ErrInvalidEventVersion, "type: %s, eventVersion: %d, projectionVersion: %d", esEvent.GetEventType(), esEvent.GetVersion(), version)
	}
	return nil
}
