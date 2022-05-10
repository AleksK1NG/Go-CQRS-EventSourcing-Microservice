package config

import (
	"flag"
	"fmt"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/constants"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/logger"
	"github.com/pkg/errors"
	"os"

	"github.com/spf13/viper"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config", "", "Writer microservice microservice config path")
}

type Config struct {
	ServiceName string        `mapstructure:"serviceName"`
	Logger      logger.Config `mapstructure:"logger"`
	GRPC        GRPC          `mapstructure:"grpc"`
	//Postgresql           *postgres.Config        `mapstructure:"postgres"`
	//Timeouts             Timeouts                `mapstructure:"timeouts" validate:"required"`
	//EventSourcingConfig  es.Config               `mapstructure:"eventSourcingConfig" validate:"required"`
	//Kafka                *kafkaClient.Config     `mapstructure:"kafka" validate:"required"`
	//KafkaTopics          KafkaTopics             `mapstructure:"kafkaTopics" validate:"required"`
	//Mongo                *mongodb.Config         `mapstructure:"mongo" validate:"required"`
	//MongoCollections     MongoCollections        `mapstructure:"mongoCollections" validate:"required"`
	//KafkaPublisherConfig es.KafkaEventsBusConfig `mapstructure:"kafkaPublisherConfig" validate:"required"`
	//Jaeger               *tracing.Config         `mapstructure:"jaeger"`
	//Elastic              elasticsearch.Config    `mapstructure:"elastic"`
	//ElasticIndexes       ElasticIndexes          `mapstructure:"elasticIndexes"`
	//Projections          Projections             `mapstructure:"projections"`
	Http Http `mapstructure:"http"`
}

type GRPC struct {
	Port        string `mapstructure:"port"`
	Development bool   `mapstructure:"development"`
}

type Timeouts struct {
	PostgresInitMilliseconds int  `mapstructure:"postgresInitMilliseconds" validate:"required"`
	PostgresInitRetryCount   uint `mapstructure:"postgresInitRetryCount" validate:"required"`
}

//type KafkaTopics struct {
//	EventCreated            kafkaClient.TopicConfig `mapstructure:"eventCreated" validate:"required"`
//	OrderAggregateTypeTopic kafkaClient.TopicConfig `mapstructure:"orderAggregateTypeTopic" validate:"required"`
//}

type MongoCollections struct {
	Orders string `mapstructure:"orders" validate:"required"`
}

type ElasticIndexes struct {
	Orders string `mapstructure:"orders" validate:"required"`
}

type Projections struct {
	MongoGroup   string `mapstructure:"mongoGroup" validate:"required"`
	ElasticGroup string `mapstructure:"elasticGroup" validate:"required"`
}

type Http struct {
	Port                string   `mapstructure:"port" validate:"required"`
	Development         bool     `mapstructure:"development"`
	BasePath            string   `mapstructure:"basePath" validate:"required"`
	OrdersPath          string   `mapstructure:"ordersPath" validate:"required"`
	DebugErrorsResponse bool     `mapstructure:"debugErrorsResponse"`
	IgnoreLogUrls       []string `mapstructure:"ignoreLogUrls"`
}

func InitConfig() (*Config, error) {
	if configPath == "" {
		configPathFromEnv := os.Getenv(constants.ConfigPath)
		if configPathFromEnv != "" {
			configPath = configPathFromEnv
		} else {
			getwd, err := os.Getwd()
			if err != nil {
				return nil, errors.Wrap(err, "os.Getwd")
			}
			configPath = fmt.Sprintf("%s/config/config.yaml", getwd)
		}
	}

	cfg := &Config{}

	viper.SetConfigType(constants.Yaml)
	viper.SetConfigFile(configPath)

	if err := viper.ReadInConfig(); err != nil {
		return nil, errors.Wrap(err, "viper.ReadInConfig")
	}

	if err := viper.Unmarshal(cfg); err != nil {
		return nil, errors.Wrap(err, "viper.Unmarshal")
	}

	grpcPort := os.Getenv(constants.GrpcPort)
	if grpcPort != "" {
		cfg.GRPC.Port = grpcPort
	}

	//postgresHost := os.Getenv(constants.PostgresqlHost)
	//if postgresHost != "" {
	//	cfg.Postgresql.Host = postgresHost
	//}
	//postgresPort := os.Getenv(constants.PostgresqlPort)
	//if postgresPort != "" {
	//	cfg.Postgresql.Port = postgresPort
	//}
	//jaegerAddr := os.Getenv(constants.JaegerHostPort)
	//if jaegerAddr != "" {
	//	cfg.Jaeger.HostPort = jaegerAddr
	//}
	//kafkaBrokers := os.Getenv(constants.KafkaBrokers)
	//if kafkaBrokers != "" {
	//	cfg.Kafka.Brokers = []string{kafkaBrokers}
	//}

	return cfg, nil
}
