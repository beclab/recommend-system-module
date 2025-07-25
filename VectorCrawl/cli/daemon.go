package cli

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bytetrade.io/web3os/vector-crawl/common"
)

func StartDaemon() {
	common.Logger.Info("Starting Page Parse Service ...")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	signal.Notify(stop, syscall.SIGTERM)

	httpServer := HttpdServe()

	<-stop
	common.Logger.Info("Shutting down the process...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if httpServer != nil {
		httpServer.Shutdown(ctx)
	}

	common.Logger.Info("Process gracefully stopped")
}
