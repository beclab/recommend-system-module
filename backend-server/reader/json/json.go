package json

import (
	"strings"
	"time"

	"bytetrade.io/web3os/backend-server/model"
	"bytetrade.io/web3os/backend-server/reader/date"
	"bytetrade.io/web3os/backend-server/reader/sanitizer"
	"go.uber.org/zap"

	"bytetrade.io/web3os/backend-server/common"
)

type jsonFeed struct {
	Version string `json:"version"`
	Title   string `json:"title"`
	//SiteURL string `json:"home_page_url"`
	SiteURL string `json:"link"`
	FeedURL string `json:"feed_url"`
	//Authors []jsonAuthor `json:"authors"`
	//Author  jsonAuthor   `json:"author"`
	//Items   []jsonItem   `json:"items"`
	Authors []string   `json:"authors"`
	Author  string     `json:"author"`
	Items   []jsonItem `json:"item"`
}

type jsonItem struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	URL     string `json:"link"`
	Summary string `json:"summary"`
	//Text          string           `json:"content_text"`
	Text          string   `json:"description"`
	HTML          string   `json:"content_html"`
	DatePublished string   `json:"pubDate"`
	DateModified  string   `json:"date_modified"`
	Authors       []string `json:"authors"`
	Author        string   `json:"author"`
	//Authors       []jsonAuthor     `json:"authors"`
	//Author        jsonAuthor       `json:"author"`
	Attachments []jsonAttachment `json:"attachments"`
}

type jsonAttachment struct {
	URL      string `json:"url"`
	MimeType string `json:"mime_type"`
	Title    string `json:"title"`
	Size     int64  `json:"size_in_bytes"`
	Duration int    `json:"duration_in_seconds"`
}

func (j *jsonFeed) GetAuthor() string {
	if len(j.Authors) > 0 {
		return j.Authors[0]
		//return (getAuthor(j.Authors[0]))
	}
	//return getAuthor(j.Author)
	return j.Author
}

func (j *jsonFeed) Transform(baseURL string) *model.Feed {
	var err error

	feed := new(model.Feed)

	feed.FeedURL, err = common.AbsoluteURL(baseURL, j.FeedURL)
	if err != nil {
		feed.FeedURL = j.FeedURL
	}

	feed.SiteURL, err = common.AbsoluteURL(baseURL, j.SiteURL)
	if err != nil {
		feed.SiteURL = j.SiteURL
	}

	feed.Title = strings.TrimSpace(j.Title)
	if feed.Title == "" {
		feed.Title = feed.SiteURL
	}

	for _, item := range j.Items {
		entry := item.Transform()
		entryURL, err := common.AbsoluteURL(feed.SiteURL, entry.URL)
		if err == nil {
			entry.URL = entryURL
		}

		if entry.Author == "" {
			entry.Author = j.GetAuthor()
		}

		feed.Entries = append(feed.Entries, entry)
	}

	return feed
}

func (j *jsonItem) GetDate() time.Time {
	for _, value := range []string{j.DatePublished, j.DateModified} {
		if value != "" {
			d, err := date.Parse(value)
			if err != nil {
				common.Logger.Error("json:", zap.Error(err))
				return time.Now()
			}

			return d
		}
	}

	return time.Now()
}

func (j *jsonItem) GetAuthor() string {
	if len(j.Authors) > 0 {
		//return getAuthor(j.Authors[0])
		return j.Authors[0]
	}
	//return getAuthor(j.Author)
	return j.Author
}

func (j *jsonItem) GetTitle() string {
	if j.Title != "" {
		return j.Title
	}

	for _, value := range []string{j.Summary, j.Text, j.HTML} {
		if value != "" {
			return sanitizer.TruncateHTML(value, 100)
		}
	}

	return j.URL
}

func (j *jsonItem) GetContent() string {
	for _, value := range []string{j.HTML, j.Text, j.Summary} {
		if value != "" {
			return value
		}
	}

	return ""
}

func (j *jsonItem) Transform() *model.Entry {
	entry := new(model.Entry)
	entry.URL = j.URL
	entry.PublishedAtParsed = j.GetDate()
	entry.Author = j.GetAuthor()
	entry.Content = j.GetContent()
	entry.Title = strings.TrimSpace(j.GetTitle())
	return entry
}
