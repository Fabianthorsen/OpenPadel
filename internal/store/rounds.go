package store

import (
	"database/sql"
	"errors"

	"github.com/fabianthorsen/nottennis/internal/domain"
)

func (s *Store) SaveRounds(sessionID string, rounds []domain.Round) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, r := range rounds {
		if _, err := tx.Exec(
			`INSERT INTO rounds (id, session_id, number) VALUES (?, ?, ?)`,
			r.ID, sessionID, r.Number,
		); err != nil {
			return err
		}
		for _, pid := range r.Bench {
			if _, err := tx.Exec(
				`INSERT INTO bench (round_id, player_id) VALUES (?, ?)`,
				r.ID, pid,
			); err != nil {
				return err
			}
		}
		for _, m := range r.Matches {
			if _, err := tx.Exec(
				`INSERT INTO matches (id, round_id, court, p1, p2, p3, p4) VALUES (?, ?, ?, ?, ?, ?, ?)`,
				m.ID, r.ID, m.Court, m.TeamA[0], m.TeamA[1], m.TeamB[0], m.TeamB[1],
			); err != nil {
				return err
			}
		}
	}
	return tx.Commit()
}

func (s *Store) GetRounds(sessionID string) ([]domain.Round, error) {
	rows, err := s.db.Query(
		`SELECT id, number FROM rounds WHERE session_id = ? ORDER BY number`,
		sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rounds []domain.Round
	for rows.Next() {
		var r domain.Round
		r.SessionID = sessionID
		if err := rows.Scan(&r.ID, &r.Number); err != nil {
			return nil, err
		}
		rounds = append(rounds, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i := range rounds {
		rounds[i].Bench, err = s.getBench(rounds[i].ID)
		if err != nil {
			return nil, err
		}
		rounds[i].Matches, err = s.getMatches(rounds[i].ID)
		if err != nil {
			return nil, err
		}
	}
	return rounds, nil
}

func (s *Store) GetCurrentRound(sessionID string) (*domain.Round, error) {
	// Use the session's tracked current_round number.
	// Falls back to auto-detect for sessions created before this field existed.
	var r domain.Round
	r.SessionID = sessionID
	err := s.db.QueryRow(`
		SELECT r.id, r.number FROM rounds r
		JOIN sessions sess ON sess.id = r.session_id
		WHERE r.session_id = ? AND r.number = sess.current_round`,
		sessionID,
	).Scan(&r.ID, &r.Number)
	if errors.Is(err, sql.ErrNoRows) {
		// Fallback: return last round (handles legacy sessions or completed sessions)
		if err2 := s.db.QueryRow(
			`SELECT id, number FROM rounds WHERE session_id = ? ORDER BY number DESC LIMIT 1`,
			sessionID,
		).Scan(&r.ID, &r.Number); errors.Is(err2, sql.ErrNoRows) {
			return nil, ErrNotFound
		} else if err2 != nil {
			return nil, err2
		}
	} else if err != nil {
		return nil, err
	}

	r.Bench, err = s.getBench(r.ID)
	if err != nil {
		return nil, err
	}
	r.Matches, err = s.getMatches(r.ID)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *Store) GetMatch(matchID string) (*domain.Match, error) {
	row := s.db.QueryRow(
		`SELECT id, round_id, court, p1, p2, p3, p4, score_a, score_b, live_a, live_b, server FROM matches WHERE id = ?`,
		matchID,
	)
	return scanMatch(row)
}

func (s *Store) UpdateScore(matchID string, scoreA, scoreB int) (*domain.Match, error) {
	_, err := s.db.Exec(
		`UPDATE matches SET score_a = ?, score_b = ?, live_a = NULL, live_b = NULL WHERE id = ?`,
		scoreA, scoreB, matchID,
	)
	if err != nil {
		return nil, err
	}
	return s.GetMatch(matchID)
}

func (s *Store) UpdateLiveScore(matchID, server string, a, b int) error {
	_, err := s.db.Exec(
		`UPDATE matches SET live_a = ?, live_b = ?, server = ? WHERE id = ?`,
		a, b, server, matchID,
	)
	return err
}

func (s *Store) GetLeaderboard(sessionID string) ([]domain.Standing, error) {
	rows, err := s.db.Query(`
		SELECT
			p.id,
			p.user_id,
			p.name,
			COALESCE(SUM(
				CASE
					WHEN m.p1 = p.id OR m.p2 = p.id THEN m.score_a
					WHEN m.p3 = p.id OR m.p4 = p.id THEN m.score_b
					ELSE 0
				END
			), 0) AS points,
			COUNT(m.id) AS games_played,
			COALESCE(SUM(
				CASE
					WHEN (m.p1 = p.id OR m.p2 = p.id) AND m.score_a > m.score_b THEN 1
					WHEN (m.p3 = p.id OR m.p4 = p.id) AND m.score_b > m.score_a THEN 1
					ELSE 0
				END
			), 0) AS wins,
			COALESCE(SUM(
				CASE WHEN m.score_a = m.score_b THEN 1 ELSE 0 END
			), 0) AS draws
		FROM players p
		LEFT JOIN rounds r ON r.session_id = p.session_id
		LEFT JOIN matches m ON m.round_id = r.id
			AND (m.p1 = p.id OR m.p2 = p.id OR m.p3 = p.id OR m.p4 = p.id)
			AND m.score_a IS NOT NULL
		WHERE p.session_id = ? AND p.active = 1
		GROUP BY p.id, p.name
		ORDER BY points DESC`,
		sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var standings []domain.Standing
	rank := 1
	for rows.Next() {
		var s domain.Standing
		if err := rows.Scan(&s.PlayerID, &s.UserID, &s.Name, &s.Points, &s.GamesPlayed, &s.Wins, &s.Draws); err != nil {
			return nil, err
		}
		s.Rank = rank
		rank++
		standings = append(standings, s)
	}
	if standings == nil {
		standings = []domain.Standing{}
	}
	return standings, rows.Err()
}

// AllRoundsComplete returns true when every match in the session has a score.
func (s *Store) AllRoundsComplete(sessionID string) (bool, error) {
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(*) FROM matches m
		JOIN rounds r ON r.id = m.round_id
		WHERE r.session_id = ? AND m.score_a IS NULL`,
		sessionID,
	).Scan(&count)
	return count == 0, err
}

func (s *Store) getBench(roundID string) ([]string, error) {
	rows, err := s.db.Query(`SELECT player_id FROM bench WHERE round_id = ?`, roundID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if ids == nil {
		ids = []string{}
	}
	return ids, rows.Err()
}

func (s *Store) getMatches(roundID string) ([]domain.Match, error) {
	rows, err := s.db.Query(
		`SELECT id, round_id, court, p1, p2, p3, p4, score_a, score_b, live_a, live_b, server FROM matches WHERE round_id = ? ORDER BY court`,
		roundID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var matches []domain.Match
	for rows.Next() {
		row := &singleRow{}
		if err := rows.Scan(
			&row.vals[0], &row.vals[1], &row.vals[2], &row.vals[3], &row.vals[4],
			&row.vals[5], &row.vals[6], &row.vals[7], &row.vals[8],
			&row.vals[9], &row.vals[10], &row.vals[11],
		); err != nil {
			return nil, err
		}
		m, err := scanMatch(row)
		if err != nil {
			return nil, err
		}
		matches = append(matches, *m)
	}
	if matches == nil {
		matches = []domain.Match{}
	}
	return matches, rows.Err()
}

// singleRow lets us reuse scanMatch for both sql.Row and sql.Rows.
type singleRow struct {
	vals [12]any
	idx  int
}

func (r *singleRow) Scan(dest ...any) error {
	for i, d := range dest {
		switch v := d.(type) {
		case *string:
			if s, ok := r.vals[i].(string); ok {
				*v = s
			}
		case *int:
			switch n := r.vals[i].(type) {
			case int64:
				*v = int(n)
			}
		case *sql.NullInt64:
			switch n := r.vals[i].(type) {
			case int64:
				v.Int64 = n
				v.Valid = true
			case nil:
				v.Valid = false
			}
		case *sql.NullString:
			switch sv := r.vals[i].(type) {
			case string:
				v.String = sv
				v.Valid = true
			case nil:
				v.Valid = false
			}
		}
	}
	return nil
}

func scanMatch(s interface {
	Scan(...any) error
}) (*domain.Match, error) {
	var m domain.Match
	var scoreA, scoreB sql.NullInt64
	var liveA, liveB sql.NullInt64
	var server sql.NullString
	if err := s.Scan(
		&m.ID, &m.RoundID, &m.Court,
		&m.TeamA[0], &m.TeamA[1], &m.TeamB[0], &m.TeamB[1],
		&scoreA, &scoreB, &liveA, &liveB, &server,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if scoreA.Valid {
		m.Score = &domain.Score{A: int(scoreA.Int64), B: int(scoreB.Int64)}
	}
	if liveA.Valid {
		m.Live = &domain.LiveScore{A: int(liveA.Int64), B: int(liveB.Int64), Server: server.String}
	}
	return &m, nil
}
