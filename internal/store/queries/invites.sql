-- name: CreateInvite :exec
INSERT INTO invites (id, session_id, from_user_id, to_user_id, status, created_at) VALUES (?, ?, ?, ?, 'pending', ?);

-- name: GetInvitesByUserID :many
SELECT id, session_id, from_user_id, status, created_at FROM invites WHERE to_user_id = ? AND status = 'pending' ORDER BY created_at DESC;

-- name: GetPendingInvitesBySessionID :many
SELECT id, from_user_id, to_user_id, status, created_at FROM invites WHERE session_id = ? AND status = 'pending' ORDER BY created_at DESC;

-- name: GetInvite :one
SELECT id, session_id, from_user_id, to_user_id, status FROM invites WHERE id = ?;

-- name: UpdateInviteStatus :exec
UPDATE invites SET status = ? WHERE id = ?;

-- name: AcceptInvite :exec
INSERT INTO players (id, session_id, user_id, name, avatar_icon, avatar_color, active, joined_at) VALUES (?, ?, ?, ?, ?, ?, 1, ?);

-- name: DeleteInvite :exec
DELETE FROM invites WHERE id = ?;
