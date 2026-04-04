package store

import (
	"database/sql"
	"errors"
	"time"

	"github.com/fabianthorsen/nottennis/internal/domain"
)

var ErrNotFound = errors.New("not found")

func (s *Store) CreateSession(courts, points int) (*domain.Session, error) {
	now := time.Now().UTC()
	sess := &domain.Session{
		ID:         newID(),
		AdminToken: newAdminToken(),
		Status:     domain.StatusLobby,
		Courts:     courts,
		Points:     points,
		Players:    []domain.Player{},
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	_, err := s.db.Exec(
		`INSERT INTO sessions (id, admin_token, status, courts, points, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		sess.ID, sess.AdminToken, sess.Status, sess.Courts, sess.Points,
		sess.CreatedAt.Format(time.RFC3339), sess.UpdatedAt.Format(time.RFC3339),
	)
	return sess, err
}

func (s *Store) GetSession(id string) (*domain.Session, error) {
	row := s.db.QueryRow(
		`SELECT id, admin_token, status, courts, points, rounds_total, creator_player_id, current_round, created_at, updated_at
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

func (s *Store) CompleteSession(id string) error {
	_, err := s.db.Exec(
		`UPDATE sessions SET status = ?, updated_at = ? WHERE id = ?`,
		domain.StatusComplete, time.Now().UTC().Format(time.RFC3339), id,
	)
	return err
}

func scanSession(row *sql.Row) (*domain.Session, error) {
	var sess domain.Session
	var roundsTotal sql.NullInt64
	var creatorPlayerID sql.NullString
	var currentRound sql.NullInt64
	var createdAt, updatedAt string
	err := row.Scan(
		&sess.ID, &sess.AdminToken, &sess.Status,
		&sess.Courts, &sess.Points, &roundsTotal,
		&creatorPlayerID, &currentRound, &createdAt, &updatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
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
	sess.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	sess.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	return &sess, nil
}

func (s *Store) DeleteSession(id string) error {
	// Delete in dependency order due to foreign keys.
	s.db.Exec(`DELETE FROM bench WHERE round_id IN (SELECT id FROM rounds WHERE session_id = ?)`, id)
	s.db.Exec(`DELETE FROM matches WHERE round_id IN (SELECT id FROM rounds WHERE session_id = ?)`, id)
	s.db.Exec(`DELETE FROM rounds WHERE session_id = ?`, id)
	s.db.Exec(`DELETE FROM players WHERE session_id = ?`, id)
	_, err := s.db.Exec(`DELETE FROM sessions WHERE id = ?`, id)
	return err
}
