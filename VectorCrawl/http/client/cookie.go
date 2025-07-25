package client

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"bytetrade.io/web3os/vector-crawl/common"
	"bytetrade.io/web3os/vector-crawl/model"
	"go.uber.org/zap"
)

func GetPrimaryDomain(u string) (string, string) {
	parsedURL, err := url.Parse(u)
	if err != nil {
		return "", ""
	}
	host := parsedURL.Hostname()

	parts := strings.Split(host, ".")
	if len(parts) >= 2 {
		return host, strings.Join(parts[len(parts)-2:], ".")
	}
	return host, host
}

func getParentDomain(domain string) string {

	parts := strings.Split(domain, ".")
	if len(parts) > 2 {
		return strings.Join(parts[1:], ".")
	}
	return ""
}

func LoadCookieInfoManager(bflUser, domain, primaryDomain string) []model.SettingDomainRespModel {
	initDomain := domain
	cookieList := make([]model.SettingDomainRespModel, 0)
	for {
		if initDomain == domain {
			list := LoadCookieInfo(bflUser, domain)
			if len(list) > 0 {
				cookieList = append(cookieList, list...)
			}
		}
		list := LoadCookieInfo(bflUser, "."+domain)
		if len(list) > 0 {
			cookieList = append(cookieList, list...)
		}
		if domain == primaryDomain {
			break
		}
		parentDomain := getParentDomain(domain)
		if domain == parentDomain {
			break
		}
		domain = parentDomain
	}
	return cookieList
}
func LoadCookieInfo(bflUser string, host string) []model.SettingDomainRespModel {
	reqData := model.SettingReqModel{Domain: host}
	reqJsonByte, err := json.Marshal(reqData)
	if err != nil {
		common.Logger.Error("add cookie json marshal  fail", zap.Error(err))
	}
	settingUrl := "http://system-server.user-system-" + bflUser + "/legacy/v1alpha1/service.settings/v1/api/cookie/retrieve"
	common.Logger.Info("start load cookie info ", zap.String("host", host))
	request, _ := http.NewRequest("POST", settingUrl, bytes.NewBuffer(reqJsonByte))
	request.Header.Set("Content-Type", "application/json")
	//request.Header.Set("Cookie", "auth_token=eyJhbGciOiJIUzUxMiJ9.eyJleHAiOjE3ODA2NDYxNzksImlhdCI6MTc0OTExMDE3OSwidXNlcm5hbWUiOiJtbWNob25nMjAyMSIsImdyb3VwcyI6WyJsbGRhcF9hZG1pbiJdfQ.JqdLAMd4SAWFdatKSyXninS98DmTvy9FXn_ma6sKHXBM1YrOiQGlZhZdY3OjU9-rdY3pfj8tqPvYhiFZWa_0nw; authelia_session=Dhwb#bgNT4SqteKabN$GN-VIgq1Lm^8U")

	client := &http.Client{Timeout: time.Second * 5}
	response, err := client.Do(request)
	if err != nil {
		common.Logger.Error("load cookie info  fail", zap.Error(err))
		return []model.SettingDomainRespModel{}
	}
	if response != nil {
		defer response.Body.Close()
	}
	responseBody, _ := io.ReadAll(response.Body)
	log.Print("get cookid result url :", settingUrl, string(responseBody))
	var resObj model.SettingResponseModel
	if err := json.Unmarshal(responseBody, &resObj); err != nil {
		log.Print("json decode failed, err", err)
		return []model.SettingDomainRespModel{}
	}
	if resObj.Code == 0 {
		return resObj.Data
	}
	return []model.SettingDomainRespModel{}

}

func RequestAddCookie(request *http.Request, reqUrl string, bflUser string) {
	urlDomain, urlPrimaryDomain := GetPrimaryDomain(reqUrl)
	if urlDomain != "" && bflUser != "" {
		domainList := LoadCookieInfoManager(bflUser, urlDomain, urlPrimaryDomain)
		for _, domain := range domainList {
			for _, record := range domain.Records {
				if strings.HasPrefix(record.Domain, ".") {
					if len(record.Domain)-len(urlDomain) > 1 {
						print("skip cookie domain:", record.Domain)
						continue
					}
				} else {
					if record.Domain != urlDomain {
						print("skip cookie domain2:", record.Domain)
						continue
					}
				}
				cookieVal := record.Value
				if urlPrimaryDomain != "zhihu.com" {
					cookieVal = url.QueryEscape(record.Value)
				}
				cookie := &http.Cookie{
					Name:    record.Name,
					Value:   cookieVal,
					Path:    record.Path,
					Domain:  record.Domain,
					Expires: time.Unix(int64(record.Expires), 0),
				}
				request.AddCookie(cookie)
			}
		}
	}
}
