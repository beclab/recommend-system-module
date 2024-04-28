package model

type FeedSyncPackageData struct {
	Name       string `json:"name"`
	Md5        string `json:"md5"`
	UpdateTime int64  `json:"update_time"`
}
type FeedSyncPackageDataList []FeedSyncPackageData

type EntrySyncPackageData struct {
	FeedName   string `json:"feed_name"`
	ModelName  string `json:"model_name"`
	Md5        string `json:"md5"`
	Language   string `json:"language"`
	StartTime  int64  `json:"start_time"`
	UpdateTime int64  `json:"update_time"`
}
type EntrySyncPackageDataList []EntrySyncPackageData

type FeedSyncData struct {
	SyncStartTimestamp int64 `json:"sync_start_timestamp"`
	SyncEndTimestamp   int64 `json:"sync_end_timestamp"`
}
