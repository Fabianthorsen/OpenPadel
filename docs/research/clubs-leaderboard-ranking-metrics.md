# Ranking metrics for a cross-session club leaderboard

**Purpose:** Survey candidate ranking metrics for OpenPadel's forthcoming cross-session **club
leaderboard** and describe how each behaves under OpenPadel's specific realities (Americano /
Mexicano rotation, fixed-sum point targets, guests, Go + SQLite single-binary).

> This document **informs but does not make** the decision. It is input for the leaderboard
> decision ticket. It deliberately ends with the open sub-decisions rather than resolving them.

The five core metrics surveyed: **cumulative points**, **win-rate**, **average points-difference
per game**, **Elo**, **Glicko-2**. **TrueSkill** and the **official padel points ladders** are
included only as reference points.

---

## OpenPadel realities these metrics must survive

These are treated as given (from the codebase / ticket), not re-derived:

- **Americano** — individual scoring, partners **rotate every round**; your personal points =
  the points your team scored in each match you played.
- **Mexicano** — pairings **adapt by standings** each round (1+4 vs 2+3, …); individual scoring;
  no bench; exactly `courts × 4` players.
- A **Session** plays to a point **target of 16 / 24 / 32** (the two team scores in a match sum to
  the target); variable number of rounds.
- **Matches** are 2v2 with an exact score `(a, b)`; the schema stores which 4 players, which side,
  and the score per match.
- **Guests** (name-only, no account) may play and appear as partners/opponents but **must NOT
  accrue** on the club leaderboard. Only registered **Users** accrue.
- Data available: every club session's matches (team composition + score + `game_mode`). Backend
  is Go + SQLite, single-writer. The leaderboard should be **fully recomputable from stored match
  data**, or incrementally updatable on score submission.

---

## Comparison table

Axes are rows; metrics are columns. Cells are terse — nuance is in the per-metric sections below.

