package api

import (
	"net/http"

	"bytetrade.io/web3os/backend-server/storage"
	"bytetrade.io/web3os/backend-server/worker"

	"github.com/gorilla/mux"
)

type handler struct {
	store  *storage.Storage
	router *mux.Router
	pool   *worker.Pool
}

func Serve(router *mux.Router, store *storage.Storage, pool *worker.Pool) {
	handler := &handler{store, router, pool}
	middleware := newMiddleware(store)
	sr := router.PathPrefix("/api").Subrouter()
	sr.Use(middleware.handleCORS)
	sr.Methods(http.MethodOptions)

	sr.HandleFunc("/knowledge/feeds/{feedID}/refresh", handler.knowledgeRefreshFeed).Methods(http.MethodPut)

	sr.HandleFunc("/entries/{entryID}/fetch-content", handler.fetchContent).Methods(http.MethodGet)
	sr.HandleFunc("/knowledge/entries/{entryID}/fetch-content", handler.knowledgeFetchContent).Methods(http.MethodGet)

	sr.HandleFunc("/knowledge/noMediaDownloadQuery", handler.noMediaDownloadQuery).Methods(http.MethodGet)

	sr.HandleFunc("/knowledge/getPvcAnnotation", handler.getPvcAnnotation).Methods(http.MethodGet)

}
