# Full-Stack Learning Tasks

These are starter tasks designed to help you develop both frontend and backend skills in the OpenPadel codebase. Each task involves touching both the Go backend and SvelteKit frontend, with progressively different focuses.

---

## Task #1: Add Tournament Win Counter to User Profile

**Focus:** Backend-heavy (migrations, store updates, API)

Add a field to track how many tournaments a user has won and display it on their profile page.

### Backend Work
- Add a `win_count` column to the `users` table (migration)
- Update `internal/store/users.go` to increment `win_count` when a user finishes in 1st place
- Modify the profile API handler to return `win_count` in `GET /api/auth/me` and `GET /api/auth/profile`

### Frontend Work
- Update `App.User` type in `src/app.d.ts` to include `win_count`
- Display the win counter on the profile page (`web/src/routes/profile/+page.svelte`) alongside existing stats

### Key Learning
- How to write goose migrations (`internal/store/migrations/`)
- How to update store layer queries (raw SQL → sqlc pattern)
- How to modify API response contracts
- How to update TypeScript types and display new data in Svelte

---

## Task #2: Add Session Invite Limit and Validation

**Focus:** Balanced (validation patterns, error handling)

Prevent users from inviting the same person to the same session twice with a friendly error message.

### Backend Work
- Add validation in `internal/api/invites.go` POST handler to check if an invite already exists with status="pending"
- Return `400 { "error": "already_invited" }` if duplicate
- Write a store query in `internal/store/invites.go` to check existing pending invites

### Frontend Work
- Update the invite button in the invite UI to disable if the invitee is already invited
- Show a toast message when invite fails with the error reason
- Make use of the `svelte-sonner` package (already installed) for error feedback

### Key Learning
- How to validate business logic at the API layer
- How to write descriptive error responses
- How to handle API errors in frontend components
- How to use toast notifications for user feedback

---

## Task #3: Add Player Last-Played Date to Leaderboard

**Focus:** Balanced (migrations + data flow + UI)

Track when each player last participated in a session and display it on the career stats leaderboard.

### Backend Work
- Add a `last_played_at` column to the `players` table (migration)
- Update it whenever a player joins a session in `internal/store/players.go`
- Include it in the history API response (`GET /api/auth/history`)

### Frontend Work
- Update the history/career view to show "Last played: X days ago"
- Format the timestamp nicely using a date utility function

### Key Learning
- How migrations work and when to apply them
- How to update store queries to include new fields
- How to format timestamps in the UI (relative time)
- How data flows from database → API → frontend

---

## Task #4: Implement Basic Player Search Filtering on Contacts Page

**Focus:** Frontend-heavy (state management, reactive filtering)

Add a client-side filter to the contacts list so users can search by name or quickly filter by recently-played-with.

### Frontend Work
- Add a search input and filter buttons to `web/src/lib/components/Contacts.svelte`
- Filter the contacts list based on input (case-insensitive name match)
- Add buttons to quickly filter by "Recently Played" (sort by `last_played_at`)
- Use Svelte 5 runes for reactive state management

### Backend Work
- Ensure `GET /api/contacts` returns the full contact list with player stats
- Verify the response shape matches what the frontend needs (no changes likely needed)

### Key Learning
- How to work with Svelte 5 runes (`$state`, `$derived`)
- How to build reactive filters and sorting
- How to structure component state for clarity
- How API responses drive UI requirements

---

## Task #5: Add Simple Validation Error Toast on Score Submission

**Focus:** Full-stack error handling and user feedback

Wire up error feedback for invalid score submissions so users see a message instead of silent failures.

### Backend Work
- Ensure score validation errors are clear (e.g., "scores must sum to X points" for Americano)
- Return `400` with a descriptive error message in `internal/api/rounds.go`
- Write tests to verify error responses (use existing test patterns in `internal/api/*_test.go`)

### Frontend Work
- Catch submit errors in `web/src/lib/components/ActiveSession.svelte`
- Display them as a toast using `svelte-sonner` instead of console errors
- Show the error message to the user immediately
- Clear the error after a timeout or user action

### Key Learning
- How to test error paths in Go (httptest patterns)
- How to propagate errors gracefully through the API
- How to handle promise rejections in Svelte
- How to provide real-time feedback to users

---

## Getting Started

### Before You Code

1. **Read the architecture:** Check `ARCHITECTURE.md` for the overall structure
2. **Find similar code:** Use Grep to find similar patterns before implementing
3. **Check tests:** Look at `internal/store/*_test.go` and `internal/api/*_test.go` for the patterns

### General Workflow

```
1. Create a feature branch (e.g., feat/win-counter)
2. Implement backend changes (migrations first, then store, then API)
3. Write tests: go test ./...
4. Implement frontend changes
5. Type-check: bunx svelte-check
6. Commit with conventional commits (feat:, fix:, etc.)
7. Create a PR
```

### Pro Tips

- **Start with Task #1 or #2** — they have clearer scope and teach important patterns first
- **Write tests as you go** — the test patterns are already set up, just follow them
- **Search the codebase** — someone has probably done something similar already
- **Ask questions** — don't get stuck; the code is well-organized
- **Keep commits small** — each commit should leave the app in a working state

### Useful File Locations

- **Migrations:** `internal/store/migrations/*.sql`
- **Store layer:** `internal/store/*.go` (data access, sqlc queries)
- **API handlers:** `internal/api/*.go` (request/response logic)
- **Frontend types:** `src/app.d.ts`
- **Components:** `web/src/lib/components/` and `web/src/routes/`
- **API client:** `web/src/lib/api/client.ts`

### Commands You'll Use

```bash
# Test backend
go test ./...

# Type-check frontend
bunx svelte-check

# Run dev server
go run ./cmd/server

# Bun commands (npm equivalent)
bun install
bun run dev
```

---

## Task Dependencies

The tasks are independent but learning-wise:
1. Start with **#1** (teaches migrations and store layer)
2. Then **#2** (applies validation patterns)
3. Then **#3** (combines both concepts)
4. Then **#4** (pure frontend practice)
5. Finally **#5** (integrates error handling everywhere)

Pick one and go! You've got this.
