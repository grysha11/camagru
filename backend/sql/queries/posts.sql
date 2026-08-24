-- name: CreatePost :one
INSERT INTO posts (user_id, image_path, overlay_id)
VALUES (
    $1,
    $2,
    $3
)
RETURNING *;

-- name: GetPostByID :one
SELECT * FROM posts
WHERE id = $1 LIMIT 1;

-- name: ListPostsByUser :many
SELECT
    posts.id,
    posts.user_id,
    users.username,
    posts.image_path,
    posts.overlay_id,
    posts.created_at,
    COUNT(DISTINCT likes.user_id) AS like_count,
    COUNT(DISTINCT comments.id) AS comment_count,
    COALESCE(BOOL_OR(likes.user_id = posts.user_id), false)::bool AS liked_by_me
FROM posts
JOIN users ON users.id = posts.user_id
LEFT JOIN likes ON likes.post_id = posts.id
LEFT JOIN comments ON comments.post_id = posts.id
WHERE posts.user_id = $1
GROUP BY posts.id, users.username
ORDER BY posts.created_at DESC;

-- name: CountPosts :one
SELECT COUNT(*) FROM posts;

-- name: DeletePost :execrows
DELETE FROM posts
WHERE id = $1 AND user_id = $2;

-- name: ListPosts :many
SELECT
    posts.id,
    posts.user_id,
    users.username,
    posts.image_path,
    posts.overlay_id,
    posts.created_at,
    COUNT(DISTINCT likes.user_id) AS like_count,
    COUNT(DISTINCT comments.id) AS comment_count,
    COALESCE(BOOL_OR(likes.user_id = sqlc.narg('viewer_id')::uuid), false)::bool AS liked_by_me
FROM posts
JOIN users ON users.id = posts.user_id
LEFT JOIN likes ON likes.post_id = posts.id
LEFT JOIN comments ON comments.post_id = posts.id
GROUP BY posts.id, users.username
ORDER BY posts.created_at DESC
LIMIT $1 OFFSET $2;

-- name: GetPostSummary :one
SELECT
    posts.id,
    posts.user_id,
    users.username,
    posts.image_path,
    posts.overlay_id,
    posts.created_at,
    COUNT(DISTINCT likes.user_id) AS like_count,
    COUNT(DISTINCT comments.id) AS comment_count,
    COALESCE(BOOL_OR(likes.user_id = sqlc.narg('viewer_id')::uuid), false)::bool AS liked_by_me
FROM posts
JOIN users ON users.id = posts.user_id
LEFT JOIN likes ON likes.post_id = posts.id
LEFT JOIN comments ON comments.post_id = posts.id
WHERE posts.id = $1
GROUP BY posts.id, users.username;

-- name: GetPostOwnerNotifyInfo :one
SELECT users.id AS user_id, users.email, users.notify_on_comment
FROM posts
JOIN users ON users.id = posts.user_id
WHERE posts.id = $1;
