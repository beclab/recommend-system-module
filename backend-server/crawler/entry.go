package crawler

import (
	"fmt"
	"io"
	"net/url"
	"strings"

	"bytetrade.io/web3os/backend-server/common"
	"bytetrade.io/web3os/backend-server/crawler/bskyapi"
	"bytetrade.io/web3os/backend-server/crawler/feishu"
	notionClient "bytetrade.io/web3os/backend-server/crawler/notionapi"
	"bytetrade.io/web3os/backend-server/crawler/notionapi/tohtml"
	"bytetrade.io/web3os/backend-server/crawler/nytimes"
	"bytetrade.io/web3os/backend-server/crawler/quora"
	"bytetrade.io/web3os/backend-server/crawler/tbilibili"
	"bytetrade.io/web3os/backend-server/crawler/threads"
	"bytetrade.io/web3os/backend-server/crawler/washingtonpost"
	"bytetrade.io/web3os/backend-server/crawler/weibo"
	wolaiapi "bytetrade.io/web3os/backend-server/crawler/wolaiapi"
	"bytetrade.io/web3os/backend-server/crawler/wsj"
	"bytetrade.io/web3os/backend-server/crawler/ximalaya"
	"bytetrade.io/web3os/backend-server/http/client"
	"bytetrade.io/web3os/backend-server/knowledge"
	"bytetrade.io/web3os/backend-server/model"
	"github.com/beclab/article-extractor/processor"
	"go.uber.org/zap"
)

func handlerGenerateEntry(entry *model.Entry, newEntry *model.Entry) {
	if newEntry != nil {
		entry.FullContent = newEntry.FullContent
		entry.MediaContent = newEntry.MediaContent
		entry.MediaUrl = newEntry.MediaUrl
		entry.MediaType = newEntry.MediaType
		entry.Author = newEntry.Author
		entry.Title = newEntry.Title
		entry.PublishedAt = newEntry.PublishedAt
		entry.ImageUrl = common.GetImageUrlFromContent(entry.FullContent)
	}
}

func handleX(entry *model.Entry) {
	twitterID := ""
	parts := strings.Split(entry.URL, "status/")
	if len(parts) > 1 {
		twitterID = strings.TrimSpace(parts[1])
	}
	fmt.Println("twitter ID:", twitterID)
	twitterEntry := knowledge.FetchTwitterContent(entry.BflUser, twitterID, entry.URL)
	handlerGenerateEntry(entry, twitterEntry)
	entry.Language = "en"
}
func handleXHS(entry *model.Entry) {
	xshEntry := knowledge.FetchXHSContent(entry.URL, entry.BflUser)
	handlerGenerateEntry(entry, xshEntry)
	entry.Language = "zh-cn"
}
func handleBsky(entry *model.Entry) {
	bskyEntry := bskyapi.Fetch(entry.BflUser, entry.URL)
	handlerGenerateEntry(entry, bskyEntry)
	entry.Language = "en"
}

func handleThreads(entry *model.Entry) {
	threadsEntry := threads.Fetch(entry.URL)
	handlerGenerateEntry(entry, threadsEntry)
	entry.Language = "en"
}

func handleQtfm(entry *model.Entry) {
	handleYtdlp(entry)
	if entry.Title != "" {
		entry.MediaUrl = entry.URL
		entry.MediaType = "audio"
	}
}

func handleTBilibili(entry *model.Entry) {
	xshEntry := tbilibili.Fetch(entry.BflUser, entry.URL)
	handlerGenerateEntry(entry, xshEntry)
	entry.Language = "zh-cn"
}

