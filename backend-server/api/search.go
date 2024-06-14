package api

import (
	"encoding/json"
	"net/http"

	respJson "bytetrade.io/web3os/backend-server/http/response/json"
	"bytetrade.io/web3os/backend-server/service/search"

	"bytetrade.io/web3os/backend-server/common"
	"bytetrade.io/web3os/backend-server/http/request"
	"bytetrade.io/web3os/backend-server/model"
	"go.uber.org/zap"
)

func (h *handler) inputRSS(w http.ResponseWriter, r *http.Request) {

	var notificationData model.NotificationData
	if err := json.NewDecoder(r.Body).Decode(&notificationData); err != nil {
		respJson.BadRequest(w, r, err)
		return
	}
	docId := search.InputRSS(&notificationData)

	//id, _ := primitive.ObjectIDFromHex(notificationData.EntryId)
	entry := &model.Entry{ID: notificationData.EntryId, DocId: docId}
	h.store.UpdateEntryDocID(entry)
	respJson.OK(w, r, docId)
}

func (h *handler) deleteRSS(w http.ResponseWriter, r *http.Request) {
	var entryIds []string
	if err := json.NewDecoder(r.Body).Decode(&entryIds); err != nil {
		respJson.BadRequest(w, r, err)
		return
	}

	entryDocIds, err := h.store.GetEntryDocList(entryIds)
	if err != nil {
		return
	}
	search.DeleteRSS(entryDocIds)
	respJson.NoContent(w, r)
}

func (h *handler) queryRSS(w http.ResponseWriter, r *http.Request) {
	query := request.QueryStringParam(r, "query", "")
	if query == "" {
		respJson.OK(w, r, "")
		return
	}
	common.Logger.Info("queryRSS", zap.String("query", query))

	content := search.QueryRSS(query)

	respJson.OK(w, r, content)
}
