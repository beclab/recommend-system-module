package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

type handler struct {
	router *mux.Router
}

func Serve(router *mux.Router) {
	handler := &handler{router}
	middleware := newMiddleware()
	sr := router.PathPrefix("/api").Subrouter()
	sr.Use(middleware.handleCORS)
	sr.Methods(http.MethodOptions)

	sr.HandleFunc("/page-filetype", handler.parseFileType).Methods(http.MethodPut)
	sr.HandleFunc("/page-parse", handler.parse).Methods(http.MethodPut)

}
