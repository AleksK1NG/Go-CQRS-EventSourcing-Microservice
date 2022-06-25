package domain

import "context"

type MongoRepository interface {
	Insert(ctx context.Context, projection *BankAccountMongoProjection) error
	Update(ctx context.Context, projection *BankAccountMongoProjection) error
	Upsert(ctx context.Context, projection *BankAccountMongoProjection) error
	DeleteByAggregateID(ctx context.Context, aggregateID string) error

	GetByAggregateID(ctx context.Context, aggregateID string) (*BankAccountMongoProjection, error)
}
