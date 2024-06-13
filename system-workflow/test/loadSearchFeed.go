package main

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"bytetrade.io/web3os/system_workflow/common"
	"bytetrade.io/web3os/system_workflow/model"
	"bytetrade.io/web3os/system_workflow/protobuf_entity"
	"bytetrade.io/web3os/system_workflow/storge"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

func syncDiscoveryFeedloadPackage(mongoClient *mongo.Client, newPackage *model.DiscoveryFeedPackagInfo) {

	client := &http.Client{Timeout: time.Second * 5}
	res, err := client.Get(newPackage.Url)
	if err != nil {
		common.Logger.Error("get discovery feed package  fail", zap.Error(err))
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		common.Logger.Error("discovery feed fail to get response", zap.Error(err))
	}

	uncompressByte := common.DoZlibUnCompress(body)
	var allPackageList protobuf_entity.ListDiscoveryFeed
	unmarshalErr := proto.Unmarshal(uncompressByte, &allPackageList)
	if unmarshalErr != nil {
		common.Logger.Error("unmarshal all discovery feed object  error", zap.Error(unmarshalErr))
	}
	storge.RemoveDiscoveryFeed(mongoClient)
	for _, discoveryFeed := range allPackageList.DiscoveryFeeds {
		storge.CreateDiscoveryFeed(mongoClient, model.GetDiscoveryModel(discoveryFeed))
	}

}

func syncDiscoveryFeedPackage(mongoClient *mongo.Client) {
	url := common.GetSyncDiscoveryFeedPackageUrl()
	common.Logger.Info("sync discovery feedPackage:", zap.String("url:", url))
	client := &http.Client{Timeout: time.Second * 5}
	res, err := client.Get(url)
	if err != nil || res.StatusCode != 200 {
		common.Logger.Error("sync discovery feedPackage error", zap.Error(err))
		return
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		common.Logger.Error("read discovery feedPackage  fail", zap.Error(err))
		return
	}
	var packages model.DiscoveryFeedPackagInfos
	errJson := json.Unmarshal(body, &packages)
	if errJson != nil {
		common.Logger.Error("get discovery feedPackage data  fail", zap.Error(errJson))
		return
	}
	if len(packages) > 0 {
		syncDiscoveryFeedloadPackage(mongoClient, packages[0])
	}

}

func main() {

	mongoClient, _, closeMongoClientFun := common.NewMongodbConnection()

	defer closeMongoClientFun()

	syncDiscoveryFeedPackage(mongoClient)

}
