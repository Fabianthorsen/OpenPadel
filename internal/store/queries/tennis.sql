-- name: DeleteTennisTeamsBySessionID :exec
DELETE FROM tennis_teams WHERE session_id = ?;

-- name: InsertTennisTeam :exec
INSERT INTO tennis_teams (session_id, player_id, team) VALUES (?, ?, ?);

-- name: GetTennisTeamsBySessionID :many
SELECT tt.player_id, p.name, tt.team
FROM tennis_teams tt
JOIN players p ON p.id = tt.player_id
WHERE tt.session_id = ?;

-- name: CreateTennisMatch :exec
INSERT INTO tennis_matches (id, session_id, state, created_at, updated_at) VALUES (?, ?, ?, ?, ?);

-- name: GetTennisMatch :one
SELECT id, session_id, state, created_at, updated_at FROM tennis_matches WHERE session_id = ?;

-- name: UpdateTennisState :exec
UPDATE tennis_matches SET state = ?, updated_at = ? WHERE id = ?;
