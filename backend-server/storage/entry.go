package storage

import (
	"context"
	"fmt"

	"bytetrade.io/web3os/backend-server/common"
	"bytetrade.io/web3os/backend-server/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

func getEntryMongodbColl(s *Storage) *mongo.Collection {
	return s.mongodb.Database(common.GetMongoDbName()).Collection(common.GetMongoEntryColl())
}
func (s *Storage) GetEntryDocList(entryIDs []string) ([]string, error) {
	coll := getEntryMongodbColl(s)
	filter := bson.D{}
	filter = append(filter, bson.E{Key: "_id", Value: bson.M{"$in": entryIDs}})

	opts := options.Find().SetProjection(bson.D{{"doc_id", 1}})
	findEntryCursor, findErr := coll.Find(context.TODO(), filter, opts)
	if findErr != nil {
		common.Logger.Error("find entry list fail", zap.Error(findErr))
		return nil, findErr
	}
	docIDList := make([]string, 0)
	for findEntryCursor.Next(context.TODO()) {
		var currentEntry model.Entry
		if decodeErr := findEntryCursor.Decode(&currentEntry); decodeErr != nil {
			common.Logger.Error("decode current_entry fail", zap.Error(decodeErr))
			return nil, decodeErr
		}
		docIDList = append(docIDList, currentEntry.DocId)
	}
	return docIDList, nil
}

func (s *Storage) GetEntryById(entryID string) (*model.Entry, error) {
	coll := getEntryMongodbColl(s)
	var entry model.Entry
	id, _ := primitive.ObjectIDFromHex(entryID)
	filter := bson.M{"_id": id}
	err := coll.FindOne(context.TODO(), filter).Decode(&entry)
	if err != nil {
		return nil, fmt.Errorf(`store: unable to fetch entry: %v`, err)
	}
	return &entry, nil

}
func (s *Storage) GetEntryByUrl(feedID, url string) *model.Entry {
	coll := getEntryMongodbColl(s)
	var entry model.Entry
	fid, _ := primitive.ObjectIDFromHex(feedID)
	query := make(map[string]interface{})
	query["feed"] = fid
	query["url"] = url

	err := coll.FindOne(context.TODO(), query).Decode(&entry)
	if err != nil {
		return nil
	}
	return &entry
}

func (s *Storage) EntryInSourceExists(feedID, url string) bool {
	coll := getEntryMongodbColl(s)
	fid, _ := primitive.ObjectIDFromHex(feedID)
	query := make(map[string]interface{})
	query["feed"] = fid
	query["sources"] = common.FeedSource
	query["url"] = url

	count, err := coll.CountDocuments(context.TODO(), query)
	if err != nil {
		return true
	}
	return count > 0
}

func (s *Storage) UpdateEntryContent(entry *model.Entry) error {
	coll := getEntryMongodbColl(s)
	filter := bson.M{"_id": entry.ID}
	update := bson.M{"$set": bson.M{"crawler": true, "title": entry.Title, "raw_content": entry.RawContent, "full_content": entry.FullContent}}

	if _, err := coll.UpdateOne(context.TODO(), filter, update); err != nil {
		common.Logger.Error("unable to update entry", zap.String("url", entry.ID.Hex()), zap.Error(err))
	}

	return nil
}

func (s *Storage) UpdateEntryDocID(entry *model.Entry) error {
	coll := getEntryMongodbColl(s)
	filter := bson.M{"_id": entry.ID}
	update := bson.M{"$set": bson.M{"doc_id": entry.DocId}}

	if _, err := coll.UpdateOne(context.TODO(), filter, update); err != nil {
		common.Logger.Error("unable to update entry", zap.String("url", entry.ID.Hex()), zap.Error(err))
	}

	return nil
}

func (s *Storage) UpdateEntryAlgorithms(entryID string, algorithms []string) error {
	coll := getEntryMongodbColl(s)
	id, _ := primitive.ObjectIDFromHex(entryID)
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"algorithms": algorithms}}

	if _, err := coll.UpdateOne(context.TODO(), filter, update); err != nil {
		common.Logger.Error("unable to update entry algorithms", zap.String("entryID", entryID), zap.Error(err))
	}

	return nil
}

func (s *Storage) createEntry(entry *model.Entry) (string, error) {
	coll := getEntryMongodbColl(s)
	entry.ID = primitive.NewObjectID()
	if _, err := coll.InsertOne(context.TODO(), entry); err != nil {
		common.Logger.Error("store: store: unable to create entry", zap.String("url", entry.URL), zap.Error(err))
		return "", err
	}
	return entry.ID.Hex(), nil
}

/*func (s *Storage) SaveFeedEntries(feedID, feedTitle, feedUrl string, entries model.Entries) {
	if len(entries) == 0 {
		return
	}
	var feedSearchRSSList []model.FeedNotification
	feedNotification := model.FeedNotification{
		FeedId:   feedID,
		FeedName: feedTitle,
		FeedIcon: "",
	}
	feedSearchRSSList = append(feedSearchRSSList, feedNotification)

	addList := make([]*model.EntryAddModel, 0)
	for _, entryModel := range entries {
		reqModel := model.GetEntryAddModel(entryModel, feedUrl)
		addList = append(addList, reqModel)
	}
	jsonByte, _ := json.Marshal(addList)
	url := common.EntryMonogoUpdateApiUrl()
	request, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonByte))
	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		common.Logger.Error("add entry in mongo  fail", zap.Error(err))
	}
	defer response.Body.Close()
	responseBody, _ := io.ReadAll(response.Body)
	var resObj model.MongoEntryApiResponseModel
	if err := json.Unmarshal(responseBody, &resObj); err != nil {
		log.Print("json decode failed, err", err)
		return
	}
	if resObj.Code == 0 {
		resEntryMap := make(map[string]string, 0)
		for _, resDataDetail := range resObj.Data {
			resEntryMap[resDataDetail.Url] = resDataDetail.ID
		}
		for _, entryModel := range entries {
			entryID, ok := resEntryMap[entryModel.URL]
			if ok {
				notificationData := model.NotificationData{
					Name:      entryModel.Title,
					EntryId:   entryID,
					Created:   entryModel.PublishedAt.Unix(),
					FeedInfos: feedSearchRSSList,
					Content:   entryModel.FullContent,
				}
				docId := search.InputRSS(&notificationData)
				entryObjID, _ := primitive.ObjectIDFromHex(entryID)
				updateDocIDEntry := &model.Entry{ID: entryObjID, DocId: docId}
				s.UpdateEntryDocID(updateDocIDEntry)
			}
		}
	}

}*/
