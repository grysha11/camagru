-- name: GetCommentByID :one
SELECT * FROM comments
WHERE id = $1 LIMIT 1;

-- name: DeleteComment :execrows
DELETE FROM comments
WHERE id = $1 AND user_id = $2;

-- name: CreateComment :one
INSERT INTO comments (post_id, user_id, content)
VALUES (
    $1,
    $2,
    $3
)
RETURNING *;

-- name: ListCommentsByPost :many
SELECT
    comments.id,
    comments.post_id,
    comments.user_id,
    users.username,
    comments.content,
    comments.created_at
FROM comments
JOIN users ON users.id = comments.user_id
WHERE comments.post_id = $1
ORDER BY comments.created_at ASC;
