# ADR 0008: Enforce SQLite Foreign Keys

## Status

Accepted

## Context

Foreign-key enforcement was **off** across the whole application until now. Both open sites —
`internal/store/store.go` and `cmd/migrate/main.go` — opened SQLite with the **mattn/go-sqlite3**
DSN syntax:

```go
sql.Open("sqlite", path+"?_journal_mode=WAL&_foreign_keys=on")
```

The project uses **modernc.org/sqlite**, which silently ignores those params. At runtime
`PRAGMA foreign_keys` read **0** and — separately — WAL journaling was never applied either. The
consequences (see #249):

- No `ON DELETE CASCADE` / `ON DELETE SET NULL` action ever fired. Deleting a parent left orphaned
  children and dangling references.
- FK constraints weren't checked on insert, so a row could reference a non-existent parent.
- Every hand-rolled cascade in the store existed only to paper over the inert `ON DELETE` clauses
  (e.g. the original `Store.DeleteClub` unwound `club_invites` / `club_members` and NULLed
  `sessions.club_id` by hand).

Flipping enforcement on an existing database is hard to reverse, so it warranted its own change and
this record: behaviour shifts for every table at once, and delete paths that silently left orphans
now either cascade or fail loudly.

## Decision

**Enforce foreign keys (and WAL) on every connection, using modernc's pragma DSN syntax.**

```go
sql.Open("sqlite", path+"?_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)")
```

The pragma form applies to every pooled connection (the pool is capped at one connection anyway).
`Store.Open` asserts `PRAGMA foreign_keys == 1` after migrating and refuses to start otherwise, so a
future DSN regression fails fast instead of silently degrading.

### Consequences for delete paths

With enforcement on, each delete path was audited and made FK-safe:

- **`DeleteClub`** now leans on the schema: a single `DELETE FROM clubs` cascades to `club_members`
  and `club_invites` and detaches Sessions via `sessions.club_id ON DELETE SET NULL`. The manual
  transaction was removed.
- **`DeleteSession`** deletes its subtree (bench/matches → rounds → players → session) in one
  transaction with error checking; pending session invites cascade via `invites.session_id`.
- **`DeleteUser`** (account deletion) is the invasive case. Several parents reference `users(id)`
  with no cascade (`clubs.created_by`, `club_members`, `club_invites`, `password_reset_tokens`), so
  a naive delete now fails. Account deletion therefore:
  - clears the creator label on Sessions (`creator_user_id` → NULL) and detaches Player rows
    (`user_id` → NULL) so history outlives the account;
  - deletes auth tokens, reset tokens and any remaining Club invites; `push_subscriptions`,
    `contacts` and session invites cascade;
  - **rehomes every Club the User owns or belongs to** rather than leaving it orphaned.

### Club ownership on account deletion

`clubs.created_by` is `NOT NULL` and identity metadata only (authorization is the `admin` role in
`club_members`, never `created_by`). When a User deletes their account, for each Club they touch:

- **No other member remains** → the Club is deleted (members and invites cascade).
- **Other members remain** → the Club survives: ownership (`created_by`) and, if needed, the Admin
  role pass to a remaining member (Admins first, then earliest joiner). The departing User's
  membership is removed.

This mirrors how `sessions.creator_user_id` is treated (the label detaches, the record survives) and
keeps a Club from ever being left ownerless or adminless, without a `NOT NULL`-dropping table
rebuild.

### Orphan cleanup

Rows orphaned while enforcement was off don't violate the pragma on their own but would trip
`PRAGMA foreign_key_check` and defeat the cascades we now rely on. Migration
`000014_fk_orphan_cleanup.sql` removes orphaned tokens / memberships / invites, NULLs nullable
dangling references, and rehomes-or-deletes Clubs whose creator was already gone. Session subtrees
are omitted: `DeleteSession` has always removed them in dependency order, so they don't orphan.

## Alternatives considered

- **Leave enforcement off, keep hand-rolling cascades.** The status quo. Rejected: every new
  relationship needs a bespoke unwind, and a missed one silently orphans data. The declared
  `ON DELETE` clauses already express intent — enforce them.
- **Make `clubs.created_by` nullable and NULL it on owner deletion.** Faithful to "reassign, else
  NULL", but SQLite can't drop a `NOT NULL` constraint without a full table rebuild, and a
  NULL-owner, zero-member Club is a useless ghost. Deleting the Club when no member remains reaches
  the same end state without the rebuild.
- **Block account deletion while the User is a sole Club Admin.** Mirrors `guardLastAdmin`, but traps
  the User until they hand off or delete Clubs manually — poor UX for an irreversible action.

## References

- Issue #249
- `internal/store/store.go`, `cmd/migrate/main.go` — DSN + startup assertion
- `internal/store/users.go` — `DeleteUser`, `rehomeClubsForUserDelete`
- `internal/store/clubs.go` — simplified `DeleteClub`
- `internal/store/sessions.go` — transactional `DeleteSession`
- `internal/store/migrations/000014_fk_orphan_cleanup.sql`
