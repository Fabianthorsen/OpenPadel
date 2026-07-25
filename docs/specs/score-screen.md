# Score entry — redesign spec

**Ticket:** [Redesign spec: Score entry](https://github.com/Fabianthorsen/OpenPadel/issues/171) (#171)
**Map:** [UI improvement](https://github.com/Fabianthorsen/OpenPadel/issues/167) (#167) · **North-star:** [visual language](../../CONTEXT.md#visual-language) · **Rubric:** [redesign-rubric.md](redesign-rubric.md) · **Audit:** [ui-audit.md](../research/ui-audit.md)
**Surfaces:** score flow in `web/src/lib/components/ActiveSession.svelte`, the numpad in `web/src/routes/s/[id]/+page.svelte`, `web/src/lib/stores/numpad.ts`

> Register: **working screen → calm.** Reached from the Active round view (#170): admin taps a court → this entry surface. Big scores are *functional* emphasis (kept prominent), not decoration. Auto-complement is canonical (#169). The old immersive dark-green unified card + court tabs + NET divider + service/receive + 4-tab nav from the previous draft of this file are **superseded** by the north-star and #170.

---

## The decisions (from grilling; entry model chosen from concrete previews)

1. **Surface: a bottom sheet over the round.** Tapping a court's "Enter score" (admin) opens a per-court entry **sheet** over the round view; Finalize (or dismiss) returns to the round. Consistent for 1-court and multi-court sessions; composes with the existing `Drawer`/`Sheet` primitives.
2. **Entry model: steppers + numpad, both (calmed).** Two team rows, each `− [score] +`, tap the number to type. `+/−` supports live point-by-point tracking; the keypad is the fast path. **Auto-complement**: entering one team's score fills the other to the points target.
3. **Keypad is integrated into the entry sheet** (not a second stacked drawer over it) — tapping a team's score focuses it and reveals the keypad within the same sheet, avoiding nested-modal jank.

---

## Layout & hierarchy (the entry sheet)

Bottom sheet, `max-w-[480px]`, safe-area aware:

1. **Header** — `Court {n} · Round {x}` + close (`×`).
2. **Team A block** — overlapping avatars + team name (`text-[15px] font-semibold`), then the entry row: circular `−` · **big score** (`text-[64px] font-[800] tabular-nums`, tappable) · `+`.
3. **Validity readout** (center) — `{a} + {b} = {target}` with a check when the sum matches; `text-text-secondary`, turns `text-primary` when valid. Optional thin split bar visualising a/b of the target.
4. **Team B block** — mirror of A (score row, name, avatars).
5. **Keypad** (revealed when a score is focused) — `1–9`, `0`, `⌫`, done; types into the focused team, auto-completes the other. `bg-surface-raised` keys, `--radius` tokens, `active:scale-95`. Over-target entry → `animate-shake` (the **token**, not `animate-[shake_0.4s_ease-in-out]`).
6. **Finalize** — full-width `Button` `cta` variant (the shared size from #170), gated until `a + b === target`. Helper below: "Scores synced live to all devices" (`text-text-disabled`).

Calm surfaces throughout — **no** `#3d7a24`, **no** court-pattern SVG. Scores are large and `tabular-nums` (functional), on neutral surface.

---

## Interaction & flow

- **`+/−`** adjust by 1, clamped `[0, target]`; each change live-saves via `api.scores.updateLive` (SSE broadcasts to other devices — keep this, per CLAUDE.md live-updates invariant).
- **Tap a score** → focus it + reveal keypad. First digit overwrites; further digits append; entry `> target` is rejected with a shake. On done/auto-complete, the *other* team fills to `target − entered`.
- **Auto-complement made visible** (audit flagged it was silent): the derived team's number updates with a subtle transition, and the `a + b = target` readout confirms the relationship.
- **Finalize** → `api.scores.submit`; success toast; sheet dismisses to the round; the court card shows the final result.
- **Re-edit** — tapping a finalized court reopens the sheet in editing mode with current scores (matches today's `editing[]` behaviour).
- **Non-admins** never see this surface (read-only scores on the round view).

---

## States to implement

- **Untouched** (0–0) — Finalize disabled.
- **Editing / focused** — keypad visible, a team focused.
- **Valid** (`a + b === target`) — readout `primary` + check; Finalize enabled.
- **Invalid** (sum ≠ target) — Finalize disabled; readout neutral.
- **Over-target entry** — shake, entry rejected.
- **Submitting** — Finalize shows progress; inputs locked.
- **Finalized (re-editable)** — court shows final; reopening returns to Editing.

---

## Components & tokens

- **Reuse** `Button` (`cta` variant from #170), `Avatar`, `Sheet`/`Drawer`, `Toaster`.
- **Shared addition (anti-divergence rule):** the numpad is currently a store (`numpad.ts`) plus inline markup in the route. Extract the entry surface into a shared component (e.g. `ScoreEntrySheet` in `web/src/lib/components/`) that owns the two-team + integrated-keypad UI, so it isn't hand-rolled inline. Keep the store or fold it in.
- **Token fixes:** `animate-shake` (not the arbitrary value), `text-primary-foreground` (not `text-white`), radius tokens, drop `#3d7a24`.

---

## A11y & responsiveness

- Steppers, keypad keys, close, Finalize: min 48×48px; `focus-visible` rings.
- `aria-label`s on `+/−` ("Increase Team A score") and keypad keys.
- Validity + auto-complement announced via `aria-live="polite"`.
- Sheet focus-trapped; Esc/backdrop/swipe-down dismisses; focus returns to the court card.
- Safe-area insets on the sheet bottom; single column ≤480px.
- All copy via `$_()`.

---

## Before → after

| Before (shipped) | After (this spec) |
|---|---|
| Two immersive `#3d7a24` cards inline in the scoring tab | Calm per-court entry **sheet** over the round |
| Global numpad **drawer** triggered separately (nested-drawer risk) | Keypad **integrated** into the entry sheet |
| Auto-complement silent | Visible `a + b = target` readout + transition |
| `animate-[shake_…]`, `text-white`, `#3d7a24` | `animate-shake`, `text-primary-foreground`, tokens |
| Numpad markup inline in the route | Shared `ScoreEntrySheet` component |

---

## Out of scope

- Round/court **display** and the affordance into scoring → Active round (#170).
- Leaderboard content → Live leaderboard (#172).
- Changing the auto-complement rule itself (ratified in #169); new modes; dark mode.

---
---

# Spec: Avatar Icon System

> _(Separate pre-existing spec kept in this file for history; not part of the #171 score-entry redesign.)_

## Goal

Replace plain initials with user-chosen icons from a curated set.
No photo uploads — icons keep it lightweight and fun.

## How It Works

- User picks an **icon** + **background color** in their profile settings
- Displayed everywhere a player avatar appears (score card, lobby, leaderboard, etc.)
- Guest players (no account) get a **random icon + color** auto-assigned on join
  — stored on their player record for the session lifetime

## Icon Set

~24 icons from `lucide-svelte`, sport/fun themed. Examples:
`Zap`, `Star`, `Flame`, `Shield`, `Crown`, `Sword`, `Trophy`,
`Target`, `Rocket`, `Ghost`, `Cat`, `Dog`, `Bird`, `Fish`,
`Leaf`, `Sun`, `Moon`, `Snowflake`, `Mountain`, `Waves`,
`Music`, `Heart`, `Smile`, `Bolt`

Rendered at ~20px inside the avatar circle.

## Color Palette

~10 curated background colors (not user-freeform):
Forest green, Sky blue, Warm orange, Coral, Purple, Teal,
Gold, Slate, Rose, Charcoal

## Avatar Component

```
┌─────────────────────┐
│  [colored circle]   │
│    [icon inside]    │   size variants: sm (28px), md (36px), lg (48px)
└─────────────────────┘
```

Fallback: if no icon set, show initials (current behaviour).

## Data Model Changes

```sql
-- users table: add two columns
avatar_icon   TEXT   -- lucide icon name, e.g. "Flame"
avatar_color  TEXT   -- color key, e.g. "forest" | "sky" | "coral" ...

-- players table: add two columns (for guests)
avatar_icon   TEXT
avatar_color  TEXT
```

## Profile Settings

Add an icon/color picker to the profile page:
- Grid of icon options (tap to select, highlight active)
- Row of color swatches below
- Live preview of resulting avatar
- Saves via existing `PUT /api/auth/profile` endpoint (add fields)

## Open Questions

- [ ] Which invite screen is being redesigned? (lobby / send invite / accept card)
- [ ] Should guest avatar assignment persist across sessions for the same device,
      or is it random each time?
- [ ] Cap the icon set at ~24 or allow more?
