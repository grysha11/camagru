package main

import (
	"os"
	"log"
	"database/sql"
	_ "github.com/lib/pq"
	"net/http"
	"github.com/grysha11/camagru-backend/internal/config"
	"github.com/grysha11/camagru-backend/internal/handler"
	"github.com/grysha11/camagru-backend/internal/router"
)

func main() {
	dbUrl := os.Getenv("DB_URL")
	postgre, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatalf("Error connecting to db: %v\n", err)
	}

	cfg := config.NewConfig(postgre)

	h := &handler.Handler{Cfg: cfg}

	mux := router.NewRouter(cfg, h)

	port := os.Getenv("BACKEND_PORT")
	server := &http.Server{
		Addr: ":" + port,
		Handler: mux,
	}

	log.Printf("Listening on port: %v\n", server.Addr)
	log.Fatal(server.ListenAndServe())
}
