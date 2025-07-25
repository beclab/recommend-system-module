package api

import (
	"net/http"

	"bytetrade.io/web3os/vector-crawl/crawler"
	"bytetrade.io/web3os/vector-crawl/http/request"
	"bytetrade.io/web3os/vector-crawl/http/response/json"
)

func (h *handler) parseFileType(w http.ResponseWriter, r *http.Request) {
	url := request.QueryStringParam(r, "url", "")
	bflUser := request.QueryStringParam(r, "bfl_user", "")
	entry := crawler.PageCrawler(url, bflUser)
	json.OK(w, r, entry)

}
func (h *handler) parse(w http.ResponseWriter, r *http.Request) {
	url := request.QueryStringParam(r, "url", "")
	bflUser := request.QueryStringParam(r, "bfl_user", "")
	entry := crawler.PageCrawler(url, bflUser)
	json.OK(w, r, entry)

}
