-- +goose Up
ALTER TABLE users ADD COLUMN avatar_path TEXT;
ALTER TABLE users ADD COLUMN pending_email TEXT;

-- +goose Down
ALTER TABLE users DROP COLUMN pending_email;
ALTER TABLE users DROP COLUMN avatar_path;
