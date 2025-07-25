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

func SaveFeedEntries(bflUser string, store *storage.Storage, entries model.Entries, feed *model.Feed) {
	common.Logger.Info("add entry in knowledge", zap.Int("len", len(entries)))
	if len(entries) == 0 {
		return
	}

	addList := make([]*model.EntryAddModel, 0)
	for _, entryModel := range entries {
		reqModel := model.GetEntryAddModel(entryModel, feed.FeedURL)
		addList = append(addList, reqModel)
	}
	doReq(bflUser, addList, entries, feed, store, true)

}

func DownloadDoReq(download model.EntryDownloadModel) {
	downloadUrl := common.DownloadApiUrl() + "/download/start" // "/termius/download"
	algoJsonByte, err := json.Marshal(download)
	if err != nil {
		common.Logger.Error("add download json marshal  fail", zap.Error(err))
	}

	common.Logger.Info("start download ", zap.String("api", downloadUrl), zap.String("url", download.DataSource), zap.String("file_type", download.FileType))
	algoReq, _ := http.NewRequest("POST", downloadUrl, bytes.NewBuffer(algoJsonByte))
	algoReq.Header.Set("Content-Type", "application/json")
	algoClient := &http.Client{Timeout: time.Second * 5}
	_, err = algoClient.Do(algoReq)
	if err != nil {
		common.Logger.Error("new download   fail", zap.Error(err))
		return
	}
	if algoReq != nil {
		defer algoReq.Body.Close()
	}
	body, _ := io.ReadAll(algoReq.Body)
	jsonStr := string(body)
	common.Logger.Info("new download response: ", zap.String("download url", download.DataSource), zap.String("body", jsonStr))
}
func NewEnclosure(entry *model.Entry, feed *model.Feed, store *storage.Storage) {
	exist := store.GetEnclosureNumByEntry(entry.ID)
	if exist > 0 {
		common.Logger.Info("new enclosure exit where entry's enclosure exist ", zap.String("entry id:", entry.ID))
		return
	}
	enclosureID, _ := store.CreateEnclosure(entry)
	if entry.MediaUrl != "" {
		var download model.EntryDownloadModel
		download.DataSource = entry.MediaUrl
		//download.TaskUser = common.CurrentUser()
		download.DownloadAPP = "wise"
		download.EnclosureId = enclosureID
		download.FileName = entry.Title
		download.FileType = entry.MediaType
		download.Path = "Downloads/Wise/Article"
		download.BflUser = entry.BflUser
		if feed != nil {
			download.Path = "Downloads/Wise/Feed/" + feed.Title
		}

		if feed == nil || feed.AutoDownload {
			DownloadDoReq(download)
		}
	} else {
		common.Logger.Error("entry mediaUrl is null")
	}
}

func doReq(bflUser string, list []*model.EntryAddModel, entries model.Entries, feed *model.Feed, store *storage.Storage, isNew bool) {
	jsonByte, _ := json.Marshal(list)
	url := common.EntryMonogoUpdateApiUrl()
	request, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonByte))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Bfl-User", bflUser)
	client := &http.Client{Timeout: 20 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		common.Logger.Error("add entry in knowledg  fail", zap.Error(err))
		return
	}
	if response != nil {
		defer response.Body.Close()
	}
	responseBody, _ := io.ReadAll(response.Body)
	var resObj model.MongoEntryApiResponseModel
	if err := json.Unmarshal(responseBody, &resObj); err != nil {
		log.Print("json decode failed, err", err)
		return
	}
	if resObj.Code == 0 {
		resEntryMap := make(map[string]string, 0)
		if isNew {
			for _, resDataDetail := range resObj.Data {
				resEntryMap[resDataDetail.Url] = resDataDetail.ID
			}
			for _, entry := range entries {
				entryID, ok := resEntryMap[entry.URL]
				if ok {
					entry.ID = entryID
					if entry.MediaContent != "" || entry.MediaUrl != "" {
						NewEnclosure(entry, feed, store)
					}
				}
			}
		}

	}

}

func UpdateFeedEntries(bflUser string, store *storage.Storage, entries model.Entries, feed *model.Feed) {
	common.Logger.Info("update entry in knowledge", zap.Int("len", len(entries)))
	if len(entries) == 0 {
		return
	}

	addList := make([]*model.EntryAddModel, 0)
	for _, entryModel := range entries {
		reqModel := model.GetEntryUpdateSourceModel(entryModel, feed.FeedURL)
		addList = append(addList, reqModel)
	}
	doReq(bflUser, addList, entries, feed, store, false)
}

