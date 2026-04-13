-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS users (
    id            TEXT PRIMARY KEY,
    email         TEXT NOT NULL UNIQUE,
    display_name  TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    created_at    TEXT NOT NULL,
    avatar_icon   TEXT NOT NULL DEFAULT '',
    avatar_color  TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS auth_tokens (
    token      TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users(id),
    created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS password_reset_tokens (
    token_hash TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users(id),
    expires_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS sessions (
    id                     TEXT PRIMARY KEY,
    admin_token            TEXT NOT NULL,
    status                 TEXT NOT NULL DEFAULT 'lobby',
    courts                 INTEGER NOT NULL,
    points                 INTEGER NOT NULL,
    rounds_total           INTEGER,
    creator_player_id      TEXT,
    created_at             TEXT NOT NULL,
    updated_at             TEXT NOT NULL,
    current_round          INTEGER NOT NULL DEFAULT 1,
    name                   TEXT NOT NULL DEFAULT '',
    scheduled_at           TEXT,
    game_mode              TEXT NOT NULL DEFAULT 'americano',
    sets_to_win            INTEGER NOT NULL DEFAULT 2,
    games_per_set          INTEGER NOT NULL DEFAULT 6,
    ended_early            INTEGER NOT NULL DEFAULT 0,
    court_duration_minutes INTEGER,
    ends_at                TEXT
);

CREATE TABLE IF NOT EXISTS players (
    id            TEXT PRIMARY KEY,
    session_id    TEXT NOT NULL REFERENCES sessions(id),
    name          TEXT NOT NULL,
    active        INTEGER NOT NULL DEFAULT 1,
    joined_at     TEXT NOT NULL,
    user_id       TEXT REFERENCES users(id),
    avatar_icon   TEXT NOT NULL DEFAULT '',
    avatar_color  TEXT NOT NULL DEFAULT ''
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_players_session_name ON players(session_id, name) WHERE active = 1;
CREATE INDEX IF NOT EXISTS idx_players_session ON players(session_id);

CREATE TABLE IF NOT EXISTS rounds (
    id         TEXT PRIMARY KEY,
    session_id TEXT NOT NULL REFERENCES sessions(id),
    number     INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_rounds_session ON rounds(session_id);

CREATE TABLE IF NOT EXISTS bench (
    round_id  TEXT NOT NULL REFERENCES rounds(id),
    player_id TEXT NOT NULL REFERENCES players(id),
    PRIMARY KEY (round_id, player_id)
);

CREATE TABLE IF NOT EXISTS matches (
    id       TEXT PRIMARY KEY,
    round_id TEXT NOT NULL REFERENCES rounds(id),
    court    INTEGER NOT NULL,
    p1       TEXT NOT NULL REFERENCES players(id),
    p2       TEXT NOT NULL REFERENCES players(id),
    p3       TEXT NOT NULL REFERENCES players(id),
    p4       TEXT NOT NULL REFERENCES players(id),
    score_a  INTEGER,
    score_b  INTEGER,
    live_a   INTEGER,
    live_b   INTEGER,
    server   TEXT
);

CREATE INDEX IF NOT EXISTS idx_matches_round ON matches(round_id);

CREATE TABLE IF NOT EXISTS push_subscriptions (
    id         TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    endpoint   TEXT NOT NULL UNIQUE,
    p256dh     TEXT NOT NULL,
    auth       TEXT NOT NULL,
    created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS tennis_teams (
    session_id TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    player_id  TEXT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    team       TEXT NOT NULL,
    PRIMARY KEY (session_id, player_id)
);

CREATE TABLE IF NOT EXISTS tennis_matches (
    id         TEXT PRIMARY KEY,
    session_id TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    state      TEXT NOT NULL DEFAULT '{}',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS contacts (
    user_id         TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    contact_user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at      TEXT NOT NULL,
    PRIMARY KEY (user_id, contact_user_id)
);

CREATE TABLE IF NOT EXISTS invites (
    id           TEXT PRIMARY KEY,
    session_id   TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    from_user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    to_user_id   TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status       TEXT NOT NULL DEFAULT 'pending',
    created_at   TEXT NOT NULL,
    UNIQUE (session_id, to_user_id)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS invites;
DROP TABLE IF EXISTS contacts;
DROP TABLE IF EXISTS tennis_matches;
DROP TABLE IF EXISTS tennis_teams;
DROP TABLE IF EXISTS push_subscriptions;
DROP INDEX IF EXISTS idx_matches_round;
DROP TABLE IF EXISTS matches;
DROP TABLE IF EXISTS bench;
DROP INDEX IF EXISTS idx_rounds_session;
DROP TABLE IF EXISTS rounds;
DROP INDEX IF EXISTS idx_players_session_name;
DROP INDEX IF EXISTS idx_players_session;
DROP TABLE IF EXISTS players;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS password_reset_tokens;
DROP TABLE IF EXISTS auth_tokens;
DROP TABLE IF EXISTS users;

-- +goose StatementEnd
