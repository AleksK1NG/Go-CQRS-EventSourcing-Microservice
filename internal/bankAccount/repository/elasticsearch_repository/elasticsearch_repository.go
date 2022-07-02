package elasticsearch_repository

import (
	"context"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/config"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/internal/bankAccount/domain"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/esclient"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/logger"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
)

type elasticRepo struct {
	log    logger.Logger
	cfg    *config.Config
	client *elasticsearch.Client
}

func NewElasticRepo(log logger.Logger, cfg *config.Config, client *elasticsearch.Client) *elasticRepo {
	return &elasticRepo{log: log, cfg: cfg, client: client}
}

func (e *elasticRepo) Index(ctx context.Context, projection *domain.ElasticSearchProjection) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "elasticRepo.Index")
	defer span.Finish()
	span.LogFields(log.String("aggregateID", projection.AggregateID))

	response, err := esclient.Index(ctx, e.client, e.cfg.ElasticIndexes.BankAccounts, projection.AggregateID, projection)
	if err != nil {
		return errors.Wrapf(err, "esclient.Index id: %s", projection.AggregateID)
	}
	defer response.Body.Close()

	if response.IsError() {
		return errors.Wrapf(errors.New("ElasticSearch request err"), "response.IsError id: %s", projection.AggregateID)
	}

	e.log.Infof("ElasticSearch index result: %s", response.String())
	return nil
}

func (e *elasticRepo) Update(ctx context.Context, projection *domain.ElasticSearchProjection) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "elasticRepo.Update")
	defer span.Finish()
	span.LogFields(log.String("aggregateID", projection.AggregateID))

	response, err := esclient.Update(ctx, e.client, e.cfg.ElasticIndexes.BankAccounts, projection.AggregateID, projection)
	if err != nil {
		return errors.Wrapf(err, "esclient.Update id: %s", projection.AggregateID)
	}
	defer response.Body.Close()

	if response.IsError() {
		return errors.Wrapf(errors.New("ElasticSearch request err"), "response.IsError id: %s", projection.AggregateID)
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
		return errors.Wrapf(err, "esclient.Delete id: %s", aggregateID)
	}
	defer response.Body.Close()

	if response.IsError() {
		return errors.Wrapf(errors.New("ElasticSearch request err"), "response.IsError id: %s", aggregateID)
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
		return nil, errors.Wrapf(err, "esclient.GetByID id: %s", aggregateID)
	}

	e.log.Infof("ElasticSearch delete result: %#v", response)
	return response.Source, nil
}

type SearchOptions struct {
	Size   int
	From   int
	Sort   []string
	Fields []string
}

func (e *elasticRepo) Search(ctx context.Context, term string, options SearchOptions) (*esclient.SearchListResponse[*domain.ElasticSearchProjection], error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "elasticRepo.Search")
	defer span.Finish()
	span.LogFields(log.String("term", term))

	response, err := esclient.SearchMultiMatchPrefix[*domain.ElasticSearchProjection](ctx, e.client, esclient.SearchMatchPrefixRequest{
		Index:  []string{e.cfg.ElasticIndexes.BankAccounts},
		Term:   term,
		Size:   options.Size,
		From:   options.From,
		Sort:   options.Sort,
		Fields: options.Fields,
		//SortMap: map[string]interface{}{"amount": "asc"},
	})
	if err != nil {
		return nil, errors.Wrapf(err, "esclient.SearchMultiMatchPrefix term: %s", term)
	}

	return response, nil
}
