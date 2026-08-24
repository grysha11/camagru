package handler

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/grysha11/camagru-backend/internal/api"
	"github.com/grysha11/camagru-backend/internal/db"
)

func seedTestPost(t *testing.T, h *Handler, userID uuid.UUID) db.Post {
	t.Helper()
	p, err := h.Cfg.DB.CreatePost(context.Background(), db.CreatePostParams{
		UserID:    userID,
		ImagePath: "/uploads/test.png",
	})
	if err != nil {
		t.Fatalf("seed post: %v", err)
	}
	return p
}

func TestDeletePostNotOwner(t *testing.T) {
	h := newTestHandler(t)

	owner := seedVerifiedUser(t, h, "postowner", "postowner@example.com", "Correct$123")
	other := seedVerifiedUser(t, h, "notowner", "notowner@example.com", "Correct$123")
	post := seedTestPost(t, h, owner.ID)

	resp, err := h.DeletePost(authContext(other.ID), api.DeletePostRequestObject{Id: post.ID})
	if err != nil {
		t.Fatalf("DeletePost: %v", err)
	}
	if _, ok := resp.(api.DeletePost403JSONResponse); !ok {
		t.Fatalf("DeletePost response = %T, want 403", resp)
	}

	if _, err := h.Cfg.DB.GetPostByID(context.Background(), post.ID); err != nil {
		t.Errorf("post should still exist after a rejected delete: %v", err)
	}
}

func TestDeletePostNotFound(t *testing.T) {
	h := newTestHandler(t)

	owner := seedVerifiedUser(t, h, "postowner2", "postowner2@example.com", "Correct$123")

	resp, err := h.DeletePost(authContext(owner.ID), api.DeletePostRequestObject{Id: uuid.New()})
	if err != nil {
		t.Fatalf("DeletePost: %v", err)
	}
	if _, ok := resp.(api.DeletePost404JSONResponse); !ok {
		t.Fatalf("DeletePost response = %T, want 404", resp)
	}
}

func TestDeletePostOwnerSucceeds(t *testing.T) {
	h := newTestHandler(t)

	owner := seedVerifiedUser(t, h, "postowner3", "postowner3@example.com", "Correct$123")
	post := seedTestPost(t, h, owner.ID)

	resp, err := h.DeletePost(authContext(owner.ID), api.DeletePostRequestObject{Id: post.ID})
	if err != nil {
		t.Fatalf("DeletePost: %v", err)
	}
	if _, ok := resp.(api.DeletePost200JSONResponse); !ok {
		t.Fatalf("DeletePost response = %T, want 200", resp)
	}

	if _, err := h.Cfg.DB.GetPostByID(context.Background(), post.ID); err == nil {
		t.Error("post should no longer exist after a successful delete")
	}
}
