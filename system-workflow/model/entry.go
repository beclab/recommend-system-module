package model

import "bytetrade.io/web3os/system_workflow/protobuf_entity"

type EntryAddModel struct {
	Source      string   `json:"source,omitempty"`
	Url         string   `json:"url,omitempty"`
	Title       string   `json:"title,omitempty"`
	FeedUrl     string   `json:"feed_url,omitempty"`
	PublishedAt int64    `json:"published_at,omitempty"`
	Author      string   `json:"author,omitempty"`
	KeywordList []string `json:"keyword,omitempty"`
	Language    string   `json:"language,omitempty"`
	ImageUrl    string   `json:"image_url,omitempty"`
	Crawler     bool     `json:"crawler,omitempty"`
	Extract     bool     `json:"extract,omitempty"`

	Starred     bool   `json:"starred,omitempty"`
	Saved       bool   `json:"saved,omitempty"`
	Unread      bool   `json:"unread,omitempty"`
	Readlater   bool   `json:"readlater,omitempty"`
	Disabled    bool   `json:"disabled,omitempty"`
	RawContent  string `json:"raw_content,omitempty"`
	FullContent string `json:"full_content,omitempty"`
}

type EntryAddResponseModel struct {
	ID     string `json:"_id,omitempty" `
	Source string `json:"source"`
	Url    string `json:"url"`
}

type MongoEntryApiResponseModel struct {
	Code    int                     `json:"code"`
	Message string                  `json:"message"`
	Data    []EntryAddResponseModel `json:"data"`
}

type EntryApiDataResponseModel struct {
	Count int             `json:"count"`
	Items []EntryAddModel `json:"items"`
}

type EntryApiResponseModel struct {
	Code    int                       `json:"code"`
	Message string                    `json:"message"`
	Data    EntryApiDataResponseModel `json:"data"`
}

func GetProtoEntryTransModel(model string, protoEntity *protobuf_entity.Entry) *protobuf_entity.EntryTrans {
	var result protobuf_entity.EntryTrans
	result.Url = protoEntity.Url
	result.CreatedAt = protoEntity.CreatedAt
	result.PublishedAt = protoEntity.PublishedAt
	result.Title = protoEntity.Title
	result.Author = protoEntity.Author
	result.FeedId = protoEntity.FeedId
	result.FeedUrl = protoEntity.FeedUrl
	result.ImageUrl = protoEntity.ImageUrl
	result.KeywordList = protoEntity.KeywordList
	result.Language = protoEntity.Language
	if model == "bert_v2" {
		result.Embedding = protoEntity.EmbeddingContentAll_MiniLM_L6V2Base
	} else if model == "bert_v3" {
		result.Embedding = protoEntity.EmbeddingContentParaphraseMultilingual_MiniLM_L12
	}
	result.RecallPoint = protoEntity.RecallPoint
	result.PublishedAtTimestamp = protoEntity.PublishedAtTimestamp
	return &result
}
