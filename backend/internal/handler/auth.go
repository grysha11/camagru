package handler

import (
	"context"
	"time"
	"fmt"
	"github.com/grysha11/camagru-backend/internal/api"
	"github.com/grysha11/camagru-backend/internal/auth"
	"github.com/grysha11/camagru-backend/internal/db"
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

}

func (h *Handler) LogoutUser(ctx context.Context, r api.LogoutUserRequestObject) (api.LogoutUserResponseObject, error) {

}
