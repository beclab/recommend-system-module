package common

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func FirstNonEmptyStr(options ...string) string {
	for _, s := range options {
		if strings.TrimSpace(s) != "" {
			return s
		}
	}
	return ""
}

func GetFirstSentence(text string) string {
	sentenceEndings := []string{".", "。", "!", "?", "？", "<br>"}
	minIndex := len(text)
	if minIndex > 50 {
		minIndex = 50
	}

	for _, ending := range sentenceEndings {
		index := strings.Index(text, ending)
		if index != -1 && index < minIndex {
			minIndex = index
		}
	}
	firstSentence := text[:minIndex]
	if len(firstSentence) < len(text) {
		firstSentence = firstSentence + "..."
	}

	return firstSentence
}

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
