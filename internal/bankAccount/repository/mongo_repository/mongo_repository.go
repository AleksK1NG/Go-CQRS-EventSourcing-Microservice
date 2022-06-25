package mongo_repository

import (
	"context"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/config"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/internal/bankAccount/domain"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/constants"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/logger"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/tracing"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type bankAccountMongoRepository struct {
	log logger.Logger
	cfg *config.Config
	db  *mongo.Client
}

func NewBankAccountMongoRepository(log logger.Logger, cfg *config.Config, db *mongo.Client) *bankAccountMongoRepository {
	return &bankAccountMongoRepository{log: log, cfg: cfg, db: db}
}

func (b *bankAccountMongoRepository) Insert(ctx context.Context, projection *domain.BankAccountMongoProjection) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "bankAccountMongoRepository.Insert")
	defer span.Finish()
	span.LogFields(log.String("aggregateID", projection.AggregateID))

	_, err := b.bankAccountsCollection().InsertOne(ctx, projection)
	if err != nil {
		return tracing.TraceWithErr(span, errors.Wrapf(err, "[InsertOne] AggregateID: %s", projection.AggregateID))
	}

	b.log.Debugf("[Insert] result AggregateID: %s", projection.AggregateID)
	return nil
}

func (b *bankAccountMongoRepository) Update(ctx context.Context, projection *domain.BankAccountMongoProjection) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "bankAccountMongoRepository.Update")
	defer span.Finish()
	span.LogFields(log.String("aggregateID", projection.AggregateID))

	projection.ID = ""
	projection.UpdatedAt = time.Now().UTC()

	ops := options.FindOneAndUpdate()
	ops.SetReturnDocument(options.After)
	ops.SetUpsert(false)
	filter := bson.M{constants.MongoAggregateID: projection.AggregateID}

	err := b.bankAccountsCollection().FindOneAndUpdate(ctx, filter, bson.M{"$set": projection}, ops).Decode(projection)
	if err != nil {
		return tracing.TraceWithErr(span, errors.Wrapf(err, "[FindOneAndUpdate] aggregateID: %s", projection.AggregateID))
	}

	b.log.Debugf("[Update] result AggregateID: %s", projection.AggregateID)
	return nil
}

func (b *bankAccountMongoRepository) GetByAggregateID(ctx context.Context, aggregateID string) (*domain.BankAccountMongoProjection, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "bankAccountMongoRepository.GetByAggregateID")
	defer span.Finish()
	span.LogFields(log.String("aggregateID", aggregateID))

	filter := bson.M{constants.MongoAggregateID: aggregateID}
	var projection domain.BankAccountMongoProjection
	err := b.bankAccountsCollection().FindOne(ctx, filter).Decode(&projection)
	if err != nil {
		return nil, tracing.TraceWithErr(span, errors.Wrapf(err, "[FindOne] aggregateID: %s", projection.AggregateID))
	}

	b.log.Debugf("[GetByAggregateID] result projection: %+v", projection)
	return &projection, nil
}

func (b *bankAccountMongoRepository) bankAccountsCollection() *mongo.Collection {
	return b.db.Database(b.cfg.Mongo.Db).Collection(b.cfg.MongoCollections.BankAccounts)
}
