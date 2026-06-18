package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/grysha11/camagru-backend/internal/api"
	"github.com/grysha11/camagru-backend/internal/auth"
	"github.com/grysha11/camagru-backend/internal/db"
	"github.com/grysha11/camagru-backend/internal/middleware"
)

func (h *Handler) RegisterUser(ctx context.Context, r api.RegisterUserRequestObject) (api.RegisterUserResponseObject, error) {
	if err := auth.IsValidPassword(r.Body.Password); err != nil {
		return api.RegisterUser400JSONResponse{Error: "Password must be at least 8 characters and include uppercase, lowercase, number, and special character"}, nil
	}

	hashedPassword, err := auth.HashPassword(r.Body.Password)
	if err != nil {
		return api.RegisterUser400JSONResponse{Error: "Failed to proccess password"}, nil
	}

	_, err = h.Cfg.DB.CreateUser(ctx, db.CreateUserParams{
		Username: r.Body.Username,
		Email: string(r.Body.Email),
		HashedPassword: hashedPassword,
	})
	if err != nil {
		return api.RegisterUser400JSONResponse{Error: "Username or email already exists"}, nil
	}

	return api.RegisterUser201JSONResponse{Message: "User registered successfully"}, nil
}

func (h *Handler) LoginUser(ctx context.Context, r api.LoginUserRequestObject) (api.LoginUserResponseObject, error) {
	user, err := h.Cfg.DB.GetUserByEmail(ctx, string(r.Body.Email))
	if err != nil {
		return api.LoginUser401JSONResponse{Error: "User does not exists with this email"}, nil
	}

	if err := auth.CheckHash(r.Body.Password, user.HashedPassword); err != nil {
		return api.LoginUser401JSONResponse{Error: "Invalid credentials"}, nil
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
		Token: refreshToken,
		UserID: user.ID,
		ExpiredAt: expiresAt,
	})
	if err != nil {
		return api.LoginUser401JSONResponse{Error: "Could not create session"}, nil
	}

	accessCookie := fmt.Sprintf("access_token=%s; Path=/; Max-Age=%d; HttpOnly; Secure; SameSite=Strict", accessToken, 15*60)
	refreshCookie := fmt.Sprintf("refresh_token=%s; Path=/; Max-Age=%d; HttpOnly; Secure; SameSite=Strict", refreshToken, 7*24*60*60)

	cookieHeader := accessCookie + "\nSet-Cookie: " + refreshCookie

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

	dbToken, err := h.Cfg.DB.GetRefreshToken(ctx, refreshToken)
	if err != nil {
		return api.RefreshToken401JSONResponse{Error: "Invalid session"}, nil
	}

	if time.Now().After(dbToken.ExpiredAt) || dbToken.RevokedAt.Valid {
		return api.RefreshToken401JSONResponse{Error: "Session expired, please log in again"}, nil
	}

	newAccessToken, err := auth.GenerateAccessToken(dbToken.UserID.String(), h.Cfg.JWTSecret, 15*time.Minute)
	if err != nil {
		return api.RefreshToken401JSONResponse{Error: "Could not generate refresh token"}, nil
	}

	accessToken := fmt.Sprintf("access_token=%s; Path=/; Max-Age=%d; HttpOnly; Secure; SameSite=Strict", newAccessToken, 15*60)

	return api.RefreshToken200JSONResponse{
		Body: api.SuccessResponse{Message: "Token refreshed"},
		Headers: api.RefreshToken200ResponseHeaders{
			SetCookie: &accessToken,
		},
	}, nil
}

func (h *Handler) LogoutUser(ctx context.Context, r api.LogoutUserRequestObject) (api.LogoutUserResponseObject, error) {
	refreshToken, ok := ctx.Value(middleware.RefreshTokenKey).(string)
	if ok || refreshToken != "" {
		_ = h.Cfg.DB.DeleteRefreshToken(ctx, refreshToken)
	}

	killAcessToken := "access_token=; Path=/; Max-Age=0; HttpOnly; Secure; SameSite=Strict"
	killRefreshToken := "refresh_token=; Path=/; Max-Age=0; HttpOnly; Secure; SameSite=Strict"
	cookieHeader := killAcessToken + "\nSet-Cookie" + killRefreshToken

	return api.LogoutUser200JSONResponse{
		Body: api.SuccessResponse{Message: "Successfully logged out"},
		Headers: api.LogoutUser200ResponseHeaders{
			SetCookie: &cookieHeader,
		},
	}, nil
}
