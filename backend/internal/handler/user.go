package handler

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/oapi-codegen/runtime/types"

	"github.com/grysha11/camagru-backend/internal/api"
	"github.com/grysha11/camagru-backend/internal/auth"
	"github.com/grysha11/camagru-backend/internal/db"
	"github.com/grysha11/camagru-backend/internal/imaging"
	"github.com/grysha11/camagru-backend/internal/mailer"
	"github.com/grysha11/camagru-backend/internal/middleware"
)

func (h *Handler) sendEmailChangeConfirmation(ctx context.Context, userID uuid.UUID, currentEmail string) error {
	rawToken, err := h.createEmailToken(ctx, userID, emailPurposeChangeEmail, emailVerifyTokenTTL)
	if err != nil {
		return err
	}

	link := fmt.Sprintf("%s/confirm-email.html?type=change_email&token=%s", h.Cfg.AppBaseURL, rawToken)
	body, err := mailer.RenderLinkEmail(
		"We received a request to change the email address on your Camagru account. If this was you, confirm the change by visiting the link below:",
		link,
		"This link works once and expires in 24 hours. If you didn't request this, you can safely ignore this email.",
	)
	if err != nil {
		return err
	}
	return h.Cfg.Mailer.Send(currentEmail, "Confirm your Camagru email change", body)
}

func (h *Handler) toUserResponse(ctx context.Context, user db.User) api.UserResponse {
	var avatarPath *string
	if user.AvatarPath.Valid {
		avatarPath = &user.AvatarPath.String
	}

	hasGithubLogin := false
	if _, err := h.Cfg.DB.GetOAuthIdentityByUser(ctx, db.GetOAuthIdentityByUserParams{UserID: user.ID, Provider: "github"}); err == nil {
		hasGithubLogin = true
	} else if !errors.Is(err, sql.ErrNoRows) {
		slog.Warn("toUserResponse: db error checking github oauth identity", slog.Any("error", err), slog.String("user_id", user.ID.String()))
	}

	return api.UserResponse{
		Id:              user.ID,
		Username:        user.Username,
		Email:           types.Email(user.Email),
		AvatarPath:      avatarPath,
		NotifyOnComment: user.NotifyOnComment,
		CreatedAt:       user.CreatedAt.Time,
		HasGithubLogin:  hasGithubLogin,
	}
}

func (h *Handler) GetMe(ctx context.Context, r api.GetMeRequestObject) (api.GetMeResponseObject, error) {
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return api.GetMe401JSONResponse{Error: "Not authenticated"}, nil
	}

	user, err := h.Cfg.DB.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.Warn("GetMe failed: user not found", slog.String("user_id", userID.String()))
		} else {
			slog.Error("GetMe failed: database error looking up user by id", slog.Any("error", err))
		}
		return api.GetMe401JSONResponse{Error: "Not authenticated"}, nil
	}

	return api.GetMe200JSONResponse(h.toUserResponse(ctx, user)), nil
}

func (h *Handler) GetUserProfile(ctx context.Context, r api.GetUserProfileRequestObject) (api.GetUserProfileResponseObject, error) {
	user, err := h.Cfg.DB.GetUserByUsername(ctx, r.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return api.GetUserProfile404JSONResponse{Error: "User not found"}, nil
		}
		slog.Error("GetUserProfile: db error", slog.Any("error", err), slog.String("username", r.Username))
		return api.GetUserProfile404JSONResponse{Error: "User not found"}, nil
	}

	var avatarPath *string
	if user.AvatarPath.Valid {
		avatarPath = &user.AvatarPath.String
	}

	return api.GetUserProfile200JSONResponse{
		Username:   user.Username,
		AvatarPath: avatarPath,
		CreatedAt:  user.CreatedAt.Time,
	}, nil
}

