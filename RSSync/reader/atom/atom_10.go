package atom

import (
	"encoding/xml"
	"html"
	"strings"
	"time"

	"bytetrade.io/web3os/RSSync/common"
	"bytetrade.io/web3os/RSSync/model"
	"bytetrade.io/web3os/RSSync/reader/date"
	"bytetrade.io/web3os/RSSync/reader/media"
	"bytetrade.io/web3os/RSSync/reader/sanitizer"
	"go.uber.org/zap"
)

// Specs:
// https://tools.ietf.org/html/rfc4287
// https://validator.w3.org/feed/docs/atom.html
type atom10Feed struct {
	XMLName xml.Name      `xml:"http://www.w3.org/2005/Atom feed"`
	ID      string        `xml:"id"`
	Title   atom10Text    `xml:"title"`
	Authors atomAuthors   `xml:"author"`
	Links   atomLinks     `xml:"link"`
	Entries []atom10Entry `xml:"entry"`
}

func (a *atom10Feed) Transform(baseURL string) *model.Feed {
	var err error

	feed := new(model.Feed)

	feedURL := a.Links.firstLinkWithRelation("self")
	feed.FeedURL, err = common.AbsoluteURL(baseURL, feedURL)
	if err != nil {
		feed.FeedURL = feedURL
	}

	siteURL := a.Links.originalLink()
	feed.SiteURL, err = common.AbsoluteURL(baseURL, siteURL)
	if err != nil {
		feed.SiteURL = siteURL
	}

	feed.Title = html.UnescapeString(a.Title.String())
	if feed.Title == "" {
		feed.Title = feed.SiteURL
	}

	for _, entry := range a.Entries {
		item := entry.Transform()
		entryURL, err := common.AbsoluteURL(feed.SiteURL, item.URL)
		if err == nil {
			item.URL = entryURL
		}

		if item.Author == "" {
			item.Author = a.Authors.String()
		}

		if item.Title == "" {
			item.Title = sanitizer.TruncateHTML(item.Content, 100)
		}

		if item.Title == "" {
			item.Title = item.URL
		}

		feed.Entries = append(feed.Entries, item)
	}

	return feed
}

type atom10Entry struct {
	ID        string      `xml:"id"`
	Title     atom10Text  `xml:"title"`
	Published string      `xml:"published"`
	Updated   string      `xml:"updated"`
	Links     atomLinks   `xml:"link"`
	Summary   atom10Text  `xml:"summary"`
	Content   atom10Text  `xml:"http://www.w3.org/2005/Atom content"`
	Authors   atomAuthors `xml:"author"`
	media.Element
}

func (a *atom10Entry) Transform() *model.Entry {
	entry := new(model.Entry)
	entry.URL = a.Links.originalLink()
	entry.PublishedAtParsed = a.entryDate()
	entry.Author = a.Authors.String()
	entry.Content = a.entryContent()
	entry.Title = a.entryTitle()
	entry.ImageUrl = a.entryEnclosures()
	entry.CommentsURL = a.entryCommentsURL()
	return entry
}

func (a *atom10Entry) entryTitle() string {
	return html.UnescapeString(a.Title.String())
}

func (a *atom10Entry) entryContent() string {
	content := a.Content.String()
	if content != "" {
		return content
	}

	summary := a.Summary.String()
	if summary != "" {
		return summary
	}

	/*mediaDescription := a.FirstMediaDescription()
	if mediaDescription != "" {
		return mediaDescription
	}*/

	return ""
}

// Note: The published date represents the original creation date for YouTube feeds.
// Example:
// <published>2019-01-26T08:02:28+00:00</published>
// <updated>2019-01-29T07:27:27+00:00</updated>
func (a *atom10Entry) entryDate() time.Time {
	dateText := a.Published
	if dateText == "" {
		dateText = a.Updated
	}

	if dateText != "" {
		result, err := date.Parse(dateText)
		if err != nil {
			common.Logger.Error("atom (entry):", zap.String("id", a.ID), zap.Error(err))
			return time.Now()
		}

		return result
	}

	return time.Now()
}

// See https://tools.ietf.org/html/rfc4685#section-4
// If the type attribute of the atom:link is omitted, its value is assumed to be "application/atom+xml".
// We accept only HTML or XHTML documents for now since the intention is to have the same behavior as RSS.
func (a *atom10Entry) entryCommentsURL() string {
	commentsURL := a.Links.firstLinkWithRelationAndType("replies", "text/html", "application/xhtml+xml")
	if common.IsAbsoluteURL(commentsURL) {
		return commentsURL
	}
	return ""
}

type atom10Text struct {
	Type             string               `xml:"type,attr"`
	CharData         string               `xml:",chardata"`
	InnerXML         string               `xml:",innerxml"`
	XHTMLRootElement atomXHTMLRootElement `xml:"http://www.w3.org/1999/xhtml div"`
}

// Text: https://datatracker.ietf.org/doc/html/rfc4287#section-3.1.1.1
// HTML: https://datatracker.ietf.org/doc/html/rfc4287#section-3.1.1.2
// XHTML: https://datatracker.ietf.org/doc/html/rfc4287#section-3.1.1.3
func (a *atom10Text) String() string {
	var content string
	switch {
	case a.Type == "", a.Type == "text", a.Type == "text/plain":
		if strings.HasPrefix(strings.TrimSpace(a.InnerXML), `<![CDATA[`) {
			content = html.EscapeString(a.CharData)
		} else {
			content = a.InnerXML
		}
	case a.Type == "xhtml":
		var root = a.XHTMLRootElement
		if root.XMLName.Local == "div" {
			content = root.InnerXML
		} else {
			content = a.InnerXML
		}
	default:
		content = a.CharData
	}

	return strings.TrimSpace(content)
}

func (a *atom10Entry) entryEnclosures() string {
	for _, mediaContent := range a.AllMediaContents() {
		//if mediaContent.Medium == "image" {
		return mediaContent.URL
		//}
	}

	for _, mediaThumbnail := range a.Element.MediaThumbnails {
		return mediaThumbnail.URL
	}
	return ""
}

type atomXHTMLRootElement struct {
	XMLName  xml.Name `xml:"div"`
	InnerXML string   `xml:",innerxml"`
}
