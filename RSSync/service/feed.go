package service

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"bytetrade.io/web3os/RSSync/common"
	"bytetrade.io/web3os/RSSync/http/client"
	"bytetrade.io/web3os/RSSync/model"
	"bytetrade.io/web3os/RSSync/reader/browser"
	"bytetrade.io/web3os/RSSync/reader/icon"
	"bytetrade.io/web3os/RSSync/reader/parser"
	"go.uber.org/zap"

	"bytetrade.io/web3os/RSSync/storage"
)

const (
	bilibiliDomain = "bilibili.com"
	twitterDomain  = ".x.com"
)

func getRssHubCookieDomain(domain string) string {
	switch {
	case strings.HasPrefix(domain, "bilibili"):
		return bilibiliDomain
	case strings.HasPrefix(domain, "twitter"):
		return twitterDomain
	default:
		return ""
	}
}

func generateRssHubCookie(bflUser, domain string) string {
	domainList := client.LoadCookieInfoManager(bflUser, domain, domain)
	cookies := ""
	for _, domainItem := range domainList {
		for _, record := range domainItem.Records {
			if domain == ".x.com" {
				if record.Name == "auth_token" {
					return record.Value
				}
			} else {
				cookies = cookies + record.Name + "=" + record.Value + ";"
			}
		}
	}
	return cookies
}

func rssRrefresh(store *storage.Storage, feed *model.Feed, feedURL string, rsshubCookie string) *model.Feed {
	common.Logger.Info("start refresh feed ", zap.String("feedId", feed.ID), zap.String("feed url:", feedURL))
	request := client.NewClientWithConfig(feedURL)
	request.WithBflUser(feed.BflUser)
	request.WithCredentials(feed.Username, feed.Password)
	request.WithUserAgent(feed.UserAgent)
	request.WithCookie(feed.Cookie)
	request.WithRssHubCookie(rsshubCookie)
	request.AllowSelfSignedCertificates = feed.AllowSelfSignedCertificates

	if !feed.IgnoreHTTPCache {
		request.WithCacheHeaders(feed.EtagHeader, feed.LastModifiedHeader)
	}

	if feed.FetchViaProxy {
		request.WithProxy()
	}

	response, requestErr := browser.Exec(request)
	if requestErr != nil {
		feed.ParsingErrorCount++
		store.UpdateFeedError(feed.ID, feed)
		common.Logger.Error("refresh feed  browser request", zap.String("feedId", feed.ID), zap.Error(requestErr))
		return nil
	}

	common.Logger.Info("[RefreshFeed] Feed step2", zap.String("etag header", feed.EtagHeader), zap.String("last modified header", feed.LastModifiedHeader))

	if feed.IgnoreHTTPCache || response.IsModified(feed.EtagHeader, feed.LastModifiedHeader) {
		updatedFeed, parseErr := parser.ParseFeed(response.EffectiveURL, response.BodyAsString())
		if parseErr != nil {
			feed.ParsingErrorCount++
			store.UpdateFeedError(feed.ID, feed)
			common.Logger.Error("refresh feed ParseFeed error id: %s,%v", zap.String("feedId", feed.ID), zap.Error(parseErr))
			return nil
		}
		updatedFeed.EtagHeader = response.ETag
		updatedFeed.LastModifiedHeader = response.LastModified
		updatedFeed.FeedURL = response.EffectiveURL
		common.Logger.Info("[RefreshFeed] Feed ", zap.String("feedId", feed.ID), zap.Int("entry size", len(updatedFeed.Entries)))
		return updatedFeed
	} else {
		common.Logger.Debug("[RefreshFeed] Feed #%s not modified", zap.String("feedId", feed.ID))
	}
	return nil
}

