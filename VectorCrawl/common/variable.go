package common

import (
	"os"
	"strconv"
)

const (
	defaultListenAddr            = "127.0.0.1:8081"
	defaultHTTPClientTimeout     = 20
	defaultHTTPClientMaxBodySize = 15 * 1024 * 1024

	defaultDownloadApiUrl = "http://localhost:3080/api"
	defaultYtdlpApiUrl    = "http://127.0.0.1:3082/api"
)

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

func GetListenAddr() string {
	return ParseString(os.Getenv("LISTEN_ADDR"), defaultListenAddr)
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

func ParseString(value string, defaultV string) string {
	if value == "" {
		return defaultV
	}
	return value
}

func GetHttpClientTimeout() int {
	return ParseInt(os.Getenv("HTTP_CLIENT_TIMEOUT"), defaultHTTPClientTimeout)
}

func GetHttpClientMaxBodySize() int {
	return ParseInt(os.Getenv("HTTP_CLIENT_MAX_BODYSIZE"), defaultHTTPClientMaxBodySize)
}
