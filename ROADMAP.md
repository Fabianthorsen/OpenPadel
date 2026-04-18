# Roadmap

## Planned

### Features

- [ ] Score screen UX redesign — court tabs, unified card, numpad entry, courts overview bottom sheet
- [ ] Invite screen UX redesign — host avatar hero, stacked player avatars, join section
- [ ] Round count control — allow admin to choose rounds as multiples of the rotation unit (e.g. 5, 10, 15 for a 5-player bench config)
- [ ] Admin can add extra rounds mid-session if players want to keep playing
- [ ] End tournament button — admin menu with options: Keep playing, End and discard, End and save results
- [ ] Timed Americano — game mode with time-based rounds instead of fixed rounds
- [ ] Winners court — winners stay on court, losers rotate (losers bench or next challengers)
- [ ] Assign score entry to other players (not admin-only)

### Design System

- [ ] **Centralise podium / semantic colors** — replace hardcoded `#3d7a24`, `#4a7856`, `#a8c5b0`, `#c0392b` in [ActiveSession.svelte](web/src/lib/components/ActiveSession.svelte), [Leaderboard.svelte](web/src/lib/components/Leaderboard.svelte), and [profile](web/src/routes/profile/+page.svelte#L490) with new tokens (`--color-primary-strong`, `--color-podium-silver`, `--color-podium-bronze`, `--color-loss`) in [app.css](web/src/app.css)
- [ ] **Move Avatar palette into theme tokens** — [ui/Avatar.svelte](web/src/lib/components/ui/Avatar.svelte) hardcodes 8 avatar hex colors; lift to CSS vars so they survive a dark-mode pass
- [ ] **Typography utility classes** — add `.text-display`, `.text-h1`, `.text-h2`, `.text-h3`, `.text-small` utilities to [app.css](web/src/app.css) matching the spec in [DESIGN.md](DESIGN.md); replace scattered `text-[28px]`/`text-[34px]`/`text-[80px]`/`font-[800]` across components
- [ ] **Radius utility aliases** — add `rounded-card` (8px), `rounded-input` (6px), `rounded-badge` (4px), `rounded-pill` (99px) to theme; stop mixing `rounded-2xl`/`rounded-3xl`/`rounded-xl` inline
- [ ] **Unify card surfaces** — [Leaderboard](web/src/lib/components/Leaderboard.svelte), [ActiveSession](web/src/lib/components/ActiveSession.svelte), [Lobby](web/src/lib/components/Lobby.svelte) use raw `<div class="rounded-2xl bg-surface-raised">` — standardise on the shadcn [Card](web/src/lib/components/ui/card) component
- [ ] **Consolidate toggle groups** — [pill-toggle-group](web/src/lib/components/ui/pill-toggle-group) and [toggle-group](web/src/lib/components/ui/toggle-group) diverged; share one primitive with a `variant="pill" | "square"` prop
- [ ] **Spacing audit** — sweep `gap-2.5`/`px-3.5`/off-scale paddings across [ActiveSession](web/src/lib/components/ActiveSession.svelte), [Lobby](web/src/lib/components/Lobby.svelte), [Leaderboard](web/src/lib/components/Leaderboard.svelte), [+page.svelte](web/src/routes/+page.svelte), [auth](web/src/routes/auth/+page.svelte), [forgot](web/src/routes/forgot/+page.svelte), [reset](web/src/routes/reset/+page.svelte); enforce 4/8/12/16/20/24/32/48 scale from [DESIGN.md](DESIGN.md)
- [ ] **Dark-mode foundation** — swap the `@theme` hex values in [app.css](web/src/app.css) for CSS custom properties under `:root` + `prefers-color-scheme: dark`; prerequisite for the V2 dark mode called out in [DESIGN.md](DESIGN.md)
- [ ] **Accessibility sweep** — add `aria-label` to emoji-only buttons (🎾 court tabs in ActiveSession); verify 48×48 tap targets; add non-color cue (icon or letter) to podium rank backgrounds so colorblind users can distinguish silver/bronze

### Backend Quality

- [ ] **Finish the sqlc migration** — residual raw SQL in [rounds.go:89](internal/store/rounds.go#L89), [rounds.go:97](internal/store/rounds.go#L97), [rounds.go:229](internal/store/rounds.go#L229), [rounds.go:328](internal/store/rounds.go#L328), [rounds.go:348](internal/store/rounds.go#L348), [contacts.go:44](internal/store/contacts.go#L44), [players.go:86](internal/store/players.go#L86), [invites.go:32-38](internal/store/invites.go#L32-L38) bypasses codegen — move into `.sql` query files
- [ ] **Session-lookup helper** — 7-line `Get → ErrNotFound → server_error` blocks repeat across [api/sessions.go](internal/api/sessions.go), [rounds.go](internal/api/rounds.go), [tennis.go](internal/api/tennis.go), [players.go](internal/api/players.go); extract `requireSession(w, r, id) *domain.Session` into [middleware.go](internal/api/middleware.go)
- [ ] **Structured error logging in handlers** — only [auth.go](internal/api/auth.go) logs store errors today; every other handler swallows context on `respondError(w, 500, "server_error")`. Tag with handler name + request id
- [ ] **Consistent sentinel handling** — some handlers use `errors.Is(err, store.ErrNotFound)`, others check error-message strings; make `errors.Is` the convention
- [ ] **SSE drop metering** — [events/hub.go](internal/events/hub.go) silently drops when a client buffer fills; log at debug and expose a counter so we can tell if it ever matters in prod

### Tooling & Infrastructure

- [ ] **Error toasts** — wire up svelte-sonner (already installed) to API client for global error feedback
- [ ] **Sentry** — add `@sentry/sveltekit` + Go SDK for production error tracking with stack traces
- [x] **Vitest** — unit tests for frontend: API client, session stream store, utility functions (35 tests across `utils.ts`, `api/client.ts`, `sessionStream.svelte.ts`); component tests deferred
- [x] **API handler tests** — `httptest`-based coverage for auth, session lifecycle, round advance, score submission, player join/deactivate (store gap tests for users, rounds, players added too); push/mexicano handler tests deferred
- [ ] **Playwright** — E2E tests for happy path (create session → join → submit scores); requires `data-testid` attributes on key interactive elements
- [ ] **sqlc-only CI check** — grep for `s.db.Exec(`/`s.db.Query(`/`s.db.QueryRow(` in `internal/store/` to prevent new raw SQL sneaking back in
- [x] **v1.9.0** — Database migrations with goose: versioned `.sql` files in `internal/store/migrations/`, embedded via `//go:embed`
- [x] **v1.9.0** — **sqlc** — generate type-safe Go from SQL queries, eliminate hand-written `rows.Scan` patterns in `internal/store/`; refactored all store files (users, sessions, players, rounds, tennis, contacts, invites, push)

## In Progress

_Nothing in progress — main is clean._

## Done

- [x] **v1.10.0** — SSE real-time updates: replaced polling with Server-Sent Events (`internal/events` Hub + handler, `sessionStream.svelte.ts` factory store). Live scores, round advances, session state changes, and tennis points now push instantly to all connected clients. 30 s fallback poll retained.
- [x] **v1.10.0** — Admin access recovery: sessions now store `creator_user_id`; logged-in session creator is recognised as admin even after localStorage is cleared or on a different device. Profile upcoming-session links restore the admin token automatically.

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
- [x] **v1.4.0** — Adaptive polling (3s lobby / 15s active) — superseded by SSE in v1.9.5
- [x] **v1.4.0** — Tournament naming and fun awards on final results
- [x] **v1.4.0** — Admin joins as player with creator crown
- [x] **v1.4.0** — Explicit round advance, score editing, draws support
- [x] **v1.3.x** — Americano game mode (V1 core — sessions, lobby, rounds, leaderboard)
- [x] **v1.2.x** — i18n — English and Norwegian translations
- [x] **v1.1.x** — PWA — installable, offline-capable, OG share tags
- [x] **v1.0.0** — Initial release — Go backend + SvelteKit frontend on Fly.io
