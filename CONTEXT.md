# OpenPadel

A mobile-first PWA for organizing padel games and tracking scores courtside.

## Language

**Session**:
A single padel event with a lifecycle (`lobby` → `playing` → `done`), a game mode, and a set of players.
_Avoid_: Tournament, game, event

**Player**:
A participant in a Session. May be a **Guest** (name only, no account) or linked to a registered **User** via `UserID`. Guest join is a deliberate simplicity — anyone can join a Session with just a name, no signup required.
_Avoid_: Participant, member

**Guest**:
A Player with no `UserID` — joined by name only, no account behind them.
_Avoid_: Anonymous player, unregistered player

**User**:
A registered account holder (email + password). Distinct from Player — a User only becomes a Player by joining a specific Session.
_Avoid_: Account, member

**Admin**:
Whoever holds a Session's `AdminToken` (a secret bearer token generated at creation). Admin status is the actual authorization gate for privileged actions (start, cancel, close, score entry, kick players) — checked via `isAdmin()`, independent of login state. A Session's join link without the token is read-only.
_Avoid_: Owner, moderator

**Creator**:
The User or Player who made the Session (`CreatorUserID` / `CreatorPlayerID`). Identity metadata only — used for "this is your session" UI (`IsCreator`) — not an authorization check. A Creator without the admin token cannot perform admin actions.
_Avoid_: Admin (these are related but distinct — see **Admin**)

**Game Mode**:
The pairing algorithm for a Session: `americano` (rotating partners, individual scoring) or `mexicano` (partners adapt based on standings, requires 2+ courts). A third mode, Timed Americano, was built and later removed entirely.
_Avoid_: Format, variant

**Round**:
One cycle of a Session — a set of Matches across all courts, plus the list of Players benched that cycle.

**Bench**:
Players sitting out a given Round. Rotation guarantees a Player benched in Round N plays in Round N+1.

**Match**:
One court's game within a Round — two teams of two Players, an optional final Score, and an optional in-progress Live score.

**Standing / Leaderboard**:
The ranked, live-updating view of a Session's Players by points — public, no auth required to view.
_Avoid_: Ranking, scoreboard

**Contact**:
A saved connection between two Users (added via search), used to streamline sending Invites.

**Invite**:
A request from one User to another to join a specific Session, with a `pending | accepted | declined` status. Distinct from the public join link — Invites target a known User, not an anonymous link.
