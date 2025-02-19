package wolaiapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3483.0 Safari/537.36"
)

// {"pageId":"smmLWANwc5QJNXtsW3nTvv","limit":100,"position":{"stack":[]},"chunkNumber":0}
type WolaiPositionReq struct {
	Stack []string `json:"stack"`
}
type WolaiReq struct {
	PageId      string           `json:"pageId"`
	Limit       int              `json:"limit"`
	Position    WolaiPositionReq `json:"position"`
	ChunkNumber int              `json:"chunkNumber"`
}
type BlockAttributes struct {
	/*
		 "title": [
			[
				"自学分为 5 个阶段",
				[
					[
						"B"
					]
				]
			],
			[
				"。",
				[]
			]
		]
	*/
	Title  []interface{} `json:"title"`
	Source []string      `json:"source"`
}
type BlockValue struct {
	Id         string          `json:"id"`
	Active     bool            `json:"active"`
	Attributes BlockAttributes `json:"attributes"`
	CloneCount int             `json:"clone_count"`
	Status     int             `json:"status"`
	SubNodes   []string        `json:"sub_nodes"`
	Type       string          `json:"type"`
	ParentType string          `json:"parent_type"`
}
type BlockData struct {
	Role  string     `json:"role"`
	Value BlockValue `json:"value"`
}
type ChunkData struct {
	Block map[string]BlockData `json:"block"`
}

type Response struct {
	Code    int       `json:"code"`
	Data    ChunkData `json:"data"`
	Message string    `json:"message"`
}

func ExtractPageIDFromURL(uri string) string {
	uri = strings.TrimSuffix(uri, "/")
	parts := strings.Split(uri, "/")
	lastPart := parts[len(parts)-1]
	return lastPart
}

func getAttributeTitle(title []interface{}) string {
	switch v := title[0].(type) {
	case []interface{}:
		if str, ok := v[0].(string); ok {
			return str
		}
	case string:
		if str, ok := title[0].(string); ok {
			return str
		}
	}
	return ""
}

func getTextHtml(titles []interface{}, head string) string {
	content := ""
	for _, title := range titles {
		switch v := title.(type) {
		case []interface{}:
			if str, ok := v[0].(string); ok {
				spanContent := "<span>" + str + "</span>"
				if head != "" {
					spanContent = "<" + head + ">" + spanContent + "</" + head + ">"
				}
				content += spanContent
			}
			if len(v) == 1 {
			}
		}
	}
	return content
}

func getImgHtml(urls []string) string {
	content := ""
	for _, url := range urls {
		content += "<figure> <img src=" + url + " style='width: 100%'></img></figure>"
	}
	return content
}

func generateDiv(block BlockValue) string {
	/*
		midHeader
		tinyHeader
		enumList
		bullList
		text
		image
	*/
	divContent := ""
	//<span class="inline-wrap">【时间】2024<span class="jill"></span>年<span class="jill"></span>12<span class="jill"></span>月<span class="jill"></span>3<span class="jill"></span>日（周二）14:00-16:00</span>
	switch block.Type {
	case "text":
		divContent = getTextHtml(block.Attributes.Title, "")
	case "image":
		divContent = getImgHtml(block.Attributes.Source)
	case "midHeader":
		divContent = getTextHtml(block.Attributes.Title, "h2")
	case "tinyHeader":
		divContent = getTextHtml(block.Attributes.Title, "h4")
	case "enumList":
		divContent = getTextHtml(block.Attributes.Title, "")
	case "bullList":
		divContent = getTextHtml(block.Attributes.Title, "")
	}

	html := fmt.Sprintf(`
        <div>
			<div>
				%s
			</div>
		</div>
    `, divContent)
	return html

}

func generateHtml(pageID string, chunkData ChunkData) string {
	title := ""
	bodyContent := ""
	if chunk, exists := chunkData.Block[pageID]; exists {
		subNodes := chunk.Value.SubNodes
		title = getAttributeTitle(chunk.Value.Attributes.Title)
		for _, node := range subNodes {
			if nodeChunk, nodeExists := chunkData.Block[node]; nodeExists {
				bodyContent += generateDiv(nodeChunk.Value)
			}
		}
	}

	html := fmt.Sprintf(`
        <!DOCTYPE html>
        <html lang="zh">
		<head>
			<title>%s</title>
		</head>
        <body>
		  	<div class="page-body">
           		%s
			</div>
        </body>
        </html>
    `, title, bodyContent)

	return html
}

func FetchPage(pageID string) string {
	//jsonData := fmt.Sprintf(`{"pageId":"%s","limit":100,"position":{"stack":[]},"chunkNumber":0}`, pageID)
	positionData := WolaiPositionReq{
		Stack: []string{},
	}
	reqData := WolaiReq{
		PageId:      pageID,
		Limit:       100,
		Position:    positionData,
		ChunkNumber: 0,
	}
	jsonByte, _ := json.Marshal(reqData)
	// 创建新的 POST 请求
	req, err := http.NewRequest("POST", "https://api.wolai.com/v1/pages/getPageChunks?pageId="+pageID, bytes.NewBuffer(jsonByte))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return ""
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)    // 替换为你的 User-Agent
	req.Header.Set("Accept-Language", "zh-CN") // 设置接受的语言
	//req.Header.Set("Cookie", "acw_tc=1-67ad6510-16e0089e-59361102db04cc9fa9d6e7cbd80cab202391148a9f7f; cna=SxExIIjMuRACAXWTMNWwapuP; isg=BOvrmoPKCJ8DJVRMbtPIykOgegnVAP-CuRuTFV1rHSol_ANe5dJJ0IIeVzySb1d6; token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiJiczFuOHZWWnF0VVFyNFN0RnRmaTZKIiwiaWF0IjoxNzM5NDE3NDA1LCJleHAiOjIwNTQ3Nzc0MDV9.0az7LI3sW5ZY3XiKsBuQz3dGsmlDUROTpZkHAQ_eO2Q; wolai_client_id=i9uekkBry2oWcwrDgMUDrr")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	responseBody, _ := io.ReadAll(resp.Body)

	var data Response
	if err := json.Unmarshal(responseBody, &data); err != nil {
		return ""
	}
	return generateHtml(pageID, data.Data)
}
