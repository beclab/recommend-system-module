package rdf

// DublinCoreFeedElement represents Dublin Core feed XML elements.
type DublinCoreFeedElement struct {
	DublinCoreCreator string `xml:"http://purl.org/dc/elements/1.1/ channel>creator"`
}

// DublinCoreEntryElement represents Dublin Core entry XML elements.
type DublinCoreEntryElement struct {
	DublinCoreDate    string `xml:"http://purl.org/dc/elements/1.1/ date"`
	DublinCoreCreator string `xml:"http://purl.org/dc/elements/1.1/ creator"`
	DublinCoreContent string `xml:"http://purl.org/rss/1.0/modules/content/ encoded"`
}
