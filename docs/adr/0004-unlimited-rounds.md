# ADR 0004: Unlimited Rounds

## Status

Proposed

## Context

Both game modes force a fixed round count chosen upfront: Mexicano hard-caps `AdvanceRound`
once `nextRound > RoundsTotal`; Americano pre-generates the full bench rotation for
`rounds_total` rounds at `Start()`. There's no way to keep a Session going once that count is
reached, and no way to play without picking a number at all. This was an open question since
the earliest project scoping ("can the admin add extra rounds mid-session if everyone wants to
keep playing?") and never resolved.

Two escape hatches were considered for a Session that hits its round cap mid-play: let the admin
add a specific batch of extra rounds (+1 / +3 / +5), or convert the Session to fully unlimited
with one tap. Batch-add is more granular but means re-prompting every time the new cap is hit
during a long open-ended game — friction that works against the actual ask ("play for as long as
I want").

## Decision

Add **Unlimited Rounds** as a valid Session config: `rounds_total` left unset, rounds generate on
demand until the Admin ends the Session via the existing `close` endpoint (which already supports
ending at any point, any status). Reachable two ways:

- Chosen upfront in setup/Lobby, alongside the numeric rounds picker.
- Entered mid-Session via a one-tap "Keep Playing" action once a fixed Session hits its round cap.

This conversion is **one-way** (fixed → unlimited) and **not incremental** — there is no
batch-add-N-rounds affordance. Applies to both Americano and Mexicano.

For Americano, this requires the round generator (`internal/gamemode/americano/rounds.go`) to move
from "compute the full rotation upfront" to "generate the next round from history reconstructed
from already-played rounds" — the existing greedy algorithm is already per-round and doesn't
structurally need the total round count, so this is a refactor, not an algorithm change.

Today, submitting the last score of the last round **auto-completes** the Session immediately
(`internal/api/rounds.go` — "Auto-complete logic" after `submitScore`), before the admin gets any
say. That auto-complete becomes deferred: at that exact moment, present a "Finish" vs "Keep
Playing" choice instead of completing unconditionally. The Session stays `playing` until the admin
picks one. This choice appears at most once per Session — picking "Keep Playing" removes the cap
entirely, so there's no cap left to hit again.

Shipped sequentially, not bundled: Mexicano first, since its backend already carries most of the
plumbing (`StartMexicanoSession` preserves a null `RoundsTotal` from creation) — mostly a frontend
change to stop forcing a default and add the toggle. Americano's generator refactor ships as a
fast-follow once ready; there's no reason to gate the easy win on the harder one.

Closing (`POST /sessions/:id/close`) an unlimited session requires at least one fully-scored round
— a new guard, since today `close` accepts any status/round count unconditionally. Prevents an
empty leaderboard/history entry from an admin closing before any round finishes.

Out of scope for v1: switching to Unlimited *before* the round cap is reached (e.g. round 3 of 7,
"let's just keep going"). `Lobby.svelte`'s editable config only renders pre-start — there's no
mid-game settings surface today, and adding one is a bigger UI change than this feature needs. The
cap-reached prompt is the only entry point into Unlimited from a fixed Session.

## Consequences

- `ended_early` (used in tournament history) is only meaningful when `rounds_total` is set — an
  unlimited Session that the admin closes was never "on track" for a target, so `ended_early` is
  always `false` for it, regardless of how many rounds were played.
- Americano's "balanced bench" guarantee is exact at full rotation-cycle boundaries. Stopping
  mid-cycle (which unlimited/live-extend inherently allows) can leave one player benched one more
  time than another. The hard constraint (no consecutive bench) still holds; only the soft
  "distributed evenly" aim can drift slightly.
- No granular "+N rounds" control. If that turns out to be needed later, it's a new decision, not
  a reversal of this one — the one-tap conversion stays as the default path.
