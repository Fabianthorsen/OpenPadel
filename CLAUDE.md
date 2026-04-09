# OpenPadel — Working Guidelines

## Git Workflow

1. `git pull` before creating any new branch
2. Branch naming mirrors commit type:
   - `feat/short-description`
   - `fix/short-description`
   - `chore/short-description`
   - `refactor/short-description`
3. One feature or fix per branch — keep scope small and focused
4. Conventional Commits required: `feat:`, `fix:`, `chore:`, `refactor:`, `docs:`
5. Scope tags encouraged: `feat(ui):`, `fix(scheduler):`, etc.
6. Never push directly to `main` — always via a PR from a feature branch
7. Merge to `main` only when the feature is complete and tested locally

## Testing

- Write tests for every store function, API handler, or business logic change where it makes sense
- Use the existing test patterns in `internal/store/*_test.go` and `internal/api/*_test.go`
- Tests must pass before opening a PR (`go test ./...`)

## Tooling

- Package manager: **bun** (not npm or npx) — always use `bun run`, `bunx`, etc.
- Go tests: `go test ./...`
- Frontend type-check: `bunx svelte-check`

## Development Philosophy

- Build the smallest working slice first — no speculative abstractions
- Each commit must leave the app in a working state
- If a feature feels large, split it into smaller independently-shippable pieces
- No new packages without discussing the tradeoff — keep dependencies minimal

## After Every Commit

Update **ROADMAP.md**:
- Move completed items from Planned → Done (with version or short description)
- Add new ideas to Planned as they come up
- Keep In Progress honest — only what's on the current branch

## Keeping Docs Up To Date

**ARCHITECTURE.md** — update when:
- Adding or removing a package under `internal/`
- Changing the database schema
- Adding new API endpoints or changing existing ones
- Changing the deployment setup

**PROJECT.md** — update when:
- A new game mode or major feature is scoped out
- An open question gets answered
- V-scope changes (something moves from V2 → active work)
