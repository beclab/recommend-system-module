package rss

import (
	"errors"
	"io"

	"bytetrade.io/web3os/RSSync/model"
	"bytetrade.io/web3os/RSSync/reader/xml"
)

// Parse returns a normalized feed struct from a RSS feed.
func Parse(baseURL string, data io.Reader) (*model.Feed, error) {
	feed := new(rssFeed)
	decoder := xml.NewDecoder(data)
	err := decoder.Decode(feed)
	if err != nil {
		return nil, errors.New("unable to parse RSS feed")
	}

	return feed.Transform(baseURL), nil
}
