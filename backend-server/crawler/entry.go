package crawler

import (
	"fmt"
	"io"
	"log"
	"net/url"
	"path/filepath"
	"regexp"
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
	if twitterEntry != nil {
		twitterEntry.ImageUrl = common.GetImageUrlFromContent(twitterEntry.FullContent)
		twitterEntry.Language = "en"
		twitterEntry.FileType = common.ArticleFileType
	}
	return twitterEntry
}
func handleXHS(url string, bflUser string) *model.Entry {
	xshEntry := knowledge.FetchXHSContent(url, bflUser)
	if xshEntry != nil {
		xshEntry.ImageUrl = common.GetImageUrlFromContent(xshEntry.FullContent)
		xshEntry.Language = "zh-cn"
		xshEntry.FileType = common.ArticleFileType
	}
	return xshEntry
}
func handleBsky(url string, bflUser string) *model.Entry {
	bskyEntry := bskyapi.Fetch(bflUser, url)
	if bskyEntry != nil {
		bskyEntry.ImageUrl = common.GetImageUrlFromContent(bskyEntry.FullContent)
		bskyEntry.Language = "en"
		bskyEntry.FileType = common.ArticleFileType
	}
	return bskyEntry
}

func handleThreads(url string) *model.Entry {
	threadsEntry := threads.Fetch(url)
	if threadsEntry != nil {
		threadsEntry.ImageUrl = common.GetImageUrlFromContent(threadsEntry.FullContent)
		threadsEntry.Language = "en"
		threadsEntry.FileType = common.ArticleFileType
	}
	return threadsEntry
}

func handleQtfm(url string, bflUser string) *model.Entry {
	entry := ytdlp.Fetch(bflUser, url)
	if entry != nil && entry.Title != "" {
		entry.DownloadFileUrl = url
		entry.FileType = common.AudioFileType
	}
	return entry
}

func handleTBilibili(url string, bflUser string) *model.Entry {
	tbilibiliEntry := tbilibili.Fetch(bflUser, url)
	if tbilibiliEntry != nil {
		tbilibiliEntry.ImageUrl = common.GetImageUrlFromContent(tbilibiliEntry.FullContent)
		tbilibiliEntry.Language = "zh-cn"
		tbilibiliEntry.FileType = common.ArticleFileType
	}
	return tbilibiliEntry
}

func handleWeibo(url string, bflUser string) *model.Entry {
	weiboEntry := weibo.Fetch(bflUser, url)
	if weiboEntry != nil {
		weiboEntry.ImageUrl = common.GetImageUrlFromContent(weiboEntry.FullContent)
		weiboEntry.Language = "zh-cn"
		weiboEntry.FileType = common.ArticleFileType
	}
	return weiboEntry
}
func EntryCrawler(url string, bflUser string, feedID string) *model.Entry {
	primaryDomain := common.GetPrimaryDomain(url)
	urlDomain := common.Domain(url)

	opusPattern := `bilibili\.com/opus`
	bilibiliOpusRe := regexp.MustCompile(opusPattern)
	common.Logger.Info("crawler entry start", zap.String("url", url), zap.String("primary domain:", primaryDomain))

	var entry *model.Entry
	switch primaryDomain {
	case "bilibili.com":
		if urlDomain == "t.bilibili.com" || bilibiliOpusRe.MatchString(url) {
			entry = handleTBilibili(url, bflUser)
		} else {
			entry = handleDefault(url, bflUser)
		}
	case "x.com":
		entry = handleX(url, bflUser)
	case "weibo.com":
		entry = handleWeibo(url, bflUser)
	case "xiaohongshu.com":
		entry = handleXHS(url, bflUser)
	case "bsky.app":
		entry = handleBsky(url, bflUser)
	case "threads.net":
		entry = handleThreads(url)
	case "qtfm.cn":
		entry = handleQtfm(url, bflUser)
	default:
		entry = handleDefault(url, bflUser)
	}
	if entry == nil {
		entry = &model.Entry{}
		entry.FileType = common.ArticleFileType
	}
	if feedID != "" {
		entry.FileType = common.ArticleFileType
	} else {
		if entry.FileType == common.VideoFileType || entry.FileType == common.ArticleFileType || entry.FileType == "" {
			handleYtdlp(bflUser, url, entry)
		}

	}
	return entry
}

