package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"bytetrade.io/web3os/system_workflow/api"
	"bytetrade.io/web3os/system_workflow/common"
	"bytetrade.io/web3os/system_workflow/model"
	"bytetrade.io/web3os/system_workflow/protobuf_entity"
	"bytetrade.io/web3os/system_workflow/storge"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

func addFeedInMongo(sources []string, feedMap map[string]*protobuf_entity.Feed) {
	addList := make([]*model.FeedAddModel, 0)
	for _, source := range sources {
		for _, currentFeed := range feedMap {
			reqModel := model.GetFeedAddModel(currentFeed)
			addList = append(addList, reqModel)
			if len(addList) >= 100 {
				api.AddFeedInMongo(source, addList)
				addList = make([]*model.FeedAddModel, 0)
				//time.Sleep(time.Second * 1)
			}
		}
		api.AddFeedInMongo(source, addList)
	}
}

func delFeedInMongo(sources []string, feedMap map[string]*protobuf_entity.Feed) {
	for _, source := range sources {
		delList := make([]string, 0)
		for feedUrl := range feedMap {
			delList = append(delList, feedUrl)
		}
		api.DelFeedInMongo(source, delList)
	}
}

func syncFeedDownloadPackage(packageUrl string, whetherAll bool) (*protobuf_entity.FeedAllPackage, *protobuf_entity.FeedIncremntPackage) {

	//feedRes, err := http.Get(packageUrl)
	client := &http.Client{Timeout: time.Second * 5}
	feedRes, err := client.Get(packageUrl)
	if err != nil {
		common.Logger.Error("get feed data  fail", zap.Error(err))
	}
	defer feedRes.Body.Close()

	body, err := io.ReadAll(feedRes.Body)
	if err != nil {
		common.Logger.Error("feed fail to get response", zap.Error(err))
	}
	uncompressByte := common.DoZlibUnCompress(body)
	var allPackageData protobuf_entity.FeedAllPackage
	var increasePackageData protobuf_entity.FeedIncremntPackage
	if whetherAll {
		unmarshalErr := proto.Unmarshal(uncompressByte, &allPackageData)
		if unmarshalErr != nil {
			common.Logger.Error("unmarshal all feed object  error", zap.Error(unmarshalErr))
		}
	} else {
		unmarshalErr := proto.Unmarshal(uncompressByte, &increasePackageData)
		if unmarshalErr != nil {
			common.Logger.Error("unmarshal increase feed object  error", zap.Error(unmarshalErr))
		}
	}
	return &allPackageData, &increasePackageData

}

func syncFeedGetPackage(feedUrl string, whetherAll bool) ([]*protobuf_entity.FeedAllPackage, []*protobuf_entity.FeedIncremntPackage, int64) {
	common.Logger.Info("sync feed:", zap.String("url", feedUrl))
	var allPackagePackTime int64
	var allPackageData []*protobuf_entity.FeedAllPackage
	var increasePackageData []*protobuf_entity.FeedIncremntPackage
	client := &http.Client{Timeout: time.Second * 5}
	//res, err := http.Get(feedUrl)
	res, err := client.Get(feedUrl)
	if err != nil {
		common.Logger.Error("get feed data  fail", zap.Error(err))
		return allPackageData, increasePackageData, allPackagePackTime
	}
	if res.StatusCode != 200 {
		common.Logger.Error("get feed data fail code")
		return allPackageData, increasePackageData, allPackagePackTime
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		common.Logger.Error("read feed data  fail", zap.Error(err))
	}
	if whetherAll {
		var feedPackages model.FeedPackageAllInfos
		errJson := json.Unmarshal(body, &feedPackages)
		if errJson != nil {
			common.Logger.Error("get feed data  fail", zap.Error(errJson))
		}
		for _, currentPackage := range feedPackages {
			allPackage, _ := syncFeedDownloadPackage(currentPackage.Url, whetherAll)
			allPackageData = append(allPackageData, allPackage)
			if allPackagePackTime < currentPackage.PackageTime {
				allPackagePackTime = currentPackage.PackageTime
			}
		}
	} else {
		var feedPackages model.FeedPackageIncrementInfos
		errJson := json.Unmarshal(body, &feedPackages)
		if errJson != nil {
			common.Logger.Error("get feed data  fail", zap.Error(errJson))
		}
		for _, currentPackage := range feedPackages {
			if currentPackage.FeedOperationSize > 0 || currentPackage.FeedNameOperationSize > 0 {
				_, increasePackage := syncFeedDownloadPackage(currentPackage.Url, whetherAll)
				increasePackageData = append(increasePackageData, increasePackage)
			}
		}
	}

	return allPackageData, increasePackageData, allPackagePackTime

}

