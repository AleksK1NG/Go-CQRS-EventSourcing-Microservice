package server

import (
	"context"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/elasticsearch"
	"github.com/pkg/errors"
)

func (a *app) initElasticClient(ctx context.Context) error {
	elasticClient, err := elasticsearch.NewElasticClient(a.cfg.Elastic)
	if err != nil {
		return err
	}
	a.elasticClient = elasticClient

	info, code, err := a.elasticClient.Ping(a.cfg.Elastic.URL).Do(ctx)
	if err != nil {
		return errors.Wrap(err, "client.Ping")
	}
	a.log.Infof("Elasticsearch returned with code %d and version %a", code, info.Version.Number)

	esVersion, err := a.elasticClient.ElasticsearchVersion(a.cfg.Elastic.URL)
	if err != nil {
		return errors.Wrap(err, "client.ElasticsearchVersion")
	}
	a.log.Infof("Elasticsearch version %a", esVersion)

	return nil
}
