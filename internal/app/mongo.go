package app

import (
	"context"
	serviceErrors "github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/service_errors"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (a *app) initMongoDBCollections(ctx context.Context) {
	err := a.mongoClient.Database(a.cfg.Mongo.Db).CreateCollection(ctx, a.cfg.MongoCollections.BankAccounts)
	if err != nil {
		if !utils.CheckErrForMessagesCaseInSensitive(err, serviceErrors.ErrMsgMongoCollectionAlreadyExists) {
			a.log.Warnf("(CreateCollection) err: %v", err)
		}
	}

	aggregateIdIndexOptions := options.Index().
		SetSparse(true).
		SetUnique(true)

	aggregateIdIndex, err := a.mongoClient.Database(a.cfg.Mongo.Db).
		Collection(a.cfg.MongoCollections.BankAccounts).
		Indexes().
		CreateOne(ctx, mongo.IndexModel{
			Keys:    bson.D{{Key: "aggregateID", Value: 1}},
			Options: aggregateIdIndexOptions,
		})
	if err != nil && !utils.CheckErrForMessagesCaseInSensitive(err, serviceErrors.ErrMsgAlreadyExists) {
		a.log.Warnf("(CreateOne) err: %v", err)
	}
	a.log.Infof("(CreatedIndex) aggregateIdIndex: %s", aggregateIdIndex)

	emailIndexOptions := options.Index().
		SetSparse(true).
		SetUnique(true)

	emailIndex, err := a.mongoClient.Database(a.cfg.Mongo.Db).
		Collection(a.cfg.MongoCollections.BankAccounts).
		Indexes().
		CreateOne(ctx, mongo.IndexModel{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: emailIndexOptions,
		})
	if err != nil && !utils.CheckErrForMessagesCaseInSensitive(err, serviceErrors.ErrMsgAlreadyExists) {
		a.log.Warnf("(CreateOne) err: %v", err)
	}
	a.log.Infof("(CreatedIndex) emailIndex: %s", emailIndex)

	list, err := a.mongoClient.Database(a.cfg.Mongo.Db).Collection(a.cfg.MongoCollections.BankAccounts).
		Indexes().
		List(ctx)
	if err != nil {
		a.log.Warnf("(initMongoDBCollections) [List] err: %v", err)
	}

	if list != nil {
		var results []bson.M
		if err := list.All(ctx, &results); err != nil {
			a.log.Warnf("(All) err: %v", err)
		}
		a.log.Infof("(indexes) results: %+v", results)
	}

	collections, err := a.mongoClient.Database(a.cfg.Mongo.Db).ListCollectionNames(ctx, bson.M{})
	if err != nil {
		a.log.Warnf("(ListCollections) err: %v", err)
	}
	a.log.Infof("(Collections) created collections: %+v", collections)
}
