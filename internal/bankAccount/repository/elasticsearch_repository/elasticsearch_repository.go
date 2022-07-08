package elasticsearch_repository

import (
	"context"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/config"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/internal/bankAccount/domain"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/esclient"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/logger"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/tracing"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
	"net/http"
	"time"
)

type elasticRepo struct {
	log    logger.Logger
	cfg    *config.Config
	client *elasticsearch.Client
}

func NewElasticRepository(log logger.Logger, cfg *config.Config, client *elasticsearch.Client) *elasticRepo {
	return &elasticRepo{log: log, cfg: cfg, client: client}
}

func (e *elasticRepo) Index(ctx context.Context, projection *domain.ElasticSearchProjection) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "elasticRepo.Index")
	defer span.Finish()
	span.LogFields(log.String("aggregateID", projection.AggregateID))

	response, err := esclient.Index(ctx, e.client, e.cfg.ElasticIndexes.BankAccounts, projection.AggregateID, projection)
	if err != nil {
		return tracing.TraceWithErr(span, errors.Wrapf(err, "esclient.Index id: %s", projection.AggregateID))
	}
	defer response.Body.Close()

	if response.IsError() {
		return tracing.TraceWithErr(span, errors.Wrapf(errors.New("ElasticSearch request err"), "response.IsError id: %s", projection.AggregateID))
	}
	if response.HasWarnings() {
		e.log.Warnf("ElasticSearch Index warnings: %+v", response.Warnings())
	}

	e.log.Infof("ElasticSearch index result: %s", response.String())
	return nil
}

func (e *elasticRepo) Update(ctx context.Context, projection *domain.ElasticSearchProjection) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "elasticRepo.Update")
	defer span.Finish()
	span.LogFields(log.String("aggregateID", projection.AggregateID))

	projection.UpdatedAt = time.Now().UTC()

	response, err := esclient.Update(ctx, e.client, e.cfg.ElasticIndexes.BankAccounts, projection.AggregateID, projection)
	if err != nil {
		return tracing.TraceWithErr(span, errors.Wrapf(err, "esclient.Update id: %s", projection.AggregateID))
	}
	defer response.Body.Close()

	if response.IsError() {
		return tracing.TraceWithErr(span, errors.Wrapf(errors.New("ElasticSearch request err"), "response.IsError id: %s", projection.AggregateID))
	}
	if response.HasWarnings() {
		e.log.Warnf("ElasticSearch Update warnings: %+v", response.Warnings())
	}

	e.log.Infof("ElasticSearch update result: %s", response.String())
	return nil
}

func (e *elasticRepo) DeleteByAggregateID(ctx context.Context, aggregateID string) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "elasticRepo.DeleteByAggregateID")
	defer span.Finish()
	span.LogFields(log.String("aggregateID", aggregateID))

	response, err := esclient.Delete(ctx, e.client, e.cfg.ElasticIndexes.BankAccounts, aggregateID)
	if err != nil {
		return tracing.TraceWithErr(span, errors.Wrapf(err, "esclient.Delete id: %s", aggregateID))
	}
	defer response.Body.Close()

	if response.IsError() && response.StatusCode != http.StatusNotFound {
		return tracing.TraceWithErr(span, errors.Wrapf(errors.New("ElasticSearch delete"), "response.IsError aggregateID: %s, status: %s", aggregateID, response.Status()))
	}
	if response.HasWarnings() {
		e.log.Warnf("ElasticSearch Delete warnings: %+v", response.Warnings())
	}

	e.log.Infof("ElasticSearch delete result: %s", response.String())
	return nil
}

func (e *elasticRepo) GetByAggregateID(ctx context.Context, aggregateID string) (*domain.ElasticSearchProjection, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "elasticRepo.GetByAggregateID")
	defer span.Finish()
	span.LogFields(log.String("aggregateID", aggregateID))

	response, err := esclient.GetByID[*domain.ElasticSearchProjection](ctx, e.client, e.cfg.ElasticIndexes.BankAccounts, aggregateID)
	if err != nil {
		return nil, tracing.TraceWithErr(span, errors.Wrapf(err, "esclient.GetByID id: %s", aggregateID))
	}

	e.log.Infof("ElasticSearch delete result: %+v", response)
	return response.Source, nil
}

func (e *elasticRepo) Search(ctx context.Context, term string, options esclient.SearchOptions) (*esclient.SearchListResponse[*domain.ElasticSearchProjection], error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "elasticRepo.Search")
	defer span.Finish()
	span.LogFields(log.String("term", term))

	searchMatchPrefixRequest := esclient.SearchMatchPrefixRequest{
		Index:   []string{e.cfg.ElasticIndexes.BankAccounts},
		Term:    term,
		Size:    options.Size,
		From:    options.From,
		Sort:    []string{"balance.amount"},
		Fields:  options.Fields,
		SortMap: map[string]interface{}{"balance.amount": "asc"},
	}

	if options.Sort != nil {
		searchMatchPrefixRequest.Sort = options.Sort
	}

	response, err := esclient.SearchMultiMatchPrefix[*domain.ElasticSearchProjection](ctx, e.client, searchMatchPrefixRequest)
	if err != nil {
		return nil, tracing.TraceWithErr(span, errors.Wrapf(err, "esclient.SearchMultiMatchPrefix term: %s", term))
	}

	return response, nil
}
