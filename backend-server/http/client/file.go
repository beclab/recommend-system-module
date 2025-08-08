package client

import (
	"log"
	"net/http"
	"net/url"
	"strings"

	"bytetrade.io/web3os/backend-server/common"
	"go.uber.org/zap"
)

func GetDownloadFile(downloadUrl string, bflUser string, fileType string) string {

	req, err := http.NewRequest("HEAD", downloadUrl, nil)
	if err != nil {
		common.Logger.Error("Error creating request", zap.Error(err))
		return ""
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Mobile Safari/537.36")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Connection", "Keep-Alive")

	RequestAddCookie(req, downloadUrl, bflUser)
	reqClient := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := reqClient.Do(req)
	if err != nil {
		common.Logger.Error("Error fetching URL", zap.Error(err))
		return ""
	}
	defer resp.Body.Close()
	log.Print("downloadfile head:", downloadUrl, resp.Header)
	//redirectURL := resp.Header.Get("Location")
	if values, ok := resp.Header["Location"]; ok {
		return GetFileNameFromUrl(values[0], fileType)
	}
	return GetFileNameFromUrl(downloadUrl, fileType)
}

func GetFileNameFromUrl(url string, fileType string) string {
	lastSlashIndex := strings.LastIndex(url, "/")
	fileName := url[lastSlashIndex+1:]
	if fileType == "ebook" && !strings.HasSuffix(fileName, ".epub") {
		fileName = fileName + ".epub"
	}
	if fileType == "pdf" && !strings.HasSuffix(fileName, ".pdf") {
		fileName = fileName + ".pdf"
	}
	return fileName
}

func particularUrlAnalysis(downloadUrl string) (string, string) {
	contentType := ""
	fileName := ""
	urlDomain := common.Domain(downloadUrl)
	if strings.Contains(urlDomain, "manybooks.net") {
		cleanPath := strings.Trim(downloadUrl, "/")
		parts := strings.Split(cleanPath, "/")
		lastPart := parts[len(parts)-1]
		if lastPart == "pdf" {
			contentType = "pdf"
			fileName = parts[len(parts)-2] + ".pdf"
		}
		if lastPart == "epub" {
			contentType = "ebook"
			fileName = parts[len(parts)-2] + ".epub"
		}

	}
	return contentType, fileName
}

func GetContentAndisposition(downloadUrl string, bflUser string) (string, string) {
	contentType := ""
	fileName := ""
	contentType, fileName = particularUrlAnalysis(downloadUrl)
	if contentType != "" {
		return contentType, fileName
	}

	req, err := http.NewRequest("HEAD", downloadUrl, nil)
	if err != nil {
		common.Logger.Error("Error creating request", zap.Error(err))
		return contentType, fileName
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Mobile Safari/537.36")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Connection", "keep-alive")
	//req.Header.Set("Accept-Encoding", "identity")
	//req.Header.Set("Accept-Encoding", "gzip, deflate, br")

	//z-lib
	//req.Header.Set("Cookie", "siteLanguage=en; refuseChangeDomain=1; remix_userkey=5fad65ce9889bb2ad717d985df7bad46; remix_userid=43395752; hide_regBonusPopup_announcement=true")
	log.Print("start contentdisposition head:")
	RequestAddCookie(req, downloadUrl, bflUser)
	reqClient := &http.Client{}
	resp, err := reqClient.Do(req)
	if err != nil {
		common.Logger.Error("Error fetching URL:", zap.Error(err))
		return contentType, fileName
	}
	defer resp.Body.Close()
	log.Print("contentdisposition head:", downloadUrl, resp.Header["Content-Type"])
	reqContentType := ""

	if headContentType, ok := resp.Header["Content-Type"]; ok {
		reqContentType = headContentType[0]
	}
	if strings.HasPrefix(reqContentType, "text/html") {
		contentType = "text/html"
	}
	if reqContentType == "application/pdf" {
		contentType = "pdf"
	}
	if reqContentType == "application/epub+zip" {
		contentType = "ebook"
	}
	if strings.HasPrefix(reqContentType, "audio/") {
		contentType = "audio"
	}
	if strings.HasPrefix(reqContentType, "video/") {
		contentType = "video"
	}
	if contentDisposition, ok := resp.Header["Content-Disposition"]; ok {
		log.Print("Content-Disposition:", contentDisposition)
		parts := strings.Split(contentDisposition[0], ";")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, "filename*=") {
				encodedPart := part[len("filename*="):]
				langAndEncoding := strings.SplitN(encodedPart, "'", 3)
				if len(langAndEncoding) == 3 {
					file, err := url.QueryUnescape(langAndEncoding[2])
					if err == nil {
						fileName = file
					}
				}
				if fileName != "" {
					break
				}
			} else if strings.HasPrefix(part, "filename=") {
				fileName = strings.Trim(part[len("filename="):], `"`)
			}
		}
	}
	log.Print("Content-Disposition filename:", fileName, "contentType:", contentType)
	return contentType, fileName
}
