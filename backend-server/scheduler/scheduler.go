package scheduler

import (
	"time"

	"bytetrade.io/web3os/backend-server/common"
	"bytetrade.io/web3os/backend-server/storage"
	"bytetrade.io/web3os/backend-server/worker"
	"go.uber.org/zap"
)

// Serve starts the internal scheduler.
func Serve(store *storage.Storage, pool *worker.Pool) {
	common.Logger.Info("Starting scheduler...")
	go feedScheduler(
		store,
		pool,
		common.GetPollingFrequency(),
		common.DefaultBatchSize,
	)

}

func feedScheduler(store *storage.Storage, pool *worker.Pool, frequency, batchSize int) {
	for range time.Tick(time.Duration(frequency) * time.Minute) {
		jobs, err := store.FeedToUpdateList(batchSize)
		if err != nil {
			common.Logger.Error("Scheduler:Feed", zap.Error(err))
		} else {
			pool.Push(jobs)
		}
		common.Logger.Info("feedScheduler...")
	}
}
