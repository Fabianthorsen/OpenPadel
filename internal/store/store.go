package store

import (
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"
	"github.com/fabianthorsen/openpadel/internal/store/db"
	_ "modernc.org/sqlite"
)

type Store struct {
	db      *sql.DB
	queries *db.Queries
}

func Open(path string) (*Store, error) {
	dbHandle, err := sql.Open("sqlite", path+"?_journal_mode=WAL&_foreign_keys=on")
	if err != nil {
		return nil, err
	}
	dbHandle.SetMaxOpenConns(1)
	s := &Store{
		db:      dbHandle,
		queries: db.New(dbHandle),
	}
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
	return goose.Up(s.db, "migrations")
}
