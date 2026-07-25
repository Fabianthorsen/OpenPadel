# Lobby & join — redesign spec

**Ticket:** [Redesign spec: Lobby & join-via-link](https://github.com/Fabianthorsen/OpenPadel/issues/173) (#173)
**Map:** [UI improvement](https://github.com/Fabianthorsen/OpenPadel/issues/167) (#167) · **North-star:** [visual language](../../CONTEXT.md#visual-language) · **Rubric:** [redesign-rubric.md](redesign-rubric.md) · **Audit:** [ui-audit.md](../research/ui-audit.md)
**Surfaces:** `web/src/lib/components/Lobby.svelte`, the lobby + join states of `web/src/routes/s/[id]/+page.svelte`
**Supersedes:** `docs/specs/invite-screen.md` (merged here).

> Register: **working screen → calm.** Two states on the same route: **(A) pre-join invitation** (link opened, not a member) and **(B) lobby** (joined). Uses the shared session shell (header + back; **no** bottom nav / Standings peek — there are no standings pre-play). Guests and registered players are both first-class (CLAUDE.md).

---

## The decisions (grilling)

1. **Pre-join = a welcoming invitation** — host, session, who's already in, then join (not a bare form).
2. **Lobby config = compact summary + Edit drawer**, backed by **one shared `SessionConfig` component**. ⚠️ **Cross-cutting consolidation** (kills the audit's "three create UIs"): **Session creation (#175)** builds/uses the same `SessionConfig`; **Home (#176)** drops its inline setup form and routes to creation. The lobby just shows a summary + Edit.

---

## State A — Pre-join invitation

For someone who opened a share link and isn't a member yet.

Layout (single column, calm surfaces, `primary` accents only on the detail chips):
1. Back `×` (top-right).
2. **Host** — host avatar (`lg`) + `"{Host} invited you to play"` (`text-text-secondary`) + **session name** (H1) + detail chips `Americano · 2 courts · 24 pts` (mode neutral, courts/points in `primary`, dot separators). If scheduled, show date/time.
3. **Players preview** — `SectionLabel "Players (n)"` + a **stacked avatar group** (overlapping, up to 4 + "+N") + name summary ("Ana, Bruno, Carl +2"). Empty → "No one's joined yet — be first!"
4. **Join** —
   - Logged-in: a profile row (avatar + name/email) + full-width **Join session →** (`cta`).
   - Guest / not logged in: `[Your name…]` `Input` + **Join →** (`cta`); an "or sign in" link below.
5. Muted footer ("Powered by OpenPadel").

Edge states: session already `playing` → "Session in progress" (no join); `done` → "Session ended"; full (`players === courts×4`, Mexicano) → "Session full".

---

## State B — Lobby (joined)

Layout, top → bottom:
1. **Header** — "Waiting to start" (or the scheduled date/time) + **session name**.
2. **Config summary** — one line: `Americano · 2 courts · 24 pts` + **Edit** (admin only) → opens the shared **`SessionConfig`** drawer (mode / courts / points / rounds / schedule). Non-admins see the summary read-only.
3. **Share card** — `SectionLabel "Join code"` + the 4-char code (tokenised tiles) + the link + **Share / Copy** (native OS share sheet).
4. **Invite** (admin) — a search field (debounced, registered Users) → **Invite**; and **add "{name}" as guest** (dashed affordance once text is typed, no match required). Both paths stay — guests and registered are first-class.
5. **Players list** — `SectionLabel "Players (n)"`; rows: avatar + name + host crown + "you" + pending-invite (clock) + remove `×` (admin). Live via `session_updated` SSE.
6. **Start** (admin) — full-width **Start →** (`cta`), gated on `can_start` (server) / fallback `players === courts×4` (Mexicano exact) or `≥` (Americano). Disabled → clear reason: "Needs 8 players (have 6)" or the mode hint. Non-admin joined → calm "Waiting for the admin to start…".
7. **Cancel session** (admin) — quiet destructive text affordance → `ConfirmDialog`.
8. Dev-only **Seed players** (`import.meta.env.DEV` gate).

---

## Components & tokens

- Reuse `Button` (`cta`), `Input`, `PillToggleGroup`, `Stepper`, `Calendar`, `Avatar`, `SectionLabel`, `Drawer`, `ConfirmDialog`, `Badge` (host/pending chips).
- **Shared additions (anti-divergence rule):**
  - **`SessionConfig`** component (mode/courts/points/rounds/schedule) — shared with Session creation (#175). Whichever ticket implements first builds it; both consume it. Add to the rubric's standing-fixes table.
  - **`AvatarGroup`** (stacked overlapping avatars + "+N") — for the players preview; reusable on other screens.
- **Kill the audit's debt here:** `Button`/`Input` class-overrides → `cta`/variants; `bg-red-50`/`text-red-900` warnings → `--color-warning` / `text-destructive`; the ~7 inline section-labels → `SectionLabel`; the duplicated selectable-card pattern → shared; join-code tiles tokenised; the two near-duplicate join forms collapse into the single State-A join.

---

## Interaction & flow

- **Config patches immediately** on change over SSE (keep the "no Save button" decision — joining players see changes live); the Edit drawer's "Done" just closes.
- **Mexicano** — selecting it bumps `courts` to 2 and disables the 1-court option (CLAUDE.md invariant).
- **Live** — roster + pending invites refresh on `session_updated` SSE (no polling; page-level 30s fallback covers buffering proxies).
- **Join** — optimistic add to the roster.

---

## States

- **Pre-join:** loading · invitation (open) · scheduled · already-started · ended · full.
- **Lobby:** admin (summary+Edit, invite, roster, Start) · non-admin (waiting) · cancelling (spinner) · loading (shared `<Spinner>`).

---

## A11y & responsiveness

- Real `<label>` on the join field + config controls; `aria-invalid` + error text on validation.
- `aria-live="polite"` announces players joining.
- Tap targets ≥48px; `focus-visible`; focus-trap in the Edit drawer / dialogs.
- Safe-area; single column ≤480px; all copy via `$_()` (kill hardcoded "Invite", "as guest", etc.).

---

## Before → after

| Before (shipped) | After (this spec) |
|---|---|
| Bare pre-join heading + form | Welcoming invitation (host, players preview, join) |
| Full config inline-editable in the lobby | Compact summary + Edit drawer → shared `SessionConfig` |
| Two near-duplicate join forms | One join path (State A) |
| `Button`/`Input` class-overridden; `bg-red-50`; 7× inline labels | `cta`/variants; `--color-warning`/`destructive`; `SectionLabel` |

---

## Out of scope

- The creation entry-point + first-run config UX → Session creation (#175) (which builds the shared `SessionConfig`).
- Played-state screens (round/leaderboard); dark mode; ranking logic.

## Cross-cutting note (for the map)

The config consolidation decided here sets the model for **Session creation (#175)** (builds `SessionConfig` + the create entry) and **Home (#176)** (drop the inline setup form; route "Start session" → creation). Flagged on the map.
