package oauth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) *GitHubClient {
	t.Helper()
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)

	return &GitHubClient{
		cfg:        GitHubConfig{ClientID: "id", ClientSecret: "secret", RedirectURL: "http://localhost/callback"},
		httpClient: &http.Client{Timeout: 5 * time.Second},
		tokenURL:   ts.URL + "/login/oauth/access_token",
		userURL:    ts.URL + "/user",
		emailsURL:  ts.URL + "/user/emails",
	}
}

func TestAuthorizeURL(t *testing.T) {
	c := NewGitHubClient(GitHubConfig{ClientID: "abc123", ClientSecret: "secret", RedirectURL: "http://localhost/callback"})
	got := c.AuthorizeURL("state-xyz")

	if got == "" {
		t.Fatal("AuthorizeURL returned empty string")
	}
	for _, want := range []string{"client_id=abc123", "state=state-xyz", "redirect_uri="} {
		if !strings.Contains(got, want) {
			t.Errorf("AuthorizeURL() = %q, want it to contain %q", got, want)
		}
	}
}

func TestExchangeSuccess(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/login/oauth/access_token" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]string{"access_token": "gh-token-123"})
	})

	token, err := c.Exchange(context.Background(), "code-123")
	if err != nil {
		t.Fatalf("Exchange: %v", err)
	}
	if token != "gh-token-123" {
		t.Errorf("token = %q, want %q", token, "gh-token-123")
	}
}

func TestExchangeError(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"error": "bad_verification_code", "error_description": "expired"})
	})

	if _, err := c.Exchange(context.Background(), "bad-code"); err == nil {
		t.Error("expected error for a failed exchange, got nil")
	}
}

func TestFetchUserWithVerifiedPrimaryEmail(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/user":
			json.NewEncoder(w).Encode(map[string]any{"id": 42, "login": "octocat", "name": "The Octocat"})
		case "/user/emails":
			json.NewEncoder(w).Encode([]map[string]any{
				{"email": "unverified@example.com", "primary": false, "verified": false},
				{"email": "verified@example.com", "primary": true, "verified": true},
			})
		default:
			http.NotFound(w, r)
		}
	})

	user, err := c.FetchUser(context.Background(), "token")
	if err != nil {
		t.Fatalf("FetchUser: %v", err)
	}
	if user.ID != "42" || user.Login != "octocat" || user.Email != "verified@example.com" {
		t.Errorf("FetchUser() = %+v, want ID=42 Login=octocat Email=verified@example.com", user)
	}
}

func TestFetchUserNoVerifiedEmail(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/user":
			json.NewEncoder(w).Encode(map[string]any{"id": 42, "login": "octocat"})
		case "/user/emails":
			json.NewEncoder(w).Encode([]map[string]any{
				{"email": "unverified@example.com", "primary": true, "verified": false},
			})
		default:
			http.NotFound(w, r)
		}
	})

	if _, err := c.FetchUser(context.Background(), "token"); err != ErrNoVerifiedEmail {
		t.Errorf("FetchUser() error = %v, want %v", err, ErrNoVerifiedEmail)
	}
}
