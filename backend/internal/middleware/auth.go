package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/grysha11/camagru-backend/internal/api"
	"github.com/grysha11/camagru-backend/internal/auth"
)

type contextKey string
const UserIDKey contextKey = "userID"
const RefreshTokenKey contextKey = "refreshToken"

func AuthMiddleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("access_token")
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

func RefreshTokenContextMiddleware() api.StrictMiddlewareFunc {
	return func(f api.StrictHandlerFunc, operationID string) api.StrictHandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request any) (any, error) {
			switch operationID {
			case "RefreshToken":
				cookie, err := r.Cookie("refresh_token")
				if err != nil || cookie.Value == "" {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusUnauthorized)
					json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Invalid session"})
					return nil, nil
				}
				ctx = context.WithValue(ctx, RefreshTokenKey, cookie.Value)
			case "LogoutUser":
				if cookie, err := r.Cookie("refresh_token"); err == nil && cookie.Value != "" {
					ctx = context.WithValue(ctx, RefreshTokenKey, cookie.Value)
				}
			}
			return f(ctx, w, r, request)
		}
	}
}

func MultiCookieMiddleware() api.StrictMiddlewareFunc {
	return func(f api.StrictHandlerFunc, operationID string) api.StrictHandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request any) (any, error) {
			response, err := f(ctx, w, r, request)
			if err != nil {
				return response, err
			}
			switch resp := response.(type) {
			case api.LoginUser200JSONResponse:
				writeSplitCookies(w, resp.Headers.SetCookie)
				resp.Headers.SetCookie = nil
				return resp, nil
			case api.LogoutUser200JSONResponse:
				writeSplitCookies(w, resp.Headers.SetCookie)
				resp.Headers.SetCookie = nil
				return resp, nil
			default:
				return response, err
			}
		}
	}
}

func writeSplitCookies(w http.ResponseWriter, combined *string) {
	if combined == nil {
		return
	}
	for _, c := range strings.Split(*combined, "\n") {
		if c != "" {
			w.Header().Add("Set-Cookie", c)
		}
	}
}
