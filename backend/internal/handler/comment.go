package handler

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/grysha11/camagru-backend/internal/api"
	"github.com/grysha11/camagru-backend/internal/db"
	"github.com/grysha11/camagru-backend/internal/middleware"
)

const maxCommentLength = 1000

func (h *Handler) ListComments(ctx context.Context, r api.ListCommentsRequestObject) (api.ListCommentsResponseObject, error) {
	if _, err := h.Cfg.DB.GetPostByID(ctx, r.Id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return api.ListComments404JSONResponse{Error: "Post not found"}, nil
		}
		slog.Error("ListComments: db error looking up post", slog.Any("error", err), slog.String("post_id", r.Id.String()))
		return api.ListComments404JSONResponse{Error: "Post not found"}, nil
	}

	rows, err := h.Cfg.DB.ListCommentsByPost(ctx, r.Id)
	if err != nil {
		slog.Error("ListComments: db error", slog.Any("error", err), slog.String("post_id", r.Id.String()))
		return api.ListComments404JSONResponse{Error: "Post not found"}, nil
	}

	resp := make([]api.CommentResponse, 0, len(rows))
	for _, c := range rows {
		resp = append(resp, api.CommentResponse{
			Id:        c.ID,
			PostId:    c.PostID,
			UserId:    c.UserID,
			Username:  c.Username,
			Content:   c.Content,
			CreatedAt: c.CreatedAt,
		})
	}

	return api.ListComments200JSONResponse(resp), nil
}

func (h *Handler) CreateComment(ctx context.Context, r api.CreateCommentRequestObject) (api.CreateCommentResponseObject, error) {
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return api.CreateComment401JSONResponse{Error: "Not authenticated"}, nil
	}

	if r.Body == nil {
		return api.CreateComment400JSONResponse{Error: "Comment content is required"}, nil
	}

	content := strings.TrimSpace(r.Body.Content)
	if content == "" {
		return api.CreateComment400JSONResponse{Error: "Comment content is required"}, nil
	}
	if len(content) > maxCommentLength {
		return api.CreateComment400JSONResponse{Error: "Comment is too long"}, nil
	}

	ownerInfo, err := h.Cfg.DB.GetPostOwnerNotifyInfo(ctx, r.Id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return api.CreateComment404JSONResponse{Error: "Post not found"}, nil
		}
		slog.Error("CreateComment: db error looking up post owner", slog.Any("error", err), slog.String("post_id", r.Id.String()))
		return api.CreateComment404JSONResponse{Error: "Post not found"}, nil
	}

	commenter, err := h.Cfg.DB.GetUserByID(ctx, userID)
	if err != nil {
		slog.Error("CreateComment: db error looking up commenter", slog.Any("error", err), slog.String("user_id", userID.String()))
		return api.CreateComment404JSONResponse{Error: "Post not found"}, nil
	}

	comment, err := h.Cfg.DB.CreateComment(ctx, db.CreateCommentParams{
		PostID:  r.Id,
		UserID:  userID,
		Content: content,
	})
	if err != nil {
		slog.Error("CreateComment: db error creating comment", slog.Any("error", err), slog.String("post_id", r.Id.String()))
		return api.CreateComment404JSONResponse{Error: "Post not found"}, nil
	}

	if ownerInfo.NotifyOnComment && ownerInfo.UserID != userID {
		subject := "New comment on your Camagru post"
		body := fmt.Sprintf("%s commented on your post:\n\n%s\n", commenter.Username, comment.Content)
		if err := h.Cfg.Mailer.Send(ownerInfo.Email, subject, body); err != nil {
			slog.Error("CreateComment: failed to send notification email", slog.Any("error", err), slog.String("post_id", r.Id.String()))
		}
	}

	return api.CreateComment201JSONResponse{
		Id:        comment.ID,
		PostId:    comment.PostID,
		UserId:    comment.UserID,
		Username:  commenter.Username,
		Content:   comment.Content,
		CreatedAt: comment.CreatedAt,
	}, nil
}

func (h *Handler) DeleteComment(ctx context.Context, r api.DeleteCommentRequestObject) (api.DeleteCommentResponseObject, error) {
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return api.DeleteComment401JSONResponse{Error: "Not authenticated"}, nil
	}

	comment, err := h.Cfg.DB.GetCommentByID(ctx, r.CommentId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return api.DeleteComment404JSONResponse{Error: "Comment not found"}, nil
		}
		slog.Error("DeleteComment: db error looking up comment", slog.Any("error", err), slog.String("comment_id", r.CommentId.String()))
		return api.DeleteComment500JSONResponse{Error: "Could not delete comment"}, nil
	}

	if comment.PostID != r.Id {
		return api.DeleteComment404JSONResponse{Error: "Comment not found"}, nil
	}

	if comment.UserID != userID {
		return api.DeleteComment403JSONResponse{Error: "You do not own this comment"}, nil
	}

	if _, err := h.Cfg.DB.DeleteComment(ctx, db.DeleteCommentParams{ID: r.CommentId, UserID: userID}); err != nil {
		slog.Error("DeleteComment: db error deleting comment", slog.Any("error", err), slog.String("comment_id", r.CommentId.String()))
		return api.DeleteComment500JSONResponse{Error: "Could not delete comment"}, nil
	}

	return api.DeleteComment200JSONResponse{Message: "Comment deleted"}, nil
}
