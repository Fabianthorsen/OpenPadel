package store

import (
	"context"
	"database/sql"
	"errors"
	"sort"
	"time"

	"github.com/fabianthorsen/openpadel/internal/domain"
	"github.com/fabianthorsen/openpadel/internal/store/db"
)

func (s *Store) SaveRounds(sessionID string, rounds []domain.Round) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	qtx := s.queries.WithTx(tx)
	for _, r := range rounds {
		if err := qtx.InsertRound(context.Background(), db.InsertRoundParams{
			ID:        r.ID,
			SessionID: sessionID,
			Number:    int64(r.Number),
		}); err != nil {
			return err
		}
		for _, pid := range r.Bench {
			if err := qtx.InsertBench(context.Background(), db.InsertBenchParams{
				RoundID:  r.ID,
				PlayerID: pid,
			}); err != nil {
				return err
			}
		}
		for _, m := range r.Matches {
			if err := qtx.InsertMatch(context.Background(), db.InsertMatchParams{
				ID:      m.ID,
				RoundID: r.ID,
				Court:   int64(m.Court),
				P1:      m.TeamA[0],
				P2:      m.TeamA[1],
				P3:      m.TeamB[0],
				P4:      m.TeamB[1],
			}); err != nil {
				return err
			}
		}
	}
	return tx.Commit()
}

