# Spec: Session Creation

## Goal

The drawer that creates a new Session from the home screen. Kept deliberately minimal —
one choice (game mode), sensible defaults for everything else, tune the rest in the Lobby
afterward. This documents the current, shipped implementation (`CreateDrawer.svelte`), not a
pending redesign.

## Layout

```
┌──────────────────────────────────────────┐
│  Let's set up a session          [×]     │  ← Drawer.Header
├──────────────────────────────────────────┤
│  [Americano] [Mexicano]                  │  ← pill toggle, Americano default
│  "Rotating partners, individual points"  │  ← hint text, swaps per mode
│                                          │
│  [ Create session ]                     │  ← full width, green, disabled while creating
└──────────────────────────────────────────┘
```

Bottom sheet on mobile, floating card (480px, bottom-anchored) on desktop — `Drawer.Root`.

## Behavior

- Defaults on create: `courts: 2, points: 24`; Mexicano additionally sets `rounds_total: 7`.
  No pickers for these here — courts/points/rounds/name/schedule are all edited inline in the
  Lobby after creation (see `docs/specs/lobby.md`), not at creation time.
- On submit: creates the Session, stores `admin_token_<id>` in `localStorage`, immediately
  joins the creator as a Player using their account display name, stores `player_id_<id>`
  and `last_session_id`, then navigates to `/s/:id?token=<adminToken>`.
- Creator must be logged in — `auth.user!.display_name` is read without a null check, so this
  drawer assumes an authenticated context (unauthenticated users don't reach it from the UI).
- Error state: inline `text-destructive` message below the toggle, from `translateApiError`.

## Key Design Decisions

| Decision | Why |
|---|---|
| No config pickers at creation | Reduces time-to-first-action; courts/points rarely known until players are actually gathering in the Lobby |
| Mexicano forces `rounds_total: 7` default | Mexicano has no bench/auto round count like Americano — needs an explicit starting value, tunable via the Lobby's rounds Stepper |
| Creator auto-joins as a Player | Creator is a participant by default, not just an organizer — consistent with the Admin ≠ Creator model in `CONTEXT.md` |

## Open Questions

- [ ] Should the creation drawer let admins name the session up front, instead of always
      renaming afterward in the Lobby?
- [ ] Is there a guest-creator path (create without an account), or is session creation always
      gated behind login? Current code assumes the latter.
