package mongo_repository

import (
	"context"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/config"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/internal/bankAccount/domain"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/logger"
	"go.mongodb.org/mongo-driver/mongo"
)

type bankAccountMongoRepository struct {
	log logger.Logger
	cfg *config.Config
	db  *mongo.Client
}

func NewBankAccountMongoRepository(log logger.Logger, cfg *config.Config, db *mongo.Client) *bankAccountMongoRepository {
	return &bankAccountMongoRepository{log: log, cfg: cfg, db: db}
}

func (b *bankAccountMongoRepository) Insert(ctx context.Context, projection domain.BankAccountMongoProjection) error {

	//TODO implement me
	panic("implement me")
}

func (b *bankAccountMongoRepository) Update(ctx context.Context, projection domain.BankAccountMongoProjection) error {
	//TODO implement me
	panic("implement me")
}

func (b *bankAccountMongoRepository) GetByAggregateID(ctx context.Context, aggregateID string) (*domain.BankAccountMongoProjection, error) {
	//TODO implement me
	panic("implement me")
}
