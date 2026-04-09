# Architecture — OpenPadel

## Overview

Single Go binary serves both the REST API and the compiled SvelteKit static build.
No separate web server needed. SQLite on disk. Deployed as a single Fly.io machine.

```
┌─────────────────────────────────────┐
│              Fly.io VM              │
│                                     │
│  Go binary                          │
│  ├── /api/*        → REST handlers  │
│  └── /*            → SvelteKit PWA  │
│                                     │
│  /data/openpadel.db   SQLite file   │
└─────────────────────────────────────┘
```

---

## Project Structure

```
openpadel/
├── cmd/server/main.go          entrypoint — wires store, email, router; reads env
├── internal/
│   ├── api/
│   │   ├── router.go           chi router, CORS, middleware, all route registrations
│   │   ├── middleware.go       requireAuth / optionalAuth, context helpers
│   │   ├── respond.go          respond() and respondError() helpers
│   │   ├── auth.go             register, login, logout, me, profile, history, deleteAccount, forgot/reset
│   │   ├── sessions.go         create, get, start, cancel, close session
│   │   ├── players.go          join, deactivate player
│   │   ├── rounds.go           get rounds, current round, advance, submit score, live score, leaderboard
│   │   ├── tennis.go           set teams, get match, add point, set server
│   │   ├── mexicano.go         Mexicano-specific handlers
│   │   ├── contacts.go         get, add, remove contacts; search users
│   │   ├── invites.go          get, send, accept, decline invites
│   │   └── push.go             VAPID key, subscribe, unsubscribe
│   ├── domain/session.go       all shared types (User, Session, Player, Round, Match, etc.)
│   ├── store/                  SQLite data access — one file per domain area
│   │   ├── store.go            DB init, WAL mode, inline schema + additive migrations
│   │   ├── id.go               newID() (4-char base32), newAdminToken() (tok_ + base58)
│   │   ├── sessions.go
│   │   ├── players.go
│   │   ├── rounds.go
│   │   ├── users.go
│   │   ├── tennis.go
│   │   ├── contacts.go
│   │   ├── invites.go
│   │   └── push.go
│   ├── scheduler/
│   │   ├── americano.go        greedy round-generation, full bench rotation pre-computed
│   │   └── mexicano.go         Mexicano variant — pairings adapt based on standings
│   ├── tennis/scoring.go       pure tennis scoring engine (sets, games, tiebreak, golden point)
│   ├── livescores/store.go     in-memory concurrent map for live/in-progress scores
│   ├── email/resend.go         Resend API client — password reset only
│   └── ui/ui.go                embed.FS wrapper — serves SPA, injects OG meta tags
├── web/                        SvelteKit frontend source
├── Dockerfile                  two-stage build: bun → go binary with embedded frontend
└── fly.toml                    Fly.io config, SQLite volume at /data
```

---

## Auth

Email/password authentication. Opaque bearer tokens stored in DB.

```
POST /api/auth/register   → creates user + issues token
POST /api/auth/login      → verifies bcrypt hash, issues token
Header: Authorization: Bearer <token>
```

Session admin tokens are separate (`tok_` + 32 base58 chars), issued at session creation
and stored in the browser's `localStorage`.

---

## Game Modes

| Mode       | Status | Description                                              |
|------------|--------|----------------------------------------------------------|
| Americano  | Live   | Rotating partners, individual scoring, pre-computed rounds |
| Mexicano   | Live   | Like Americano, but pairings adapt each round by standings |
| Tennis     | Live   | Regular 2v2 with sets, games, serve tracking             |
| Round Robin| Planned| Every pair plays every other pair                        |

---

## API

Base path: `/api`. Content-Type: `application/json` throughout.
Errors: `{ "error": "human readable message" }`

### Auth
```
POST   /api/auth/register
POST   /api/auth/login
POST   /api/auth/logout
GET    /api/auth/me
PUT    /api/auth/profile
GET    /api/auth/history
DELETE /api/auth/account
POST   /api/auth/forgot-password
POST   /api/auth/reset-password
```

### Sessions
```
POST   /api/sessions
GET    /api/sessions/:id
POST   /api/sessions/:id/start
POST   /api/sessions/:id/cancel
POST   /api/sessions/:id/close
```

