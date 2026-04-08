package store

import (
	"database/sql"
	"errors"
	"time"

	"github.com/fabianthorsen/openpadel/internal/domain"
)

var ErrNotFound = errors.New("not found")

func (s *Store) CreateSession(courts, points int, name, gameMode string, setsToWin, gamesPerSet int, roundsTotal *int, scheduledAt *time.Time) (*domain.Session, error) {
	now := time.Now().UTC()
	if gameMode == "" {
		gameMode = "americano"
	}
	if setsToWin == 0 {
		setsToWin = 2
	}
	if gamesPerSet != 4 && gamesPerSet != 6 {
		gamesPerSet = 6
	}
	sess := &domain.Session{
		ID:          newID(),
		AdminToken:  newAdminToken(),
		Status:      domain.StatusLobby,
		Name:        name,
		GameMode:    gameMode,
		SetsToWin:   setsToWin,
		GamesPerSet: gamesPerSet,
		Courts:      courts,
		Points:      points,
		RoundsTotal: roundsTotal,
		ScheduledAt: scheduledAt,
		Players:     []domain.Player{},
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	var scheduledAtStr *string
	if scheduledAt != nil {
		s := scheduledAt.UTC().Format(time.RFC3339)
		scheduledAtStr = &s
	}
	_, err := s.db.Exec(
		`INSERT INTO sessions (id, admin_token, status, name, game_mode, sets_to_win, games_per_set, courts, points, rounds_total, scheduled_at, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		sess.ID, sess.AdminToken, sess.Status, sess.Name, sess.GameMode, sess.SetsToWin, sess.GamesPerSet,
		sess.Courts, sess.Points, roundsTotal, scheduledAtStr,
		sess.CreatedAt.Format(time.RFC3339), sess.UpdatedAt.Format(time.RFC3339),
	)
	return sess, err
}

func (s *Store) GetSession(id string) (*domain.Session, error) {
	row := s.db.QueryRow(
		`SELECT id, admin_token, status, name, game_mode, sets_to_win, games_per_set, courts, points, rounds_total, creator_player_id, current_round, scheduled_at, created_at, updated_at
		 FROM sessions WHERE id = ?`, id,
	)
	sess, err := scanSession(row)
	if err != nil {
		return nil, err
	}
	players, err := s.GetPlayers(id)
	if err != nil {
		return nil, err
	}
	sess.Players = players
	return sess, nil
}

func (s *Store) SetCreatorPlayer(sessionID, playerID string) error {
	_, err := s.db.Exec(
		`UPDATE sessions SET creator_player_id = ?, updated_at = ? WHERE id = ?`,
		playerID, time.Now().UTC().Format(time.RFC3339), sessionID,
	)
	return err
}

func (s *Store) StartSession(id string, roundsTotal int) error {
	_, err := s.db.Exec(
		`UPDATE sessions SET status = ?, rounds_total = ?, current_round = 1, updated_at = ? WHERE id = ?`,
		domain.StatusActive, roundsTotal, time.Now().UTC().Format(time.RFC3339), id,
	)
	return err
}

// StartMexicanoSession activates the session with current_round=1.
// rounds_total is preserved from creation time (null = open-ended, N = preset).
func (s *Store) StartMexicanoSession(id string) error {
	_, err := s.db.Exec(
		`UPDATE sessions SET status = ?, current_round = 1, updated_at = ? WHERE id = ?`,
		domain.StatusActive, time.Now().UTC().Format(time.RFC3339), id,
	)
	return err
}

func (s *Store) AdvanceRound(id string) error {
	_, err := s.db.Exec(
		`UPDATE sessions SET current_round = current_round + 1, updated_at = ? WHERE id = ?`,
		time.Now().UTC().Format(time.RFC3339), id,
	)
	return err
}

func (s *Store) CurrentRoundAllScored(sessionID string) (bool, error) {
	var unscored int
	err := s.db.QueryRow(`
		SELECT COUNT(*) FROM matches m
		JOIN rounds r ON r.id = m.round_id
		JOIN sessions s ON s.id = r.session_id
		WHERE s.id = ? AND r.number = s.current_round AND m.score_a IS NULL`,
		sessionID,
	).Scan(&unscored)
	return unscored == 0, err
}

func (s *Store) CompleteSession(id string, endedEarly bool) error {
	endedEarlyInt := 0
	if endedEarly {
		endedEarlyInt = 1
	}
	_, err := s.db.Exec(
		`UPDATE sessions SET status = ?, ended_early = ?, updated_at = ? WHERE id = ?`,
		domain.StatusComplete, endedEarlyInt, time.Now().UTC().Format(time.RFC3339), id,
	)
	return err
}

func scanSession(row *sql.Row) (*domain.Session, error) {
	var sess domain.Session
	var roundsTotal sql.NullInt64
	var creatorPlayerID sql.NullString
	var currentRound sql.NullInt64
	var name, gameMode sql.NullString
	var setsToWin, gamesPerSet sql.NullInt64
	var scheduledAt sql.NullString
	var createdAt, updatedAt string
	err := row.Scan(
		&sess.ID, &sess.AdminToken, &sess.Status, &name,
		&gameMode, &setsToWin, &gamesPerSet,
		&sess.Courts, &sess.Points, &roundsTotal,
		&creatorPlayerID, &currentRound, &scheduledAt, &createdAt, &updatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if name.Valid {
		sess.Name = name.String
	}
	sess.GameMode = "americano"
	if gameMode.Valid && gameMode.String != "" {
		sess.GameMode = gameMode.String
	}
	sess.SetsToWin = 2
	if setsToWin.Valid && setsToWin.Int64 > 0 {
		sess.SetsToWin = int(setsToWin.Int64)
	}
	sess.GamesPerSet = 6
	if gamesPerSet.Valid && gamesPerSet.Int64 > 0 {
		sess.GamesPerSet = int(gamesPerSet.Int64)
	}
	if roundsTotal.Valid {
		v := int(roundsTotal.Int64)
		sess.RoundsTotal = &v
	}
	if creatorPlayerID.Valid {
		sess.CreatorPlayerID = creatorPlayerID.String
	}
	if currentRound.Valid {
		v := int(currentRound.Int64)
		sess.CurrentRound = &v
	}
	if scheduledAt.Valid {
		t, _ := time.Parse(time.RFC3339, scheduledAt.String)
		sess.ScheduledAt = &t
	}
	sess.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	sess.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	return &sess, nil
}

func (s *Store) DeleteSession(id string) error {
	// Delete in dependency order due to foreign keys.
	s.db.Exec(`DELETE FROM tennis_matches WHERE session_id = ?`, id)
	s.db.Exec(`DELETE FROM tennis_teams WHERE session_id = ?`, id)
	s.db.Exec(`DELETE FROM bench WHERE round_id IN (SELECT id FROM rounds WHERE session_id = ?)`, id)
	s.db.Exec(`DELETE FROM matches WHERE round_id IN (SELECT id FROM rounds WHERE session_id = ?)`, id)
	s.db.Exec(`DELETE FROM rounds WHERE session_id = ?`, id)
	s.db.Exec(`DELETE FROM players WHERE session_id = ?`, id)
	_, err := s.db.Exec(`DELETE FROM sessions WHERE id = ?`, id)
	return err
}
