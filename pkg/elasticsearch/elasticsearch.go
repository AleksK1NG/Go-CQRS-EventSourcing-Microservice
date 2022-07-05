package elasticsearch

import (
	"context"
	"fmt"
	v7 "github.com/olivere/elastic/v7"
)

type Config struct {
	URL string `mapstructure:"url"`
}

func NewElasticClient(cfg Config) (*v7.Client, error) {
	// Starting with elastic.v5, you must pass a context to execute each service
	ctx := context.Background()

	// Obtain a client and connect to the default Elasticsearch installation
	// on 127.0.0.1:9200. Of course you can configure your client to connect
	// to other hosts and configure it in various other ways.
	client, err := v7.NewClient(
		v7.SetURL(cfg.URL),
		v7.SetSniff(false),
		v7.SetGzip(true),
	)
	if err != nil {
		return nil, err
	}

	// Ping the Elasticsearch app to get e.g. the version number
	info, code, err := client.Ping(cfg.URL).Do(ctx)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Elasticsearch returned with code %d and version %s\n", code, info.Version.Number)

	// Getting the ES version number is quite common, so there's a shortcut
	esVersion, err := client.ElasticsearchVersion(cfg.URL)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Elasticsearch version: %s\n", esVersion)

	return client, nil
}
