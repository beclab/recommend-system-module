package search

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"bytetrade.io/web3os/backend-server/common"
	"bytetrade.io/web3os/backend-server/model"
	"go.uber.org/zap"
)

func InputRSS(notificationData *model.NotificationData) string {

	if !common.GetZincRpcStart() {
		return ""
	}
	requestBytes, err := json.Marshal(notificationData)
	if err != nil {
		common.Logger.Error("InputRSS request marshal error", zap.String("name:", notificationData.Name))
		return ""
	}
	//common.Logger.Info("InputRSS request ", zap.String("body:", string(requestBytes)))
	common.Logger.Info("InputRSS request ", zap.String("name:", notificationData.Name))
	bodyReader := bytes.NewReader(requestBytes)
	requestUrl := "http://localhost:6317/api/input?index=Rss"
	req, err := http.NewRequest(http.MethodPost, requestUrl, bodyReader)

	if err != nil {
		common.Logger.Error("client: could not create request", zap.Error(err))
		return ""
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Access-Control-Allow-Origin", "*")
	req.Header.Set("Access-Control-Allow-Headers", "X-Requested-With,Content-Type")
	req.Header.Set("Access-Control-Allow-Methods", "PUT,POST,GET,DELETE,OPTIONS")
	//req.Header.Set("X-Access-Token", accessToken)

	client := http.Client{
		Timeout: 3 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		common.Logger.Error("client: error making http request", zap.Error(err))
		return ""
	}

	if resp.StatusCode != http.StatusOK {
		common.Logger.Error("status code error:", zap.String("status", resp.Status))
		return ""
	}
	defer resp.Body.Close()

	var r model.MessageNotificationResponse
	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		common.Logger.Error("response decode error")
		return ""
	}

	if r.Code != 0 {
		common.Logger.Error("response decode error:", zap.Int("code", r.Code))
		return ""
	}
	common.Logger.Info("Input RSS ", zap.String("data", r.Data))
	return r.Data

}

type DelRssReqStru struct {
	DocId string `json:"docId"`
}

func DeleteRSS(entryDocIds []string) {

	requestUrl := "http://localhost:6317/api/delete?index=Rss"
	for _, docId := range entryDocIds {

		reqStr := DelRssReqStru{
			DocId: docId,
		}

		requestBytes, _ := json.Marshal(reqStr)
		common.Logger.Info("request bodys", zap.String("body", string(requestBytes)))

		bodyReader := bytes.NewReader(requestBytes)

		req, err := http.NewRequest("POST", requestUrl, bodyReader)
		common.Logger.Info("request deleteRss docId", zap.String("docID", docId))
		if err != nil {
			common.Logger.Error("client: could not create request", zap.Error(err))
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Access-Control-Allow-Origin", "*")
		req.Header.Set("Access-Control-Allow-Headers", "X-Requested-With,Content-Type")
		req.Header.Set("Access-Control-Allow-Methods", "PUT,POST,GET,DELETE,OPTIONS")
		//req.Header.Set("X-Access-Token", accessToken)

		client := http.Client{
			Timeout: 3 * time.Second,
		}

		resp, err := client.Do(req)
		if err != nil {
			common.Logger.Error("client: error making http request", zap.Error(err))
			return
		}
		defer resp.Body.Close()

		body2, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		common.Logger.Info("delete rss resp body2", zap.String("body", string(body2)))
	}

}

type QueryRssReqStru struct {
	Query string `json:"query"`
	Limit int    `json:"limit"`
}

func QueryRSS(query string) string {
	common.Logger.Info("queryRSS", zap.String("query", query))

	requestUrl := "http://localhost:6317/api/query?index=Rss"

	reqStr := QueryRssReqStru{
		Query: query,
		Limit: 10,
	}

	requestBytes, _ := json.Marshal(reqStr)
	common.Logger.Info("query Rss request bodys", zap.String("body", string(requestBytes)))

	bodyReader := bytes.NewReader(requestBytes)
	req, err := http.NewRequest("POST", requestUrl, bodyReader)
	if err != nil {
		common.Logger.Error("client: could not create request", zap.Error(err))
		return ""
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Access-Control-Allow-Origin", "*")
	req.Header.Set("Access-Control-Allow-Headers", "X-Requested-With,Content-Type")
	req.Header.Set("Access-Control-Allow-Methods", "PUT,POST,GET,DELETE,OPTIONS")
	//req.Header.Set("X-Access-Token", accessToken)

	client := http.Client{
		Timeout: 3 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		common.Logger.Error("client: error making http request", zap.Error(err))
		return ""
	}
	defer resp.Body.Close()

	body2, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	return string(body2)

}
