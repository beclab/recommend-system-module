package worker

import (
	"time"

	"bytetrade.io/web3os/backend-server/common"
	"bytetrade.io/web3os/backend-server/model"
	"bytetrade.io/web3os/backend-server/service"
	"bytetrade.io/web3os/backend-server/storage"
	"go.uber.org/zap"
)

// Worker refreshes a feed in the background.
type Worker struct {
	id int
	//contentPool *contentworker.ContentPool
	store *storage.Storage
}

// Run wait for a job and refresh the given feed.
func (w *Worker) Run(c chan model.Job) {

	for {
		job := <-c
		common.Logger.Info("[Worker ] Received feed #%d ", zap.Int("id", w.id), zap.String("feed id", job.FeedID))
		//service.RefreshFeed(w.store, w.contentPool, job.FeedID)
		service.RefreshFeed(w.store, job.FeedID)
		time.Sleep(1 * time.Minute)
	}
}
