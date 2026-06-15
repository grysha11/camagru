package middleware

import (
	"context"
	"net/http"

	"github.com/grysha11/camagru-backend/internal/auth"
)

type contextKey string
const UserIDKey contextKey = "userID"

func AuthMiddleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("acccess_token")
			if err != nil {
				http.Error(w, "unathorized", http.StatusUnauthorized)
				return
			}

			userID, err := auth.ValidateAccessToken(cookie.Value, secret)
			if err != nil {
				http.Error(w, "unathorized", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			reqWithCtx := r.WithContext(ctx)
			next.ServeHTTP(w, reqWithCtx)
		})
	}
}