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

func CheckCookRequired(host string) bool {

	if _, ok := COOKIE_RULES[host]; ok {
		log.Print("check cookie true :", host)
		return true
	}
	return false
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
	//request.Header.Set("Cookie", "auth_token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MjY1Nzk0OTEsImlhdCI6MTcyNTM2OTg5MSwiaXNzIjoia3ViZXNwaGVyZSIsInN1YiI6Im1tY2hvbmcyMDIxIiwidG9rZW5fdHlwZSI6ImFjY2Vzc190b2tlbiIsInVzZXJuYW1lIjoibW1jaG9uZzIwMjEiLCJleHRyYSI6eyJ1bmluaXRpYWxpemVkIjpbInRydWUiXX19.SsN5bZKoZ_eKBo5_I0FlaJcSBXMvbVQjiTPyb6sHq4A; authelia_session=nO1TW3lVuUUQI^_DA1Ez_btJvc5!NPgt")

	client := &http.Client{Timeout: time.Second * 5}
	response, err := client.Do(request)
	if err != nil {
		common.Logger.Error("load cookie info  fail", zap.Error(err))
		return []model.SettingDomainRespModel{}
	}
	defer response.Body.Close()
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
