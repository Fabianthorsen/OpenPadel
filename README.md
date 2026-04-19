# OpenPadel

A lightweight padel tournament app. Supports Americano (rotating partners, point-based) and Mexicano (pairings adapt by standings).

## Features

- **Americano** — round-robin rotation, configurable points per game, live leaderboard
- **Mexicano** — dynamic pairings based on standings, no bench players
- **Timed Americano** — fixed duration tournaments with drift-corrected round timings
- Join by 4-character session code, no account required
- Real-time score updates via polling
- Push notifications when a tournament starts (PWA)
- User accounts with career stats split by game mode

## Stack

- **Backend** — Go, SQLite (via `mattn/go-sqlite3`), deployed on [Fly.io](https://fly.io) (Stockholm region)
- **Frontend** — SvelteKit 5, Tailwind CSS, Bun
- **Auth** — JWT, bcrypt password hashing

## Project structure

```
.
├── cmd/server/          # Entrypoint
├── internal/
│   ├── api/             # HTTP handlers and router
│   ├── domain/          # Shared types
│   ├── gamemode/        # Game mode services (americano, mexicano, timed)
│   ├── store/           # SQLite queries
│   └── ui/              # Embedded SvelteKit build
└── web/                 # SvelteKit frontend
    └── src/
        ├── lib/
        │   ├── api/     # API client
        │   ├── components/
        │   └── i18n/    # English + Norwegian
        └── routes/
```

## Running locally

```bash
# Terminal 1 — backend (serves API on :8080)
make dev/api

# Terminal 2 — frontend (proxies API, serves on :5173)
make dev/web
```

## Building for production

```bash
make build   # builds frontend then compiles Go binary to bin/openpadel
```

## Deploying

```bash
fly deploy
```

The database is persisted on a Fly volume at `/data/openpadel.db`.

## Useful commands

```bash
make db/reset   # clear all game data (keeps user accounts)
make test       # run Go tests
make fmt        # format Go + frontend
```
