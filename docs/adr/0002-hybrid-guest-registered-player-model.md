# ADR 0002: Hybrid guest/registered Player model

## Status

Accepted

## Context

Every Session has Players, and every Player either carries a `UserID` (a registered, logged-in
account) or is a Guest — name only, no account, no login. The public join link lets anyone become
a Player by typing a name, with no signup step.

This could instead have required an account to join any Session, which would give every Player a
durable identity, cross-session stats, and a simpler auth model (one kind of Player, not two).

## Decision

Keep guest join as a permanent, first-class path, not a legacy affordance to phase out. In the
user's words: "one of the simplicities of the app is the ability to just join without having an
account." A Player only becomes tied to a User when someone who's logged in joins — guest and
registered Players are otherwise treated identically everywhere a Player appears (rounds, scoring,
leaderboard).

## Consequences

- Every Player-touching code path (scoring, leaderboard, bench rotation) must work with a
  `UserID`-less Player — there's no "assume a User" shortcut anywhere in the Session lifecycle.
- Guests have no cross-session identity: stats, contacts, and invites only exist for registered
  Users. A guest who plays five Sessions has five disconnected Player records, not one history.
- Friction reduction wins over completeness here — this is deliberate, not a gap to close. Don't
  propose "require an account to join" as a simplification; it removes the feature this ADR exists
  to protect.
- See **Player**, **Guest**, **User** in `CONTEXT.md` for the exact terminology.
