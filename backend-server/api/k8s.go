package api

import (
	"net/http"

	"bytetrade.io/web3os/backend-server/common"
	"bytetrade.io/web3os/backend-server/http/response/json"
	"go.uber.org/zap"

	"bytetrade.io/web3os/backend-server/http/request"
)

func (h *handler) getPvcAnnotation(w http.ResponseWriter, r *http.Request) {
	bflUser := request.QueryStringParam(r, "bfl_user", "")
	common.Logger.Error("get pvc annotation", zap.String("bfl_user", bflUser))
	annotation, _ := common.GetPvcAnnotation(bflUser)

	json.OK(w, r, annotation)
}
