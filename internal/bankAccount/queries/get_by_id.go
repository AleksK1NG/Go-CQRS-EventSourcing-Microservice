package queries

import (
	"context"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/internal/bankAccount/domain"
	bankAccountErrors "github.com/AleksK1NG/go-cqrs-eventsourcing/internal/bankAccount/errors"
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
	AggregateID    string `json:"aggregateID" validate:"required,gte=0"`
	FromEventStore bool   `json:"fromEventStore"`
}

type GetBankAccountByID interface {
	Handle(ctx context.Context, query GetBankAccountByIDQuery) (*domain.BankAccountMongoProjection, error)
}

type getBankAccountByIDQuery struct {
	log             logger.Logger
	aggregateStore  es.AggregateStore
	mongoRepository domain.MongoRepository
}

func NewGetBankAccountByIDQuery(log logger.Logger, aggregateStore es.AggregateStore, mongoRepository domain.MongoRepository) *getBankAccountByIDQuery {
	return &getBankAccountByIDQuery{log: log, aggregateStore: aggregateStore, mongoRepository: mongoRepository}
}

func (q *getBankAccountByIDQuery) Handle(ctx context.Context, query GetBankAccountByIDQuery) (*domain.BankAccountMongoProjection, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "getBankAccountByIDQuery.Handle")
	defer span.Finish()
	span.LogFields(log.Object("query", query))

	if query.FromEventStore {
		return q.loadFromAggregateStore(ctx, query)
	}

	projection, err := q.mongoRepository.GetByAggregateID(ctx, query.AggregateID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			bankAccountAggregate := domain.NewBankAccountAggregate(query.AggregateID)
			if err = q.aggregateStore.Load(ctx, bankAccountAggregate); err != nil {
				return nil, tracing.TraceWithErr(span, err)
			}
			if bankAccountAggregate.GetVersion() == 0 {
				return nil, tracing.TraceWithErr(span, errors.Wrapf(bankAccountErrors.ErrBankAccountNotFound, "id: %s", query.AggregateID))
			}

			mongoProjection := mappers.BankAccountToMongoProjection(bankAccountAggregate)
			err = q.mongoRepository.Upsert(ctx, mongoProjection)
			if err != nil {
				q.log.Errorf("(GetBankAccountByIDQuery) mongo Upsert AggregateID: %s, err: %v", query.AggregateID, tracing.TraceWithErr(span, err))
			}
			q.log.Debugf("(GetBankAccountByIDQuery) Upsert %+v", query)
			return mongoProjection, nil

		}
		return nil, tracing.TraceWithErr(span, err)
	}

	q.log.Debugf("(GetBankAccountByIDQuery) from mongo %+v", query)
	return projection, nil
}

func (q *getBankAccountByIDQuery) loadFromAggregateStore(ctx context.Context, query GetBankAccountByIDQuery) (*domain.BankAccountMongoProjection, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "getBankAccountByIDQuery.loadFromAggregateStore")
	defer span.Finish()

	bankAccountAggregate := domain.NewBankAccountAggregate(query.AggregateID)
	if err := q.aggregateStore.Load(ctx, bankAccountAggregate); err != nil {
		return nil, tracing.TraceWithErr(span, err)
	}
	if bankAccountAggregate.GetVersion() == 0 {
		return nil, tracing.TraceWithErr(span, errors.Wrapf(bankAccountErrors.ErrBankAccountNotFound, "id: %s", query.AggregateID))
	}

	q.log.Debugf("(GetBankAccountByIDQuery) from aggregateStore bankAccountAggregate: %+v", bankAccountAggregate.BankAccount)
	return mappers.BankAccountToMongoProjection(bankAccountAggregate), nil
}
