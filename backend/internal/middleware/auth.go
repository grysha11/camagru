package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/grysha11/camagru-backend/internal/api"
	"github.com/grysha11/camagru-backend/internal/auth"
)

type contextKey string

const UserIDKey contextKey = "userID"
const RefreshTokenKey contextKey = "refreshToken"
const OAuthStateKey contextKey = "oauthState"
const OAuthIntentKey contextKey = "oauthIntent"

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userIDStr, ok := ctx.Value(UserIDKey).(string)
	if !ok || userIDStr == "" {
		return uuid.UUID{}, false
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.UUID{}, false
	}
	return userID, true
}

var requiredAuthOps = map[string]bool{
	"GetMe":                 true,
	"RequestPasswordChange": true,
	"UpdateProfile":         true,
	"UploadAvatar":          true,
	"DeleteAccount":         true,
	"CreatePost":            true,
	"ListMyPosts":           true,
	"DeletePost":            true,
	"LikePost":              true,
	"UnlikePost":            true,
	"CreateComment":         true,
	"DeleteComment":         true,
	"GitHubOAuthLink":       true,
}

var optionalAuthOps = map[string]bool{
	"ListPosts":    true,
	"ListComments": true,
	"GetPost":      true,
}

func AuthMiddleware(secret string) api.StrictMiddlewareFunc {
	return func(f api.StrictHandlerFunc, operationID string) api.StrictHandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request any) (any, error) {
			if !requiredAuthOps[operationID] && !optionalAuthOps[operationID] {
				return f(ctx, w, r, request)
			}

			cookie, err := r.Cookie("access_token")
			if err != nil {
				if optionalAuthOps[operationID] {
					return f(ctx, w, r, request)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Not authenticated"})
				return nil, nil
			}

			userID, err := auth.ValidateAccessToken(cookie.Value, secret)
			if err != nil {
				if optionalAuthOps[operationID] {
					return f(ctx, w, r, request)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Not authenticated"})
				return nil, nil
			}

			ctx = context.WithValue(ctx, UserIDKey, userID)
			return f(ctx, w, r, request)
		}
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
			case "GitHubOAuthCallback":
				if cookie, err := r.Cookie("oauth_state"); err == nil && cookie.Value != "" {
					ctx = context.WithValue(ctx, OAuthStateKey, cookie.Value)
				}
				if cookie, err := r.Cookie("oauth_intent"); err == nil && cookie.Value != "" {
					ctx = context.WithValue(ctx, OAuthIntentKey, cookie.Value)
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
				splitAndClear(w, &resp.Headers.SetCookie)
				return resp, nil
			case api.LogoutUser200JSONResponse:
				splitAndClear(w, &resp.Headers.SetCookie)
				return resp, nil
			case api.DeleteAccount200JSONResponse:
				splitAndClear(w, &resp.Headers.SetCookie)
				return resp, nil
			case api.RefreshToken200JSONResponse:
				splitAndClear(w, &resp.Headers.SetCookie)
				return resp, nil
			case api.GitHubOAuthLogin302Response:
				splitAndClear(w, &resp.Headers.SetCookie)
				return resp, nil
			case api.GitHubOAuthLink302Response:
				splitAndClear(w, &resp.Headers.SetCookie)
				return resp, nil
			case api.GitHubOAuthCallback302Response:
				splitAndClear(w, &resp.Headers.SetCookie)
				return resp, nil
			default:
				return response, err
			}
		}
	}
}

func splitAndClear(w http.ResponseWriter, cookie **string) {
	writeSplitCookies(w, *cookie)
	*cookie = nil
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
