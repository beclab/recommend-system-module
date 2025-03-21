package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"bytetrade.io/web3os/system_workflow/api"
	"bytetrade.io/web3os/system_workflow/common"
	"bytetrade.io/web3os/system_workflow/model"
	"bytetrade.io/web3os/system_workflow/protobuf_entity"
	"bytetrade.io/web3os/system_workflow/storge"

	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

func addFeedInMongo(source string, feedMap map[string]*protobuf_entity.Feed) {
	addList := make([]*model.FeedAddModel, 0)
	for _, currentFeed := range feedMap {
		reqModel := model.GetFeedAddModel(currentFeed)
		addList = append(addList, reqModel)
		if len(addList) >= 100 {
			api.AddFeedInMongo(source, addList)
			addList = make([]*model.FeedAddModel, 0)
		}
	}
	api.AddFeedInMongo(source, addList)
}

func delFeedInMongo(source string, feedMap map[string]*protobuf_entity.Feed) {
	delList := make([]string, 0)
	for feedUrl := range feedMap {
		delList = append(delList, feedUrl)
	}
	api.DelFeedInMongo(source, delList)
}

func syncFeedDownloadPackage(packageUrl string, whetherAll bool) (*protobuf_entity.FeedAllPackage, *protobuf_entity.FeedIncremntPackage) {
	var allPackageData protobuf_entity.FeedAllPackage
	var increasePackageData protobuf_entity.FeedIncremntPackage

	client := &http.Client{Timeout: time.Second * 5}
	feedRes, err := client.Get(packageUrl)
	if err != nil {
		common.Logger.Error("get feed data  fail", zap.Error(err))
		return &allPackageData, &increasePackageData
	}
	defer feedRes.Body.Close()

	body, err := io.ReadAll(feedRes.Body)
	if err != nil {
		common.Logger.Error("feed fail to get response", zap.Error(err))
	}
	uncompressByte := common.DoZlibUnCompress(body)

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
			if allPackage != nil {
				allPackageData = append(allPackageData, allPackage)
			}
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
				if increasePackage != nil {
					increasePackageData = append(increasePackageData, increasePackage)
				}
			}
		}
	}

	return allPackageData, increasePackageData, allPackagePackTime

}

func handleFullSync(provider model.AlgoSyncProviderResponseModel, source string) {
	allPackageURL := fmt.Sprintf("%s&package_type=all", provider.FeedProvider.Url)
	allPackages, _, packageTime := syncFeedGetPackage(allPackageURL, true)

	incrementalURL := fmt.Sprintf("%s&package_type=increment&start=%d", provider.FeedProvider.Url, packageTime)
	_, incrementalPackages, _ := syncFeedGetPackage(incrementalURL, false)

	mergedFeeds := mergeAllAndIncrementalFeeds(allPackages, incrementalPackages)
	addFeedInMongo(source, mergedFeeds)
}

func mergeAllAndIncrementalFeeds(allPackages []*protobuf_entity.FeedAllPackage, incrementalPackages []*protobuf_entity.FeedIncremntPackage) map[string]*protobuf_entity.Feed {
	feeds := make(map[string]*protobuf_entity.Feed)
	for _, pkg := range allPackages {
		for _, feed := range pkg.Feeds {
			feeds[feed.FeedUrl] = feed
		}
	}
	for _, pkg := range incrementalPackages {
		processFeedNameOperations(pkg.FeedNameOperations, feeds)
		processFeedOperations(pkg.FeedOperations, feeds)
	}
	return feeds
}

func parseOperationData(data string, target interface{}) error {
	return json.Unmarshal([]byte(data), target)
}

func processFeedNameOperations(operations []*protobuf_entity.FeedNameOperation, feeds map[string]*protobuf_entity.Feed) {
	for _, op := range operations {
		var feed protobuf_entity.Feed
		if err := parseOperationData(op.Data, &feed); err != nil {
			common.Logger.Error("unmarshal increase feed name operation data  fail", zap.String("data", op.Data), zap.Error(err))
			continue
		}
		switch op.Action {
		case "add":
			feeds[feed.FeedUrl] = &feed
		case "delete":
			delete(feeds, feed.FeedUrl)
		}
	}
}

func processFeedOperations(operations []*protobuf_entity.FeedOperation, feeds map[string]*protobuf_entity.Feed) {
	for _, op := range operations {
		var feed protobuf_entity.Feed
		if err := parseOperationData(op.Data, &feed); err != nil {
			common.Logger.Error("unmarshal increase feed name operation data  fail", zap.String("data", op.Data), zap.Error(err))
			continue
		}

		if existingFeed, exists := feeds[feed.FeedUrl]; exists && op.Action == "update" {
			feeds[feed.FeedUrl] = model.GetUpdateProtoFeed(existingFeed, &feed)
		}
	}
}

