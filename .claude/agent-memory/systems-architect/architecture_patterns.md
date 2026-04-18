---
name: Architecture patterns
description: Key conventions for adding new game modes — sqlc queries, scheduler files, API handler branching, SSE event types, domain types, frontend routing
type: project
---

New game modes follow a consistent pattern:
- **Scheduler**: separate file in `internal/scheduler/` (e.g., `americano.go`, `mexicano.go`) with pure functions
- **API branching**: `startSession` in `internal/api/sessions.go` switches on `sess.GameMode`; mode-specific helpers in dedicated files (e.g., `mexicano.go`)
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
