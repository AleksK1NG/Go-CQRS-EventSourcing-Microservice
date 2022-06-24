package mongo_projection

import (
	"context"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/config"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/internal/bankAccount/domain"
	bankAccountErrors "github.com/AleksK1NG/go-cqrs-eventsourcing/internal/bankAccount/errors"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/internal/bankAccount/events"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/es"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/logger"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
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

func (b *bankAccountMongoProjection) When(ctx context.Context, event any) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "bankAccountMongoProjection.bankAccountMongoProjection")
	defer span.Finish()

	switch evt := event.(type) {

	case *events.BankAccountCreatedEventV1:
		return b.onBankAccountCreated(ctx, evt)

	case *events.BalanceDepositedEventV1:
		return b.onBalanceDeposited(ctx, evt)

	case *events.EmailChangedEventV1:
		return b.onEmailChanged(ctx, evt)

	default:
		return errors.Wrapf(bankAccountErrors.ErrUnknownEventType, "event: %+v", event)
	}
}

func (b *bankAccountMongoProjection) onBankAccountCreated(ctx context.Context, event *events.BankAccountCreatedEventV1) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "bankAccountMongoProjection.onBankAccountCreated")
	defer span.Finish()

	return nil
}

func (b *bankAccountMongoProjection) onBalanceDeposited(ctx context.Context, event *events.BalanceDepositedEventV1) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "bankAccountMongoProjection.onBalanceDeposited")
	defer span.Finish()

	return nil
}

func (b *bankAccountMongoProjection) onEmailChanged(ctx context.Context, event *events.EmailChangedEventV1) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "bankAccountMongoProjection.onEmailChanged")
	defer span.Finish()

	return nil
}
