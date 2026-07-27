package config

import (
	"log"
	"os"
	"database/sql"
	"github.com/grysha11/camagru-backend/internal/db"
)
type Config struct {
	DB	*db.Queries
	JWTSecret string
}

func NewConfig(database *sql.DB) *Config {
	jwt := os.Getenv("JWT_SECRET")
	if jwt == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}
	return &Config{
		DB: db.New(database),
		JWTSecret: jwt,
	}
}
