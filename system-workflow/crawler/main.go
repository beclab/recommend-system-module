package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"bytetrade.io/web3os/system_workflow/api"
	"github.com/robfig/cron/v3"

	"sync"

	"bytetrade.io/web3os/system_workflow/common"
	"bytetrade.io/web3os/system_workflow/model"

	"go.uber.org/zap"
)

func loadSources() []string {
	//url := "http://recommend-service.os-system:6755/recommend-service/v1/status/recommenddev/" + common.GetTermiusUserName()
	url := "http://app-service.os-system:6755/app-service/v1/recommenddev/" + common.GetTermiusUserName() + "/status"
	client := &http.Client{Timeout: time.Second * 5}
	res, err := client.Get(url)
	//res, err := http.Get(url)
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
	}
	sourceArr := make([]string, 0)
	for _, argo := range response.Data {
		sourceArr = append(sourceArr, argo.Metadata.Name)
	}
	return sourceArr
}

func doCrawler(source string, list []model.EntryCrawlerModel) {
	if len(list) > 0 {
		addList := make([]*model.EntryAddModel, 0)
		for _, entry := range list {
			rawContent := common.GetUTF8ValidString(fetchRawContnt(entry.Url))
			if rawContent != "" {
				var addEntry model.EntryAddModel
				addEntry.Url = entry.Url
				addEntry.Source = source
				addEntry.RawContent = rawContent
				addEntry.Crawler = true

				addList = append(addList, &addEntry)
				if len(addList) >= 10 {
					api.UpdateEntriesInMongo(addList)
					addList = make([]*model.EntryAddModel, 0)
				}
			}
		}
		api.UpdateEntriesInMongo(addList)
	}

}

func fetchRawContnt(url string) string {

	client := &http.Client{Timeout: 5 * time.Second}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Connection", "close")

	response, err := client.Do(req)

	if err != nil {
		common.Logger.Error("crawling entry rawContent error", zap.String("url", url), zap.Error(err))
		return ""
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		common.Logger.Error("scraper: unable to download web page", zap.Int("statuscode", response.StatusCode), zap.String("url", url))
		return ""
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		common.Logger.Error("scraper fail to get response", zap.String("url", url), zap.Error(err))

		return ""
	}
	return string(body)
}

func doCrawlerTask() {
	sources := loadSources()
	startTimestamp := int64(time.Now().UTC().Unix())
	workNum := common.ParseInt(os.Getenv("CRAWLER_WORKER_POOL"), 6)
	for _, source := range sources {
		lastPrerankTimeStr, _ := api.GetRedisConfig(source, "last_prerank_time").(string)
		lastPrerankTime, _ := strconv.ParseInt(lastPrerankTimeStr, 10, 64)
		lastCrawlerTimeStr, _ := api.GetRedisConfig(source, "last_crawler_time").(string)
		lastCrawlerTime, _ := strconv.ParseInt(lastCrawlerTimeStr, 10, 64)
		common.Logger.Info("crawler  start ", zap.String("source:", source), zap.Int64("last prerank time:", lastPrerankTime), zap.Int64("last crawler time:", lastCrawlerTime))
		if lastPrerankTimeStr != "" && (lastCrawlerTimeStr == "" || lastPrerankTime > lastCrawlerTime) {
			limit := 100
			offset := 0
			crawlerList := make([]model.EntryCrawlerModel, 0)
			sum, crawlerData := api.GetUncrawleredList(offset, limit, source)
			crawlerList = append(crawlerList, crawlerData...)
			//sum := crawlerData.Count
			for i := 1; i*limit < sum; i++ {
				common.Logger.Info("get crawler data ", zap.String("source:", source), zap.Int("page", i))
				_, crawlerData := api.GetUncrawleredList(limit*i, limit, source)
				crawlerList = append(crawlerList, crawlerData...)
			}

			if len(crawlerList) > limit {
				var wg sync.WaitGroup
				wg.Add(workNum)
				perCount := len(crawlerList) / workNum
				for i := 0; i < workNum; i++ {
					start := i * perCount
					end := start + perCount
					if i == workNum-1 {
						end = len(crawlerList)
					}
					go func() {
						common.Logger.Info(fmt.Sprintf("start:%d,end:%d", start, end))
						list := crawlerList[start:end]
						doCrawler(source, list)
						//time.Sleep(time.Second * 1)
						wg.Done()
					}()
				}
				wg.Wait()

			} else {
				doCrawler(source, crawlerList)
			}
			api.SetRedisConfig(source, "last_crawler_time", startTimestamp)
			common.Logger.Info("crawler  end ", zap.String("source:", source), zap.Int("rank len:", len(crawlerList)), zap.Int64("change last_crawler_time time:", startTimestamp))
		}
	}

	common.Logger.Info("crawler fetch content end")

}

func main() {
	common.Logger.Info("crawler task start ...")
	c := cron.New(cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger)))

	argoCheckCr := "@every " + common.GetCrawlerFrequency() + "m"
	c.AddFunc(argoCheckCr, func() {
		common.Logger.Info("do crawler task  ...")
		doCrawlerTask()
	})
	c.Start()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	signal.Notify(stop, syscall.SIGTERM)
	<-stop
	common.Logger.Info("crawler task end...")
}
