package parser

import (
	"encoding/xml"
	"strings"

	rxml "bytetrade.io/web3os/RSSync/reader/xml"
)

// List of feed formats.
const (
	FormatRDF     = "rdf"
	FormatRSS     = "rss"
	FormatAtom    = "atom"
	FormatJSON    = "json"
	FormatUnknown = "unknown"
)

// DetectFeedFormat tries to guess the feed format from input data.
func DetectFeedFormat(data string) string {
	if strings.HasPrefix(strings.TrimSpace(data), "{") {
		return FormatJSON
	}

	decoder := rxml.NewDecoder(strings.NewReader(data))

	for {
		token, _ := decoder.Token()
		if token == nil {
			break
		}

		if element, ok := token.(xml.StartElement); ok {
			switch element.Name.Local {
			case "rss":
				return FormatRSS
			case "feed":
				return FormatAtom
			case "RDF":
				return FormatRDF
			}
		}
	}

	return FormatUnknown
}
