package client

import (
	"log"
	"net/http"
	"net/url"
	"strings"
)

func GetDownloadFile(downloadUrl string, bflUser string, fileType string) string {

	req, err := http.NewRequest("HEAD", downloadUrl, nil)
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Mobile Safari/537.36")
	RequestAddCookie(req, downloadUrl, bflUser)
	reqClient := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := reqClient.Do(req)
	if err != nil {
		log.Fatalf("Error fetching URL: %v", err)
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

func GetContentAndisposition(downloadUrl string, bflUser string) (string, string) {

	req, err := http.NewRequest("HEAD", downloadUrl, nil)
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Mobile Safari/537.36")
	//z-lib
	//req.Header.Set("Cookie", "siteLanguage=en; refuseChangeDomain=1; remix_userkey=5fad65ce9889bb2ad717d985df7bad46; remix_userid=43395752; hide_regBonusPopup_announcement=true")
	RequestAddCookie(req, downloadUrl, bflUser)
	reqClient := &http.Client{}
	resp, err := reqClient.Do(req)
	if err != nil {
		log.Fatalf("Error fetching URL: %v", err)
	}
	defer resp.Body.Close()
	log.Print("contentdisposition head:", downloadUrl, resp.Header)
	contentType := ""
	reqContentType := ""
	fileName := ""
	if headContentType, ok := resp.Header["Content-Type"]; ok {
		reqContentType = headContentType[0]
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
		log.Print("Content-Disposition filename:", fileName)
	}
	return contentType, fileName
}
