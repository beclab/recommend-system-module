package service

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"bytetrade.io/web3os/RSSync/common"
	"bytetrade.io/web3os/RSSync/model"
	"go.uber.org/zap"
)

func RefreshWeChatFeed(wechatAcc string) (*model.Feed, string) {
	var feed model.Feed
	avatar := ""
	url := common.GetWeChatFeedRefrshUrl() + "?wechatAccount=" + wechatAcc //+ "&lasttime=" + fmt.Sprintf("%d", checkAt.Unix())
	client := &http.Client{Timeout: time.Second * 5}
	res, err := client.Get(url)
	if err != nil {
		common.Logger.Error("wechat feed refresh error", zap.Error(err))
		return &feed, avatar
	}
	if res != nil {
		defer res.Body.Close()
	}
	body, _ := io.ReadAll(res.Body)

	var wechatEntries model.WeChatEntries
	if err := json.Unmarshal(body, &wechatEntries); err != nil {
		log.Print("json decode failed, err", err)
	}
	if len(wechatEntries) > 0 {
		entries := make([]*model.Entry, 0)
		for _, wechatEntry := range wechatEntries {
			entries = append(entries, model.GetEntryFromWeChatEntry(wechatEntry))
		}
		feed.Title = wechatEntries[0].AccountNickname
		avatar = wechatEntries[0].AccountAvatar
		feed.Entries = entries
	}
	common.Logger.Info("wechat feed refresh", zap.String("wechatAcc", wechatAcc), zap.Int("len", len(wechatEntries)))
	return &feed, avatar
}

func GetWeChatContent(entryUrl string) string {
	url := common.GetWeChatEntryContentUrl() + "?url=" + url.QueryEscape(entryUrl)
	client := &http.Client{Timeout: time.Second * 5}
	res, err := client.Get(url)
	if err != nil {
		common.Logger.Error("wechat entry get error: %v", zap.Error(err))
		return ""
	}
	if res != nil {
		defer res.Body.Close()
	}
	body, _ := io.ReadAll(res.Body)

	var response model.WechatEntryResponse
	if err := json.Unmarshal(body, &response); err != nil {
		log.Print("json decode failed, err", err)
	}

	return response.RawContent
}