func syncFeed(postgresClient *sql.DB, redisClient *redis.Client, provider *model.SyncProvider) {
	syncStartTime := time.Now()
	saveData, _ := storge.GetFeedSync(redisClient, provider.Provider, provider.FeedName)
	if saveData == nil {
		packageFeeds := make(map[string]*protobuf_entity.Feed, 0)
		allPackageDataList, _, packageTime := syncFeedGetPackage(fmt.Sprintf("%s&package_type=all", provider.FeedUrl), true)
		_, increasePackageDataList, _ := syncFeedGetPackage(fmt.Sprintf("%s&package_type=increment&start=%d", provider.FeedUrl, packageTime), false)
		for _, allPackage := range allPackageDataList {
			for _, feed := range allPackage.Feeds {
				packageFeeds[feed.FeedUrl] = feed
			}
		}
		for _, increasePackage := range increasePackageDataList {
			for _, operation := range increasePackage.FeedNameOperations {
				var updateFeed protobuf_entity.Feed
				errJson := json.Unmarshal([]byte(operation.Data), &updateFeed)
				if errJson != nil {
					common.Logger.Error("unmarshal increase feed name operation data  fail", zap.String("data", operation.Data), zap.Error(errJson))
					continue
				}
				if operation.Action == "add" {
					packageFeeds[updateFeed.FeedUrl] = &updateFeed
				}
				if operation.Action == "delete" {
					delete(packageFeeds, updateFeed.FeedUrl)
				}

			}
			for _, operation := range increasePackage.FeedOperations {
				var updateFeed protobuf_entity.Feed
				errJson := json.Unmarshal([]byte(operation.Data), &updateFeed)
				if errJson != nil {
					common.Logger.Error("unmarshal increase feed name operation data  fail", zap.String("data", operation.Data), zap.Error(errJson))
					continue
				}
				feed, feedOK := packageFeeds[updateFeed.FeedUrl]
				if operation.Action == "update" && feedOK {
					packageFeeds[updateFeed.FeedUrl] = model.GetUpdateProtoFeed(feed, &updateFeed)
				}
			}
		}
		addFeedInMongo(provider.Source, packageFeeds)
	} else {
		_, increasePackageDataList, _ := syncFeedGetPackage(fmt.Sprintf("%s&package_type=increment&start=%d", provider.FeedUrl, saveData.SyncStartTimestamp), false)

		for _, increasePackage := range increasePackageDataList {
			addPackageFeeds := make(map[string]*protobuf_entity.Feed, 0)
			deletePackageFeeds := make(map[string]*protobuf_entity.Feed, 0)
			for _, operation := range increasePackage.FeedNameOperations {
				var updateFeed protobuf_entity.Feed
				errJson := json.Unmarshal([]byte(operation.Data), &updateFeed)
				if errJson != nil {
					common.Logger.Error("unmarshal increase feed name operation data  fail", zap.String("data", operation.Data), zap.Error(errJson))
					continue
				}
				if operation.Action == "new" {
					common.Logger.Info("new feed in sync", zap.String("feed url:", updateFeed.FeedUrl))
					_, delExist := deletePackageFeeds[updateFeed.FeedUrl]
					if delExist {
						delete(deletePackageFeeds, updateFeed.FeedUrl)
					}
					addPackageFeeds[updateFeed.FeedUrl] = &updateFeed
				}
				if operation.Action == "remove" {
					common.Logger.Info("remove feed in sync", zap.String("feed url:", updateFeed.FeedUrl))
					_, addExist := addPackageFeeds[updateFeed.FeedUrl]
					if addExist {
						delete(addPackageFeeds, updateFeed.FeedUrl)
					}
					deletePackageFeeds[updateFeed.FeedUrl] = &updateFeed
				}
			}
			addFeedInMongo(provider.Source, addPackageFeeds)
			delFeedInMongo(provider.Source, deletePackageFeeds)
			updateFeedList := make(map[string]map[string]interface{}, 0)
			for _, operation := range increasePackage.FeedOperations {
				var curUpdateFeed map[string]interface{}
				errJson := json.Unmarshal([]byte(operation.Data), &curUpdateFeed)
				if errJson != nil {
					common.Logger.Error("unmarshal increase feed update operation data  fail", zap.String("data", operation.Data), zap.Error(errJson))
					continue
				}
				feedUrl, ok := curUpdateFeed["feed_url"]
				if ok {
					updateFeed, isFeedExist := updateFeedList[fmt.Sprintf("%v", feedUrl)]
					if isFeedExist {
						for key := range curUpdateFeed {
							updateFeed[key] = curUpdateFeed[key]
						}
					} else {
						updateFeedList[fmt.Sprintf("%v", feedUrl)] = curUpdateFeed
					}
				}
			}
			storge.UpdateFeed(postgresClient, provider.Source, updateFeedList)
		}
	}
	var redisSaveData model.FeedSyncData
	redisSaveData.SyncEndTimestamp = time.Now().UTC().Unix()
	redisSaveData.SyncStartTimestamp = syncStartTime.UTC().Unix()
	storge.SaveFeedSync(redisClient, provider.Provider, provider.FeedName, redisSaveData)
}

