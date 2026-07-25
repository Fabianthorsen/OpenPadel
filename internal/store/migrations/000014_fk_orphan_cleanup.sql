-- +goose Up
-- +goose StatementBegin
-- Foreign-key enforcement was silently off until #249 (the modernc.org/sqlite
-- driver ignored the mattn-style `_foreign_keys=on` DSN param). Rows orphaned
-- while it was off never violate the pragma on their own, but they would trip a
-- PRAGMA foreign_key_check and defeat the cascades we now rely on. Clean them up
-- before enforcement matters. Session subtrees (players/rounds/matches/bench) are
-- omitted: DeleteSession has always removed them in dependency order, so they do
-- not orphan.

DELETE FROM auth_tokens WHERE user_id NOT IN (SELECT id FROM users);
DELETE FROM password_reset_tokens WHERE user_id NOT IN (SELECT id FROM users);
DELETE FROM push_subscriptions WHERE user_id NOT IN (SELECT id FROM users);
DELETE FROM contacts WHERE user_id NOT IN (SELECT id FROM users) OR contact_user_id NOT IN (SELECT id FROM users);
DELETE FROM invites WHERE from_user_id NOT IN (SELECT id FROM users) OR to_user_id NOT IN (SELECT id FROM users) OR session_id NOT IN (SELECT id FROM sessions);
DELETE FROM club_invites WHERE inviter_id NOT IN (SELECT id FROM users) OR invitee_id NOT IN (SELECT id FROM users) OR club_id NOT IN (SELECT id FROM clubs);
DELETE FROM club_members WHERE user_id NOT IN (SELECT id FROM users) OR club_id NOT IN (SELECT id FROM clubs);

UPDATE players SET user_id = NULL WHERE user_id IS NOT NULL AND user_id NOT IN (SELECT id FROM users);
UPDATE sessions SET creator_user_id = NULL WHERE creator_user_id IS NOT NULL AND creator_user_id NOT IN (SELECT id FROM users);
UPDATE sessions SET club_id = NULL WHERE club_id IS NOT NULL AND club_id NOT IN (SELECT id FROM clubs);

-- Clubs whose creator was deleted: hand created_by to a remaining member (Admins
-- first, then earliest joiner); delete the Club when no member remains. This is
-- the one-shot backfill of the same rule rehomeClub (users.go) applies going
-- forward — keep the successor ordering in sync if either changes.
UPDATE clubs SET created_by = (
    SELECT cm.user_id FROM club_members cm
    WHERE cm.club_id = clubs.id
    ORDER BY (cm.role = 'admin') DESC, cm.joined_at ASC
    LIMIT 1
)
WHERE created_by NOT IN (SELECT id FROM users)
  AND EXISTS (SELECT 1 FROM club_members cm WHERE cm.club_id = clubs.id);

DELETE FROM clubs WHERE created_by NOT IN (SELECT id FROM users);

-- +goose StatementEnd

-- +goose Down
-- Irreversible data cleanup — deleted orphan rows cannot be recovered.
