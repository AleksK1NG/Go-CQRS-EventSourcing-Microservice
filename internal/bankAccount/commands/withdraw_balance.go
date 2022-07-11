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

type WithdrawBalanceCommand struct {
	AggregateID string `json:"aggregateID" validate:"required,gte=0"`
	Amount      int64  `json:"amount" validate:"required,gte=0"`
	PaymentID   string `json:"paymentID" validate:"required,gte=0"`
}

type WithdrawBalance interface {
	Handle(ctx context.Context, cmd WithdrawBalanceCommand) error
}

type withdrawBalanceCommandHandler struct {
	log            logger.Logger
	aggregateStore es.AggregateStore
}

func NewWithdrawBalanceCommandHandler(log logger.Logger, aggregateStore es.AggregateStore) *withdrawBalanceCommandHandler {
	return &withdrawBalanceCommandHandler{log: log, aggregateStore: aggregateStore}

}

func (c *withdrawBalanceCommandHandler) Handle(ctx context.Context, cmd WithdrawBalanceCommand) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "withdrawBalanceCommandHandler.Handle")
	defer span.Finish()
	span.LogFields(log.Object("command", cmd))

	bankAccountAggregate := domain.NewBankAccountAggregate(cmd.AggregateID)
	err := c.aggregateStore.Load(ctx, bankAccountAggregate)
	if err != nil {
		return tracing.TraceWithErr(span, err)
	}

	if err := bankAccountAggregate.WithdrawBalance(ctx, cmd.Amount, cmd.PaymentID); err != nil {
		return tracing.TraceWithErr(span, err)
	}

	return c.aggregateStore.Save(ctx, bankAccountAggregate)
}
