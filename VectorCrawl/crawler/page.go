package crawler

import (
	"fmt"
	"io"
	"strings"

	"bytetrade.io/web3os/vector-crawl/common"
	"bytetrade.io/web3os/vector-crawl/crawler/bskyapi"
	"bytetrade.io/web3os/vector-crawl/crawler/feishu"
	notionClient "bytetrade.io/web3os/vector-crawl/crawler/notionapi"
	"bytetrade.io/web3os/vector-crawl/crawler/notionapi/tohtml"
	"bytetrade.io/web3os/vector-crawl/crawler/nytimes"
	"bytetrade.io/web3os/vector-crawl/crawler/quora"
	"bytetrade.io/web3os/vector-crawl/crawler/tbilibili"
	"bytetrade.io/web3os/vector-crawl/crawler/threads"
	"bytetrade.io/web3os/vector-crawl/crawler/twitter"
	"bytetrade.io/web3os/vector-crawl/crawler/washingtonpost"
	"bytetrade.io/web3os/vector-crawl/crawler/weibo"
	wolaiapi "bytetrade.io/web3os/vector-crawl/crawler/wolaiapi"
	"bytetrade.io/web3os/vector-crawl/crawler/wsj"
	"bytetrade.io/web3os/vector-crawl/crawler/ximalaya"
	"bytetrade.io/web3os/vector-crawl/crawler/ytdlp"
	"bytetrade.io/web3os/vector-crawl/http/client"
	"bytetrade.io/web3os/vector-crawl/model"
	"github.com/beclab/article-extractor/processor"
	"go.uber.org/zap"
)

func handleX(url string, bflUser string) *model.Entry {
	twitterID := ""
	parts := strings.Split(url, "status/")
	if len(parts) > 1 {
		twitterID = strings.TrimSpace(parts[1])
	}
	fmt.Println("twitter ID:", twitterID)
	twitterEntry := twitter.Fetch(bflUser, twitterID, url)
	return twitterEntry
}

func handleBsky(url string, bflUser string) *model.Entry {
	bskyEntry := bskyapi.Fetch(bflUser, url)
	bskyEntry.ImageUrl = common.GetImageUrlFromContent(bskyEntry.FullContent)
	return bskyEntry
}

func handleThreads(url string) *model.Entry {
	threadsEntry := threads.Fetch(url)
	threadsEntry.Language = "en"
	threadsEntry.ImageUrl = common.GetImageUrlFromContent(threadsEntry.FullContent)
	return threadsEntry

}

func handleQtfm(url string, bflUser string) *model.Entry {
	entry := ytdlp.Fetch(bflUser, url)
	if entry.Title != "" {
		entry.DownloadFileUrl = url
		entry.DownloadFileType = "audio"
	}
	return entry
}

func handleTBilibili(url string, bflUser string) *model.Entry {
	tbilibiliEntry := tbilibili.Fetch(bflUser, url)
	tbilibiliEntry.Language = "zh-cn"
	tbilibiliEntry.ImageUrl = common.GetImageUrlFromContent(tbilibiliEntry.FullContent)
	return tbilibiliEntry
}

func handleWeibo(url string, bflUser string) *model.Entry {
	weiboEntry := weibo.Fetch(bflUser, url)
	weiboEntry.Language = "zh-cn"
	weiboEntry.ImageUrl = common.GetImageUrlFromContent(weiboEntry.FullContent)
	return weiboEntry
}

func PageCrawler(url string, bflUser string) *model.Entry {
	primaryDomain := common.GetPrimaryDomain(url)
	urlDomain := common.Domain(url)
	common.Logger.Info("crawler entry start", zap.String("url", url), zap.String("primary domain:", primaryDomain))

	switch primaryDomain {
	case "bilibili.com":
		if urlDomain == "t.bilibili.com" {
			return handleTBilibili(url, bflUser)
		} else {
			return handleDefault(url, bflUser)
		}
	case "x.com":
		return handleX(url, bflUser)
	case "weibo.com":
		return handleWeibo(url, bflUser)
	case "bsky.app":
		return handleBsky(url, bflUser)
	case "threads.net":
		return handleThreads(url)
	case "qtfm.cn":
		return handleQtfm(url, bflUser)
	default:
		return handleDefault(url, bflUser)
	}
}

