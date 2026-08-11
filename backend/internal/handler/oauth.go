package handler

import (
	"context"
	"crypto/subtle"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/lib/pq"

	"github.com/grysha11/camagru-backend/internal/api"
	"github.com/grysha11/camagru-backend/internal/auth"
	"github.com/grysha11/camagru-backend/internal/db"
	"github.com/grysha11/camagru-backend/internal/middleware"
	"github.com/grysha11/camagru-backend/internal/oauth"
)

const githubProvider = "github"

var (
	errEmailConflictUnverified  = errors.New("oauth: email matches an unverified local account")
	errUsernameRetriesExhausted = errors.New("oauth: could not find a free username")
)

func (h *Handler) GitHubOAuthLogin(ctx context.Context, r api.GitHubOAuthLoginRequestObject) (api.GitHubOAuthLoginResponseObject, error) {
	state, err := auth.GenerateRandomToken()
	if err != nil {
		slog.Error("oauth: failed to generate state", slog.Any("error", err))
		return api.GitHubOAuthLogin500JSONResponse{Error: "Could not start GitHub login"}, nil
	}

	stateCookie := fmt.Sprintf("oauth_state=%s; Path=/api/oauth/github; Max-Age=300; HttpOnly; Secure; SameSite=Lax", state)
	location := h.Cfg.GitHub.AuthorizeURL(state)

	return api.GitHubOAuthLogin302Response{
		Headers: api.GitHubOAuthLogin302ResponseHeaders{
			Location:  &location,
			SetCookie: &stateCookie,
		},
	}, nil
}

func (h *Handler) GitHubOAuthCallback(ctx context.Context, r api.GitHubOAuthCallbackRequestObject) (api.GitHubOAuthCallbackResponseObject, error) {
	clearState := "oauth_state=; Path=/api/oauth/github; Max-Age=0; HttpOnly; Secure; SameSite=Lax"

	fail := func(reason string) api.GitHubOAuthCallbackResponseObject {
		location := fmt.Sprintf("%s/index.html?oauth_error=%s", h.Cfg.AppBaseURL, reason)
		cookie := clearState
		return api.GitHubOAuthCallback302Response{
			Headers: api.GitHubOAuthCallback302ResponseHeaders{
				Location:  &location,
				SetCookie: &cookie,
			},
		}
	}

	if r.Params.Error != nil && *r.Params.Error != "" {
		slog.Warn("oauth: github returned an error", slog.String("error", *r.Params.Error))
		return fail("access_denied"), nil
	}

	stateCookie, _ := ctx.Value(middleware.OAuthStateKey).(string)
	if stateCookie == "" || len(stateCookie) != len(r.Params.State) ||
		subtle.ConstantTimeCompare([]byte(r.Params.State), []byte(stateCookie)) != 1 {
		slog.Warn("oauth: state mismatch")
		return fail("state_mismatch"), nil
	}

	if r.Params.Code == nil || *r.Params.Code == "" {
		return fail("missing_code"), nil
	}

	ghToken, err := h.Cfg.GitHub.Exchange(ctx, *r.Params.Code)
	if err != nil {
		slog.Error("oauth: github code exchange failed", slog.Any("error", err))
		return fail("exchange_failed"), nil
	}

	ghUser, err := h.Cfg.GitHub.FetchUser(ctx, ghToken)
	if err != nil {
		if errors.Is(err, oauth.ErrNoVerifiedEmail) {
			slog.Warn("oauth: github account has no verified email")
			return fail("no_email"), nil
		}
		slog.Error("oauth: failed to fetch github user", slog.Any("error", err))
		return fail("server_error"), nil
	}

	user, err := h.resolveGitHubUser(ctx, ghUser)
	if err != nil {
		if errors.Is(err, errEmailConflictUnverified) {
			slog.Warn("oauth: email collides with unverified local account", slog.String("email", ghUser.Email))
			return fail("email_conflict"), nil
		}
		slog.Error("oauth: failed to resolve/create user", slog.Any("error", err))
		return fail("server_error"), nil
	}

	sessionCookies, err := h.issueSessionCookies(ctx, user.ID)
	if err != nil {
		slog.Error("oauth: failed to issue session", slog.Any("error", err), slog.String("user_id", user.ID.String()))
		return fail("server_error"), nil
	}

	cookieHeader := clearState + "\n" + sessionCookies
	location := fmt.Sprintf("%s/gallery.html", h.Cfg.AppBaseURL)

	return api.GitHubOAuthCallback302Response{
		Headers: api.GitHubOAuthCallback302ResponseHeaders{
			Location:  &location,
			SetCookie: &cookieHeader,
		},
	}, nil
}

func (h *Handler) resolveGitHubUser(ctx context.Context, ghUser oauth.GitHubUser) (db.User, error) {
	identity, err := h.Cfg.DB.GetOAuthIdentityByProvider(ctx, db.GetOAuthIdentityByProviderParams{
		Provider:       githubProvider,
		ProviderUserID: ghUser.ID,
	})
	if err == nil {
		return h.Cfg.DB.GetUserByID(ctx, identity.UserID)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return db.User{}, fmt.Errorf("lookup oauth identity: %w", err)
	}

	existing, err := h.Cfg.DB.GetUserByEmail(ctx, ghUser.Email)
	if err == nil {
		if !existing.EmailVerifiedAt.Valid {
			return db.User{}, errEmailConflictUnverified
		}
		if _, err := h.Cfg.DB.CreateOAuthIdentity(ctx, db.CreateOAuthIdentityParams{
			UserID:         existing.ID,
			Provider:       githubProvider,
			ProviderUserID: ghUser.ID,
		}); err != nil {
			return db.User{}, fmt.Errorf("link oauth identity: %w", err)
		}
		return existing, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return db.User{}, fmt.Errorf("lookup user by email: %w", err)
	}

	return h.createUserFromGitHub(ctx, ghUser)
}

func (h *Handler) createUserFromGitHub(ctx context.Context, ghUser oauth.GitHubUser) (db.User, error) {
	username := ghUser.Login

	for attempt := 0; attempt < 3; attempt++ {
		user, err := h.Cfg.DB.CreateUser(ctx, db.CreateUserParams{
			Username:       username,
			Email:          ghUser.Email,
			HashedPassword: sql.NullString{},
		})
		if err == nil {
			if err := h.Cfg.DB.MarkEmailVerified(ctx, user.ID); err != nil {
				slog.Error("oauth: failed to mark github-created user verified", slog.Any("error", err), slog.String("user_id", user.ID.String()))
			}
			if _, err := h.Cfg.DB.CreateOAuthIdentity(ctx, db.CreateOAuthIdentityParams{
				UserID:         user.ID,
				Provider:       githubProvider,
				ProviderUserID: ghUser.ID,
			}); err != nil {
				return db.User{}, fmt.Errorf("create oauth identity: %w", err)
			}
			return user, nil
		}

		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Constraint == "users_username_key" {
			suffix, genErr := auth.GenerateRandomToken()
			if genErr != nil {
				return db.User{}, fmt.Errorf("create user from github: %w", err)
			}
			username = ghUser.Login + "-" + suffix[:6]
			continue
		}

		return db.User{}, fmt.Errorf("create user from github: %w", err)
	}

	return db.User{}, errUsernameRetriesExhausted
}
