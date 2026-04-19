-- +goose Up
DROP TABLE IF EXISTS tennis_matches;
DROP TABLE IF EXISTS tennis_teams;

-- +goose Down
-- (intentionally empty — no rollback)
