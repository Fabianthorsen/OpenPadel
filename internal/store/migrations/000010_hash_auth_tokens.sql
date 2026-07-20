-- +goose Up
-- +goose StatementBegin

-- Auth tokens are now stored as a SHA-256 hash (never the raw bearer token) and
-- carry a sliding expiry, mirroring password_reset_tokens (#240). Existing
-- plaintext rows are dropped, so everyone re-logs in once after this deploy.
DROP TABLE IF EXISTS auth_tokens;
CREATE TABLE auth_tokens (
    token_hash TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users(id),
    created_at TEXT NOT NULL,
    expires_at TEXT NOT NULL
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS auth_tokens;
CREATE TABLE auth_tokens (
    token      TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users(id),
    created_at TEXT NOT NULL
);

-- +goose StatementEnd
