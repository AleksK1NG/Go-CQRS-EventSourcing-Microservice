package commands

import (
	"context"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/internal/bankAccount/domain"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/es"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/logger"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

type ChangeEmailCommand struct {
	AggregateID string `json:"aggregateID"`
	NewEmail    string `json:"newEmail"`
}

type ChangeEmail interface {
	Handle(ctx context.Context, cmd ChangeEmailCommand) error
}

type changeEmailCmdHandler struct {
	log            logger.Logger
	aggregateStore es.AggregateStore
}

func NewChangeEmailCmdHandler(log logger.Logger, aggregateStore es.AggregateStore) *changeEmailCmdHandler {
	return &changeEmailCmdHandler{log: log, aggregateStore: aggregateStore}
}

func (c *changeEmailCmdHandler) Handle(ctx context.Context, cmd ChangeEmailCommand) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "changeEmailCmdHandler.Handle")
	defer span.Finish()
	span.LogFields(log.Object("command", cmd))

	bankAccountAggregate := domain.NewBankAccountAggregate(cmd.AggregateID)
	err := c.aggregateStore.Load(ctx, bankAccountAggregate)
	if err != nil {
		return err
	}

	if err := bankAccountAggregate.ChangeEmail(ctx, cmd.NewEmail); err != nil {
		return err
	}

	return c.aggregateStore.Save(ctx, bankAccountAggregate)
}
