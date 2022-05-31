package server

import (
	"context"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/elasticsearch"
	"github.com/pkg/errors"
)

func (s *server) initElasticClient(ctx context.Context) error {
	elasticClient, err := elasticsearch.NewElasticClient(s.cfg.Elastic)
	if err != nil {
		return err
	}
	s.elasticClient = elasticClient

	info, code, err := s.elasticClient.Ping(s.cfg.Elastic.URL).Do(ctx)
	if err != nil {
		return errors.Wrap(err, "client.Ping")
	}
	s.log.Infof("Elasticsearch returned with code %d and version %s", code, info.Version.Number)

	esVersion, err := s.elasticClient.ElasticsearchVersion(s.cfg.Elastic.URL)
	if err != nil {
		return errors.Wrap(err, "client.ElasticsearchVersion")
	}
	s.log.Infof("Elasticsearch version %s", esVersion)

	return nil
}
