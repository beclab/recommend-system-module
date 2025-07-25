package twitter

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"bytetrade.io/web3os/vector-crawl/common"
	"bytetrade.io/web3os/vector-crawl/model"
	"go.uber.org/zap"
)

func Fetch(bfl_user, twitterID, url string) *model.Entry {
	apiUrl := common.DownloadApiUrl() + "/twitter/fetch-content?twitter_id=" + twitterID + "&url=" + url + "&bfl_user=" + bfl_user
	client := &http.Client{Timeout: time.Second * 120}
	res, err := client.Get(apiUrl)
	if err != nil {
		common.Logger.Error("fetch twitter content error", zap.String("id", twitterID), zap.String("url", url), zap.Error(err))
		return nil
	}
	if res.StatusCode != 200 {
		common.Logger.Error("fetch twitter content error", zap.String("id", twitterID), zap.String("url", url))
		return nil
	}
	if res != nil {
		defer res.Body.Close()
	}
	body, _ := io.ReadAll(res.Body)

	var resObj model.EntryFetchResponseModel
	if err := json.Unmarshal(body, &resObj); err != nil {
		common.Logger.Error("fetch twitter content ,json decode failed, err", zap.Error(err))
		return nil
	}

	return &resObj.Data

}
