package server

import (
	"context"
	"fmt"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/config"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/constants"
	serviceErrors "github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/service_errors"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strings"
	"time"
)

const (
	waitShotDownDuration = 3 * time.Second
)

func (s *server) initMongoDBCollections(ctx context.Context) {
	err := s.mongoClient.Database(s.cfg.Mongo.Db).CreateCollection(ctx, s.cfg.MongoCollections.BankAccounts)
	if err != nil {
		if !utils.CheckErrForMessagesCaseInSensitive(err, serviceErrors.ErrMsgMongoCollectionAlreadyExists) {
			s.log.Warnf("(CreateCollection) err: {%v}", err)
		}
	}

	indexOptions := options.Index().SetSparse(true).SetUnique(true)
	index, err := s.mongoClient.Database(s.cfg.Mongo.Db).Collection(s.cfg.MongoCollections.BankAccounts).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: constants.BankAccountIndex, Value: 1}},
		Options: indexOptions,
	})
	if err != nil && !utils.CheckErrForMessagesCaseInSensitive(err, serviceErrors.ErrMsgAlreadyExists) {
		s.log.Warnf("(CreateOne) err: {%v}", err)
	}
	s.log.Infof("(CreatedIndex) index: {%s}", index)

	list, err := s.mongoClient.Database(s.cfg.Mongo.Db).Collection(s.cfg.MongoCollections.BankAccounts).Indexes().List(ctx)
	if err != nil {
		s.log.Warnf("(initMongoDBCollections) [List] err: {%v}", err)
	}

	if list != nil {
		var results []bson.M
		if err := list.All(ctx, &results); err != nil {
			s.log.Warnf("(All) err: {%v}", err)
		}
		s.log.Infof("(indexes) results: {%#v}", results)
	}

	collections, err := s.mongoClient.Database(s.cfg.Mongo.Db).ListCollectionNames(ctx, bson.M{})
	if err != nil {
		s.log.Warnf("(ListCollections) err: {%v}", err)
	}
	s.log.Infof("(Collections) created collections: {%v}", collections)
}

func (s *server) getHttpMetricsCb() func(err error) {
	return func(err error) {
		if err != nil {
			s.metrics.ErrorHttpRequests.Inc()
		} else {
			s.metrics.SuccessHttpRequests.Inc()
		}
	}
}

func (s *server) getGrpcMetricsCb() func(err error) {
	return func(err error) {
		if err != nil {
			s.metrics.ErrorGrpcRequests.Inc()
		} else {
			s.metrics.SuccessGrpcRequests.Inc()
		}
	}
}

func (s *server) waitShootDown(duration time.Duration) {
	go func() {
		time.Sleep(duration)
		s.doneCh <- struct{}{}
	}()
}

func GetMicroserviceName(cfg config.Config) string {
	return fmt.Sprintf("(%s)", strings.ToUpper(cfg.ServiceName))
}
