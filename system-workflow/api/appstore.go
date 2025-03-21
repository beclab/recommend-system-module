package api

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"bytetrade.io/web3os/system_workflow/common"
	"bytetrade.io/web3os/system_workflow/model"
	"go.uber.org/zap"
)

func LoadSources(name string) map[string][]model.AlgoSyncProviderResponseModel {
	sourceMap := make(map[string][]model.AlgoSyncProviderResponseModel)
	url := "http://app-service.os-system:6755/app-service/v1/recommenddev/" + name + "/status"
	client := &http.Client{Timeout: time.Second * 5}
	res, err := client.Get(url)
	if err != nil {
		common.Logger.Error("get recommend service error", zap.String("url", url), zap.Error(err))
		return sourceMap
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	jsonStr := string(body)
	common.Logger.Info("get recommend service response: ", zap.String("url", url), zap.String("body", jsonStr))

	var response model.RecommendServiceResponseModel
	if err := json.Unmarshal(body, &response); err != nil {
		common.Logger.Error("json decode failed ", zap.String("url", url), zap.Error(err))
	}

	for _, argo := range response.Data {
		sourceMap[argo.Metadata.Name] = argo.SyncProvider
	}
	return sourceMap
}