func fileToSave(path string, fileBytes []byte) {
	tempFile, createTempFileErr := os.Create(path)
	if createTempFileErr != nil {
		common.Logger.Error("create temp file err", zap.String("currentFeedFilePath", path), zap.Error(createTempFileErr))
		return
	}
	writer := bufio.NewWriter(tempFile)
	_, writeErr := writer.Write(fileBytes)
	if writeErr != nil {
		common.Logger.Error("write file error", zap.Error(writeErr))
		return
	}
	syncErr := writer.Flush()
	if syncErr != nil {
		common.Logger.Error("sync file error", zap.Error(syncErr))
		return
	}
}

func syncEntryDownloadPackage(provider string, newPackage *model.EntryPackage) {
	startTime := time.Unix(newPackage.StartTime, 0)
	dayStart := common.GetSpecificDayOneDayStart(startTime).Unix()
	timeStr := strconv.FormatInt(dayStart, 10)

	path := filepath.Join(common.SyncEntryDirectory(provider, newPackage.FeedName, newPackage.ModelName), timeStr) // newPackage.Language, timeStr)
	common.CreateNotExistDirectory(path, newPackage.ModelName+"_"+timeStr)

	//entryRes, err := http.Get(newPackage.URL)
	client := &http.Client{Timeout: time.Second * 5}
	entryRes, err := client.Get(newPackage.URL)
	if err != nil {
		common.Logger.Error("get entry data  fail", zap.Error(err))
	}
	defer entryRes.Body.Close()

	body, err := io.ReadAll(entryRes.Body)
	if err != nil {
		common.Logger.Error("feed fail to get response", zap.Error(err))
	}

	fileName := fmt.Sprintf("%d.zlib", newPackage.StartTime)
	fileToSave(filepath.Join(path, fileName), body)

}

