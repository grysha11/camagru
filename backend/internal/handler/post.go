package handler

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"io"
	"log/slog"
	"strings"

	"github.com/google/uuid"

	"github.com/grysha11/camagru-backend/internal/api"
	"github.com/grysha11/camagru-backend/internal/db"
	"github.com/grysha11/camagru-backend/internal/imaging"
	"github.com/grysha11/camagru-backend/internal/middleware"
)

const maxImageUploadSize = 8 << 20 // 8 MiB
const postsPerPage = 10

func (h *Handler) CreatePost(ctx context.Context, r api.CreatePostRequestObject) (api.CreatePostResponseObject, error) {
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return api.CreatePost401JSONResponse{Error: "Not authenticated"}, nil
	}

	if r.Body == nil {
		return api.CreatePost400JSONResponse{Error: "Missing multipart body"}, nil
	}

	var overlayID string
	var imgData []byte

	for {
		part, err := r.Body.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return api.CreatePost400JSONResponse{Error: "Malformed multipart body"}, nil
		}

		switch part.FormName() {
		case "overlay_id":
			b, err := io.ReadAll(io.LimitReader(part, 256))
			if err != nil {
				return api.CreatePost400JSONResponse{Error: "Could not read overlay_id"}, nil
			}
			overlayID = strings.TrimSpace(string(b))
		case "image":
			b, err := io.ReadAll(io.LimitReader(part, maxImageUploadSize+1))
			if err != nil {
				return api.CreatePost400JSONResponse{Error: "Could not read image"}, nil
			}
			if len(b) > maxImageUploadSize {
				return api.CreatePost400JSONResponse{Error: "Image too large"}, nil
			}
			imgData = b
		}
	}

	if overlayID == "" || len(imgData) == 0 {
		return api.CreatePost400JSONResponse{Error: "image and overlay_id are both required"}, nil
	}

	entry, ok := h.Cfg.Overlays.Get(overlayID)
	if !ok {
		return api.CreatePost400JSONResponse{Error: "Unknown overlay_id"}, nil
	}

	img, err := imaging.Decode(bytes.NewReader(imgData))
	if err != nil {
		return api.CreatePost400JSONResponse{Error: "Invalid image content"}, nil
	}

	if img.Bounds().Dx() != entry.Image.Bounds().Dx() || img.Bounds().Dy() != entry.Image.Bounds().Dy() {
		return api.CreatePost400JSONResponse{Error: "Image dimensions must match the overlay"}, nil
	}

	composite := imaging.Composite(img, entry.Image)

	var buf bytes.Buffer
	if err := imaging.EncodePNG(&buf, composite); err != nil {
		slog.Error("CreatePost: encode failed", slog.Any("error", err))
		return api.CreatePost500JSONResponse{Error: "Could not process image"}, nil
	}

	imagePath, err := h.Cfg.Storage.Save(buf.Bytes(), ".png")
	if err != nil {
		slog.Error("CreatePost: save failed", slog.Any("error", err))
		return api.CreatePost500JSONResponse{Error: "Could not save image"}, nil
	}

	post, err := h.Cfg.DB.CreatePost(ctx, db.CreatePostParams{
		UserID:    userID,
		ImagePath: imagePath,
		OverlayID: sql.NullString{String: overlayID, Valid: true},
	})
	if err != nil {
		slog.Error("CreatePost: db insert failed", slog.Any("error", err), slog.String("user_id", userID.String()))
		return api.CreatePost500JSONResponse{Error: "Could not save post"}, nil
	}

	return api.CreatePost201JSONResponse{
		Id:        post.ID,
		ImagePath: post.ImagePath,
		OverlayId: &overlayID,
		CreatedAt: post.CreatedAt,
	}, nil
}

func (h *Handler) ListMyPosts(ctx context.Context, r api.ListMyPostsRequestObject) (api.ListMyPostsResponseObject, error) {
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return api.ListMyPosts401JSONResponse{Error: "Not authenticated"}, nil
	}

	posts, err := h.Cfg.DB.ListPostsByUser(ctx, userID)
	if err != nil {
		slog.Error("ListMyPosts: db error", slog.Any("error", err), slog.String("user_id", userID.String()))
		return api.ListMyPosts500JSONResponse{Error: "Could not load posts"}, nil
	}

	resp := make([]api.PostSummary, 0, len(posts))
	for _, p := range posts {
		var overlayID *string
		if p.OverlayID.Valid {
			overlayID = &p.OverlayID.String
		}
		resp = append(resp, api.PostSummary{
			Id:           p.ID,
			UserId:       p.UserID,
			Username:     p.Username,
			ImagePath:    p.ImagePath,
			OverlayId:    overlayID,
			CreatedAt:    p.CreatedAt,
			LikeCount:    int(p.LikeCount),
			CommentCount: int(p.CommentCount),
			LikedByMe:    p.LikedByMe,
		})
	}

	return api.ListMyPosts200JSONResponse(resp), nil
}

