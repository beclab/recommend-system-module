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

type ThreadPostEmbedRecordRecordValueEmbeds struct {
	EmbedType  string                 `json:"$type"`
	EmbedMedia []ThreadPostEmbedImage `json:"media"`
}

type ThreadPostEmbedRecordRecordValue struct {
	CreatedAt string `json:"createdAt"`
	Text      string `json:"text"`
	//Embed     ThreadPostEmbedRecordRecordValueEmbed `json:"embed"`
}
type ThreadPostEmbedRecordRecord struct {
	Author ThreadPostAuthor                 `json:"author"`
	Value  ThreadPostEmbedRecordRecordValue `json:"value"`
	Embeds []ThreadPostMedia                `json:"embeds"`
}
type ThreadPostEmbedRecord struct {
	EmbedRecordRecord ThreadPostEmbedRecordRecord `json:"record"`
}
type ThreadPostMedia struct {
	EmbedType     string                 `json:"$type"`
	EmbedPlaylist string                 `json:"playlist"`
	EmbedImages   []ThreadPostEmbedImage `json:"images"`
}
type ThreadPostEmbed struct {
	EmbedType     string                 `json:"$type"`
	EmbedMedia    ThreadPostMedia        `json:"media"`
	EmbedRecord   ThreadPostEmbedRecord  `json:"record"`
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

type ThreadPostReply struct {
	ThreadPost ThreadPost        `json:"post"`
	Replies    []ThreadPostReply `json:"replies"`
}

type Thread struct {
	Post    ThreadPost        `json:"post"`
	Replies []ThreadPostReply `json:"replies"`
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

func getEmbedContent(embed ThreadPostEmbed) string {
	content := ""
	if embed.EmbedType == "app.bsky.embed.images#view" {
		for _, image := range embed.EmbedImages {
			content = content + "<img src='" + image.Thumb + "' /><br>"
		}
	} else if embed.EmbedType == "app.bsky.embed.video#view" {
		content = content + "<video controls=''><source src='" + embed.EmbedPlaylist + "' type='application/x-mpegURL'>Your browser does not support the video tag.</video>"
	} else if embed.EmbedType == "app.bsky.embed.recordWithMedia#view" {
		if embed.EmbedMedia.EmbedType == "app.bsky.embed.images#view" {
			for _, image := range embed.EmbedMedia.EmbedImages {
				content = content + "<img src='" + image.Thumb + "' /><br>"
			}
		} else if embed.EmbedMedia.EmbedType == "app.bsky.embed.video#view" {
			content = content + "<video controls=''><source src='" + embed.EmbedMedia.EmbedPlaylist + "' type='application/x-mpegURL'>Your browser does not support the video tag.</video>"
		}
	}
	return content
}

func getPostEmbedContent(embed ThreadPostMedia) string {
	content := ""
	if embed.EmbedType == "app.bsky.embed.images#view" {
		for _, image := range embed.EmbedImages {
			content = content + "<img src='" + image.Thumb + "' /><br>"
		}
	} else if embed.EmbedType == "app.bsky.embed.video#view" {
		content = content + "<video controls=''><source src='" + embed.EmbedPlaylist + "' type='application/x-mpegURL'>Your browser does not support the video tag.</video>"
	}
	return content
}

func getReplyContent(replies []ThreadPostReply, author string) string {
	content := ""
	for _, reply := range replies {
		if reply.ThreadPost.Author.Name == author {
			content = content + "<div class='bskyReplyClass'>" + author + "<br>" + strings.ReplaceAll(reply.ThreadPost.Record.Text, "\n", "<br>") + "</div>"
			content = content + getEmbedContent(reply.ThreadPost.Embed) + "</div>"
		}
		for _, reply2 := range reply.Replies {
			if reply2.ThreadPost.Author.Name == author {
				content = content + "<div class='bskyReplyClass'>" + author + "<br>" + strings.ReplaceAll(reply2.ThreadPost.Record.Text, "\n", "<br>") + "</div>"
				content = content + getEmbedContent(reply2.ThreadPost.Embed) + "</div>"
			}
			if len(reply2.Replies) > 0 {
				content = content + getReplyContent(reply2.Replies, author)
			}
		}
	}
	return content
}
func generateEntry(resp *Response) *model.Entry {
	entry := new(model.Entry)

	entry.Author = resp.Thread.Post.Author.Name
	entry.FullContent = "<p>" + "<div class='bskyMainClass'>" + strings.ReplaceAll(resp.Thread.Post.Record.Text, "\n", "<br>") + "</p>"
	/*if resp.Thread.Post.Embed.EmbedType == "app.bsky.embed.images#view" {
		for _, image := range resp.Thread.Post.Embed.EmbedImages {
			entry.FullContent = entry.FullContent + "<img src='" + image.Thumb + "' /><br>"
		}
	} else if resp.Thread.Post.Embed.EmbedType == "app.bsky.embed.video#view" {
		entry.FullContent = entry.FullContent + "<video controls=''><source src='" + resp.Thread.Post.Embed.EmbedPlaylist + "' type='application/x-mpegURL'>Your browser does not support the video tag.</video>"
	} else if resp.Thread.Post.Embed.EmbedType == "app.bsky.embed.recordWithMedia#view" {
		entry.FullContent = entry.FullContent + "<video controls=''><source src='" + resp.Thread.Post.Embed.EmbedPlaylist + "' type='application/x-mpegURL'>Your browser does not support the video tag.</video>"
	}*/
	entry.FullContent = entry.FullContent + getEmbedContent(resp.Thread.Post.Embed) + "</div>"
	if len(resp.Thread.Post.Embed.EmbedRecord.EmbedRecordRecord.Embeds) > 0 {
		quoteContent := resp.Thread.Post.Embed.EmbedRecord.EmbedRecordRecord.Author.Name + "<br>" + strings.ReplaceAll(resp.Thread.Post.Embed.EmbedRecord.EmbedRecordRecord.Value.Text, "\n", "<br>") + "<br>"
		for _, quoteEmbed := range resp.Thread.Post.Embed.EmbedRecord.EmbedRecordRecord.Embeds {
			quoteContent = quoteContent + getPostEmbedContent(quoteEmbed) + "<br>"
		}
		entry.FullContent = entry.FullContent + "<div class='bskyQuoteClass' style='padding-left: 10ch;'>" + quoteContent + "</div>"
	}
	/*for _, reply := range resp.Thread.Replies {
		if reply.ThreadPost.Author.Name == entry.Author {
			entry.FullContent = entry.FullContent + "<div class='bskyReplyClass'>" + entry.Author + "<br>" + strings.ReplaceAll(reply.ThreadPost.Record.Text, "\n", "<br>") + "</div>"
			entry.FullContent = entry.FullContent + getEmbedContent(reply.ThreadPost.Embed) + "</div>"
		}
		for _, reply2 := range reply.Replies {
			if reply2.ThreadPost.Author.Name == entry.Author {
				entry.FullContent = entry.FullContent + "<div class='bskyReplyClass'>" + entry.Author + "<br>" + strings.ReplaceAll(reply2.ThreadPost.Record.Text, "\n", "<br>") + "</div>"
				entry.FullContent = entry.FullContent + getEmbedContent(reply2.ThreadPost.Embed) + "</div>"
			}
		}
	}*/
	entry.FullContent = entry.FullContent + getReplyContent(resp.Thread.Replies, entry.Author)
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
