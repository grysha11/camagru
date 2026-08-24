package handler

import (
	"context"
	"testing"

	"github.com/grysha11/camagru-backend/internal/api"
	"github.com/grysha11/camagru-backend/internal/db"
)

func TestLikePostIdempotent(t *testing.T) {
	h := newTestHandler(t)

	owner := seedVerifiedUser(t, h, "likeowner", "likeowner@example.com", "Correct$123")
	liker := seedVerifiedUser(t, h, "liker", "liker@example.com", "Correct$123")
	post := seedTestPost(t, h, owner.ID)

	resp1, err := h.LikePost(authContext(liker.ID), api.LikePostRequestObject{Id: post.ID})
	if err != nil {
		t.Fatalf("first LikePost: %v", err)
	}
	if _, ok := resp1.(api.LikePost200JSONResponse); !ok {
		t.Fatalf("first LikePost response = %T, want 200", resp1)
	}

	resp2, err := h.LikePost(authContext(liker.ID), api.LikePostRequestObject{Id: post.ID})
	if err != nil {
		t.Fatalf("second LikePost: %v", err)
	}
	if _, ok := resp2.(api.LikePost200JSONResponse); !ok {
		t.Fatalf("second LikePost response = %T, want 200", resp2)
	}

	if _, err := h.Cfg.DB.GetLike(context.Background(), db.GetLikeParams{PostID: post.ID, UserID: liker.ID}); err != nil {
		t.Errorf("expected exactly one like row to exist: %v", err)
	}
}

func TestUnlikePostThenRelike(t *testing.T) {
	h := newTestHandler(t)

	owner := seedVerifiedUser(t, h, "likeowner2", "likeowner2@example.com", "Correct$123")
	liker := seedVerifiedUser(t, h, "liker2", "liker2@example.com", "Correct$123")
	post := seedTestPost(t, h, owner.ID)

	if _, err := h.LikePost(authContext(liker.ID), api.LikePostRequestObject{Id: post.ID}); err != nil {
		t.Fatalf("LikePost: %v", err)
	}
	if _, err := h.UnlikePost(authContext(liker.ID), api.UnlikePostRequestObject{Id: post.ID}); err != nil {
		t.Fatalf("UnlikePost: %v", err)
	}
	if _, err := h.Cfg.DB.GetLike(context.Background(), db.GetLikeParams{PostID: post.ID, UserID: liker.ID}); err == nil {
		t.Error("expected like row to be gone after unlike")
	}

	resp, err := h.LikePost(authContext(liker.ID), api.LikePostRequestObject{Id: post.ID})
	if err != nil {
		t.Fatalf("re-LikePost: %v", err)
	}
	if _, ok := resp.(api.LikePost200JSONResponse); !ok {
		t.Fatalf("re-LikePost response = %T, want 200", resp)
	}
}
