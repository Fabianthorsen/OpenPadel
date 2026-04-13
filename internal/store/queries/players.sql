-- name: CreatePlayer :exec
INSERT INTO players (id, session_id, user_id, name, avatar_icon, avatar_color, active, joined_at)
VALUES (?, ?, ?, ?, ?, ?, 1, ?);

-- name: GetPlayersBySessionID :many
SELECT id, session_id, COALESCE(user_id, ''), name, avatar_icon, avatar_color, active, joined_at
FROM players WHERE session_id = ? ORDER BY joined_at;

-- name: GetCreatorName :one
SELECT p.name FROM players p
JOIN sessions s ON s.creator_player_id = p.id
WHERE s.id = ?;

-- name: DeactivatePlayer :exec
UPDATE players SET active = 0 WHERE id = ?;

-- name: GetUserAvatarByUserID :one
SELECT avatar_icon, avatar_color FROM users WHERE id = ?;
