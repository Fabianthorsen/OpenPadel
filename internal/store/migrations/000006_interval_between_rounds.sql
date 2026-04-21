-- +goose Up
-- +goose StatementBegin

ALTER TABLE sessions ADD COLUMN interval_between_rounds_minutes INTEGER;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE sessions DROP COLUMN interval_between_rounds_minutes;

-- +goose StatementEnd
