package bskyapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"

	"bytetrade.io/web3os/backend-server/common"
	"bytetrade.io/web3os/backend-server/http/client"
	"bytetrade.io/web3os/backend-server/model"
	"bytetrade.io/web3os/backend-server/reader/date"
	"go.uber.org/zap"
)

const (
	userAgent  = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3483.0 Safari/537.36"
	acceptLang = "en-US,en;q=0.9"
)

type ThreadPostRecord struct {
	CreatedAt string `json:"createdAt"`
	Text      string `json:"text"`
}

type ThreadPostEmbedImage struct {
	Thumb string `json:"thumb"`
}
type ThreadPostEmbed struct {
	EmbedType     string                 `json:"$type"`
	EmbedPlaylist string                 `json:"playlist"`
	EmbedImages   []ThreadPostEmbedImage `json:"images"`
}

type ThreadPostAuthor struct {
	Name string `json:"displayName"`
}

type ThreadPost struct {
	Author ThreadPostAuthor `json:"author"`
	Record ThreadPostRecord `json:"record"`
	Embed  ThreadPostEmbed  `json:"embed"`
}

type Thread struct {
	Post ThreadPost `json:"post"`
}
type Response struct {
	Thread Thread `json:"thread"`
}

func Fetch(websiteURL string) *model.Entry {
	resp := fetchPage(websiteURL)
	if resp != nil {
		return generateEntry(resp)
	}
	return nil

}

func generateEntry(resp *Response) *model.Entry {
	entry := new(model.Entry)

	entry.FullContent = "<p>" + strings.ReplaceAll(resp.Thread.Post.Record.Text, "\n", "<br>") + "</p>"
	if resp.Thread.Post.Embed.EmbedType == "app.bsky.embed.images#view" {
		for _, image := range resp.Thread.Post.Embed.EmbedImages {
			entry.FullContent = entry.FullContent + "<img src='" + image.Thumb + "' /><br>"
		}
	} else if resp.Thread.Post.Embed.EmbedType == "app.bsky.embed.video#view" {
		entry.FullContent = entry.FullContent + "<video controls=''><source src='" + resp.Thread.Post.Embed.EmbedPlaylist + "' type='application/x-mpegURL'>Your browser does not support the video tag.</video>"
	}

	entry.Author = resp.Thread.Post.Author.Name
	entry.Title = common.GetFirstSentence(resp.Thread.Post.Record.Text)
	publishedAt, dateParseErr := date.Parse(resp.Thread.Post.Record.CreatedAt)
	if dateParseErr == nil {
		entry.PublishedAt = publishedAt.Unix()
	}

	return entry

}
func fetchPage(websiteURL string) *Response {
	//https://bsky.app/profile/plantepigenetics.ch/post/3ljfvtgf63c24 转换成
	//https://public.api.bsky.app/xrpc/app.bsky.feed.getPostThread?uri=at://plantepigenetics.ch/app.bsky.feed.post/3ljfvtgf63c24&depth=10
	parsedURL, err := url.Parse(websiteURL)
	if err != nil {
		fmt.Println("Error parsing URL:", err)
		return nil
	}
	profile := ""
	postID := ""
	segments := strings.Split(parsedURL.Path, "/")
	for i := 0; i < len(segments); i++ {
		if segments[i] == "profile" {
			profile = segments[i+1]
		}
		if segments[i] == "post" {
			postID = segments[i+1]
		}
	}
	encodedURI := "at://" + profile + "/app.bsky.feed.post/" + postID
	fetchUrl := fmt.Sprintf("https://public.api.bsky.app/xrpc/app.bsky.feed.getPostThread?uri=%s&depth=10", encodedURI)
	common.Logger.Info("bsky fetch ", zap.String("url", fetchUrl))
	clt := client.NewClientWithConfig(fetchUrl)
	clt.WithUserAgent(userAgent)

	response, err := clt.Get()
	if err != nil {
		common.Logger.Error("crawling entry rawContent error ", zap.String("url", websiteURL), zap.Error(err))
		return nil
	}

	if response.HasServerFailure() {
		common.Logger.Error("crawling entry rawContent error ", zap.String("url", websiteURL))
		return nil
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		common.Logger.Error("crawling entry rawContent error ", zap.String("url", websiteURL), zap.Error(err))
		return nil
	}

	var data Response
	if err := json.Unmarshal(body, &data); err != nil {
		return nil
	}
	return &data
}
