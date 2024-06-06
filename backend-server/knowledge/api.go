package knowledge

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"bytetrade.io/web3os/backend-server/common"
	"bytetrade.io/web3os/backend-server/model"
	"bytetrade.io/web3os/backend-server/service/search"
	"bytetrade.io/web3os/backend-server/storage"
	"go.uber.org/zap"
)

func SaveFeedEntries(store *storage.Storage, entries model.Entries, feed *model.Feed, feedSearchRSSList []model.FeedNotification) {
	common.Logger.Info("add entry in mongo", zap.Int("len", len(entries)))
	if len(entries) == 0 {
		return
	}

	addList := make([]*model.EntryAddModel, 0)
	for _, entryModel := range entries {
		reqModel := model.GetEntryAddModel(entryModel, feed.FeedURL)
		addList = append(addList, reqModel)
	}
	doReq(addList, entries, feedSearchRSSList, store)

}

func doReq(list []*model.EntryAddModel, entries model.Entries, feedSearchRSSList []model.FeedNotification, store *storage.Storage) {
	jsonByte, _ := json.Marshal(list)
	url := common.EntryMonogoUpdateApiUrl()
	request, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonByte))
	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 5 * time.Second}
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
		for _, entry := range entries {
			entryID, ok := resEntryMap[entry.URL]
			if ok {
				notificationData := model.NotificationData{
					Name:      entry.Title,
					EntryId:   entryID,
					Created:   entry.PublishedAt,
					FeedInfos: feedSearchRSSList,
					Content:   entry.FullContent,
				}
				docId := search.InputRSS(&notificationData)
				//entryObjID, _ := primitive.ObjectIDFromHex(entryID)
				//updateDocIDEntry := &model.Entry{ID: entryObjID, DocId: docId}
				updateDocIDEntry := &model.Entry{ID: entryID, DocId: docId}
				store.UpdateEntryDocID(updateDocIDEntry)
			}
		}
	}

}

func UpdateFeedEntries(store *storage.Storage, entries model.Entries, feed *model.Feed, feedSearchRSSList []model.FeedNotification) {
	common.Logger.Info("update entry in mongo", zap.Int("len", len(entries)))
	if len(entries) == 0 {
		return
	}

	addList := make([]*model.EntryAddModel, 0)
	for _, entryModel := range entries {
		reqModel := model.GetEntryUpdateSourceModel(entryModel, feed.FeedURL)
		addList = append(addList, reqModel)
	}
	doReq(addList, entries, feedSearchRSSList, store)

}
