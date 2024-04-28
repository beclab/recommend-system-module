package rdf

import (
	"encoding/xml"
	"html"
	"strings"
	"time"

	"bytetrade.io/web3os/backend-server/common"
	"bytetrade.io/web3os/backend-server/model"
	"bytetrade.io/web3os/backend-server/reader/date"
	"bytetrade.io/web3os/backend-server/reader/sanitizer"

	"go.uber.org/zap"
)

type rdfFeed struct {
	XMLName xml.Name  `xml:"RDF"`
	Title   string    `xml:"channel>title"`
	Link    string    `xml:"channel>link"`
	Items   []rdfItem `xml:"item"`
	DublinCoreFeedElement
}

func (r *rdfFeed) Transform(baseURL string) *model.Feed {
	var err error
	feed := new(model.Feed)
	feed.Title = sanitizer.StripTags(r.Title)
	feed.FeedURL = baseURL
	feed.SiteURL, err = common.AbsoluteURL(baseURL, r.Link)
	if err != nil {
		feed.SiteURL = r.Link
	}

	for _, item := range r.Items {
		entry := item.Transform()
		if entry.Author == "" && r.DublinCoreCreator != "" {
			entry.Author = strings.TrimSpace(r.DublinCoreCreator)
		}

		if entry.URL == "" {
			entry.URL = feed.SiteURL
		} else {
			entryURL, err := common.AbsoluteURL(feed.SiteURL, entry.URL)
			if err == nil {
				entry.URL = entryURL
			}
		}

		feed.Entries = append(feed.Entries, entry)
	}

	return feed
}

type rdfItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	DublinCoreEntryElement
}

func (r *rdfItem) Transform() *model.Entry {
	entry := new(model.Entry)
	entry.Title = r.entryTitle()
	entry.Author = r.entryAuthor()
	entry.URL = r.entryURL()
	entry.Content = r.entryContent()
	entry.PublishedAtParsed = r.entryDate()
	return entry
}

func (r *rdfItem) entryTitle() string {
	return html.UnescapeString(strings.TrimSpace(r.Title))
}

func (r *rdfItem) entryContent() string {
	switch {
	case r.DublinCoreContent != "":
		return r.DublinCoreContent
	default:
		return r.Description
	}
}

func (r *rdfItem) entryAuthor() string {
	return strings.TrimSpace(r.DublinCoreCreator)
}

func (r *rdfItem) entryURL() string {
	return strings.TrimSpace(r.Link)
}

func (r *rdfItem) entryDate() time.Time {
	if r.DublinCoreDate != "" {
		result, err := date.Parse(r.DublinCoreDate)
		if err != nil {
			common.Logger.Error("rss (entry):", zap.String("link", r.Link), zap.Error(err))
			return time.Now()
		}

		return result
	}

	return time.Now()
}
