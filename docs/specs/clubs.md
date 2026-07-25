# Spec: Clubs

## Goal

Clubs are persistent groups of registered **Users** who play together (a workplace, a friend group,
a venue). A Club eases inviting people to games, keeps a cross-session **leaderboard**, and lets
members organise **club events**. This is a **proposed** feature spec — it assembles the decisions
made across the *Clubs — wayfinder map* (#113); no code is shipped yet.

**Decision record this spec is built from:**

- Foundation — membership, roles, joining: #114
- Leaderboard metric: #116 → [`docs/adr/0005-club-leaderboard-metric.md`](../adr/0005-club-leaderboard-metric.md) (folds in the #115 ranking-metric research)
- Club-scoped invites: #117
- Club events: #118
- UI shape (prototype): #119

**Load-bearing constraint (from `CLAUDE.md`):** `AdminToken` is the only authorization gate for
*Session* actions. Club roles are a **separate, account-scoped** gate for *club-management* actions
only, and never grant power over any Session — including a Club's own events.

---

## Domain (additions to `CONTEXT.md`)

**Club** — a named, persistent group of registered Users who play together. Users-only: Guests may
play in a Club's Sessions but are not members and do not accrue on the Club leaderboard. A User may
belong to many Clubs. _Avoid_: team, group, league.

**Club Admin** — a Club member with management rights: edit/delete the Club, manage the roster
(remove, promote/demote), manage the join link. Distinct from **Admin** (the session `AdminToken`
holder); a Club Admin has no special power over any Session, including the Club's own events.
_Avoid_: owner, moderator, game master.

**Club event** — a Session owned by a Club (`club_id`), created by any Member. Administered like any
Session via its `AdminToken`. _Avoid_: tournament.

**Club invite** — a pending request for a specific User to join a Club (`pending | accepted |
declined`). Distinct from **Invite** (which targets a Session) and from the Club join link (which
needs no pending state).

---

## Data model

One change to an existing table; three new tables. All times `TEXT` RFC3339, IDs via `newID()`
(4-char base32) except where noted, consistent with `internal/store/`.

```sql
-- existing table, one added column
ALTER TABLE sessions ADD COLUMN club_id TEXT NULL REFERENCES clubs(id);
-- club_id NULL  → an ordinary Session
-- club_id set   → a club event

CREATE TABLE clubs (
    id           TEXT PRIMARY KEY,            -- short base32, used in /c/:id
    name         TEXT NOT NULL,
    description  TEXT NOT NULL DEFAULT '',
    avatar_icon  TEXT NOT NULL DEFAULT '',    -- same avatar model as users/players
    avatar_color TEXT NOT NULL DEFAULT '',
    join_code    TEXT NOT NULL UNIQUE,        -- rotatable opaque token; gates the join link
    created_by   TEXT NOT NULL REFERENCES users(id),
    created_at   TEXT NOT NULL
);

CREATE TABLE club_members (
    club_id   TEXT NOT NULL REFERENCES clubs(id) ON DELETE CASCADE,
    user_id   TEXT NOT NULL REFERENCES users(id),
    role      TEXT NOT NULL DEFAULT 'member',  -- 'admin' | 'member'
    joined_at TEXT NOT NULL,
    PRIMARY KEY (club_id, user_id)
);

CREATE TABLE club_invites (
    id         TEXT PRIMARY KEY,
    club_id    TEXT NOT NULL REFERENCES clubs(id) ON DELETE CASCADE,
    inviter_id TEXT NOT NULL REFERENCES users(id),
    invitee_id TEXT NOT NULL REFERENCES users(id),
    status     TEXT NOT NULL DEFAULT 'pending', -- pending | accepted | declined
    created_at TEXT NOT NULL,
    UNIQUE (club_id, invitee_id)
);
```

Shipped as one goose migration (`internal/store/migrations/`). The creator is inserted into
`club_members` with `role='admin'` in the same transaction as the `clubs` row.

---

## API surface

New handlers in `internal/api/clubs.go`; membership-invite handlers alongside. All require auth
(`requireAuth`) except the join-link *preview*. Club-management endpoints additionally check the
caller's `club_members.role`.

```
# Clubs
POST   /api/clubs                     create; caller becomes first Club Admin
GET    /api/clubs                     my clubs (roster counts, my role)
GET    /api/clubs/:id                 club detail — roster, is_admin, my membership
PATCH  /api/clubs/:id                 edit name/description/avatar          (Club Admin)
DELETE /api/clubs/:id                 delete club                            (Club Admin)
GET    /api/clubs/:id/leaderboard     computed leaderboard (see below)
GET    /api/clubs/:id/events          upcoming club events (lobby|playing)

# Membership
POST   /api/clubs/join                join via { join_code } from the link → Member
DELETE /api/clubs/:id/members/:userID leave (self) or remove (Club Admin)
PATCH  /api/clubs/:id/members/:userID promote/demote { role }               (Club Admin)
POST   /api/clubs/:id/join-code/rotate revoke + reissue the join link        (Club Admin)

# Club membership invites (distinct from Session invites)
POST   /api/clubs/:id/invites         invite a User { to_user_id }           (any Member)
GET    /api/clubs/invites             my pending club invites
PUT    /api/clubs/invites/:id/accept  accept → become a Member
PUT    /api/clubs/invites/:id/decline

# Club events & session/event invites
POST   /api/sessions                  existing; now accepts optional { club_id }
POST   /api/sessions/:id/invites/club bulk fan-out: { club_id } → one Invite per member  (#117)
```

Everything else in the invite flow is **reused unchanged**: `POST /api/sessions/:id/invites`
(targeted), `GET /api/invites`, `PUT /api/invites/:id/accept|decline`. Accepting any Session invite
creates a `Player` as today.

---

## Leaderboard computation

**Metric (ADR 0005):** average normalized point-margin per game. For each fully-scored `Match` a
member played in a qualifying club event, `margin = (memberTeamPoints − opponentPoints) / points`
∈ `[−1, +1]`; the member's value is the **mean** of `margin` over their qualifying games.

- **Counts:** only `Session`s with this `club_id`, only scored matches (`score_a` not null), only
  within the **rolling window** (default 90 days, by session date). Members map from a match's
  player slots → `players.user_id` → `club_members`; Guests (`user_id NULL`) never produce a row
  but do count as the opponents/partners in a member's games.
- **Qualification:** ranked only with **≥ `MinGames`** qualifying games (default 5). Below → a
  **provisional / not-yet-ranked** list with an "N more to rank" count.
- **Tie-break:** higher average → more qualifying games → earlier first-qualifying date → name.
- **Computation:** recompute on read (the mean is order-independent) — a single SQLite aggregate
  joining `matches → rounds → sessions → players → club_members`. No materialised state.
- **Deferred (ADR 0005):** a Glicko-2 skill layer; retro-computable from stored matches when wanted.

`MinGames` and the window length are tunable constants (candidate: `internal/domain`).

---

## Screens

Visual language per `CONTEXT.md § Visual Language`. UI shape chosen in #119 — prototype:
<https://claude.ai/code/artifact/a4fca59f-0974-4874-88ef-7dd13b7799d1>.

### Club home — "action dashboard" (Variant C)

```
┌──────────────────────────────────────────┐
│  ‹   Bouvet Padel                    ⚙   │  ← name + avatar; ⚙ → admin sheet (admins)
│      18 members                          │
├──────────────────────────────────────────┤
│  NEXT UP                                 │
│  ┌────────────────────────────────────┐ │
│  │ Tuesday · 18:00          [6/8 in]  │ │  ← next event = hero (green border)
│  │ Americano · 2 courts · Ekeberg     │ │
│  │ (facepile of who's in)             │ │
│  │ [ Join ]            [ Share ]      │ │  ← one-tap join
│  └────────────────────────────────────┘ │
│                                          │
│  THIS SEASON'S TOP        See full board→│
│  1  [av] Fabian   14   +0.48             │  ← compact top-3, podium colours
│  2  [av] Ingrid   12   +0.41             │
│  3  [av] Ola      16   +0.33             │
│                                          │
│  MEMBERS                      [ Invite ] │
│  (facepile)  +13                         │
└──────────────────────────────────────────┘
```

### Full leaderboard page (via "See full board")

```
[Leaderboard]                    [ Last 90 days ▾ ]
Average point margin per game · 5+ games to rank
  #  MEMBER            GP   AVG ±
  1  [av] Fabian       14   +0.48    ← podium rows (green / green / #A8C5B0)
  2  [av] Ingrid       12   +0.41
  3  [av] Ola          16   +0.33
  4  [av] Kari          9   +0.27    ← alternating rows below
  ...
  ┌ Not yet ranked (3) ───────────────┐
  │ [av] Sofie      2 more   ▓▓▓░░     │  ← provisional, progress to MinGames
  │ [av] Erik       3 more   ▓▓░░░     │
  └───────────────────────────────────┘
```

### Supporting flows (accepted as mocked, #119)

- **Create club event** — the existing session-create flow (`CreateDrawer`) with a club banner and
  `club_id` set; any Member; creator holds the `AdminToken` as today.
- **Join a club** — a direct **Club invite** card (Join / Decline) *and* a distinct join link
  `openpadel.app/c/join/:join_code`. Kept separate from the session join link.
- **Bulk "invite my whole club"** — an "Invite all" row (fan-out) above the existing targeted
  contact/search invites, on any session's invite surface.
- **Club Admin sheet** — edit name/photo, manage join link, promote/remove members, delete club;
  admins only.

## Behavior

- **A club event is a Session** (`club_id` set) — the entire `lobby → playing → done` lifecycle,
  round generation, scoring, live scores, and SSE are reused unchanged.
- **Discovery is pull + notify** — club events appear on the club page (`GET /clubs/:id/events`);
  on creation, members are notified via the existing web-push + user-SSE fan-out over
  `club_members` (a new `club_event_created` event type). The global home feed
  (`GetUpcomingTournaments`) is **not** modified.
- **Attendance = the player list** — no separate RSVP. A member joins the event's lobby (becomes a
  `Player` linked to their `user_id`) ahead of time; that *is* their commitment. Mexicano's exact
  `courts×4` and Americano's flexibility are the existing Session rules.
