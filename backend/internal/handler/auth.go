package handler

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
	"log/slog"

	"github.com/grysha11/camagru-backend/internal/api"
	"github.com/grysha11/camagru-backend/internal/auth"
	"github.com/grysha11/camagru-backend/internal/db"
	"github.com/grysha11/camagru-backend/internal/middleware"
)

func (h *Handler) RegisterUser(ctx context.Context, r api.RegisterUserRequestObject) (api.RegisterUserResponseObject, error) {
	if err := auth.ValidateEmail(string(r.Body.Email)); err != nil {
		return api.RegisterUser400JSONResponse{Error: "Invalid email address format"}, nil
	}

	if err := auth.ValidatePassword(r.Body.Password); err != nil {
		return api.RegisterUser400JSONResponse{Error: "Password must be at least 8 characters and include uppercase, lowercase, number, and special character"}, nil
	}

	hashedPassword, err := auth.HashPassword(r.Body.Password)
	if err != nil {
		slog.Error("Registration failed: password hashing error", slog.Any("error", err), slog.String("email", string(r.Body.Email)))
		return api.RegisterUser400JSONResponse{Error: "Failed to process password"}, nil
	}

	_, err = h.Cfg.DB.CreateUser(ctx, db.CreateUserParams{
		Username: r.Body.Username,
		Email: string(r.Body.Email),
		HashedPassword: hashedPassword,
	})
	if err != nil {
		slog.Warn("Registration rejected: database unique constraint conflict", slog.Any("error", err), slog.String("email", string(r.Body.Email)), slog.String("username", r.Body.Username))
		return api.RegisterUser400JSONResponse{Error: "Username or email already exists"}, nil
	}

	return api.RegisterUser201JSONResponse{Message: "User registered successfully"}, nil
}

func (h *Handler) LoginUser(ctx context.Context, r api.LoginUserRequestObject) (api.LoginUserResponseObject, error) {
	user, err := h.Cfg.DB.GetUserByEmail(ctx, string(r.Body.Email))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.Warn("Login failed: unknown email", slog.String("email", string(r.Body.Email)))
		} else {
			slog.Error("Login failed: database error looking up user by email", slog.Any("error", err))
		}
		return api.LoginUser401JSONResponse{Error: "Invalid email or password"}, nil
	}

	if err := auth.CheckHash(r.Body.Password, user.HashedPassword); err != nil {
		slog.Warn("Login failed: incorrect password", slog.String("email", string(r.Body.Email)))
		return api.LoginUser401JSONResponse{Error: "Invalid email or password"}, nil
	}

	accessToken, err := auth.GenerateAccessToken(user.ID.String(), h.Cfg.JWTSecret, 15*time.Minute)
	if err != nil {
		return api.LoginUser401JSONResponse{Error: "Could not generate access token"}, nil
	}

	refreshToken, err := auth.GenerateRefreshToken()
	if err != nil {
		return api.LoginUser401JSONResponse{Error: "Could not generate refresh token"}, nil
	}

	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	err = h.Cfg.DB.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{
		TokenHash: auth.HashRefreshToken(refreshToken),
		UserID: user.ID,
		ExpiredAt: expiresAt,
	})
	if err != nil {
		return api.LoginUser401JSONResponse{Error: "Could not create session"}, nil
	}

	accessCookie := fmt.Sprintf("access_token=%s; Path=/; Max-Age=%d; HttpOnly; Secure; SameSite=Strict", accessToken, 15*60)
	refreshCookie := fmt.Sprintf("refresh_token=%s; Path=/; Max-Age=%d; HttpOnly; Secure; SameSite=Strict", refreshToken, 7*24*60*60)

	cookieHeader := accessCookie + "\n" + refreshCookie

	return api.LoginUser200JSONResponse{
		Body: api.SuccessResponse{Message: "Successfully logged in"},
		Headers: api.LoginUser200ResponseHeaders{
			SetCookie: &cookieHeader,
		},
	}, nil
}

