package weibo

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
	userAgent  = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Safari/537.36"
	acceptLang = "en-US,en;q=0.9"
)

type EmojiNodeData struct {
	EmojiUrl  string `json:"icon_url"`
	EmojiSize int    `json:"size"`
	EmojiText string `json:"text"`
	EmojiType int    `json:"type"`
}
type RichNode struct {
	NodeOrigText string        `json:"orig_text"`
	NodeText     string        `json:"text"`
	NodeType     string        `json:"type"`
	EmojiNode    EmojiNodeData `json:"emoji"`
}
type DynamicDescData struct {
	DescNodes   []RichNode `json:"rich_text_nodes"`
	DynamicText string     `json:"text"`
}

type DynamicMajorOpusData struct {
	MajorOpusSummary DynamicDescData `json:"summary"`
}
type DynamicMajorData struct {
	MajorOpus DynamicMajorOpusData `json:"opus"`
}

type DynamicData struct {
	DynamicDesc  DynamicDescData  `json:"desc"`
	DynamicMajor DynamicMajorData `json:"major"`
}

type AuthorData struct {
	Name  string `json:"name"`
	Face  string `json:"face"`
	Pubts int64  `json:"pub_ts"`
}

type ModulesData struct {
	AuthorModule  AuthorData  `json:"module_author"`
	DynamicModule DynamicData `json:"module_dynamic"`
}

type OrigData struct {
	Modules ModulesData `json:"modules"`
}

type ItemData struct {
	Modules ModulesData `json:"modules"`
	Orig    OrigData    `json:"orig"`
}

type UserData struct {
	Name string `json:"screen_name"`
}
type RetweetData struct {
	User    UserData `json:"user"`
	Text    string   `json:"text"`
	TextRaw string   `json:"text_raw"`
}

type Response struct {
	User      UserData    `json:"user"`
	Text      string      `json:"text"`
	TextRaw   string      `json:"text_raw"`
	CreatedAt string      `json:"created_at"`
	Retweet   RetweetData `json:"retweeted_status"`
}

func Fetch(bflUser, websiteURL string) *model.Entry {
	resp := fetchPage(bflUser, websiteURL)
	if resp != nil {
		return generateEntry(resp)
	}
	return nil

}

func generateEntry(resp *Response) *model.Entry {
	entry := new(model.Entry)

	entry.Author = resp.User.Name
	publicDate, parseErr := date.Parse(resp.CreatedAt)
	if parseErr != nil {
		common.Logger.Error("date parse err:", zap.Error(parseErr))
	}
	entry.PublishedAt = publicDate.Unix()
	entry.FullContent = "<div class='mainClass'>" + resp.Text + "</div>"
	if resp.Retweet.Text != "" {
		entry.FullContent = entry.FullContent + "<div class='quoteClass'>" + resp.Retweet.Text + "</div>"
	}

	entry.Title = common.GetFirstSentence(resp.TextRaw)
	common.Logger.Info("weibo entry", zap.String("author", string(entry.Author)), zap.String("title", string(entry.Title)))
	return entry

}
func fetchPage(bflUser, websiteURL string) *Response {
	//https://weibo.com/ajax/statuses/show?id=Pu2xetkaJ&locale=zh-CN&isGetLongText=true
	parsedURL, err := url.Parse(websiteURL)
	if err != nil {
		fmt.Println("Error parsing URL:", err)
		return nil
	}
	pathSegments := strings.Split(parsedURL.Path, "/")
	if len(pathSegments) > 1 {
		id := pathSegments[len(pathSegments)-1]
		fetchUrl := fmt.Sprintf("https://weibo.com/ajax/statuses/show?id=%s&locale=zh-CN&isGetLongText=true", id)
		common.Logger.Info("weibo fetch ", zap.String("url", fetchUrl))
		clt := client.NewClientWithConfig(fetchUrl)
		clt.WithBflUser(bflUser)
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
		//common.Logger.Info("fetch weibo result", zap.String("url", string(body)))
		var data Response
		if err := json.Unmarshal(body, &data); err != nil {
			return nil
		}
		return &data
	}
	return nil
}
