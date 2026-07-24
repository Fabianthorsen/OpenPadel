-- +goose Up
-- +goose StatementBegin

-- Per-player secret (stored as a SHA-256 hash) issued to the joining client at
-- join time and required to self-remove a guest, replacing the guessable
-- X-Player-Id header as the authorization signal (#241). Never returned in the
-- shared session listing. Pre-existing players default to an empty hash, so they
-- can no longer self-remove by token until they rejoin (admins can still remove
-- them); acceptable given self-removal is lobby-only and reversible.
ALTER TABLE players ADD COLUMN token_hash TEXT NOT NULL DEFAULT '';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE players DROP COLUMN token_hash;

-- +goose StatementEnd
