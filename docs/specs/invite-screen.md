# Spec: Invite / Join Screen Redesign

## Goal

The screen shown when someone clicks an invite link (pre-join, State 1 in Lobby.svelte).
Currently a bare heading + form. Should feel like an actual invitation — show who's
hosting, who's already in, and make joining feel welcoming.

## Reference Design Language

Same visual system as the score screen redesign:
- Off-white background (`#F7F6F3`)
- Dark forest green accents
- Avatar icon circles (from avatar system spec)
- Bold typography for key info
- Card-based layout, generous spacing

---

## Layout (top to bottom)

```
[Back button ×]                                          (top right)

[Host section]
  [avatar icon — large, 56px]
  "{Host name} invited you to play"                     (14px, muted)
  "{Session name or game mode}"                         (28px, bold)
  "Americano • 2 courts • 32 pts"                       (14px, muted, green accents)

[Players card]
  PLAYERS (3)                                           (label, 11px uppercase)
  ┌──────────────────────────────────────────┐
  │  [avatar] [avatar] [avatar] + 2 more     │  ← stacked avatar row
  │  "Mads, Søren, Jakob and 2 others"       │  ← names summary, muted
  └──────────────────────────────────────────┘

[Join section]
  ┌──────────────────────────────────────────┐
  │  If logged in:                           │
  │    [avatar]  Your Name         (bold)    │
  │              your@email.com    (muted)   │
  │                                          │
  │    [Join Session →]  (full width, green) │
  ├──────────────────────────────────────────┤
  │  If not logged in:                       │
  │    [Sign in]          (full width, green)│
  │    No account? Create one                │
  │                                          │
  │    ────────── or ──────────              │
  │                                          │
  │    [Your name...        ] [Join]         │
  └──────────────────────────────────────────┘

[Footer — muted, small]
  "Powered by OpenPadel"
```

---

## Key Changes from Current

| Current | Redesigned |
|---|---|
| Plain heading: "Join X's Americano" | Host avatar + "X invited you" + bold session name |
| Bare details line | Styled detail chips (mode • courts • pts) with green accents |
| No player visibility | Players card showing stacked avatars + name summary |
| Avatar = initials only | Avatar icon system (icon + color) |
| Sign in / guest form unstyled | Cleaner card with logged-in profile preview |

---

## Players Card

Show up to 4 stacked avatar circles (overlapping, like avatar groups).
If more than 4 players: show 4 avatars + "+N more" label.
Below avatars: first-name list summary — "Mads, Søren, Jakob and 2 others".
If 0 players (just created): "No one has joined yet — be first!"

Stacked avatar layout:
```
[●][●][●][●] +2       ← each circle offset -8px, 36px diameter
```

**Live updates**: polls `GET /api/sessions/:id` every 3s — avatars animate in
as new players join while you're looking at the screen. No refresh needed.

---

## Session Detail Chips

```
Americano  •  Court 1  •  32 pts
```
- Game mode in default text
- Courts and points in primary green
- Separator dots in muted gray

---

## Open Questions

- [x] Show scheduled time/date if set
- [x] If session already started: show "Session in progress" message, no join action
