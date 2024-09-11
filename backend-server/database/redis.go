package database

import (
	"log"

	"bytetrade.io/web3os/backend-server/common"
	"github.com/go-redis/redis"
)

func NewRedisConnection() *redis.Client {
	add := common.GetRedisAddr()
	password := common.GetRedisPassword()
	log.Printf("redis connection uri:%s,password:%s", add, password)
	var rdb = redis.NewClient(&redis.Options{
		Addr:     add,
		Password: password,
		//Username: os.Getenv("TERMINUS_RECOMMEND_REDIS_USERNAME"),
		DB: 0,
	})
	return rdb
}
