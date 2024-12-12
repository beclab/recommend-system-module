package common

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func GetImageUrlFromContent(content string) string {
	imageUrl := ""
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	if err == nil {
		doc.Find("img").Each(func(i int, s *goquery.Selection) {
			img, _ := s.Attr("src")
			if strings.HasPrefix(img, "http") {
				imageUrl = img
			}
		})
	}
	return imageUrl
}
