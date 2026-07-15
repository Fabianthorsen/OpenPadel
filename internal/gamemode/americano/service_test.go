package americano

import (
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fabianthorsen/openpadel/internal/domain"
)

type mockStore struct {
	savedRounds             []domain.Round
	sessionStarted          bool
	startedRoundsTotal      int
	saveRoundsErr           error
	startSessionErr         error
	allRoundsCompleteResult bool
	advanceRoundErr         error
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
	m.startedRoundsTotal = roundsTotal
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

func (m *mockStore) StartAmericanoSession(id string, endsAt *time.Time) error {
	if m.startSessionErr != nil {
		return m.startSessionErr
	}
	m.sessionStarted = true
	return nil
}

// TestStart_FixedAmericano_10p2c is a regression test for the round count of a
// fair fixed ("limited") Americano tournament.
//
// For 10 players on 2 courts there are 8 playing slots, so 2 players bench each
// round. The fair tournament length is TotalRounds(10,2) = 10 rounds: over 10
// rounds with 2 benched each, all 10 players sit out exactly twice. (The UI's
// fallback default of 7 is NOT the fair count — this pins the correct value.)
//
// It drives the real Service.Start → GenerateRounds pipeline and asserts the
// session starts with exactly 10 rounds, each with 2 courts and a fair bench.
func TestStart_FixedAmericano_10p2c(t *testing.T) {
	const players, courts = 10, 2

	// The expected value: 10. Documented here so a formula change trips this test.
	wantRounds := TotalRounds(players, courts)
	if wantRounds != 10 {
		t.Fatalf("precondition: TotalRounds(%d, %d) = %d, want 10", players, courts, wantRounds)
	}

	store := &mockStore{}
	svc := New(store)

	active := make([]domain.Player, players)
	for i := range active {
		active[i] = domain.Player{ID: fmt.Sprintf("P%02d", i+1), Name: fmt.Sprintf("Player %d", i+1)}
	}

	// Deliberately configure the WRONG count the buggy client used to send (7).
	// A limited Americano only signals "limited" via a non-nil rounds_total; the
	// backend must recompute the fair count (10) from the roster and ignore this 7.
	staleClientValue := 7
	sess := &domain.Session{
		ID:          "sess-1",
		GameMode:    domain.ModeAmericano,
		Courts:      courts,
		RoundsTotal: &staleClientValue,
	}

	rec := httptest.NewRecorder()
	if err := svc.Start(rec, "sess-1", sess, active, nil); err != nil {
		t.Fatalf("Start failed: %v (status %d, body %q)", err, rec.Code, rec.Body.String())
	}

	if !store.sessionStarted {
		t.Fatal("session was not started")
	}
	if got := len(store.savedRounds); got != 10 {
		t.Fatalf("started with %d rounds, want 10 (backend must recompute, not use the client's 7)", got)
	}
	if store.startedRoundsTotal != 10 {
		t.Errorf("persisted rounds_total = %d, want 10 (fair count, not the client's %d)",
			store.startedRoundsTotal, staleClientValue)
	}

	// Every round fills both courts and benches exactly (players - courts*4) = 2.
	const wantBench = players - courts*4
	benchCounts := map[string]int{}
	prevBench := map[string]bool{}
	for i, r := range store.savedRounds {
		if len(r.Matches) != courts {
			t.Errorf("round %d: %d matches, want %d", i+1, len(r.Matches), courts)
		}
		if len(r.Bench) != wantBench {
			t.Errorf("round %d: benched %d, want %d", i+1, len(r.Bench), wantBench)
		}
		curBench := map[string]bool{}
		for _, id := range r.Bench {
			benchCounts[id]++
			curBench[id] = true
			if prevBench[id] {
				t.Errorf("round %d: player %s benched two rounds in a row", i+1, id)
			}
		}
		prevBench = curBench
	}

	// Fair rotation: over 10 rounds with 2 benched each, all 10 sit out twice.
	for _, p := range active {
		if benchCounts[p.ID] != 2 {
			t.Errorf("player %s benched %d times, want 2 (fair rotation)", p.ID, benchCounts[p.ID])
		}
	}
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
