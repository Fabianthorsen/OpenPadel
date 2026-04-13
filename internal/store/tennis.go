package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/fabianthorsen/openpadel/internal/domain"
	"github.com/fabianthorsen/openpadel/internal/store/db"
)

// SaveTennisTeams replaces all team assignments for a session.
func (s *Store) SaveTennisTeams(sessionID string, teams []domain.TennisTeam) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	qtx := s.queries.WithTx(tx)
	if err := qtx.DeleteTennisTeams(context.Background(), sessionID); err != nil {
		return err
	}
	for _, t := range teams {
		if err := qtx.InsertTennisTeam(context.Background(), db.InsertTennisTeamParams{
			SessionID: sessionID,
			PlayerID:  t.PlayerID,
			Team:      t.Team,
		}); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// GetTennisTeams returns all team assignments for a session, with player names joined.
func (s *Store) GetTennisTeams(sessionID string) ([]domain.TennisTeam, error) {
	rows, err := s.queries.GetTennisTeamsBySessionID(context.Background(), sessionID)
	if err != nil {
		return nil, err
	}

	var teams []domain.TennisTeam
	for _, row := range rows {
		t := domain.TennisTeam{
			PlayerID: row.PlayerID,
			Name:     row.Name,
			Team:     row.Team,
		}
		teams = append(teams, t)
	}
	return teams, nil
}

// CreateTennisMatch creates the initial match record for a tennis session.
func (s *Store) CreateTennisMatch(sessionID string) (*domain.TennisMatch, error) {
	now := time.Now().UTC()
	state := domain.TennisState{Sets: [][2]int{}}
	stateJSON, _ := json.Marshal(state)
	id := newID()
	err := s.queries.CreateTennisMatch(context.Background(), db.CreateTennisMatchParams{
		ID:        id,
		SessionID: sessionID,
		State:     string(stateJSON),
		CreatedAt: now.Format(time.RFC3339),
		UpdatedAt: now.Format(time.RFC3339),
	})
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
	row, err := s.queries.GetTennisMatch(context.Background(), sessionID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	m := &domain.TennisMatch{
		ID:        row.ID,
		SessionID: row.SessionID,
		CreatedAt: parseTime(row.CreatedAt),
		UpdatedAt: parseTime(row.UpdatedAt),
	}

	m.State.Sets = [][2]int{} // ensure non-null JSON array
	if err := json.Unmarshal([]byte(row.State), &m.State); err != nil {
		return nil, err
	}
	if m.State.Sets == nil {
		m.State.Sets = [][2]int{}
	}

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

	return m, nil
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
	return s.queries.UpdateTennisState(context.Background(), db.UpdateTennisStateParams{
		State:     string(stateJSON),
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
		ID:        matchID,
	})
}