| Axis | Cumulative points | Win-rate | Avg points-diff / game | Elo | Glicko-2 |
|---|---|---|---|---|---|
| **1. Input needed** | Aggregate points only | Match W/L/D | Aggregate score margins | Pairwise/team outcomes + opponent ratings | Pairwise/team outcomes + opponent rating **and RD**, batched into rating periods |
| **2. Rotating-partner attribution** | Non-issue — personal points are already individual (your team's score credited to you); but partner/opponent-blind | Team W/L shared by both partners; rotation averages partner luck over many rounds | Team margin credited to you; rotation averages partner effect | Needs a **team heuristic** (team rating = avg of the 2 members; both get same delta). Many fresh 2v2s = lots of paired-comparison signal | Same team-avg heuristic; the paper's "player vs *m* opponents in a period" maps cleanly onto rotation (see notes) |
| **3. Cold start** | Starts at 0; fair but rewards volume, not skill | Undefined / wildly noisy at low *n* (0% or 100% after 1 game) | Undefined at *n*=0; noisy but steadier than win-rate | Everyone 1500, **no uncertainty** — a rookie's 1500 treated as reliable as a veteran's (Glickman's own critique of Elo) | 1500 / **RD 350** / σ 0.06; RD *encodes* "unknown", moves fast, shown as CI. Principled cold start |
| **4. Inactivity / decay** | None (rewards attendance); bolt on rolling window | None (recency-blind) | None (recency-blind) | **None built in** — no uncertainty term; rust must be bolted on | **Built in**: idle players' RD grows `φ*=√(φ²+σ²)` each period. No bolt-on |
| **5. Guests / unrated opponents** | Trivial: only sum registered users' rows (but a strong guest partner still inflates your total) | Trivial exclusion; but opponent strength invisible | Trivial exclusion; margins vs weak guests inflate | Must give guests *some* rating (provisional 1500, fixed) or drop the match — both lossy | Give guest 1500 + **high RD**; `g(φ)` auto-down-weights the game. Best principled fit |
| **6. Variable target (16/24/32) & margin** | Is margin-bearing, but raw totals **not comparable** across targets (32 ≈ 2× a 16); normalize | **Ignores** margin (16–14 == 16–2); draws only at exact half | **Is** margin; must normalize by target; rewards score-running | Native = W/L only; margin needs a MOV extension | Native = W/L/D (draws via s=0.5); margin needs a MOV extension |
| **7. Explainability** | Highest — "points you scored" | High — "% of games you win" | Medium — "you outscore opponents by N/game" (target caveat is subtle) | Medium — 1500=avg, higher=better; per-game deltas feel opaque | Lowest — 3 numbers; surface rating only, hide RD/σ |
| **8. Impl cost & recompute** | Trivial; order-independent `SUM`; incremental & fully recomputable | Trivial; needs min-*n* / shrinkage | Trivial; order-independent mean | Moderate; **order-dependent** (replay in match order); incremental is natural | Highest; needs rating-period design + iterative volatility solver; batch recompute per period; deterministic given fixed periods |

---

## Per-metric notes

### Cumulative points
Sum of your personal points across every club match. **Best explainability and lowest cost** of
any option — it is literally what Americano/Mexicano already compute, so the rotating-partner
problem simply does not arise (each round's team score is credited to you). It is fully
recomputable with a single order-independent `SUM`, and guests are handled by excluding their
rows. Weaknesses: it is **opponent- and partner-blind** (beating a strong pair counts the same as
a weak one; a strong partner inflates you), it **rewards attendance over skill** (whoever plays
most tops the board), and raw totals are **not comparable across point targets** — a 32-target
session yields roughly twice the points of a 16-target one for equal performance. Both weaknesses
are fixable in design: normalize (points **per game**, or as a fraction of the target) and apply a
**rolling recency window** — which is exactly what the pro padel ladders do (see reference points).

### Win-rate
Fraction of matches won, with the round result attributed to both partners. Cheap and
understandable. Rotation is a mild *help* here: over many rounds, partner luck averages out. But
it is **very noisy at low sample sizes** (100% after one win) and needs a minimum-games threshold
or a shrinkage prior (Beta/Wilson) before it is trustworthy. It **discards margin entirely** —
which may be desirable in a social setting (don't reward running up the score) or a loss of signal,
depending on product intent. In OpenPadel a draw is only possible when a team scores exactly half
the target (e.g. 8–8 at 16).

### Average points-difference per game
Mean of `(yourTeamScore − opponentScore)` per match. A sensible middle ground between "points"
and "win/loss": it keeps margin granularity while normalizing for the variable number of rounds
(it's a per-game average). It is steadier than win-rate at low *n* and still trivial and
order-independent to compute. Two caveats: margins **scale with the target** (a +8 margin is
dominant at target 16 but moderate at 32), so it must be normalized by target to be comparable;
and it remains **opponent-blind** — a big margin against a weak guest pair looks like skill.

### Elo
Classic paired-comparison rating. Expected score is the logistic
`E = 1 / (1 + 10^((R_opp − R_self)/400))` and the update is `R' = R + K·(S − E)` with `S ∈ {0, ½, 1}`
([FIDE Rating Regulations](https://handbook.fide.com/chapter/B022024);
[Elo overview](https://en.wikipedia.org/wiki/Elo_rating_system)). FIDE's K is **40** for new
players (until 30 games) and U-18s under 2300, **20** under 2400, **10** at/after 2400 — i.e. K is
the convergence-speed knob.

- **Doubles attribution** (the hard part) is handled by a widely-used heuristic: treat each team's
  rating as the **average of its two members**, compute each side's expected score, then move
  **both** members of a team by the same delta `K·(S − E)`
  ([2v2 Elo write-up](https://towardsdatascience.com/developing-an-elo-based-data-driven-ranking-system-for-2v2-multiplayer-games-7689f7d42a53/);
  [SliceWin](https://slicewin.com/elo-system)). This ignores the intra-team skill split (partners
  always move together) but fits OpenPadel well: Americano/Mexicano generate a large number of
  varied 2v2 matchups, which is exactly the paired-comparison data Elo consumes.
- **Its two structural gaps are exactly OpenPadel's pain points.** Elo has **no uncertainty term**,
  so a rookie's 1500 is treated as reliable as a veteran's — this is the specific deficiency
  Glickman built Glicko to fix ([Glicko paper](http://www.glicko.net/glicko/glicko.pdf)). And it
  has **no decay**: rust after weeks off must be bolted on. Guests must be handed *some* number
  (provisional fixed 1500) or their matches dropped — both lossy, and dropping is costly in a
  format where opponents are often guests.
- **Recompute** is order-dependent (a rating at match time depends on all prior matches), so a full
  rebuild must replay matches in a fixed chronological order — deterministic and cheap at club
  scale, and incremental update on submission is the native mode.

### Glicko-2
Glickman's extension of Glicko/Elo that adds an explicit **uncertainty** (rating deviation, RD) and
a **volatility** σ. Unrated players start at **rating 1500, RD 350, σ 0.06**; the system constant
**τ ∈ [0.3, 1.2]** constrains volatility change. Ratings are mapped to an internal scale by
dividing by **173.7178**, updated against *m* opponents in a **rating period**, then mapped back
([Glicko-2 paper](http://www.glicko.net/glicko/glicko2.pdf)). Outcomes are `s ∈ {0, ½, 1}`.

Why it fits OpenPadel's realities better than plain Elo:

- **Cold start & explainable confidence.** RD 350 literally means "we don't know this player yet";
  the paper recommends summarizing strength as a 95% CI of **rating ± 2·RD**, which is a
  player-friendly "we're 95% sure you're 1600 ± 120."
- **Inactivity is native, not bolted on.** A player who sits out a rating period has RD grown by
  `φ* = √(φ² + σ²)` (Glicko-1's equivalent: `RD = min(√(RD² + c²), 350)`). This directly models
  "hasn't played in weeks," which matters for a club with sporadic attendance.
- **Guests handled principledly.** Give a guest rating 1500 with a **large RD**; the term
  `g(φ) = 1/√(1 + 3φ²/π²)` shrinks a high-RD opponent's influence, so games against unknown guests
  move a member's rating only a little — an information-weighted answer to the unrated-opponent
  problem that Elo can't express.
- **Rotation maps onto the model.** Glicko-2 already processes one player against *many* opponents
  within a period; the paper even labels the variance term *v* as "the estimated variance of the
  **team's**/player's rating." You still need the team-average heuristic for the partner, but the
  many-opponents-per-period shape matches Americano/Mexicano rotation naturally.

Costs: it is the **most complex** to implement (an iterative volatility solver — though the paper
gives a fully-specified, deterministic "Illinois algorithm" for it), it needs a **rating-period
design**, margin is not native, and it is the **least explainable** (three numbers — you'd surface
only the rating). Recompute is order-dependent at the period level but deterministic given a fixed
period assignment: re-bucket matches into periods and replay.

**Rating-period granularity is the key design choice.** Glicko-2 is a *batch* system and "works
best … at least 10–15 games per player in a rating period" ([Glicko-2 paper](http://www.glicko.net/glicko/glicko2.pdf)).
OpenPadel's natural batches are a **Session**, a **week**, or a **month** of sessions. A single
session likely gives a player only ~5–8 rounds (below the guidance), so a **weekly or monthly**
rating period is the better fit and keeps "recompute from match history" clean (just re-bucket and
replay periods in order).

---

## Reference points (not core candidates)

### TrueSkill (Microsoft Research)
The most principled answer to the rotating-partner problem: TrueSkill models a **team's performance
as the sum of its members' performances** and **infers individual skills from team results**, while
tracking per-player uncertainty (Gaussian `μ, σ`; defaults μ=25, σ=25/3≈8.333, β=4.167, τ=0.083,
draw prob 0.1) and converging faster than Elo
([Microsoft Research](https://www.microsoft.com/en-us/research/publication/trueskilltm-a-bayesian-skill-rating-system/);
[trueskill.org](https://trueskill.org/)). It "works well with any … N:N team game or free-for-all."
Downside for us: factor-graph / expectation-propagation math is heavy to implement and recompute in
a small Go + SQLite binary, and it is the least explainable of all. A good conceptual north star,
almost certainly over-engineered for a social club board.

### Official padel ladders (FIP / Premier Padel)
Pure **points accumulation**: a player's ranking is the sum of their **best 22 results** over a
**rolling 52-week window**, with points awarded by tournament tier and round reached and **expiring**
after 52 weeks ([Padel FIP](https://www.padelfip.com/ranking-system-points-breakdown/)). This is a
cumulative-points ladder with a recency window and best-N filter — **not** a skill rating (no
opponent adjustment, no uncertainty). Useful validation that "cumulative points + rolling recency
window" is a legitimate, explainable design even at the top of the sport.

---

## Recommendation

For OpenPadel's audience (casual, social — explainability is a first-class requirement), its
realities (rotating partners, sporadic attendance, frequent guests) and its constraints (small
Go + SQLite single binary, recompute-from-history), I would lean toward a **hybrid, two-track**
leaderboard rather than one metric:

1. **Headline board = a normalized points/results metric with a rolling recency window.** Concretely,
   **average points-difference per game (normalized by target)** or **win-rate with a min-games
   shrinkage** as the visible ranking, over a rolling window (mirroring the FIP model). Rationale:
   it is what the game modes already produce, instantly explainable, trivial and order-independent
   to recompute, and guests fall out by simple exclusion. This is the safe default for a social app.

2. **Optional "skill rating" layer = Glicko-2.** It is the only surveyed metric that natively solves
   two of OpenPadel's actual problems — **cold start** (RD) and **inactivity** (RD growth) — and it
   handles guests in an information-weighted way (high-RD provisional opponents). Its batch
   rating-period model maps cleanly onto "a week/month of club sessions" and stays fully recomputable
   from match history. Surface only the rating (with an optional confidence band); hide RD/σ.

I would **not** pick **plain Elo** as the sole skill metric — its lack of an uncertainty term gives
a poor cold start and no decay (Glickman's own motivation for Glicko), and Glicko-2 supersedes it
for the same data at modest extra cost. **Win-rate alone** is too noisy at low *n* and margin-blind;
**cumulative totals alone** reward attendance and aren't comparable across targets; **avg
points-diff** is a strong, cheap headline but opponent-blind, which is precisely the gap the
optional Glicko-2 layer fills.

### Sub-decisions this leaves open for the decision ticket
1. **Guests:** exclude guest matches from the rating entirely, or include guests as fixed-strength
   / high-RD provisional opponents so their matches still inform members? (High-RD provisional is
   the principled Glicko-2 answer; plain exclusion is fine for the points-based headline.)
2. **Does margin count?** Margin-bearing metrics reward running up the score; W/L-based metrics
   ignore it. This is a product-intent call. If margin counts, it **must** be normalized by the
   16/24/32 target to be comparable.
3. **Glicko-2 rating-period granularity** (if adopted): per-session vs weekly vs monthly — trades
   convergence speed against the paper's ~10–15-games-per-period guidance.
4. **Headline metric identity:** total points (celebrates participation) vs per-game average
   (celebrates quality) vs win-rate — which behaviour do we want to reward?

---

## Sources

Primary sources:

- Mark E. Glickman, *Example of the Glicko-2 system* (2022) — equations, defaults (1500 / RD 350 /
  σ 0.06), τ ∈ [0.3, 1.2], 173.7178 conversion, 10–15 games/period, inactivity `φ*=√(φ²+σ²)`:
  <http://www.glicko.net/glicko/glicko2.pdf>
- Mark E. Glickman, *The Glicko system* — RD as reliability, RD growth over time
  `RD = min(√(RD² + c²), 350)`, and the critique of Elo's missing reliability measure:
  <http://www.glicko.net/glicko/glicko.pdf>
- FIDE Rating Regulations (effective 1 March 2024), FIDE Handbook B.02 — Elo update
  `Rn = Ro + K·(S − PD)` and K-factors (40 / 20 / 10): <https://handbook.fide.com/chapter/B022024>
- Herbrich, Minka & Graepel, *TrueSkill™: A Bayesian Skill Rating System*, Microsoft Research
  (NIPS 2006) — team model, individual-skill inference, faster convergence than Elo:
  <https://www.microsoft.com/en-us/research/publication/trueskilltm-a-bayesian-skill-rating-system/>
- TrueSkill implementation reference (default μ/σ/β/τ/draw prob, N:N team support):
  <https://trueskill.org/>
- Padel FIP, *Ranking System & Points Breakdown* — best-22 results, rolling 52-week window, points
  expiry: <https://www.padelfip.com/ranking-system-points-breakdown/>

Supporting (authoritative but not canonical) — doubles/2v2 Elo attribution heuristic:

- *Elo overview* (expected-score logistic, zero-sum updates): <https://en.wikipedia.org/wiki/Elo_rating_system>
- *Developing an Elo-based ranking for 2v2 games* (team-average rating, average of expected scores):
  <https://towardsdatascience.com/developing-an-elo-based-data-driven-ranking-system-for-2v2-multiplayer-games-7689f7d42a53/>
- *SliceWin Elo for racket sports* (team-average Elo for doubles): <https://slicewin.com/elo-system>