### Players
```
POST   /api/sessions/:id/players
DELETE /api/sessions/:id/players/:player_id
```

### Rounds & Scores
```
GET    /api/sessions/:id/rounds
GET    /api/sessions/:id/rounds/current
POST   /api/sessions/:id/rounds/advance
PUT    /api/sessions/:id/matches/:match_id/score
PUT    /api/sessions/:id/matches/:match_id/live-score
GET    /api/sessions/:id/leaderboard
```

### Tennis
```
POST   /api/sessions/:id/tennis/teams
GET    /api/sessions/:id/tennis
POST   /api/sessions/:id/tennis/point
PUT    /api/sessions/:id/tennis/server
```

### Contacts & Invites
```
GET    /api/contacts
POST   /api/contacts
DELETE /api/contacts/:contact_id
GET    /api/users/search

GET    /api/invites
POST   /api/sessions/:id/invites
PUT    /api/invites/:invite_id/accept
PUT    /api/invites/:invite_id/decline
GET    /api/sessions/:id/invites
```

### Push Notifications
```
GET    /api/push/vapid-public-key
POST   /api/push/subscribe
DELETE /api/push/unsubscribe
```

---

## Database

Schema is defined inline in `internal/store/store.go`. Migrations are additive
`ALTER TABLE` statements in a `var migrations []string` slice — "duplicate column"
errors are silently swallowed. No migration framework.

### Core tables

```sql
sessions  (id, admin_token, status, game_mode, name, courts, points,
           rounds_total, created_at, ended_at, ended_early)

players   (id, session_id, user_id, name, active, joined_at)

users     (id, email, display_name, password_hash, created_at)

auth_tokens          (token, user_id, created_at)
password_reset_tokens(token, user_id, expires_at, used)

rounds  (id, session_id, number)
bench   (round_id, player_id)
matches (id, round_id, court, p1, p2, p3, p4, score_a, score_b)

tennis_matches (id, session_id, ...)

contacts (user_id, contact_id, created_at)
invites  (id, session_id, inviter_id, invitee_id, status, created_at)

push_subscriptions (user_id, endpoint, p256dh, auth, created_at)
```

All times stored as `TEXT` in RFC3339. IDs are 4-char base32 (sessions) or
UUID-style (users/players). `crypto/rand` throughout.

---

## Scheduler

### Americano (`internal/scheduler/americano.go`)
Greedy round-by-round with a scoring function. All rounds pre-computed at session start.

Constraints (priority order):
1. **No consecutive bench** — benched in round N → must play round N+1
2. **Balanced bench** — bench slots distributed evenly
3. **Partner variety** — penalise recent partner repeats
4. **Opponent variety** — penalise recent opponent repeats

### Mexicano (`internal/scheduler/mexicano.go`)
Same constraints, but pairings are recalculated each round based on current standings.
No bench — requires exactly `courts × 4` players.

---

## Frontend

SvelteKit 5 SPA, compiled to `web/build/` and embedded into the Go binary via
`//go:embed all:build` in `internal/ui/`. Served at `/*` — the Go binary is the
only process needed.

Key patterns:
- **Svelte 5 runes** enforced globally (`$state`, `$props`, `$derived`, `$effect`)
- **API client** — single typed `api` object in `$lib/api/client.ts`
- **Auth store** — runes-based, token in `localStorage` under `auth_token`
- **Types** — all shared interfaces in `src/app.d.ts` under the `App` namespace
- **i18n** — English + Norwegian via `svelte-i18n`
- **Polling** — 3s in lobby, 15s during active play

---

## Deployment

Two-stage Docker build: Bun builds the frontend, Go embeds it and compiles the binary.
Final image is a single static binary on Alpine (~20 MB).

```
fly deploy         # builds and deploys to Fly.io (Stockholm / arn region)
```

SQLite lives on a persistent Fly volume mounted at `/data/openpadel.db`.
WAL mode enabled, single connection (`SetMaxOpenConns(1)`).
Litestream replicates the database continuously to Tigris (S3-compatible) via the
`litestream replicate` sidecar. On container start, the DB is restored from the
replica if no local file exists.
