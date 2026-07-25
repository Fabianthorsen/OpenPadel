# ADR 0005: Club Leaderboard Metric

## Status

Proposed

## Context

Part of the Clubs feature (see the `Clubs — wayfinder map` issue). A **Club** is a group of
registered **Users**; the club leaderboard ranks members across the club's sessions. No
cross-session rating exists today — within-session `Standing` ranks Players by cumulative points,
and that's it.

The ranking-metrics research (cited to
Glickman's Glicko papers, FIDE, TrueSkill, and the FIP padel ladder) surveyed **Elo, Glicko-2,
cumulative points, win-rate, and average points-difference** against OpenPadel's realities:
rotating partners (Americano), standings-adaptive pairings (Mexicano), variable point targets
(16 / 24 / 32), 2v2 matches with a per-team `Score`, and **guests** who play but must not accrue.
Its lean was a hybrid two-track board (an explainable headline metric + an optional Glicko-2 skill
layer), leaving five sub-decisions open. This ADR closes them.

## Decision

Ship a **single, explainable headline metric** as the club leaderboard. Glicko-2 is recorded here
as a *deferred* alternative, not built now.

### The metric — average normalized points-difference per game

For each **fully-scored** `Match` a member played in a qualifying club event:

```
margin = (memberTeamPoints − opponentTeamPoints) / pointsTarget        ∈ [−1, +1]
```

`pointsTarget` is the session's 16/24/32. Because the two team scores sum to the target, this is
equivalently `2·(memberTeamPoints / target) − 1`, so it's automatically comparable across targets.
A member's leaderboard value is the **mean of `margin`** over their qualifying games: `+1` = won
every game to a shutout, `0` = broke even, `−1` = lost every game to a shutout.

*Rejected:* **Glicko-2 / Elo** (opponent-adjusted skill, but least explainable and highest
implementation cost — deferred, see below; plain Elo further rejected for having no uncertainty
term ⇒ poor cold start, per Glickman); **win-rate** (margin-blind and noisy at low game counts);
**cumulative total points** (rewards attendance over skill, not comparable across targets).

### What counts — club events only, within a rolling window

A game counts toward a club's leaderboard **iff** it belongs to a Session owned by that club
(`club_id`). This avoids all multi-club attribution ambiguity and keeps a member's casual pickup
games off the club board. Only **fully-scored** matches count (live/in-progress scores excluded).

**Rolling window:** only club events dated within the last **N days (default 90, tunable)** count;
older games age out. The window is applied by session date, so a whole session's scored matches are
in or out together.

### Guests / mixed games — all games count for members

Every scored game a member plays in a qualifying club event counts toward their average, **whether
or not** teammates/opponents were guests. Guests never get a leaderboard row. The metric is
points-based and opponent-identity-blind, so there's no principled reason to exclude guest games;
guest-driven luck averages out over many games.

### Qualification & display — provisional below a games gate

A member is **ranked** only after **≥ MinGames** qualifying games in the window (default **5**,
tunable). Below that, they appear in a separate **"provisional / not yet ranked"** section showing
progress ("N more games to rank"). Ranked members sort by average `margin`, descending.

### Tie-breaks

Equal average → **more qualifying games** ranks higher (rewards engagement) → earlier
first-qualifying date → display name (stable, deterministic).

### Cold start, inactivity, seasons

Cold start is handled by the provisional gate. Inactivity is handled **implicitly by the rolling
window** — a member with no qualifying games in the window simply drops off the board; there is no
explicit decay term. **No seasons or resets.**

### Computation — recompute on read

The metric is a mean, hence **order-independent**, so no materialized state or incremental update is
needed. Compute on read: a single SQLite aggregate over the club's club-event matches within the
window, joined to members via `user_id`. Materialize only if profiling later demands it.

## Consequences

- **Opponent strength is invisible** — beating strong and weak opponents count equally. Accepted for
  v1 explainability; because raw matches are stored, the deferred Glicko-2 layer can be computed
  **retroactively** with no data loss if opponent-adjusted skill is later wanted.
- **Margin is rewarded**, normalized across targets. In Americano/Mexicano running up the score *is*
  the game, so this is intended, not a loophole.
- **Rank can change without you playing** — as your old games (or others') age out of the window.
  This is the cost of a "current form" board; explainable and accepted.
- **Guests can nudge a member's average** (a strong guest partner flatters it; a weak guest opponent
  inflates margin). A known, accepted imprecision of a points-based metric; averages out over games.
- **New members are provisional before ranking** — intended, to keep the ranked board meaningful.

### Deferred (considered, not chosen for v1)

- **Glicko-2 skill-rating layer** — natively handles cold start (RD), inactivity (RD growth),
  opponent strength (`g(φ)`), and guests (high-RD provisional). Superior as a *skill* measure but
  least explainable and most complex (rating periods + iterative volatility solver). Recompute-able
  retroactively from stored matches. See the research doc.
- **Seasons with resets** and season champions.
- **Opponent-strength weighting / margin caps** on the headline metric.
