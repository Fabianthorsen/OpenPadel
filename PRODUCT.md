# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Users

The primary user is the **casual social organizer** — the person in a group of
friends who sets up a padel session courtside, phone in one hand, sun overhead.
No club, no formality, often organizing on the spot. When a design trade-off
arises, their need for speed and ease wins.

Secondary participants:

- **Guests** — anyone who joins a Session by name only, no account. The largest
  group by headcount and the on-ramp to everything else.
- **Registered Users** — organizers and regulars who want saved identity, career
  stats, contacts, and Clubs. A superset of the casual organizer over time.
- **Club organizers** — Users running a persistent group's recurring games.
  Served, but not the tie-breaking audience.

## Product Purpose

OpenPadel organizes padel games and tracks scores courtside. It runs a Session
through its lifecycle (`lobby` → `playing` → `done`), auto-generates fair
round-by-round pairings for two Game Modes (Americano, Mexicano), and shows a
live, glanceable leaderboard. Success is the organizer getting a group from
"let's play" to a running, scored game in seconds, standing on the court, and
everyone being able to read the standings at a glance between rounds.

## Positioning

The differentiator is **zero-friction courtside play with no account required**:
a group joins a Session by a 4-character code, by name only, and starts scoring —
while registered Users still accrue durable identity, career stats, and Clubs on
top of the same flow. Guest join is a first-class, permanent path, not a trial
mode. The scoring model is deliberately margin-aware and scale-free
(**Point-share** / "Point Win %": the average per-Match share of points), so
stats stay honest across rotating partners and different point targets — a
neighboring "tournament bracket" app could not truthfully copy either the
no-account on-ramp or this per-Match point-share metric.

## Operating Context

- **Courtside, outdoors, one-handed.** Every working screen is used standing up,
  phone in one hand, sun overhead. Big tap targets, no tiny text, no fiddly
  inputs. The most frequent action — entering a score — must take ~3 taps.
- **Real-time, multi-device.** Several phones watch one Session; score and
  standings updates propagate live (SSE-first, with a 30s poll fallback for
  buffering proxies).
- **Installable PWA**, mobile-first. Push notifications fire when a Session
  starts.
- **Two registers of use:** repeated glances at the live leaderboard mid-session
  (calm, fast) and the one celebratory finale at Session complete.

## Capabilities and Constraints

Capabilities:

- Create a Session with a Game Mode, courts, and points target; join by
  4-character code or share link.
- **Americano** (rotating partners, individual scoring) and **Mexicano**
  (standings-adaptive pairings; requires ≥2 courts, no bench, player count =
  courts × 4 exactly).
- Fixed or **Unlimited Rounds** (one-way switch from fixed → unlimited via "Keep
  Playing"); automatic Bench rotation guaranteeing a benched Player plays the
  next Round.
- Live per-Match scoring and a public, no-auth Session leaderboard.
- **Ratings** — a self-assessed 1–5 skill level that biases match-ups (not
  partnerships) toward competitive games; unrated counts as the median (3).
- Registered Users: career stats segmented per Game Mode, Contacts, Invites, and
  **Clubs** (persistent Users-only groups with roster roles, club events, club
  invites, and a club leaderboard).

Constraints & authorization:

- **AdminToken is the only authorization gate** for privileged Session actions
  (start, cancel, close, score entry, kick), checked via `isAdmin()` and
  independent of login. Creator / Club-Admin identity is display metadata, never
  an auth check.
- Guests contribute to play and opponents' points but do not accrue on Club
  leaderboards (Users-only) and have no account behind them.

Terminology is governed by `CONTEXT.md` (the ubiquitous language: Session,
Player, Guest, User, Admin, Creator, Game Mode, Round, Bench, Match,
Standing/Leaderboard, Point-share, Club, and their "avoid" lists). Future work
must use those terms and avoid the listed synonyms (e.g. never "tournament",
"scoreboard", "team/group/league" for Club).

## Brand Commitments

- **Name:** OpenPadel.
- **Guest join, no account** is a permanent, deliberate simplicity — never gate
  core play behind signup.
- **Bilingual EN + NO** — all user-facing copy ships in English and Norwegian
  (`web/src/lib/i18n`).
- **Light-mode only** — an intentional stance (per `CONTEXT.md § Visual Language`), not a
  temporary state; do not introduce a dark theme without an explicit decision.
- **Single-binary, self-hostable** — the whole product ships as one Go binary
  embedding the SvelteKit build; open and self-hostable deploy is part of the
  identity.
- Voice and visual language live in `CONTEXT.md § Visual Language` ("Calm by default, bold when
  we celebrate"); treat expressive treatments as celebratory-only.

## Evidence on Hand

- `CONTEXT.md` — domain glossary / ubiquitous language (source of truth for
  terms).
- `CONTEXT.md § Visual Language` + `docs/specs/redesign-rubric.md` — visual north-star and
  the per-screen redesign rubric.
- `docs/adr/` — hard-to-reverse decisions (e.g. removing Timed Americano;
  rating balances match-ups not partnerships; Career Stats redesign).
- `docs/research/ui-audit.md` — the audit that reconciled the visual language with the
  shipped app.
- `README.md` — stack and run instructions.
- No customer testimonials, usage benchmarks, pricing, or press exist; future
  work must not fabricate them.

## Product Principles

1. **Courtside first.** Every working screen is designed for standing use, one
   hand, sunlight. Legibility and tap-target size beat density and decoration.
2. **The casual organizer's speed wins.** From "let's play" to a running scored
   game in seconds; the score-entry path stays ~3 taps.
3. **Guest-first, account-optional.** No-account play is the front door;
   registered identity is an additive reward, never a gate.
4. **Honest, scale-free stats.** Prefer margin-aware, per-Match Point-share over
   winrate so numbers stay fair across rotating partners and point targets.
5. **Calm by default, bold only to celebrate.** Restraint on working screens;
   expressive treatment reserved for the Session-complete finale.

## Accessibility & Inclusion

No formal external standard is mandated. The durable, product-specific
requirement is courtside usability: large tap targets and sunlight-legible
contrast/typography on every working screen, per the visual-language principles (`CONTEXT.md`).
