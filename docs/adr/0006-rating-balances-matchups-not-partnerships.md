# ADR 0006: Rating Balances Match-ups, Not Partnerships

## Status

Proposed

## Context

Players want more competitive Americano matches — fewer blowouts where a strong pair
steamrolls a weak one. A natural instinct is to add a skill **Rating** and "make the pairings
more even." But Americano's defining property is that **partners rotate so everyone partners
everyone** — the live scheduler (`internal/pairing/americano/scheduler.go`) minimises partner
repeats as its **primary** objective, with scoring done individually.

That property is in direct tension with the obvious reading of "balance by rating." If teams
were formed strong+weak (an 8 always partnered with a 3), then strong players would *never*
partner other strong players, and "everyone partners everyone" would be abandoned. At that point
it is no longer Americano.

The scheduler runs in two steps:

1. **Partner step** (`bestPartnerMatching`) — decides *who partners whom*; minimises partner
   repeats. This is what makes it Americano.
2. **Court-assignment step** (`bestCourtAssignment`) — decides *which two pairs share a court and
   therefore face each other*; today minimises court co-occurrence (how often the same people
   share a court).

The insight is that a competitive *match* does not require balanced *teams*. It requires the two
teams on a court to be **of similar combined strength** — a strong pair facing another strong
pair, a weak pair facing another weak pair. That is entirely a **court-assignment** concern and
leaves the partner step untouched.

Ratings are self-set and unreliable, and much of a real lobby is **Guests** with no account.
Any use of rating has to survive missing data and a mostly-unrated field without doing worse than
today.

## Decision

Rating influences **match-ups, not partnerships**. The partner step is unchanged. Rating enters
only at the **court-assignment step**, where the objective becomes a **weighted blend**:

```
score = ratingGap · W + coOccurrence
```

`ratingGap` is `|teamA_total − teamB_total|` for the court; `W` is tuned so rating usually
dominates while co-occurrence breaks near-ties. Rejected alternatives: making rating the *primary*
objective at the partner step (abandons everyone-partners-everyone — no longer Americano), and a
pure tiebreaker (too weak — the optimal co-occurrence grouping is often forced, so it rarely
changes anything).

Supporting decisions:

- **Scale:** 5 levels, self-selected by label at registration, stored and displayed as the
  number **1–5**. Median is **3**. Each level is anchored to a concrete, self-checkable padel
  milestone so adjacent levels don't blur (the classic Beginner/Improver/Intermediate confusion):

  | # | Name | Description |
  |---|------|-------------|
  | 1 | Beginner | New to padel — still learning the rules, scoring and basic contact. Rallies are short. |
  | 2 | Improver | Knows the rules and can keep a basic rally going, but shots are inconsistent and not yet using the back wall. |
  | 3 | Intermediate | Reliably uses the back wall and lob, hits a bandeja, and understands court positioning. |
  | 4 | Advanced | Controlled and consistent — viboras, smashes, tactical wall play. Competes in local leagues/tournaments. |
  | 5 | Expert | Strong competitive player with a full shot repertoire under pressure; high club / regional level. |

  "Pro" was rejected as the top name — self-selecting "Pro" is intimidating, so few honest players
  pick it, squashing the distribution into 1–4 and starving the balancer of a top rung. "Expert"
  is the deliberately-more-approachable ceiling.
- **Editable in settings.** `User.self_rating` is changeable any time from profile/settings (skill
  changes over time). Editing it updates the User's default seed only; it does **not** retroactively
  change `Player.rating` in sessions already joined (those are per-session snapshots).
- **Storage:** `Player.rating` (per-session, admin-editable in the lobby), seeded from an optional
  `User.self_rating` when a registered User joins. Guests remain first-class — the admin can rate
  any player, including guests.
- **How each Player gets a rating:**
  - *Registered User joins* — seeded from their (always-set) `self_rating`.
  - *Guest self-joins via link* — picks a level **alongside their name, in the same join step**;
    it is **required** to join. This revises the "join by name only" simplicity to "name + level."
  - *Admin adds a guest* — sets the level **next to the name** while adding; **optional**, blank
    defaults to median **3**, editable later in the lobby. This is the one path that can produce an
    unrated Player, so the median fallback stays a real (not merely defensive) case.
- **Rating is required at registration.** A new User must pick a rating to create an account —
  `self_rating` is a required field at signup, so new accounts are never null.
