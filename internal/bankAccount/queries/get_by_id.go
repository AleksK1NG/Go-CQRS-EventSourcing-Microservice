package queries

import (
	"context"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/internal/bankAccount/domain"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/es"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/logger"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

type GetBankAccountByIDQuery struct {
	AggregateID string `json:"aggregateID"`
}

type GetBankAccountByID interface {
	Handle(ctx context.Context, query GetBankAccountByIDQuery) (*domain.BankAccount, error)
}

type getBankAccountByIDQuery struct {
	log logger.Logger
	es  es.AggregateStore
}

func NewGetBankAccountByIDQuery(log logger.Logger, es es.AggregateStore) *getBankAccountByIDQuery {
	return &getBankAccountByIDQuery{log: log, es: es}
}

func (q *getBankAccountByIDQuery) Handle(ctx context.Context, query GetBankAccountByIDQuery) (*domain.BankAccount, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "getBankAccountByIDQuery.Handle")
	defer span.Finish()
	span.LogFields(log.Object("query", query))

	bankAccountAggregate := domain.NewBankAccountAggregate(query.AggregateID)
	if err := q.es.Load(ctx, bankAccountAggregate); err != nil {
		return nil, err
	}

	return bankAccountAggregate.BankAccount, nil
}
