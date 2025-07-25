package model

import (
	"bytetrade.io/web3os/RSSync/protobuf_entity"
)

type DiscoveryFeedPackagInfo struct {
	ID  string `json:"id"`
	Url string `json:"url"`
	MD5 string `json:"md5"`
}
type DiscoveryFeedPackagInfos []*DiscoveryFeedPackagInfo

type Discovery struct {
	ID          string `json:"id"`
	FeedUrl     string `json:"feed_url"`
	SiteUrl     string `json:"site_url"`
	Title       string `json:"title"`
	Description string `json:"description"`
	IconType    string `json:"icon_type"`
	IconContent string `json:"icon_content"`
}

func GetDiscoveryModel(protoFeed *protobuf_entity.DiscoveryFeed) *Discovery {
	var model Discovery
	model.FeedUrl = protoFeed.FeedUrl
	model.SiteUrl = protoFeed.SiteUrl
	model.Title = protoFeed.Title
	model.Description = protoFeed.Description

	model.IconType = protoFeed.IconType
	model.IconContent = protoFeed.IconContent
	return &model
}
