package commands

import (
	"context"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/internal/bankAccount/domain"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/es"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/logger"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

type DepositBalanceCommand struct {
	AggregateID string `json:"aggregateID"`
	Amount      int64  `json:"amount"`
	PaymentID   string `json:"paymentID"`
}

type DepositBalance interface {
	Handle(ctx context.Context, cmd DepositBalanceCommand) error
}

type depositBalanceCmdHandler struct {
	log logger.Logger
	es  es.AggregateStore
}

func NewDepositBalanceCmdHandler(log logger.Logger, es es.AggregateStore) *depositBalanceCmdHandler {
	return &depositBalanceCmdHandler{log: log, es: es}
}

func (c *depositBalanceCmdHandler) Handle(ctx context.Context, cmd DepositBalanceCommand) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "depositBalanceCmdHandler.Handle")
	defer span.Finish()
	span.LogFields(log.Object("command", cmd))

	bankAccountAggregate := domain.NewBankAccountAggregate(cmd.AggregateID)
	err := c.es.Load(ctx, bankAccountAggregate)
	if err != nil {
		return err
	}

	if err := bankAccountAggregate.DepositBalance(ctx, cmd.Amount, cmd.PaymentID); err != nil {
		return err
	}

	return c.es.Save(ctx, bankAccountAggregate)
}
