package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"bytetrade.io/web3os/argo-task/common"
	"github.com/robfig/cron/v3"
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

func main() {
	log.Print("argo task start ...")

	c := cron.New()
	argoCheckCr := "@every " + common.GetTaskFrequency() + "m"
	c.AddFunc(argoCheckCr, func() {
		log.Print("do check task...")
		argoSyncTaskCheck()
		argoCrawlerTaskCheck()
	})
	c.Start()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	signal.Notify(stop, syscall.SIGTERM)
	<-stop
	log.Print("argo task end... ")
}
