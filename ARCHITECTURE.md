# Architecture — NotTennis
### By Marco

---

## Overview

Single Go binary serves both the REST API and the compiled SvelteKit static build.
No separate web server needed. SQLite on disk, backed up via litestream to S3.
Deployed as a single Fly.io machine.

```
┌─────────────────────────────────────┐
│              Fly.io VM              │
│                                     │
│  Go binary                          │
│  ├── /api/*        → REST handlers  │
│  └── /*            → SvelteKit PWA  │
│                                     │
│  nottennis.db      SQLite file      │
│  litestream        → S3 backup      │
└─────────────────────────────────────┘
```

---

## Project Structure

```
nottennis/
├── cmd/
│   └── server/
│       └── main.go           entry point, wires everything together
├── internal/
│   ├── api/
│   │   ├── sessions.go       session CRUD handlers
│   │   ├── players.go        join / deactivate handlers
│   │   ├── rounds.go         round + score handlers
│   │   ├── leaderboard.go    standings computation handler
│   │   └── middleware.go     admin auth, CORS, logging
│   ├── domain/
│   │   ├── session.go        Session, Player, Round, Match types
│   │   └── leaderboard.go    standings calculation logic
│   ├── scheduler/
│   │   └── americano.go      pairing + bench rotation algorithm
│   └── store/
│       ├── store.go          DB init, migrations
│       ├── sessions.go       session queries
│       ├── players.go        player queries
│       └── rounds.go         round + score queries
├── web/                      compiled SvelteKit output (gitignored, built at deploy)
├── fly.toml
├── Dockerfile
└── litestream.yml
```

---

## Admin Auth

The admin receives a token at session creation. It is stored by the SvelteKit client
in `localStorage` and sent as a header on write requests.

```
Header: Authorization: Bearer <admin_token>
```

The admin share URL (`/s/:id?token=xxx`) is used once — SvelteKit reads the query param,
stores it in localStorage, strips it from the URL. All subsequent API calls use the header.

Non-admin requests (leaderboard, join) require no auth.

---

## API

Base path: `/api`
Content-Type: `application/json` throughout.
Errors follow: `{ "error": "human readable message" }`

---

### Sessions

#### Create session
```
POST /api/sessions

Body:
{
  "courts": 2,
  "points": 24
}

Response 201:
{
  "id": "abc123",
  "admin_token": "tok_xxxxxxxxxxxxxxxx",
  "status": "lobby",
  "config": {
    "courts": 2,
    "points": 24
  },
  "players": [],
  "created_at": "2026-04-03T10:00:00Z"
}
```

#### Get session
```
GET /api/sessions/:id

Response 200:
{
  "id": "abc123",
  "status": "lobby" | "active" | "complete",
  "config": {
    "courts": 2,
    "points": 24,
    "rounds": 9        ← populated once session starts (= player count)
  },
  "players": [
    { "id": "p1", "name": "Ana",   "active": true },
    { "id": "p2", "name": "Bruno", "active": true }
  ],
  "current_round": 3,  ← null if lobby
  "updated_at": "2026-04-03T10:05:00Z"
}
```

This is the polling endpoint. SvelteKit calls it every 15s on the session page.

#### Start session
```
POST /api/sessions/:id/start
Authorization: Bearer <admin_token>

Response 200:
{
  "id": "abc123",
  "status": "active",
  "config": { "courts": 2, "points": 24, "rounds": 9 },
  "players": [ ... ]
}
```

Triggers round generation. All rounds pre-calculated and stored at this point.
Requires ≥ 5 active players.

---

### Players

#### Join session
```
POST /api/sessions/:id/players

Body:
{ "name": "Ana" }

Response 201:
{
  "id": "p1",
  "name": "Ana",
  "session_id": "abc123"
}

Errors:
  409 if session is not in lobby status
  409 if name already taken in this session — return "Ops, somebody already took that name"
```

#### Deactivate player (pre-start only)
```
DELETE /api/sessions/:id/players/:player_id
Authorization: Bearer <admin_token>

Response 200:
{ "id": "p1", "active": false }
```

Only valid while session is in lobby. No dropout handling mid-session in V1.

---

### Rounds

#### Get all rounds
```
GET /api/sessions/:id/rounds

Response 200:
{
  "rounds": [
    {
      "number": 1,
      "bench": ["p9"],
      "matches": [
        {
          "id": "m1",
          "court": 1,
          "team_a": ["p1", "p2"],
          "team_b": ["p3", "p4"],
          "score": null
        },
        {
          "id": "m2",
          "court": 2,
          "team_a": ["p5", "p6"],
          "team_b": ["p7", "p8"],
          "score": null
        }
      ]
    },
    ...
  ]
}
```

#### Get current round
```
GET /api/sessions/:id/rounds/current

Response 200: single round object (same shape as above)
Response 404: if session not yet active
```

---

### Scores

#### Submit score
```
PUT /api/sessions/:id/matches/:match_id/score
Authorization: Bearer <admin_token>

Body:
{
  "score_a": 15,
  "score_b": 9
}

Response 200:
{
  "id": "m1",
  "court": 1,
  "team_a": ["p1", "p2"],
  "team_b": ["p3", "p4"],
  "score": { "a": 15, "b": 9 }
}

Errors:
  400 if score_a + score_b != session points target
  403 if not admin
```

