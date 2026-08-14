package handler

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/lib/pq"

	"github.com/grysha11/camagru-backend/internal/api"
	"github.com/grysha11/camagru-backend/internal/db"
	"github.com/grysha11/camagru-backend/internal/middleware"
)

func (h *Handler) LikePost(ctx context.Context, r api.LikePostRequestObject) (api.LikePostResponseObject, error) {
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return api.LikePost401JSONResponse{Error: "Not authenticated"}, nil
	}

	if _, err := h.Cfg.DB.GetPostByID(ctx, r.Id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return api.LikePost404JSONResponse{Error: "Post not found"}, nil
		}
		slog.Error("LikePost: db error looking up post", slog.Any("error", err), slog.String("post_id", r.Id.String()))
		return api.LikePost404JSONResponse{Error: "Post not found"}, nil
	}

	if err := h.Cfg.DB.CreateLike(ctx, db.CreateLikeParams{PostID: r.Id, UserID: userID}); err != nil {
		var pqErr *pq.Error
		if !errors.As(err, &pqErr) || pqErr.Constraint != "likes_pkey" {
			slog.Error("LikePost: db error creating like", slog.Any("error", err), slog.String("post_id", r.Id.String()))
			return api.LikePost404JSONResponse{Error: "Post not found"}, nil
		}
	}

	return api.LikePost200JSONResponse{Message: "Post liked"}, nil
}

func (h *Handler) UnlikePost(ctx context.Context, r api.UnlikePostRequestObject) (api.UnlikePostResponseObject, error) {
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return api.UnlikePost401JSONResponse{Error: "Not authenticated"}, nil
	}

	if _, err := h.Cfg.DB.GetPostByID(ctx, r.Id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return api.UnlikePost404JSONResponse{Error: "Post not found"}, nil
		}
		slog.Error("UnlikePost: db error looking up post", slog.Any("error", err), slog.String("post_id", r.Id.String()))
		return api.UnlikePost404JSONResponse{Error: "Post not found"}, nil
	}

	if err := h.Cfg.DB.DeleteLike(ctx, db.DeleteLikeParams{PostID: r.Id, UserID: userID}); err != nil {
		slog.Error("UnlikePost: db error deleting like", slog.Any("error", err), slog.String("post_id", r.Id.String()))
	}

	return api.UnlikePost200JSONResponse{Message: "Post unliked"}, nil
}
