.PHONY: dev build fmt lint test

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

## Build production binary (embeds web/build)
build:
	cd web && bun run build
	go build -o $(BIN) ./cmd/server

## Run Go tests
test:
	go test ./...

## Tidy Go deps
tidy:
	go mod tidy