Validation: scores must sum to the session points target. Enforced both client and server.
Score can be re-submitted at any time to correct a mistake — overwrites the previous entry.
Leaderboard recomputes on the next read automatically (computed from raw scores, never cached).

---

### Leaderboard

#### Get standings
```
GET /api/sessions/:id/leaderboard

Response 200:
{
  "session_id": "abc123",
  "status": "active",
  "current_round": 3,
  "total_rounds": 9,
  "standings": [
    { "rank": 1, "player_id": "p1", "name": "Ana",   "points": 38 },
    { "rank": 2, "player_id": "p2", "name": "Bruno", "points": 35 },
    ...
  ],
  "updated_at": "2026-04-03T10:12:00Z"
}
```

Standings are computed on read (simple SQL sum), not stored.
Cheap enough for this scale — no cache needed in V1.

---

## Scheduler — Americano Algorithm

Lives in `internal/scheduler/americano.go`. Called once at session start.
Returns all rounds fully pre-computed.

### Constraints (in priority order)

1. **No consecutive bench** — if a player sat out round N, they must play round N+1. Hard constraint.
2. **Balanced bench** — bench slots distributed as evenly as possible across all players.
3. **Partner variety** — avoid same-partner pairing in back-to-back rounds where possible.
4. **Opponent variety** — avoid same opponents in back-to-back rounds where possible.

### Approach

Greedy round-by-round with a scoring function:

```
For each round:
  1. Forced players = those who were benched last round → must play
  2. Fill remaining active slots from remaining players
  3. Score candidate pairings penalising recent partner/opponent repeats
  4. Pick lowest-penalty assignment
  5. Remaining players sit bench
```

No backtracking needed for V1 — greedy is good enough for groups of 5–20.

---

## Database Schema

```sql
CREATE TABLE sessions (
  id           TEXT PRIMARY KEY,
  admin_token  TEXT NOT NULL,
  status       TEXT NOT NULL DEFAULT 'lobby',  -- lobby | active | complete
  courts       INTEGER NOT NULL,
  points       INTEGER NOT NULL,
  rounds_total INTEGER,                         -- set on start
  created_at   TEXT NOT NULL
);

CREATE TABLE players (
  id         TEXT PRIMARY KEY,
  session_id TEXT NOT NULL REFERENCES sessions(id),
  name       TEXT NOT NULL,
  active     INTEGER NOT NULL DEFAULT 1,        -- bool
  joined_at  TEXT NOT NULL
);

CREATE TABLE rounds (
  id         TEXT PRIMARY KEY,
  session_id TEXT NOT NULL REFERENCES sessions(id),
  number     INTEGER NOT NULL
);

CREATE TABLE bench (
  round_id  TEXT NOT NULL REFERENCES rounds(id),
  player_id TEXT NOT NULL REFERENCES players(id),
  PRIMARY KEY (round_id, player_id)
);

CREATE TABLE matches (
  id        TEXT PRIMARY KEY,
  round_id  TEXT NOT NULL REFERENCES rounds(id),
  court     INTEGER NOT NULL,
  p1        TEXT NOT NULL REFERENCES players(id),  -- team_a[0]
  p2        TEXT NOT NULL REFERENCES players(id),  -- team_a[1]
  p3        TEXT NOT NULL REFERENCES players(id),  -- team_b[0]
  p4        TEXT NOT NULL REFERENCES players(id),  -- team_b[1]
  score_a   INTEGER,                               -- null = not yet played
  score_b   INTEGER
);

CREATE UNIQUE INDEX idx_players_session_name ON players(session_id, name);
CREATE INDEX idx_matches_round ON matches(round_id);
CREATE INDEX idx_rounds_session ON rounds(session_id);
```

---

## Session ID Generation

Short, URL-safe, unguessable. 6 character base58 string (~38 billion combinations).
Not sequential — no enumeration risk.

```go
// Example output: "abc123", "xK9mPq"
func generateSessionID() string // crypto/rand + base58 alphabet
func generateAdminToken() string // "tok_" + 32 random base58 chars
```

---

## Deployment

```dockerfile
# Two-stage build
# Stage 1: build SvelteKit → web/
# Stage 2: build Go binary, embed web/ via embed.FS
# Final image: single binary, ~20MB
```

```yaml
# fly.toml
[mounts]
  source = "nottennis_data"
  destination = "/data"   # SQLite file lives here
```

litestream runs as a sidecar process, replicating `/data/nottennis.db` to S3 every 1s.

---

## V2 Notes (out of scope now, captured for later)

- User accounts: `users` table, bcrypt passwords or magic link auth
- Historical scores: `user_id` FK on players, aggregate views per user
- Win rate, stats: computed from matches table — no schema change needed
- Round count control: enforce `rounds % rotation_unit == 0` in session creation
- Assign score entry: role column on players table
- Mexicano: new scheduler variant, same round/match schema
