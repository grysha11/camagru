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
SELECT * FROM posts
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: CountPosts :one
SELECT COUNT(*) FROM posts;

-- name: DeletePost :execrows
DELETE FROM posts
WHERE id = $1 AND user_id = $2;

-- name: ListPosts :many
SELECT
    posts.id,
    posts.user_id,
    posts.image_path,
    posts.overlay_id,
    posts.created_at,
    COUNT(DISTINCT likes.user_id) AS like_count,
    COUNT(DISTINCT comments.id) AS comment_count
FROM posts
LEFT JOIN likes ON likes.post_id = posts.id
LEFT JOIN comments ON comments.post_id = posts.id
GROUP BY posts.id
ORDER BY posts.created_at DESC
LIMIT $1 OFFSET $2;
