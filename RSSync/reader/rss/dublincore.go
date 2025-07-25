package rss

// DublinCoreElement represents Dublin Core XML elements.
type DublinCoreElement struct {
	DublinCoreDate    string `xml:"http://purl.org/dc/elements/1.1/ date"`
	DublinCoreCreator string `xml:"http://purl.org/dc/elements/1.1/ creator"`
	DublinCoreContent string `xml:"http://purl.org/rss/1.0/modules/content/ encoded"`
}