func (h *Handler) ListPosts(ctx context.Context, r api.ListPostsRequestObject) (api.ListPostsResponseObject, error) {
	page := 1
	if r.Params.Page != nil && *r.Params.Page > 0 {
		page = *r.Params.Page
	}

	var viewerID uuid.NullUUID
	if userID, ok := middleware.UserIDFromContext(ctx); ok {
		viewerID = uuid.NullUUID{UUID: userID, Valid: true}
	}

	rows, err := h.Cfg.DB.ListPosts(ctx, db.ListPostsParams{
		Limit:    postsPerPage,
		Offset:   int32((page - 1) * postsPerPage),
		ViewerID: viewerID,
	})
	if err != nil {
		slog.Error("ListPosts: db error", slog.Any("error", err))
		return api.ListPosts500JSONResponse{Error: "Could not load posts"}, nil
	}

	totalPosts, err := h.Cfg.DB.CountPosts(ctx)
	if err != nil {
		slog.Error("ListPosts: count db error", slog.Any("error", err))
		return api.ListPosts500JSONResponse{Error: "Could not load posts"}, nil
	}

	totalPages := int((totalPosts + postsPerPage - 1) / postsPerPage)

	posts := make([]api.PostSummary, 0, len(rows))
	for _, p := range rows {
		var overlayID *string
		if p.OverlayID.Valid {
			overlayID = &p.OverlayID.String
		}
		posts = append(posts, api.PostSummary{
			Id:           p.ID,
			UserId:       p.UserID,
			Username:     p.Username,
			ImagePath:    p.ImagePath,
			OverlayId:    overlayID,
			CreatedAt:    p.CreatedAt,
			LikeCount:    int(p.LikeCount),
			CommentCount: int(p.CommentCount),
			LikedByMe:    p.LikedByMe,
		})
	}

	return api.ListPosts200JSONResponse{
		Posts:      posts,
		Page:       page,
		PageSize:   postsPerPage,
		TotalPosts: int(totalPosts),
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}, nil
}

func (h *Handler) GetPost(ctx context.Context, r api.GetPostRequestObject) (api.GetPostResponseObject, error) {
	var viewerID uuid.NullUUID
	if userID, ok := middleware.UserIDFromContext(ctx); ok {
		viewerID = uuid.NullUUID{UUID: userID, Valid: true}
	}

	p, err := h.Cfg.DB.GetPostSummary(ctx, db.GetPostSummaryParams{ID: r.Id, ViewerID: viewerID})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return api.GetPost404JSONResponse{Error: "Post not found"}, nil
		}
		slog.Error("GetPost: db error", slog.Any("error", err), slog.String("post_id", r.Id.String()))
		return api.GetPost500JSONResponse{Error: "Could not load post"}, nil
	}

	var overlayID *string
	if p.OverlayID.Valid {
		overlayID = &p.OverlayID.String
	}

	return api.GetPost200JSONResponse{
		Id:           p.ID,
		UserId:       p.UserID,
		Username:     p.Username,
		ImagePath:    p.ImagePath,
		OverlayId:    overlayID,
		CreatedAt:    p.CreatedAt,
		LikeCount:    int(p.LikeCount),
		CommentCount: int(p.CommentCount),
		LikedByMe:    p.LikedByMe,
	}, nil
}

func (h *Handler) DeletePost(ctx context.Context, r api.DeletePostRequestObject) (api.DeletePostResponseObject, error) {
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return api.DeletePost401JSONResponse{Error: "Not authenticated"}, nil
	}

	post, err := h.Cfg.DB.GetPostByID(ctx, r.Id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return api.DeletePost404JSONResponse{Error: "Post not found"}, nil
		}
		slog.Error("DeletePost: db error looking up post", slog.Any("error", err), slog.String("post_id", r.Id.String()))
		return api.DeletePost500JSONResponse{Error: "Could not delete post"}, nil
	}

	if post.UserID != userID {
		return api.DeletePost403JSONResponse{Error: "You do not own this post"}, nil
	}

	if _, err := h.Cfg.DB.DeletePost(ctx, db.DeletePostParams{ID: r.Id, UserID: userID}); err != nil {
		slog.Error("DeletePost: db error deleting post", slog.Any("error", err), slog.String("post_id", r.Id.String()))
		return api.DeletePost500JSONResponse{Error: "Could not delete post"}, nil
	}

	if err := h.Cfg.Storage.Delete(post.ImagePath); err != nil {
		slog.Error("DeletePost: failed to remove image file", slog.Any("error", err), slog.String("post_id", r.Id.String()))
	}

	return api.DeletePost200JSONResponse{Message: "Post deleted"}, nil
}
