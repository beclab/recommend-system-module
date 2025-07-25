package worker

import (
	"bytetrade.io/web3os/RSSync/common"
	"bytetrade.io/web3os/RSSync/model"
	"bytetrade.io/web3os/RSSync/service"
	"bytetrade.io/web3os/RSSync/storage"
	"go.uber.org/zap"
)

// Worker refreshes a feed in the background.
type Worker struct {
	id    int
	store *storage.Storage
}

// Run wait for a job and refresh the given feed.
func (w *Worker) Run(c chan model.Job) {

	for {
		job := <-c
		common.Logger.Info("[Worker ] Received feed #%d ", zap.Int("id", w.id), zap.String("feed id", job.FeedID))
		service.RefreshFeed(w.store, job.FeedID)
	}
}
