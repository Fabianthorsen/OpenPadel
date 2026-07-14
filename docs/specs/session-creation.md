# Session creation — redesign spec

**Ticket:** [Redesign spec: Session creation (CreateDrawer)](https://github.com/Fabianthorsen/OpenPadel/issues/175) (#175)
**Map:** [UI improvement](https://github.com/Fabianthorsen/OpenPadel/issues/167) (#167) · **North-star:** [DESIGN.md](../../DESIGN.md) §2 · **Rubric:** [redesign-rubric.md](redesign-rubric.md) · **Audit:** [ui-audit.md](../research/ui-audit.md) · **Drawer primitive:** [drawer-design.md](drawer-design.md)
**Surfaces:** `web/src/lib/components/CreateDrawer.svelte`; the home entry in `web/src/routes/+page.svelte` (**remove** its inline setup step).

> Register: **working screen → calm.** A bottom-sheet `Drawer`, opened from Home's "Start session". Login-gated (the creator auto-joins using their account name) — the signed-out → sign-in entry is Home's job (#176).

---

## The decision (grilling)

**Minimal creation:** pick game mode + optionally name it → Create → land in the lobby and tune everything else there (via the lobby's Edit drawer / `SessionConfig`, #173). This keeps time-to-first-action low and **consolidates the audit's "three create UIs"**: the home full-setup form is **removed**; the lobby owns full config. `SessionConfig` lives **only in the lobby** — creation does *not* use it.

---

## Layout (the create drawer)

`Drawer` (bottom sheet, `max-w-[480px]` centred on desktop):
1. **Header** — title "Set up a session" + close `×`.
2. **Game mode** — `PillToggleGroup`: Americano / Mexicano (Americano default) + a hint line that swaps per mode.
3. **Name** — optional `Input`, "Name it (optional)". (Resolves the old open question — you *can* name up front; still editable in the lobby.)
4. **Create** — full-width `Button` `cta`, loading state while creating; inline `text-destructive` error below.

Defaults applied on create (no pickers here): `courts: 2`, `points: 24`; Mexicano also `rounds_total: 7`. (Mexicano's ≥2-courts invariant is satisfied by the default.)

---

## Interaction & flow

- **Create** → `api.sessions.create({ game_mode, name, ...defaults })` → store `admin_token_<id>` → creator **auto-joins** as a Player (account `display_name`) → store `player_id_<id>` + `last_session_id` → `goto('/s/:id?token=<adminToken>')` (into the lobby). Flow unchanged; restyled.
- **Login-gated** — assumes `auth.user`. Signed-out users don't reach this drawer; Home (#176) presents sign-in first. A guest-creator path is backend capability → out of scope.
- This is the app's **single** creation entry (Home routes here; the lobby's "Edit" is a *different* affordance — editing an existing session, not creating).

---

## Components & tokens

- Reuse `Drawer`, `PillToggleGroup`, `Input`, `Button` (`cta` — replaces the raw `bg-primary … text-white` CTA string; `text-primary-foreground` not `text-white`).
- **No `SessionConfig` here** (creation is minimal). `SessionConfig` is the lobby's Edit component (#173, built by the lobby implement #182).

---

## States

- **Idle** — mode + optional name.
- **Creating** — `Button` loading; inputs locked.
- **Error** — inline `text-destructive` from `translateApiError`.

---

## A11y & responsiveness

- Labels on the mode toggle + name input; `Drawer` handles focus-trap / Esc / restore.
- `cta` ≥48px; `focus-visible`.
- Safe-area; single column ≤480px; copy via `$_()` (already localised).

---

## Before → after

| Before (shipped) | After (this spec) |
|---|---|
| Two create paths: minimal `CreateDrawer` **and** the home full-setup form | **One** minimal drawer; home setup form removed |
| No name at creation (rename in lobby) | Optional name field up front |
| Raw `bg-primary … text-white` CTA `<button>` | `Button` `cta` variant |

---

## Out of scope

- **Full config at creation** (chose minimal); **guest-creator** (backend capability).
- The **Home entry + signed-out sign-in gate** → Home (#176).
- The **lobby Edit / `SessionConfig`** → Lobby & join (#173).
