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

	"bytetrade.io/web3os/backend-server/common"
	"bytetrade.io/web3os/backend-server/model"
	"go.uber.org/zap"
)

var COOKIE_RULES = map[string]string{
	"bilibili.com": "recommend",
	"spotify.com":  "required",
	"reuters.com":  "required",
	"wsj.com":      "required",
	"ft.com":       "required",
}

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

func CheckCookRequired(host string) bool {

	if _, ok := COOKIE_RULES[host]; ok {
		log.Print("check cookie true :", host)
		return true
	}
	return false
}

func LoadCookieInfoManager(domain, primaryDomain string) []model.SettingDomainRespModel {
	initDomain := domain
	cookieList := make([]model.SettingDomainRespModel, 0)
	for {
		if initDomain == domain {
			list := LoadCookieInfo(domain)
			if len(list) > 0 {
				cookieList = append(cookieList, list...)
			}
		}
		list := LoadCookieInfo("." + domain)
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
func LoadCookieInfo(host string) []model.SettingDomainRespModel {
	reqData := model.SettingReqModel{Domain: host}
	reqJsonByte, err := json.Marshal(reqData)
	if err != nil {
		common.Logger.Error("add cookie json marshal  fail", zap.Error(err))
	}
	settingUrl := common.SettingApiUrl()
	common.Logger.Info("start load cookie info ", zap.String("host", host))
	request, _ := http.NewRequest("POST", settingUrl, bytes.NewBuffer(reqJsonByte))
	request.Header.Set("Content-Type", "application/json")
	//request.Header.Set("Cookie", "auth_token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MjgzMDQ5MTYsImlhdCI6MTcyNzA5NTMxNiwiaXNzIjoia3ViZXNwaGVyZSIsInN1YiI6Im1tY2hvbmcyMDIxIiwidG9rZW5fdHlwZSI6ImFjY2Vzc190b2tlbiIsInVzZXJuYW1lIjoibW1jaG9uZzIwMjEiLCJleHRyYSI6eyJ1bmluaXRpYWxpemVkIjpbInRydWUiXX19.Namzne2m9ZwChnFygILh18XUv-3wh6_4vnACykquOVc; authelia_session=F9im*hFguSW5CwTkXA9-7N!6zIgMcSTq")

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
