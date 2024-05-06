package crawler

import (
	"io"
	"strings"

	"bytetrade.io/web3os/backend-server/common"
	"bytetrade.io/web3os/backend-server/http/client"
	"bytetrade.io/web3os/backend-server/model"
	"github.com/PuerkitoBio/goquery"
	"github.com/beclab/article-extractor/processor"
	"go.uber.org/zap"
)

func EntryCrawler(entry *model.Entry, feedUrl, userAgent, cookie string, certificates, fetchViaProxy bool) {
	//entryID, entryUrl, entryTitle, imageUrl, author string, entryPublishedAt int64, feed *model.Feed) (string, string, int64) {

	entry.RawContent = fetchRawContnt(
		entry.URL,
		entry.Title,
		userAgent,
		cookie,
		certificates,
		fetchViaProxy,
	)

	if entry.RawContent != "" {
		if entry.Title == "" {
			entry.Title = extractTitleByHtml(entry.RawContent)
		}
		fullContent, pureContent, _, imageUrlFromContent, _, templateAuthor, _, publishedAtTimestamp := processor.ArticleReadabilityExtractor(entry.RawContent, entry.URL, feedUrl, "", true)

		entry.FullContent = fullContent
		if entry.ImageUrl == "" {
			entry.ImageUrl = imageUrlFromContent
		}
		if templateAuthor != "" {
			entry.Author = templateAuthor
		}
		/*if templateDate != nil {
			entry.PublishedAt = (*templateDate).Unix()
		}*/
		if publishedAtTimestamp != 0 {
			entry.PublishedAt = publishedAtTimestamp
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
	//return rawContent, rtContent, entryPublishedAt
}

func fetchRawContnt(websiteURL, title, userAgent string, cookie string, allowSelfSignedCertificates, useProxy bool) string {
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
