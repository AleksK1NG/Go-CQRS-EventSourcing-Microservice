package queries

import (
	"context"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/internal/bankAccount/domain"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/internal/mappers"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/es"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/logger"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/tracing"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
)

type GetBankAccountByIDQuery struct {
	AggregateID    string `json:"aggregateID"`
	FromEventStore bool   `json:"fromEventStore"`
}

type GetBankAccountByID interface {
	Handle(ctx context.Context, query GetBankAccountByIDQuery) (*domain.BankAccountMongoProjection, error)
}

type getBankAccountByIDQuery struct {
	log logger.Logger
	es  es.AggregateStore
	mr  domain.MongoRepository
}

func NewGetBankAccountByIDQuery(log logger.Logger, es es.AggregateStore, mr domain.MongoRepository) *getBankAccountByIDQuery {
	return &getBankAccountByIDQuery{log: log, es: es, mr: mr}
}

func (q *getBankAccountByIDQuery) Handle(ctx context.Context, query GetBankAccountByIDQuery) (*domain.BankAccountMongoProjection, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "getBankAccountByIDQuery.Handle")
	defer span.Finish()
	span.LogFields(log.Object("query", query))

	if query.FromEventStore {
		bankAccountAggregate := domain.NewBankAccountAggregate(query.AggregateID)
		if err := q.es.Load(ctx, bankAccountAggregate); err != nil {
			return nil, tracing.TraceWithErr(span, err)
		}
		q.log.Debugf("GetBankAccountByIDQuery from es bankAccountAggregate: %#+v", bankAccountAggregate.BankAccount)
		return mappers.BankAccountToMongoProjection(bankAccountAggregate.BankAccount), nil
	}

	projection, err := q.mr.GetByAggregateID(ctx, query.AggregateID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			bankAccountAggregate := domain.NewBankAccountAggregate(query.AggregateID)
			if err = q.es.Load(ctx, bankAccountAggregate); err != nil {
				return nil, tracing.TraceWithErr(span, err)
			}

			mongoProjection := mappers.BankAccountToMongoProjection(bankAccountAggregate.BankAccount)
			err = q.mr.Upsert(ctx, mongoProjection)
			if err != nil {
				q.log.Errorf("GetBankAccountByIDQuery mongo Upsert AggregateID: %s, err: %v", query.AggregateID, tracing.TraceWithErr(span, err))
			}
			q.log.Debugf("GetBankAccountByIDQuery Upsert %+v", query)
			return mongoProjection, nil

		}
		return nil, tracing.TraceWithErr(span, err)
	}

	q.log.Debugf("GetBankAccountByIDQuery from mongo %+v", query)
	return projection, nil
}
