package parser

import (
	"errors"
	"strings"

	"bytetrade.io/web3os/backend-server/model"
	"bytetrade.io/web3os/backend-server/reader/atom"
	"bytetrade.io/web3os/backend-server/reader/json"
	"bytetrade.io/web3os/backend-server/reader/rdf"
	"bytetrade.io/web3os/backend-server/reader/rss"
)

// ParseFeed analyzes the input data and returns a normalized feed object.
func ParseFeed(baseURL, data string) (*model.Feed, error) {
	switch DetectFeedFormat(data) {
	case FormatAtom:
		return atom.Parse(baseURL, strings.NewReader(data))
	case FormatRSS:
		return rss.Parse(baseURL, strings.NewReader(data))
	case FormatJSON:
		return json.Parse(baseURL, strings.NewReader(data))
	case FormatRDF:
		return rdf.Parse(baseURL, strings.NewReader(data))
	default:
		return nil, errors.New("Unsupported feed format")
	}
}
