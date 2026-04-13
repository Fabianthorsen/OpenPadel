package main

import (
	"database/sql"
	"flag"
	"log"
	"os"

	"github.com/pressly/goose/v3"
	"github.com/fabianthorsen/openpadel/internal/store"
	_ "modernc.org/sqlite"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: migrate <command> [args]\ncommands: up, down, status, reset")
	}

	command := os.Args[1]
	dbPath := flag.String("db", "openpadel.db", "database path")
	flag.Parse()

	db, err := sql.Open("sqlite", *dbPath+"?_journal_mode=WAL&_foreign_keys=on")
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(1)

	goose.SetBaseFS(store.MigrationsFS)
	if err := goose.SetDialect("sqlite3"); err != nil {
		log.Fatalf("set dialect: %v", err)
	}

	switch command {
	case "up":
		if err := goose.Up(db, "migrations"); err != nil {
			log.Fatalf("migrate up: %v", err)
		}
		log.Println("migrations applied")
	case "down":
		if err := goose.Down(db, "migrations"); err != nil {
			log.Fatalf("migrate down: %v", err)
		}
		log.Println("migration rolled back")
	case "status":
		if err := goose.Status(db, "migrations"); err != nil {
			log.Fatalf("status: %v", err)
		}
	case "reset":
		if err := goose.Reset(db, "migrations"); err != nil {
			log.Fatalf("reset: %v", err)
		}
		log.Println("all migrations rolled back")
	default:
		log.Fatalf("unknown command: %s", command)
	}
}
