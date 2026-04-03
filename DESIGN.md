# Design — NotTennis
### By Sofia

---

## Principles

1. **Courtside first.** Every screen is used standing up, phone in one hand, sun overhead.
   Big tap targets. No tiny text. No fiddly inputs.
2. **Score in 3 taps.** The most frequent action (entering a score) must be instant.
3. **Glanceable leaderboard.** A player should read the standings in under 2 seconds.
4. **Nordic restraint.** Muted, calm, typographically led. Colour is used sparingly and with purpose.

---

## Color System

Light mode first. Warm off-white base — not pure white, which feels clinical.
Dark mode in V2.

```
Background       #F7F6F3   warm off-white, main surface
Surface          #FFFFFF   cards, sheets
Surface raised   #EFEDE8   inputs, segmented controls, hover
Border           #E0DDD7   subtle dividers
Border strong    #C8C4BC   emphasized borders, active inputs

Primary          #4A7856   forest green — buttons, active states, rank #1
Primary hover    #3D6348   pressed/hover
Primary muted    #EDF2EE   tinted backgrounds, selected states, tags

Text primary     #1C1B19   headings, scores — warm near-black
Text secondary   #6B6860   labels, meta, bench
Text disabled    #B5B2AB

Positive         #4A7856   same as primary — score confirmed
Destructive      #C0392B   (reserved, use sparingly)
```

Only one accent color. No gradients, no shadows beyond `box-shadow: 0 1px 3px rgba(0,0,0,0.07)`.

---

## Typography

```
Font: Inter — geometric, readable, excellent on mobile
      Load via fontsource (@fontsource/inter) — no Google Fonts request

Scale:
  Display   48px / 700   match scores on score entry screen
  H1        26px / 650   screen titles
  H2        18px / 600   section headers, player names in leaderboard
  H3        15px / 600   match cards, court labels
  Body      15px / 400   regular content
  Small     13px / 400   meta, bench label, timestamps
  Mono      14px / 500   round indicators, point totals (tabular-nums)
```

Letter spacing: `-0.01em` on headings. `0` on body.
Scores always use `font-variant-numeric: tabular-nums` — digits stay fixed width.
Line height: `1.5` body, `1.2` display/headings.

---

## Spacing & Layout

Base unit: 4px. Use multiples: 4, 8, 12, 16, 20, 24, 32, 48.

```
Screen padding      16px sides (safe area inset aware)
Card padding        16px
Card gap            10px
Section gap         24px
Bottom nav height   60px + safe area inset
Minimum tap target  48×48px
Border radius       8px cards, 6px inputs, 4px badges, 99px pills
```

Single column layout. Max content width 480px, centered.
This is a phone app — no responsive grid needed.

---

## Key Screens

### 1. Home / Create session

Minimal. Logo + one action.

```
┌─────────────────────────┐
│                         │
│                         │
│   NotTennis             │  ← wordmark, H1, Text primary
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
  nottennis.app/s/abc123  [Copy]

  [ Share ]                      ← native OS share sheet

  ─────────────────────────────

  Waiting for players (2)...
  [ Start when ready → ]         ← disabled until ≥ 5 players
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
│  │   Start →       │   │  ← green, enabled at ≥ 5 players
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

No keyboard by default. Tap number to type if needed.

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
│  │  −      9    +  │   │
│  └─────────────────┘   │
│                         │
│  24 / 24               │  ← Small, green when valid, grey when not
│                         │
│  ┌─────────────────┐   │
│  │   Confirm       │   │  ← Disabled until sum == points target
│  └─────────────────┘   │
└─────────────────────────┘
```

+/− adjusts by 1. Tapping the number opens a numeric input (no full keyboard — `inputmode="numeric"`).
Confirm stays disabled until both scores sum to the points target.

---

### 6. Leaderboard

Rank, name, points. Nothing else.

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
│   4   Diana        28   │
│   5   Erik         24   │
│   6   Fiona        22   │
│   7   Gio          19   │
│   8   Hanna        17   │
│   9   Ivan         12   │
│                         │
│  Updated just now       │  ← Small, Text disabled, bottom
└─────────────────────────┘
```

Rank 1 gets the Primary color on the number — no badge, no bold name. Subtle distinction.

---

### 7. Session complete

Quiet celebration. Typography does the work.

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
│  4   Diana      28      │
│  5   Erik       24      │
│  ...                    │
│                         │
│  [ Share results ]      │  ← Secondary button (outlined)
│                         │
└─────────────────────────┘
```

---

## Component Notes (shadcn-svelte)

| Need              | shadcn component | Notes                                             |
|-------------------|------------------|---------------------------------------------------|
| Primary button    | `Button`         | Custom green variant                              |
| Secondary button  | `Button`         | Outlined, border color, transparent bg            |
| Segmented control | `ToggleGroup`    | Courts and points selection, pill style           |
| Match cards       | `Card`           | White surface, 1px border, subtle shadow          |
| Score input       | Custom           | +/− with display number, tap for numpad           |
| Player list row   | Custom           | Simple flex row, 48px min height                  |
| Toast             | `Sonner`         | Score confirmed, player joined                    |
| Sheet / drawer    | `Drawer`         | Numpad overlay when tapping score number          |

---

## Micro-interactions

- Player joins lobby → row fades in from below, name appears
- Score confirmed → match card border flashes green briefly, then settles
- Leaderboard re-sort → rows animate to new positions (150ms ease-out)
- Round advances → previous round card fades, new one fades in
- All animations 150–200ms, `ease-out`. Nothing decorative.

---

## What this is not

- No gradients
- No shadows beyond a single 1px border or `0 1px 3px rgba(0,0,0,0.07)`
- No emojis
- No decorative icons beyond functional ones (copy, share, chevron)
- No dark mode in V1
