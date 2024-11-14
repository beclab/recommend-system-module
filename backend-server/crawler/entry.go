package crawler

import (
	"context"
	"io"
	"net/url"
	"os"
	"strings"
	"time"

	"bytetrade.io/web3os/backend-server/common"
	"bytetrade.io/web3os/backend-server/http/client"
	"bytetrade.io/web3os/backend-server/knowledge"
	"bytetrade.io/web3os/backend-server/model"
	"github.com/PuerkitoBio/goquery"
	"github.com/beclab/article-extractor/processor"
	"github.com/chromedp/chromedp"
	"go.uber.org/zap"
)

func EntryCrawler(entry *model.Entry, feedUrl, userAgent, cookie string, certificates, fetchViaProxy bool) {
	//entryID, entryUrl, entryTitle, imageUrl, author string, entryPublishedAt int64, feed *model.Feed) (string, string, int64) {
	common.Logger.Info("crawler entry start", zap.String("url", entry.URL))
	entry.RawContent = FetchRawContnt(
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
			metaEntry := knowledge.LoadMetaFromYtdlp(entry.URL)
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
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(fullContent))
			if err == nil {
				doc.Find("img").Each(func(i int, s *goquery.Selection) {
					img, _ := s.Attr("src")
					if strings.HasPrefix(img, "http") {
						entry.ImageUrl = img
					}
				})
			}
		}

	} else {
		common.Logger.Error("crawler raw content is null", zap.String("url", entry.URL))
	}
	common.Logger.Info("crawler entry finished", zap.String("url", entry.URL))
	//return rawContent, rtContent, entryPublishedAt
}

func notionFetchByheadless(websiteURL string) string {
	var allocCtx context.Context
	var cancelCtx context.CancelFunc
	allocOpts := chromedp.DefaultExecAllocatorOptions[:]
	allocOpts = append(allocOpts,
		chromedp.DisableGPU,
		chromedp.Flag("blink-settings", "imagesEnabled=false"),
		chromedp.UserAgent(`Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.55 Safari/537.36`),
		//chromedp.Flag("accept-language", `zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7,zh-TW;q=0.6`),
	)
	headlessSer := os.Getenv("HEADLESS_SERVER_URL")
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
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
		chromedp.WaitVisible(`.notion-page-content`),
		chromedp.OuterHTML("html", &htmlContent),
	)
	if err != nil {
		common.Logger.Error("notion headless fetch error", zap.String("url", websiteURL), zap.Error(err))
	}
	common.Logger.Info("notion headless fetch end...", zap.Int("content len", len(htmlContent)))
	return htmlContent
}

func FetchRawContnt(websiteURL, title, userAgent string, cookie string, allowSelfSignedCertificates, useProxy bool) string {
	urlDomain := domain(websiteURL)
	common.Logger.Info("fatch raw contnet", zap.String("domain", websiteURL))
	if strings.Contains(urlDomain, "notion.site") {
		return notionFetchByheadless(websiteURL)
	}

	clt := client.NewClientWithConfig(websiteURL)
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
