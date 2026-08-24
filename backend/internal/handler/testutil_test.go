package handler

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"

	"github.com/grysha11/camagru-backend/internal/auth"
	"github.com/grysha11/camagru-backend/internal/config"
	"github.com/grysha11/camagru-backend/internal/db"
	"github.com/grysha11/camagru-backend/internal/mailer"
	"github.com/grysha11/camagru-backend/internal/middleware"
	"github.com/grysha11/camagru-backend/internal/storage"
	"github.com/grysha11/camagru-backend/internal/testutil"
)

func newTestHandler(t *testing.T) *Handler {
	t.Helper()
	h, _ := newTestHandlerWithStorageDir(t)
	return h
}

func newTestHandlerWithStorageDir(t *testing.T) (*Handler, string) {
	t.Helper()

	sqlDB := testutil.OpenTestDB(t)
	testutil.TruncateAll(t, sqlDB)

	storageDir := t.TempDir()

	h := &Handler{
		Cfg: &config.Config{
			DB:         db.New(sqlDB),
			JWTSecret:  "test-secret",
			Mailer:     mailer.New(mailer.Config{Host: "127.0.0.1", Port: "1", From: "test@example.com"}),
			AppBaseURL: "http://localhost:3000",
			Storage:    storage.New(storage.Config{BasePath: storageDir, URLPrefix: "/uploads"}),
		},
	}
	return h, storageDir
}

func authContext(userID uuid.UUID) context.Context {
	return context.WithValue(context.Background(), middleware.UserIDKey, userID.String())
}

func seedVerifiedUser(t *testing.T, h *Handler, username, email, password string) db.User {
	t.Helper()

	hashed, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}

	u, err := h.Cfg.DB.CreateUser(context.Background(), db.CreateUserParams{
		Username:       username,
		Email:          email,
		HashedPassword: sql.NullString{String: hashed, Valid: true},
	})
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}

	if err := h.Cfg.DB.MarkEmailVerified(context.Background(), u.ID); err != nil {
		t.Fatalf("MarkEmailVerified: %v", err)
	}
	u.EmailVerifiedAt = sql.NullTime{Valid: true}

	return u
}