func (h *Handler) UpdateProfile(ctx context.Context, r api.UpdateProfileRequestObject) (api.UpdateProfileResponseObject, error) {
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return api.UpdateProfile401JSONResponse{Error: "Not authenticated"}, nil
	}

	if r.Body == nil {
		return api.UpdateProfile400JSONResponse{Error: "Nothing to update"}, nil
	}

	user, err := h.Cfg.DB.GetUserByID(ctx, userID)
	if err != nil {
		slog.Error("UpdateProfile: db error looking up user", slog.Any("error", err), slog.String("user_id", userID.String()))
		return api.UpdateProfile401JSONResponse{Error: "Not authenticated"}, nil
	}

	var message *string

	if r.Body.Username != nil {
		username := strings.TrimSpace(*r.Body.Username)
		if username == "" {
			return api.UpdateProfile400JSONResponse{Error: "Username cannot be empty"}, nil
		}
		if username != user.Username {
			if existing, err := h.Cfg.DB.GetUserByUsername(ctx, username); err == nil && existing.ID != userID {
				return api.UpdateProfile409JSONResponse{Error: "Username already in use"}, nil
			} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
				slog.Error("UpdateProfile: db error checking username", slog.Any("error", err), slog.String("user_id", userID.String()))
				return api.UpdateProfile400JSONResponse{Error: "Could not update username"}, nil
			}
			if err := h.Cfg.DB.UpdateUsername(ctx, db.UpdateUsernameParams{Username: username, ID: userID}); err != nil {
				slog.Error("UpdateProfile: db error updating username", slog.Any("error", err), slog.String("user_id", userID.String()))
				return api.UpdateProfile400JSONResponse{Error: "Could not update username"}, nil
			}
			user.Username = username
		}
	}

	if r.Body.Email != nil {
		email := string(*r.Body.Email)
		if err := auth.ValidateEmail(email); err != nil {
			return api.UpdateProfile400JSONResponse{Error: "Invalid email address format"}, nil
		}
		if email != user.Email {
			if existing, err := h.Cfg.DB.GetUserByEmail(ctx, email); err == nil && existing.ID != userID {
				return api.UpdateProfile409JSONResponse{Error: "Email already in use"}, nil
			} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
				slog.Error("UpdateProfile: db error checking email", slog.Any("error", err), slog.String("user_id", userID.String()))
				return api.UpdateProfile400JSONResponse{Error: "Could not update email"}, nil
			}
			if err := h.Cfg.DB.SetPendingEmail(ctx, db.SetPendingEmailParams{
				PendingEmail: sql.NullString{String: email, Valid: true},
				ID:           userID,
			}); err != nil {
				slog.Error("UpdateProfile: db error setting pending email", slog.Any("error", err), slog.String("user_id", userID.String()))
				return api.UpdateProfile400JSONResponse{Error: "Could not update email"}, nil
			}
			if err := h.sendEmailChangeConfirmation(ctx, userID, user.Email); err != nil {
				slog.Error("UpdateProfile: failed to send email-change confirmation", slog.Any("error", err), slog.String("user_id", userID.String()))
			}
			msg := "Confirmation email sent to your current address. The change takes effect once you confirm it."
			message = &msg
		}
	}

	if r.Body.NotifyOnComment != nil {
		if err := h.Cfg.DB.UpdateNotifyOnComment(ctx, db.UpdateNotifyOnCommentParams{
			NotifyOnComment: *r.Body.NotifyOnComment,
			ID:              userID,
		}); err != nil {
			slog.Error("UpdateProfile: db error updating notify_on_comment", slog.Any("error", err), slog.String("user_id", userID.String()))
			return api.UpdateProfile400JSONResponse{Error: "Could not update notification preference"}, nil
		}
		user.NotifyOnComment = *r.Body.NotifyOnComment
	}

	return api.UpdateProfile200JSONResponse{
		User:    h.toUserResponse(ctx, user),
		Message: message,
	}, nil
}

func (h *Handler) ConfirmEmailChange(ctx context.Context, r api.ConfirmEmailChangeRequestObject) (api.ConfirmEmailChangeResponseObject, error) {
	tokenHash := auth.HashToken(r.Params.Token)
	emailToken, err := h.Cfg.DB.GetEmailToken(ctx, tokenHash)
	if err != nil || emailToken.Purpose != emailPurposeChangeEmail || emailToken.UsedAt.Valid || time.Now().After(emailToken.ExpiredAt) {
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			slog.Error("Confirm email change failed: database error looking up token", slog.Any("error", err))
		}
		return api.ConfirmEmailChange400JSONResponse{Error: "Invalid or expired token"}, nil
	}

	if err := h.Cfg.DB.ApplyPendingEmail(ctx, emailToken.UserID); err != nil {
		slog.Error("Confirm email change failed: could not apply pending email", slog.Any("error", err), slog.String("user_id", emailToken.UserID.String()))
		return api.ConfirmEmailChange400JSONResponse{Error: "Could not confirm email change"}, nil
	}

	if err := h.Cfg.DB.MarkEmailTokenUsed(ctx, tokenHash); err != nil {
		slog.Error("Confirm email change: failed to mark token used", slog.Any("error", err), slog.String("user_id", emailToken.UserID.String()))
	}

	if user, err := h.Cfg.DB.GetUserByID(ctx, emailToken.UserID); err != nil {
		slog.Error("Confirm email change: db error looking up user for notice", slog.Any("error", err), slog.String("user_id", emailToken.UserID.String()))
	} else {
		body, err := mailer.RenderLinkEmail(
			"Your Camagru account email was just changed to this address. If you didn't request this, contact us immediately.",
			h.Cfg.AppBaseURL+"/settings.html",
			"",
		)
		if err != nil {
			slog.Error("Confirm email change: failed to render notice email", slog.Any("error", err), slog.String("user_id", emailToken.UserID.String()))
		} else if err := h.Cfg.Mailer.Send(user.Email, "Your Camagru email address changed", body); err != nil {
			slog.Error("Confirm email change: failed to send notice email", slog.Any("error", err), slog.String("user_id", emailToken.UserID.String()))
		}
	}

	return api.ConfirmEmailChange200JSONResponse{Message: "Email address changed successfully"}, nil
}

