package api

import (
	"net/http"

	"bytetrade.io/web3os/backend-server/common"
	"bytetrade.io/web3os/backend-server/crawler"
	"bytetrade.io/web3os/backend-server/http/request"
	"bytetrade.io/web3os/backend-server/http/response/json"
	"bytetrade.io/web3os/backend-server/model"
	"bytetrade.io/web3os/backend-server/service/search"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	feed, _ := h.store.GetFeedById(entry.FeedID)
	//crawler.EntryCrawler(entry, feed) //entry.ID.Hex(), entry.URL, entry.Title, entry.ImageUrl, entry.Author, entry.PublishedAt, feed)

	feedUrl := ""
	userAgent := ""
	cookie := ""
	certificates := false
	fetchViaProxy := false

	var feedSearchRSSList []model.FeedNotification
	if feed != nil {

		feedUrl = feed.FeedURL
		userAgent = feed.UserAgent
		cookie = feed.Cookie
		certificates = feed.AllowSelfSignedCertificates
		fetchViaProxy = feed.FetchViaProxy

		if feed.ID != primitive.NilObjectID {
			feedNotification := model.FeedNotification{
				FeedId:   feed.ID.Hex(),
				FeedName: feed.Title,
				FeedIcon: "",
			}
			feedSearchRSSList = append(feedSearchRSSList, feedNotification)
		}
	}
	crawler.EntryCrawler(entry, feedUrl, userAgent, cookie, certificates, fetchViaProxy)

	notificationData := model.NotificationData{
		Name:      entry.Title,
		EntryId:   entry.ID.Hex(),
		Created:   entry.PublishedAt,
		FeedInfos: feedSearchRSSList,
		Content:   entry.FullContent,
	}
	docId := search.InputRSS(&notificationData)
	updateDocIDEntry := &model.Entry{ID: entry.ID, DocId: docId, Title: entry.Title, Language: entry.Language, Author: entry.Author, RawContent: entry.RawContent, FullContent: entry.FullContent}
	h.store.UpdateEntryContent(updateDocIDEntry)

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
