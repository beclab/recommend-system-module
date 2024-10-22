package model

import (
	"time"

	"bytetrade.io/web3os/backend-server/reader/date"
)

type Entry struct {
	ID                string  `json:"id"`
	FeedID            *string `json:"feed"`
	Status            string  `json:"status"`
	Title             string  `json:"title"`
	URL               string  `json:"url"`
	CommentsURL       string  `json:"comments_url"`
	PublishedAtParsed time.Time
	PublishedAt       int64 `json:"published_at"`

	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
	Content    string    `json:"content"`
	RawContent string    `json:"raw_content"`
	//PureContent string             `bson:"pure_content"`
	FullContent string `json:"full_content"`
	//DocId       string   `json:"doc_id"`
	Author       string `json:"author"`
	ImageUrl     string `json:"image_url"`
	Readlater    bool   `json:"readlater"`
	Crawler      bool   `json:"crawler"`
	Starred      bool   `json:"starred"`
	Disabled     bool   `json:"disabled"`
	Saved        bool   `json:"saved"`
	Unread       bool   `json:"unread"`
	Language     string `json:"language"`
	MediaContent string `json:"media_content"`
	MediaUrl     string `json:"media_url"`
	MediaType    string `json:"media_type"`

	Sources []string `json:"sources"`
}

type EntryAddModel struct {
	Url         string `json:"url,omitempty"`
	Title       string `json:"title,omitempty"`
	FeedUrl     string `json:"feed_url,omitempty"`
	PublishedAt int64  `json:"published_at,omitempty"`
	RawContent  string `json:"raw_content,omitempty"`
	FullContent string `json:"full_content,omitempty"`
	Author      string `json:"author,omitempty"`
	ImageUrl    string `json:"image_url,omitempty"`
	Starred     bool   `json:"starred,omitempty"`
	Disabled    bool   `json:"disabled,omitempty"`
	Saved       bool   `json:"saved,omitempty"`
	Unread      bool   `json:"unread,omitempty"`
	Crawler     bool   `json:"crawler,omitempty"`
	Extract     bool   `json:"extract,omitempty"`
	Readlater   bool   `json:"readlater,omitempty"`
	Language    string `json:"language,omitempty"`
	Source      string `json:"source"`
}

type EntryAddResponseModel struct {
	ID     string `json:"id,omitempty" `
	Source string `json:"source"`
	Url    string `json:"url"`
}

type MongoEntryApiResponseModel struct {
	Code    int                     `json:"code"`
	Message string                  `json:"message"`
	Data    []EntryAddResponseModel `json:"data"`
}

type EntryFetchResponseModel struct {
	Code int   `json:"code"`
	Data Entry `json:"data"`
}

func GetEntryAddModel(entryModel *Entry, feedUrl string) *EntryAddModel {
	var result EntryAddModel
	result.Url = entryModel.URL
	result.Title = entryModel.Title
	result.FeedUrl = feedUrl
	result.PublishedAt = entryModel.PublishedAt
	result.Author = entryModel.Author
	result.RawContent = entryModel.RawContent
	result.FullContent = entryModel.FullContent
	result.ImageUrl = entryModel.ImageUrl
	result.Crawler = true
	result.Extract = true
	result.Language = entryModel.Language

	result.Readlater = false
	result.Starred = false
	result.Disabled = false
	result.Saved = false
	result.Unread = true

	result.Source = "wise"
	return &result
}

func GetEntryUpdateSourceModel(entryModel *Entry, feedUrl string) *EntryAddModel {
	var result EntryAddModel
	result.Url = entryModel.URL
	result.FeedUrl = feedUrl
	result.Source = "wise"
	return &result
}

// Entries represents a list of entries.
type Entries []*Entry

type WeChatEntry struct {
	Title       string `json:"title"`
	URL         string `bson:"url"`
	PublishedAt string `json:"published_at"`
	CreatedAt   time.Time
	Content     string `bson:"content"`
	Author      string `bson:"author"`
	ImageUrl    string `bson:"image_url"`

	ReadNum       int `json:"read_num"`
	ShareLikeNum  int `json:"share_like_num"`
	LikeNum       int `json:"like_num"`
	Idx           int `json:"idx"`
	CopyrightStat int `json:"copyright_stat"`

	AccountUsername string `json:"account_username"`
	AccountNickname string `json:"account_nickname"`
	AccountAvatar   string `json:"account_avatar"`
}

type WechatEntryResponse struct {
	RawContent  string `json:"raw_content"`
	FullContent string `json:"full_content"`
}

func GetEntryFromWeChatEntry(wechatEntry *WeChatEntry) *Entry {
	var entry Entry

	publishedDate, _ := date.Parse(wechatEntry.PublishedAt)

	entry.Title = wechatEntry.Title
	entry.URL = wechatEntry.URL
	entry.PublishedAtParsed = publishedDate
	entry.Content = wechatEntry.Content
	entry.Author = wechatEntry.Author
	entry.ImageUrl = wechatEntry.ImageUrl

	return &entry
}

type WeChatEntries []*WeChatEntry

type EntryDownloadModel struct {
	DataSource  string `json:"data_source"`
	TaskUser    string `json:"task_user"`
	DownloadAPP string `json:"download_app"`
	EnclosureId string `json:"enclosure_id"`
	FileName    string `json:"file_name"`
	FileType    string `json:"file_type"`
	Path        string `json:"path"`
}
