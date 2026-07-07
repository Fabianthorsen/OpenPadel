package americano

import (
	"fmt"
	"testing"
	"time"

	"github.com/fabianthorsen/openpadel/internal/domain"
)

type mockStore struct {
	savedRounds    []domain.Round
	sessionStarted bool
	saveRoundsErr  error
	startSessionErr error
	allRoundsCompleteResult bool
	advanceRoundErr error
}

func (m *mockStore) SaveRounds(sessionID string, rounds []domain.Round) error {
	if m.saveRoundsErr != nil {
		return m.saveRoundsErr
	}
	m.savedRounds = append(m.savedRounds, rounds...)
	return nil
}

func (m *mockStore) StartSession(id string, roundsTotal int, endsAt *time.Time) error {
	if m.startSessionErr != nil {
		return m.startSessionErr
	}
	m.sessionStarted = true
	return nil
}

func (m *mockStore) AllRoundsComplete(sessionID string) (bool, error) {
	return m.allRoundsCompleteResult, nil
}

func (m *mockStore) AdvanceAmericanoRound(sessionID string, round domain.Round) error {
	if m.advanceRoundErr != nil {
		return m.advanceRoundErr
	}
	m.savedRounds = append(m.savedRounds, round)
	return nil
}

func (m *mockStore) GetRounds(sessionID string) ([]domain.Round, error) {
	return m.savedRounds, nil
}

// TestAdvanceAmericanoRound_Unlimited verifies that AdvanceAmericanoRound generates
// a new round given previously-played rounds, respecting all constraints.
func TestAdvanceAmericanoRound_Unlimited(t *testing.T) {
	store := &mockStore{}
	svc := New(store)

	// Simulate: 9 players, 2 courts, unlimited (no fixed total)
	players := make([]domain.Player, 9)
	for i := 0; i < 9; i++ {
		players[i] = domain.Player{
			ID:   fmt.Sprintf("P%02d", i+1),
			Name: fmt.Sprintf("Player %d", i+1),
		}
	}

	// AdvanceAmericanoRound should exist and take previousRounds, players, courts
	// This is a placeholder test that will fail until AdvanceAmericanoRound is implemented
	err := svc.AdvanceAmericanoRound("sess-1", []domain.Round{}, players, 2)
	if err != nil {
		t.Fatalf("AdvanceAmericanoRound failed: %v", err)
	}
}