func handleYtdlp(bflUser string, entry *model.Entry) {
	ytdlpEntry := ytdlp.Fetch(bflUser, entry.URL)
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

func handleDefault(url string, bflUser string) *model.Entry {
	entry := new(model.Entry)
	entry.RawContent = FetchRawContnt(
		url,
		bflUser,
	)

	common.Logger.Info("crawler entry start to extract", zap.String("url", url))
	fullContent, pureContent, dateInArticle, imageUrlFromContent, title, templateAuthor, publishedAtTimestamp, mediaContent, downloadFileUrl, downloadFileType := processor.ArticleReadabilityExtractor(entry.RawContent, url, "", "", true)
	if strings.TrimSpace(entry.Title) == "" {
		entry.Title = title
	}
	entry.FullContent = fullContent
	entry.MediaContent = mediaContent
	entry.DownloadFileUrl = downloadFileUrl
	entry.DownloadFileType = downloadFileType
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
	if ytdlp.IsMetaFromYtdlp(url) {
		handleYtdlp(bflUser, entry)
	}

	languageLen := len(pureContent)
	if languageLen > 100 {
		languageLen = 100
	}
	entry.Language = common.GetLanguage(pureContent[:languageLen])

	if entry.ImageUrl == "" && fullContent != "" {
		entry.ImageUrl = common.GetImageUrlFromContent(fullContent)
	}

	return entry
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

func FetchRawContnt(url string, bflUser string) string {
	url = fetchUrlToChange(url)
	urlDomain := common.Domain(url)
	common.Logger.Info("fatch raw contnet", zap.String("domain", url))
	if strings.Contains(urlDomain, "notion.site") {
		return notionFetchByApi(url)
	}
	if strings.Contains(urlDomain, "wolai.com") {
		return wolaiFetchByApi(url)
	}

	if strings.Contains(urlDomain, "quora.com") {
		return quora.QuoraByheadless(url)
	}
	if strings.Contains(urlDomain, "feishu.cn") {
		return feishu.FeishuByheadless(url)
	}
	if strings.Contains(urlDomain, "ximalaya.com") {
		return ximalaya.XimalayaByheadless(url)
	}
	if strings.Contains(urlDomain, "washingtonpost.com") {
		return washingtonpost.WashingtonpostByheadless(url)
	}
	//nytimes no success
	if strings.Contains(urlDomain, "nytimes.com") {
		return nytimes.NytimesByheadless(bflUser, url)
	}

	if strings.Contains(urlDomain, "wsj.com") {
		return wsj.WsjByheadless(url)
	}
	if strings.Contains(urlDomain, "youtube.com") {
		return ""
	}

	clt := client.NewClientWithConfig(url)
	clt.WithBflUser(bflUser)
	/*clt.WithUserAgent(userAgent)
	clt.WithCookie(cookie)
	if useProxy {
		clt.WithProxy()
	}
	clt.AllowSelfSignedCertificates = allowSelfSignedCertificates*/

	response, err := clt.Get()
	if err != nil {
		common.Logger.Error("crawling entry rawContent error ", zap.String("url", url), zap.Error(err))
		return ""
	}

	if response.HasServerFailure() {
		common.Logger.Error("crawling entry rawContent error ", zap.String("url", url))
		return ""
	}

	if !isAllowedContentType(response.ContentType) {
		common.Logger.Error("scraper: this resource is not a HTML document ", zap.String("url", url))
		return ""
	}

	if err = response.EnsureUnicodeBody(); err != nil {
		common.Logger.Error("scraper: this response check unicodeBody error ")
		return ""
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		common.Logger.Error("crawling entry rawContent error ", zap.String("url", url), zap.Error(err))
		return ""
	}
	common.Logger.Info("crawle raw content", zap.Int("length:", len(body)))
	return string(body)
}

func isAllowedContentType(contentType string) bool {
	contentType = strings.ToLower(contentType)
	return strings.HasPrefix(contentType, "text/html") ||
		strings.HasPrefix(contentType, "application/xhtml+xml")
}

func fetchUrlToChange(websiteURL string) string {
	urlDomain := common.Domain(websiteURL)
	switch urlDomain {
	case "web.okjike.com":
		parts := strings.Split(websiteURL, "/")
		id := parts[len(parts)-1]
		return "https://m.okjike.com/originalPosts/" + id
	default:
		return websiteURL
	}

}
