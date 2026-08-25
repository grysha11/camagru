-- name: CreateOAuthIdentity :one
INSERT INTO oauth_identities (user_id, provider, provider_user_id)
VALUES (
    $1,
    $2,
    $3
)
RETURNING *;

-- name: GetOAuthIdentityByProvider :one
SELECT * FROM oauth_identities
WHERE provider = $1 AND provider_user_id = $2 LIMIT 1;

-- name: GetOAuthIdentityByUser :one
SELECT * FROM oauth_identities
WHERE user_id = $1 AND provider = $2 LIMIT 1;

-- name: DeleteOAuthIdentity :execrows
DELETE FROM oauth_identities
WHERE user_id = $1 AND provider = $2;