func processIncrementalFeedOperations(operations []*protobuf_entity.FeedOperation) map[string]map[string]interface{} {
	updateFields := make(map[string]map[string]interface{})
	for _, operation := range operations {
		var curUpdateFeed map[string]interface{}
		errJson := json.Unmarshal([]byte(operation.Data), &curUpdateFeed)
		if errJson != nil {
			common.Logger.Error("unmarshal increase feed update operation data  fail", zap.String("data", operation.Data), zap.Error(errJson))
			continue
		}
		feedUrl, ok := curUpdateFeed["feed_url"]
		if ok {
			updateFeed, isFeedExist := updateFields[fmt.Sprintf("%v", feedUrl)]
			if isFeedExist {
				for key := range curUpdateFeed {
					updateFeed[key] = curUpdateFeed[key]
				}
			} else {
				updateFields[fmt.Sprintf("%v", feedUrl)] = curUpdateFeed
			}
		}
	}

	return updateFields
}

func handleIncrementalSync(postgresClient *sql.DB, provider model.AlgoSyncProviderResponseModel, source string, startTimestamp int64) {
	incrementalURL := fmt.Sprintf("%s&package_type=increment&start=%d", provider.FeedProvider.Url, startTimestamp)
	_, incrementalPackages, _ := syncFeedGetPackage(incrementalURL, false)

	for _, pkg := range incrementalPackages {
		addPackageFeeds := make(map[string]*protobuf_entity.Feed, 0)
		deletePackageFeeds := make(map[string]*protobuf_entity.Feed, 0)
		for _, operation := range pkg.FeedNameOperations {
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

		addFeedInMongo(source, addPackageFeeds)
		delFeedInMongo(source, deletePackageFeeds)
		updateFields := processIncrementalFeedOperations(pkg.FeedOperations)
		storge.UpdateFeed(postgresClient, source, updateFields)
	}
}

func syncFeed(postgresClient *sql.DB, redisClient *redis.Client, provider model.AlgoSyncProviderResponseModel, source string) {
	syncStartTime := time.Now()
	saveData, _ := storge.GetFeedSync(redisClient, provider.Provider, provider.FeedName, source)
	if saveData == nil {
		handleFullSync(provider, source)

	} else {
		handleIncrementalSync(postgresClient, provider, source, saveData.SyncStartTimestamp)
	}
	var redisSaveData model.FeedSyncData
	redisSaveData.SyncEndTimestamp = time.Now().UTC().Unix()
	redisSaveData.SyncStartTimestamp = syncStartTime.UTC().Unix()
	storge.SaveFeedSync(redisClient, provider.Provider, provider.FeedName, source, redisSaveData)
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

	client := &http.Client{Timeout: time.Second * 5}
	entryRes, err := client.Get(newPackage.URL)
	if err != nil {
		common.Logger.Error("get entry data  fail", zap.Error(err))
		return
	}
	defer entryRes.Body.Close()

	body, err := io.ReadAll(entryRes.Body)
	if err != nil {
		common.Logger.Error("feed fail to get response", zap.Error(err))
	}
	uncompressByte := common.DoZlibUnCompress(body)
	var allPackageData protobuf_entity.ListEntry
	transEntryList := make([]*protobuf_entity.EntryTrans, 0)
	unmarshalErr := proto.Unmarshal(uncompressByte, &allPackageData)
	if unmarshalErr != nil {
		common.Logger.Error("unmarshal all feed object  error", zap.Error(unmarshalErr))
		return
	}
	for _, entry := range allPackageData.Entries {
		entryTrans := model.GetProtoEntryTransModel(newPackage.ModelName, entry)
		transEntryList = append(transEntryList, entryTrans)

	}
	var transProtobuf protobuf_entity.ListEntryTrans
	transProtobuf.Entries = transEntryList

	currentProtoByte, marshalErr := proto.Marshal(&transProtobuf)
	if marshalErr != nil {
		common.Logger.Error("save to file marshal Err ", zap.Error(marshalErr))
		return
	}

	fileName := fmt.Sprintf("%d.zlib", newPackage.StartTime)
	fileToSave(filepath.Join(path, fileName), currentProtoByte)

}

func fetchEntryData(baseURL string, startTime int64) (model.EntryPackages, error) {
	url := baseURL + "&start=" + strconv.FormatInt(startTime, 10)
	common.Logger.Info("sync entry:", zap.String("url:", url))
	client := &http.Client{Timeout: time.Second * 5}
	res, err := client.Get(url)
	if err != nil {
		common.Logger.Error("get entry data  fail", zap.Error(err))
		return nil, err
	}
	if res.StatusCode != 200 {
		common.Logger.Error("get entry data fail code")
		return nil, err
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	var entryPackages model.EntryPackages
	errJson := json.Unmarshal(body, &entryPackages)
	if errJson != nil {
		common.Logger.Error("get entry data  fail", zap.Error(errJson))
		return nil, err
	}
	return entryPackages, nil
}

func syncEntry(redisClient *redis.Client, provider *model.SyncProvider, lastSyncTime int64) error {
	if lastSyncTime == 0 {
		currentUtcTime := time.Now().UTC()
		checkUtcTime := currentUtcTime.AddDate(0, 0, -int(provider.EntrySyncDate))
		lastSyncTime = int64(checkUtcTime.Unix())
	} else {
		lastSyncTime = lastSyncTime - 6*60*60
	}
	entryPackages, err := fetchEntryData(provider.EntryUrl, lastSyncTime)
	if err != nil {
		return fmt.Errorf("failed to fetch entry data: %w", err)
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
			storge.SaveEntrySyncPackageData(redisClient, provider.Provider, saveData)
		}

	}
	return nil

}

func fetchModelNameFromUrl(url string) string {
	modelName := ""
	start := strings.Index(url, "model_name=")
	if start != -1 {
		start += len("model_name=")
		end := strings.Index(url[start:], "&")
		if end != -1 {
			modelName = url[start : start+end]
		} else {
			modelName = url[start:]
		}
	}
	return modelName
}

type SourceDataStruct struct {
	Users     []string
	Providers []model.AlgoSyncProviderResponseModel
}

func getUserSource() map[string]SourceDataStruct {
	userList := common.GetUserList()
	userSourceMap := make(map[string]SourceDataStruct)
	for _, user := range userList {
		sources := api.LoadSources(user)
		for source := range sources {
			if _, exists := userSourceMap[source]; !exists {
				sourceData := SourceDataStruct{
					Users:     []string{},
					Providers: sources[source],
				}
				userSourceMap[source] = sourceData
			}
			mapData := userSourceMap[source]
			mapData.Users = append(mapData.Users, user)
			userSourceMap[source] = mapData
		}
	}
	return userSourceMap
}

func doSyncTask() {
	common.Logger.Info("package sync  start...")
	startTimestamp := int64(time.Now().UTC().Unix())

	providerList := make(map[string]*model.SyncProvider, 0)
	redisClient := common.GetRDBClient()
	defer redisClient.Close()
	postgresClient := common.NewPostgresClient()
	defer postgresClient.Close()

	userSourceData := getUserSource()
	for source := range userSourceData {
		sourceData := userSourceData[source]
		for _, provider := range sourceData.Providers {
			entryProviderUrl := provider.EntryProvider.Url
			modelName := fetchModelNameFromUrl(entryProviderUrl)
			key := provider.Provider + provider.FeedName + "_" + modelName
			common.Logger.Info("generate sync provider", zap.String("entry url", entryProviderUrl), zap.String("key", key))
			p, exist := providerList[key]
			if exist {
				if p.EntrySyncDate < provider.EntryProvider.SyncDate {
					p.EntrySyncDate = provider.EntryProvider.SyncDate
				}
			} else {
				var providerSetting model.SyncProvider
				providerSetting.FeedName = provider.FeedName
				providerSetting.Provider = provider.Provider
				providerSetting.FeedUrl = provider.FeedProvider.Url
				providerSetting.EntrySyncDate = provider.EntryProvider.SyncDate
				providerSetting.EntryUrl = provider.EntryProvider.Url
				providerList[key] = &providerSetting
			}
			syncFeed(postgresClient, redisClient, provider, source)
		}
	}

	for key, provider := range providerList {
		lastSyncTimeStr, _ := api.GetRedisConfig(key, "last_sync_time").(string)
		lastSyncTime, _ := strconv.ParseInt(lastSyncTimeStr, 10, 64)
		common.Logger.Info("sync  start", zap.String("last sync time str", lastSyncTimeStr), zap.Int64("last sync time", lastSyncTime), zap.Int64("now time", startTimestamp))
		if lastSyncTimeStr == "" || startTimestamp > lastSyncTime+10*60 {
			syncErr := syncEntry(redisClient, provider, lastSyncTime)
			if syncErr == nil {
				api.SetRedisConfig(key, "last_sync_time", startTimestamp)
			}
		}

	}
	common.Logger.Info("feed and entry packages sync  end")
	common.Logger.Info("package sync  end")
}

func main() {
	//common.Logger.Info("crawler task start 10...")
	//doSyncTask()
	//common.Logger.Info("crawler task end...")
	common.K8sTest()
	c := cron.New(cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger)))
	argoCheckCr := "@every 1m"
	c.AddFunc(argoCheckCr, func() {
		common.Logger.Info("do task  ...")
		common.K8sTest()
	})
}

func main1() {
	common.Logger.Info("sync task start 10...")
	//c := cron.New()
	c := cron.New(cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger)))
	argoCheckCr := "@every " + common.GeSyncFrequency() + "m"
	c.AddFunc(argoCheckCr, func() {
		common.Logger.Info("do crawler task  ...")
		doSyncTask()
	})
	c.Start()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	signal.Notify(stop, syscall.SIGTERM)
	<-stop
	common.Logger.Info("crawler task end...")
}
