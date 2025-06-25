package service

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"bytetrade.io/web3os/backend-server/common"
	"bytetrade.io/web3os/backend-server/model"
	"bytetrade.io/web3os/backend-server/storage"
	"go.uber.org/zap"
)

type YoutubeResponseItem struct {
	Total     string `json:"id"`
	Title     string `json:"title"`
	URL       string `json:"url"`
	Timestamp int64  `json:"timestamp"`
}

type YoutubeListResponseData struct {
	//Total  int                   `json:"total"`
	Title  string                `json:"title"`
	Author string                `json:"author"`
	Avatar string                `json:"avatar"`
	List   []YoutubeResponseItem `json:"list"`
}

type YoutubeListResponse struct {
	Code int                     `json:"code"`
	Data YoutubeListResponseData `json:"data"`
}

func GetEntryFromYoutubeEntry(youtubeEntry YoutubeResponseItem, author string) *model.Entry {
	var entry model.Entry

	entry.Title = youtubeEntry.Title
	entry.URL = youtubeEntry.URL
	entry.PublishedAt = youtubeEntry.Timestamp
	entry.Author = author
	return &entry
}

func youtubeFeedRefreshExec(url string, start int, end int) YoutubeListResponse {
	youtubeListUrl := common.YTDLPApiUrl() + "/v1/get_youtube_entry_list?" + fmt.Sprintf("url=%s&start=%d&end=%d", url, start, end)
	client := &http.Client{Timeout: time.Second * 60}
	var responseData YoutubeListResponse
	res, err := client.Get(youtubeListUrl)
	if err != nil {
		common.Logger.Error("youtube feed refresh error", zap.Error(err))
		return responseData
	}
	if res != nil {
		defer res.Body.Close()
	}
	body, _ := io.ReadAll(res.Body)

	if err := json.Unmarshal(body, &responseData); err != nil {
		log.Print("json decode failed, err", err)
	}
	if responseData.Code != 0 {
		log.Print("youtube feed code err")
	}
	return responseData

}
func RefreshYoutubeFeed(store *storage.Storage, url string, feedID string) (*model.Feed, string) {
	var feed model.Feed
	avatar := ""
	start := 0
	limit := 10
	if store.GetEntryNumByFeed(feedID) == 0 {
		limit = 30
	}
	entries := make([]*model.Entry, 0)
	responseData := youtubeFeedRefreshExec(url, start, start+limit)
	author := responseData.Data.Author
	for len(responseData.Data.List) > 0 {
		entrySize := len(responseData.Data.List)
		for _, respEntry := range responseData.Data.List {
			entries = append(entries, GetEntryFromYoutubeEntry(respEntry, author))
		}
		//if the oldest one is fetched,task is over
		//else fetch the next list
		lastEntry := responseData.Data.List[entrySize-1]
		savedEntry := store.GetEntryByUrl(feedID, lastEntry.URL)
		if savedEntry != nil {
			break
		} else if len(responseData.Data.List) == limit {
			start = start + limit
			responseData = youtubeFeedRefreshExec(url, start, start+limit)
		}
	}
	feed.Title = responseData.Data.Title
	//avatar = responseData.Data.Avatar
	feed.Entries = entries
	common.Logger.Info("youtebe feed refresh", zap.String("url", url), zap.Int("len", len(responseData.Data.List)))

	return &feed, avatar
}
