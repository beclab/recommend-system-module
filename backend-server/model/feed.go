package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Feed represents a feed in the application.
type Feed struct {
	ID                          primitive.ObjectID `bson:"_id"`
	FeedURL                     string             `bson:"feed_url"`
	SiteURL                     string             `bson:"site_url"`
	Title                       string             `bson:"title"`
	Description                 string             `bson:"description"`
	Language                    string             `bson:"language"`
	IconMimeType                string             `bson:"icon_mime_type"`
	IconContent                 string             `bson:"icon_content"`
	CheckedAt                   time.Time          `bson:"checked_at"`
	ParsingErrorMsg             string             `bson:"parsing_error_message"`
	Readings                    int                `bson:"readings"`
	Likes                       int                `bson:"likes"`
	Followers                   int                `bson:"followers"`
	Velocity                    int                `bson:"velocity"`
	ParsingErrorCount           int                `bson:"parsing_error_count"`
	Status                      int                `bson:"status"`
	Remark                      string             `bson:"remark"`
	CategoryID                  string             `bson:"category_id"`
	UserAgent                   string             `bson:"user_agent"`
	Cookie                      string             `bson:"cookie"`
	Username                    string             `bson:"username"`
	Password                    string             `bson:"password"`
	AllowSelfSignedCertificates bool               `bson:"allow_self_signed_certificates"`
	FetchViaProxy               bool               `bson:"fetch_via_proxy"`
	IgnoreHTTPCache             bool               `bson:"ignore_http_cache"`
	EtagHeader                  string             `bson:"etag_header"`
	LastModifiedHeader          string             `bson:"last_modified_header"`
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
