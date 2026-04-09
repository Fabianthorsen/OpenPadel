# Spec: Score Screen Redesign

## Goal

Redesign the active match score entry screen to feel more immersive and court-like.
Single unified dark match card, player avatars, cleaner role/server indicators.

## Reference

Screenshot: dark green unified score card with avatars, SERVICE/RECEIVE badges,
NET divider, role badge, official card, and Finalize button outside the card.

---

## Screen Layout (top to bottom)

```
[Header]
  [session avatar]  Session name              [settings icon]

[Role badge]
  ✏ ACTIVE SCOREKEEPER            ← only shown to the scorekeeper

[Match info row]
  AMERICANO TOURNAMENT            (small caps, muted)
  Match 4 of 15                   [Courts Overview ▦]
  (large, bold)

[Court tabs]  ← horizontal scroll
  [🎾 COURT 1 •]  [🎾 COURT 2]  [🎾 COURT 3]
  active tab has dot indicator + filled background

[Score card]  ← single dark green card, one per active court tab
  [avatar][avatar]                ← overlapping circles, centered top
  Marcus & Sofia                  ← first names joined with &, centered
  TEAM A                          ← small caps, muted, centered

  [−]        21        [+]

  ── NET ──                       ← bold italic pill, VERSUS styling

  [−]        17        [+]

  Anders & Elena                  ← centered
  TEAM B                          ← small caps, muted, centered
  [avatar][avatar]                ← overlapping circles, centered bottom

[Finalize Result ✓]  ← full-width, dark green, with checkmark icon
"Scores are synced live to all player devices"  ← muted helper text

[Official card]
  [avatar]  Official: Lukas H.    [✓ shield]
```

---

## Component Changes

### Court tabs
- Horizontal scrollable tab row between match header and score card
- Each tab: racket icon + "COURT N" label
- Active tab: filled/highlighted background + dot indicator
- Tapping switches the score card to that court
- Replaces current stacked multi-card layout entirely

### Courts Overview
- Chip/button next to "Match X of Y" header
- Opens a summary view of all courts at a glance (scores, status)
- Useful for admins managing multiple courts simultaneously
- Can be a bottom sheet or full-screen overlay

### Score card
- **Background**: dark forest green, single rounded card
- **Team layout**: centered, names at top and bottom (not left-aligned)
- **Player names**: "First & First" format joined with ampersand
- **Avatars**: two overlapping circles (36px), centered above Team A name and below Team B name
- **NET divider**: centered pill badge, bold italic styling (prominent like VERSUS, label stays "NET")
- **Score numbers**: white, very large (~80px), centered
- **+/− buttons**: translucent dark circles, flanking the score

### Score entry interaction
Two input modes, always available:

**Incremental** — tap +/− to adjust by 1 (good for live tracking during play)

**Direct entry** — tap the score number itself → numpad appears → type score → confirm
- On confirm: if scores don't sum to target, the *other* team auto-adjusts to `target − entered`
- Example: target 24, enter 23 for Team A → Team B auto-fills to 1
- If entered score exceeds target: reject with brief shake animation
- Numpad shows target as a hint: "Target: 24"

**Finalize** stays disabled until both scores sum to target.
Auto-fill means in practice you only ever need to enter one number.

### Role badge
- Small pill above match title: "ACTIVE SCOREKEEPER" with pencil icon
- Only visible to the scorekeeper / admin
- Hidden for spectators

### Match header
- "AMERICANO TOURNAMENT" small caps label above
- "Match X of Y" large bold
- "Courts Overview" chip on the right

### Finalize button + helper text
- Full width, dark green, "Finalize Result" + checkmark icon
- Below: "Scores are synced live to all player devices" in muted small text
- Disabled + muted until scores sum to session.points

### Official card
- Below the finalize button
- Scorekeeper avatar + "Official: [name]" + verified shield

### Bottom nav
- 4 tabs: **SCORING** / **STANDINGS** / **PLAYERS** / **SCHEDULE**
- SCORING replaces current "LIVE" label
- SCHEDULE tab: shows round schedule / upcoming matchups

---

## Open Questions

- [x] Courts Overview: bottom sheet, shows court number + team names + live score + status icon (not started / in progress / finalized)
- [x] SERVICE/RECEIVE indicators — removed from Americano, kept in Tennis only
- [x] No SCHEDULE tab — "Match X of Y" header is tappable, opens a bottom sheet with the full round list
- [x] "Official" = always admin for now, delegation is a future roadmap item
- [x] PLAYERS tab: bench players prominently at top, then full active player list below

---

## Out of Scope

- Score entry flow changes (still +/− tapping, still finalise on sum match)
- Mexicano / Tennis modes — spec Americano first, extend after

---

# Spec: Avatar Icon System

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
