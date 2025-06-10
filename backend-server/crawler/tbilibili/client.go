package tbilibili

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"

	"bytetrade.io/web3os/backend-server/common"
	"bytetrade.io/web3os/backend-server/http/client"
	"bytetrade.io/web3os/backend-server/model"
	"go.uber.org/zap"
)

const (
	userAgent  = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3483.0 Safari/537.36"
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

type DynamicData struct {
	DynamicDesc DynamicDescData `json:"desc"`
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

type ChunkData struct {
	Item ItemData `json:"item"`
}

type Response struct {
	Code int       `json:"code"`
	Data ChunkData `json:"data"`
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

	entry.Author = resp.Data.Item.Modules.AuthorModule.Name
	entry.PublishedAt = resp.Data.Item.Modules.AuthorModule.Pubts
	moduleFullContent := ""
	for _, node := range resp.Data.Item.Modules.DynamicModule.DynamicDesc.DescNodes {
		if node.NodeType == "RICH_TEXT_NODE_TYPE_TEXT" {
			moduleFullContent = moduleFullContent + "<span>" + node.NodeOrigText + "</span>"
		}
		if node.NodeType == "RICH_TEXT_NODE_TYPE_EMOJI" {
			moduleFullContent = moduleFullContent + "<img src='" + node.EmojiNode.EmojiUrl + "' />"
		}
	}
	entry.FullContent = "<div class='mainClass'>" + moduleFullContent + "</div>"
	if len(resp.Data.Item.Orig.Modules.DynamicModule.DynamicDesc.DescNodes) > 0 {
		origFullContent := "<div class='quoteAuthor'><div class='quoteAuthorImg'><img src='" + resp.Data.Item.Orig.Modules.AuthorModule.Face +
			"' /></div><div class='quoteAuthorLabel'>" + resp.Data.Item.Orig.Modules.AuthorModule.Name + "</div></div>"

		for _, node := range resp.Data.Item.Orig.Modules.DynamicModule.DynamicDesc.DescNodes {
			if node.NodeType == "RICH_TEXT_NODE_TYPE_TEXT" {
				origFullContent = origFullContent + "<span>" + node.NodeOrigText + "</span>"
			}
			if node.NodeType == "RICH_TEXT_NODE_TYPE_EMOJI" {
				origFullContent = origFullContent + "<img src='" + node.EmojiNode.EmojiUrl + "' />"
			}
		}
		entry.FullContent = entry.FullContent + "<div class='quoteClass'>" + origFullContent + "</div>"
	}
	entry.Title = common.GetFirstSentence(resp.Data.Item.Modules.DynamicModule.DynamicDesc.DynamicText)

	return entry

}
func fetchPage(bflUser, websiteURL string) *Response {
	//https://api.bilibili.com/x/polymer/web-dynamic/v1/detail?id=1066311590723190791&features=itemOpusStyle,opusBigCover,onlyfansVote,endFooterHidden,decorationCard,onlyfansAssetsV2,ugcDelete,onlyfansQaCard,editable,opusPrivateVisible&timezone_offset=-480&platform=web&gaia_source=main_web&web_location=333.1368&x-bili-device-req-json=%7B%22platform%22:%22web%22,%22device%22:%22pc%22%7D&x-bili-web-req-json=%7B%22spm_id%22:%22333.1368%22%7D&w_rid=7550572b615e9ab897b3684e3e891889&wts=1749198999
	parsedURL, err := url.Parse(websiteURL)
	if err != nil {
		fmt.Println("Error parsing URL:", err)
		return nil
	}
	pathSegments := strings.Split(parsedURL.Path, "/")
	if len(pathSegments) > 1 {
		id := pathSegments[len(pathSegments)-1]
		fetchUrl := fmt.Sprintf("https://api.bilibili.com/x/polymer/web-dynamic/v1/detail?id=%s&features=itemOpusStyle,opusBigCover,onlyfansVote,endFooterHidden,decorationCard,onlyfansAssetsV2,ugcDelete,onlyfansQaCard,editable,opusPrivateVisible", id)
		common.Logger.Info("tbilibili fetch ", zap.String("url", fetchUrl))
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

		var data Response
		if err := json.Unmarshal(body, &data); err != nil {
			return nil
		}
		return &data
	}
	return nil
}
