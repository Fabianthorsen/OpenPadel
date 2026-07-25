package store

import (
	"database/sql"
	"fmt"

	"github.com/fabianthorsen/openpadel/internal/store/db"
	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

type Store struct {
	db      *sql.DB
	queries *db.Queries
}

func (s *Store) DB() *sql.DB {
	return s.db
}

func Open(path string) (*Store, error) {
	dbHandle, err := sql.Open("sqlite", path+"?_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)")
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
	if err := s.assertForeignKeys(); err != nil {
		return nil, err
	}
	return s, nil
}

// assertForeignKeys fails fast if the connection did not come up with foreign
// key enforcement on. The modernc.org/sqlite driver silently ignores the
// mattn-style `_foreign_keys=on` DSN param, so a wrong DSN degrades to no
// enforcement without any error — this turns that silent failure into a loud one.
func (s *Store) assertForeignKeys() error {
	var fk int
	if err := s.db.QueryRow("PRAGMA foreign_keys").Scan(&fk); err != nil {
		return fmt.Errorf("read foreign_keys pragma: %w", err)
	}
	if fk != 1 {
		return fmt.Errorf("foreign key enforcement is off (PRAGMA foreign_keys=%d); check the sqlite DSN", fk)
	}
	return nil
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
