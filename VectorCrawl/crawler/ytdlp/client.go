package ytdlp

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"bytetrade.io/web3os/vector-crawl/common"
	"bytetrade.io/web3os/vector-crawl/model"
	"go.uber.org/zap"
)

func IsMetaFromYtdlp(url string) bool {
	mediaList := map[string]struct{}{
		"bilibili.com": {},
		"youtube.com":  {},
		"vimeo.com":    {},
		"rumble.com":   {},
	}

	for domain := range mediaList {
		if strings.Contains(url, domain) {
			return true
		}
	}
	return false
}

func Fetch(bfl_user, url string) *model.Entry {
	apiUrl := common.YTDLPApiUrl() + "/v1/get_metadata?url=" + url + "&bfl_user=" + bfl_user
	common.Logger.Info("load meta from ytdlp", zap.String("url", apiUrl))
	client := &http.Client{Timeout: time.Second * 50}
	res, err := client.Get(url)
	if err != nil {
		common.Logger.Error("load ytdlp meta error", zap.Error(err))
		return nil
	}
	if res.StatusCode != 200 {
		common.Logger.Error("load ytdlp meta error")
		return nil
	}
	if res != nil {
		defer res.Body.Close()
	}
	body, _ := io.ReadAll(res.Body)

	var resObj model.EntryFetchResponseModel
	if err := json.Unmarshal(body, &resObj); err != nil {
		common.Logger.Error("load ytdlp meta ,json decode failed, err", zap.Error(err))
		return nil
	}

	return &resObj.Data

}
