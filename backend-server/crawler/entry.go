package crawler

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"bytetrade.io/web3os/backend-server/common"
	"bytetrade.io/web3os/backend-server/crawler/bskyapi"
	"bytetrade.io/web3os/backend-server/crawler/feishu"
	notionClient "bytetrade.io/web3os/backend-server/crawler/notionapi"
	"bytetrade.io/web3os/backend-server/crawler/notionapi/tohtml"
	"bytetrade.io/web3os/backend-server/crawler/quora"
	"bytetrade.io/web3os/backend-server/crawler/threads"
	wolaiapi "bytetrade.io/web3os/backend-server/crawler/wolaiapi"
	"bytetrade.io/web3os/backend-server/http/client"
	"bytetrade.io/web3os/backend-server/knowledge"
	"bytetrade.io/web3os/backend-server/model"
	"github.com/PuerkitoBio/goquery"
	"github.com/beclab/article-extractor/processor"
	"go.uber.org/zap"
)

func writeFullContent(content string) {
	file, err := os.Create("content.html")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()
	file.WriteString(content)
}

func EntryCrawler(entry *model.Entry, feedUrl, userAgent, cookie string, certificates, fetchViaProxy bool) {
	//entryID, entryUrl, entryTitle, imageUrl, author string, entryPublishedAt int64, feed *model.Feed) (string, string, int64) {
	primaryDomain := common.GetPrimaryDomain(entry.URL)
	common.Logger.Info("crawler entry start", zap.String("url", entry.URL), zap.String("primary domain:", primaryDomain))
	if primaryDomain == "bilibili.com" {
		entry.FullContent = entry.Content
		entry.Language = "zh-cn"

		domain := common.Domain(entry.URL)
		if domain == "t.bilibili.com" {
			if entry.ImageUrl == "" && entry.Content != "" {
				entry.ImageUrl = common.GetImageUrlFromContent(entry.Content)
			}
			return
		}
	}
	if primaryDomain == "x.com" {
		twitterID := ""
		parts := strings.Split(entry.URL, "status/")
		if len(parts) > 1 {
			twitterID = strings.TrimSpace(parts[1])
		}
		fmt.Println("twitter ID:", twitterID)
		twitterEntry := knowledge.FetchTwitterContent(entry.BflUser, twitterID, entry.URL)
		if twitterEntry != nil {
			entry.FullContent = twitterEntry.FullContent
			entry.MediaContent = twitterEntry.MediaContent
			entry.MediaUrl = twitterEntry.MediaUrl
			entry.MediaType = twitterEntry.MediaType
			entry.Author = twitterEntry.Author
			entry.Title = twitterEntry.Title
			entry.PublishedAt = twitterEntry.PublishedAt
			entry.ImageUrl = common.GetImageUrlFromContent(entry.FullContent)
			entry.Language = "en"
		}
		return
	}

	if primaryDomain == "xiaohongshu.com" {
		xshEntry := knowledge.FetchXHSContent(entry.URL)
		if xshEntry != nil {
			entry.FullContent = xshEntry.FullContent
			entry.MediaContent = xshEntry.MediaContent
			entry.MediaUrl = xshEntry.MediaUrl
			entry.MediaType = xshEntry.MediaType
			entry.Author = xshEntry.Author
			entry.Title = xshEntry.Title
			entry.PublishedAt = xshEntry.PublishedAt
			entry.ImageUrl = common.GetImageUrlFromContent(entry.FullContent)
			entry.Language = "zh-cn"
		}
		return
	}

	if primaryDomain == "bsky.app" {
		bskyEntry := bskyapi.Fetch(entry.BflUser, entry.URL)
		if bskyEntry != nil {
			entry.FullContent = bskyEntry.FullContent
			entry.Author = bskyEntry.Author
			entry.Title = bskyEntry.Title
			entry.PublishedAt = bskyEntry.PublishedAt
			entry.ImageUrl = common.GetImageUrlFromContent(entry.FullContent)
			entry.Language = "en"
		}
		return
	}

	if primaryDomain == "threads.net" {
		threadsEntry := threads.Fetch(entry.URL)
		if threadsEntry != nil {
			entry.FullContent = threadsEntry.FullContent
			entry.Author = threadsEntry.Author
			entry.Title = threadsEntry.Title
			entry.PublishedAt = threadsEntry.PublishedAt
			entry.ImageUrl = common.GetImageUrlFromContent(entry.FullContent)
			entry.Language = "en"
		}
		return
	}
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
		if isMetaFromYtdlp(entry.URL) {
			metaEntry := knowledge.LoadMetaFromYtdlp(entry.BflUser, entry.URL)
			if metaEntry != nil {
				if metaEntry.Author != "" {
					entry.Author = metaEntry.Author
				}
				if metaEntry.Title != "" {
					entry.Title = metaEntry.Title
				}
				if metaEntry.PublishedAt != 0 {
					entry.PublishedAt = metaEntry.PublishedAt
				}
				if metaEntry.FullContent != "" {
					entry.FullContent = metaEntry.FullContent
				}

			}
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
	common.Logger.Info("crawler entry finished", zap.String("url", entry.URL))
	//return rawContent, rtContent, entryPublishedAt
}

/*func notionFetchByheadless(websiteURL string) string {
	var allocCtx context.Context
	var cancelCtx context.CancelFunc
	allocOpts := chromedp.DefaultExecAllocatorOptions[:]

	allocOpts = append(allocOpts,
		chromedp.DisableGPU,
		chromedp.Flag("blink-settings", "imagesEnabled=false"),
		chromedp.Flag("no-first-run", true),
		//chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.UserAgent(`Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.55 Safari/537.36`),
		//chromedp.Flag("accept-language", `zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7,zh-TW;q=0.6`),
	)

	headlessSer := os.Getenv("HEADLESS_SERVER_URL")
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	if headlessSer != "" {
		c, cancelAlloc := chromedp.NewRemoteAllocator(ctx, headlessSer)
		defer cancelAlloc()
		allocCtx, cancelCtx = chromedp.NewContext(c)
	} else {
		c, cancelAlloc := chromedp.NewExecAllocator(ctx, allocOpts...)
		defer cancelAlloc()

		allocCtx, cancelCtx = chromedp.NewContext(c)
	}
	//ctx, cancel := chromedp.NewContext(context.Background())
	defer cancelCtx()
	htmlContent := ""
	common.Logger.Info("notion headless fetch 1 ")
	err := chromedp.Run(allocCtx,
		chromedp.Navigate(websiteURL),
		//chromedp.WaitVisible(`.notion-page-content`),
		chromedp.OuterHTML("html", &htmlContent),
	)
	if err != nil {
		common.Logger.Error("notion headless fetch error", zap.String("url", websiteURL), zap.Error(err))
	}
	common.Logger.Info("notion headless fetch end...", zap.Int("content len", len(htmlContent)))
	return htmlContent
}*/

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

func extractTitleByHtml(content string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	if err != nil {
		return ""
	}
	return doc.Find("title").Text()
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
