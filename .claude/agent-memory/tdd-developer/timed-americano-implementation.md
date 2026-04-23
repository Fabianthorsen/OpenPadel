---
name: Timed Americano Frontend Implementation Guide
description: TDD patterns, testing conventions, and code structure for timed_americano feature
type: reference
---

## Testing Patterns Established

### API Client Tests (client.test.ts)
- Use `mockFetch()` helper to stub global fetch with Vitest
- Mock responses with status and body
- Extract actual fetch call from mock to verify JSON body: `JSON.parse(call[1].body)`
- Use `toMatchObject()` for JSON comparisons (order-independent)

### Component Logic Tests (RoundTimer.test.ts)
- Test pure utility functions separately from Svelte components
- Functions: `calculateRemaining()`, `getColorClass()`, `formatTimer()`
- Use `vi.useFakeTimers()` / `vi.useRealTimers()` in beforeEach/afterEach
- No testing-library needed for logic-only tests

## Code Structure Decisions

### Session Interface (app.d.ts)
- Added to Session: `total_duration_minutes?`, `round_duration_seconds?`, `round_started_at?`
- Updated game_mode union to include `'timed_americano'`
- All timed fields are optional (backward compatible)

### API Client Refactoring (client.ts)
- Changed `sessions.create(courts, points, ...)` to `sessions.create({...})`
- Builder pattern with optional fields simplifies future param additions
- Default values for sets_to_win=2, games_per_set=6 if not provided
- Conditionally includes optional fields in request body only if provided

### i18n Conventions (en.json, no.json)
- Prefix groups: `create_*`, `timer_*`, `rules_*`, `api_error_*`
- Keys use snake_case, values support {n} interpolation
- Norwegian (no.json) mirrors English structure and placement

## Next Steps (Phase 3+)

### ActiveSession.svelte Changes
- Import RoundTimer component
- Add derived: `const isTimedAmericano = $derived(session.game_mode === 'timed_americano')`
- Render RoundTimer when `isTimedAmericano && session.round_started_at`
- Modify numpad logic:
  - For timed mode: allow any score 0-99 (no sum constraint)
  - For timed mode: collect scores separately (a first, then b)
  - Keep original behavior for americano/mexicano

### CreateDrawer.svelte Changes
- Add state: `totalDurationMinutes = $state<number>(90)`, `bufferSeconds = $state<number>(120)`
- Add toggle pill group to select duration (60, 90, 120, 150, 180 minutes)
- Add toggle pill group to select buffer (120 = 2 min, 180 = 3 min)
- Only show these controls when `gameMode === 'timed_americano'`
- Pass params to `api.sessions.create()` when creating

### Session Page SSE (s/[id]/+page.svelte)
- Add event listener for `timer_sync` event type
- Payload: `{round_duration_seconds, round_started_at, remaining_rounds}`
- Update session state when received to sync timer across clients

## Testing Checklist
- [ ] RoundTimer countdown logic tests pass
- [ ] RoundTimer color state machine tests pass
- [ ] API client test passes (timed_americano params)
- [ ] ActiveSession score entry allows free scoring
- [ ] CreateDrawer collects duration/buffer params
- [ ] Session page SSE listener updates timer_sync props
- [ ] Svelte-check: no type errors on new code
- [ ] All existing tests still pass
