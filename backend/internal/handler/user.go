package handler

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/google/uuid"
	"github.com/oapi-codegen/runtime/types"

	"github.com/grysha11/camagru-backend/internal/api"
	"github.com/grysha11/camagru-backend/internal/middleware"
)

func (h *Handler) GetMe(ctx context.Context, r api.GetMeRequestObject) (api.GetMeResponseObject, error) {
	userIDStr, ok := ctx.Value(middleware.UserIDKey).(string)
	if !ok || userIDStr == "" {
		return api.GetMe401JSONResponse{Error: "Not authenticated"}, nil
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
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

	return api.GetMe200JSONResponse{
		Id:       user.ID,
		Username: user.Username,
		Email:    types.Email(user.Email),
	}, nil
}
