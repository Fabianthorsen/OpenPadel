# Spec: Lobby Screen

## Goal

The main screen for a Session in `lobby` state — where admins tune config, players join,
and everyone waits for "Start". This documents the current, shipped implementation
(`Lobby.svelte`, State 2 — pre-join is covered separately in `docs/specs/invite-screen.md`).

## Layout (admin view, top to bottom)

```
[Header]
  "Waiting for players" or scheduled date/time      [session name, tap to edit]
  "2 courts · 24 pts · Americano"                    (ⓘ rules)  [× back]

[Config card]  ← admin only
  CONFIG
  GAME MODE        [Americano] [Mexicano]            ← 2-up cards, hint text each
  COURTS           (1) (2) (3) (4)                    ← pill toggles, "1" disabled if Mexicano
  POINTS           (16) (24) (32)                     ← pill toggles
  ROUNDS           [ − 7 + ]                           ← Mexicano only, Stepper
  SCHEDULE         [tap to add ▾]                     ← expands to Calendar + hour/min Stepper

[Join code card]
  JOIN CODE
  [A][B][C][D]                                        ← 4-char code, one tile per char
  openpadel.app/s/abc123              [share/copy]

[Invite card]  ← admin only
  [🔍 Search or add player...]
  (search results → [Invite] buttons)
  ["name" as guest] dashed button, shown once typed and no match

[Players list]
  PLAYERS (n)
  [avatar] Name          👑(if creator) "you"(if self) [×](admin, remove)
  [avatar] Pending Name  ⏱ Invited                     ← greyed, pending invite

[Join form]  ← non-admin, not yet joined
  [Your name...]  [Join]

[Admin controls]                                       [Non-admin, already joined]
  [Start]  (disabled until canStart)                    "Waiting for the admin to start..."
  reason text if disabled (validation_errors or mode-specific hint)
  [Cancel session]  outlined, destructive on hover
  (dev-only) [Seed test players]
```

## Behavior

- **Inline config editing** — every config field (`name`, `game_mode`, `courts`, `points`,
  `rounds_total`, `scheduled_at`) patches the Session immediately on change via
  `api.sessions.update`, no separate "save" step. On error, local state reverts to the
  server's last-known values.
- **Mexicano min 2 courts** — switching to Mexicano bumps `courts` to 2 if it was 1; the "1"
  pill is disabled while Mexicano is selected. See `CLAUDE.md` invariants.
- **`canStart`** — prefers server-computed `session.can_start`; falls back to
  `activePlayers.length === requiredPlayers` (Mexicano, exact) or `>= requiredPlayers`
  (Americano, at-least). `requiredPlayers = courts * 4`.
- **Player add paths** — admin can search registered Users (debounced, 300ms, min 2 chars) and
  send an Invite, or add anyone by name as a Guest directly (no search match required).
  Non-admins add themselves via the join form.
- **Live updates** — player list and pending invites refresh on the `session_updated` SSE event
  (via `stream.onEvent`), not polling.
- **Dev-only seeding** — `import.meta.env.DEV` gate reveals a "seed test players" button that
  fills the roster with a fixed name list, up to `courts * 4 + 2`.

## Key Design Decisions

| Decision | Why |
|---|---|
| No "Save" button on config | Session is still being assembled — every admin action should be visible to joining players instantly over SSE |
| Guest add doesn't require a search miss | Admin-added guests are common (friends without the app) — the dashed "add as guest" affordance is always available once text is typed |
| `can_start` computed server-side, UI has a fallback | Keeps the Start-button gating logic in one place (`domain` constraints) while still rendering *something* reasonable if that field is ever absent |

## Open Questions

- [ ] Should the rules dialog (ⓘ) content differ between admin and player, or is the shared
      `rules_<mode>` i18n string sufficient long-term?
- [ ] Pending invites currently show with a static clock icon and no cancel/resend action —
      is that sufficient, or does an admin need a way to retract a sent invite?
