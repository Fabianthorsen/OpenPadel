# ADR 0001: Build then remove Timed Americano entirely

## Status

Accepted (implemented in `c1efa6d`, closes #58, parent #57)

## Context

Timed Americano was a third game mode: fixed tournament duration, free-form scoring, and
drift-corrected round timers (`RecalculateRoundDuration`, `timer_sync` SSE events, buffer/interval
config). It shipped with a full UI (`RoundTimer`, buffer/interval pickers in `CreateDrawer`).

While redesigning the session state machine (#57 — moving to intent-driven `lobby → playing →
done`, extracting pairing logic into reusable modules), the mode's configuration knobs didn't
align with the rest of the app's game-rule model. Its buffer/interval parameters had already been
simplified once (`T = (D*60 - (R-1)*I*60 - R*B) / R` → `T = (D*60) / R`) before the decision was
made to cut it rather than keep simplifying it.

## Decision

Delete Timed Americano entirely — domain constants, API handlers, gamemode service/rounds package,
timer UI components, `timer_sync` SSE event — rather than keep it as a maintained third mode.
Per #58: "it's experimental and adds complexity without clear use case."

## Consequences

- Smaller surface area: two game modes (Americano, Mexicano) share one mental model — a fixed
  points target per match, no wall-clock timers to keep in sync across clients.
- The `timer_sync` event type and drift-correction logic are gone; any future timed mode would be
  a fresh design, not a resurrection of this code.
- Precedent set: an experimental mode without a clear use case gets removed outright rather than
  kept behind a flag or half-maintained. Don't re-add timed play without a concrete driver.
