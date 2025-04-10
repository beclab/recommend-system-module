package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"bytetrade.io/web3os/system_workflow/common"
	"bytetrade.io/web3os/system_workflow/model"

	"go.uber.org/zap"
)

func AddFeedInKnowledge(bfl_user, source string, list []*model.FeedAddModel) {
	if len(list) > 0 {
		for _, reqModel := range list {
			reqModel.Source = source
		}
		url := common.FeedMonogoApiUrl() + source
		jsonByte, _ := json.Marshal(list)
		request, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonByte))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("X-Bfl-User", bfl_user)
		client := &http.Client{Timeout: 5 * time.Second}
		response, err := client.Do(request)
		if err != nil {
			common.Logger.Error("add feed in mongo  fail", zap.Error(err))
			return
		}
		defer response.Body.Close()
		responseBody, _ := io.ReadAll(response.Body)
		common.Logger.Info("add feed in mongo", zap.String("url", url), zap.String("content", string(responseBody)))
	}

}
func DelFeedInKnowledge(bfl_user, source string, list []string) {
	if len(list) > 0 {
		common.Logger.Info("del feed in mongo", zap.Int("list size", len(list)))
		reqData := model.MongoFeedDelModel{FeedUrls: list}
		jsonByte, _ := json.Marshal(reqData)

		url := common.FeedMonogoApiUrl() + source
		request, _ := http.NewRequest("DELETE", url, bytes.NewBuffer(jsonByte))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("X-Bfl-User", bfl_user)
		client := &http.Client{Timeout: 5 * time.Second}
		response, err := client.Do(request)
		if err != nil {
			common.Logger.Error("del feed in mongo  fail", zap.Error(err))
			return
		}
		defer response.Body.Close()
		responseBody, _ := io.ReadAll(response.Body)
		common.Logger.Info("del feed in mongo", zap.String("content", string(responseBody)))
	}
}

func getAllEntries(bfl_user, reqParam string) *model.EntryApiDataResponseModel {
	url := common.EntryMonogoEntryApiUrl() + "?" + reqParam
	request, _ := http.NewRequest("GET", url, nil)
	request.Header.Set("X-Bfl-User", bfl_user)
	client := &http.Client{Timeout: time.Second * 5}
	//res, err := http.Get(url)
	res, err := client.Do(request)
	if err != nil {
		common.Logger.Error("get entry data  fail", zap.Error(err))
		return nil
	}
	if res.StatusCode != 200 {
		common.Logger.Error("get entry data fail code")
		return nil
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	var resObj model.EntryApiResponseModel
	if err := json.Unmarshal(body, &resObj); err != nil {
		log.Print("json decode failed, err", err)
		return nil
	}
	return &resObj.Data
}

func GetUncrawleredList(bfl_user string, offset, limit int, source string) (int, []model.EntryCrawlerModel) {
	param := "offset=" + fmt.Sprintf("%d", offset) + "&limit=" + fmt.Sprintf("%d", limit) + "&crawler=false&source=" + source
	queryData := getAllEntries(bfl_user, param)
	crawlerList := make([]model.EntryCrawlerModel, 0)
	for _, entry := range queryData.Items {
		var crawlerEntry model.EntryCrawlerModel
		crawlerEntry.Url = entry.Url
		crawlerList = append(crawlerList, crawlerEntry)
	}
	return queryData.Count, crawlerList
}

func UpdateEntriesInMongo(addList []*model.EntryAddModel) {

	if len(addList) > 0 {
		jsonByte, _ := json.Marshal(addList)
		url := common.EntryMonogoEntryApiUrl()
		request, newReqErr := http.NewRequest("POST", url, bytes.NewBuffer(jsonByte))
		if newReqErr != nil {
			log.Print("new http request fail url", url, newReqErr)
			return
		}
		request.Header.Set("Content-Type", "application/json")
		client := &http.Client{Timeout: 5 * time.Second}
		response, err := client.Do(request)
		if err != nil {
			log.Print("add entry in mongo  fail", err)
			return
		}
		defer response.Body.Close()
		responseBody, _ := io.ReadAll(response.Body)
		var resObj model.MongoEntryApiResponseModel
		if err := json.Unmarshal(responseBody, &resObj); err != nil {
			log.Print("json decode failed, err", err)
			return
		}
		if resObj.Code != 0 {
			common.Logger.Info("update feed in mongo code err", zap.Int("result code", resObj.Code))
		}
		common.Logger.Info("update entries in mongo finish all...", zap.Int("entry size:", len(addList)))
	}
}

func GetRedisConfig(bfl_user, provider, key string) interface{} {
	url := common.RedisConfigApiUrl() + provider + "/" + key
	request, _ := http.NewRequest("GET", url, nil)
	request.Header.Set("X-Bfl-User", bfl_user)
	client := &http.Client{Timeout: time.Second * 5}
	res, err := client.Do(request)
	if err != nil {
		common.Logger.Error("get redis config  fail", zap.Error(err))
		return ""
	}
	if res.StatusCode != 200 {
		common.Logger.Error("get redis config fail code")
		return ""
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	var resObj model.RedisConfigResponseModel
	if err := json.Unmarshal(body, &resObj); err != nil {
		log.Print("json decode failed, err", err)
		return ""
	}
	return resObj.Data
}

func SetRedisConfig(bfl_user, provider, key string, val interface{}) {
	var c model.RedisConfig
	c.Value = val
	url := common.RedisConfigApiUrl() + provider + "/" + key
	common.Logger.Info("set redis config", zap.String("url", url))

	jsonByte, err := json.Marshal(c)
	if err != nil {
		common.Logger.Error("set redis configjson marshal  fail", zap.Error(err))
	}

	common.Logger.Info("set redis config  ", zap.String("url", url), zap.String("key", key), zap.Any("val", val))

	algoReq, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonByte))
	algoReq.Header.Set("Content-Type", "application/json")
	algoReq.Header.Set("X-Bfl-User", bfl_user)
	algoClient := &http.Client{Timeout: 5 * time.Second}
	_, err = algoClient.Do(algoReq)
	if err != nil {
		common.Logger.Error("set redis configjson req  fail", zap.Error(err))
		return
	}

	defer algoReq.Body.Close()
	body, _ := io.ReadAll(algoReq.Body)
	jsonStr := string(body)
	common.Logger.Info("update redis config response: ", zap.String("body", jsonStr))

}
