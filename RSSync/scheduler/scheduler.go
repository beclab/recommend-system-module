package scheduler

import (
	"time"

	"bytetrade.io/web3os/RSSync/common"
	"bytetrade.io/web3os/RSSync/service"
	"bytetrade.io/web3os/RSSync/storage"
	"bytetrade.io/web3os/RSSync/worker"
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

	go discoveryFeedSyncScheduler(
		store,
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

func discoveryFeedSyncScheduler(store *storage.Storage) {
	for range time.Tick(time.Duration(10) * time.Minute) {
		service.SyncDiscoveryFeedPackage(store)
		common.Logger.Info("discovery feed sync ...")
	}
}
