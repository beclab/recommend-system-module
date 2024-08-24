package common

import (
	"os"
	"path/filepath"
	"strconv"
)

const (
	FeedPathPrefix  = "/feed/"
	EntryPathPrefix = "/entry/"

	defaultSource          = "algo"
	defaultKnowledgeApiUrl = "http://localhost:3010"

	defaultNfSDir     = "/Users/simon/Desktop/workspace/pp/apps/rss-termius-v2/recommend_protocol/data1/nfs"
	defaultJUICEFSDir = "/Users/simon/Desktop/workspace/pp/apps/rss-termius-v2/recommend_protocol/data1/juicefs"
)

func GetFeedSyncPackageDataRedisKey() string {
	return "feed_sync_data"
}

func GetEntrySyncPackageDataRedisKey() string {
	return "entrysync"
}

func GetTermiusUserName() string {
	return os.Getenv("TERMIUS_USER_NAME")
}

func GeSyncFrequency() string {
	defaultFeq := "4"
	frequency := os.Getenv("SYNC_TASK_FREQUENCY")
	if frequency == "" {
		return defaultFeq
	}
	return frequency
}

func GetCrawlerFrequency() string {
	defaultFeq := "3"
	frequency := os.Getenv("CRAWLER_TASK_FREQUENCY")
	if frequency == "" {
		return defaultFeq
	}
	return frequency
}

func GetSyncTemplatePluginsUrl() string {
	envDir := os.Getenv("SYNC_TEMPLATE_PLUGINS_URL")
	if envDir == "" {
		return "https://recommend-provider-prd.bttcdn.com/api/provider/templatePlugins"
	}
	return envDir
}
func GetSyncDiscoveryFeedPackageUrl() string {
	envDir := os.Getenv("SYNC_DISCOVERY_FEEDPACKAGE_URL")
	if envDir == "" {
		return "https://recommend-provider-prd.bttcdn.com/api/provider/discoveryFeedPackages"
	}
	return envDir
}

func NFSRootDirectory() string {
	envDir := os.Getenv("NFS_ROOT_DIRECTORY")
	if envDir == "" {
		return defaultNfSDir
	}
	return envDir
}

func JUICEFSRootDirectory() string {

	envDir := os.Getenv("JUICEFS_ROOT_DIRECTORY")
	if envDir == "" {
		return defaultJUICEFSDir
	}
	return envDir
}

func knowledgeBaseUrl() string {
	env := os.Getenv("KNOWLEDGE_BASE_API_URL")
	if env == "" {
		return defaultKnowledgeApiUrl
	}
	return env
}

func FeedMonogoApiUrl() string {
	return knowledgeBaseUrl() + "/knowledge/feed/algorithm/"
}

func EntryMonogoEntryApiUrl() string {
	return knowledgeBaseUrl() + "/knowledge/entry/"
}

func RedisConfigApiUrl() string {
	return knowledgeBaseUrl() + "/knowledge/config/"
}

func SyncFeedDirectory(provider, packageName string) string {
	path := filepath.Join(JUICEFSRootDirectory(), FeedPathPrefix, provider, packageName)
	return path
}

func SyncEntryDirectory(provider, feedName, modelName string) string {
	path := filepath.Join(JUICEFSRootDirectory(), EntryPathPrefix, provider, feedName, modelName)
	return path
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
