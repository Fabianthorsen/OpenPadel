-- +goose Up
-- +goose StatementBegin

ALTER TABLE sessions ADD COLUMN total_duration_minutes INTEGER;
ALTER TABLE sessions ADD COLUMN buffer_seconds INTEGER;
ALTER TABLE sessions ADD COLUMN round_duration_seconds INTEGER;
ALTER TABLE sessions ADD COLUMN round_started_at TEXT;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE sessions DROP COLUMN total_duration_minutes;
ALTER TABLE sessions DROP COLUMN buffer_seconds;
ALTER TABLE sessions DROP COLUMN round_duration_seconds;
ALTER TABLE sessions DROP COLUMN round_started_at;

-- +goose StatementEnd
