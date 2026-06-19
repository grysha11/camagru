package auth

import (
	"net/mail"
	"unicode"
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func CheckHash(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err
}

func IsValidPassword(password string) error {
	var (
		hasMinLen = false
		hasUpper = false
		hasLower = false
		hasNumber = false
		hasSpecial = false
	)

	if len(password) >= 8 {
		hasMinLen = true
	}

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if hasMinLen && hasUpper && hasLower && hasNumber && hasSpecial {
		return nil
	}

	return fmt.Errorf("Not valid password")
}

func IsValidEmail(email string) error {
	addr, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("invalid email format")
	}

	if addr.Address != email {
		return fmt.Errorf("invalid email format")
	}
	
	return nil
}
