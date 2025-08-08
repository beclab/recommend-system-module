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

func (h *handler) newFetchContent(entry *model.Entry) string {

	crawler.EntryCrawler(entry.URL, entry.BflUser, entry.FeedID)

	updateEntry := &model.Entry{ID: entry.ID, URL: entry.URL, ImageUrl: entry.ImageUrl, PublishedAt: entry.PublishedAt, Title: entry.Title, Language: entry.Language, Author: entry.Author, RawContent: entry.RawContent, FullContent: entry.FullContent}
	if entry.MediaContent != "" || entry.DownloadFileUrl != "" {
		updateEntry.Attachment = true
	}
	knowledge.UpdateLibraryEntryContent(entry.BflUser, updateEntry, false)
	if entry.DownloadFileUrl != "" {
		knowledge.DownloadDoReq(entry, nil, h.store)
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
	go func() {
		h.newFetchContent(entry)
	}()
	json.NoContent(w, r)

}

func (h *handler) parse(w http.ResponseWriter, r *http.Request) {

}

func (h *handler) asyncParse(w http.ResponseWriter, r *http.Request) {

}
