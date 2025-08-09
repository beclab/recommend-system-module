package common

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	ArticleFileType = "article"
	VideoFileType   = "video"
	AudioFileType   = "audio"
	PdfFileType     = "pdf"
	EbookFileType   = "ebook"
)

const (
	defaultListenAddr            = "127.0.0.1:8081"
	defaultHTTPClientTimeout     = 20
	defaultHTTPClientMaxBodySize = 15 * 1024 * 1024

	defaultDatabaseURL = "host=127.0.0.1  user=postgres password=postgres dbname=rss sslmode=disable"
	//defaultDatabaseURL = "host=124.222.40.95  user=postgres password=liujx123 dbname=rss_v4 sslmode=disable"
	defaultPGHost    = "127.0.0.1"
	defaultPGUser    = "postgres"
	defaultPGPass    = "postgres"
	defaultPGPDBName = "rss"
	defaultPGPort    = 5432

	defaultDatabaseMaxConns           = 20
	defaultDatabaseMinConns           = 1
	defaultDatabaseConnectionLifetime = 5

	defaultEntryUpdateApiUrl = "http://localhost:3010/knowledge/entry/"

	defaultDownloadApiUrl = "http://localhost:3080/api"
	defaultYtdlpApiUrl    = "http://127.0.0.1:3082/api"

	FeedSource                   = "wise"
	DefaultWorkerPoolSize        = 1
	DefaultContentWorkerPoolSize = 3
	DefaultPollingFrequency      = 15
	DefaultBatchSize             = 100
)

func EntryUpdateApiUrl() string {
	env := os.Getenv("ENTRY_UPDATE_API_URL")
	if env == "" {
		return defaultEntryUpdateApiUrl
	}
	return env
}

func DownloadApiUrl() string {
	env := os.Getenv("DOWNLOAD_API_URL")
	if env == "" {
		return defaultDownloadApiUrl
	}
	return env
}

func YTDLPApiUrl() string {
	env := os.Getenv("YT_DLP_API_URL")
	if env == "" {
		return defaultYtdlpApiUrl
	}
	return env
}

func CurrentUser() string {
	env := ParseString(os.Getenv("CURRENT_USER"), "")
	return env
}

func GetListenAddr() string {
	return ParseString(os.Getenv("LISTEN_ADDR"), defaultListenAddr)
}

func GetPollingFrequency() int {
	return ParseInt(os.Getenv("POLLING_FREQUENCY"), DefaultPollingFrequency)
}

func GetPGHost() string {
	return ParseString(os.Getenv("PG_HOST"), defaultPGHost)
}

func GetPGUser() string {
	return ParseString(os.Getenv("PG_USERNAME"), defaultPGUser)
}

func GetPGPass() string {
	return ParseString(os.Getenv("PG_PASSWORD"), defaultPGPass)
}

func GetPGDbName() string {
	return ParseString(os.Getenv("PG_DATABASE"), defaultPGPDBName)
}

func GetPGPort() int {
	return ParseInt(os.Getenv("PG_PORT"), defaultPGPort)
}

func DatabaseURL() string {
	return fmt.Sprintf("host=%s  port=%d user=%s password=%s dbname=%s sslmode=disable", GetPGHost(), GetPGPort(), GetPGUser(), GetPGPass(), GetPGDbName())
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
	return ParseString(os.Getenv("RSS_HUB_URL"), "http://127.0.0.1:1200/")
}

func GetSyncDiscoveryFeedPackageUrl() string {
	envDir := os.Getenv("SYNC_DISCOVERY_FEEDPACKAGE_URL")
	if envDir == "" {
		return "https://recommend-provider-prd.bttcdn.com/api/provider/discoveryFeedPackages"
	}
	return envDir
}

func GetRedisAddr() string {
	return ParseString(os.Getenv("REDIS_ADDR"), "127.0.0.1:6379")
}

func GetRedisPassword() string {
	return ParseString(os.Getenv("REDIS_PASSWORD"), "")
}

func GetWatchDir() string {
	return ParseString(os.Getenv("WATCH_DIR"), "/data/Home/Downloads")
}
