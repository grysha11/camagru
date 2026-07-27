package handler

import (
	"context"

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
		return api.GetMe401JSONResponse{Error: "Not authenticated"}, nil
	}

	return api.GetMe200JSONResponse{
		Id:       user.ID,
		Username: user.Username,
		Email:    types.Email(user.Email),
	}, nil
}