func syncEntry(redisClient *redis.Client, provider *model.SyncProvider, lastSyncTime int64) {
	if lastSyncTime == 0 {
		currentUtcTime := time.Now().UTC()
		checkUtcTime := currentUtcTime.AddDate(0, 0, -int(provider.EntrySyncDate))
		lastSyncTime = int64(checkUtcTime.Unix())
	} else {
		lastSyncTime = lastSyncTime - 6*60*60
	}

	url := provider.EntryUrl + "&start=" + strconv.FormatInt(lastSyncTime, 10)
	common.Logger.Info("sync entry:", zap.String("url:", url))
	client := &http.Client{Timeout: time.Second * 5}
	res, err := client.Get(url)
	//res, err := http.Get(url)
	if err != nil {
		common.Logger.Error("get entry data  fail", zap.Error(err))
	}
	if res.StatusCode != 200 {
		common.Logger.Error("get entry data fail code")
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	var entryPackages model.EntryPackages
	errJson := json.Unmarshal(body, &entryPackages)
	if errJson != nil {
		common.Logger.Error("get entry data  fail", zap.Error(errJson))
	}
	for _, currentEntryPackage := range entryPackages {
		saveData, _ := storge.GetEntrySyncPackageData(redisClient, provider.Provider, currentEntryPackage.FeedName, currentEntryPackage.ModelName, currentEntryPackage.StartTime)
		if saveData == nil || saveData.Md5 != currentEntryPackage.MD5 {
			syncEntryDownloadPackage(provider.Provider, currentEntryPackage)
			var saveData model.EntrySyncPackageData
			saveData.Md5 = currentEntryPackage.MD5
			saveData.Language = currentEntryPackage.Language
			saveData.StartTime = currentEntryPackage.StartTime
			saveData.FeedName = currentEntryPackage.FeedName
			saveData.ModelName = currentEntryPackage.ModelName
			saveData.UpdateTime = int64(time.Now().UTC().Unix())
			storge.SaveEntrySyncPackageData(redisClient, common.GetEntrySyncPackageDataRedisKey(), saveData)
		}

	}

}

func checkExistAlgorithmInFirstRun(resp model.RecommendServiceResponseModel) (bool, string) {
	for _, argo := range resp.Data {
		source := argo.Metadata.Name
		lastExtractorTimeStr, _ := api.GetRedisConfig(source, "last_extractor_time").(string)
		if lastExtractorTimeStr == "" {
			return true, source
		}
	}
	return false, ""

}
func main() {
	common.Logger.Info("package sync  start...")

	startTimestamp := int64(time.Now().UTC().Unix())

	providerList := make(map[string]*model.SyncProvider, 0)
	url := "http://app-service.os-system:6755/app-service/v1/recommenddev/" + common.GetTermiusUserName() + "/status"
	client := &http.Client{Timeout: time.Second * 5}
	//res, err := http.Get(url)
	res, err := client.Get(url)
	if err != nil {
		common.Logger.Error("get recommend service error", zap.String("url", url), zap.Error(err))
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	jsonStr := string(body)
	common.Logger.Info("get recommend service response: ", zap.String("url", url), zap.String("body", jsonStr))

	var response model.RecommendServiceResponseModel
	if err := json.Unmarshal(body, &response); err != nil {
		common.Logger.Error("json decode failed ", zap.String("url", url), zap.Error(err))
		return
	}
	//inFirstRun, runSource := checkExistAlgorithmInFirstRun(response)
	for _, argo := range response.Data {
		source := argo.Metadata.Name
		/*lastExtractorTimeStr, _ := api.GetRedisConfig(source, "last_extractor_time").(string)
		if lastExtractorTimeStr == "" && inFirstRun && source != runSource {
			common.Logger.Info("source not sync because exist algorithm in first run : ", zap.String("run source:", runSource), zap.String("skip source:", source))
			continue
		}*/
		for _, provider := range argo.SyncProvider {
			key := provider.Provider + provider.FeedName
			p, exist := providerList[key]
			if exist {
				if !common.IsInStringArray(p.Source, source) {
					p.Source = append(p.Source, source)
				}
				if p.EntrySyncDate < provider.EntryProvider.SyncDate {
					p.EntrySyncDate = provider.EntryProvider.SyncDate
				}
			} else {
				var providerSetting model.SyncProvider
				sourceArr := make([]string, 0)
				sourceArr = append(sourceArr, source)
				providerSetting.Source = sourceArr
				providerSetting.FeedName = provider.FeedName
				providerSetting.Provider = provider.Provider
				providerSetting.FeedUrl = provider.FeedProvider.Url
				providerSetting.EntrySyncDate = provider.EntryProvider.SyncDate
				providerSetting.EntryUrl = provider.EntryProvider.Url
				providerList[key] = &providerSetting
			}
		}
	}

	redisClient := common.GetRDBClient()
	defer redisClient.Close()
	postgresClient := common.NewPostgresClient()
	defer postgresClient.Close()
	for key, provider := range providerList {
		lastSyncTimeStr, _ := api.GetRedisConfig(key, "last_sync_time").(string)
		lastSyncTime, _ := strconv.ParseInt(lastSyncTimeStr, 10, 64)
		common.Logger.Info("sync  start", zap.String("last sync time str", lastSyncTimeStr), zap.Int64("last sync time", lastSyncTime), zap.Int64("now time", startTimestamp))
		if lastSyncTimeStr == "" || startTimestamp > lastSyncTime+10*60 {

			syncFeed(postgresClient, redisClient, provider)
			syncEntry(redisClient, provider, lastSyncTime)
			api.SetRedisConfig(key, "last_sync_time", startTimestamp)
		}

	}

	common.Logger.Info("package sync  end")
}
