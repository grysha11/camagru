-- name: CreateEmailToken :exec
INSERT INTO email_tokens (token_hash, user_id, purpose, expired_at)
VALUES (
    $1,
    $2,
    $3,
    $4
);

-- name: GetEmailToken :one
SELECT * FROM email_tokens
WHERE token_hash = $1 LIMIT 1;

-- name: MarkEmailTokenUsed :exec
UPDATE email_tokens
SET used_at = NOW()
WHERE token_hash = $1;

-- name: DeleteEmailTokensByUserAndPurpose :exec
DELETE FROM email_tokens
WHERE user_id = $1 AND purpose = $2;
