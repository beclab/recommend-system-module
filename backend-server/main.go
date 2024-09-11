package main

import (
	"bytetrade.io/web3os/backend-server/cli"
	"bytetrade.io/web3os/backend-server/common"
	"bytetrade.io/web3os/backend-server/database"
	"bytetrade.io/web3os/backend-server/storage"
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

	/*mongodb, err := database.NewMongodbConnection()
	if err != nil {
		common.Logger.Error("mongodb connect fail", zap.Error(err))
	}

	defer func() {
		if err := mongodb.Disconnect(context.Background()); err != nil {
			panic(err)
		}
	}()*/

	store := storage.NewStorage(db, redisdb)
	cli.StartDaemon(store)
}
