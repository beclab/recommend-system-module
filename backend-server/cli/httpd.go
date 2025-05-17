package cli

import (
	"context"
	"log"
	"net/http"
	"time"

	"bytetrade.io/web3os/backend-server/api"
	"bytetrade.io/web3os/backend-server/common"
	"bytetrade.io/web3os/backend-server/http/request"
	"bytetrade.io/web3os/backend-server/storage"
	"bytetrade.io/web3os/backend-server/worker"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// Serve starts a new HTTP server.
func HttpdServe(store *storage.Storage, pool *worker.Pool) *http.Server {
	listenAddr := common.GetListenAddr()
	server := &http.Server{
		ReadTimeout:  300 * time.Second,
		WriteTimeout: 300 * time.Second,
		IdleTimeout:  300 * time.Second,
		Handler:      setupHandler(store, pool),
	}

	server.Addr = listenAddr
	startHTTPServer(server)

	return server
}

func startHTTPServer(server *http.Server) {
	go func() {
		common.Logger.Info(`Listening on  without TLS`, zap.String("addrsss:", server.Addr))
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			common.Logger.Fatal(`Server failed to start: %v`, zap.Error(err))
		}
	}()
}

func middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := request.FindClientIP(r)
		start := time.Now()
		ctx := r.Context()
		ctx = context.WithValue(ctx, request.ClientIPContextKey, clientIP)
		next.ServeHTTP(w, r.WithContext(ctx))

		duration := time.Since(start)
		log.Printf("time: %s, method: %s, URL: %s, cost: %v\n", start.Format(time.RFC3339), r.Method, r.URL.Path, duration)
	})
}

func setupHandler(store *storage.Storage, pool *worker.Pool) *mux.Router {
	router := mux.NewRouter()

	router.Use(middleware)

	api.Serve(router, store, pool)

	return router
}
