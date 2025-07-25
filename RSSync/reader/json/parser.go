package json

import (
	"encoding/json"
	"errors"
	"io"

	"bytetrade.io/web3os/RSSync/model"
)

// Parse returns a normalized feed struct from a JSON feed.
func Parse(baseURL string, data io.Reader) (*model.Feed, error) {
	feed := new(jsonFeed)
	decoder := json.NewDecoder(data)
	if err := decoder.Decode(&feed); err != nil {
		return nil, errors.New("unable to parse JSON Feed")
	}

	return feed.Transform(baseURL), nil
}
