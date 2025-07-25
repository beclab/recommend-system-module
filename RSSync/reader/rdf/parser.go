package rdf

import (
	"errors"
	"io"

	"bytetrade.io/web3os/RSSync/model"
	"bytetrade.io/web3os/RSSync/reader/xml"
)

// Parse returns a normalized feed struct from a RDF feed.
func Parse(baseURL string, data io.Reader) (*model.Feed, error) {
	feed := new(rdfFeed)
	decoder := xml.NewDecoder(data)
	err := decoder.Decode(feed)
	if err != nil {
		return nil, errors.New("unable to parse RDF feed")
	}

	return feed.Transform(baseURL), nil
}
