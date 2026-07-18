# CLAUDE.md

OpenPadel — Go + SQLite backend, SvelteKit 5 frontend, single binary deploy to Fly.io.

## Commands

- `make dev/api` / `make dev/web` — run backend/frontend in dev
- `make lint` — gofmt/vet/golangci-lint (new code only, see below) + prettier check
- `make test` — Go tests (`cd web && bun run test` for frontend)
- `make build` — build frontend, embed into Go binary
- `make fmt` — format Go + frontend
- `bun run check` (in `web/`) — svelte-check types

## Invariants

- **golangci-lint is scoped to `--new-from-rev=main`.** Pre-existing findings in old code are
  intentionally grandfathered — don't try to "fix" unrelated lint debt in an unrelated PR.
- **Mexicano requires ≥2 courts, no bench.** Player count must equal `courts × 4` exactly
  (`domain.MexicanoConstraints`). The UI disables the 1-court option for this mode
  (`Lobby.svelte`) even though the domain layer technically allows it at `courts=1`.
- **Guest and registered Players are both first-class.** Anyone can join a Session by name only,
  no account — this is a deliberate simplicity, not a stopgap. See **Player** in `CONTEXT.md`.
- **AdminToken is the only authorization gate**, checked via `isAdmin()`. `CreatorUserID` /
  `CreatorPlayerID` is identity metadata for UI display only — never treat it as an auth check.
- **Live updates are SSE-first**, with a 30s poll as fallback insurance for buffering proxies —
  don't "simplify" by ripping out the poll or by replacing SSE with polling.

## Docs

- `CONTEXT.md` — domain glossary (terms to use, terms to avoid)
- `ARCHITECTURE.md` — system design, API surface, data model
- `DESIGN.md` — visual/UX language
- `docs/adr/` — hard-to-reverse decisions and why
- `docs/specs/` — per-screen UI specs

## Agent skills

### Issue tracker

Issues live in GitHub Issues (`Fabianthorsen/OpenPadel`), managed via the `gh` CLI. See `docs/agents/issue-tracker.md`.

### Triage labels

Default canonical labels (`needs-triage`, `needs-info`, `ready-for-agent`, `ready-for-human`, `wontfix`). See `docs/agents/triage-labels.md`.

### Domain docs

Single-context — `CONTEXT.md` + `docs/adr/` at the repo root. See `docs/agents/domain.md`.
