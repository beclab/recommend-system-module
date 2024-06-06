package common

import (
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultListenAddr            = "127.0.0.1:8081"
	defaultHTTPClientTimeout     = 20
	defaultHTTPClientMaxBodySize = 15 * 1024 * 1024

	defaultMongodbURI          = "mongodb://localhost:27017"
	defaultMongodbName         = "document"
	defaultMongoFeedColl       = "feeds"
	defaultMongoEntryColl      = "entries"
	defaultMongoAlgorithmsColl = "algorithms"

	defaultDatabaseURL                = "host=124.222.40.95  user=postgres password=liujx123 dbname=rss_v3 sslmode=disable"
	defaultDatabaseMaxConns           = 20
	defaultDatabaseMinConns           = 1
	defaultDatabaseConnectionLifetime = 5

	defaultEntryMongoUpdateApiUrl = "http://localhost:3010/knowledge/entry/"

	FeedSource                   = "wise"
	DefaultWorkerPoolSize        = 1
	DefaultContentWorkerPoolSize = 3
	DefaultPollingFrequency      = 15
	DefaultBatchSize             = 100
)

func EntryMonogoUpdateApiUrl() string {
	env := os.Getenv("ENTRY_MONGO_UPDATE_API_URL")
	if env == "" {
		return defaultEntryMongoUpdateApiUrl
	}
	return env
}

func GetListenAddr() string {
	return ParseString(os.Getenv("LISTEN_ADDR"), defaultListenAddr)
}

func GetPollingFrequency() int {
	return ParseInt(os.Getenv("POLLING_FREQUENCY"), DefaultPollingFrequency)
}

func DatabaseURL() string {
	return ParseString(os.Getenv("DATABASE_URL"), defaultDatabaseURL)
}

func DatabaseMaxConns() int {
	return ParseInt(os.Getenv("DATABASE_MAX_CONNS"), defaultDatabaseMaxConns)
}

func DatabaseMinConns() int {
	return ParseInt(os.Getenv("DATABASE_MIN_CONNS"), defaultDatabaseMinConns)
}

func DatabaseConnectionLifetime() time.Duration {
	lifeTIme := ParseInt(os.Getenv("DATABASE_LIFETIME"), defaultDatabaseConnectionLifetime)
	return time.Duration(lifeTIme) * time.Minute
}

func GetMongoURI() string {
	return ParseString(os.Getenv("MONGODB_URI"), defaultMongodbURI)
}
func GetMongoDbName() string {
	return ParseString(os.Getenv("MONGODB_NAME"), defaultMongodbName)
}

func GetMongoFeedColl() string {
	return ParseString(os.Getenv("MONGODB_FEED_COLL"), defaultMongoFeedColl)
}

func GetMongoEntryColl() string {
	return ParseString(os.Getenv("MONGODB_ENTRY_COLL"), defaultMongoEntryColl)
}

func GetMongoAlgorithmsColl() string {
	return ParseString(os.Getenv("MONGODB_ALGORITHMS_COLL"), defaultMongoAlgorithmsColl)
}

func GetZincRpcStart() bool {
	return parseBool(os.Getenv("ZINC_RPC_START"), false)
}

func GetWorkPoolSize() int {
	return ParseInt(os.Getenv("WORK_POOL_SIZE"), DefaultWorkerPoolSize)
}

func GetContentWorkPoolSize() int {
	return ParseInt(os.Getenv("CONTENT_WORK_POOL_SIZE"), DefaultContentWorkerPoolSize)
}

func parseBool(value string, fallback bool) bool {
	if value == "" {
		return fallback
	}

	value = strings.ToLower(value)
	if value == "1" || value == "yes" || value == "true" || value == "on" {
		return true
	}

	return false
}

func ParseInt(value string, defaultV int) int {
	if value == "" {
		return defaultV
	}

	v, err := strconv.Atoi(value)
	if err != nil {
		return defaultV
	}
	return v
}

func ParseBool(value string, defaultV bool) bool {
	if value == "" {
		return defaultV
	}

	value = strings.ToLower(value)
	if value == "1" || value == "yes" || value == "true" || value == "on" {
		return true
	}

	return false
}

func ParseString(value string, defaultV string) string {
	if value == "" {
		return defaultV
	}
	return value
}

func GetPollingParsingErrorLimit() int {
	return ParseInt(os.Getenv("POLLING_PARSING_ERROR_LIMIT"), 3)
}

func GetHttpClientTimeout() int {
	return ParseInt(os.Getenv("HTTP_CLIENT_TIMEOUT"), defaultHTTPClientTimeout)
}

func GetHttpClientMaxBodySize() int {
	return ParseInt(os.Getenv("HTTP_CLIENT_MAX_BODYSIZE"), defaultHTTPClientMaxBodySize)
}

func GetWeChatFeedRefrshUrl() string {
	return ParseString(os.Getenv("WE_CHAT_REFRESH_FEED_URL"), "https://recommend-wechat-test.bttcdn.com/api/wechat/entries")
	//return ParseString(os.Getenv("WE_CHAT_REFRESH_FEED_URL"), "http://127.0.0.1:8080/api/wechat/entries")
}

func GetWeChatEntryContentUrl() string {
	return ParseString(os.Getenv("WECHAT_ENTRY_CONTENT_GET_API_URL"), "http://127.0.0.1:8080/api/wechat/entry/content")
}

func GetRSSHubUrl() string {
	return ParseString(os.Getenv("RSS_HUB_URL"), "http://127.0.0.1:3000/rss")
}
