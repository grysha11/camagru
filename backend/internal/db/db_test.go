package db

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/grysha11/camagru-backend/internal/testutil"
)

func setupQueries(t *testing.T) *Queries {
	t.Helper()
	sqlDB := testutil.OpenTestDB(t)
	testutil.TruncateAll(t, sqlDB)
	return New(sqlDB)
}

func seedUser(t *testing.T, q *Queries, username, email string) User {
	t.Helper()
	u, err := q.CreateUser(context.Background(), CreateUserParams{
		Username:       username,
		Email:          email,
		HashedPassword: sql.NullString{String: "hash", Valid: true},
	})
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return u
}

func seedPost(t *testing.T, q *Queries, userID uuid.UUID) Post {
	t.Helper()
	p, err := q.CreatePost(context.Background(), CreatePostParams{
		UserID:    userID,
		ImagePath: "path.png",
		OverlayID: sql.NullString{String: "overlay.png", Valid: true},
	})
	if err != nil {
		t.Fatalf("seed post: %v", err)
	}
	return p
}

func TestCreateUserUniqueEmail(t *testing.T) {
	q := setupQueries(t)
	ctx := context.Background()

	seedUser(t, q, "alice", "alice@example.com")

	_, err := q.CreateUser(ctx, CreateUserParams{
		Username:       "alice2",
		Email:          "alice@example.com",
		HashedPassword: sql.NullString{String: "hash", Valid: true},
	})
	var pqErr *pq.Error
	if !errors.As(err, &pqErr) {
		t.Fatalf("expected *pq.Error for duplicate email, got %v", err)
	}
}

func TestCreateUserUniqueUsername(t *testing.T) {
	q := setupQueries(t)
	ctx := context.Background()

	seedUser(t, q, "bob", "bob@example.com")

	_, err := q.CreateUser(ctx, CreateUserParams{
		Username:       "bob",
		Email:          "bob2@example.com",
		HashedPassword: sql.NullString{String: "hash", Valid: true},
	})
	var pqErr *pq.Error
	if !errors.As(err, &pqErr) {
		t.Fatalf("expected *pq.Error for duplicate username, got %v", err)
	}
}

func TestApplyPendingEmail(t *testing.T) {
	q := setupQueries(t)
	ctx := context.Background()

	u := seedUser(t, q, "carol", "carol@example.com")
	if err := q.SetPendingEmail(ctx, SetPendingEmailParams{
		PendingEmail: sql.NullString{String: "carol-new@example.com", Valid: true},
		ID:           u.ID,
	}); err != nil {
		t.Fatalf("SetPendingEmail: %v", err)
	}

	if err := q.ApplyPendingEmail(ctx, u.ID); err != nil {
		t.Fatalf("ApplyPendingEmail: %v", err)
	}

	got, err := q.GetUserByID(ctx, u.ID)
	if err != nil {
		t.Fatalf("GetUserByID: %v", err)
	}
	if got.Email != "carol-new@example.com" {
		t.Errorf("email = %q, want %q", got.Email, "carol-new@example.com")
	}
	if got.PendingEmail.Valid {
		t.Errorf("pending_email still set: %+v", got.PendingEmail)
	}
}

func TestListPostsPaginationAndLikedByMe(t *testing.T) {
	q := setupQueries(t)
	ctx := context.Background()

	author := seedUser(t, q, "dave", "dave@example.com")
	viewer := seedUser(t, q, "erin", "erin@example.com")
	other := seedUser(t, q, "frank", "frank@example.com")

	var posts []Post
	for i := 0; i < 3; i++ {
		posts = append(posts, seedPost(t, q, author.ID))
	}

	if err := q.CreateLike(ctx, CreateLikeParams{PostID: posts[0].ID, UserID: viewer.ID}); err != nil {
		t.Fatalf("CreateLike: %v", err)
	}
	if err := q.CreateLike(ctx, CreateLikeParams{PostID: posts[0].ID, UserID: other.ID}); err != nil {
		t.Fatalf("CreateLike (other): %v", err)
	}

	rows, err := q.ListPosts(ctx, ListPostsParams{Limit: 2, Offset: 0, ViewerID: uuid.NullUUID{UUID: viewer.ID, Valid: true}})
	if err != nil {
		t.Fatalf("ListPosts page 1: %v", err)
	}
	if len(rows) != 2 || rows[0].ID != posts[2].ID || rows[1].ID != posts[1].ID {
		t.Fatalf("page 1 = %+v, want newest-first [%v, %v]", rows, posts[2].ID, posts[1].ID)
	}

	rows2, err := q.ListPosts(ctx, ListPostsParams{Limit: 2, Offset: 2, ViewerID: uuid.NullUUID{UUID: viewer.ID, Valid: true}})
	if err != nil {
		t.Fatalf("ListPosts page 2: %v", err)
	}
	if len(rows2) != 1 || rows2[0].ID != posts[0].ID {
		t.Fatalf("page 2 = %+v, want just posts[0]", rows2)
	}
	if !rows2[0].LikedByMe {
		t.Error("liked_by_me for viewer who liked = false, want true")
	}

	anonRows, err := q.ListPosts(ctx, ListPostsParams{Limit: 2, Offset: 2, ViewerID: uuid.NullUUID{}})
	if err != nil {
		t.Fatalf("ListPosts anonymous: %v", err)
	}
	if len(anonRows) != 1 || anonRows[0].LikedByMe {
		t.Errorf("liked_by_me for anonymous viewer = %+v, want false", anonRows)
	}
}

func TestLikeUniqueConstraint(t *testing.T) {
	q := setupQueries(t)
	ctx := context.Background()

	author := seedUser(t, q, "gina", "gina@example.com")
	liker := seedUser(t, q, "hank", "hank@example.com")
	post := seedPost(t, q, author.ID)

	if err := q.CreateLike(ctx, CreateLikeParams{PostID: post.ID, UserID: liker.ID}); err != nil {
		t.Fatalf("first CreateLike: %v", err)
	}

	err := q.CreateLike(ctx, CreateLikeParams{PostID: post.ID, UserID: liker.ID})
	var pqErr *pq.Error
	if !errors.As(err, &pqErr) {
		t.Fatalf("expected *pq.Error on duplicate like, got %v", err)
	}
}

func TestCommentOwnership(t *testing.T) {
	q := setupQueries(t)
	ctx := context.Background()

	author := seedUser(t, q, "ivan", "ivan@example.com")
	commenter := seedUser(t, q, "jane", "jane@example.com")
	post := seedPost(t, q, author.ID)

	comment, err := q.CreateComment(ctx, CreateCommentParams{PostID: post.ID, UserID: commenter.ID, Content: "hello"})
	if err != nil {
		t.Fatalf("CreateComment: %v", err)
	}

	got, err := q.GetCommentByID(ctx, comment.ID)
	if err != nil {
		t.Fatalf("GetCommentByID: %v", err)
	}
	if got.UserID != commenter.ID {
		t.Errorf("comment owner = %v, want %v", got.UserID, commenter.ID)
	}
	if got.PostID != post.ID {
		t.Errorf("comment post = %v, want %v", got.PostID, post.ID)
	}
}
