package api

import (
	encodeJson "encoding/json"
	"net/http"
	"strings"

	"bytetrade.io/web3os/backend-server/common"
	"bytetrade.io/web3os/backend-server/crawler"
	"bytetrade.io/web3os/backend-server/http/client"
	"bytetrade.io/web3os/backend-server/http/request"
	"bytetrade.io/web3os/backend-server/http/response/json"
	"bytetrade.io/web3os/backend-server/knowledge"
	"bytetrade.io/web3os/backend-server/model"
	"github.com/beclab/article-extractor/processor"
	"go.uber.org/zap"
)

func (h *handler) fetchContent(w http.ResponseWriter, r *http.Request) {
	entryID := request.RouteStringParam(r, "entryID")

	entry, err := h.store.GetEntryById(entryID)
	if err != nil {
		common.Logger.Error("load entry error", zap.String("entryID", entryID), zap.Error(err))
	}
	if entry == nil {
		common.Logger.Error("load entry error entry is nil", zap.String("feedId", entryID))
		json.OK(w, r, "")
		return
	}
	if strings.TrimSpace(entry.FullContent) == "" {
		entry.FullContent = h.newFetchContent(entry)
	}
	json.OK(w, r, entry.FullContent)

}

func (h *handler) newFetchContent(entry *model.Entry) string {
	var feed *model.Feed
	if entry.FeedID != nil {
		feed, _ = h.store.GetFeedById(*entry.FeedID)
	}
	//crawler.EntryCrawler(entry, feed) //entry.ID.Hex(), entry.URL, entry.Title, entry.ImageUrl, entry.Author, entry.PublishedAt, feed)

	feedUrl := ""
	userAgent := ""
	cookie := ""
	certificates := false
	fetchViaProxy := false

	if feed != nil {

		feedUrl = feed.FeedURL
		userAgent = feed.UserAgent
		cookie = feed.Cookie
		certificates = feed.AllowSelfSignedCertificates
		fetchViaProxy = feed.FetchViaProxy

	}
	crawler.EntryCrawler(entry, feedUrl, userAgent, cookie, certificates, fetchViaProxy)

	updateEntry := &model.Entry{ID: entry.ID, URL: entry.URL, ImageUrl: entry.ImageUrl, PublishedAt: entry.PublishedAt, Title: entry.Title, Language: entry.Language, Author: entry.Author, RawContent: entry.RawContent, FullContent: entry.FullContent}
	//h.store.UpdateEntryContent(updateDocIDEntry)
	if entry.MediaContent != "" || entry.MediaUrl != "" {
		updateEntry.Attachment = true
	}
	knowledge.UpdateLibraryEntryContent(entry.BflUser, updateEntry, false)
	if entry.MediaContent != "" || entry.MediaUrl != "" {
		knowledge.NewEnclosure(entry, nil, h.store)
	}

	return entry.FullContent
}

func (h *handler) knowledgeFetchContent(w http.ResponseWriter, r *http.Request) {
	entryID := request.RouteStringParam(r, "entryID")

	common.Logger.Info("knowledge fetch  entry content", zap.String("entryID", entryID))
	entry, err := h.store.GetEntryById(entryID)
	if err != nil {
		common.Logger.Error("load entry error", zap.String("entryID", entryID), zap.Error(err))
	}
	if entry == nil {
		common.Logger.Error("load entry error entry is nil", zap.String("feedId", entryID))
		json.OK(w, r, "")
		return
	}
	if strings.TrimSpace(entry.FullContent) == "" {
		go func() {
			h.newFetchContent(entry)
		}()
	}
	json.NoContent(w, r)

}

