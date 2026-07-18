-- +goose Up
-- +goose StatementBegin

ALTER TABLE players ADD COLUMN rating INTEGER NOT NULL DEFAULT 3;
ALTER TABLE users ADD COLUMN self_rating INTEGER;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE users DROP COLUMN self_rating;
ALTER TABLE players DROP COLUMN rating;

-- +goose StatementEnd
