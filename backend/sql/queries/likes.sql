-- name: CreateLike :exec
INSERT INTO likes (post_id, user_id)
VALUES (
    $1,
    $2
);

-- name: DeleteLike :exec
DELETE FROM likes
WHERE post_id = $1 AND user_id = $2;

-- name: GetLike :one
SELECT * FROM likes
WHERE post_id = $1 AND user_id = $2 LIMIT 1;
