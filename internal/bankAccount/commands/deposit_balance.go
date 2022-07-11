package commands

import (
	"context"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/internal/bankAccount/domain"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/es"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/logger"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/tracing"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

type DepositBalanceCommand struct {
	AggregateID string `json:"aggregateID" validate:"required,gte=0"`
	Amount      int64  `json:"amount" validate:"required,gte=0"`
	PaymentID   string `json:"paymentID" validate:"required,gte=0"`
}

type DepositBalance interface {
	Handle(ctx context.Context, cmd DepositBalanceCommand) error
}

type depositBalanceCmdHandler struct {
	log            logger.Logger
	aggregateStore es.AggregateStore
}

func NewDepositBalanceCmdHandler(log logger.Logger, aggregateStore es.AggregateStore) *depositBalanceCmdHandler {
	return &depositBalanceCmdHandler{log: log, aggregateStore: aggregateStore}
}

func (c *depositBalanceCmdHandler) Handle(ctx context.Context, cmd DepositBalanceCommand) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "depositBalanceCmdHandler.Handle")
	defer span.Finish()
	span.LogFields(log.Object("command", cmd))

	bankAccountAggregate := domain.NewBankAccountAggregate(cmd.AggregateID)
	err := c.aggregateStore.Load(ctx, bankAccountAggregate)
	if err != nil {
		return tracing.TraceWithErr(span, err)
	}

	if err := bankAccountAggregate.DepositBalance(ctx, cmd.Amount, cmd.PaymentID); err != nil {
		return tracing.TraceWithErr(span, err)
	}

	return c.aggregateStore.Save(ctx, bankAccountAggregate)
}