func (s *Store) GetRounds(sessionID string) ([]domain.Round, error) {
	rows, err := s.queries.GetRoundsBySessionID(context.Background(), sessionID)
	if err != nil {
		return nil, err
	}

	var rounds []domain.Round
	for _, row := range rows {
		r := domain.Round{
			ID:        row.ID,
			SessionID: sessionID,
			Number:    int(row.Number),
		}
		rounds = append(rounds, r)
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
	row, err := s.queries.GetMatchByID(context.Background(), matchID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return rowToMatch(row), nil
}

func (s *Store) UpdateScore(matchID string, scoreA, scoreB int) (*domain.Match, error) {
	err := s.queries.UpdateMatchScore(context.Background(), db.UpdateMatchScoreParams{
		ScoreA: sql.NullInt64{Int64: int64(scoreA), Valid: true},
		ScoreB: sql.NullInt64{Int64: int64(scoreB), Valid: true},
		ID:     matchID,
	})
	if err != nil {
		return nil, err
	}
	return s.GetMatch(matchID)
}

func (s *Store) UpdateLiveScore(matchID, server string, a, b int) error {
	return s.queries.UpdateMatchLiveScore(context.Background(), db.UpdateMatchLiveScoreParams{
		LiveA:  sql.NullInt64{Int64: int64(a), Valid: true},
		LiveB:  sql.NullInt64{Int64: int64(b), Valid: true},
		Server: sql.NullString{String: server, Valid: true},
		ID:     matchID,
	})
}

func (s *Store) GetLeaderboard(sessionID string) ([]domain.Standing, error) {
	rows, err := s.queries.GetLeaderboard(context.Background(), sessionID)
	if err != nil {
		return nil, err
	}

	var standings []domain.Standing
	for _, row := range rows {
		st := domain.Standing{
			PlayerID:       row.ID,
			Name:           row.Name,
			Points:         int(row.Points),
			PointsConceded: int(row.PointsConceded),
			GamesPlayed:    int(row.GamesPlayed),
			Wins:           int(row.Wins),
			Draws:          int(row.Draws),
			AvatarIcon:     row.AvatarIcon,
			AvatarColor:    row.AvatarColor,
		}
		if row.UserID.Valid {
			st.UserID = &row.UserID.String
		}
		standings = append(standings, st)
	}

	// Build head-to-head lookup: h2h[playerA][playerB] = points playerA scored against playerB's team.
	h2h, err := s.getH2H(sessionID)
	if err != nil {
		return nil, err
	}

	// Sort with tiebreaker chain:
	// 1. Total points (desc)
	// 2. Points per game (desc) — handles bench inequality
	// 3. Head-to-head points (desc)
	// 4. Point spread / net points (desc)
	sort.SliceStable(standings, func(i, j int) bool {
		a, b := standings[i], standings[j]
		if a.Points != b.Points {
			return a.Points > b.Points
		}
		// Points per game
		avgA := perGame(a.Points, a.GamesPlayed)
		avgB := perGame(b.Points, b.GamesPlayed)
		if avgA != avgB {
			return avgA > avgB
		}
		// Head-to-head
		h2hA := h2h[a.PlayerID][b.PlayerID]
		h2hB := h2h[b.PlayerID][a.PlayerID]
		if h2hA != h2hB {
			return h2hA > h2hB
		}
		// Point spread
		spreadA := a.Points - a.PointsConceded
		spreadB := b.Points - b.PointsConceded
		return spreadA > spreadB
	})

	for i := range standings {
		standings[i].Rank = i + 1
	}
	if standings == nil {
		standings = []domain.Standing{}
	}
	return standings, nil
}

func perGame(points, games int) float64 {
	if games == 0 {
		return 0
	}
	return float64(points) / float64(games)
}

// getH2H returns a map[playerID][opponentID] = points scored by playerID against opponentID's team.
func (s *Store) getH2H(sessionID string) (map[string]map[string]int, error) {
	rows, err := s.db.Query(`
		SELECT m.p1, m.p2, m.p3, m.p4, m.score_a, m.score_b
		FROM matches m
		JOIN rounds r ON r.id = m.round_id
		WHERE r.session_id = ? AND m.score_a IS NOT NULL`,
		sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	h2h := map[string]map[string]int{}
	ensure := func(id string) {
		if h2h[id] == nil {
			h2h[id] = map[string]int{}
		}
	}

	for rows.Next() {
		var p1, p2, p3, p4 string
		var scoreA, scoreB int
		if err := rows.Scan(&p1, &p2, &p3, &p4, &scoreA, &scoreB); err != nil {
			return nil, err
		}
		// Team A (p1,p2) scored scoreA against team B (p3,p4)
		for _, a := range []string{p1, p2} {
			ensure(a)
			for _, opp := range []string{p3, p4} {
				h2h[a][opp] += scoreA
			}
		}
		// Team B (p3,p4) scored scoreB against team A (p1,p2)
		for _, b := range []string{p3, p4} {
			ensure(b)
			for _, opp := range []string{p1, p2} {
				h2h[b][opp] += scoreB
			}
		}
	}
	return h2h, rows.Err()
}


// AdvanceMexicanoRound saves a newly generated round and updates current_round atomically.
func (s *Store) AdvanceMexicanoRound(sessionID string, round domain.Round) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	qtx := s.queries.WithTx(tx)
	if err := qtx.InsertRound(context.Background(), db.InsertRoundParams{
		ID:        round.ID,
		SessionID: sessionID,
		Number:    int64(round.Number),
	}); err != nil {
		return err
	}
	for _, pid := range round.Bench {
		if err := qtx.InsertBench(context.Background(), db.InsertBenchParams{
			RoundID:  round.ID,
			PlayerID: pid,
		}); err != nil {
			return err
		}
	}
	for _, m := range round.Matches {
		if err := qtx.InsertMatch(context.Background(), db.InsertMatchParams{
			ID:      m.ID,
			RoundID: round.ID,
			Court:   int64(m.Court),
			P1:      m.TeamA[0],
			P2:      m.TeamA[1],
			P3:      m.TeamB[0],
			P4:      m.TeamB[1],
		}); err != nil {
			return err
		}
	}
	now := time.Now().UTC().Format(time.RFC3339)
	if err := qtx.UpdateSessionCurrentRound(context.Background(), db.UpdateSessionCurrentRoundParams{
		CurrentRound: int64(round.Number),
		UpdatedAt:    now,
		ID:           sessionID,
	}); err != nil {
		return err
	}
	return tx.Commit()
}

// AllRoundsComplete returns true when every match in the session has a score.
func (s *Store) AllRoundsComplete(sessionID string) (bool, error) {
	count, err := s.queries.AllRoundsComplete(context.Background(), sessionID)
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

func rowToMatch(row db.Match) *domain.Match {
	m := &domain.Match{
		ID:      row.ID,
		RoundID: row.RoundID,
		Court:   int(row.Court),
		TeamA:   [2]string{row.P1, row.P2},
		TeamB:   [2]string{row.P3, row.P4},
	}
	if row.ScoreA.Valid {
		m.Score = &domain.Score{A: int(row.ScoreA.Int64), B: int(row.ScoreB.Int64)}
	}
	if row.LiveA.Valid {
		m.Live = &domain.LiveScore{A: int(row.LiveA.Int64), B: int(row.LiveB.Int64), Server: row.Server.String}
	}
	return m
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
