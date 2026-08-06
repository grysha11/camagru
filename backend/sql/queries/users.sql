-- name: CreateUser :one
INSERT INTO users (username, email, hashed_password)
VALUES (
    $1,
    $2,
    $3
)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: MarkEmailVerified :exec
UPDATE users
SET email_verified_at = NOW()
WHERE id = $1;

-- name: UpdateNotifyOnComment :exec
UPDATE users
SET notify_on_comment = $1
WHERE id = $2;

-- name: UpdateUserPassword :exec
UPDATE users
SET hashed_password = $1, updated_at = NOW()
WHERE id = $2;