package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"

	"bytetrade.io/web3os/system_workflow/api"
	"github.com/robfig/cron/v3"

	"bytetrade.io/web3os/system_workflow/common"
	"bytetrade.io/web3os/system_workflow/model"

	"go.uber.org/zap"
)

func urlToUniqueString(url string) string {
	hash := sha256.New()
	hash.Write([]byte(url))
	return hex.EncodeToString(hash.Sum(nil))
}

func doCrawler(source string, list []model.EntryCrawlerModel) {
	cacheDir := "/appCache/rss/"
	if len(list) > 0 {
		addList := make([]*model.EntryAddModel, 0)
		for _, entry := range list {
			primaryDomain := common.GetPrimaryDomain(entry.Url)
			fileName := urlToUniqueString(entry.Url)
			path := filepath.Join(cacheDir, primaryDomain, fileName)
			rawContent := ""
			if common.IsFileExist(path) {
				rawContent, _ = common.ReadFile(path)
			}
			if rawContent == "" {
				rawContent = common.GetUTF8ValidString(fetchRawContnt(entry.Url))
				if rawContent != "" {
					common.CreateNotExistDirectory(filepath.Join(cacheDir, primaryDomain), "save raw content"+primaryDomain)
					common.FileToSave(path, []byte(rawContent))
				}
			}

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
	userList := common.GetUserList()
	for _, user := range userList {
		sources := api.LoadSources(user)
		startTimestamp := int64(time.Now().UTC().Unix())
		workNum := common.ParseInt(os.Getenv("CRAWLER_WORKER_POOL"), 6)
		for source := range sources {
			lastPrerankTimeStr, _ := api.GetRedisConfig(user, source, "last_prerank_time").(string)
			lastPrerankTime, _ := strconv.ParseInt(lastPrerankTimeStr, 10, 64)
			lastCrawlerTimeStr, _ := api.GetRedisConfig(user, source, "last_crawler_time").(string)
			lastCrawlerTime, _ := strconv.ParseInt(lastCrawlerTimeStr, 10, 64)
			common.Logger.Info("crawler  start ", zap.String("user:", user), zap.String("source:", source), zap.Int64("last prerank time:", lastPrerankTime), zap.Int64("last crawler time:", lastCrawlerTime))
			if lastPrerankTimeStr != "" && (lastCrawlerTimeStr == "" || lastPrerankTime > lastCrawlerTime) {
				limit := 100
				offset := 0
				crawlerList := make([]model.EntryCrawlerModel, 0)
				sum, crawlerData := api.GetUncrawleredList(user, offset, limit, source)
				crawlerList = append(crawlerList, crawlerData...)
				//sum := crawlerData.Count
				for i := 1; i*limit < sum; i++ {
					common.Logger.Info("get crawler data ", zap.String("source:", source), zap.Int("page", i))
					_, crawlerData := api.GetUncrawleredList(user, limit*i, limit, source)
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
				api.SetRedisConfig(user, source, "last_crawler_time", startTimestamp)
				common.Logger.Info("crawler  end ", zap.String("source:", source), zap.Int("rank len:", len(crawlerList)), zap.Int64("change last_crawler_time time:", startTimestamp))
			}
		}
	}

	common.Logger.Info("crawler fetch content end")

}

func main1() {
	common.Logger.Info("crawler task start ...")
	doCrawlerTask()
	common.Logger.Info("crawler task end...")
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
