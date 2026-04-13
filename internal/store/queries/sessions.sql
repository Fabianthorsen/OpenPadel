-- name: CreateSession :exec
INSERT INTO sessions (id, admin_token, status, name, game_mode, sets_to_win, games_per_set, courts, points, rounds_total, scheduled_at, court_duration_minutes, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetSession :one
SELECT id, admin_token, status, name, game_mode, sets_to_win, games_per_set, courts, points, rounds_total, creator_player_id, current_round, scheduled_at, court_duration_minutes, ends_at, created_at, updated_at
FROM sessions WHERE id = ?;

-- name: SetCreatorPlayer :exec
UPDATE sessions SET creator_player_id = ?, updated_at = ? WHERE id = ?;

-- name: StartSession :exec
UPDATE sessions SET status = ?, rounds_total = ?, current_round = 1, ends_at = ?, updated_at = ? WHERE id = ?;

-- name: StartMexicanoSession :exec
UPDATE sessions SET status = ?, current_round = 1, ends_at = ?, updated_at = ? WHERE id = ?;

-- name: AdvanceRound :exec
UPDATE sessions SET current_round = current_round + 1, updated_at = ? WHERE id = ?;

-- name: CurrentRoundAllScored :one
SELECT COUNT(*) FROM matches m
JOIN rounds r ON r.id = m.round_id
JOIN sessions s ON s.id = r.session_id
WHERE s.id = ? AND r.number = s.current_round AND m.score_a IS NULL;

-- name: CompleteSession :exec
UPDATE sessions SET status = ?, ended_early = ?, updated_at = ? WHERE id = ?;

-- name: DeleteTennisMatches :exec
DELETE FROM tennis_matches WHERE session_id = ?;

-- name: DeleteTennisTeams :exec
DELETE FROM tennis_teams WHERE session_id = ?;

-- name: DeleteBench :exec
DELETE FROM bench WHERE round_id IN (SELECT id FROM rounds WHERE session_id = ?);

-- name: DeleteMatches :exec
DELETE FROM matches WHERE round_id IN (SELECT id FROM rounds WHERE session_id = ?);

-- name: DeleteRounds :exec
DELETE FROM rounds WHERE session_id = ?;

-- name: DeletePlayers :exec
DELETE FROM players WHERE session_id = ?;

-- name: DeleteSession :exec
DELETE FROM sessions WHERE id = ?;
