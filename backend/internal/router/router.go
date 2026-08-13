package router

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/grysha11/camagru-backend/internal/api"
	"github.com/grysha11/camagru-backend/internal/config"
	"github.com/grysha11/camagru-backend/internal/handler"
	"github.com/grysha11/camagru-backend/internal/middleware"
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

	mux.HandleFunc(http.MethodGet+" /api/overlays/{id}", func(w http.ResponseWriter, r *http.Request) {
		entry, ok := cfg.Overlays.Get(r.PathValue("id"))
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		w.Write(entry.Bytes)
	})

	loginLimiter := middleware.NewRateLimiter(5, time.Minute)
	strictHandler := api.NewStrictHandler(h, []api.StrictMiddlewareFunc{
		middleware.RateLimitMiddleware(loginLimiter),
		middleware.AuthMiddleware(cfg.JWTSecret),
		middleware.RefreshTokenContextMiddleware(),
		middleware.MultiCookieMiddleware(),
	})
	api.HandlerFromMux(strictHandler, mux)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)

		json.NewEncoder(w).Encode(api.ErrorResponse{
			Error: "Endpoint not found: " + r.URL.Path,
		})
	})

	return mux
}