func handleWeibo(entry *model.Entry) {
	weiboEntry := weibo.Fetch(entry.BflUser, entry.URL)
	handlerGenerateEntry(entry, weiboEntry)
	entry.Language = "zh-cn"
}
func EntryCrawler(entry *model.Entry, feedUrl, userAgent, cookie string, certificates, fetchViaProxy bool) {
	primaryDomain := common.GetPrimaryDomain(entry.URL)
	urlDomain := domain(entry.URL)
	common.Logger.Info("crawler entry start", zap.String("url", entry.URL), zap.String("primary domain:", primaryDomain))

	switch primaryDomain {
	case "bilibili.com":
		entry.FullContent = entry.Content
		entry.Language = "zh-cn"
		if urlDomain == "t.bilibili.com" {
			handleTBilibili(entry)
		} else {
			handleDefault(entry, feedUrl, userAgent, cookie, certificates, fetchViaProxy)
		}
	case "x.com":
		handleX(entry)
	case "weibo.com":
		handleWeibo(entry)
	case "xiaohongshu.com":
		handleXHS(entry)
	case "bsky.app":
		handleBsky(entry)
	case "threads.net":
		handleThreads(entry)
	case "qtfm.cn":
		handleQtfm(entry)
	default:
		handleDefault(entry, feedUrl, userAgent, cookie, certificates, fetchViaProxy)
	}
	common.Logger.Info("crawler entry finished", zap.String("url", entry.URL))
}

func handleYtdlp(entry *model.Entry) {
	ytdlpEntry := knowledge.LoadMetaFromYtdlp(entry.BflUser, entry.URL)
	if ytdlpEntry != nil {
		if ytdlpEntry.Author != "" {
			entry.Author = ytdlpEntry.Author
		}
		if ytdlpEntry.Title != "" {
			entry.Title = ytdlpEntry.Title
		}
		if ytdlpEntry.PublishedAt != 0 {
			entry.PublishedAt = ytdlpEntry.PublishedAt
		}
		if ytdlpEntry.FullContent != "" {
			entry.FullContent = ytdlpEntry.FullContent
		}

	}
}

func handleDefault(entry *model.Entry, feedUrl, userAgent, cookie string, certificates, fetchViaProxy bool) {
	entry.RawContent = FetchRawContnt(
		entry.BflUser,
		entry.URL,
		entry.Title,
		userAgent,
		cookie,
		certificates,
		fetchViaProxy,
	)

	if entry.RawContent != "" {
		common.Logger.Info("crawler entry start to extract", zap.String("url", entry.URL))
		fullContent, pureContent, dateInArticle, imageUrlFromContent, title, templateAuthor, publishedAtTimestamp, mediaContent, mediaUrl, mediaType := processor.ArticleReadabilityExtractor(entry.RawContent, entry.URL, feedUrl, "", true)
		if strings.TrimSpace(entry.Title) == "" {
			entry.Title = title
		}
		entry.FullContent = fullContent
		entry.MediaContent = mediaContent
		entry.MediaUrl = mediaUrl
		entry.MediaType = mediaType
		if entry.ImageUrl == "" {
			entry.ImageUrl = imageUrlFromContent
		}
		if templateAuthor != "" {
			entry.Author = templateAuthor
		}
		if publishedAtTimestamp != 0 {
			entry.PublishedAt = publishedAtTimestamp
		} else {
			if dateInArticle != nil {
				entry.PublishedAt = (*dateInArticle).Unix()
			}
		}
		//if youtube feed don't fetch metadata
		if isMetaFromYtdlp(entry.URL) && (feedUrl == "" || !strings.Contains(entry.URL, "youtube.com")) {
			handleYtdlp(entry)
		}

		languageLen := len(pureContent)
		if languageLen > 100 {
			languageLen = 100
		}
		entry.Language = common.GetLanguage(pureContent[:languageLen])

		if entry.ImageUrl == "" && fullContent != "" {
			entry.ImageUrl = common.GetImageUrlFromContent(fullContent)
		}

	} else {
		common.Logger.Error("crawler raw content is null", zap.String("url", entry.URL))
	}
}

func notionFetchByApi(websiteURL string) string {
	client := notionClient.Client{}
	notionID := notionClient.ExtractNoDashIDFromNotionURL(websiteURL)
	common.Logger.Info("notion fetch", zap.String("id", notionID))
	if notionID != "" {
		page, _ := client.DownloadPage(notionID)
		bytes := tohtml.ToHTML(page)
		return string(bytes)
	}
	return ""
}

