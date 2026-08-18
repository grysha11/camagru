package handler

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/grysha11/camagru-backend/internal/api"
	"github.com/grysha11/camagru-backend/internal/auth"
	"github.com/grysha11/camagru-backend/internal/db"
	"github.com/grysha11/camagru-backend/internal/mailer"
	"github.com/grysha11/camagru-backend/internal/middleware"
)

const (
	emailPurposeVerify = "verify_email"
	emailPurposeReset  = "reset_password"
)

const (
	emailVerifyTokenTTL   = 24 * time.Hour
	passwordResetTokenTTL = 1 * time.Hour
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

	user, err := h.Cfg.DB.CreateUser(ctx, db.CreateUserParams{
		Username:       r.Body.Username,
		Email:          string(r.Body.Email),
		HashedPassword: sql.NullString{String: hashedPassword, Valid: true},
	})
	if err != nil {
		slog.Warn("Registration rejected: database unique constraint conflict", slog.Any("error", err), slog.String("email", string(r.Body.Email)), slog.String("username", r.Body.Username))
		return api.RegisterUser400JSONResponse{Error: "Username or email already exists"}, nil
	}

	if err := h.sendEmailToken(ctx, user.ID, user.Email, emailPurposeVerify, emailVerifyTokenTTL,
		"Confirm your Camagru account", "confirm-email.html",
		"Please confirm your email address by visiting the link below:",
		"This link works once and expires in 24 hours."); err != nil {
		slog.Error("Registration: failed to send confirmation email", slog.Any("error", err), slog.String("user_id", user.ID.String()))
	}

	return api.RegisterUser201JSONResponse{Message: "User registered successfully. Please check your email to confirm your account."}, nil
}

func (h *Handler) createEmailToken(ctx context.Context, userID uuid.UUID, purpose string, ttl time.Duration) (string, error) {
	if err := h.Cfg.DB.DeleteEmailTokensByUserAndPurpose(ctx, db.DeleteEmailTokensByUserAndPurposeParams{
		UserID:  userID,
		Purpose: purpose,
	}); err != nil {
		slog.Error("Failed to invalidate previous tokens", slog.Any("error", err), slog.String("user_id", userID.String()), slog.String("purpose", purpose))
	}

	rawToken, err := auth.GenerateRandomToken()
	if err != nil {
		return "", err
	}

	if err := h.Cfg.DB.CreateEmailToken(ctx, db.CreateEmailTokenParams{
		TokenHash: auth.HashToken(rawToken),
		UserID:    userID,
		Purpose:   purpose,
		ExpiredAt: time.Now().Add(ttl),
	}); err != nil {
		return "", err
	}

	return rawToken, nil
}

func (h *Handler) sendEmailToken(ctx context.Context, userID uuid.UUID, email, purpose string, ttl time.Duration, subject, page, bodyIntro, footnote string) error {
	rawToken, err := h.createEmailToken(ctx, userID, purpose, ttl)
	if err != nil {
		return err
	}

	link := fmt.Sprintf("%s/%s?token=%s", h.Cfg.AppBaseURL, page, rawToken)
	body, err := mailer.RenderLinkEmail(bodyIntro, link, footnote)
	if err != nil {
		return err
	}
	return h.Cfg.Mailer.Send(email, subject, body)
}

func (h *Handler) issueSessionCookies(ctx context.Context, userID uuid.UUID) (string, error) {
	accessToken, err := auth.GenerateAccessToken(userID.String(), h.Cfg.JWTSecret, 15*time.Minute)
	if err != nil {
		return "", err
	}

	refreshToken, err := auth.GenerateRandomToken()
	if err != nil {
		return "", err
	}

	if err := h.Cfg.DB.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{
		TokenHash: auth.HashToken(refreshToken),
		UserID:    userID,
		ExpiredAt: time.Now().Add(7 * 24 * time.Hour),
	}); err != nil {
		return "", err
	}

	accessCookie := fmt.Sprintf("access_token=%s; Path=/; Max-Age=%d; HttpOnly; Secure; SameSite=Strict", accessToken, 15*60)
	refreshCookie := fmt.Sprintf("refresh_token=%s; Path=/; Max-Age=%d; HttpOnly; Secure; SameSite=Strict", refreshToken, 7*24*60*60)

	return accessCookie + "\n" + refreshCookie, nil
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

	if !user.HashedPassword.Valid {
		slog.Warn("Login failed: account has no password set", slog.String("email", string(r.Body.Email)))
		return api.LoginUser401JSONResponse{Error: "Invalid email or password"}, nil
	}

	if err := auth.CheckHash(r.Body.Password, user.HashedPassword.String); err != nil {
		slog.Warn("Login failed: incorrect password", slog.String("email", string(r.Body.Email)))
		return api.LoginUser401JSONResponse{Error: "Invalid email or password"}, nil
	}

	if !user.EmailVerifiedAt.Valid {
		slog.Warn("Login rejected: email not verified", slog.String("email", string(r.Body.Email)))
		return api.LoginUser401JSONResponse{Error: "Please confirm your email before logging in"}, nil
	}

	cookieHeader, err := h.issueSessionCookies(ctx, user.ID)
	if err != nil {
		slog.Error("Login failed: could not issue session", slog.Any("error", err), slog.String("user_id", user.ID.String()))
		return api.LoginUser401JSONResponse{Error: "Could not create session"}, nil
	}

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

	tokenHash := auth.HashToken(refreshToken)
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

	cookieHeader, err := h.issueSessionCookies(ctx, dbToken.UserID)
	if err != nil {
		slog.Error("Refresh failed: could not issue session", slog.Any("error", err), slog.String("user_id", dbToken.UserID.String()))
		return api.RefreshToken401JSONResponse{Error: "Could not create session"}, nil
	}
	if err := h.Cfg.DB.RevokeRefreshToken(ctx, tokenHash); err != nil {
		slog.Error("Failed to revoke old refresh token after rotation", slog.Any("error", err), slog.String("user_id", dbToken.UserID.String()))
	}

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
		if err := h.Cfg.DB.RevokeRefreshToken(ctx, auth.HashToken(refreshToken)); err != nil {
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

func (h *Handler) ConfirmEmail(ctx context.Context, r api.ConfirmEmailRequestObject) (api.ConfirmEmailResponseObject, error) {
	tokenHash := auth.HashToken(r.Params.Token)
	emailToken, err := h.Cfg.DB.GetEmailToken(ctx, tokenHash)
	if err != nil || emailToken.Purpose != emailPurposeVerify || emailToken.UsedAt.Valid || time.Now().After(emailToken.ExpiredAt) {
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			slog.Error("Confirm email failed: database error looking up token", slog.Any("error", err))
		}
		return api.ConfirmEmail400JSONResponse{Error: "Invalid or expired token"}, nil
	}

	if err := h.Cfg.DB.MarkEmailVerified(ctx, emailToken.UserID); err != nil {
		slog.Error("Confirm email failed: could not mark user verified", slog.Any("error", err), slog.String("user_id", emailToken.UserID.String()))
		return api.ConfirmEmail400JSONResponse{Error: "Could not confirm email"}, nil
	}

	if err := h.Cfg.DB.MarkEmailTokenUsed(ctx, tokenHash); err != nil {
		slog.Error("Confirm email: failed to mark token used", slog.Any("error", err), slog.String("user_id", emailToken.UserID.String()))
	}

	return api.ConfirmEmail200JSONResponse{Message: "Email confirmed successfully"}, nil
}

