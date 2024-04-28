package main

import (
	"context"

	"bytetrade.io/web3os/backend-server/cli"
	"bytetrade.io/web3os/backend-server/common"
	"bytetrade.io/web3os/backend-server/database"
	"bytetrade.io/web3os/backend-server/storage"

	"go.uber.org/zap"
)

func main() {
	mongodb, err := database.NewMongodbConnection()
	if err != nil {
		common.Logger.Error("mongodb connect fail", zap.Error(err))
	}

	defer func() {
		if err := mongodb.Disconnect(context.Background()); err != nil {
			panic(err)
		}
	}()

	store := storage.NewStorage(mongodb)
	cli.StartDaemon(store)
}
