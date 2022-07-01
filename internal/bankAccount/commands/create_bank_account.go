package commands

import (
	"context"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/internal/bankAccount/domain"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/es"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/logger"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/tracing"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
)

type CreateBankAccountCommand struct {
	AggregateID string `json:"id"`
	Email       string `json:"email"`
	Address     string `json:"address"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	Balance     int64  `json:"balance"`
	Status      string `json:"status"`
}

type CreateBankAccount interface {
	Handle(ctx context.Context, cmd CreateBankAccountCommand) error
}

type createBankAccountCmdHandler struct {
	log logger.Logger
	es  es.AggregateStore
}

func NewCreateBankAccountCmdHandler(log logger.Logger, es es.AggregateStore) *createBankAccountCmdHandler {
	return &createBankAccountCmdHandler{log: log, es: es}
}

func (c *createBankAccountCmdHandler) Handle(ctx context.Context, cmd CreateBankAccountCommand) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "createBankAccountCmdHandler.Handle")
	defer span.Finish()
	span.LogFields(log.Object("command", cmd))

	exists, err := c.es.Exists(ctx, cmd.AggregateID)
	if err != nil {
		return tracing.TraceWithErr(span, err)
	}
	if exists {
		return tracing.TraceWithErr(span, errors.New("already exists"))
	}

	bankAccountAggregate := domain.NewBankAccountAggregate(cmd.AggregateID)
	err = bankAccountAggregate.CreateNewBankAccount(ctx, cmd.Email, cmd.Address, cmd.FirstName, cmd.LastName, cmd.Status, cmd.Balance)
	if err != nil {
		return tracing.TraceWithErr(span, err)
	}

	return c.es.Save(ctx, bankAccountAggregate)
}
