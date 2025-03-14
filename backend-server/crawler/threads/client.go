package threads

import (
	"context"
	"encoding/json"
	"html"
	"net/url"
	"os"
	"strings"
	"time"

	"bytetrade.io/web3os/backend-server/common"
	"bytetrade.io/web3os/backend-server/model"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
	"go.uber.org/zap"
)

func findThreadItems(data interface{}) []interface{} {
	if m, ok := data.(map[string]interface{}); ok {
		if items, exists := m["thread_items"]; exists {
			if extMap, ok := items.([]interface{}); ok {
				return extMap
			}
		}

		for _, value := range m {
			if result := findThreadItems(value); result != nil {
				return result
			}
		}
	}

	if slice, ok := data.([]interface{}); ok {
		for _, item := range slice {
			if result := findThreadItems(item); result != nil {
				return result
			}
		}
	}

	return nil
}

func decodeMsg(str string) []interface{} {
	var result map[string]interface{}

	err := json.Unmarshal([]byte(str), &result)
	if err != nil {
		return nil
	}
	require, ok := result["require"].([]interface{})
	if !ok || len(require) == 0 {
		return nil
	}
	require0, ok := require[0].([]interface{})
	if !ok || len(require0) == 0 {
		return nil
	}

	scheduledServerJS := require0[0]
	if scheduledServerJS != "ScheduledServerJS" {
		return nil
	}
	items := findThreadItems(require0)

	return items
}

func getImgageContent(image map[string]interface{}) string {
	imageContent := ""
	imageCandidates, candidatesOK := image["candidates"].([]interface{})
	if candidatesOK && len(imageCandidates) > 0 {
		firstImage, firstImageOK := imageCandidates[0].(map[string]interface{})
		if firstImageOK {
			imageUrl, imageUrlOK := firstImage["url"].(string)
			if imageUrlOK {
				decodedImage, err := url.QueryUnescape(imageUrl)
				if err != nil {
					common.Logger.Error("url decode error", zap.Error(err))
				}
				//todo跨域问题，先加个临时代理服务器https://cors-anywhere.herokuapp.com/  https://cors-proxy.fringe.zone/
				imageContent = imageContent + "<img crossorigin='anonymous' src='https://cors-anywhere.herokuapp.com/" + decodedImage + "' />"
			}
		}
	}
	return imageContent
}

func extractPostID(url string) string {
	url = strings.TrimSuffix(url, "/")
	parts := strings.Split(url, "/")
	return parts[len(parts)-1]
}
func generateEntry(url, rawContent string) *model.Entry {
	entry := new(model.Entry)

	postID := extractPostID(url)
	templateRawData := strings.NewReader(string(rawContent))
	doc, _ := goquery.NewDocumentFromReader(templateRawData)
	userName := ""
	created_at := float64(0)
	fullContent := ""
	videoContent := ""
	imageContent := ""
	title := ""
	doc.Find("script[type='application/json'][data-sjs]").Each(func(i int, s *goquery.Selection) {
		var content string
		content, _ = s.Html()
		decodedStr := html.UnescapeString(content)
		threadItems := decodeMsg(decodedStr)
		if threadItems != nil && len(threadItems) > 0 {
			firstItem, firstOK := threadItems[0].(map[string]interface{})
			if firstOK {
				postItem, postOK := firstItem["post"].(map[string]interface{})
				if postOK {
					code := postItem["code"].(string)
					if code == postID {
						user, userOK := postItem["user"].(map[string]interface{})
						if userOK {
							userName = user["username"].(string)
						}
						caption, captionOK := postItem["caption"].(map[string]interface{})
						if captionOK {
							fullContent, _ = caption["text"].(string)
						}
						created_at, _ = postItem["taken_at"].(float64)
						videos, videoOK := postItem["video_versions"].([]interface{})
						if videoOK && len(videos) > 0 {
							firstVideo, firstVideoOK := videos[0].(map[string]interface{})
							if firstVideoOK {
								videoUrl, videoUrlOK := firstVideo["url"].(string)
								if videoUrlOK {
									videoContent = " <video controls> <source src='" + videoUrl + "' />Your browser does not support the video tag.</video>"
								}

							}
						} else {
							carousels, carouselOK := postItem["carousel_media"].([]interface{})
							if carouselOK {
								for _, carousel := range carousels {
									carouselImage, carouselImageOK := carousel.(map[string]interface{})["image_versions2"].(map[string]interface{})
									if carouselImageOK {
										imageContent = imageContent + getImgageContent(carouselImage)
									}
								}
							} else {
								imageItem, imageOK := postItem["image_versions2"].(map[string]interface{})
								if imageOK {
									imageContent = imageContent + getImgageContent(imageItem)
								}
							}
						}

					}
				}
			}
		}

	})

	doc.Find(`meta[property="og:description"]`).Each(func(i int, s *goquery.Selection) {
		titleContent, exists := s.Attr("content")
		if exists {
			title = titleContent
		}
	})

	entry.FullContent = fullContent + imageContent + videoContent
	entry.Author = userName
	entry.Title = title
	entry.PublishedAt = int64(created_at)

	return entry

}
func Fetch(websiteURL string) *model.Entry {
	rawContent := threadsByheadless(websiteURL)
	if rawContent != "" {
		return generateEntry(websiteURL, rawContent)
	}
	return nil

}

func threadsByheadless(websiteURL string) string {
	var allocCtx context.Context
	var cancelCtx context.CancelFunc
	allocOpts := chromedp.DefaultExecAllocatorOptions[:]

	allocOpts = append(allocOpts,
		chromedp.DisableGPU,
		chromedp.Flag("blink-settings", "imagesEnabled=false"),
		chromedp.Flag("no-first-run", true),
		//chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.UserAgent(`Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.55 Safari/537.36`),
		//chromedp.Flag("accept-language", `zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7,zh-TW;q=0.6`),
	)

	headlessSer := os.Getenv("HEADLESS_SERVER_URL")
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	if headlessSer != "" {
		c, cancelAlloc := chromedp.NewRemoteAllocator(ctx, headlessSer)
		defer cancelAlloc()
		allocCtx, cancelCtx = chromedp.NewContext(c)
	} else {
		c, cancelAlloc := chromedp.NewExecAllocator(ctx, allocOpts...)
		defer cancelAlloc()

		allocCtx, cancelCtx = chromedp.NewContext(c)
	}
	//ctx, cancel := chromedp.NewContext(context.Background())
	defer cancelCtx()
	htmlContent := ""
	common.Logger.Info("threads headless fetch 1 ")
	err := chromedp.Run(allocCtx,
		chromedp.Navigate(websiteURL),
		chromedp.WaitVisible(`[data-pressable-container=true]`, chromedp.ByQuery),
		chromedp.OuterHTML("html", &htmlContent),
	)
	if err != nil {
		common.Logger.Error("threads headless fetch error", zap.String("url", websiteURL), zap.Error(err))
	}
	common.Logger.Info("threads headless fetch end...", zap.Int("content len", len(htmlContent)))

	/*fileWriteErr := os.WriteFile("threads.txt", []byte(htmlContent), 0644)
	if fileWriteErr != nil {
		fmt.Println("Error writing file:", fileWriteErr)
	}*/

	return htmlContent
}
