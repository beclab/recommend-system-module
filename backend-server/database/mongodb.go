package database

import (
	"context"

	"bytetrade.io/web3os/backend-server/common"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func NewMongodbConnection() (*mongo.Client, error) {
	uri := common.GetMongoURI()
	var mongoClient *mongo.Client
	var connectMongoClientError error

	mongoClient, connectMongoClientError = mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if connectMongoClientError != nil {
		return nil, connectMongoClientError
	}

	if connectMongoClientError = mongoClient.Ping(context.Background(), nil); connectMongoClientError != nil {
		return nil, connectMongoClientError
	}
	return mongoClient, connectMongoClientError
}
