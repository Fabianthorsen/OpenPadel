-- name: CreateClubInvite :exec
INSERT INTO club_invites (id, club_id, inviter_id, invitee_id, status, created_at)
VALUES (?, ?, ?, ?, 'pending', ?);

-- name: GetClubInvitesByInviteeID :many
SELECT
    ci.id,
    ci.club_id,
    c.name AS club_name,
    c.avatar_icon AS club_avatar_icon,
    c.avatar_color AS club_avatar_color,
    ci.inviter_id,
    iu.display_name AS inviter_display_name,
    ci.status,
    ci.created_at
FROM club_invites ci
JOIN clubs c ON c.id = ci.club_id
JOIN users iu ON iu.id = ci.inviter_id
WHERE ci.invitee_id = ? AND ci.status = 'pending'
ORDER BY ci.created_at DESC;

-- name: GetClubInvite :one
SELECT
    ci.id,
    ci.club_id,
    c.name AS club_name,
    c.avatar_icon AS club_avatar_icon,
    c.avatar_color AS club_avatar_color,
    ci.inviter_id,
    iu.display_name AS inviter_display_name,
    ci.invitee_id,
    ci.status,
    ci.created_at
FROM club_invites ci
JOIN clubs c ON c.id = ci.club_id
JOIN users iu ON iu.id = ci.inviter_id
WHERE ci.id = ?;

-- name: UpdateClubInviteStatus :exec
UPDATE club_invites SET status = ? WHERE id = ?;

-- name: ResetClubInvite :exec
UPDATE club_invites SET status = 'pending', inviter_id = ? WHERE id = ?;

-- name: GetClubInviteByClubAndInvitee :one
SELECT id, status FROM club_invites WHERE club_id = ? AND invitee_id = ?;
