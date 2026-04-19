-- name: CreateUser :exec
INSERT INTO users (id, email, display_name, avatar_icon, avatar_color, password_hash, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: UpdateProfile :exec
UPDATE users SET display_name = ?, avatar_icon = ?, avatar_color = ? WHERE id = ?;

-- name: UpdateUserPassword :exec
UPDATE users SET password_hash = ? WHERE id = ?;

-- name: UpdateProfileAvatarOnPlayers :exec
UPDATE players SET avatar_icon = ?, avatar_color = ? WHERE user_id = ?;

-- name: GetUserByEmail :one
SELECT id, email, display_name, avatar_icon, avatar_color, password_hash, created_at
FROM users WHERE email = ?;

-- name: GetUserByID :one
SELECT id, email, display_name, avatar_icon, avatar_color, password_hash, created_at
FROM users WHERE id = ?;

-- name: CreateAuthToken :exec
INSERT INTO auth_tokens (token, user_id, created_at) VALUES (?, ?, ?);

-- name: GetUserIDByToken :one
SELECT user_id FROM auth_tokens WHERE token = ?;

-- name: DeleteAuthToken :exec
DELETE FROM auth_tokens WHERE token = ?;

-- name: DeleteAuthTokensByUserID :exec
DELETE FROM auth_tokens WHERE user_id = ?;

-- name: GetAmericanoCareerStats :one
SELECT
    COUNT(DISTINCT p.session_id) AS tournaments,
    COUNT(m.id) AS games_played,
    CAST(COALESCE(SUM(
        CASE
            WHEN (m.p1 = p.id OR m.p2 = p.id) AND m.score_a > m.score_b THEN 1
            WHEN (m.p3 = p.id OR m.p4 = p.id) AND m.score_b > m.score_a THEN 1
            ELSE 0
        END
    ), 0) AS INTEGER) AS wins,
    CAST(COALESCE(SUM(
        CASE WHEN m.score_a = m.score_b THEN 1 ELSE 0 END
    ), 0) AS INTEGER) AS draws,
    CAST(COALESCE(SUM(
        CASE
            WHEN m.p1 = p.id OR m.p2 = p.id THEN m.score_a
            WHEN m.p3 = p.id OR m.p4 = p.id THEN m.score_b
            ELSE 0
        END
    ), 0) AS INTEGER) AS total_points
FROM players p
JOIN sessions s ON s.id = p.session_id AND s.status = 'complete' AND s.game_mode IN ('americano', 'timed_americano')
LEFT JOIN rounds r ON r.session_id = p.session_id
LEFT JOIN matches m ON m.round_id = r.id
    AND (m.p1 = p.id OR m.p2 = p.id OR m.p3 = p.id OR m.p4 = p.id)
    AND m.score_a IS NOT NULL
WHERE p.user_id = ? AND p.active = 1;

-- name: CreatePasswordResetToken :exec
INSERT INTO password_reset_tokens (token_hash, user_id, expires_at) VALUES (?, ?, ?);

-- name: DeletePasswordResetTokensByUserID :exec
DELETE FROM password_reset_tokens WHERE user_id = ?;

-- name: GetPasswordResetToken :one
SELECT user_id, expires_at FROM password_reset_tokens WHERE token_hash = ?;

-- name: DeletePasswordResetToken :exec
DELETE FROM password_reset_tokens WHERE token_hash = ?;

-- name: GetTournamentHistorySessions :many
SELECT
    s.id,
    CAST(COALESCE(NULLIF(s.name, ''), 'OpenPadel') AS TEXT) AS name,
    s.status,
    s.created_at,
    COALESCE(s.ended_early, 0) AS ended_early
FROM players p
JOIN sessions s ON s.id = p.session_id
WHERE p.user_id = ? AND p.active = 1 AND s.status = 'complete'
GROUP BY s.id
ORDER BY s.created_at DESC;

-- name: GetUpcomingTournaments :many
SELECT
    s.id,
    CAST(COALESCE(NULLIF(s.name, ''), 'OpenPadel') AS TEXT) AS name,
    s.status,
    s.game_mode,
    s.courts,
    COUNT(p2.id) AS player_count,
    s.scheduled_at
FROM players p
JOIN sessions s ON s.id = p.session_id
LEFT JOIN players p2 ON p2.session_id = s.id AND p2.active = 1
WHERE p.user_id = ? AND p.active = 1 AND s.status IN ('lobby', 'active')
GROUP BY s.id
ORDER BY s.status DESC, COALESCE(s.scheduled_at, s.created_at) ASC;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = ?;

-- name: UpdatePlayerUserIDToNull :exec
UPDATE players SET user_id = NULL WHERE user_id = ?;

-- name: IncrementTournamentWinCount :exec
UPDATE users SET win_count = win_count + 1 WHERE id = ?;
