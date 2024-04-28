package storge

import (
	"context"
	"fmt"

	"bytetrade.io/web3os/system_workflow/common"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

func UpdateFeed(mongoClient *mongo.Client, sources []string, updateFeedList map[string]map[string]interface{}) {

	feedCollection := mongoClient.Database(common.MONGO_DB_NAME()).Collection(common.FEED_COLLECTION_NAME())
	for _, source := range sources {
		for _, updateFeed := range updateFeedList {
			filters := make([]bson.M, 0)
			filters = append(filters, bson.M{"sources": source})
			filters = append(filters, bson.M{"feed_url": updateFeed["feed_url"]})
			filter := bson.M{"$and": filters}

			var updateDoc bson.D

			for field, value := range updateFeed {
				if field != "feed_url" {
					updateDoc = append(updateDoc, bson.E{Key: field, Value: value})
				}
			}
			/*_, siteUrlExist := updateFeed["site_url"]
			if siteUrlExist {
				updates = append(updates, bson.M{"site_url": updateFeed["site_url"]})
			}
			_, titleExist := updateFeed["title"]
			if titleExist {
				updates = append(updates, bson.M{"title": updateFeed["title"]})
			}
			_, iconExist := updateFeed["icon_content"]
			if iconExist {
				updates = append(updates, bson.M{"icon_content": updateFeed["icon_content"]})
			}
			_, lastModifyExist := updateFeed["last_modify_time"]
			if lastModifyExist {
				updates = append(updates, bson.M{"last_modify_time": updateFeed["last_modify_time"]})
			}

			update := bson.M{"$set": updates}*/
			if _, err := feedCollection.UpdateOne(context.TODO(), filter, bson.D{{Key: "$set", Value: updateDoc}}); err != nil {
				common.Logger.Error("unable to update feed ", zap.String("feed url", fmt.Sprintf("%v", updateFeed["feed_url"])), zap.Error(err))
			}
		}
	}
}
