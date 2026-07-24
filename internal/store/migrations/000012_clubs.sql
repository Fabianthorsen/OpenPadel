-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS clubs (
    id           TEXT PRIMARY KEY,
    name         TEXT NOT NULL,
    description  TEXT NOT NULL DEFAULT '',
    avatar_icon  TEXT NOT NULL DEFAULT '',
    avatar_color TEXT NOT NULL DEFAULT '',
    join_code    TEXT NOT NULL UNIQUE,
    created_by   TEXT NOT NULL REFERENCES users(id),
    created_at   TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS club_members (
    club_id  TEXT NOT NULL REFERENCES clubs(id) ON DELETE CASCADE,
    user_id  TEXT NOT NULL REFERENCES users(id),
    role     TEXT NOT NULL DEFAULT 'member',
    joined_at TEXT NOT NULL,
    PRIMARY KEY (club_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_club_members_user ON club_members(user_id);

ALTER TABLE sessions ADD COLUMN club_id TEXT REFERENCES clubs(id) ON DELETE SET NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS club_members;
DROP TABLE IF EXISTS clubs;

ALTER TABLE sessions DROP COLUMN club_id;

-- +goose StatementEnd
