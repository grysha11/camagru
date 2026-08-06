package config

import (
	"log"
	"os"
	"database/sql"
	"github.com/grysha11/camagru-backend/internal/db"
	"github.com/grysha11/camagru-backend/internal/mailer"
)
type Config struct {
	DB	*db.Queries
	JWTSecret string
	Mailer *mailer.Mailer
	AppBaseURL string
}

func mustEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("%s environment variable is required", key)
	}
	return value
}

func NewConfig(database *sql.DB) *Config {
	jwt := mustEnv("JWT_SECRET")
	appBaseURL := mustEnv("APP_BASE_URL")

	m := mailer.New(mailer.Config{
		Host:     mustEnv("SMTP_HOST"),
		Port:     mustEnv("SMTP_PORT"),
		User:     os.Getenv("SMTP_USER"),
		Password: os.Getenv("SMTP_PASSWORD"),
		From:     mustEnv("SMTP_FROM"),
	})

	return &Config{
		DB: db.New(database),
		JWTSecret: jwt,
		Mailer: m,
		AppBaseURL: appBaseURL,
	}
}