func (h *Handler) ForgotPassword(ctx context.Context, r api.ForgotPasswordRequestObject) (api.ForgotPasswordResponseObject, error) {
	const genericMessage = "If an account exists with that email, a password reset link has been sent."

	user, err := h.Cfg.DB.GetUserByEmail(ctx, string(r.Body.Email))
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			slog.Error("Forgot password: database error looking up user by email", slog.Any("error", err))
		}
		return api.ForgotPassword200JSONResponse{Message: genericMessage}, nil
	}

	if err := h.sendEmailToken(ctx, user.ID, user.Email, emailPurposeReset, passwordResetTokenTTL,
		"Reset your Camagru password", "reset-password.html",
		"Reset your password by visiting the link below:",
		"This link works once and expires in 1 hour."); err != nil {
		slog.Error("Forgot password: failed to send reset email", slog.Any("error", err), slog.String("user_id", user.ID.String()))
	}

	return api.ForgotPassword200JSONResponse{Message: genericMessage}, nil
}

func (h *Handler) ResetPassword(ctx context.Context, r api.ResetPasswordRequestObject) (api.ResetPasswordResponseObject, error) {
	if err := auth.ValidatePassword(r.Body.Password); err != nil {
		return api.ResetPassword400JSONResponse{Error: "Password must be at least 8 characters and include uppercase, lowercase, number, and special character"}, nil
	}

	tokenHash := auth.HashToken(r.Body.Token)
	emailToken, err := h.Cfg.DB.GetEmailToken(ctx, tokenHash)
	if err != nil || emailToken.Purpose != emailPurposeReset || emailToken.UsedAt.Valid || time.Now().After(emailToken.ExpiredAt) {
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			slog.Error("Reset password failed: database error looking up token", slog.Any("error", err))
		}
		return api.ResetPassword400JSONResponse{Error: "Invalid or expired token"}, nil
	}

	hashedPassword, err := auth.HashPassword(r.Body.Password)
	if err != nil {
		slog.Error("Reset password failed: password hashing error", slog.Any("error", err), slog.String("user_id", emailToken.UserID.String()))
		return api.ResetPassword400JSONResponse{Error: "Failed to process password"}, nil
	}

	if err := h.Cfg.DB.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
		HashedPassword: sql.NullString{String: hashedPassword, Valid: true},
		ID:             emailToken.UserID,
	}); err != nil {
		slog.Error("Reset password failed: could not update password", slog.Any("error", err), slog.String("user_id", emailToken.UserID.String()))
		return api.ResetPassword400JSONResponse{Error: "Could not reset password"}, nil
	}

	if err := h.Cfg.DB.MarkEmailTokenUsed(ctx, tokenHash); err != nil {
		slog.Error("Reset password: failed to mark token used", slog.Any("error", err), slog.String("user_id", emailToken.UserID.String()))
	}

	if err := h.Cfg.DB.DeleteAllUserRefreshTokens(ctx, emailToken.UserID); err != nil {
		slog.Error("Reset password: failed to revoke existing sessions", slog.Any("error", err), slog.String("user_id", emailToken.UserID.String()))
	}

	return api.ResetPassword200JSONResponse{Message: "Password reset successfully"}, nil
}

func (h *Handler) RequestPasswordChange(ctx context.Context, r api.RequestPasswordChangeRequestObject) (api.RequestPasswordChangeResponseObject, error) {
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return api.RequestPasswordChange401JSONResponse{Error: "Not authenticated"}, nil
	}

	rawToken, err := h.createEmailToken(ctx, userID, emailPurposeReset, passwordResetTokenTTL)
	if err != nil {
		slog.Error("Request password change: failed to create token", slog.Any("error", err), slog.String("user_id", userID.String()))
		return api.RequestPasswordChange401JSONResponse{Error: "Could not generate token"}, nil
	}

	return api.RequestPasswordChange200JSONResponse{Token: rawToken}, nil
}
