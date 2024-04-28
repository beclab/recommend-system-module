package model

type EntryPackage struct {
	ID        string `json:"id"`
	URL       string `json:"url"`
	StartTime int64  `json:"start_time"`
	EndTime   int64  `json:"end_time"`
	Language  string `json:"language"`
	MD5       string `json:"md5"`
	ModelName string `json:"model_name"`
	FeedName  string `json:"feed_name"`
}
type EntryPackages []*EntryPackage

type FeedPackageAllInfo struct {
	ID          string `json:"id"`
	Provider    string `json:"provider"`
	FeedName    string `json:"feed_name"`
	PackageTime int64  `json:"package_time"`
	Url         string `json:"url"`
	MD5         string `json:"md5"`
}

type FeedPackageIncrementInfo struct {
	ID                    string `json:"id"`
	Provider              string `json:"provider"`
	FeedName              string `json:"feed_name"`
	FromTime              int64  `json:"from_time"`
	EndTime               int64  `json:"end_time"`
	Url                   string `json:"url"`
	MD5                   string `json:"md5"`
	Interval              int64  `json:"interval"`
	UnionHash             string `json:"union_hash"`
	FeedOperationSize     int32  `json:"feed_operation_size"`
	FeedNameOperationSize int32  `json:"feed_name_operation_size"`
}

type FeedPackageAllInfos []*FeedPackageAllInfo
type FeedPackageIncrementInfos []*FeedPackageIncrementInfo
