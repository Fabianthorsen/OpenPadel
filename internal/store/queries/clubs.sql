-- name: CreateClub :exec
INSERT INTO clubs (id, name, description, avatar_icon, avatar_color, join_code, created_by, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetClub :one
SELECT id, name, description, avatar_icon, avatar_color, join_code, created_by, created_at
FROM clubs WHERE id = ?;

-- name: GetClubByJoinCode :one
SELECT id, name, description, avatar_icon, avatar_color, join_code, created_by, created_at
FROM clubs WHERE join_code = ?;

-- name: GetUserClubs :many
SELECT DISTINCT c.id, c.name, c.description, c.avatar_icon, c.avatar_color, c.join_code, c.created_by, c.created_at
FROM clubs c
JOIN club_members cm ON cm.club_id = c.id
WHERE cm.user_id = ?
ORDER BY c.created_at DESC;

-- name: GetClubMembers :many
SELECT cm.user_id, cm.role, u.display_name, u.avatar_icon, u.avatar_color, cm.joined_at
FROM club_members cm
JOIN users u ON u.id = cm.user_id
WHERE cm.club_id = ?
ORDER BY cm.joined_at ASC;

-- name: GetClubMemberCount :one
SELECT COUNT(*) FROM club_members WHERE club_id = ?;

-- name: GetClubMember :one
SELECT cm.user_id, cm.role, cm.joined_at
FROM club_members cm
WHERE cm.club_id = ? AND cm.user_id = ?;

-- name: InsertClubMember :exec
INSERT INTO club_members (club_id, user_id, role, joined_at)
VALUES (?, ?, ?, ?);

-- name: UpdateClub :exec
UPDATE clubs SET name = ?, description = ?, avatar_icon = ?, avatar_color = ?
WHERE id = ?;

-- name: UpdateJoinCode :exec
UPDATE clubs SET join_code = ? WHERE id = ?;

-- name: DeleteClub :exec
DELETE FROM clubs WHERE id = ?;

-- name: DeleteClubMember :exec
DELETE FROM club_members WHERE club_id = ? AND user_id = ?;

-- name: UpdateClubMemberRole :exec
UPDATE club_members SET role = ? WHERE club_id = ? AND user_id = ?;

-- name: GetClubAdminCount :one
SELECT COUNT(*) FROM club_members WHERE club_id = ? AND role = 'admin';

-- name: GetClubEvents :many
SELECT
    s.id,
    CAST(COALESCE(NULLIF(s.name, ''), 'OpenPadel') AS TEXT) AS name,
    s.status,
    s.game_mode,
    s.courts,
    COUNT(p.id) AS player_count,
    s.scheduled_at
FROM sessions s
LEFT JOIN players p ON p.session_id = s.id AND p.active = 1
WHERE s.club_id = ? AND s.status IN ('lobby', 'playing')
GROUP BY s.id
ORDER BY COALESCE(s.scheduled_at, s.created_at) ASC;
