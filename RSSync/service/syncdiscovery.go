package service

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"bytetrade.io/web3os/RSSync/common"
	"bytetrade.io/web3os/RSSync/model"
	"bytetrade.io/web3os/RSSync/protobuf_entity"
	"bytetrade.io/web3os/RSSync/storage"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

func syncDiscoveryFeedloadPackage(store *storage.Storage, newPackage *model.DiscoveryFeedPackagInfo) error {

	client := &http.Client{Timeout: time.Second * 5}
	res, err := client.Get(newPackage.Url)
	if err != nil {
		common.Logger.Error("get discovery feed package  fail", zap.Error(err))
		return err
	}
	if res != nil {
		defer res.Body.Close()
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		common.Logger.Error("discovery feed fail to get response", zap.Error(err))
	}

	uncompressByte := common.DoZlibUnCompress(body)
	var allPackageList protobuf_entity.ListDiscoveryFeed
	unmarshalErr := proto.Unmarshal(uncompressByte, &allPackageList)
	if unmarshalErr != nil {
		common.Logger.Error("unmarshal all discovery feed object  error", zap.Error(unmarshalErr))
		return unmarshalErr
	}
	store.RemoveDiscoveryFeed()
	for _, discoveryFeed := range allPackageList.DiscoveryFeeds {
		store.CreateDiscoveryFeed(model.GetDiscoveryModel(discoveryFeed))
	}
	return nil
}

func SyncDiscoveryFeedPackage(store *storage.Storage) {
	saveData, _ := store.GetDiscoveryFeedPackage()
	if saveData != nil {
		common.Logger.Info("discovery feed is exist,no sync!")
		return
	}
	url := common.GetSyncDiscoveryFeedPackageUrl()
	common.Logger.Info("sync discovery feedPackage:", zap.String("url:", url))
	client := &http.Client{Timeout: time.Second * 5}
	res, err := client.Get(url)
	if err != nil || res.StatusCode != 200 {
		common.Logger.Error("sync discovery feedPackage error", zap.Error(err))
		return
	}
	if res != nil {
		defer res.Body.Close()
	}
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
		saveData, _ := store.GetDiscoveryFeedPackage()
		if saveData == nil || saveData.MD5 != packages[0].MD5 {
			loadErr := syncDiscoveryFeedloadPackage(store, packages[0])
			if loadErr == nil {
				store.SaveDiscoveryFeedPackage(*packages[0])
			}

		}
	}

}