func updateFeedIconByIconUrl(originalFeed *model.Feed, avatar string) {
	if avatar != "" {
		icon, _ := icon.DownloadIcon(avatar, originalFeed.UserAgent, originalFeed.FetchViaProxy, originalFeed.AllowSelfSignedCertificates)
		if icon != nil && icon.MimeType != "" {
			originalFeed.IconMimeType = icon.MimeType
			originalFeed.IconContent = fmt.Sprintf("%s;base64,%s", icon.MimeType, base64.StdEncoding.EncodeToString(icon.Content))
		}
	}
}

func handleDefaultFeed(store *storage.Storage, originalFeed *model.Feed, feedUrl string) *model.Feed {
	rsshubCookie := ""
	if strings.HasPrefix(feedUrl, "rsshub://") {
		rsshubCookie = generateRssHubCookie(originalFeed.BflUser, getRssHubCookieDomain(feedUrl[9:]))
		feedUrl = common.GetRSSHubUrl() + feedUrl[9:]
	}
	return rssRrefresh(store, originalFeed, feedUrl, rsshubCookie)
}

// RefreshFeed refreshes a feed.
// func RefreshFeed(store *storage.Storage, contentPool *contentworker.ContentPool, feedID string) {
func RefreshFeed(store *storage.Storage, feedID string) {
	originalFeed, storeErr := store.GetFeedById(feedID)
	if storeErr != nil || originalFeed == nil {
		common.Logger.Error("refresh feed load from db error id", zap.String("feedId", feedID), zap.Error(storeErr))
		return
	}
	common.Logger.Info("refresh feed", zap.String("feedurl", originalFeed.FeedURL), zap.String("etag header", originalFeed.EtagHeader), zap.String("last modified header", originalFeed.LastModifiedHeader))
	feedUrl := originalFeed.FeedURL
	feedDomain := common.Domain(feedUrl)
	var avatar string
	var updatedFeed *model.Feed
	switch {
	case strings.HasPrefix(feedUrl, "wechat://"):
		updatedFeed, avatar = RefreshWeChatFeed(feedUrl[9:])
		updateFeedIconByIconUrl(originalFeed, avatar)
	case feedDomain == "www.youtube.com":
		updatedFeed, avatar = RefreshYoutubeFeed(store, feedUrl, originalFeed.ID)
		updateFeedIconByIconUrl(originalFeed, avatar)
	default:
		updatedFeed = handleDefaultFeed(store, originalFeed, feedUrl)
		if updatedFeed != nil {
			originalFeed.EtagHeader = updatedFeed.EtagHeader
			originalFeed.LastModifiedHeader = updatedFeed.LastModifiedHeader
			originalFeed.SiteURL = updatedFeed.SiteURL
			if originalFeed.SiteURL == "" {
				originalFeed.SiteURL = originalFeed.FeedURL
			}
			icon := CheckFeedIcon(originalFeed.SiteURL, originalFeed.UserAgent, originalFeed.FetchViaProxy, originalFeed.AllowSelfSignedCertificates)
			if icon != nil && icon.MimeType != "" {
				originalFeed.IconMimeType = icon.MimeType
				originalFeed.IconContent = fmt.Sprintf("%s;base64,%s", icon.MimeType, base64.StdEncoding.EncodeToString(icon.Content))
			}
		}
	}

	if updatedFeed != nil {
		if originalFeed.Title == "" {
			originalFeed.Title = updatedFeed.Title
		}
		ProcessFeedEntries(store, originalFeed, updatedFeed.Entries)
	}

	originalFeed.CheckedAt = time.Now()
	originalFeed.ParsingErrorCount = 0

	if storeErr := store.UpdateFeed(feedID, originalFeed); storeErr != nil {
		originalFeed.ParsingErrorCount++
		store.UpdateFeedError(feedID, originalFeed)
	}

}

func CheckFeedIcon(websiteURL, userAgent string, fetchViaProxy, allowSelfSignedCertificates bool) *model.Icon {
	iconO, _ := icon.FindIcon(websiteURL, userAgent, fetchViaProxy, allowSelfSignedCertificates)

	return iconO

}
