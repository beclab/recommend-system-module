package storge

import (
	"context"
	"fmt"

	"bytetrade.io/web3os/system_workflow/common"
	"bytetrade.io/web3os/system_workflow/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func RemoveDiscoveryFeed(mongoClient *mongo.Client) error {
	coll := mongoClient.Database(common.MONGO_DB_NAME()).Collection("discoveries")
	_, err := coll.DeleteMany(context.TODO(), bson.D{})
	if err != nil {
		return err
	}
	return nil
}

func CreateDiscoveryFeed(mongoClient *mongo.Client, discovery *model.Discovery) error {
	coll := mongoClient.Database(common.MONGO_DB_NAME()).Collection("discoveries")
	discovery.ID = primitive.NewObjectID()
	if _, err := coll.InsertOne(context.TODO(), discovery); err != nil {
		return fmt.Errorf(`store: unable to create discovery feed  (%s): %v`, discovery.FeedUrl, err)

	}
	return nil
}
