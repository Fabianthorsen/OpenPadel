package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/fabianthorsen/nottennis/internal/domain"
)

// SaveTennisTeams replaces all team assignments for a session.
func (s *Store) SaveTennisTeams(sessionID string, teams []domain.TennisTeam) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	if _, err := tx.Exec(`DELETE FROM tennis_teams WHERE session_id = ?`, sessionID); err != nil {
		return err
	}
	for _, t := range teams {
		if _, err := tx.Exec(
			`INSERT INTO tennis_teams (session_id, player_id, team) VALUES (?, ?, ?)`,
			sessionID, t.PlayerID, t.Team,
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// GetTennisTeams returns all team assignments for a session, with player names joined.
func (s *Store) GetTennisTeams(sessionID string) ([]domain.TennisTeam, error) {
	rows, err := s.db.Query(`
		SELECT tt.player_id, p.name, tt.team
		FROM tennis_teams tt
		JOIN players p ON p.id = tt.player_id
		WHERE tt.session_id = ?`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []domain.TennisTeam
	for rows.Next() {
		var t domain.TennisTeam
		if err := rows.Scan(&t.PlayerID, &t.Name, &t.Team); err != nil {
			return nil, err
		}
		teams = append(teams, t)
	}
	return teams, rows.Err()
}

// CreateTennisMatch creates the initial match record for a tennis session.
func (s *Store) CreateTennisMatch(sessionID string) (*domain.TennisMatch, error) {
	now := time.Now().UTC()
	state := domain.TennisState{Sets: [][2]int{}}
	stateJSON, _ := json.Marshal(state)
	id := newID()
	_, err := s.db.Exec(
		`INSERT INTO tennis_matches (id, session_id, state, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
		id, sessionID, string(stateJSON),
		now.Format(time.RFC3339), now.Format(time.RFC3339),
	)
	if err != nil {
		return nil, err
	}
	return &domain.TennisMatch{
		ID:        id,
		SessionID: sessionID,
		State:     state,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// GetTennisMatch retrieves the tennis match state and teams for a session.
func (s *Store) GetTennisMatch(sessionID string) (*domain.TennisMatch, error) {
	var m domain.TennisMatch
	var stateJSON string
	var createdAt, updatedAt string

	err := s.db.QueryRow(
		`SELECT id, session_id, state, created_at, updated_at FROM tennis_matches WHERE session_id = ?`,
		sessionID,
	).Scan(&m.ID, &m.SessionID, &stateJSON, &createdAt, &updatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	m.State.Sets = [][2]int{} // ensure non-null JSON array
	if err := json.Unmarshal([]byte(stateJSON), &m.State); err != nil {
		return nil, err
	}
	if m.State.Sets == nil {
		m.State.Sets = [][2]int{}
	}
	m.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	m.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

	teams, err := s.GetTennisTeams(sessionID)
	if err != nil {
		return nil, err
	}
	m.Teams = domain.TennisTeams{A: []domain.TennisTeam{}, B: []domain.TennisTeam{}}
	for _, t := range teams {
		if t.Team == "a" {
			m.Teams.A = append(m.Teams.A, t)
		} else {
			m.Teams.B = append(m.Teams.B, t)
		}
	}

	return &m, nil
}

// SaveTennisState persists updated match state.
func (s *Store) SaveTennisState(matchID string, state domain.TennisState) error {
	if state.Sets == nil {
		state.Sets = [][2]int{}
	}
	stateJSON, err := json.Marshal(state)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(
		`UPDATE tennis_matches SET state = ?, updated_at = ? WHERE id = ?`,
		string(stateJSON), time.Now().UTC().Format(time.RFC3339), matchID,
	)
	return err
}
