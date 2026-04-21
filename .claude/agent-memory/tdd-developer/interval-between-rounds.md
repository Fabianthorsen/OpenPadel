---
name: Interval Between Rounds Implementation
description: TDD patterns and design decisions for configurable round intervals (1-5 minutes)
type: reference
---

## Implementation Completed

### Database Layer
- Migration 006: Added `interval_between_rounds_minutes INTEGER` to sessions table (nullable)
- sqlc generated correct types: `IntervalBetweenRoundsMinutes` as `sql.NullInt64`
- All existing data backward compatible (null values supported)

### Domain Layer
- Session struct field: `IntervalBetweenRoundsMin *int` (optional pointer)
- Persisted through store: rowToSession() unpacks interval from DB row
- Default to 3 when nil during business logic

### Round Duration Math
**New formula:** `T = (D*60 - (R-1)*I*60 - R*B) / R`
- D = total duration (minutes)
- R = number of rounds
- I = interval between rounds (minutes)
- B = buffer per round (seconds)
- Explanation: Total time minus interval time minus all buffer time, divided by rounds

**Test coverage:**
- Even/odd player counts (8, 9 players)
- Large groups (16 players)
- Boundary intervals (1, 5 minutes)
- Minimum validation (120 seconds per round enforced)
- Mid-tournament recalculation with remaining rounds/seconds

### API Validation
**Request body field:** `interval_between_rounds_minutes: number (1-5)`
- Timed Americano mode only
- Default: 3 when omitted
- Validation: 1 ≤ value ≤ 5
- Error: 400 Bad Request if out of bounds

**Test coverage:**
- Default interval (3) when omitted
- Custom interval (1-5) accepted
- Below minimum (0) rejected
- Above maximum (6) rejected

### Store/Service Integration
- Store.CreateSession() accepts intervalBetweenRoundsMin param
- Store.StartTimedAmericanoSession() accepts and persists interval
- Service.Start() reads interval from session, defaults to 3 if nil
- Service.AdvanceRound() uses interval for RecalculateRoundDuration()
- All store tests updated to pass interval parameter

## Test Counts
- Timed gamemode tests: 8 tests (6 existing + 2 new boundary tests)
- API tests: 4 new tests (default, custom, too low, too high)
- Store tests: 5 updated to include interval parameter
- Total: All tests pass, no regressions

## Key Design Decisions

**Why default to 3 minutes?**
- Provides reasonable rest between intense rounds
- Allows cleanup/court setup without being excessive
- Configurable for user preference (1-5 range)

**Why 1-5 minute bounds?**
- 1 min: Minimal breaks for back-to-back fast play
- 5 min: Significant rest without excessive downtime
- Beyond 5: Tournament duration prediction becomes unrealistic

**Why optional field?**
- Backward compatibility for existing sessions
- nil value converted to default 3 during calculations
- Stored in DB allows future session inspection/replay

## Next Steps (Frontend/Countdown)

These features are documented in memory for frontend implementation:
1. CreateDrawer: PillToggleGroup for intervals 1/2/3/4/5 (default 3)
2. CountdownTimer: Show interval countdown before round auto-advance
3. Admin action: Click to start round early (skip countdown)
