-- +goose Up
-- +goose StatementBegin

-- Marks a Player added by hand via an admin token (a guest the admin created),
-- as opposed to a self-joined guest or a registered user. Only these players
-- may have their rating inline-edited by the admin (#211).
ALTER TABLE players ADD COLUMN added_by_admin INTEGER NOT NULL DEFAULT 0;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE players DROP COLUMN added_by_admin;

-- +goose StatementEnd
