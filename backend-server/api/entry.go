package api

import (
	"net/http"
	"strings"

	"bytetrade.io/web3os/backend-server/common"
	"bytetrade.io/web3os/backend-server/crawler"
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
	if entry.FullContent == "" {
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

	updateEntry := &model.Entry{ID: entry.ID, URL: entry.URL, PublishedAt: entry.PublishedAt, Title: entry.Title, Language: entry.Language, Author: entry.Author, RawContent: entry.RawContent, FullContent: entry.FullContent}
	//h.store.UpdateEntryContent(updateDocIDEntry)
	knowledge.UpdateLibraryEntryContent(updateEntry)

	if entry.MediaContent != "" || entry.MediaUrl != "" {
		knowledge.NewEnclosure(entry, h.store)
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

func (h *handler) radioDetection(w http.ResponseWriter, r *http.Request) {
	url := request.QueryStringParam(r, "url", "")
	useAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	rawContent := crawler.FetchRawContnt(
		url,
		"",
		useAgent,
		"",
		false,
		false,
	)
	result := processor.RadioDetectionInArticle(rawContent, url)
	json.OK(w, r, model.StrResponseModel{Code: 0, Data: result})
}
