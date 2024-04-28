package storge

import (
	"encoding/json"
	"fmt"

	"bytetrade.io/web3os/system_workflow/common"
	"bytetrade.io/web3os/system_workflow/model"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func SaveFeedSync(rdb *redis.Client, provider, feedName string, data model.FeedSyncData) error {
	jsonFeedSyncSetting, err := json.Marshal(data)
	if err != nil {
		common.Logger.Error("marshal entrySyncSetting fail", zap.Error(err))
		return err
	}

	err = rdb.HSet(common.RedisCtx, fmt.Sprintf("feed_sync_%s", provider), feedName, jsonFeedSyncSetting).Err()

	if err != nil {
		common.Logger.Error("set feed sync setting fail", zap.Error(err))
		return err
	}
	return nil
}
func GetFeedSync(rdb *redis.Client, provider, feedName string) (*model.FeedSyncData, error) {

	exists, _ := rdb.HExists(common.RedisCtx, fmt.Sprintf("feed_sync_%s", provider), feedName).Result()
	if !exists {
		return nil, nil
	}

	jsonData, err := rdb.HGet(common.RedisCtx, fmt.Sprintf("feed_sync_%s", provider), feedName).Result()
	if err != nil {
		common.Logger.Error("get feed sync setting fail", zap.Error(err))
		return nil, err
	}
	var redisFeedSyncPackageData model.FeedSyncData
	unmarshalErr := json.Unmarshal([]byte(jsonData), &redisFeedSyncPackageData)
	if unmarshalErr != nil {
		common.Logger.Error("unmarshal feed sync setting fail", zap.Error(err))
		return nil, err
	}
	return &redisFeedSyncPackageData, nil
}

func SaveFeedSyncPackageData(rdb *redis.Client, provider string, data model.FeedSyncPackageData) error {

	jsonFeedSyncSetting, err := json.Marshal(data)
	if err != nil {
		common.Logger.Error("marshal entrySyncSetting fail", zap.Error(err))
		return err
	}

	err = rdb.HSet(common.RedisCtx, provider, data.Name, jsonFeedSyncSetting).Err()

	if err != nil {
		common.Logger.Error("set feed sync setting fail", zap.Error(err))
		return err
	}
	return nil
}

func GetFeedSyncPackageData(rdb *redis.Client, provider, name string) (*model.FeedSyncPackageData, error) {

	exists, _ := rdb.HExists(common.RedisCtx, provider, name).Result()
	if !exists {
		return nil, nil
	}

	jsonData, err := rdb.HGet(common.RedisCtx, provider, name).Result()
	if err != nil {
		common.Logger.Error("get feed sync setting fail", zap.Error(err))
		return nil, err
	}
	var redisFeedSyncPackageData model.FeedSyncPackageData
	unmarshalErr := json.Unmarshal([]byte(jsonData), &redisFeedSyncPackageData)
	if unmarshalErr != nil {
		common.Logger.Error("unmarshal feed sync setting fail", zap.Error(err))
		return nil, err
	}
	return &redisFeedSyncPackageData, nil
}

func SaveEntrySyncPackageData(rdb *redis.Client, provider string, data model.EntrySyncPackageData) error {

	jsonEntrySyncSetting, err := json.Marshal(data)
	if err != nil {
		common.Logger.Error("marshal entrySyncSetting fail", zap.Error(err))
		return err
	}

	subkey := fmt.Sprintf("%s_%s_%d", data.FeedName, data.ModelName, data.StartTime)
	err = rdb.HSet(common.RedisCtx, provider, subkey, jsonEntrySyncSetting).Err()

	if err != nil {
		common.Logger.Error("set entry sync setting fail", zap.Error(err))
		return err
	}
	return nil
}

func GetEntrySyncPackageData(rdb *redis.Client, provider, feedName, modelName string, startTime int64) (*model.EntrySyncPackageData, error) {

	subkey := fmt.Sprintf("%s_%s_%d", feedName, modelName, startTime)

	exists, _ := rdb.HExists(common.RedisCtx, provider, subkey).Result()
	if !exists {
		return nil, nil
	}

	jsonData, err := rdb.HGet(common.RedisCtx, provider, subkey).Result()
	if err != nil {
		common.Logger.Error("get entry sync setting fail", zap.Error(err))
		return nil, err
	}
	var redisEntrySyncPackageData model.EntrySyncPackageData
	unmarshalErr := json.Unmarshal([]byte(jsonData), &redisEntrySyncPackageData)
	if unmarshalErr != nil {
		common.Logger.Error("unmarshal entry sync setting fail", zap.Error(err))
		return nil, err
	}
	return &redisEntrySyncPackageData, nil
}
