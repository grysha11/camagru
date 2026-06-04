package router

import (
	"net/http"

	"github.com/grysha11/camagru-backend/internal/api"
	"github.com/grysha11/camagru-backend/internal/config"
	"github.com/grysha11/camagru-backend/internal/handler"
)

func NewRouter(cfg *config.Config, h *handler.Handler) *http.ServeMux {
	strictHandler := api.NewStrictHandler(h, nil)
	mux := http.NewServeMux()
	api.HandlerFromMux(strictHandler, mux)

	return mux
}