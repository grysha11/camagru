-- +goose Up
ALTER TABLE users ADD COLUMN email_verified_at TIMESTAMP;
ALTER TABLE users ADD COLUMN notify_on_comment BOOLEAN NOT NULL DEFAULT TRUE;
ALTER TABLE users ALTER COLUMN hashed_password DROP NOT NULL;

-- +goose Down
ALTER TABLE users ALTER COLUMN hashed_password SET NOT NULL;
ALTER TABLE users DROP COLUMN notify_on_comment;
ALTER TABLE users DROP COLUMN email_verified_at;
