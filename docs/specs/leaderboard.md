# Spec: Leaderboard Screen

## Goal

The public, no-auth standings view for a Session — shown live during `playing` and as the
final results screen once `done`. This documents the current, shipped implementation
(`Leaderboard.svelte`).

## Layout — Live (during `playing`)

```
[Leader hero card]  ← forest green, faint court-line pattern background
  [avatar lg]  LEADER
               {Name}                    (24px, bold, white)
               {points} PTS   |   {W}/{D}/{L}  W-L

[Current round label]                     "Round 3 of 7"
  #  PLAYER              GAMES   W-L    PTS
  1  [avatar] Name        4      3/0/1   18   ← rank 1-3 get colored row bg (green shades)
  2  [avatar] Name        4      2/1/1   15
  3  [avatar] Name        4      2/0/2   12
  4  [avatar] Name        4      1/1/2    9   ← alternating row bg below podium
  ...
```

## Layout — Final (once `done`)

```
[Heading]
  FINAL RESULTS
  {Session name}

[Podium]  ← 2nd(left) / 1st(centre, taller+trophy) / 3rd(right)
  🏆
  [avatar xl]      [avatar lg]      [avatar lg]
  ①                ②                ③
  Name              Name             Name
  N PTS             N PTS            N PTS
  W·D·L             W·D·L            W·D·L
  [+ Add]           [+ Add]          [+ Add]      ← contacts, logged-in viewer only
  ▇▇▇ bar          ▇▇ bar           ▇ bar          ← podium bars, height by rank

[Full ranking, 4th+]
  #4  [avatar] Name   W/D/L   N PTS  [+]
  ...

  [✕ Close]
```

## Behavior

- **Live vs Final is one component, `complete` prop switches rendering** — not two routes. Same
  data shape (`App.Leaderboard`), different layout.
- **Live refresh** — reloads on the `round_updated` SSE event via `stream.onEvent`; no polling
  inside this component (the page-level 30s fallback poll covers the SSE-buffered-proxy case,
  see `ARCHITECTURE.md`).
- **Final-only: contacts** — "Add" button next to each standing (except the viewer's own row)
  calls `api.contacts.add`; only rendered when `auth.token` is present and the player has a
  `user_id` (Guests without an account can't be added as a contact).
- **W/D/L is derived, not stored per-row** — `losses = games_played - wins - draws` computed
  inline everywhere it's shown.
- **Podium ordering** — visually 2nd/1st/3rd (`podiumOrder` array), not rank order, so the
  winner sits centered and tallest.

## Key Design Decisions

| Decision | Why |
|---|---|
| Single component for live + final | Standings data and row rendering are identical; only the header/podium framing differs — avoids duplicating the standings list markup |
| No "Add contact" during live play | Contacts are a post-game, session-summary action — adding mid-session isn't a driving use case, and the final screen is where the viewer's full attention is already on standings |
| Podium reorders visually | Matches physical podium expectations (1st in the middle) over simple left-to-right rank order |

## Open Questions

- [ ] Does a Guest player (no `user_id`) get any equivalent of "Add contact" for a User who
      wants to remember playing with them, or is that intentionally one-directional (Users
      can add each other, Guests can't be added)?
- [ ] Should the live leaderboard show anything for a tied rank (equal points), or is silent
      insertion order (by player_id) acceptable?
