package main

import (
	"database/sql"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/grysha11/camagru-backend/internal/config"
	"github.com/grysha11/camagru-backend/internal/handler"
	"github.com/grysha11/camagru-backend/internal/middleware"
	"github.com/grysha11/camagru-backend/internal/router"
	_ "github.com/lib/pq"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	dbUrl := os.Getenv("DB_URL")
	if dbUrl == "" {
		log.Fatal("DB_URL environment variable is required")
	}

	postgre, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatalf("Error connecting to db: %v\n", err)
	}

	cfg := config.NewConfig(postgre)
	h := &handler.Handler{Cfg: cfg}
	mux := router.NewRouter(cfg, h)

	port := os.Getenv("BACKEND_PORT")
	if port == "" {
		log.Fatal("BACKEND_PORT environment variable is required")
	}

	loggedMux := middleware.LoggerMiddleware(mux)

	server := &http.Server{
		Addr: ":" + port,
		Handler: loggedMux,
	}

	slog.Info("Starting backend", slog.String("port", port))
	log.Fatal(server.ListenAndServe())
}
