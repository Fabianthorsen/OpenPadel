-- name: CreateInvite :exec
INSERT INTO invites (id, session_id, from_user_id, to_user_id, status, created_at) VALUES (?, ?, ?, ?, 'pending', ?);

-- name: GetInvitesByUserID :many
SELECT
    i.id,
    i.session_id,
    s.name AS session_name,
    i.from_user_id,
    fu.display_name AS from_display_name,
    i.status,
    i.created_at
FROM invites i
JOIN sessions s ON s.id = i.session_id
JOIN users fu ON fu.id = i.from_user_id
WHERE i.to_user_id = ? AND i.status = 'pending'
ORDER BY i.created_at DESC;

-- name: GetPendingInvitesBySessionID :many
SELECT
    i.id,
    i.from_user_id,
    fu.display_name AS from_display_name,
    i.to_user_id,
    tu.display_name AS to_display_name,
    i.status,
    i.created_at
FROM invites i
JOIN users fu ON fu.id = i.from_user_id
JOIN users tu ON tu.id = i.to_user_id
WHERE i.session_id = ? AND i.status = 'pending'
ORDER BY i.created_at DESC;

-- name: GetInvite :one
SELECT
    i.id,
    i.session_id,
    s.name AS session_name,
    i.from_user_id,
    fu.display_name AS from_display_name,
    fu.avatar_icon AS from_avatar_icon,
    fu.avatar_color AS from_avatar_color,
    i.to_user_id,
    i.status,
    i.created_at
FROM invites i
JOIN sessions s ON s.id = i.session_id
JOIN users fu ON fu.id = i.from_user_id
WHERE i.id = ?;

-- name: UpdateInviteStatus :exec
UPDATE invites SET status = ? WHERE id = ?;

-- name: AcceptInvite :exec
INSERT INTO players (id, session_id, user_id, name, avatar_icon, avatar_color, active, joined_at) VALUES (?, ?, ?, ?, ?, ?, 1, ?);

-- name: DeleteInvite :exec
DELETE FROM invites WHERE id = ?;