func wolaiFetchByApi(websiteURL string) string {
	pageID := wolaiapi.ExtractPageIDFromURL(websiteURL)
	common.Logger.Info("wolai fetch", zap.String("id", pageID))
	if pageID != "" {
		page := wolaiapi.FetchPage(pageID)
		return page
	}
	return ""
}

func FetchRawContnt(bflUser, websiteURL, title, userAgent string, cookie string, allowSelfSignedCertificates, useProxy bool) string {
	websiteURL = fetchUrlToChange(websiteURL)
	urlDomain := domain(websiteURL)
	common.Logger.Info("fatch raw contnet", zap.String("domain", websiteURL))
	if strings.Contains(urlDomain, "notion.site") {
		return notionFetchByApi(websiteURL)
	}
	if strings.Contains(urlDomain, "wolai.com") {
		return wolaiFetchByApi(websiteURL)
	}

	if strings.Contains(urlDomain, "quora.com") {
		return quora.QuoraByheadless(websiteURL)
	}
	if strings.Contains(urlDomain, "feishu.cn") {
		return feishu.FeishuByheadless(websiteURL)
	}
	if strings.Contains(urlDomain, "ximalaya.com") {
		return ximalaya.XimalayaByheadless(websiteURL)
	}
	if strings.Contains(urlDomain, "washingtonpost.com") {
		return washingtonpost.WashingtonpostByheadless(websiteURL)
	}
	//nytimes no success
	if strings.Contains(urlDomain, "nytimes.com") {
		return nytimes.NytimesByheadless(bflUser, websiteURL)
	}

	if strings.Contains(urlDomain, "wsj.com") {
		return wsj.WsjByheadless(websiteURL)
	}

	clt := client.NewClientWithConfig(websiteURL)
	clt.WithBflUser(bflUser)
	clt.WithUserAgent(userAgent)
	clt.WithCookie(cookie)
	if useProxy {
		clt.WithProxy()
	}
	clt.AllowSelfSignedCertificates = allowSelfSignedCertificates

	response, err := clt.Get()
	if err != nil {
		common.Logger.Error("crawling entry rawContent error ", zap.String("url", websiteURL), zap.Error(err))
		return ""
	}

	if response.HasServerFailure() {
		common.Logger.Error("crawling entry rawContent error ", zap.String("url", websiteURL))
		return ""
	}

	if !isAllowedContentType(response.ContentType) {
		common.Logger.Error("scraper: this resource is not a HTML document ", zap.String("url", websiteURL))
		return ""
	}

	if err = response.EnsureUnicodeBody(); err != nil {
		common.Logger.Error("scraper: this response check unicodeBody error ")
		return ""
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		common.Logger.Error("crawling entry rawContent error ", zap.String("url", websiteURL), zap.Error(err))
		return ""
	}
	common.Logger.Info("crawle raw content", zap.Int("length:", len(body)))
	return string(body)
}

func isMetaFromYtdlp(url string) bool {
	mediaList := []string{"bilibili.com", "youtube.com", "vimeo.com", "rumble.com"}
	for _, urlDomain := range mediaList {
		if strings.Contains(url, urlDomain) {
			return true
		}
	}

	return false
}

func isAllowedContentType(contentType string) bool {
	contentType = strings.ToLower(contentType)
	return strings.HasPrefix(contentType, "text/html") ||
		strings.HasPrefix(contentType, "application/xhtml+xml")
}

func domain(websiteURL string) string {
	parsedURL, err := url.Parse(websiteURL)
	if err != nil {
		return websiteURL
	}

	return parsedURL.Host
}

func fetchUrlToChange(websiteURL string) string {
	urlDomain := domain(websiteURL)
	switch urlDomain {
	case "web.okjike.com":
		parts := strings.Split(websiteURL, "/")
		id := parts[len(parts)-1]
		return "https://m.okjike.com/originalPosts/" + id
	default:
		return websiteURL
	}

}
