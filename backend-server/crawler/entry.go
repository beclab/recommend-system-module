package crawler

import (
	"fmt"
	"io"
	"log"
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
	"bytetrade.io/web3os/backend-server/crawler/ytdlp"
	"bytetrade.io/web3os/backend-server/http/client"
	"bytetrade.io/web3os/backend-server/knowledge"
	"bytetrade.io/web3os/backend-server/model"
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
	twitterEntry := knowledge.FetchTwitterContent(bflUser, twitterID, url)
	twitterEntry.ImageUrl = common.GetImageUrlFromContent(twitterEntry.FullContent)
	twitterEntry.Language = "en"
	return twitterEntry
}
func handleXHS(url string, bflUser string) *model.Entry {
	xshEntry := knowledge.FetchXHSContent(url, bflUser)
	xshEntry.ImageUrl = common.GetImageUrlFromContent(xshEntry.FullContent)
	xshEntry.Language = "zh-cn"
	return xshEntry
}
func handleBsky(url string, bflUser string) *model.Entry {
	bskyEntry := bskyapi.Fetch(bflUser, url)
	bskyEntry.ImageUrl = common.GetImageUrlFromContent(bskyEntry.FullContent)
	bskyEntry.Language = "en"
	return bskyEntry
}

func handleThreads(url string) *model.Entry {
	threadsEntry := threads.Fetch(url)
	threadsEntry.ImageUrl = common.GetImageUrlFromContent(threadsEntry.FullContent)
	threadsEntry.Language = "en"
	return threadsEntry
}

func handleQtfm(url string, bflUser string) *model.Entry {
	entry := ytdlp.Fetch(bflUser, url)
	if entry.Title != "" {
		entry.DownloadFileUrl = url
		entry.FileType = common.AudioFileType
	}
	return entry
}

func handleTBilibili(url string, bflUser string) *model.Entry {
	tbilibiliEntry := tbilibili.Fetch(bflUser, url)
	tbilibiliEntry.ImageUrl = common.GetImageUrlFromContent(tbilibiliEntry.FullContent)
	tbilibiliEntry.Language = "zh-cn"
	return tbilibiliEntry
}

