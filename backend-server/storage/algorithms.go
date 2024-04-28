package storage

import (
	"context"
	"time"

	"bytetrade.io/web3os/backend-server/common"
	"bytetrade.io/web3os/backend-server/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

func getAlgorithmsMongodbColl(s *Storage) *mongo.Collection {
	return s.mongodb.Database(common.GetMongoDbName()).Collection(common.GetMongoAlgorithmsColl())
}
func (s *Storage) createAlgorithms(algorithms *model.Algorithms) (string, error) {
	coll := getAlgorithmsMongodbColl(s)
	algorithms.ID = primitive.NewObjectID()
	algorithms.CreatedAt = time.Now()
	algorithms.UpdatedAt = time.Now()
	if _, err := coll.InsertOne(context.TODO(), algorithms); err != nil {
		common.Logger.Error("store: store: unable to create algorithms", zap.String("entryId", algorithms.Entry), zap.Error(err))
		return "", err
	}
	return algorithms.ID.Hex(), nil
}
