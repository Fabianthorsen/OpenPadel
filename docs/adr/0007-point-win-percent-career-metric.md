# ADR 0007: Point Win % as the Career Skill Metric

## Status

Proposed

## Context

The profile shows a per-User career stat headlined by **winrate** (`wins / games_played`, Americano
only — `GetAmericanoCareerStats`). Winrate is a poor "how good am I" signal for this game:

- **It isn't yours.** Americano rotates partners so everyone partners everyone; half your winrate is
  your rotating partner's play, not yours. (Mexicano adapts partners to standings — same dilution,
  different cause.)
- **It's margin-blind.** A 24–8 demolition and a 16–15 squeaker are both "1 win."
- **It's noisy at low game counts** — one game reads 0% or 100%.

Scoring in both modes is fundamentally *individual points* (Americano credits points to the player;
Mexicano still records a per-team `Score`), so a points-based metric expresses skill far better than
a binary win. This is the same conclusion the club-leaderboard work reached — see
**[ADR 0005](./0005-club-leaderboard-metric.md)**, whose headline metric is the mean of
`margin = (memberTeamPoints − opponentTeamPoints) / pointsTarget ∈ [−1, +1]`.

Two design pressures are specific to a *career/profile* surface (vs. the club board):

1. **Sessions use different points targets** (16 / 24 / 32), so any raw-points metric mixes scales
   across a career. Normalization is mandatory, not optional.
2. **A profile spans Game Modes.** Blending Americano and Mexicano *raw points* is meaningless, but
   a *normalized fraction* stays a valid fraction in either mode — so a unified cross-mode summary
   is defensible where a unified raw total would not be.

## Decision

**Adopt ADR 0005's normalized-margin metric as the career skill metric, presented on the profile as
"Point Win %".** It is the same underlying number, rescaled for legibility and applied at career
(rather than club) scope.

### The metric

For each **fully-scored** `Match` a player played:

```
share = yourTeamScore / (yourTeamScore + opponentTeamScore)          ∈ [0, 1]
```

Because the two team scores sum to the session's points target, this is identical to ADR 0005's
`share = memberTeamPoints / pointsTarget`, and `Point Win % = share × 100 = (margin + 1) / 2 × 100`.
**They are the same metric**; Point Win % is its 0–100% presentation, chosen because "56% of points
won" reads as a familiar win-percentage to a casual user where "+0.12 margin" does not.

The career figure is the **mean of per-Match `share`** (each Match weighted equally, regardless of
its target). Order-independent, so compute-on-read; no materialized state.

*Rejected:* **raw points-per-game** (mixes scales across targets, and rewards long-target sessions);
**point-pooled share** (`Σ yours / Σ all` — re-weights toward long-target games, reintroducing the
scale problem normalization removes); **replacing winrate entirely** (see below).

### Presentation — unified summary, segmented detail

- **Profile:** a single **cross-mode summary** of `{Point Win %, winrate, games}`. Point Win % and
  games blend across modes cleanly (fraction / count); winrate is retained here **only as a familiar
  secondary read**, not as the skill signal ADR 0005 rejected it as.
- **Career Stats page:** the full metric set **segmented per Game Mode**. Point-share is only
  apples-to-apples *within* a scoring model, so per-mode sections keep every number meaningful.
  Rendered from a data-driven metric catalog over a per-Session results series (so trend/form/streak
  stats derive client-side without new endpoints).
- **Placement stats are the one cross-mode exception** (titles, podiums, best / average finish;
  #229). A *finishing rank* compares like-for-like across scoring models — 1st is 1st in Americano
  or Mexicano — so unlike point-share it is honest to blend. These live in the cross-mode summary
  (`CareerSummary`), not the per-mode sections. Derived on read from each done Session's leaderboard
  rank, reusing the tournament-history walk.

### Low sample counts

No minimum-games gate beyond hiding the stat at zero games. Unlike winrate, a normalized share
degrades gracefully — one game reads ~54%, never a jarring 0% / 100%.

## Consequences

- **One canonical skill number across the app.** Profile Point Win % and the club board are the same
  metric at different scopes/windows; they cannot drift apart in definition. A future reader must
  treat 0005 and 0007 as one metric, two presentations.
- **Mexicano mutes the signal slightly.** Mexicano hands stronger players weaker partners to balance
  standings, compressing everyone's share toward 50%. Accepted for the unified profile summary;
  the per-mode Career Stats page keeps Mexicano's numbers in their own section so the compression
  isn't blamed on Americano.
- **Winrate survives as a secondary stat**, contradicting 0005's rejection of it *as a ranking
  metric* — but only as a legible companion on a personal profile, never as the ranked/headline skill
  signal. No conflict in intent.
- **Margin is rewarded** (running up the score raises your share). In both modes that *is* the game,
  so it's intended, not a loophole — consistent with 0005.
- **Guest opponents/partners nudge the number** (points-based, opponent-identity-blind). Averages out
  over games; same accepted imprecision as 0005.

### Deferred (considered, not chosen now)

- **Relational stats** (best partner, head-to-head / nemesis) — need per-Match opponent identity.
  The data already supports it (`Match.p1–p4 → players → user_id`); it's an unwritten query, not a
  schema gap. Registered-vs-registered only, since Guests have no stable identity. A future reach.
- **Opponent-strength weighting / Glicko-2 skill layer** — same deferral and rationale as 0005;
  recompute-able retroactively from stored matches.
