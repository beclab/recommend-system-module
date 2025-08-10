package api

import (
	"net/http"

	"bytetrade.io/web3os/backend-server/common"
	"bytetrade.io/web3os/backend-server/crawler"
	"bytetrade.io/web3os/backend-server/http/request"
	"bytetrade.io/web3os/backend-server/http/response/json"
	"bytetrade.io/web3os/backend-server/knowledge"
	"bytetrade.io/web3os/backend-server/model"
	"bytetrade.io/web3os/backend-server/service"
	"go.uber.org/zap"
)

func (h *handler) newFetchContent(entry *model.Entry) string {

	updateEntry := crawler.EntryCrawler(entry.URL, entry.BflUser, entry.FeedID)
	service.CopyEntry(entry, updateEntry)
	if entry.MediaContent != "" || entry.DownloadFileUrl != "" {
		entry.Attachment = true
	}
	knowledge.UpdateLibraryEntryContent(entry.BflUser, entry)
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
