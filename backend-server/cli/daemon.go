package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bytetrade.io/web3os/backend-server/common"
	"bytetrade.io/web3os/backend-server/scheduler"
	"bytetrade.io/web3os/backend-server/service/rpc"
	"bytetrade.io/web3os/backend-server/worker"

	"bytetrade.io/web3os/backend-server/storage"
)

const DefaultPort = "6317"

func startZincRpc() {
	zincHost := os.Getenv("ZINC_HOST")
	zincPort := os.Getenv("ZINC_PORT")
	url := "http://" + zincHost + ":" + zincPort
	if zincHost == "" || zincPort == "" {
		url = "http://localhost:4080"
	}
	port := os.Getenv("W_PORT")
	if port == "" {
		port = DefaultPort
	}
	username := os.Getenv("ZINC_USER")
	if username == "" {
		username = "admin"
	}
	password := os.Getenv("ZINC_PASSWORD")
	if password == "" {
		password = "User#123"
	}

	fmt.Println("Init RPCSERVER!")
	rpc.InitRpcService(url, port, username, password, map[string]string{})

	fmt.Println("RPCSERVER to start!")
	contx := context.Background()
	rpcErr := rpc.RpcServer.Start(contx)

	if rpcErr != nil {
		panic(rpcErr)
	}

}
func StartDaemon(store *storage.Storage) {
	common.Logger.Info("Starting Service...")

	if common.GetZincRpcStart() {
		startZincRpc()
	} else {
		common.Logger.Info("zinc not start")
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	signal.Notify(stop, syscall.SIGTERM)

	//contentPool := contentworker.NewContentPool(store, common.GetContentWorkPoolSize())
	//pool := worker.NewPool(store, contentPool, common.GetWorkPoolSize())
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
