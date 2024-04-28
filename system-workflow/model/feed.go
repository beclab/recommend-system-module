package model

import (
	"bytetrade.io/web3os/system_workflow/protobuf_entity"
)

type FeedAddModel struct {
	FeedUrl     string `json:"feed_url"`
	SiteUrl     string `json:"site_url"`
	Source      string `json:"source"`
	Crawler     bool   `json:"crawler"`
	Title       string `json:"title"`
	Description string `json:"description"`
	IconType    string `json:"icon_type"`
	IconContent string `json:"icon_content"`
}

type MongoApiResponseModel struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Data    []string `json:"data"`
}

type MongoFeedDelModel struct {
	FeedUrls []string `json:"feed_urls"`
}

func GetFeedAddModel(protoFeed *protobuf_entity.Feed) *FeedAddModel {
	var model FeedAddModel
	model.FeedUrl = protoFeed.FeedUrl
	model.SiteUrl = protoFeed.SiteUrl
	model.Title = protoFeed.Title
	model.Description = protoFeed.Description

	model.IconType = protoFeed.IconType
	//model.IconContent = fmt.Sprintf("%s;base64,%s", model.IconType, base64.StdEncoding.EncodeToString(protoFeed.IconContent))
	model.IconContent = protoFeed.IconContent
	model.Crawler = true
	return &model
}

func GetUpdateProtoFeed(oldFeed *protobuf_entity.Feed, updateFeed *protobuf_entity.Feed) *protobuf_entity.Feed {
	if updateFeed.SiteUrl != "" {
		oldFeed.SiteUrl = updateFeed.SiteUrl
	}
	if updateFeed.Title != "" {
		oldFeed.Title = updateFeed.Title
	}
	if updateFeed.Description != "" {
		oldFeed.Description = updateFeed.Description
	}
	if updateFeed.IconContent != "" {
		oldFeed.IconContent = updateFeed.IconContent
	}
	if updateFeed.LastModifyTime != 0 {
		oldFeed.LastModifyTime = updateFeed.LastModifyTime
	}
	if updateFeed.Status != 0 {
		oldFeed.Status = updateFeed.Status
	}
	if updateFeed.Reading != 0 {
		oldFeed.Reading = updateFeed.Reading
	}
	if updateFeed.Likes != 0 {
		oldFeed.Likes = updateFeed.Likes
	}
	if updateFeed.Followers != 0 {
		oldFeed.Followers = updateFeed.Followers
	}

	return oldFeed
}
