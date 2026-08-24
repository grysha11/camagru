package auth

import "testing"

func TestHashPasswordRoundTrip(t *testing.T) {
	hash, err := HashPassword("Sup3r$ecret")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}

	if err := CheckHash("Sup3r$ecret", hash); err != nil {
		t.Errorf("CheckHash with correct password: %v", err)
	}

	if err := CheckHash("wrongpassword", hash); err == nil {
		t.Error("CheckHash with wrong password: expected error, got nil")
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"valid", "Sup3r$ecret", false},
		{"too short", "Sh0rt!", true},
		{"missing upper", "sup3r$ecret", true},
		{"missing lower", "SUP3R$ECRET", true},
		{"missing number", "Super$ecret", true},
		{"missing special", "Sup3rsecret", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePassword(%q) error = %v, wantErr %v", tt.password, err, tt.wantErr)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{"valid", "user@example.com", false},
		{"missing at", "userexample.com", true},
		{"missing domain", "user@", true},
		{"empty", "", true},
		{"trailing garbage rejected by exact-match check", "user@example.com (Name)", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEmail(%q) error = %v, wantErr %v", tt.email, err, tt.wantErr)
			}
		})
	}
}
