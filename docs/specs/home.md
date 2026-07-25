# Home / landing — redesign spec

**Ticket:** [Redesign spec: Home / landing](https://github.com/Fabianthorsen/OpenPadel/issues/176) (#176)
**Map:** [UI improvement](https://github.com/Fabianthorsen/OpenPadel/issues/167) (#167) · **North-star:** [visual language](../../CONTEXT.md#visual-language) · **Rubric:** [redesign-rubric.md](redesign-rubric.md) · **Audit:** [ui-audit.md](../research/ui-audit.md)
**Surfaces:** `web/src/routes/+page.svelte` (**remove** its inline setup step), `web/src/lib/components/Footer.svelte`.

> Register: **working screen → calm.** Decision (#176): **`/` is the signed-out landing**; logged-in users are **redirected to `/profile`** (their home). The inline setup form is **removed** — creation is the minimal drawer (#175). ⚠️ **Cross-cutting:** this makes **Profile (#178) the logged-in home/dashboard** (Start session drawer + rejoin + upcoming + history + contacts + invites + settings).

---

## Layout (signed-out `/`)

Centered, `max-w-sm`, single column, `pt-safe-page`:
1. **Brand** — `OpenPadel` wordmark (H1, `text-primary`) + tagline (`text-text-secondary`).
2. **Rejoin pill** (if a last session is active — also serves a returning guest): a status dot + "Rejoin {name}" → the session. Top of the actions.
3. **Primary — Sign in →** (`Button` `cta`) → `/auth`.
4. **"No account? Create one"** → `/auth?register=1`.
5. **"or" divider** (shared divider component).
6. **Join by code** — `Input` (4-char, uppercased) + **Join** `Button` → `/s/{CODE}`.
7. **Footer**.

---

## Removed

- **The entire setup step** (courts/points/name/schedule/organiser card/info-note/range-slider). Creation is the minimal drawer (#175), opened from Profile.
- **The dead logged-in pill** — logged-in users are redirected to `/profile`, so they never render `/`.
- **The `ℹ` emoji** (was in the info-note) and the raw `<input type="range">` scheduler.

---

## Behaviour

- **Redirect** — `auth.ready && auth.user` → `goto('/profile')`, forwarding `?create=1` (Profile opens the creation drawer) and `?notfound=1`.
- **Toasts** — `?deleted=1` (account deleted), `?notfound=1` (session not found).
- **Join by code** → `/s/{CODE}` (uppercased, 4-char).
- **Rejoin** — from `localStorage.last_session_id` + admin token, shown only if the session is `lobby`/`playing` (unchanged logic).

---

## Components & tokens

- `Button` `cta` (Sign in) + `Button` secondary (Join); `Input` for the join code (use the primitive's styling, **drop** the `border-0` override); a **shared "or" divider** component (kills the duplicated `bg-border h-px flex-1` pattern); no emoji.
- Kill: the raw `bg-primary … text-white` CTA `<a>`, the `Input` `border-0` override, the hand-rolled initials circle (dead pill).

---

## States

- **Signed-out** — the landing (above).
- **Signed-in** — redirected to `/profile` (no `/` render).
- **Auth loading** (`!auth.ready`) — minimal placeholder (shared `<Spinner>` / brand only), not a flash of the signed-out actions.

---

## A11y & responsiveness

- Real `<label>` on the join-code input; Sign in / Join ≥48px; `focus-visible`.
- Safe-area; single column ≤480px; copy via `$_()` (already localised).

---

## Before → after

| Before (shipped) | After (this spec) |
|---|---|
| Two steps: signed-out landing **+** a full inline setup form | Signed-out landing only; setup removed (creation = #175 drawer) |
| Raw CTA `<a>`, `Input` `border-0` override, dead logged-in pill | `Button` `cta`, `Input` primitive, pill removed |
| Duplicated `bg-border h-px flex-1` "or" divider | Shared divider component |

---

## Out of scope

- **The logged-in home/dashboard** → Profile (#178).
- **The creation drawer** → Session creation (#175).
- **Auth screens** (`/auth`, `/forgot`, `/reset`) → Auth (#177).
