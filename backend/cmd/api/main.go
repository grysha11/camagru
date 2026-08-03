package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	postgre.SetMaxOpenConns(25)
	postgre.SetMaxIdleConns(25)
	postgre.SetConnMaxLifetime(5 * time.Minute)

	pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
	err = postgre.PingContext(pingCtx)
	pingCancel()
	if err != nil {
		log.Fatalf("Error pinging db: %v\n", err)
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
		Addr:    ":" + port,
		Handler: loggedMux,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		slog.Info("Starting backend", slog.String("port", port))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Server failed", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	stop()
	slog.Info("Shutting down server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("Graceful shutdown failed", slog.Any("error", err))
	}

	if err := postgre.Close(); err != nil {
		slog.Error("Error closing database connection", slog.Any("error", err))
	}
}
