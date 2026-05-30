package config

import "github.com/grysha11/camagru-backend/internal/db"

type Config struct {
	DB	*db.Queries
}