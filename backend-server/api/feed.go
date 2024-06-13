package api

import (
	"net/http"

	"bytetrade.io/web3os/backend-server/common"
	"bytetrade.io/web3os/backend-server/http/response/json"
	"bytetrade.io/web3os/backend-server/model"
	"bytetrade.io/web3os/backend-server/service"
	"go.uber.org/zap"

	"bytetrade.io/web3os/backend-server/http/request"
)

func (h *handler) refreshFeed(w http.ResponseWriter, r *http.Request) {
	feedID := request.RouteStringParam(r, "feedID")

	if !h.store.FeedExists(feedID) {
		json.NotFound(w, r)
		return
	}
	h.store.ResetFeedHeader(feedID)
	jobs := make([]model.Job, 0)
	jobs = append(jobs, model.Job{FeedID: feedID})
	go func() {
		h.pool.Push(jobs)
	}()
	//service.RefreshFeed(h.store, feedID)

	common.Logger.Info("refresh feed", zap.String("feedID", feedID))
	json.NoContent(w, r)
}

func (h *handler) knowledgeRefreshFeed(w http.ResponseWriter, r *http.Request) {
	feedID := request.RouteStringParam(r, "feedID")

	if !h.store.FeedExists(feedID) {
		json.NotFound(w, r)
		return
	}
	h.store.ResetFeedHeader(feedID)
	jobs := make([]model.Job, 0)
	jobs = append(jobs, model.Job{FeedID: feedID})
	go func() {
		h.pool.Push(jobs)
	}()

	common.Logger.Info("knowledge refresh feed", zap.String("feedID", feedID))
	json.NoContent(w, r)
}

func (h *handler) rssParse(w http.ResponseWriter, r *http.Request) {

	url := request.QueryStringParam(r, "url", "")
	feed := service.RssParseFromURL(url)
	if feed == nil {
		json.OK(w, r, model.ParseFeedResponseModel{Code: -1})
	} else {
		json.OK(w, r, model.ParseFeedResponseModel{Code: 0, Data: model.GetFeedParseModel(feed)})
	}
}
