package rss

// FeedBurnerElement represents FeedBurner XML elements.
type FeedBurnerElement struct {
	FeedBurnerLink          string `xml:"http://rssnamespace.org/feedburner/ext/1.0 origLink"`
	FeedBurnerEnclosureLink string `xml:"http://rssnamespace.org/feedburner/ext/1.0 origEnclosureLink"`
}
