package handler

import (
	"bytes"
	"context"
	"database/sql"
	"io"
	"log/slog"
	"strings"

	"github.com/grysha11/camagru-backend/internal/api"
	"github.com/grysha11/camagru-backend/internal/db"
	"github.com/grysha11/camagru-backend/internal/imaging"
	"github.com/grysha11/camagru-backend/internal/middleware"
)

const maxImageUploadSize = 8 << 20 // 8 MiB

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

	resp := make([]api.PostResponse, 0, len(posts))
	for _, p := range posts {
		var overlayID *string
		if p.OverlayID.Valid {
			overlayID = &p.OverlayID.String
		}
		resp = append(resp, api.PostResponse{
			Id:        p.ID,
			ImagePath: p.ImagePath,
			OverlayId: overlayID,
			CreatedAt: p.CreatedAt,
		})
	}

	return api.ListMyPosts200JSONResponse(resp), nil
}
