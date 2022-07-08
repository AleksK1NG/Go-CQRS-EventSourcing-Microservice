package queries

import (
	"context"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/internal/bankAccount/domain"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/es"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/esclient"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/logger"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/tracing"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/utils"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

type SearchQueryResult struct {
	List               []*domain.ElasticSearchProjection `json:"list"`
	PaginationResponse *utils.PaginationResponse         `json:"paginationResponse"`
}

type SearchBankAccountsQuery struct {
	QueryTerm  string            `json:"queryTerm" validate:"required,gte=0"`
	Pagination *utils.Pagination `json:"pagination"`
}

type SearchBankAccounts interface {
	Handle(ctx context.Context, query SearchBankAccountsQuery) (*SearchQueryResult, error)
}

type searchBankAccountsQuery struct {
	log               logger.Logger
	aggregateStore    es.AggregateStore
	elasticRepository domain.ElasticSearchRepository
}

func NewSearchBankAccountsQuery(log logger.Logger, aggregateStore es.AggregateStore, elasticRepository domain.ElasticSearchRepository) *searchBankAccountsQuery {
	return &searchBankAccountsQuery{log: log, aggregateStore: aggregateStore, elasticRepository: elasticRepository}
}

func (q *searchBankAccountsQuery) Handle(ctx context.Context, query SearchBankAccountsQuery) (*SearchQueryResult, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "searchBankAccountsQuery.Handle")
	defer span.Finish()
	span.LogFields(log.Object("query", query))

	options := esclient.SearchOptions{
		Size:   query.Pagination.GetSize(),
		From:   query.Pagination.GetOffset(),
		Fields: []string{"firstName", "lastName", "email"},
	}

	result, err := q.elasticRepository.Search(ctx, query.QueryTerm, options)
	if err != nil {
		return nil, tracing.TraceWithErr(span, err)
	}

	return &SearchQueryResult{
		List:               result.List,
		PaginationResponse: utils.NewPaginationResponse(result.Total, query.Pagination),
	}, nil
}
