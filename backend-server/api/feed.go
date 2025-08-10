package api

import (
	"net/http"

	"bytetrade.io/web3os/backend-server/common"
	"bytetrade.io/web3os/backend-server/http/response/json"
	"bytetrade.io/web3os/backend-server/model"
	"go.uber.org/zap"

	"bytetrade.io/web3os/backend-server/http/request"
)

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