func handleWeibo(url string, bflUser string) *model.Entry {
	weiboEntry := weibo.Fetch(bflUser, url)
	weiboEntry.ImageUrl = common.GetImageUrlFromContent(weiboEntry.FullContent)
	weiboEntry.Language = "zh-cn"
	return weiboEntry
}
func EntryCrawler(url string, bflUser string, feedID string) *model.Entry {
	primaryDomain := common.GetPrimaryDomain(url)
	urlDomain := common.Domain(url)
	common.Logger.Info("crawler entry start", zap.String("url", url), zap.String("primary domain:", primaryDomain))

	switch primaryDomain {
	case "bilibili.com":
		if urlDomain == "t.bilibili.com" {
			return handleTBilibili(url, bflUser)
		} else {
			return handleDefault(url, bflUser, feedID)
		}
	case "x.com":
		return handleX(url, bflUser)
	case "weibo.com":
		return handleWeibo(url, bflUser)
	case "xiaohongshu.com":
		return handleXHS(url, bflUser)
	case "bsky.app":
		return handleBsky(url, bflUser)
	case "threads.net":
		return handleThreads(url)
	case "qtfm.cn":
		return handleQtfm(url, bflUser)
	default:
		return handleDefault(url, bflUser, feedID)
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
	if ytdlpEntry.DownloadFileType != "" {
		entry.DownloadFileType = ytdlpEntry.DownloadFileType
	}
	if ytdlpEntry.DownloadFileUrl != "" {
		entry.DownloadFileUrl = ytdlpEntry.DownloadFileUrl
	}
	if ytdlpEntry.DownloadFileName != "" {
		entry.DownloadFileName = ytdlpEntry.DownloadFileName
	}

}

func setFileInfo(
	entry *model.Entry,
	url,
	extractFileType,
	extractFileUrl,
	extractFileName,
	contentTypeFileType,
	contentTypeFileName string,
) {
	if extractFileType == "" {
		extractFileUrl, extractFileName, extractFileType = processor.DownloadTypeQueryByUrl(url)
	}

	switch {
	case extractFileType != "":
		entry.FileType = extractFileType
		entry.DownloadFileUrl = extractFileUrl
		entry.DownloadFileType = extractFileType
		if extractFileName == "" {
			extractFileName = entry.Title
		}
		entry.DownloadFileName = extractFileName
	case contentTypeFileType != "":
		entry.FileType = contentTypeFileType
		entry.DownloadFileUrl = url
		entry.DownloadFileType = contentTypeFileType
		entry.DownloadFileName = contentTypeFileName
	default:
		entry.FileType = "article"
	}
}

func handleDefault(url string, bflUser string, feedID string) *model.Entry {
	entry := new(model.Entry)
	rawContent, fileTypeFromContentType, fileNameFromContentType := FetchRawContent(
		bflUser,
		url,
	)
	entry.RawContent = rawContent

	common.Logger.Info("crawler entry start to extract", zap.String("url", entry.URL))
	fullContent, pureContent, dateInArticle, imageUrlFromContent, title, templateAuthor, publishedAtTimestamp, mediaContent, downloadFileUrl, downloadFileType := processor.ArticleExtractor(entry.RawContent, url)
	entry.Title = common.FirstNonEmptyStr(entry.Title, title)
	entry.FullContent = fullContent
	entry.MediaContent = mediaContent
	entry.Author = common.FirstNonEmptyStr(entry.Author, templateAuthor)

	setFileInfo(
		entry,
		url,
		downloadFileType,
		downloadFileUrl,
		"",
		fileTypeFromContentType,
		fileNameFromContentType,
	)

	if feedID != "" {
		handleYtdlp(bflUser, entry)
	}

	entry.ImageUrl = common.FirstNonEmptyStr(
		entry.ImageUrl,
		imageUrlFromContent,
		common.GetImageUrlFromContent(fullContent),
	)

	if publishedAtTimestamp != 0 {
		entry.PublishedAt = publishedAtTimestamp
	} else {
		if dateInArticle != nil {
			entry.PublishedAt = (*dateInArticle).Unix()
		}
	}

	entry.Language = common.DetectLanguage(pureContent)
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

func FetchRawContent(bflUser, websiteURL string) (string, string, string) {
	url := fetchUrlToChange(websiteURL)
	urlDomain := common.Domain(url)
	common.Logger.Info("fatch raw contnet", zap.String("domain", websiteURL))
	switch {
	case strings.Contains(urlDomain, "notion.site"):
		return notionFetchByApi(url), "", ""
	case strings.Contains(urlDomain, "wolai.com"):
		return wolaiFetchByApi(url), "", ""
	case strings.Contains(urlDomain, "quora.com"):
		return quora.QuoraByheadless(url), "", ""
	case strings.Contains(urlDomain, "feishu.cn"):
		return feishu.FeishuByheadless(url), "", ""
	case strings.Contains(urlDomain, "ximalaya.com"):
		return ximalaya.XimalayaByheadless(url), "", ""
	case strings.Contains(urlDomain, "washingtonpost.com"):
		return washingtonpost.WashingtonpostByheadless(url), "", ""
	case strings.Contains(urlDomain, "nytimes.com"):
		return nytimes.NytimesByheadless(bflUser, url), "", ""
	case strings.Contains(urlDomain, "wsj.com"):
		return wsj.WsjByheadless(url), "", ""
	case strings.Contains(urlDomain, "youtube.com"):
		return "", "", ""
	default:
		return defaultFetchRawContent(url, bflUser)
	}
}

func defaultFetchRawContent(url string, bflUser string) (string, string, string) {
	clt := client.NewClientWithConfig(url)
	clt.WithBflUser(bflUser)
	response, err := clt.Get()
	if err != nil {
		common.Logger.Error("crawling entry rawContent error ", zap.String("url", url), zap.Error(err))
		return "", "", ""
	}
	if response.HasServerFailure() {
		common.Logger.Error("crawling entry rawContent error ", zap.String("url", url))
		return "", "", ""
	}

	fileType := determineFileType(response.ContentType)
	fileName := extractFileName(response.ContentDisposition)
	if fileType != "" && fileName == "" {
		fileName = getFileNameFromUrl(url, fileType)
	}
	if !isAllowedContentType(response.ContentType) {
		common.Logger.Error("scraper: this resource is not a HTML document ", zap.String("url", url))
		return "", fileType, fileName
	}
	if err = response.EnsureUnicodeBody(); err != nil {
		common.Logger.Error("scraper: this response check unicodeBody error ")
		return "", fileType, fileName
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		common.Logger.Error("crawling entry rawContent error ", zap.String("url", url), zap.Error(err))
		return "", fileType, fileName
	}
	common.Logger.Info("crawle raw content", zap.Int("length:", len(body)))
	return string(body), fileType, fileName
}

func determineFileType(reqContentType string) string {
	switch {
	case strings.HasPrefix(reqContentType, "text/html"):
		return ""
	case reqContentType == "application/pdf":
		return common.PdfFileType
	case reqContentType == "application/epub+zip":
		return common.EbookFileType
	case strings.HasPrefix(reqContentType, "audio/"):
		return common.AudioFileType
	case strings.HasPrefix(reqContentType, "video/"):
		return common.VideoFileType
	}
	return ""
}

func extractFileName(contentDisposition string) string {
	if contentDisposition != "" {
		log.Print("Content-Disposition:", contentDisposition)
		parts := strings.Split(contentDisposition, ";")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, "filename*=") {
				encodedPart := part[len("filename*="):]
				langAndEncoding := strings.SplitN(encodedPart, "'", 3)
				if len(langAndEncoding) == 3 {
					file, err := url.QueryUnescape(langAndEncoding[2])
					if err == nil {
						return file
					}
				}
			} else if strings.HasPrefix(part, "filename=") {
				return strings.Trim(part[len("filename="):], `"`)
			}
		}
	}
	return ""
}
func getFileNameFromUrl(url string, fileType string) string {
	lastSlashIndex := strings.LastIndex(url, "/")
	fileName := url[lastSlashIndex+1:]
	if fileType == common.EbookFileType && !strings.HasSuffix(fileName, ".epub") {
		fileName = fileName + ".epub"
	}
	if fileType == common.PdfFileType && !strings.HasSuffix(fileName, ".pdf") {
		fileName = fileName + ".pdf"
	}
	return fileName
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
