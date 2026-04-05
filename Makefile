.PHONY: dev build fmt lint test setup db/reset

# Go binary output
BIN := bin/nottennis

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
	sqlite3 nottennis.db "DELETE FROM bench; DELETE FROM matches; DELETE FROM rounds; DELETE FROM players; DELETE FROM sessions;"
	@echo "Game data cleared."

## Install git hooks (run once after cloning)
setup:
	cp scripts/commit-msg .git/hooks/commit-msg
	chmod +x .git/hooks/commit-msg
	@echo "Git hooks installed."
