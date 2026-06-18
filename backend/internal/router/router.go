package router

import (
	"net/http"

	"github.com/grysha11/camagru-backend/internal/api"
	"github.com/grysha11/camagru-backend/internal/config"
	"github.com/grysha11/camagru-backend/internal/handler"
)

func NewRouter(cfg *config.Config, h *handler.Handler) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc(http.MethodGet+" /api/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./api/openapi.yaml")
	})

	mux.HandleFunc(http.MethodGet+" /docs", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./api/docs.html")
	})

	mux.HandleFunc(http.MethodGet+" /swagger", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs", http.StatusMovedPermanently)
	})
	
	strictHandler := api.NewStrictHandler(h, nil)
	api.HandlerFromMux(strictHandler, mux)

	return mux
}