- **Guests** join a club event via the public join link (by name); they play but never accrue.
- **Leaderboard is live** — recomputed on read; a `round_updated` on a club event's session
  invalidates any cached club board (or simply always recompute — it's cheap).
- **Club membership is Users-only** — every membership path requires a logged-in User.

## Key Design Decisions

| Decision | Why | Source |
|---|---|---|
| A club event is a `Session` + `club_id`, not a new object | Reuses the whole lifecycle/scoring/SSE; keeps one code path | #118 |
| Club Admin never bridges into session authz | Preserves the "`AdminToken` is the only gate" invariant; club events run exactly like any Session | #114, `CLAUDE.md` |
| Users-only membership | A cross-session leaderboard needs stable identity, which only Users have; avoids a guest identity-stitching sub-project | map destination, #114 |
| Single explainable metric (avg point margin), not Elo/Glicko now | Casual/social audience — explainability first; Glicko-2 deferred and retro-computable | ADR 0005 |
| Only club events count toward the board | No multi-club attribution ambiguity; a member's pickup games never leak in | #116 |
| Reuse `Invite` unchanged; bulk fan-out = N rows | "Don't fork the model"; one new endpoint instead of a new invite type | #117 |
| Discovery is auto-surface + push, not per-member invites | Members shouldn't need an invite to their own club's events | #117, #118 |
| No RSVP — joining the lobby is attendance | One "am I in?" signal; reuses `Player`; fits the app's lobby model | #118 |
| Distinct club join link (`/c/join/:code`), separate from session joins | Users must never confuse "join this game" with "join this club" | #114 |
| Club home foregrounds the next event, board is a glance | The point is getting a game together; ranking is secondary for a social club | #119 |

## Open Questions

- [ ] `Session` currently has no explicit "date" beyond `scheduled_at` / `created_at`; confirm which
      the 90-day leaderboard window and the club-events ordering key off (proposal:
      `COALESCE(scheduled_at, created_at)`).
- [ ] Can a member **leave** a Club freely, and what happens to a Club whose last Admin leaves
      (auto-promote the oldest member, or block the last Admin from leaving)?
- [ ] Should `MinGames` / the window length be per-Club settings or global constants? (Spec assumes
      global constants for v1.)
- [ ] Does deleting a Club also unset `club_id` on its past sessions (keep the games, lose the
      attribution) or is delete soft? (Proposal: `ON DELETE` unsets `sessions.club_id`; games and
      history survive as ordinary sessions.)
