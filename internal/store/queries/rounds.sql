-- name: InsertRound :exec
INSERT INTO rounds (id, session_id, number) VALUES (?, ?, ?);

-- name: InsertBench :exec
INSERT INTO bench (round_id, player_id) VALUES (?, ?);

-- name: InsertMatch :exec
INSERT INTO matches (id, round_id, court, p1, p2, p3, p4) VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: GetRoundsBySessionID :many
SELECT id, number FROM rounds WHERE session_id = ? ORDER BY number;

-- name: GetCurrentRoundBySessionID :one
SELECT r.id, r.number FROM rounds r
JOIN sessions sess ON sess.id = r.session_id
WHERE r.session_id = ? AND r.number = sess.current_round;

-- name: GetLastRoundBySessionID :one
SELECT id, number FROM rounds WHERE session_id = ? ORDER BY number DESC LIMIT 1;

-- name: GetMatchByID :one
SELECT id, round_id, court, p1, p2, p3, p4, score_a, score_b, live_a, live_b, server FROM matches WHERE id = ?;

-- name: UpdateMatchScore :exec
UPDATE matches SET score_a = ?, score_b = ?, live_a = NULL, live_b = NULL WHERE id = ?;

-- name: UpdateMatchLiveScore :exec
UPDATE matches SET live_a = ?, live_b = ?, server = ? WHERE id = ?;

-- name: GetLeaderboard :many
SELECT
    p.id,
    p.user_id,
    p.name,
    COALESCE(SUM(CASE WHEN m.p1 = p.id OR m.p2 = p.id THEN m.score_a WHEN m.p3 = p.id OR m.p4 = p.id THEN m.score_b ELSE 0 END), 0) AS points,
    COALESCE(SUM(CASE WHEN m.p1 = p.id OR m.p2 = p.id THEN m.score_b WHEN m.p3 = p.id OR m.p4 = p.id THEN m.score_a ELSE 0 END), 0) AS points_conceded,
    COUNT(m.id) AS games_played,
    COALESCE(SUM(CASE WHEN (m.p1 = p.id OR m.p2 = p.id) AND m.score_a > m.score_b THEN 1 WHEN (m.p3 = p.id OR m.p4 = p.id) AND m.score_b > m.score_a THEN 1 ELSE 0 END), 0) AS wins,
    COALESCE(SUM(CASE WHEN m.score_a = m.score_b THEN 1 ELSE 0 END), 0) AS draws,
    p.avatar_icon,
    p.avatar_color
FROM players p
LEFT JOIN rounds r ON r.session_id = p.session_id
LEFT JOIN matches m ON m.round_id = r.id
    AND (m.p1 = p.id OR m.p2 = p.id OR m.p3 = p.id OR m.p4 = p.id)
    AND m.score_a IS NOT NULL
WHERE p.session_id = ? AND p.active = 1
GROUP BY p.id, p.name;

-- name: GetHeadToHead :many
SELECT m.p1, m.p2, m.p3, m.p4, m.score_a, m.score_b
FROM matches m
JOIN rounds r ON r.id = m.round_id
WHERE r.session_id = ? AND m.score_a IS NOT NULL;

-- name: AdvanceMexicanoRound :exec
INSERT INTO rounds (id, session_id, number) VALUES (?, ?, ?);

-- name: UpdateSessionCurrentRound :exec
UPDATE sessions SET current_round = ?, updated_at = ? WHERE id = ?;

-- name: CountUnscored :one
SELECT COUNT(*) FROM matches m
JOIN rounds r ON r.id = m.round_id
WHERE r.session_id = ? AND m.score_a IS NULL;

-- name: GetBenchByRoundID :many
SELECT player_id FROM bench WHERE round_id = ?;

-- name: GetMatchesByRoundID :many
SELECT id, round_id, court, p1, p2, p3, p4, score_a, score_b, live_a, live_b, server FROM matches WHERE round_id = ? ORDER BY court;
