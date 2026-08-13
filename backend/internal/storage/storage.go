package storage

import (
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/google/uuid"
)

type Config struct {
	BasePath  string
	URLPrefix string
}

type Storage struct {
	cfg Config
}

func New(cfg Config) *Storage {
	if err := os.MkdirAll(cfg.BasePath, 0o755); err != nil {
		log.Fatalf("storage: could not create base path %q: %v", cfg.BasePath, err)
	}
	return &Storage{cfg: cfg}
}

func (s *Storage) Save(data []byte, ext string) (string, error) {
	filename := uuid.NewString() + ext
	if err := os.WriteFile(filepath.Join(s.cfg.BasePath, filename), data, 0o644); err != nil {
		return "", err
	}
	return path.Join(s.cfg.URLPrefix, filename), nil
}
