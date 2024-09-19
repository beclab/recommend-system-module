package cli

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"bytetrade.io/web3os/backend-server/common"
	"bytetrade.io/web3os/backend-server/scheduler"
	"bytetrade.io/web3os/backend-server/service"
	"bytetrade.io/web3os/backend-server/worker"

	"bytetrade.io/web3os/backend-server/storage"
)

const DefaultPort = "6317"

func StartDaemon(store *storage.Storage) {
	common.Logger.Info("Starting Service...")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	signal.Notify(stop, syscall.SIGTERM)

	//contentPool := contentworker.NewContentPool(store, common.GetContentWorkPoolSize())
	//pool := worker.NewPool(store, contentPool, common.GetWorkPoolSize())
	pool := worker.NewPool(store, common.GetWorkPoolSize())
	scheduler.Serve(store, pool)

	httpServer := HttpdServe(store, pool)

	watchDirStr := common.GetWatchDir()
	watchDirs := strings.Split(watchDirStr, ",")
	for i, dir := range watchDirs {
		watchDirs[i] = strings.TrimSpace(dir)
	}
	if len(watchDirs) > 0 {
		service.WatchPath(store, watchDirs)
	}

	<-stop
	common.Logger.Info("Shutting down the process...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if httpServer != nil {
		httpServer.Shutdown(ctx)
	}

	common.Logger.Info("Process gracefully stopped")
}
