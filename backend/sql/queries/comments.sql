-- name: CreateComment :one
INSERT INTO comments (post_id, user_id, content)
VALUES (
    $1,
    $2,
    $3
)
RETURNING *;

-- name: ListCommentsByPost :many
SELECT * FROM comments
WHERE post_id = $1
ORDER BY created_at ASC;
