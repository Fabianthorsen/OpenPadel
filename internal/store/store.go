package store

import (
	"database/sql"
	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path+"?_journal_mode=WAL&_foreign_keys=on")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) migrate() error {
	if _, err := s.db.Exec(schema); err != nil {
		return err
	}
	// Additive column migrations — ignore "duplicate column" errors.
	for _, stmt := range migrations {
		s.db.Exec(stmt) //nolint:errcheck
	}
	return nil
}

// migrations contains ALTER TABLE statements for columns added after initial schema.
// SQLite has no IF NOT EXISTS for ALTER TABLE, so we run and ignore duplicate errors.
var migrations = []string{
	`ALTER TABLE sessions ADD COLUMN creator_player_id TEXT`,
	`ALTER TABLE sessions ADD COLUMN current_round INTEGER NOT NULL DEFAULT 1`,
	`ALTER TABLE sessions ADD COLUMN name TEXT NOT NULL DEFAULT ''`,
	`ALTER TABLE players ADD COLUMN user_id TEXT REFERENCES users(id)`,
	`ALTER TABLE sessions ADD COLUMN scheduled_at TEXT`,
	`DROP INDEX IF EXISTS idx_players_session_name`,
	`CREATE UNIQUE INDEX IF NOT EXISTS idx_players_session_name ON players(session_id, name) WHERE active = 1`,
	`ALTER TABLE matches ADD COLUMN live_a INTEGER`,
	`ALTER TABLE matches ADD COLUMN live_b INTEGER`,
	`ALTER TABLE matches ADD COLUMN server TEXT`,
	`CREATE TABLE IF NOT EXISTS push_subscriptions (
		id         TEXT PRIMARY KEY,
		user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		endpoint   TEXT NOT NULL UNIQUE,
		p256dh     TEXT NOT NULL,
		auth       TEXT NOT NULL,
		created_at TEXT NOT NULL
	)`,
	`ALTER TABLE sessions ADD COLUMN game_mode TEXT NOT NULL DEFAULT 'americano'`,
	`ALTER TABLE sessions ADD COLUMN sets_to_win INTEGER NOT NULL DEFAULT 2`,
	`ALTER TABLE sessions ADD COLUMN games_per_set INTEGER NOT NULL DEFAULT 6`,
	`CREATE TABLE IF NOT EXISTS tennis_teams (
		session_id TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
		player_id  TEXT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
		team       TEXT NOT NULL,
		PRIMARY KEY (session_id, player_id)
	)`,
	`CREATE TABLE IF NOT EXISTS tennis_matches (
		id         TEXT PRIMARY KEY,
		session_id TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
		state      TEXT NOT NULL DEFAULT '{}',
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL
	)`,
	`CREATE TABLE IF NOT EXISTS contacts (
		user_id         TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		contact_user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		created_at      TEXT NOT NULL,
		PRIMARY KEY (user_id, contact_user_id)
	)`,
	`CREATE TABLE IF NOT EXISTS invites (
		id          TEXT PRIMARY KEY,
		session_id  TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
		from_user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		to_user_id  TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		status      TEXT NOT NULL DEFAULT 'pending',
		created_at  TEXT NOT NULL,
		UNIQUE (session_id, to_user_id)
	)`,
	`ALTER TABLE sessions ADD COLUMN ended_early INTEGER NOT NULL DEFAULT 0`,
	`ALTER TABLE sessions ADD COLUMN court_duration_minutes INTEGER`,
	`ALTER TABLE sessions ADD COLUMN ends_at TEXT`,
}

const schema = `
CREATE TABLE IF NOT EXISTS users (
	id            TEXT PRIMARY KEY,
	email         TEXT NOT NULL UNIQUE,
	display_name  TEXT NOT NULL,
	password_hash TEXT NOT NULL,
	created_at    TEXT NOT NULL
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
	id                TEXT PRIMARY KEY,
	admin_token       TEXT NOT NULL,
	status            TEXT NOT NULL DEFAULT 'lobby',
	courts            INTEGER NOT NULL,
	points            INTEGER NOT NULL,
	rounds_total      INTEGER,
	creator_player_id TEXT,
	created_at        TEXT NOT NULL,
	updated_at        TEXT NOT NULL
);


CREATE TABLE IF NOT EXISTS players (
	id         TEXT PRIMARY KEY,
	session_id TEXT NOT NULL REFERENCES sessions(id),
	name       TEXT NOT NULL,
	active     INTEGER NOT NULL DEFAULT 1,
	joined_at  TEXT NOT NULL
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
	id      TEXT PRIMARY KEY,
	round_id TEXT NOT NULL REFERENCES rounds(id),
	court   INTEGER NOT NULL,
	p1      TEXT NOT NULL REFERENCES players(id),
	p2      TEXT NOT NULL REFERENCES players(id),
	p3      TEXT NOT NULL REFERENCES players(id),
	p4      TEXT NOT NULL REFERENCES players(id),
	score_a INTEGER,
	score_b INTEGER
);

CREATE INDEX IF NOT EXISTS idx_matches_round ON matches(round_id);
`
