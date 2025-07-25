package cli

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bytetrade.io/web3os/RSSync/common"
	"bytetrade.io/web3os/RSSync/scheduler"
	"bytetrade.io/web3os/RSSync/worker"

	"bytetrade.io/web3os/RSSync/storage"
)

const DefaultPort = "6317"

func StartDaemon(store *storage.Storage) {
	common.Logger.Info("Starting Service v0.0.25...")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	signal.Notify(stop, syscall.SIGTERM)

	pool := worker.NewPool(store, common.GetWorkPoolSize())
	scheduler.Serve(store, pool)

	httpServer := HttpdServe(store, pool)

	<-stop
	common.Logger.Info("Shutting down the process...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if httpServer != nil {
		httpServer.Shutdown(ctx)
	}

	common.Logger.Info("Process gracefully stopped")
}
