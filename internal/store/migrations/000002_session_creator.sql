-- +goose Up
ALTER TABLE sessions ADD COLUMN creator_user_id TEXT REFERENCES users(id);

-- +goose Down
-- SQLite <3.35 does not support DROP COLUMN; leave the column in place.
