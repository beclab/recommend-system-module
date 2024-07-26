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
	"bytetrade.io/web3os/backend-server/storage"
	"go.uber.org/zap"
)

func SaveFeedEntries(store *storage.Storage, entries model.Entries, feed *model.Feed) {
	common.Logger.Info("add entry in mongo", zap.Int("len", len(entries)))
	if len(entries) == 0 {
		return
	}

	addList := make([]*model.EntryAddModel, 0)
	for _, entryModel := range entries {
		reqModel := model.GetEntryAddModel(entryModel, feed.FeedURL)
		addList = append(addList, reqModel)
	}
	doReq(addList, entries, store)

}

func doDownloadReq(download model.EntryDownloadModel) {
	downloadUrl := common.DownloadApiUrl()
	algoJsonByte, err := json.Marshal(download)
	if err != nil {
		common.Logger.Error("add download json marshal  fail", zap.Error(err))
	}

	common.Logger.Info("start download ", zap.String("url", download.DataSource))
	algoReq, _ := http.NewRequest("POST", downloadUrl, bytes.NewBuffer(algoJsonByte))
	algoReq.Header.Set("Content-Type", "application/json")
	algoClient := &http.Client{Timeout: time.Second * 5}
	_, err = algoClient.Do(algoReq)

	defer algoReq.Body.Close()
	body, _ := io.ReadAll(algoReq.Body)
	jsonStr := string(body)
	common.Logger.Info("new download response: ", zap.String("body", jsonStr))

	if err != nil {
		common.Logger.Error("new download   fail", zap.Error(err))
	}

	common.Logger.Info("update download finish ", zap.String("download url", download.DataSource))
}
func NewEnclosure(entry *model.Entry, store *storage.Storage) {
	enclosureID, _ := store.CreateEnclosure(entry)
	if entry.MediaUrl != "" {
		var download model.EntryDownloadModel
		download.DataSource = entry.MediaUrl
		download.TaskUser = common.CurrentUser()
		download.DownloadAPP = "wise"
		download.EnclosureId = enclosureID
		doDownloadReq(download)
	} else {
		common.Logger.Error("entry mediaUrl is null")
	}
}

func doReq(list []*model.EntryAddModel, entries model.Entries, store *storage.Storage) {
	jsonByte, _ := json.Marshal(list)
	url := common.EntryMonogoUpdateApiUrl()
	request, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonByte))
	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 5 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		common.Logger.Error("add entry in knowledg  fail", zap.Error(err))
		return
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
				entry.ID = entryID
				if entry.MediaContent != "" || entry.MediaUrl != "" {
					NewEnclosure(entry, store)
				}
			}
		}

	}

}

func UpdateFeedEntries(store *storage.Storage, entries model.Entries, feed *model.Feed) {
	common.Logger.Info("update entry in mongo", zap.Int("len", len(entries)))
	if len(entries) == 0 {
		return
	}

	addList := make([]*model.EntryAddModel, 0)
	for _, entryModel := range entries {
		reqModel := model.GetEntryUpdateSourceModel(entryModel, feed.FeedURL)
		addList = append(addList, reqModel)
	}
	doReq(addList, entries, store)
}

func UpdateLibraryEntryContent(entry *model.Entry) {
	updateList := make([]*model.EntryAddModel, 0)
	var updateEntry model.EntryAddModel
	updateEntry.Url = entry.URL
	updateEntry.Title = entry.Title
	updateEntry.PublishedAt = entry.PublishedAt
	updateEntry.Author = entry.Author
	updateEntry.Language = entry.Language
	updateEntry.RawContent = entry.RawContent
	updateEntry.FullContent = entry.FullContent
	updateEntry.Crawler = true
	updateEntry.Extract = true
	updateList = append(updateList, &updateEntry)
	jsonByte, _ := json.Marshal(updateList)
	url := common.EntryMonogoUpdateApiUrl()
	request, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonByte))
	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 5 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		common.Logger.Error("add entry in knowledg  fail", zap.Error(err))
		return
	}
	defer response.Body.Close()
	responseBody, _ := io.ReadAll(response.Body)
	jsonStr := string(responseBody)
	common.Logger.Info("update content response: ", zap.String("body", jsonStr))

}
