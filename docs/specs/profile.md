# Profile — redesign spec (dashboard + settings)

**Ticket:** [Redesign spec: Profile](https://github.com/Fabianthorsen/OpenPadel/issues/178) (#178)
**Map:** [UI improvement](https://github.com/Fabianthorsen/OpenPadel/issues/167) (#167) · **Rubric:** [redesign-rubric.md](redesign-rubric.md) · **Audit:** [ui-audit.md](../research/ui-audit.md)
**Surfaces:** `web/src/routes/profile/+page.svelte` (→ the **dashboard**), a **new** `web/src/routes/profile/settings/+page.svelte`, `CreateDrawer`, `LocaleSwitcher`.

> Register: **working screen → calm.** Profile is the **logged-in home/dashboard** (per #176). Decision (#178): **split** it into the dashboard (`/profile`) + an account **Settings subpage** (`/profile/settings`). Auth-gated (→ `/auth` if no token). This screen carried the most audit debt — all of it is cleaned here.

---

## The decisions

1. **Split: `/profile` = dashboard; `/profile/settings` = account/settings.**
2. **All debt cleaned:** one modal pattern (`ConfirmDialog` for confirms, `Dialog` for the picker — kill the hand-rolled `fixed inset-0` delete modal); **`Switch`** for the push toggle; **`Avatar`** for contact rows (kill hand-rolled initials); tokens (**live → `--color-primary`**, losses → `text-destructive`, drop `emerald-500`/`#c0392b`); **`SectionLabel`** for the ~10 inline labels; a **shared `JoinCodeInput`** (consistent with Home #176).
3. **A reusable `Section` component.** The dashboard hand-rolls the same collapsible block ~5× (Contacts / Stats / History / Upcoming — a `Collapsible.Root` + inline label + chevron each). Collapse them into one shared **`Section`** primitive (see below) — reusable well beyond Profile.

---

## `/profile` — the dashboard

Single column, `max-w-[480px]`, `pt-safe-page`:
1. **Header** — `Avatar` + display name + "Member since {month year}"; a **Settings** affordance (gear, top-right) → `/profile/settings`.
2. **Start session** — full-width `Button` `cta` → the `CreateDrawer` (#175). Plus **join-by-code** via the shared **`JoinCodeInput`** (same component as Home #176).
3. **Invites** (if any) — `SectionLabel "Invites"` + rows: session name + from-name, **Accept** / **Decline**. Live via the `invite_received` SSE event.
4. **Upcoming / Live** — rows linking to each session; the **Live** badge uses `--color-primary` (pulsing dot / `primary-muted` tint) — **not** `emerald-500`.
5. **Contacts** — collapsible: a search `Input` (with a search icon) + saved list; **`Avatar`** on every row (not hand-rolled initials); add/remove; delete via `ConfirmDialog`.
6. **Stats** (Americano) — collapsible: the 2×2 grid (tournaments / win-rate% / games / V·U·T); losses → `text-destructive`.
7. **History** — collapsible: rows with a rank-ordinal badge (rank 1 in `primary`).

Contacts / Stats / History / Upcoming are each a **`Section`** (below), default-open per data presence (as today), with long lists (contacts, history) scrolling inside the section rather than stretching the page.

### Reusable `Section` component (shared)

One primitive replaces the ~5 hand-rolled collapsible blocks:

- **`title`** — rendered via `SectionLabel`.
- **`collapsible`** (default `true`) + **`open`** (`$bindable`, configurable default) — a chevron toggle; `collapsible={false}` renders a plain titled block (no chevron).
- **`maxHeight` / `scroll`** — when set, the body scrolls internally (`overflow-y: auto`) so long lists (contacts, history) don't run the page long; otherwise it grows to fit.
- Optional **trailing slot** in the header (e.g. a count, a search toggle).

Composes `Collapsible` + `SectionLabel` + chevron; lives in `web/src/lib/components/ui/` so Lobby, Settings, and future screens reuse it (anti-divergence rule).

---

## `/profile/settings` — account & settings (new subpage)

1. **Back** → `/profile`.
2. **Identity** — the avatar picker (icon grid + "use initials"; **+ colour swatches** per the avatar-system spec — a modest completion of the half-built picker, optional) + display-name edit; Save via `api.auth.updateProfile`.
3. **Notifications** — push toggle → **`Switch`** (keep the blocked / SW-timeout messaging).
4. **Install app** (PWA) — the iOS instructions / Android prompt (existing logic).
5. **Language** — `LocaleSwitcher`.
6. **Account** — **Sign out** (`Button` secondary); **Delete account** (destructive → `ConfirmDialog`, replacing the hand-rolled modal).

---

## Components & tokens

- Reuse `Avatar`, `Button` (`cta`/secondary), `Switch`, `ConfirmDialog`, `Dialog` (picker), `Collapsible`, `Input`, `SectionLabel`, `Footer`, `LocaleSwitcher`, `CreateDrawer`.
- **Shared additions (anti-divergence rule):** the reusable **`Section`** (titled block; optional collapse w/ default-open/closed; optional internal scroll — replaces the ~5 hand-rolled `Collapsible` blocks); **`JoinCodeInput`** (shared with Home #176 — recommend the OTP-boxes + paste UX); a shared list-row / `Card` for the invite/upcoming/history/contact rows.
- **Kill:** hand-rolled push toggle → `Switch`; hand-rolled initials → `Avatar`; hand-rolled delete-account modal → `ConfirmDialog`; `text-[#c0392b]` → `text-destructive`; `emerald-500` (live) → `--color-primary`; the ~10 inline `text-[11px]…uppercase` labels → `SectionLabel`.

---

## Live indicator (decision, recorded in DESIGN.md)

A "live/playing" session is signalled with **`--color-primary`** — a pulsing dot + `primary-muted` tint — consistent with the active-round live-court tint (#170) and the leaderboard live dot (#172). No `emerald-*`, no new token.

---

## Interaction & flow

- **Auth gate:** no token → `/auth`. `?create=1` → open the `CreateDrawer`. `?notfound=1` → toast.
- **Invites** live via `invite_received` SSE; accept → `/s/:id`; decline → remove.
- **Contacts:** debounced search (300ms, ≥2 chars); add/remove; delete via `ConfirmDialog`.
- **Push:** subscribe/unsubscribe with blocked/timeout messaging (keep).
- **Delete account** → `ConfirmDialog` → `deleteAccount` → `/?deleted=1`.
- **Start session** → `CreateDrawer` (#175).

---

## States

- **Dashboard:** loading (shared `<Spinner>`, not the bare spinner div) · loaded · empty (friendly empties for no invites/upcoming/history/contacts) · signed-out (redirect).
- **Settings:** loaded · saving-avatar · push-toggling · deleting.

---

## A11y & responsiveness

- Real `<label>` on search / name / join-code; `Switch` labelled; `ConfirmDialog` focus-trap; picker buttons labelled (already).
- `aria-live="polite"` for incoming invites.
- Tap targets ≥48px; `focus-visible`; safe-area; single column ≤480px.
- All copy via `$_()` (kill hardcoded "Invites", "Accept", "Search results", "Your contacts", "Live", "Choose avatar", "Save", "Delete Contact?", "Americano", "Member since", etc.).

---

## Before → after

| Before (shipped) | After (this spec) |
|---|---|
| One crammed scroll (dashboard + settings mixed) | `/profile` dashboard + `/profile/settings` subpage |
| 3 modal patterns (ConfirmDialog, Dialog, hand-rolled) | `ConfirmDialog` (confirms) + `Dialog` (picker) |
| Hand-rolled push toggle; hand-rolled initials | `Switch`; `Avatar` |
| `emerald-500` live; `text-[#c0392b]` losses | `--color-primary` live; `text-destructive` |
| ~10 inline labels; 4-box OTP join-code | `SectionLabel`; shared `JoinCodeInput` |
| ~5 hand-rolled `Collapsible` blocks (inconsistent) | one reusable `Section` (collapse + internal scroll) |

---

## Out of scope

- New account features (2FA, email-change flows) — backend / new capability.
- Guest "profile" (guests have no account).
- Dark mode.
