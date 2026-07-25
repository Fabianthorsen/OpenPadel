-- name: CreateUser :exec
INSERT INTO users (id, email, display_name, avatar_icon, avatar_color, password_hash, self_rating, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateUserSelfRating :exec
UPDATE users SET self_rating = ? WHERE id = ?;

-- name: UpdateProfile :exec
UPDATE users SET display_name = ?, avatar_icon = ?, avatar_color = ? WHERE id = ?;

-- name: UpdateUserPassword :exec
UPDATE users SET password_hash = ? WHERE id = ?;

-- name: UpdateProfileAvatarOnPlayers :exec
UPDATE players SET avatar_icon = ?, avatar_color = ? WHERE user_id = ?;

-- name: GetUserByEmail :one
SELECT id, email, display_name, avatar_icon, avatar_color, password_hash, self_rating, created_at
FROM users WHERE email = ?;

-- name: GetUserByID :one
SELECT id, email, display_name, avatar_icon, avatar_color, password_hash, self_rating, created_at
FROM users WHERE id = ?;

-- name: CreateAuthToken :exec
INSERT INTO auth_tokens (token_hash, user_id, created_at, expires_at) VALUES (?, ?, ?, ?);

-- name: GetAuthTokenByHash :one
SELECT user_id, expires_at FROM auth_tokens WHERE token_hash = ?;

-- name: UpdateAuthTokenExpiry :exec
UPDATE auth_tokens SET expires_at = ? WHERE token_hash = ?;

-- name: DeleteAuthToken :exec
DELETE FROM auth_tokens WHERE token_hash = ?;

-- name: DeleteAuthTokensByUserID :exec
DELETE FROM auth_tokens WHERE user_id = ?;

-- name: GetCareerSummary :one
-- Cross-mode career summary: games played, wins, and mean per-Match point-share.
-- point_share is the average of yourTeamScore / (yourTeamScore + opponentTeamScore)
-- over every fully-scored Match in a done Session, each match weighted equally
-- regardless of its points target (ADR 0007). Guests fill seats but only the
-- user's own player rows are aggregated.
SELECT
    COUNT(m.id) AS games,
    CAST(COALESCE(SUM(
        CASE
            WHEN (m.p1 = p.id OR m.p2 = p.id) AND m.score_a > m.score_b THEN 1
            WHEN (m.p3 = p.id OR m.p4 = p.id) AND m.score_b > m.score_a THEN 1
            ELSE 0
        END
    ), 0) AS INTEGER) AS wins,
    CAST(COALESCE(AVG(
        CASE
            WHEN (m.score_a + m.score_b) = 0 THEN 0.5
            WHEN (m.p1 = p.id OR m.p2 = p.id) THEN CAST(m.score_a AS REAL) / (m.score_a + m.score_b)
            ELSE CAST(m.score_b AS REAL) / (m.score_a + m.score_b)
        END
    ), 0.0) AS REAL) AS point_share
FROM players p
JOIN sessions s ON s.id = p.session_id AND s.status = 'done'
LEFT JOIN rounds r ON r.session_id = p.session_id
LEFT JOIN matches m ON m.round_id = r.id
    AND (m.p1 = p.id OR m.p2 = p.id OR m.p3 = p.id OR m.p4 = p.id)
    AND m.score_a IS NOT NULL
WHERE p.user_id = ? AND p.active = 1;

-- name: GetModeStats :many
-- Per-Game-Mode career aggregate for the Career Stats page (ADR 0007). One row
-- per Game Mode the user has a done Session in; modes with no history are absent
-- and filled in by the store. point_share is the mean per-Match share (same as
-- GetCareerSummary) but grouped by mode, since point-share only compares like-for
-- -like within one scoring model. total_points / points_conceded are the user's
-- own-team and opponent-team score totals across fully-scored matches. tournaments
-- counts distinct done Sessions the user joined in that mode. Guests fill seats
-- but only the user's own player rows are aggregated.
SELECT
    s.game_mode AS mode,
    COUNT(DISTINCT p.session_id) AS tournaments,
    COUNT(m.id) AS games,
    CAST(COALESCE(SUM(
        CASE
            WHEN (m.p1 = p.id OR m.p2 = p.id) AND m.score_a > m.score_b THEN 1
            WHEN (m.p3 = p.id OR m.p4 = p.id) AND m.score_b > m.score_a THEN 1
            ELSE 0
        END
    ), 0) AS INTEGER) AS wins,
    CAST(COALESCE(SUM(
        CASE WHEN m.id IS NOT NULL AND m.score_a = m.score_b THEN 1 ELSE 0 END
    ), 0) AS INTEGER) AS draws,
    CAST(COALESCE(SUM(
        CASE
            WHEN m.p1 = p.id OR m.p2 = p.id THEN m.score_a
            WHEN m.p3 = p.id OR m.p4 = p.id THEN m.score_b
            ELSE 0
        END
    ), 0) AS INTEGER) AS total_points,
    CAST(COALESCE(SUM(
        CASE
            WHEN m.p1 = p.id OR m.p2 = p.id THEN m.score_b
            WHEN m.p3 = p.id OR m.p4 = p.id THEN m.score_a
            ELSE 0
        END
    ), 0) AS INTEGER) AS points_conceded,
    CAST(COALESCE(AVG(
        CASE
            WHEN (m.score_a + m.score_b) = 0 THEN 0.5
            WHEN (m.p1 = p.id OR m.p2 = p.id) THEN CAST(m.score_a AS REAL) / (m.score_a + m.score_b)
            ELSE CAST(m.score_b AS REAL) / (m.score_a + m.score_b)
        END
    ), 0.0) AS REAL) AS point_share
FROM players p
JOIN sessions s ON s.id = p.session_id AND s.status = 'done'
LEFT JOIN rounds r ON r.session_id = p.session_id
LEFT JOIN matches m ON m.round_id = r.id
    AND (m.p1 = p.id OR m.p2 = p.id OR m.p3 = p.id OR m.p4 = p.id)
    AND m.score_a IS NOT NULL
WHERE p.user_id = ? AND p.active = 1
GROUP BY s.game_mode;

-- name: GetMatchResultsSeries :many
-- Per-Match results series for the Career Stats page's recent-form curve
-- (ADR 0007). One row per fully-scored Match the user played in a done Session,
-- oldest-first (by Session date, then round, then court) so the client reads it
-- as a time series. INNER JOINs on rounds/matches keep out ended-early Sessions
-- with no scored Match. points / conceded are the user's own-team and opponent
-- -team score for that Match; the win/draw/loss outcome is derived in Go. date is
-- the Session date (Matches carry no timestamp of their own). Guests fill seats
-- but only the user's player rows are aggregated.
SELECT
    m.id AS match_id,
    s.game_mode AS mode,
    s.created_at AS date,
    CAST(CASE
        WHEN m.p1 = p.id OR m.p2 = p.id THEN m.score_a
        ELSE m.score_b
    END AS INTEGER) AS points,
    CAST(CASE
        WHEN m.p1 = p.id OR m.p2 = p.id THEN m.score_b
        ELSE m.score_a
    END AS INTEGER) AS conceded
FROM players p
JOIN sessions s ON s.id = p.session_id AND s.status = 'done'
JOIN rounds r ON r.session_id = p.session_id
JOIN matches m ON m.round_id = r.id
    AND (m.p1 = p.id OR m.p2 = p.id OR m.p3 = p.id OR m.p4 = p.id)
    AND m.score_a IS NOT NULL
WHERE p.user_id = ? AND p.active = 1
ORDER BY s.created_at ASC, s.rowid ASC, r.number ASC, m.court ASC;

-- name: CreatePasswordResetToken :exec
INSERT INTO password_reset_tokens (token_hash, user_id, expires_at) VALUES (?, ?, ?);

-- name: DeletePasswordResetTokensByUserID :exec
DELETE FROM password_reset_tokens WHERE user_id = ?;

-- name: GetPasswordResetToken :one
SELECT user_id, expires_at FROM password_reset_tokens WHERE token_hash = ?;

-- name: DeletePasswordResetToken :exec
DELETE FROM password_reset_tokens WHERE token_hash = ?;

-- name: GetTournamentHistorySessions :many
-- Done Sessions the user played, newest first, for both the tournament history
-- timeline and the cross-mode placement stats (ADR 0007). scored is 1 when the
-- user has at least one fully-scored Match in the Session, so placement can
-- ignore ended-early Sessions the user never actually finished a game in; the
-- finishing rank itself is resolved from the leaderboard in Go.
-- name is returned raw (may be empty); the display fallback (club-aware) lives
-- in the frontend sessionName() so there is a single source of truth.
SELECT
    s.id,
    CAST(s.name AS TEXT) AS name,
    s.game_mode,
    CAST(COALESCE(c.name, '') AS TEXT) AS club_name,
    s.status,
    s.created_at,
    COALESCE(s.ended_early, 0) AS ended_early,
    CAST(MAX(CASE WHEN m.score_a IS NOT NULL THEN 1 ELSE 0 END) AS INTEGER) AS scored
FROM players p
JOIN sessions s ON s.id = p.session_id
LEFT JOIN clubs c ON c.id = s.club_id
LEFT JOIN rounds r ON r.session_id = p.session_id
LEFT JOIN matches m ON m.round_id = r.id
    AND (m.p1 = p.id OR m.p2 = p.id OR m.p3 = p.id OR m.p4 = p.id)
WHERE p.user_id = ? AND p.active = 1 AND s.status = 'done'
GROUP BY s.id
ORDER BY s.created_at DESC;

-- name: GetUpcomingTournaments :many
-- name is returned raw (may be empty); the club-aware display fallback lives in
-- the frontend sessionName().
SELECT
    s.id,
    CAST(s.name AS TEXT) AS name,
    CAST(COALESCE(c.name, '') AS TEXT) AS club_name,
    s.status,
    s.game_mode,
    s.courts,
    COUNT(p2.id) AS player_count,
    s.scheduled_at
FROM players p
JOIN sessions s ON s.id = p.session_id
LEFT JOIN clubs c ON c.id = s.club_id
LEFT JOIN players p2 ON p2.session_id = s.id AND p2.active = 1
WHERE p.user_id = ? AND p.active = 1 AND s.status IN ('lobby', 'playing')
GROUP BY s.id
ORDER BY s.status DESC, COALESCE(s.scheduled_at, s.created_at) ASC;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = ?;

-- name: UpdatePlayerUserIDToNull :exec
UPDATE players SET user_id = NULL WHERE user_id = ?;

-- name: IncrementTournamentWinCount :exec
UPDATE users SET win_count = win_count + 1 WHERE id = ?;
