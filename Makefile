.PHONY: dev build fmt lint lint/go lint/web test setup db/reset db/migrate/status db/migrate/up db/migrate/down

# Go binary output
BIN := bin/openpadel

## Run backend in dev mode
dev/api:
	go run ./cmd/server

## Run frontend in dev mode
dev/web:
	cd web && bun run dev

## Format all code
fmt:
	gofmt -w .
	cd web && bun run format

## Check Go formatting, vet, and lint (new/changed code only vs main)
lint/go:
	@test -z "$$(gofmt -l .)" || (echo "gofmt needs to be run on:" && gofmt -l . && exit 1)
	go vet ./...
	golangci-lint run --new-from-rev=main ./...

## Check frontend formatting
lint/web:
	cd web && bun run lint

## Run all lint checks (no writes)
lint: lint/go lint/web

## Build production binary (embeds web/build into Go binary)
build:
	cd web && bun run build
	cp -r web/build internal/ui/build
	go build -o $(BIN) ./cmd/server

## Run Go tests
test:
	go test ./...

## Tidy Go deps
tidy:
	go mod tidy

## Clear all game data (sessions, rounds, matches, players) — keeps users & auth
db/reset:
	sqlite3 openpadel.db "DELETE FROM bench; DELETE FROM matches; DELETE FROM rounds; DELETE FROM players; DELETE FROM sessions;"
	@echo "Game data cleared."

## Show migration status
db/migrate/status:
	go run ./cmd/migrate status

## Run pending migrations
db/migrate/up:
	go run ./cmd/migrate up

## Rollback last migration
db/migrate/down:
	go run ./cmd/migrate down

## Install git hooks (run once after cloning; also runs automatically via `bun install`)
setup:
	scripts/install-hooks.sh
