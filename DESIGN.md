# Design — OpenPadel

> **Foundation updated 2026-07-14** by [Lock the design north-star & page-redesign rubric](https://github.com/Fabianthorsen/OpenPadel/issues/169) (#169),
> reconciling this doc with the shipped app per the [UI audit](docs/research/ui-audit.md).
> The **Principles → What this is not** sections below are current. The **Key Screens** mockups
> are being redrawn one at a time by their per-page spec tickets — treat them as *historical* until
> each is refreshed. The shared **[page-redesign rubric](docs/specs/redesign-rubric.md)** operationalises this doc into a checklist.

---

## Principles

1. **Courtside first.** Every working screen is used standing up, phone in one hand, sun overhead.
   Big tap targets. No tiny text. No fiddly inputs.
2. **Score in 3 taps.** The most frequent action (entering a score) must be instant.
3. **Glanceable leaderboard.** A player should read the standings in under 2 seconds.
4. **Calm by default, bold when we celebrate.** See below — the reconciled north-star.

### Calm default, bold celebratory

OpenPadel has two registers, and the design language switches between them deliberately:

- **Working screens** — Home, Session creation, Lobby & join, Active round, Score entry, Auth,
  Profile. These get **Nordic restraint**: muted surfaces, one accent used sparingly, typographic
  hierarchy, functional icons only, no decorative motifs. Calm, fast, legible in sunlight.
- **Celebratory surfaces** — the **Session complete** finale (and the winner moment). This is where
  the product gets to feel like a win: podium, trophy, larger display type, and decorative motifs
  (court-line SVG, trophy flourishes). **Note (decided in #172):** the **live** leaderboard, though it
  shows standings, stays **calm & glanceable** — it's read repeatedly mid-session, so it follows
  working-screen restraint (rank · name · points, rank-1 in `primary`, no hero card, no colored rows).
  The drama is reserved for the finale.

The failure mode the audit caught was expressive treatments *leaking into working screens*
(saturated score cards, court-tab emoji). The rule: **expressive == celebratory-only.** When in
doubt, a screen is a working screen.

---

## Color System

Light mode only (see *What this is not*). Source of truth is `web/src/app.css` `@theme`, mirrored in
`web/src/lib/design-tokens.ts`. **Use tokens, never literal hex.**

```
Background       #edeee8   warm off-white, main surface        --color-background
Surface          #f4f4f0   cards, sheets                        --color-surface
Surface raised   #e3e3dc   inputs, segmented controls, hover    --color-surface-raised
Border           #d2d2cb   subtle dividers                      --color-border
Border strong    #b2b2aa   emphasized borders, active inputs    --color-border-strong

Primary          #2d5a1a   THE accent — buttons, active, rank#1 --color-primary
Primary hover    #234d13   pressed/hover                        --color-primary-hover
Primary muted    #e2ede0   tinted backgrounds, selected, tags   --color-primary-muted
Primary foreground #ffffff text on primary (use this, not text-white) --color-primary-foreground

Text primary     #1a1a16   headings, scores — warm near-black   --color-text-primary
Text secondary   #68685e   labels, meta, bench                  --color-text-secondary
Text disabled    #aeaea4   —                                    --color-text-disabled

Positive         #2d5a1a   same as primary — score confirmed    --color-positive
Destructive      #c0392b   use sparingly                        --color-destructive
Destructive fg   #ffffff   text on destructive                  --color-destructive-foreground
```

**One accent.** `--color-primary` is the *only* green on working screens. The stray greens the audit
found — `#3d7a24` (score cards), `#4A7856` (`app.html` `theme-color` meta) — **fold into
`--color-primary`.** Update `app.html`'s `theme-color` to `#2d5a1a`.

**Celebratory palette (Session complete finale only).** The podium uses real **medal colours**
(decided #174): `--color-medal-gold` / `--color-medal-silver` / `--color-medal-bronze` for 1st / 2nd /
3rd — as ring / bar / rank-badge accents, not body text (ensure contrast). Added to `@theme` +
`design-tokens.ts` when the finale first needs them. This is the *only* non-primary palette, and only
on the finale.

**Semantic states.** Warning ("time expired") uses `--color-warning` (added #170) — not raw `amber-*`.
The **"live/playing" indicator uses `--color-primary`** — a pulsing dot / `primary-muted` tint,
consistent across the active round (#170), the leaderboard (#172) and profile (#178) — not raw
`emerald-*`. Losses/negatives use `--color-destructive` (never `text-[#c0392b]`).

**shadcn-compat tokens.** `app.css` also defines `--color-{card,popover,secondary,muted,accent,input,ring,foreground}`
(+ `*-foreground`) for the vendored bits-ui components. These are plumbing — don't reach for them in
app code; use the named tokens above.

No gradients. Shadows are subtle and reserved (see *What this is not*).

---

## Typography

**Two families, two registers** (self-hosted via `@fontsource`, no Google Fonts request):

```
UI / body / data   Geist Sans          --font-sans     (400–800)  — all working screens
Display            Schibsted Grotesk   --font-display  (600–800)  — wordmark + finale only
```

- **Geist Sans** (`--font-sans`) is the default face for everything: body, labels, and all
  numeric/data (scores, standings) — it has clean tabular figures. Use `font-variant-numeric:
  tabular-nums` on scores as before.
- **Schibsted Grotesk** (`--font-display`) is the celebratory display face. Reach for it via the
  `font-display` utility, and **only** on: the **OpenPadel wordmark** (Home, Auth) and the
  **Session-complete finale** (finale title + podium names). It is a Nordic grotesk chosen to match
  the product's origin. Do **not** put it on working-screen titles, page headers, or score numerals —
  those stay Geist. Expressive == celebratory-only still holds.

There are **no font-size/weight tokens yet** — the app scatters arbitrary values (`text-[28px]`,
`font-[800]`). Target scale to tokenize (add as `@theme` tokens when a page first touches type):

```
Display   64–80 / 800   celebratory only (winner, big scores)   line-height 1.0
H1        28    / 700    screen titles                           line-height 1.1
H2        18    / 600    section headers, player names           line-height 1.2
H3        15    / 600    match cards, court labels
Body      15    / 400    regular content                         line-height 1.5
Small     13    / 400    meta, timestamps
Micro     11    / 600    uppercase section labels (SectionLabel), tracking 0.1em
```

Weight **800** is allowed for *display* (celebratory); working screens top out at 700.
Headings: `letter-spacing: -0.02em`, `line-height: 1.1` (matches `app.css`).
Scores always use `font-variant-numeric: tabular-nums`.

---

## Spacing & Layout

Base unit 4px. Tokens `--spacing-0..4` cover 0/4/8/12/16px; larger steps (20/24/32/48) use Tailwind
utilities directly for now.

```
Screen padding      16px sides (safe-area inset aware — always honor pt-safe / pb-safe)
Card padding        16px
Section gap         24px
Minimum tap target  48×48px
```

**Border radius — reconciled with real usage** (this mismatch is why primitives kept getting
class-overridden; add the missing tokens when touched):

```
sm    4px   (0.25rem)  badges                          --radius-sm
md    8px   (0.5rem)   inputs, small controls          --radius-md
base  12px  (0.75rem)  —                               --radius
lg    16px  (rounded-2xl)  cards, CTAs   ← most common — ADD TOKEN
xl    24px  (rounded-3xl)  score cards, modals, sheets  ← ADD TOKEN
full  99px  pills                                       (rounded-full)
```

Single column layout. Max content width 480px, centered. Phone app — no responsive grid.

---

## Component & tech stack

**No `shadcn-svelte`, no Skeleton UI** (the Skeleton experiment was installed and reverted). The UI is
a set of **vendored components** in `web/src/lib/components/ui/*`, built on:

- **`bits-ui`** (^2.16.3) — accessible primitives (Dialog, Toggle, Switch, Tabs, Label, Separator…)
- **`tailwind-variants`** — the Phase-1 variant pattern (all new/refactored components use `tv`)
- **`clsx` + `tailwind-merge`** (via `cn()`) — class composition
- **`svelte-sonner`** — toasts (`Toaster` in `+layout.svelte`; not "Sonner")
- **`@lucide/svelte`** — icons (note: legacy `lucide-svelte` still lingers in `Avatar` — consolidate)

**Score numpad** is a custom store (`$lib/stores/numpad`) with **auto-complement** — you enter one
team's score and the other fills to the points target. This (not a native `inputmode=numeric` input)
is the canonical model.

Phase-1 components on the tv+JSDoc pattern: **Button, Input, Label, Switch, Toggle, Badge, Drawer.**
The other 12 primitives are being migrated as touched (see `CONTEXT.md` component table + the
[UI audit](docs/research/ui-audit.md) inventory).

| Need              | Use                                                        |
|-------------------|------------------------------------------------------------|
| Buttons / CTAs    | `Button` — **needs** a full-width `cta` size + solid `destructive` (see rubric) |
| Text inputs       | `Input`                                                    |
| Segmented control | `PillToggleGroup` / `ToggleGroup`                          |
| Cards             | `Card` (currently under-used — prefer it over hand-rolled `bg-surface-raised rounded-2xl`) |
| Section labels    | `SectionLabel` (do **not** re-type `text-[11px] … uppercase` inline) |
| Numeric stepper   | `Stepper`                                                  |
| Toast             | `svelte-sonner`                                            |
| Sheet / drawer    | `Drawer` (bottom slide-up), `Sheet` (sides), `Dialog` (centered) |

---

## Micro-interactions

Shipped, keep: `active:scale-95/98` press feedback; `animate-shake` (invalid / over-max score);
`animate-ptr-fade` (pull-to-refresh); a **toast** on score confirm; `transition-colors` on state
changes. All 150–200ms, `ease-out`, nothing decorative.

The older aspirational animations (row fade-in-from-below, green border-flash on confirm, leaderboard
re-sort animation) are **not implemented** — a page spec may add them, but they're not a baseline.

---

## What this is not

- **No dark mode in V1.** There is no dark palette. The vendored components leak `dark:` utility
  classes that activate via `prefers-color-scheme` on dark devices (an inconsistent half-dark render) —
  **strip that leakage** so light mode renders consistently. A real dark theme is a future effort.
- **No emojis.** The shipped `🎾` / `⏱` / `ℹ` are replaced with `@lucide/svelte` icons.
- **Icons:** the full *functional* lucide set is allowed app-wide. **Decorative** treatments —
  court-line SVG patterns, trophy flourishes — are **celebratory-surface only** (leaderboard/complete).
- **No gradients.**
- **Shadows are subtle and reserved.** `--shadow-sm` for resting cards; `--shadow-md/lg` only for
  lifted/celebratory chrome (bottom nav, drawers, winner card). Not decorative drop-shadows.

---

## Key Screens

> ⚠️ **Historical — being redrawn per page-spec.** These mockups predate the shipped app and the
> reconcile decision above. Each is authoritative only once its per-page spec ticket refreshes it.
> Where a mockup conflicts with the Principles/Color/Type sections above, those sections win.

### 1. Home / Create session

Minimal. Logo + one action.

```
┌─────────────────────────┐
│                         │
│                         │
│   OpenPadel             │  ← wordmark, H1, Text primary
│   Padel, organised.     │  ← Small, Text secondary
│                         │
│                         │
│  ┌─────────────────┐   │
│  │  Start session  │   │  ← Primary button (green bg, white text)
│  └─────────────────┘   │
│                         │
│  Join with a link →     │  ← Text link, small, centered
│                         │
└─────────────────────────┘
```

---

### 2. Session setup (admin)

Grouped, not a form dump. Segmented controls over dropdowns.

```
  Courts
  ┌────┐ ┌────┐ ┌────┐ ┌────┐
  │ 1  │ │ 2  │ │ 3  │ │ 4  │   ← ToggleGroup, pill style
  └────┘ └────┘ └────┘ └────┘

  Points per game
  ┌────┐ ┌────────┐ ┌────┐
  │ 16 │ │   24   │ │ 32 │      ← 24 selected by default
  └────┘ └────────┘ └────┘

  ─────────────────────────────

  Share link
  openpadel.app/s/abc123  [Copy]

  [ Share ]                      ← native OS share sheet

  ─────────────────────────────

  Waiting for players (2)...
  [ Start when ready → ]         ← disabled until player count == courts × 4
```

---

### 3. Lobby (all players)

Player joins via link, types their name. Admin view shows Start button.

```
┌─────────────────────────┐
│  Waiting to start       │
│  2 courts · 24 pts      │  ← Small, Text secondary
│                         │
│  Players (6)            │
│                         │
│  ┌───┐ Ana             │  ← list, not grid — easier to scan names
│  ├───┤ Bruno           │
│  ├───┤ Carl            │
│  ├───┤ Diana           │
│  ├───┤ Erik            │
│  └───┘ Fiona           │
│                         │
│  · · ·                  │  ← subtle pulse, waiting indicator
│                         │
└─────────────────────────┘

Admin sees:
│  ┌─────────────────┐   │
│  │   Start →       │   │  ← green, enabled when player count == courts × 4
│  └─────────────────┘   │
```

---

### 4. Active round view (everyone)

Primary info: court assignments. Bench quiet but visible.

```
┌─────────────────────────┐
│  Round 3 of 9      LIVE │  ← LIVE in Primary color, no dot
│                         │
│  Court 1                │  ← H3, Text secondary
│  ┌─────────────────┐   │
│  │  Ana · Bruno    │   │  ← H2 weight names
│  │  vs             │   │  ← Small, secondary, centered
│  │  Carl · Diana   │   │
│  └─────────────────┘   │
│                         │
│  Court 2                │
│  ┌─────────────────┐   │
│  │  Erik · Fiona   │   │
│  │  vs             │   │
│  │  Gio · Hanna    │   │
│  └─────────────────┘   │
│                         │
│  Bench — Ivan           │  ← Small, Text disabled
│                         │
└─────────────────────────┘

Admin also sees:
│  [ Enter scores ]        │  ← Primary button, bottom
```

---

### 5. Score entry (admin only)

Auto-complement numpad (custom store). Tap number to type; the other team's score fills to target.

```
┌─────────────────────────┐
│  Court 1 · Round 3      │
│                         │
│  Ana + Bruno            │
│  ┌─────────────────┐   │
│  │  −     15    +  │   │  ← Display size, tap number → opens numpad
│  └─────────────────┘   │
│                         │
│  Carl + Diana           │
│  ┌─────────────────┐   │
│  │  −      9    +  │   │  ← auto-fills to points target
│  └─────────────────┘   │
│                         │
│  ┌─────────────────┐   │
│  │   Confirm       │   │  ← Disabled until both scores sum to target
│  └─────────────────┘   │
└─────────────────────────┘
```

+/− adjusts by 1. Tapping a number opens the custom numpad; entering one score auto-complements the
other. Confirm stays disabled until both scores sum to the points target.

---

### 6. Leaderboard  *(celebratory surface — expressive treatments allowed)*

```
┌─────────────────────────┐
│  Leaderboard            │
│  Round 3 of 9 · Live    │
│                         │
│   #   Name        Pts   │  ← Small caps header, Text secondary
│   ─────────────────     │
│   1   Ana          38   │  ← #1: Primary color rank number
│   2   Bruno        35   │
│   3   Carl         31   │
│   ...                   │
│                         │
│  Updated just now       │  ← Small, Text disabled, bottom
└─────────────────────────┘
```

The shipped leaderboard is richer than this (leader hero card, W/L columns, avatars, podium). As a
celebratory surface it's allowed to be — the spec ticket redraws it deliberately.

---

### 7. Session complete  *(celebratory surface)*

```
┌─────────────────────────┐
│                         │
│  Session complete       │  ← H1
│                         │
│  Ana                    │  ← Display size, Primary color
│  Winner · 38 pts        │  ← Small, secondary
│                         │
│  ─────────────────      │
│                         │
│  2   Bruno      35      │
│  3   Carl       31      │
│  ...                    │
│                         │
│  [ Share results ]      │  ← Secondary button (outlined)
│                         │
└─────────────────────────┘
```
