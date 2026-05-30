-- name: CreateUser :one
INSERT INTO users (username, email)
VALUES (
    $1,
    $2
)
RETURNING *;
