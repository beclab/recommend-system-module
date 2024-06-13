package model

import (
	"bytetrade.io/web3os/system_workflow/protobuf_entity"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Discovery struct {
	ID          primitive.ObjectID `bson:"_id"`
	FeedUrl     string             `bson:"feed_url"`
	SiteUrl     string             `bson:"site_url"`
	Title       string             `bson:"title"`
	Description string             `bson:"description"`
	IconType    string             `bson:"icon_type"`
	IconContent string             `bson:"icon_content"`
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
