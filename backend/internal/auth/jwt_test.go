package auth

import (
	"testing"
	"time"
)

func TestGenerateAndValidateAccessToken(t *testing.T) {
	token, err := GenerateAccessToken("user-123", "secret", 15*time.Minute)
	if err != nil {
		t.Fatalf("GenerateAccessToken: %v", err)
	}

	subject, err := ValidateAccessToken(token, "secret")
	if err != nil {
		t.Fatalf("ValidateAccessToken: %v", err)
	}
	if subject != "user-123" {
		t.Errorf("subject = %q, want %q", subject, "user-123")
	}
}

func TestValidateAccessTokenExpired(t *testing.T) {
	token, err := GenerateAccessToken("user-123", "secret", -1*time.Minute)
	if err != nil {
		t.Fatalf("GenerateAccessToken: %v", err)
	}

	if _, err := ValidateAccessToken(token, "secret"); err == nil {
		t.Error("expected error for expired token, got nil")
	}
}

func TestValidateAccessTokenTampered(t *testing.T) {
	token, err := GenerateAccessToken("user-123", "secret", 15*time.Minute)
	if err != nil {
		t.Fatalf("GenerateAccessToken: %v", err)
	}

	tampered := token[:len(token)-1] + "x"
	if tampered == token {
		tampered = token[:len(token)-1] + "y"
	}

	if _, err := ValidateAccessToken(tampered, "secret"); err == nil {
		t.Error("expected error for tampered token, got nil")
	}
}

func TestValidateAccessTokenWrongSecret(t *testing.T) {
	token, err := GenerateAccessToken("user-123", "secret", 15*time.Minute)
	if err != nil {
		t.Fatalf("GenerateAccessToken: %v", err)
	}

	if _, err := ValidateAccessToken(token, "wrong-secret"); err == nil {
		t.Error("expected error for wrong secret, got nil")
	}
}

func TestGenerateRandomTokenUniqueness(t *testing.T) {
	a, err := GenerateRandomToken()
	if err != nil {
		t.Fatalf("GenerateRandomToken: %v", err)
	}
	b, err := GenerateRandomToken()
	if err != nil {
		t.Fatalf("GenerateRandomToken: %v", err)
	}

	if a == "" || b == "" {
		t.Fatal("expected non-empty tokens")
	}
	if a == b {
		t.Error("expected two calls to produce different tokens")
	}
}

func TestHashTokenDeterministic(t *testing.T) {
	a := HashToken("raw-token")
	b := HashToken("raw-token")
	c := HashToken("different-token")

	if a != b {
		t.Errorf("HashToken not deterministic: %q != %q", a, b)
	}
	if a == c {
		t.Error("HashToken produced the same hash for different inputs")
	}
}
