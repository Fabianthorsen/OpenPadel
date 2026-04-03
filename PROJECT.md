# NotTennis — Padel Game Manager

## Overview

A mobile-first Progressive Web App (PWA) for organizing padel games and tracking scores.
Designed to be used courtside on a phone — fast, offline-capable, and simple.
Initial target: personal use, friends and family. Hosted on Fly.io.

## Stack

| Layer      | Technology          | Notes                                            |
|------------|---------------------|--------------------------------------------------|
| Backend    | Go                  | REST API, single binary                          |
| Frontend   | SvelteKit           | PWA, mobile-first, small bundle                  |
| Styling    | Tailwind CSS        | Utility-first, pairs perfectly with SvelteKit    |
| Components | shadcn-svelte       | Own your components, full design control         |
| Database   | SQLite + litestream | Single file DB, S3 backup via litestream         |
| Hosting    | Fly.io              | Go binary + SQLite, free tier, easy deploys      |

## AI Personas

### Dev — "Marco"

> Senior Go + SvelteKit engineer. Pragmatic. Prefers simple, boring solutions over clever ones.
> Cares deeply about API contracts, type safety, and not over-engineering V1.
> Will push back on scope creep and advocate for clean separation of concerns.

Responsibilities:
- Backend API design (Go, REST, SQLite)
- SvelteKit routing, data loading, PWA config
- Session ID generation, shareable link logic
- Deployment (single Go binary serving static SvelteKit build)

### Design — "Sofia"

> Mobile UX designer with a love for sports apps and real-world courtside usage.
> Obsessed with thumb-friendly layouts, high contrast in sunlight, and one-tap actions.
> Pushes for strong visual hierarchy and delightful micro-interactions.
> References: Strava, Duolingo, and premium sports scoring apps.

Responsibilities:
- Visual language: typography, color palette, spacing system
- Mobile layout: lobby, score entry screens, leaderboard, session setup
- Component decisions within shadcn-svelte
- PWA install experience and offline state UI

---

## V1 Scope — Americano

Americano is a rotating-partner format where players accumulate individual points across multiple rounds.
Matches are played to a fixed number of points (both team scores are recorded, e.g. 15-9).

### Player & bench model

```
Active players per round = courts × 4
Bench players per round  = total players − (courts × 4)
Default rounds           = number of players
```

Any number of players ≥ 5 is valid.

**V1 — rounds are fixed to player count, no override.**

**V2 — round count control:**
```
bench_size = players - (courts * 4)

if bench_size == 0:
  admin can set any number of rounds freely

if bench_size > 0:
  rotation_unit = players / bench_size   # rounds for everyone to bench once
  valid round counts = rotation_unit, rotation_unit*2, rotation_unit*3, ...
  (enforced in UI — only multiples of rotation_unit are selectable)
```

Examples:
- 5 players, 1 court → bench_size=1, rotation=5 → valid: 5, 10, 15...
- 6 players, 1 court → bench_size=2, rotation=3 → valid: 3, 6, 9...
- 10 players, 2 courts → bench_size=2, rotation=5 → valid: 5, 10, 15...

**Hard bench constraint:** a player who sat out round N must play in round N+1.
No exceptions — this guarantees fairness without needing a complex scoring system.

Examples:
- 1 court, 6 players → 2 benched per round
- 2 courts, 9 players → 1 benched per round
- 2 courts, 8 players → nobody benched (clean rotation)
- 3 courts, 14 players → 2 benched per round

Player dropout mid-session = mark as inactive → treated as a permanent bench slot,
remaining active players fill future rounds normally.

### Scoring options

| Points | Use case              |
|--------|-----------------------|
| 16     | Quick / warm-up games |
| 24     | Standard (default)    |
| 32     | Long format           |

### Session lifecycle

```
1. SETUP     Admin creates session, picks courts, rounds, points target
                └── Gets a join link (e.g. nottennis.app/s/abc123)

2. LOBBY     Admin shares link (WhatsApp etc.)
                └── Players open link, enter their name → join the session
                └── Admin sees players appearing in real time (polling)
                └── Admin hits "Start" when everyone is in

3. ACTIVE    Rounds generated upfront, bench rotation pre-calculated
                └── Each round shows: court assignments + who is on bench
                └── Admin enters scores after each match
                └── Players (and spectators) watch live leaderboard on same link

4. COMPLETE  All rounds done → final standings shown
                └── Shareable summary
```

### Core features

- **Session creation** — admin sets courts (1–4), rounds, points target
- **Join flow** — players join via shared link by entering their name
- **Auto-pairing with bench** — fair rotation, everyone plays roughly equally
- **Score entry** — admin-only in V1, quick input per match
- **Live leaderboard** — public, auto-refreshes every 15s, shows points + bench count
- **Final summary** — standings + match history when complete

### Out of scope for V1

- Player photos (V2)
- Assigning score entry to other players (V2)
- User accounts / authentication
- Persistent history across sessions
- WebSocket real-time push (polling is fine)
- Other game modes

---

## Session Data Model

```
Session
  ├── id             short URL-safe string, e.g. "abc123"
  ├── admin_token    secret query param, only creator has it
  ├── status         setup | lobby | active | complete
  ├── config
  │     ├── courts       int (1–4)
  │     ├── rounds       int
  │     └── points       int (16 | 24 | 32)
  ├── players []
  │     ├── id
  │     ├── name
  │     └── active       bool (false = dropped out)
  └── rounds []
        ├── bench        [player_id, ...]
        └── matches []
              ├── court        int
              ├── team_a       [player_id, player_id]
              ├── team_b       [player_id, player_id]
              └── score        { a: int, b: int } | null
```

Public URL: `/s/:id` — leaderboard + join, no auth
Admin URL: `/s/:id?token=xxx` — score entry + session controls

---

## Game Modes (Roadmap)

| Mode        | Description                                                  | Priority |
|-------------|--------------------------------------------------------------|----------|
| Americano   | Rotating partners, individual scoring, auto-scheduled rounds | V1       |
| Mexicano    | Like Americano but pairings adapt based on current standings | V2       |
| Team        | Fixed pairs compete through a bracket or round robin         | V2       |
| Round Robin | Every pair plays every other pair                            | V3       |

---

## PWA Requirements

- Installable on iOS and Android home screens
- Readable in direct sunlight (contrast-aware design)
- Works on flaky court WiFi — optimistic UI, syncs when online
- One-handed thumb-friendly layout throughout

---

## Open Questions

- [ ] Can the admin add extra rounds mid-session if everyone wants to keep playing?
- [ ] Minimum player count: 5 (1 court + 1 bench) — enforce hard or just warn?
