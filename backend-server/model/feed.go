package model

import (
	"time"
)

// Feed represents a feed in the application.
type Feed struct {
	ID                          string    `json:"id"`
	FeedURL                     string    `json:"feed_url"`
	SiteURL                     string    `json:"site_url"`
	Title                       string    `json:"title"`
	Description                 string    `json:"description"`
	Language                    string    `json:"language"`
	IconMimeType                string    `json:"icon_type"`
	IconContent                 string    `json:"icon_content"`
	CheckedAt                   time.Time `json:"checked_at"`
	ParsingErrorMsg             string    `json:"parsing_error_message"`
	Readings                    int       `json:"readings"`
	Likes                       int       `json:"likes"`
	Followers                   int       `json:"followers"`
	Velocity                    int       `json:"velocity"`
	ParsingErrorCount           int       `json:"parsing_error_count"`
	Status                      int       `json:"status"`
	Remark                      string    `json:"remark"`
	CategoryID                  string    `json:"category_id"`
	UserAgent                   string    `json:"user_agent"`
	Cookie                      string    `json:"cookie"`
	Username                    string    `json:"username"`
	Password                    string    `json:"password"`
	AllowSelfSignedCertificates bool      `json:"allow_self_signed_certificates"`
	FetchViaProxy               bool      `json:"fetch_via_proxy"`
	IgnoreHTTPCache             bool      `json:"ignore_http_cache"`
	EtagHeader                  string    `json:"etag_header"`
	LastModifiedHeader          string    `json:"last_modified_header"`
	AutoDownload                bool      `json:"auto_download"`
	Entries                     Entries
}

type FeedCreationRequest struct {
	ID                          string  `json:"id"`
	FeedURL                     string  `json:"feed_url"`
	CategoryID                  string  `json:"category_id"`
	UserAgent                   string  `json:"user_agent"`
	Cookie                      string  `json:"cookie"`
	Username                    string  `json:"username"`
	Password                    string  `json:"password"`
	UrlRewriteRules             *string `json:"urlrewrite_rules"`
	IgnoreHTTPCache             *bool   `json:"ignore_http_cache"`
	AllowSelfSignedCertificates bool    `json:"allow_self_signed_certificates"`
	FetchViaProxy               bool    `json:"fetch_via_proxy"`
	Status                      int     `json:"status"`
	Remark                      string  `bson:"remark"`
	Description                 string  `json:"description"`
}

type FeedModificationRequest struct {
	ID                          string  `json:"id"`
	FeedURL                     *string `json:"feed_url"`
	SiteURL                     *string `json:"site_url"`
	Title                       *string `json:"title"`
	CategoryID                  *string `json:"category_id"`
	UserAgent                   *string `json:"user_agent"`
	Cookie                      *string `json:"cookie"`
	Username                    *string `json:"username"`
	Password                    *string `json:"password"`
	UrlRewriteRules             *string `json:"urlrewrite_rules"`
	IgnoreHTTPCache             *bool   `json:"ignore_http_cache"`
	AllowSelfSignedCertificates *bool   `json:"allow_self_signed_certificates"`
	FetchViaProxy               *bool   `json:"fetch_via_proxy"`
	Status                      int     `json:"status"`
	Remark                      string  `bson:"remark"`
	Description                 *string `json:"description"`
}

// WithError adds a new error message and increment the error counter.
func (f *Feed) WithError(message string) {
	f.ParsingErrorCount++
	f.ParsingErrorMsg = message
}

// ResetErrorCounter removes all previous errors.
func (f *Feed) ResetErrorCounter() {
	f.ParsingErrorCount = 0
	f.ParsingErrorMsg = ""
}

// Patch updates a feed with modified values.
func (f *FeedModificationRequest) Patch(feed *Feed) {
	if f.FeedURL != nil && *f.FeedURL != "" {
		feed.FeedURL = *f.FeedURL
	}

	if f.SiteURL != nil && *f.SiteURL != "" {
		feed.SiteURL = *f.SiteURL
	}

	if f.Title != nil && *f.Title != "" {
		feed.Title = *f.Title
	}

	if f.UserAgent != nil {
		feed.UserAgent = *f.UserAgent
	}

	if f.Cookie != nil {
		feed.Cookie = *f.Cookie
	}

	if f.Username != nil {
		feed.Username = *f.Username
	}

	if f.Password != nil {
		feed.Password = *f.Password
	}

	if f.CategoryID != nil {
		feed.CategoryID = *f.CategoryID
	}

	if f.IgnoreHTTPCache != nil {
		feed.IgnoreHTTPCache = *f.IgnoreHTTPCache
	}

	if f.AllowSelfSignedCertificates != nil {
		feed.AllowSelfSignedCertificates = *f.AllowSelfSignedCertificates
	}

	if f.FetchViaProxy != nil {
		feed.FetchViaProxy = *f.FetchViaProxy
	}

	if f.Description != nil && *f.Description != "" {
		feed.Description = *f.Description
	}

}

// Feeds is a list of feed
type Feeds []*Feed

type FeedParseModel struct {
	FeedURL      string `json:"feed_url"`
	SiteURL      string `json:"site_url"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	IconMimeType string `json:"icon_type"`
	IconContent  string `json:"icon_content"`
}

type ParseFeedResponseModel struct {
	Code int            `json:"code"`
	Data FeedParseModel `json:"data"`
}

type StrResponseModel struct {
	Code int    `json:"code"`
	Data string `json:"data"`
}

func GetFeedParseModel(feedModel *Feed) FeedParseModel {
	var result FeedParseModel

	result.Title = feedModel.Title
	result.FeedURL = feedModel.FeedURL
	result.SiteURL = feedModel.SiteURL
	result.Description = feedModel.Description
	result.IconMimeType = feedModel.IconMimeType
	result.IconContent = feedModel.IconContent
	return result
}
