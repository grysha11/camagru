package config

import (
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
	return &Config{
		DB: db.New(database),
		JWTSecret: jwt,
	}
}
