package store

import (
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"
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
	goose.SetBaseFS(MigrationsFS)
	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("goose dialect: %w", err)
	}

	// Bootstrap existing databases: if the users table exists but goose's
	// goose_db_version table does not, force version to 1 so goose skips
	// the initial migration (schema is already applied).
	if err := bootstrapIfNeeded(s.db); err != nil {
		return err
	}

	return goose.Up(s.db, "migrations")
}

func bootstrapIfNeeded(db *sql.DB) error {
	// Check if goose version table already exists.
	var versionTableExists int
	err := db.QueryRow(
		`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='goose_db_version'`,
	).Scan(&versionTableExists)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if versionTableExists > 0 {
		return nil // goose already managing this DB
	}

	// Check if users table exists (indicates a pre-goose production DB).
	var usersExists int
	err = db.QueryRow(
		`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='users'`,
	).Scan(&usersExists)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if usersExists == 0 {
		return nil // fresh DB — let goose apply everything from scratch
	}

	// Existing DB, no goose table: create version table and mark version 1 as applied.
	createTableSQL := `
	CREATE TABLE goose_db_version (
		id       INTEGER PRIMARY KEY,
		version_id BIGINT NOT NULL,
		is_applied BOOLEAN NOT NULL,
		tstamp   TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`
	if _, err := db.Exec(createTableSQL); err != nil {
		return fmt.Errorf("create goose_db_version table: %w", err)
	}

	insertSQL := `INSERT INTO goose_db_version (version_id, is_applied) VALUES (1, 1)`
	if _, err := db.Exec(insertSQL); err != nil {
		return fmt.Errorf("bootstrap goose version: %w", err)
	}

	return nil
}
