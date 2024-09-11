package storage

import (
	"encoding/json"

	"bytetrade.io/web3os/backend-server/common"
	"bytetrade.io/web3os/backend-server/model"

	"go.uber.org/zap"
)

func (s *Storage) SaveDiscoveryFeedPackage(data model.DiscoveryFeedPackagInfo) error {
	redisCacheData, err := json.Marshal(data)
	if err != nil {
		common.Logger.Error("marshal entrySyncSetting fail", zap.Error(err))
		return err
	}

	err = s.redisdb.Set("discovery_feed", redisCacheData, 0).Err()

	if err != nil {
		common.Logger.Error("set feed sync setting fail", zap.Error(err))
		return err
	}
	return nil
}
func (s *Storage) GetDiscoveryFeedPackage() (*model.DiscoveryFeedPackagInfo, error) {

	jsonData, err := s.redisdb.Get("discovery_feed").Result()
	if err != nil {
		common.Logger.Error("get feed sync setting fail", zap.Error(err))
		return nil, err
	}
	if jsonData == "" {
		return nil, nil
	}
	var redisCacheData model.DiscoveryFeedPackagInfo
	unmarshalErr := json.Unmarshal([]byte(jsonData), &redisCacheData)
	if unmarshalErr != nil {
		common.Logger.Error("unmarshal feed sync setting fail", zap.Error(err))
		return nil, err
	}
	return &redisCacheData, nil
}
