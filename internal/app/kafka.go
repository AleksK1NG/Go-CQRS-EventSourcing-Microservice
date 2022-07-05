package app

import (
	"context"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/internal/bankAccount/domain"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/constants"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/es"
	kafkaClient "github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/kafka"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	"net"
	"strconv"
)

func (a *app) connectKafkaBrokers(ctx context.Context) error {
	kafkaConn, err := kafkaClient.NewKafkaConn(ctx, a.cfg.Kafka)
	if err != nil {
		return errors.Wrap(err, "kafka.NewKafkaCon")
	}

	a.kafkaConn = kafkaConn

	brokers, err := kafkaConn.Brokers()
	if err != nil {
		return errors.Wrap(err, "kafkaConn.Brokers")
	}

	a.log.Infof("(kafka connected) brokers: %+v", brokers)
	return nil
}

func (a *app) initKafkaTopics(ctx context.Context) {
	controller, err := a.kafkaConn.Controller()
	if err != nil {
		a.log.Error("kafkaConn.Controller err: %v", err)
		return
	}

	controllerURI := net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port))
	a.log.Infof("(kafka controller uri) controllerURI: %s", controllerURI)

	conn, err := kafka.DialContext(ctx, constants.TCP, controllerURI)
	if err != nil {
		a.log.Errorf("initKafkaTopics.DialContext err: %v", err)
		return
	}
	defer conn.Close() // nolint: errcheck

	a.log.Infof("(established new kafka controller connection) controllerURI: %s", controllerURI)

	bankAccountAggregateTopic := es.GetKafkaAggregateTypeTopic(a.cfg.KafkaPublisherConfig, string(domain.BankAccountAggregateType))

	if err := conn.CreateTopics(bankAccountAggregateTopic); err != nil {
		a.log.WarnErrMsg("kafkaConn.CreateTopics", err)
		return
	}

	if err := conn.CreateTopics(bankAccountAggregateTopic); err != nil {
		a.log.Errorf("(kafkaConn.CreateTopics) err: %v", err)
		return
	}

	a.log.Infof("(kafka topics created or already exists): %+v", []kafka.TopicConfig{bankAccountAggregateTopic})
}

func (a *app) getConsumerGroupTopics() []string {
	topics := []string{
		es.GetTopicName(a.cfg.KafkaPublisherConfig.TopicPrefix, string(domain.BankAccountAggregateType)),
	}

	a.log.Infof("(Consumer Topics) topics: %+v", topics)
	return topics
}
