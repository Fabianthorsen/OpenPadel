package store

import (
	"database/sql"
	"time"

	"github.com/fabianthorsen/openpadel/internal/domain"
)

func (s *Store) CreatePlayer(sessionID, name, userID string) (*domain.Player, error) {
	now := time.Now().UTC()
	p := &domain.Player{
		ID:        newID(),
		SessionID: sessionID,
		UserID:    userID,
		Name:      name,
		Active:    true,
		JoinedAt:  now,
	}
	var uid *string
	if userID != "" {
		uid = &userID
	}
	_, err := s.db.Exec(
		`INSERT INTO players (id, session_id, user_id, name, active, joined_at) VALUES (?, ?, ?, ?, 1, ?)`,
		p.ID, p.SessionID, uid, p.Name, p.JoinedAt.Format(time.RFC3339),
	)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Store) GetPlayers(sessionID string) ([]domain.Player, error) {
	rows, err := s.db.Query(
		`SELECT id, session_id, COALESCE(user_id, ''), name, active, joined_at FROM players WHERE session_id = ? ORDER BY joined_at`,
		sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPlayers(rows)
}

// GetCreatorName returns the name of the creator player for the given session,
// or an empty string if the creator hasn't joined yet.
func (s *Store) GetCreatorName(sessionID string) string {
	var name string
	s.db.QueryRow(`
		SELECT p.name FROM players p
		JOIN sessions s ON s.creator_player_id = p.id
		WHERE s.id = ?`, sessionID).Scan(&name)
	return name
}

func (s *Store) DeactivatePlayer(playerID string) error {
	res, err := s.db.Exec(`UPDATE players SET active = 0 WHERE id = ?`, playerID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func scanPlayers(rows *sql.Rows) ([]domain.Player, error) {
	var players []domain.Player
	for rows.Next() {
		var p domain.Player
		var active int
		var joinedAt string
		if err := rows.Scan(&p.ID, &p.SessionID, &p.UserID, &p.Name, &active, &joinedAt); err != nil {
			return nil, err
		}
		p.Active = active == 1
		p.JoinedAt, _ = time.Parse(time.RFC3339, joinedAt)
		players = append(players, p)
	}
	if players == nil {
		players = []domain.Player{}
	}
	return players, rows.Err()
}

