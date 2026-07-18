package mexicano

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fabianthorsen/openpadel/internal/domain"
)

type mockStore struct {
	savedRounds    []domain.Round
	sessionStarted bool
	leaderboard    []domain.Standing
	session        *domain.Session
}

func (m *mockStore) GetLeaderboard(sessionID string) ([]domain.Standing, error) {
	return m.leaderboard, nil
}

func (m *mockStore) GetSession(id string) (*domain.Session, error) {
	return m.session, nil
}

func (m *mockStore) SaveRounds(sessionID string, rounds []domain.Round) error {
	m.savedRounds = append(m.savedRounds, rounds...)
	return nil
}

func (m *mockStore) StartMexicanoSession(id string, endsAt *time.Time) error {
	m.sessionStarted = true
	return nil
}

func (m *mockStore) AdvanceMexicanoRound(sessionID string, round domain.Round) error {
	m.savedRounds = append(m.savedRounds, round)
	return nil
}

// courtPlayerSet returns the set of player IDs on a given court of a round.
func courtPlayerSet(r domain.Round, courtIdx int) map[string]bool {
	m := r.Matches[courtIdx]
	return map[string]bool{
		m.TeamA[0]: true, m.TeamA[1]: true,
		m.TeamB[0]: true, m.TeamB[1]: true,
	}
}

// TestStart_SeedsRound1ByRatingDescending verifies that Start orders the initial
// standings by rating (descending) before generating round 1, so the four
// strongest players share the top court and the four weakest share the other —
// replacing the old random shuffle. Per ADR 0006 / issue #208.
func TestStart_SeedsRound1ByRatingDescending(t *testing.T) {
	store := &mockStore{}
	svc := New(store)

	// Deliberately interleave strong and weak in join order so a rating-blind
	// seed would NOT strength-group them. Ratings: strong=5, weak=1.
	active := []domain.Player{
		{ID: "s1", Rating: 5},
		{ID: "w1", Rating: 1},
		{ID: "s2", Rating: 5},
		{ID: "w2", Rating: 1},
		{ID: "s3", Rating: 5},
		{ID: "w3", Rating: 1},
		{ID: "s4", Rating: 5},
		{ID: "w4", Rating: 1},
	}
	sess := &domain.Session{ID: "sess-1", GameMode: domain.ModeMexicano, Courts: 2}

	rec := httptest.NewRecorder()
	if err := svc.Start(rec, "sess-1", sess, active, nil); err != nil {
		t.Fatalf("Start failed: %v (body %q)", err, rec.Body.String())
	}
	if !store.sessionStarted {
		t.Fatal("session was not started")
	}
	if len(store.savedRounds) != 1 {
		t.Fatalf("expected 1 saved round, got %d", len(store.savedRounds))
	}

	r := store.savedRounds[0]
	if len(r.Matches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(r.Matches))
	}

	strong := map[string]bool{"s1": true, "s2": true, "s3": true, "s4": true}
	weak := map[string]bool{"w1": true, "w2": true, "w3": true, "w4": true}

	// Exactly one court is all-strong and the other all-weak.
	c0 := courtPlayerSet(r, 0)
	c1 := courtPlayerSet(r, 1)
	allStrong := func(s map[string]bool) bool {
		for id := range s {
			if !strong[id] {
				return false
			}
		}
		return true
	}
	allWeak := func(s map[string]bool) bool {
		for id := range s {
			if !weak[id] {
				return false
			}
		}
		return true
	}
	strengthSeeded := (allStrong(c0) && allWeak(c1)) || (allWeak(c0) && allStrong(c1))
	if !strengthSeeded {
		t.Errorf("round 1 not strength-seeded: court0=%v court1=%v", c0, c1)
	}
}

// TestAdvanceRound_UsesLiveStandingsUnchanged verifies that round 2+ is driven by
// the live leaderboard (not ratings) exactly as before.
func TestAdvanceRound_UsesLiveStandingsUnchanged(t *testing.T) {
	store := &mockStore{
		session: &domain.Session{ID: "sess-1", GameMode: domain.ModeMexicano, Courts: 2},
		leaderboard: []domain.Standing{
			{Rank: 1, PlayerID: "p1", Points: 24},
			{Rank: 2, PlayerID: "p2", Points: 24},
			{Rank: 3, PlayerID: "p3", Points: 24},
			{Rank: 4, PlayerID: "p4", Points: 24},
			{Rank: 5, PlayerID: "p5", Points: 0},
			{Rank: 6, PlayerID: "p6", Points: 0},
			{Rank: 7, PlayerID: "p7", Points: 0},
			{Rank: 8, PlayerID: "p8", Points: 0},
		},
	}
	svc := New(store)

	rec := httptest.NewRecorder()
	if err := svc.AdvanceRound(rec, "sess-1", 2); err != nil {
		t.Fatalf("AdvanceRound failed: %v", err)
	}
	if len(store.savedRounds) != 1 {
		t.Fatalf("expected 1 saved round, got %d", len(store.savedRounds))
	}
	// Court 1 = top-4 (winners), court 2 = bottom-4 (losers), from live standings.
	top := map[string]bool{"p1": true, "p2": true, "p3": true, "p4": true}
	for id := range courtPlayerSet(store.savedRounds[0], 0) {
		if !top[id] {
			t.Errorf("court 1 should hold live top-4, found %s", id)
		}
	}
}
