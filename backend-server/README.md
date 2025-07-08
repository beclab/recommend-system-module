# backend server

Provide feeds update and article extract.

# Run

| 变量名                     | 描述                      | 示例                  |
|---------------------------|--------------------------|---------------------- |
| `LISTEN_ADDR`             | 本服务地址                | `127.0.0.1:8081`       |
| `REDIS_ADDR`              | redis地址                | `127.0.0.1:6379`       |
| `REDIS_PASSWORD`。        | redis密码                | `redis password`       |
| `PG_USERNAME`             | pg用户名                 | `root`                 |
| `PG_PASSWORD`             | pg密码                   | `your password`        |
| `PG_HOST`                 | pg主机                   | `localhost`            |
| `PG_PORT`                 | pg端口                   | `root`                 |
| `PG_DATABASE`             | pg数据库                 | `pg_database`           |
| `DOWNLOAD_API_URL`        | download服务地址          | `http://127.0.0.1:3080`|
| `YT_DLP_API_URL`          | yt_dlp服务地址            | `http://127.0.0.1:3082`|
| `RSS_HUB_URL`             | rsshub服务地址            | `http://127.0.0.1:1200`|
| `WE_CHAT_REFRESH_FEED_URL`| 微信服务地址。             | `https://recommend-wechat-prd.bttcdn.com/api/wechat/entries`|

localhost PG_PORT=5432 PG_USERNAME=postgres PG_PASSWORD=password PG_DATABASE=rss

```
export LISTEN_ADDR="127.0.0.1:8081"
export REDIS_ADDR="127.0.0.1:6379"
export REDIS_PASSWORD="password"
export PG_USERNAME="postgres"
export PG_PASSWORD="password"
export PG_HOST="localhost"
export PG_PORT=5432
export PG_DATABASE="rss"
export DOWNLOAD_API_URL="http://127.0.0.1:3080"
export YT_DLP_API_URL="http://127.0.0.1:3082"
export RSS_HUB_URL="http://127.0.0.1:3010/rss"
export WE_CHAT_REFRESH_FEED_URL="https://recommend-wechat-prd.bttcdn.com/api/wechat/entries"
go run main.go
```

    
