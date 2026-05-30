package router

import (
	"net/http"
	"github.com/grysha11/camagru-backend/internal/config"
	"github.com/grysha11/camagru-backend/internal/handler"
)

func NewRouter(cfg *config.Config, h *handler.Handler) *http.ServeMux {
	mux := http.NewServeMux()

	apiRouter := http.NewServeMux()
	apiRouter.HandleFunc("GET /healthz", h.Healthz)

	mux.Handle("/api/", http.StripPrefix("/api", apiRouter))

	return mux
}