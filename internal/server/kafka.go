package server

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

func (s *server) connectKafkaBrokers(ctx context.Context) error {
	kafkaConn, err := kafkaClient.NewKafkaConn(ctx, s.cfg.Kafka)
	if err != nil {
		return errors.Wrap(err, "kafka.NewKafkaCon")
	}

	s.kafkaConn = kafkaConn

	brokers, err := kafkaConn.Brokers()
	if err != nil {
		return errors.Wrap(err, "kafkaConn.Brokers")
	}

	s.log.Infof("(kafka connected) brokers: {%+v}", brokers)
	return nil
}

func (s *server) initKafkaTopics(ctx context.Context) {
	controller, err := s.kafkaConn.Controller()
	if err != nil {
		s.log.WarnErrMsg("kafkaConn.Controller", err)
		return
	}

	controllerURI := net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port))
	s.log.Infof("(kafka controller uri): {%s}", controllerURI)

	conn, err := kafka.DialContext(ctx, constants.TCP, controllerURI)
	if err != nil {
		s.log.WarnErrMsg("initKafkaTopics.DialContext", err)
		return
	}
	defer conn.Close() // nolint: errcheck

	s.log.Infof("(established new kafka controller connection) controllerURI: %s", controllerURI)

	bankAccountAggregateTopic := es.GetKafkaAggregateTypeTopic(s.cfg.KafkaPublisherConfig, string(domain.BankAccountAggregateType))

	if err := conn.CreateTopics(bankAccountAggregateTopic); err != nil {
		s.log.WarnErrMsg("kafkaConn.CreateTopics", err)
		return
	}

	if err := conn.CreateTopics(bankAccountAggregateTopic); err != nil {
		s.log.Errorf("(kafkaConn.CreateTopics) err: %v", err)
		return
	}

	s.log.Infof("(kafka topics created or already exists): {%+v}", []kafka.TopicConfig{bankAccountAggregateTopic})
}

func (s *server) getConsumerGroupTopics() []string {
	topics := []string{
		es.GetTopicName(s.cfg.KafkaPublisherConfig.TopicPrefix, string(domain.BankAccountAggregateType)),
	}

	s.log.Infof("(Consumer Topics) topics: {%+v}", topics)
	return topics
}