- **A home-load gate backfills existing accounts.** After the migration, every pre-existing User
  has a null `self_rating`. Such a User is shown a **blocking interstitial on home/dashboard load**
  and must pick a rating to proceed past home. The gate is deliberately **narrow**: it fires *only*
  on the home surface. Every deep link — a session join link, a session view — **passes through
  unblocked**, so a returning user is never trapped at the court. New signups never see the gate
  (they can't be null); it exists purely to backfill the legacy user base.
- **Missing data:** unrated players are treated as the neutral median (**3**). An all-unrated
  lobby therefore has zero rating gaps everywhere, the rating term vanishes, and behaviour is
  identical to today. The feature is **self-cancelling** without data, so it is **always on** —
  no session toggle.
- **Visibility:** the 1–5 number is visible to everyone; the admin edits it in the lobby.
- **1 court:** no effect — with a single court the four active players form two pairs that must
  face each other, so there is no assignment freedom to exploit.
- **Mexicano:** ratings seed **round 1 only** (replacing the current random shuffle of initial
  standings); rounds 2+ already balance from live standings and are unchanged.

Gender / mixed-doubles composition was considered alongside this and **deferred entirely** to its
own design pass — it is a partner-step constraint (it *does* eliminate same-gender partnerships)
and needs a balanced-lobby rule, so it does not belong bundled with this change.

## Amendment (2026-07): Match-up variety, and mode-dependent priority

The original decision made rating **primary** at the court-assignment step and pairwise
co-occurrence a mere tie-breaker. That is correct for a **limited** tournament (a fixed, small
number of Rounds) where every *individual* Match should be competitive *now* — there aren't
enough Rounds for unfairness to average out.

It is the wrong priority for **Unlimited Rounds** Americano. There, once every partnership has
occurred (partner-repeat counts saturate to a flat objective), the court-assignment search
returns the deterministic first-found grouping every cycle — so the *identical* Match-ups recur.
Rating being primary makes this worse: with a set field the balanced groupings are a small set,
so the same balanced Match-ups repeat. Players' actual complaint is not "unbalanced" but "we keep
facing the same pair."

**Insight:** maximal Match-up variety is itself a fairness mechanism. Because the partner step is
rating-agnostic, teams already vary in strength Round to Round; adding opponent variety makes each
Player face a representative spread of opponents, so skill imbalance **averages out cumulatively**
over a long session — without balancing any single Match. Variety carries fairness precisely in
the mode (Unlimited) that has enough Rounds for averaging to work; per-Match rating balance carries
it in the mode (limited) that does not.

**Amended decision:** the court-assignment step scores the same two signals — **Match-up variety**
and **rating gap** — in a **lexicographic key whose priority depends on the mode**:

- **Unlimited:** `matchupCount → matchupRecency → ratingGap → random`
  (variety first: use every distinct Match-up before any repeat; when forced to repeat, reuse the
  stalest; only then prefer the more balanced; random breaks true ties so Rounds aren't identical.)
- **Limited:** `ratingGap → matchupCount → matchupRecency → random`
  (rating first, preserving this ADR's original intent; the old pairwise `coOccurrence` tie-breaker
  is to be **replaced** by the strictly-better Match-up-tuple signal — no behaviour regression, better
  variety among equally-balanced groupings.)

The pairwise `coOccurrence` term is retired in favour of the four-Player **Match-up** tuple
(`matchupKey`), which targets the exact opposition Players care about (`A+B vs C+D`) rather than
individual co-occurrence (`A` shares a court with `C`). Rating stays **self-cancelling**: an
unrated/all-median field has zero rating gaps, so both orderings collapse to variety-only.

**Scope of the amendment:** Unlimited Americano only for the priority flip (issue #271); Limited
Americano gets just the tie-breaker upgrade (issue #272, not yet implemented — the limited path
still scores pairwise `coOccurrence` until then); Mexicano is untouched (its teams are derived from
Standings, not chosen by a court-assignment search, so there is no grouping freedom to vary). The
`random` final key only fires on genuine ties, so existing deterministic rating tests are
unaffected; new variety behaviour is covered by property-based tests.

The unlimited guarantee is *local*: because the partner step fixes a round's pairs before court
assignment, "use every distinct Match-up before repeating" holds over the groupings reachable from
those pairs, not over all conceivable Match-ups.

### Amendment (2026-07): Bench selection is Match-up-aware on low-freedom courts (#274)

The court-assignment step can only vary *who faces whom* when there is more than one court. On **1
court** the four active Players form the single possible Match-up, so the variety objective above is
inert — a recurring partnership kept meeting the same opponent even in unlimited mode. The lever
there is not the court step but **which Players sit out**: rotating the Bench changes the active
pool and therefore the opponent a recurring pair faces.

**Amended decision:** in unlimited mode, **Bench selection becomes Match-up-aware**. Bench
*fairness stays a hard constraint* — the must-play rule (benched last Round ⇒ plays this Round) and
even Bench counts are enforced first, exactly as before; nobody ever sits out more to gain variety.
Among the **equally-fair** Bench choices, the scheduler picks the one whose resulting Round is
freshest, ranking whole Rounds lexicographically `partnerRepeats → matchupCount → matchupRecency →
ratingGap`. Partner repeats stay strictly above variety, so the everyone-partners-everyone invariant
is untouched.

To make this effective on 1 court, low-freedom fields (**≤2 courts**) also drop the two-step
partner-then-court search in favour of a **joint search over partners and opponents together** —
otherwise the partner step would fix a repeat pairing before the freedomless court step could react.
Larger fields keep the cheaper two-step search, where regrouping pairs across courts already
supplies the variety and the joint search would be too costly.

**Consequence:** on a single court the scheduler now provably returns the lexicographic-optimum
Round over all fair Bench choices, so a recurring pair gets a fresh opponent whenever fairness and
partner-repeat priority allow one. Any remaining repeat is forced by those higher-priority
constraints (a small, unavoidable warm-up effect with very few Players), not a scheduling miss.
Still unlimited Americano only; Mexicano and the limited/upfront path are untouched.

## Consequences

- **Rating is a lobby concern; finalize before Start.** Fixed-mode Americano pre-generates all
  rounds at `Start()` (`GenerateRounds`), so editing a rating mid-session does not rewrite
  already-generated rounds. In unlimited mode, only round 1 is generated at start and later rounds
  come from `GenerateNextRound` at advance time, so mid-session edits *do* influence future rounds
  there. This mirrors how roster changes already behave.
- The court-assignment search space is unchanged (same backtracking over pair groupings), so the
  performance characteristics are unaffected; only the per-grouping score changes.
- Because rating never touches the partner step, the "everyone partners everyone" guarantee and
  all existing partner-repeat/bench invariants hold exactly as before.
- Self-set ratings can be wrong or gamed; the admin override in the lobby is the intended
  correction path, which is why the admin must be able to see and edit them (rules out a fully
  hidden rating).
