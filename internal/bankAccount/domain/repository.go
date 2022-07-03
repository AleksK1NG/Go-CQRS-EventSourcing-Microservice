package domain

import (
	"context"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/esclient"
)

type UpdateProjectionCallback func(projection *BankAccountMongoProjection) *BankAccountMongoProjection

type MongoRepository interface {
	Insert(ctx context.Context, projection *BankAccountMongoProjection) error
	Update(ctx context.Context, projection *BankAccountMongoProjection) error
	Upsert(ctx context.Context, projection *BankAccountMongoProjection) error
	DeleteByAggregateID(ctx context.Context, aggregateID string) error
	UpdateConcurrently(ctx context.Context, aggregateID string, updateCb UpdateProjectionCallback, expectedVersion uint64) error

	GetByAggregateID(ctx context.Context, aggregateID string) (*BankAccountMongoProjection, error)
}

type ElasticSearchRepository interface {
	Index(ctx context.Context, projection *ElasticSearchProjection) error
	Update(ctx context.Context, projection *ElasticSearchProjection) error
	DeleteByAggregateID(ctx context.Context, aggregateID string) error

	GetByAggregateID(ctx context.Context, aggregateID string) (*ElasticSearchProjection, error)
	Search(ctx context.Context, term string, options esclient.SearchOptions) (*esclient.SearchListResponse[*ElasticSearchProjection], error)
}
