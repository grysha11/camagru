package testutil

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

func OpenTestDB(t *testing.T) *sql.DB {
	t.Helper()

	url := os.Getenv("TEST_DB_URL")
	if url == "" {
		t.Skip("TEST_DB_URL not set, skipping DB test")
	}

	db, err := sql.Open("postgres", url)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := db.Ping(); err != nil {
		t.Fatalf("ping test db: %v", err)
	}

	return db
}

func TruncateAll(t *testing.T, db *sql.DB) {
	t.Helper()

	_, err := db.Exec(`TRUNCATE comments, likes, posts, oauth_identities, email_tokens, refresh_tokens, users RESTART IDENTITY CASCADE`)
	if err != nil {
		t.Fatalf("truncate: %v", err)
	}
}