func (h *Handler) UploadAvatar(ctx context.Context, r api.UploadAvatarRequestObject) (api.UploadAvatarResponseObject, error) {
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return api.UploadAvatar401JSONResponse{Error: "Not authenticated"}, nil
	}

	if r.Body == nil {
		return api.UploadAvatar400JSONResponse{Error: "Missing multipart body"}, nil
	}

	var imgData []byte
	for {
		part, err := r.Body.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return api.UploadAvatar400JSONResponse{Error: "Malformed multipart body"}, nil
		}
		if part.FormName() != "image" {
			continue
		}
		b, err := io.ReadAll(io.LimitReader(part, maxImageUploadSize+1))
		if err != nil {
			return api.UploadAvatar400JSONResponse{Error: "Could not read image"}, nil
		}
		if len(b) > maxImageUploadSize {
			return api.UploadAvatar400JSONResponse{Error: "Image too large"}, nil
		}
		imgData = b
	}

	if len(imgData) == 0 {
		return api.UploadAvatar400JSONResponse{Error: "image is required"}, nil
	}

	img, err := imaging.Decode(bytes.NewReader(imgData))
	if err != nil {
		return api.UploadAvatar400JSONResponse{Error: "Invalid image content"}, nil
	}

	var buf bytes.Buffer
	if err := imaging.EncodePNG(&buf, img); err != nil {
		slog.Error("UploadAvatar: encode failed", slog.Any("error", err), slog.String("user_id", userID.String()))
		return api.UploadAvatar500JSONResponse{Error: "Could not process image"}, nil
	}

	user, err := h.Cfg.DB.GetUserByID(ctx, userID)
	if err != nil {
		slog.Error("UploadAvatar: db error looking up user", slog.Any("error", err), slog.String("user_id", userID.String()))
		return api.UploadAvatar401JSONResponse{Error: "Not authenticated"}, nil
	}

	avatarPath, err := h.Cfg.Storage.Save(buf.Bytes(), ".png")
	if err != nil {
		slog.Error("UploadAvatar: save failed", slog.Any("error", err), slog.String("user_id", userID.String()))
		return api.UploadAvatar500JSONResponse{Error: "Could not save avatar"}, nil
	}

	if err := h.Cfg.DB.UpdateUserAvatar(ctx, db.UpdateUserAvatarParams{
		AvatarPath: sql.NullString{String: avatarPath, Valid: true},
		ID:         userID,
	}); err != nil {
		slog.Error("UploadAvatar: db error updating avatar", slog.Any("error", err), slog.String("user_id", userID.String()))
		return api.UploadAvatar500JSONResponse{Error: "Could not save avatar"}, nil
	}

	if user.AvatarPath.Valid {
		if err := h.Cfg.Storage.Delete(user.AvatarPath.String); err != nil {
			slog.Error("UploadAvatar: failed to remove old avatar file", slog.Any("error", err), slog.String("user_id", userID.String()))
		}
	}

	user.AvatarPath = sql.NullString{String: avatarPath, Valid: true}
	return api.UploadAvatar200JSONResponse(h.toUserResponse(ctx, user)), nil
}

func (h *Handler) DeleteAccount(ctx context.Context, r api.DeleteAccountRequestObject) (api.DeleteAccountResponseObject, error) {
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return api.DeleteAccount401JSONResponse{Error: "Not authenticated"}, nil
	}

	user, err := h.Cfg.DB.GetUserByID(ctx, userID)
	if err != nil {
		slog.Error("DeleteAccount: db error looking up user", slog.Any("error", err), slog.String("user_id", userID.String()))
		return api.DeleteAccount401JSONResponse{Error: "Not authenticated"}, nil
	}

	posts, err := h.Cfg.DB.ListPostsByUser(ctx, db.ListPostsByUserParams{UserID: userID})
	if err != nil {
		slog.Error("DeleteAccount: db error listing posts for cleanup", slog.Any("error", err), slog.String("user_id", userID.String()))
	}

	if err := h.Cfg.DB.DeleteUser(ctx, userID); err != nil {
		slog.Error("DeleteAccount: db error deleting user", slog.Any("error", err), slog.String("user_id", userID.String()))
		return api.DeleteAccount401JSONResponse{Error: "Could not delete account"}, nil
	}

	for _, p := range posts {
		if err := h.Cfg.Storage.Delete(p.ImagePath); err != nil {
			slog.Error("DeleteAccount: failed to remove post image file", slog.Any("error", err), slog.String("post_id", p.ID.String()))
		}
	}
	if user.AvatarPath.Valid {
		if err := h.Cfg.Storage.Delete(user.AvatarPath.String); err != nil {
			slog.Error("DeleteAccount: failed to remove avatar file", slog.Any("error", err), slog.String("user_id", userID.String()))
		}
	}

	killAccessToken := "access_token=; Path=/; Max-Age=0; HttpOnly; Secure; SameSite=Strict"
	killRefreshToken := "refresh_token=; Path=/; Max-Age=0; HttpOnly; Secure; SameSite=Strict"
	cookieHeader := killAccessToken + "\n" + killRefreshToken

	return api.DeleteAccount200JSONResponse{
		Body:    api.SuccessResponse{Message: "Account deleted"},
		Headers: api.DeleteAccount200ResponseHeaders{SetCookie: &cookieHeader},
	}, nil
}