func handleYtdlp(bflUser string, url string, entry *model.Entry) {
	if ytdlpEntry := ytdlp.Fetch(bflUser, url); ytdlpEntry != nil {
		updateIfNotEmpty := func(dst *string, src string) {
			if src != "" {
				*dst = src
			}
		}

		updateIfNotEmpty(&entry.Author, ytdlpEntry.Author)
		updateIfNotEmpty(&entry.Title, ytdlpEntry.Title)
		updateIfNotEmpty(&entry.FullContent, ytdlpEntry.FullContent)
		updateIfNotEmpty(&entry.DownloadFileType, ytdlpEntry.DownloadFileType)
		updateIfNotEmpty(&entry.DownloadFileName, ytdlpEntry.DownloadFileName)
		if ytdlpEntry.DownloadFileType != "" {
			entry.DownloadFileUrl = url
		}
		if ytdlpEntry.PublishedAt != 0 {
			entry.PublishedAt = ytdlpEntry.PublishedAt
		}
		common.Logger.Info("yt-dlp fetch", zap.String("download url:", entry.DownloadFileUrl), zap.String("download filetype:", ytdlpEntry.DownloadFileType))
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
		entry.FileType = common.ArticleFileType
	}
	common.Logger.Info("set file type", zap.String("extractFileType", extractFileType), zap.String("contentTypeFileType:", contentTypeFileType), zap.String("final file type", entry.FileType))
}

func handleDefault(url string, bflUser string) *model.Entry {
	entry := new(model.Entry)
	rawContent, fileTypeFromContentType, fileNameFromContentType := FetchRawContent(
		bflUser,
		url,
	)
	entry.URL = url
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
		return "", common.VideoFileType, ""
	default:
		return defaultFetchRawContent(url, bflUser)
	}
}

func nonHtmlExtract(url string, bflUser string) (string, string) {
	clt := client.NewClientWithConfig(url)
	clt.WithBflUser(bflUser)
	response, err := clt.Head()
	if err != nil {
		common.Logger.Error("non html extract error ", zap.String("url", url), zap.Error(err))
		return "", ""
	}
	fileType := determineFileType(response.ContentType)
	fileName := extractFileName(response.ContentDisposition)
	if fileType != "" && fileName == "" {
		fileName = GetFileNameFromUrl(url, response.ContentType)
	}
	return fileType, fileName
}
func defaultFetchRawContent(url string, bflUser string) (string, string, string) {
	fileType, fileName := nonHtmlExtract(url, bflUser)
	if fileType != "" {
		return "", fileType, fileName
	}

	common.Logger.Info("fetch raw content ", zap.String("url", url))
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
	case strings.HasPrefix(reqContentType, "text/html") || strings.HasPrefix(reqContentType, "application/xhtml+xml"):
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
	return reqContentType
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
func GetFileNameFromUrl(urlString string, contentType string) string {
	u, err := url.Parse(urlString)
	if err != nil {
		return "download"
	}

	path := u.Path
	filename := filepath.Base(path)
	if filename == "." || filename == "/" {
		filename = "index"
	}

	if idx := strings.Index(filename, "?"); idx != -1 {
		filename = filename[:idx]
	}

	filename = strings.ReplaceAll(filename, "%20", "_")
	filename = strings.ReplaceAll(filename, "+", "_")
	filename = strings.ReplaceAll(filename, " ", "_")

	base := filename
	ext := filepath.Ext(filename)

	if ext != "" {
		return filename
	}

	switch {
	case strings.HasPrefix(contentType, "image/jpeg"):
		return base + ".jpg"
	case strings.HasPrefix(contentType, "image/png"):
		return base + ".png"
	case strings.HasPrefix(contentType, "image/gif"):
		return base + ".gif"
	case strings.HasPrefix(contentType, "image/svg+xml"):
		return base + ".svg"
	case strings.HasPrefix(contentType, "image/webp"):
		return base + ".webp"
	case strings.HasPrefix(contentType, "application/pdf"):
		return base + ".pdf"
	case strings.HasPrefix(contentType, "application/epub+zip"):
		return base + ".epub"
	case strings.HasPrefix(contentType, "application/json"):
		return base + ".json"
	case strings.HasPrefix(contentType, "text/csv"):
		return base + ".csv"
	case strings.HasPrefix(contentType, "text/plain"):
		return base + ".txt"
	case strings.HasPrefix(contentType, "application/zip"):
		return base + ".zip"
	case strings.HasPrefix(contentType, "application/vnd.ms-excel"):
		return base + ".xls"
	case strings.HasPrefix(contentType, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"):
		return base + ".xlsx"
	case strings.HasPrefix(contentType, "application/msword"):
		return base + ".doc"
	case strings.HasPrefix(contentType, "application/vnd.openxmlformats-officedocument.wordprocessingml.document"):
		return base + ".docx"
	case strings.HasPrefix(contentType, "audio/mpeg"):
		return base + ".mp3"
	case strings.HasPrefix(contentType, "video/mp4"):
		return base + ".mp4"
	case strings.HasPrefix(contentType, "application/octet-stream"):
		return base + ".bin"
	default:
		return base + ".download"
	}
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
