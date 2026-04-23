-- +goose Up
-- +goose StatementBegin

UPDATE sessions SET status = 'lobby' WHERE status = 'setup' OR status = 'lobby';
UPDATE sessions SET status = 'playing' WHERE status = 'active';
UPDATE sessions SET status = 'done' WHERE status = 'complete';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

UPDATE sessions SET status = 'lobby' WHERE status = 'lobby';
UPDATE sessions SET status = 'active' WHERE status = 'playing';
UPDATE sessions SET status = 'complete' WHERE status = 'done';

-- +goose StatementEnd
