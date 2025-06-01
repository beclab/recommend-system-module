package api

import (
	"net/http"

	"bytetrade.io/web3os/backend-server/common"
	"bytetrade.io/web3os/backend-server/http/response/json"

	"bytetrade.io/web3os/backend-server/http/request"
)

func (h *handler) getPvcAnnotation(w http.ResponseWriter, r *http.Request) {
	bflUser := request.RouteStringParam(r, "bfl_user")

	annotation, _ := common.GetPvcAnnotation(bflUser)

	json.OK(w, r, annotation)
}