func (h *handler) exceptYTdlpDownloadQuery(w http.ResponseWriter, r *http.Request) {
	url := request.QueryStringParam(r, "url", "")
	bflUser := request.QueryStringParam(r, "bfl_user", "")
	common.Logger.Info("knowledge download file query", zap.String("url", url), zap.String("bfl_user", bflUser))

	downloadUrl, donwloadFile, donwloadFileType := processor.NonRawContentDownloadQueryInArticle(url)
	if donwloadFileType != "" {
		json.OK(w, r, model.DownloadFetchResponseModel{
			Code: 0,
			Data: model.DownloadFetchReqModel{
				DownloadUrl: downloadUrl,
				FileType:    donwloadFileType,
				FileName:    donwloadFile,
			},
		})
		return
	}

	useAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	rawContent := crawler.FetchRawContnt(
		bflUser,
		url,
		"",
		useAgent,
		"",
		false,
		false,
	)
	parseUrl, _ := processor.ExceptYTdlpDownloadQueryInArticle(rawContent, url)
	if url != parseUrl && parseUrl != "" {
		url = parseUrl
	}
	urlType, fileName := client.GetContentAndisposition(url, bflUser)
	if urlType != "" && urlType != "text/html" {
		if fileName == "" {
			fileName = client.GetDownloadFile(url, bflUser, urlType)
		}
		json.OK(w, r, model.DownloadFetchResponseModel{
			Code: 0,
			Data: model.DownloadFetchReqModel{
				DownloadUrl: url,
				FileType:    urlType,
				FileName:    fileName,
			},
		})
		return
	}
	json.OK(w, r, model.DownloadFetchResponseModel{
		Code: 0,
		Data: model.DownloadFetchReqModel{},
	})
}

/*func (h *handler) FetchMetaData(w http.ResponseWriter, r *http.Request) {
	url := request.QueryStringParam(r, "url", "")
	useAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	rawContent := crawler.FetchRawContnt(
		"",
		url,
		"",
		useAgent,
		"",
		false,
		false,
	)
	fullContent, _, dateInArticle, imageUrlFromContent, title, templateAuthor, publishedAtTimestamp, _, _, _ := processor.ArticleReadabilityExtractor(rawContent, url, url, "", true)
	if publishedAtTimestamp == 0 && dateInArticle != nil {
		publishedAtTimestamp = (*dateInArticle).Unix()
	}
	entry := model.Entry{FullContent: fullContent, Title: title, Author: templateAuthor, PublishedAt: publishedAtTimestamp, ImageUrl: imageUrlFromContent}

	json.OK(w, r, model.EntryFetchResponseModel{Code: 0, Data: entry})

	json.NoContent(w, r)

}*/

func (h *handler) knowledgeVideoFetchContent(w http.ResponseWriter, r *http.Request) {
	entryID := request.RouteStringParam(r, "entryID")

	var reqObj model.DownloadFetchReqModel

	err := encodeJson.NewDecoder(r.Body).Decode(&reqObj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	common.Logger.Info("knowledge fetch  entry content", zap.Any("obj", reqObj))
	entry, err := h.store.GetEntryById(entryID)
	if err != nil {
		common.Logger.Error("load entry error", zap.String("entryID", entryID), zap.Error(err))
	}
	if entry == nil {
		common.Logger.Error("load entry error entry is nil", zap.String("feedId", entryID))
		json.OK(w, r, "")
		return
	}
	go func() {
		h.newVideoFetchContent(entry, reqObj.DownloadUrl, reqObj.FileName, reqObj.FileType, reqObj.LarepassId, reqObj.Folder)
	}()
	json.NoContent(w, r)

}
func (h *handler) newVideoFetchContent(entry *model.Entry, downloadUrl string, fileName string, fileType string, larepassId string, folder string) string {
	var feed *model.Feed
	if entry.FeedID != nil {
		feed, _ = h.store.GetFeedById(*entry.FeedID)
	}

	feedUrl := ""
	userAgent := ""
	cookie := ""
	certificates := false
	fetchViaProxy := false

	if feed != nil {
		feedUrl = feed.FeedURL
		userAgent = feed.UserAgent
		cookie = feed.Cookie
		certificates = feed.AllowSelfSignedCertificates
		fetchViaProxy = feed.FetchViaProxy
	}
	crawler.EntryCrawler(entry, feedUrl, userAgent, cookie, certificates, fetchViaProxy)

	updateEntry := &model.Entry{ID: entry.ID, URL: entry.URL, ImageUrl: entry.ImageUrl, PublishedAt: entry.PublishedAt, Title: entry.Title, Language: entry.Language, Author: entry.Author, RawContent: entry.RawContent, FullContent: entry.FullContent}
	knowledge.UpdateLibraryEntryContent(entry.BflUser, updateEntry, true)

	var download model.EntryDownloadModel
	download.DataSource = downloadUrl
	download.DownloadAPP = "wise"
	download.EnclosureId = ""
	download.FileName = fileName
	download.FileType = fileType
	download.EntryId = entry.ID
	download.Path = "Downloads/Wise/" + folder
	download.BflUser = entry.BflUser
	download.LarepassId = larepassId
	if larepassId != "" {
		download.DownloadAPP = "larepass"
	}
	knowledge.DownloadDoReq(download)
	return entry.FullContent
}
