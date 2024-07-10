package api

import (
	"net/http"

	"bytetrade.io/web3os/backend-server/common"
	"bytetrade.io/web3os/backend-server/crawler"
	"bytetrade.io/web3os/backend-server/http/request"
	"bytetrade.io/web3os/backend-server/http/response/json"
	"bytetrade.io/web3os/backend-server/knowledge"
	"bytetrade.io/web3os/backend-server/model"
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

	updateDocIDEntry := &model.Entry{ID: entry.ID, PublishedAt: entry.PublishedAt, Title: entry.Title, Language: entry.Language, Author: entry.Author, RawContent: entry.RawContent, FullContent: entry.FullContent}
	h.store.UpdateEntryContent(updateDocIDEntry)

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
	if entry.FullContent == "" {
		go func() {
			h.newFetchContent(entry)
		}()
	}
	json.NoContent(w, r)

}
