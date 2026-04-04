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
}

const schema = `
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

CREATE UNIQUE INDEX IF NOT EXISTS idx_players_session_name ON players(session_id, name);
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
