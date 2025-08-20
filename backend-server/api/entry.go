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
	content := h.newFetchContent(entry)
	json.OK(w, r, content)
}

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

func (h *handler) noMediaDownloadQuery(w http.ResponseWriter, r *http.Request) {
	url := request.QueryStringParam(r, "url", "")
	bflUser := request.QueryStringParam(r, "bfl_user", "")
	common.Logger.Info("knowledge download file query", zap.String("url", url), zap.String("bfl_user", bflUser))

	downloadUrl, downloadFile, downloadFileType := processor.DownloadTypeQueryByUrl(url)
	if downloadFileType != "" {
		h.respondWithJSON(w, r, downloadUrl, downloadFileType, downloadFile, "")
		return
	}

	rawContent, fileTypeFromContentType, fileNameFromContentType := crawler.FetchRawContent(bflUser, url)
	if fileTypeFromContentType != "" {
		h.respondWithJSON(w, r, url, fileTypeFromContentType, fileNameFromContentType, "")
		return
	}

	_, _, _, imageUrlFromContent, title, _, _, _, downloadFileUrl, downloadFileType := processor.ArticleExtractor(rawContent, url)
	if downloadFileType != "" {
		fileName := crawler.GetFileNameFromUrl(downloadFileUrl, downloadFileType)
		h.respondWithJSON(w, r, downloadFileUrl, downloadFileType, fileName, "")
		return
	} else {
		h.respondWithJSON(w, r, downloadFileUrl, "text/html", title, imageUrlFromContent)
	}

}

func (h *handler) respondWithJSON(w http.ResponseWriter, r *http.Request, downloadUrl, fileType, fileName string, thumbnail string) {
	json.OK(w, r, model.DownloadFetchResponseModel{
		Code: 0,
		Data: model.DownloadFetchReqModel{
			DownloadUrl: downloadUrl,
			FileType:    fileType,
			FileName:    fileName,
			Thumbnail:   thumbnail,
		},
	})
}
