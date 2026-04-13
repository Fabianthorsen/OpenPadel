# Roadmap

## Planned

- [ ] Score screen UX redesign — court tabs, unified card, numpad entry, courts overview bottom sheet
- [ ] Invite screen UX redesign — host avatar hero, stacked player avatars, join section
- [ ] Round count control — allow admin to choose rounds as multiples of the rotation unit (e.g. 5, 10, 15 for a 5-player bench config)
- [ ] Admin can add extra rounds mid-session if players want to keep playing
- [ ] Round Robin game mode — every pair plays every other pair
- [ ] Assign score entry to other players (not admin-only)
### Tooling & Infrastructure

- [ ] **Error toasts** — wire up svelte-sonner (already installed) to API client for global error feedback
- [ ] **Sentry** — add `@sentry/sveltekit` + Go SDK for production error tracking with stack traces
- [ ] **API handler tests** — scheduler is well-tested; add coverage for critical API handlers (start session, submit score, advance round)
- [x] **v1.9.0** — Database migrations with goose: versioned `.sql` files in `internal/store/migrations/`, embedded via `//go:embed`
- [x] **v1.9.0** — **sqlc** — generate type-safe Go from SQL queries, eliminate hand-written `rows.Scan` patterns in `internal/store/`; refactored users.go, sessions.go, players.go
- [ ] **Playwright** — E2E tests for happy path (create session → join → submit scores)

## In Progress

_Nothing in progress — main is clean._

## Done

- [x] **v1.8.0** — Pull-to-refresh on home, session, and profile screens
- [x] **v1.8.0** — Structured logging: `log/slog` JSON handler, request logger middleware (mutations + errors only)
- [x] **v1.7.1** — Player avatar system: lucide icon picker on profile, avatar shown in lobby, game, and leaderboard; guests get slate Bot icon
- [x] **v1.7.0** — Litestream continuous replication to Tigris (S3)
- [x] **v1.6.0** — Mexicano game mode (backend, scheduler, UI)
- [x] **v1.6.0** — Court booking timer with rounds-or-time duration picker
- [x] **v1.6.0** — Randomise player order and Team A/B sides on tournament start
- [x] **v1.5.0** — Early end flag (`ended_early`) on tournaments
- [x] **v1.4.2** — Leaderboard tiebreaker chain
- [x] **v1.4.0** — Tennis game mode (2v2, sets, serve tracking)
- [x] **v1.4.0** — User accounts, email/password auth, password reset
- [x] **v1.4.0** — Career stats and profile page (split by game mode)
- [x] **v1.4.0** — Contacts system with search and profile UI
- [x] **v1.4.0** — Invite system — contacts must accept before joining
- [x] **v1.4.0** — Web push notifications for tournament start
- [x] **v1.4.0** — Live score sync (in-memory store)
- [x] **v1.4.0** — 4-char uppercase join codes with home page entry
- [x] **v1.4.0** — Adaptive polling (3s lobby / 15s active)
- [x] **v1.4.0** — Tournament naming and fun awards on final results
- [x] **v1.4.0** — Admin joins as player with creator crown
- [x] **v1.4.0** — Explicit round advance, score editing, draws support
- [x] **v1.3.x** — Americano game mode (V1 core — sessions, lobby, rounds, leaderboard)
- [x] **v1.2.x** — i18n — English and Norwegian translations
- [x] **v1.1.x** — PWA — installable, offline-capable, OG share tags
- [x] **v1.0.0** — Initial release — Go backend + SvelteKit frontend on Fly.io
