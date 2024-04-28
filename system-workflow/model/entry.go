package model

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
