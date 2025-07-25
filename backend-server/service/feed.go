package service

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"bytetrade.io/web3os/backend-server/common"
	"bytetrade.io/web3os/backend-server/http/client"
	"bytetrade.io/web3os/backend-server/model"
	"bytetrade.io/web3os/backend-server/reader/browser"
	"bytetrade.io/web3os/backend-server/reader/icon"
	"bytetrade.io/web3os/backend-server/reader/parser"
	"go.uber.org/zap"

	"bytetrade.io/web3os/backend-server/storage"
)

func RssParseFromURL(feedURL string) *model.Feed {
	request := client.NewClientWithConfig(feedURL)
	response, requestErr := browser.Exec(request)
	if requestErr != nil {
		common.Logger.Error("rss parse browse request error", zap.String("feedURL", feedURL), zap.Error(requestErr))
		return nil
	}
	updatedFeed, parseErr := parser.ParseFeed(response.EffectiveURL, response.BodyAsString())
	if parseErr != nil {
		common.Logger.Error("rss parse  error", zap.String("feedURL", feedURL), zap.Error(requestErr))
		return nil
	}
	updatedFeed.FeedURL = feedURL
	icon := CheckFeedIcon(updatedFeed.SiteURL, "", false, false)
	if icon != nil {
		updatedFeed.IconMimeType = icon.MimeType
		updatedFeed.IconContent = fmt.Sprintf("%s;base64,%s", icon.MimeType, base64.StdEncoding.EncodeToString(icon.Content))
	}
	return updatedFeed
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

func getRssHubCookieDomain(domain string) string {
	if strings.HasPrefix(domain, "bilibili") {
		return "bilibili.com"
	}
	if strings.HasPrefix(domain, "twitter") {
		return ".x.com"
	}
	return ""
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

// RefreshFeed refreshes a feed.
// func RefreshFeed(store *storage.Storage, contentPool *contentworker.ContentPool, feedID string) {
func RefreshFeed(store *storage.Storage, feedID string) {
	originalFeed, storeErr := store.GetFeedById(feedID)
	if storeErr != nil {
		common.Logger.Error("refresh feed load from db error id", zap.String("feedId", feedID), zap.Error(storeErr))
	}

	if originalFeed == nil {
		common.Logger.Error("Feed  not found", zap.String("feedId", feedID))
		return
	}
	common.Logger.Info("refresh feed", zap.String("feedurl", originalFeed.FeedURL), zap.String("etag header", originalFeed.EtagHeader), zap.String("last modified header", originalFeed.LastModifiedHeader))
	feedUrl := originalFeed.FeedURL
	feedDomain := common.Domain(feedUrl)
	var updatedFeed *model.Feed
	if strings.HasPrefix(feedUrl, "wechat://") {
		wechatAcc := feedUrl[9:]
		var avatar string
		updatedFeed, avatar = RefreshWeChatFeed(wechatAcc)
		if avatar != "" {
			icon, _ := icon.DownloadIcon(avatar, originalFeed.UserAgent, originalFeed.FetchViaProxy, originalFeed.AllowSelfSignedCertificates)
			if icon != nil && icon.MimeType != "" {
				originalFeed.IconMimeType = icon.MimeType
				originalFeed.IconContent = fmt.Sprintf("%s;base64,%s", icon.MimeType, base64.StdEncoding.EncodeToString(icon.Content))
			}
		}

	} else if feedDomain == "www.youtube.com" {
		var avatar string
		updatedFeed, avatar = RefreshYoutubeFeed(store, feedUrl, originalFeed.ID)
		if avatar != "" {
			icon, _ := icon.DownloadIcon(avatar, originalFeed.UserAgent, originalFeed.FetchViaProxy, originalFeed.AllowSelfSignedCertificates)
			if icon != nil && icon.MimeType != "" {
				originalFeed.IconMimeType = icon.MimeType
				originalFeed.IconContent = fmt.Sprintf("%s;base64,%s", icon.MimeType, base64.StdEncoding.EncodeToString(icon.Content))
			}
		}
	} else {
		rsshubCookie := ""
		feedURL := originalFeed.FeedURL
		if strings.HasPrefix(feedUrl, "rsshub://") {
			//rsshub sdk:
			//feedURL = common.GetRSSHubUrl() + "?path=/" + feedUrl[9:]
			//deploy rsshub
			common.Logger.Info(" rsshub feed refresh ", zap.String("feedpath", feedUrl[9:]))
			cookieDomain := getRssHubCookieDomain(feedUrl[9:])
			if cookieDomain != "" {
				cookie := generateRssHubCookie(originalFeed.BflUser, cookieDomain)
				common.Logger.Info(" rsshub feed cookie ", zap.String("domain", cookieDomain), zap.String("cookie", cookie))
				if len(cookie) > 0 {
					rsshubCookie = cookie
				}
			}
			feedURL = common.GetRSSHubUrl() + feedUrl[9:]

		}
		updatedFeed = rssRrefresh(store, originalFeed, feedURL, rsshubCookie)
		//updatedFeed = rssRrefresh(store, originalFeed)
		if updatedFeed != nil {
			originalFeed.EtagHeader = updatedFeed.EtagHeader
			originalFeed.LastModifiedHeader = updatedFeed.LastModifiedHeader
			//originalFeed.FeedURL = updatedFeed.FeedURL
			originalFeed.SiteURL = updatedFeed.SiteURL
			if originalFeed.SiteURL == "" {
				originalFeed.SiteURL = originalFeed.FeedURL
			}
			icon := CheckFeedIcon(originalFeed.SiteURL, originalFeed.UserAgent, originalFeed.FetchViaProxy, originalFeed.AllowSelfSignedCertificates)
			if icon != nil && icon.MimeType != "" {
				originalFeed.IconMimeType = icon.MimeType
				originalFeed.IconContent = fmt.Sprintf("%s;base64,%s", icon.MimeType, base64.StdEncoding.EncodeToString(icon.Content))
			} else {
				common.Logger.Error("feed icon get null!!!", zap.String("siteurl", originalFeed.SiteURL))
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