func UpdateLibraryEntryContent(bflUser string, entry *model.Entry, isVideo bool) {
	updateList := make([]*model.EntryAddModel, 0)
	var updateEntry model.EntryAddModel
	updateEntry.Url = entry.URL
	updateEntry.Title = entry.Title
	updateEntry.ImageUrl = entry.ImageUrl
	updateEntry.PublishedAt = entry.PublishedAt
	updateEntry.Author = entry.Author
	updateEntry.Language = entry.Language
	updateEntry.RawContent = entry.RawContent
	updateEntry.FullContent = entry.FullContent
	if isVideo == false {
		updateEntry.Crawler = true
		updateEntry.Extract = true
		updateEntry.Attachment = entry.Attachment
	}
	updateList = append(updateList, &updateEntry)
	jsonByte, _ := json.Marshal(updateList)
	url := common.EntryMonogoUpdateApiUrl()
	request, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonByte))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Bfl-User", bflUser)
	client := &http.Client{Timeout: 5 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		common.Logger.Error("add entry in knowledg  fail", zap.Error(err))
		return
	}
	if response != nil {
		defer response.Body.Close()
	}
	responseBody, _ := io.ReadAll(response.Body)
	jsonStr := string(responseBody)
	common.Logger.Info("update content response: ", zap.String("body", jsonStr))

}

func LoadMetaFromYtdlp(bflUser, entryUrl string) *model.Entry {
	url := common.YTDLPApiUrl() + "/v1/get_metadata?url=" + entryUrl + "&bfl_user=" + bflUser
	common.Logger.Info("load meta from ytdlp", zap.String("url", url))
	client := &http.Client{Timeout: time.Second * 50}
	res, err := client.Get(url)
	if err != nil {
		common.Logger.Error("load ytdlp meta error", zap.Error(err))
		return nil
	}
	if res.StatusCode != 200 {
		common.Logger.Error("load ytdlp meta error")
		return nil
	}
	if res != nil {
		defer res.Body.Close()
	}
	body, _ := io.ReadAll(res.Body)

	var resObj model.EntryFetchResponseModel
	if err := json.Unmarshal(body, &resObj); err != nil {
		common.Logger.Error("load ytdlp meta ,json decode failed, err", zap.Error(err))
		return nil
	}

	return &resObj.Data

}

func FetchTwitterContent(bfl_user, twitterID, url string) *model.Entry {
	apiUrl := common.DownloadApiUrl() + "/twitter/fetch-content?twitter_id=" + twitterID + "&url=" + url + "&bfl_user=" + bfl_user
	client := &http.Client{Timeout: time.Second * 120}
	res, err := client.Get(apiUrl)
	if err != nil {
		common.Logger.Error("fetch twitter content error", zap.String("id", twitterID), zap.String("url", url), zap.Error(err))
		return nil
	}
	if res.StatusCode != 200 {
		common.Logger.Error("fetch twitter content error", zap.String("id", twitterID), zap.String("url", url))
		return nil
	}
	if res != nil {
		defer res.Body.Close()
	}
	body, _ := io.ReadAll(res.Body)

	var resObj model.EntryFetchResponseModel
	if err := json.Unmarshal(body, &resObj); err != nil {
		common.Logger.Error("fetch twitter content ,json decode failed, err", zap.Error(err))
		return nil
	}

	return &resObj.Data

}

type XHSReq struct {
	Url     string `json:"url"`
	BflUser string `json:"bfl_user"`
}

func FetchXHSContent(url string, bfl_user string) *model.Entry {
	//apiUrl := common.DownloadApiUrl() + "/xhs/fetch-content?url=" + url + "&bfl_user=" + bfl_user
	//client := &http.Client{Timeout: time.Second * 120}
	//res, err := client.Post(apiUrl)
	req := XHSReq{Url: url, BflUser: bfl_user}
	jsonByte, _ := json.Marshal(req)
	apiUrl := common.DownloadApiUrl() + "/xhs/fetch-content"
	request, _ := http.NewRequest("POST", apiUrl, bytes.NewBuffer(jsonByte))
	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 20 * time.Second}
	res, err := client.Do(request)

	if err != nil {
		common.Logger.Error("fetch xhs content error", zap.String("url", url), zap.Error(err))
		return nil
	}
	if res.StatusCode != 200 {
		common.Logger.Error("fetch xhs content error", zap.String("url", url))
		return nil
	}
	if res != nil {
		defer res.Body.Close()
	}
	body, _ := io.ReadAll(res.Body)

	var resObj model.EntryFetchResponseModel
	if err := json.Unmarshal(body, &resObj); err != nil {
		common.Logger.Error("fetch xhs content ,json decode failed, err", zap.Error(err))
		return nil
	}

	return &resObj.Data

}
