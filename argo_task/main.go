package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bytetrade.io/web3os/argo-task/common"
	"bytetrade.io/web3os/argo-task/model"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

func newArgoSyncTask() {
	nameSpace := common.GetNameSpace()
	body := common.GenerateArgoSyncPostData(nameSpace)
	log.Print(string(body))

	url := common.GetArgoUrl() + "/v1/cron-workflows/" + nameSpace
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))

	if err != nil {
		log.Print("new argo task  fail", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		jsonStr := string(body)
		fmt.Println("Response: ", jsonStr)

	} else {
		body, _ := io.ReadAll(resp.Body)
		jsonStr := string(body)
		fmt.Println("Get failed with error response: ", jsonStr)
	}
}
func argoSyncTaskCheck() {
	nameSpace := common.GetNameSpace()
	url := common.GetArgoUrl() + "/v1/cron-workflows/" + nameSpace + "/" + "recommend-task-sync"

	log.Print("url:", url)
	res, err := http.Get(url)
	if err != nil {
		log.Print("get argo data  fail", err)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	jsonStr := string(body)
	fmt.Println("sync Response: ", len(jsonStr))

	if res.StatusCode == 404 {
		newArgoSyncTask()
	}

}

func newArgoCrawlerTask() {
	nameSpace := common.GetNameSpace()
	body := common.GenerateArgoCrawlercPostData(nameSpace)
	log.Print(string(body))

	url := common.GetArgoUrl() + "/v1/cron-workflows/" + nameSpace
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))

	if err != nil {
		log.Print("new argo task  fail", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		jsonStr := string(body)
		fmt.Println("Response: ", jsonStr)

	} else {
		body, _ := io.ReadAll(resp.Body)
		jsonStr := string(body)
		fmt.Println("Get failed with error response: ", jsonStr)
	}
}
func argoCrawlerTaskCheck() {
	nameSpace := common.GetNameSpace()
	url := common.GetArgoUrl() + "/v1/cron-workflows/" + nameSpace + "/" + "recommend-task-crawler"

	log.Print("url:", url)
	res, err := http.Get(url)
	if err != nil {
		log.Print("get argo data  fail", err)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	jsonStr := string(body)
	fmt.Println("sync Response: ", len(jsonStr))

	if res.StatusCode == 404 {
		newArgoCrawlerTask()
	}

}

func getIsInstallRecommend() bool {
	url := "http://app-service.os-system:6755/app-service/v1/recommenddev/" + common.GetTermiusUserName() + "/status"
	client := &http.Client{Timeout: time.Second * 5}
	res, err := client.Get(url)
	if err != nil {
		log.Print("get appservice  fail", err)
		return true
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	jsonStr := string(body)
	log.Print("get recommend service response: ", url, jsonStr)
	var response model.RecommendServiceResponseModel
	if err := json.Unmarshal(body, &response); err != nil {
		log.Print("json decode failed ", zap.Error(err))
		return true
	}
	if len(response.Data) == 0 {
		return false
	}
	return true
}

func main() {
	log.Print("argo task start ...")

	c := cron.New()
	argoCheckCr := "@every " + common.GetTaskFrequency() + "m"
	c.AddFunc(argoCheckCr, func() {

		isInstallRecommended := getIsInstallRecommend()
		log.Print("do check task...is install", isInstallRecommended)
		if isInstallRecommended {
			argoSyncTaskCheck()
			argoCrawlerTaskCheck()
		}

	})
	c.Start()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	signal.Notify(stop, syscall.SIGTERM)
	<-stop
	log.Print("argo task end... ")
}