func (h *Handler) RefreshToken(ctx context.Context, r api.RefreshTokenRequestObject) (api.RefreshTokenResponseObject, error) {
	refreshToken, ok := ctx.Value(middleware.RefreshTokenKey).(string)
	if !ok || refreshToken == "" {
		return api.RefreshToken401JSONResponse{Error: "Invalid session"}, nil
	}

	tokenHash := auth.HashRefreshToken(refreshToken)
	dbToken, err := h.Cfg.DB.GetRefreshToken(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.Warn("Refresh rejected: token not found")
		} else {
			slog.Error("Refresh failed: database error looking up refresh token", slog.Any("error", err))
		}
		return api.RefreshToken401JSONResponse{Error: "Invalid session"}, nil
	}

	if dbToken.RevokedAt.Valid {
		slog.Warn("Refresh token reuse detected, revoking all sessions", slog.String("user_id", dbToken.UserID.String()))
		if err := h.Cfg.DB.DeleteAllUserRefreshTokens(ctx, dbToken.UserID); err != nil {
			slog.Error("Failed to revoke all user refresh tokens after reuse detection", slog.Any("error", err), slog.String("user_id", dbToken.UserID.String()))
		}
		return api.RefreshToken401JSONResponse{Error: "Session expired, please log in again"}, nil
	}

	if time.Now().After(dbToken.ExpiredAt) {
		return api.RefreshToken401JSONResponse{Error: "Session expired, please log in again"}, nil
	}

	newAccessToken, err := auth.GenerateAccessToken(dbToken.UserID.String(), h.Cfg.JWTSecret, 15*time.Minute)
	if err != nil {
		return api.RefreshToken401JSONResponse{Error: "Could not generate access token"}, nil
	}

	newRefreshToken, err := auth.GenerateRefreshToken()
	if err != nil {
		return api.RefreshToken401JSONResponse{Error: "Could not generate refresh token"}, nil
	}

	newExpiresAt := time.Now().Add(7 * 24 * time.Hour)
	if err := h.Cfg.DB.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{
		TokenHash: auth.HashRefreshToken(newRefreshToken),
		UserID: dbToken.UserID,
		ExpiredAt: newExpiresAt,
	}); err != nil {
		return api.RefreshToken401JSONResponse{Error: "Could not create session"}, nil
	}
	if err := h.Cfg.DB.RevokeRefreshToken(ctx, tokenHash); err != nil {
		slog.Error("Failed to revoke old refresh token after rotation", slog.Any("error", err), slog.String("user_id", dbToken.UserID.String()))
	}

	accessCookie := fmt.Sprintf("access_token=%s; Path=/; Max-Age=%d; HttpOnly; Secure; SameSite=Strict", newAccessToken, 15*60)
	refreshCookie := fmt.Sprintf("refresh_token=%s; Path=/; Max-Age=%d; HttpOnly; Secure; SameSite=Strict", newRefreshToken, 7*24*60*60)
	cookieHeader := accessCookie + "\n" + refreshCookie

	return api.RefreshToken200JSONResponse{
		Body: api.SuccessResponse{Message: "Token refreshed"},
		Headers: api.RefreshToken200ResponseHeaders{
			SetCookie: &cookieHeader,
		},
	}, nil
}

func (h *Handler) LogoutUser(ctx context.Context, r api.LogoutUserRequestObject) (api.LogoutUserResponseObject, error) {
	refreshToken, ok := ctx.Value(middleware.RefreshTokenKey).(string)
	if ok && refreshToken != "" {
		if err := h.Cfg.DB.RevokeRefreshToken(ctx, auth.HashRefreshToken(refreshToken)); err != nil {
			slog.Error("Failed to revoke refresh token on logout", slog.Any("error", err))
		}
	}

	killAccessToken := "access_token=; Path=/; Max-Age=0; HttpOnly; Secure; SameSite=Strict"
	killRefreshToken := "refresh_token=; Path=/; Max-Age=0; HttpOnly; Secure; SameSite=Strict"
	cookieHeader := killAccessToken + "\n" + killRefreshToken

	return api.LogoutUser200JSONResponse{
		Body: api.SuccessResponse{Message: "Successfully logged out"},
		Headers: api.LogoutUser200ResponseHeaders{
			SetCookie: &cookieHeader,
		},
	}, nil
}
