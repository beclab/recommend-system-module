package main

import (
	"bytetrade.io/web3os/RSSync/cli"
	"bytetrade.io/web3os/RSSync/common"
	"bytetrade.io/web3os/RSSync/database"
	"bytetrade.io/web3os/RSSync/storage"
	"go.uber.org/zap"
)

func main() {

	redisdb := database.NewRedisConnection()

	db, err := database.NewConnectionPool(
		common.DatabaseURL(),
		common.DatabaseMinConns(),
		common.DatabaseMaxConns(),
		common.DatabaseConnectionLifetime(),
	)
	if err != nil {
		common.Logger.Error("Unable to initialize database connection pool", zap.Error(err))
	}
	defer db.Close()

	store := storage.NewStorage(db, redisdb)
	cli.StartDaemon(store)
}
