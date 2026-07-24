-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS club_invites (
    id         TEXT PRIMARY KEY,
    club_id    TEXT NOT NULL REFERENCES clubs(id) ON DELETE CASCADE,
    inviter_id TEXT NOT NULL REFERENCES users(id),
    invitee_id TEXT NOT NULL REFERENCES users(id),
    status     TEXT NOT NULL DEFAULT 'pending',
    created_at TEXT NOT NULL,
    UNIQUE (club_id, invitee_id)
);

CREATE INDEX IF NOT EXISTS idx_club_invites_invitee ON club_invites(invitee_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS club_invites;

-- +goose StatementEnd
