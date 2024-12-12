package parser

import (
	"errors"
	"strings"

	"bytetrade.io/web3os/backend-server/common"
	"bytetrade.io/web3os/backend-server/model"
	"bytetrade.io/web3os/backend-server/reader/atom"
	"bytetrade.io/web3os/backend-server/reader/json"
	"bytetrade.io/web3os/backend-server/reader/rdf"
	"bytetrade.io/web3os/backend-server/reader/rss"
	"go.uber.org/zap"
)

// ParseFeed analyzes the input data and returns a normalized feed object.
func ParseFeed(baseURL, data string) (*model.Feed, error) {
	format := DetectFeedFormat(data)
	common.Logger.Info("parse feed,", zap.String("format:", format))
	switch format {
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
