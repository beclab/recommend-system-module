package storage

import (
	"context"
	"fmt"
	"sort"
	"time"

	"bytetrade.io/web3os/backend-server/common"
	"bytetrade.io/web3os/backend-server/model"
	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func getFeedMongodbColl(s *Storage) *mongo.Collection {
	return s.mongodb.Database(common.GetMongoDbName()).Collection(common.GetMongoFeedColl())
}

func (s *Storage) FeedExists(feedID string) bool {
	coll := getFeedMongodbColl(s)
	id, _ := primitive.ObjectIDFromHex(feedID)
	filter := bson.M{"_id": id}
	count, err := coll.CountDocuments(context.TODO(), filter)
	if err != nil {
		return false
	}
	return count > 0
}

func (s *Storage) GetFeedById(feedID string) (*model.Feed, error) {
	coll := getFeedMongodbColl(s)
	var feed model.Feed
	id, _ := primitive.ObjectIDFromHex(feedID)
	filter := bson.M{"_id": id}
	err := coll.FindOne(context.TODO(), filter).Decode(&feed)
	if err != nil {
		return nil, fmt.Errorf(`store: unable to fetch feed: %v`, err)
	}
	return &feed, nil

}

type CheckFeed struct {
	feedId    string
	checkTime time.Time
}

func (s *Storage) FeedToUpdateList(batchSize int) ([]model.Job, error) {
	coll := getFeedMongodbColl(s)
	filter := bson.M{"sources": common.FeedSource}

	//opts := options.Find().SetProjection(bson.M{"parsing_error_count": 1, "checked_at": 1})

	feeds := make(model.Feeds, 0)
	cur, err := coll.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(context.Background())
	err = cur.All(context.Background(), &feeds)
	if err != nil {
		return nil, err
	}
	common.Logger.Info("FeedToUpdateList ..", zap.Int("size", len(feeds)))
	jobFeedIDs := make([]model.Job, 0)
	checkFeeds := make([]*CheckFeed, 0)
	for _, feed := range feeds {
		errorLimit := common.GetPollingParsingErrorLimit()
		if errorLimit == 0 || feed.ParsingErrorCount < errorLimit {

			t := feed.CheckedAt
			checkFeeds = append(checkFeeds, &CheckFeed{feedId: feed.ID.Hex(), checkTime: t})
		}
	}
	sort.SliceStable(checkFeeds, func(i, j int) bool {
		return checkFeeds[i].checkTime.Before(checkFeeds[j].checkTime)
	})
	for i := 0; i < batchSize && i < len(checkFeeds); i++ {

		jobFeedIDs = append(jobFeedIDs, model.Job{FeedID: checkFeeds[i].feedId})
	}
	return jobFeedIDs, nil
}

func (s *Storage) UpdateFeedError(feedID string, feed *model.Feed) error {
	coll := getFeedMongodbColl(s)
	id, _ := primitive.ObjectIDFromHex(feedID)
	filter := bson.M{"_id": id}

	update := bson.M{"$set": bson.M{"update_at": time.Now(), "parsing_error_count": feed.ParsingErrorCount}}

	if _, err := coll.UpdateOne(context.TODO(), filter, update); err != nil {
		return fmt.Errorf(`store: unable to update feed  (%s): %v`, feed.FeedURL, err)
	}
	return nil
}

func (s *Storage) ResetFeedHeader(feedID string) error {
	coll := getFeedMongodbColl(s)
	id, _ := primitive.ObjectIDFromHex(feedID)
	filter := bson.M{"_id": id}

	update := bson.M{"$set": bson.M{"etag_header": "", "last_modified_header": ""}}

	if _, err := coll.UpdateOne(context.TODO(), filter, update); err != nil {
		return fmt.Errorf(`store: unable to update feed  (%s): %v`, feedID, err)
	}
	return nil
}

// UpdateFeed updates an existing feed.
func (s *Storage) UpdateFeed(feedID string, feed *model.Feed) (err error) {
	coll := getFeedMongodbColl(s)
	id, _ := primitive.ObjectIDFromHex(feedID)
	filter := bson.M{"_id": id}
	update := bson.M{"$set": feed}
	if _, err := coll.UpdateOne(context.TODO(), filter, update); err != nil {
		return fmt.Errorf(`store: unable to update feed  (%s): %v`, feed.FeedURL, err)
	}
	return nil
}
