---
name: Architecture patterns
description: Key conventions for adding new game modes — sqlc queries, scheduler files, API handler branching, SSE event types, domain types, frontend routing
type: project
---

New game modes follow a consistent pattern:
- **Game mode service layer**: `internal/gamemode/{mode}/service.go` orchestrates Start() and AdvanceRound() per mode (replaces old `internal/scheduler/` pattern)
- **Round generation**: `internal/gamemode/{mode}/rounds.go` contains mode-specific pairing logic and constraints
- **API branching**: `startSession` and `advanceRound` handlers in `internal/api/` dispatch to mode-specific services via interface injection
- **Store**: sqlc queries in `internal/store/queries/*.sql`, generated code in `internal/store/db/`, hand-written store methods in `internal/store/*.go`
- **Domain types**: all in `internal/domain/session.go` — Session, Round, Match, Score, Standing, Leaderboard
- **Frontend types**: mirrored in `web/src/app.d.ts` under `App` namespace
- **SSE events**: defined in `internal/events/envelope.go`, emitted post-mutation in handlers
- **Session table**: `game_mode` TEXT column discriminates modes; mode-specific columns are nullable
- **Scoring**: `submitScore` validates `score_a + score_b == session.points`; auto-complete checks differ by mode
- **Round advance**: Americano pre-generates all rounds; Mexicano generates one round at a time based on standings
- **CreateDrawer.svelte**: single drawer handles all mode creation with conditional UI sections

**Why:** Understanding these patterns is essential for planning any new game mode to ensure consistency.
**How to apply:** When planning Timed Americano or any new mode, follow these exact branching points.
