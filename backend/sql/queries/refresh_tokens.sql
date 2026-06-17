-- name: CreateRefreshToken :exec
INSERT INTO refresh_tokens (token, user_id, expired_at)
VALUES (
    $1,
    $2,
    $3
);

-- name: GetRefreshToken :one
SELECT * FROM refresh_tokens
WHERE token = $1 LIMIT 1;

-- name: DeleteRefreshToken :exec
DELETE FROM refresh_tokens
WHERE token = $1;

-- name: DeleteAllUserRefreshTokens :exec
DELETE FROM refresh_tokens
WHERE user_id = $1;
