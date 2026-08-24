package handler

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/oapi-codegen/runtime/types"

	"github.com/grysha11/camagru-backend/internal/api"
	"github.com/grysha11/camagru-backend/internal/db"
)

func TestUpdateProfileStagesPendingEmail(t *testing.T) {
	h := newTestHandler(t)

	u := seedVerifiedUser(t, h, "profileuser", "old@example.com", "Correct$123")

	newEmail := types.Email("new@example.com")
	resp, err := h.UpdateProfile(authContext(u.ID), api.UpdateProfileRequestObject{
		Body: &api.UpdateProfileRequest{Email: &newEmail},
	})
	if err != nil {
		t.Fatalf("UpdateProfile: %v", err)
	}
	success, ok := resp.(api.UpdateProfile200JSONResponse)
	if !ok {
		t.Fatalf("UpdateProfile response = %T, want 200", resp)
	}
	if success.Message == nil || *success.Message == "" {
		t.Error("expected a confirmation message when staging an email change")
	}
	if success.User.Email != types.Email("old@example.com") {
		t.Errorf("response email = %q, want the still-current address %q", success.User.Email, "old@example.com")
	}

	got, err := h.Cfg.DB.GetUserByID(context.Background(), u.ID)
	if err != nil {
		t.Fatalf("GetUserByID: %v", err)
	}
	if got.Email != "old@example.com" {
		t.Errorf("db email = %q, want unchanged %q", got.Email, "old@example.com")
	}
	if !got.PendingEmail.Valid || got.PendingEmail.String != "new@example.com" {
		t.Errorf("db pending_email = %+v, want %q", got.PendingEmail, "new@example.com")
	}
}

func TestDeleteAccountRemovesUser(t *testing.T) {
	h := newTestHandler(t)

	u := seedVerifiedUser(t, h, "deleteme", "deleteme@example.com", "Correct$123")

	resp, err := h.DeleteAccount(authContext(u.ID), api.DeleteAccountRequestObject{})
	if err != nil {
		t.Fatalf("DeleteAccount: %v", err)
	}
	if _, ok := resp.(api.DeleteAccount200JSONResponse); !ok {
		t.Fatalf("DeleteAccount response = %T, want 200", resp)
	}

	if _, err := h.Cfg.DB.GetUserByID(context.Background(), u.ID); err == nil {
		t.Error("expected user row to be gone after DeleteAccount")
	}
}

func TestDeleteAccountCleansUpPostsAndAvatar(t *testing.T) {
	h, storageDir := newTestHandlerWithStorageDir(t)

	u := seedVerifiedUser(t, h, "deletewithposts", "deletewithposts@example.com", "Correct$123")

	avatarPath, err := h.Cfg.Storage.Save([]byte("avatar-bytes"), ".png")
	if err != nil {
		t.Fatalf("seed avatar: %v", err)
	}
	if err := h.Cfg.DB.UpdateUserAvatar(context.Background(), db.UpdateUserAvatarParams{
		AvatarPath: sql.NullString{String: avatarPath, Valid: true},
		ID:         u.ID,
	}); err != nil {
		t.Fatalf("UpdateUserAvatar: %v", err)
	}

	postImagePath, err := h.Cfg.Storage.Save([]byte("post-bytes"), ".png")
	if err != nil {
		t.Fatalf("seed post image: %v", err)
	}
	if _, err := h.Cfg.DB.CreatePost(context.Background(), db.CreatePostParams{UserID: u.ID, ImagePath: postImagePath}); err != nil {
		t.Fatalf("seed post: %v", err)
	}

	avatarOnDisk := filepath.Join(storageDir, filepath.Base(avatarPath))
	postImageOnDisk := filepath.Join(storageDir, filepath.Base(postImagePath))
	if _, err := os.Stat(avatarOnDisk); err != nil {
		t.Fatalf("avatar file should exist before delete: %v", err)
	}
	if _, err := os.Stat(postImageOnDisk); err != nil {
		t.Fatalf("post image file should exist before delete: %v", err)
	}

	resp, err := h.DeleteAccount(authContext(u.ID), api.DeleteAccountRequestObject{})
	if err != nil {
		t.Fatalf("DeleteAccount: %v", err)
	}
	if _, ok := resp.(api.DeleteAccount200JSONResponse); !ok {
		t.Fatalf("DeleteAccount response = %T, want 200", resp)
	}

	if _, err := os.Stat(avatarOnDisk); !os.IsNotExist(err) {
		t.Errorf("avatar file should be removed after DeleteAccount, stat err = %v", err)
	}
	if _, err := os.Stat(postImageOnDisk); !os.IsNotExist(err) {
		t.Errorf("post image file should be removed after DeleteAccount, stat err = %v", err)
	}
}
