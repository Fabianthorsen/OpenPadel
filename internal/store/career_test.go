package store_test

import (
	"math"
	"testing"

	"github.com/fabianthorsen/openpadel/internal/domain"
	"github.com/fabianthorsen/openpadel/internal/store"
)

// seedFinishedMatch creates a single-match session with `userID` on team A and
// three guests filling the other seats, optionally scores it and marks it done.
// id must be unique within a test (used to derive round/match IDs).
func seedFinishedMatch(t *testing.T, s *store.Store, id string, mode domain.GameMode, points int, userID string, scoreFor, scoreAgainst int, scored, done bool) {
	t.Helper()
	sess, err := s.CreateSession(domain.SessionInput{Courts: 2, Points: points, GameMode: mode}, "")
	if err != nil {
		t.Fatalf("CreateSession(%s): %v", id, err)
	}
	alice, _ := s.CreatePlayer(sess.ID, "Alice", userID, false)
	bob, _ := s.CreatePlayer(sess.ID, "Bob", "", false)
	carol, _ := s.CreatePlayer(sess.ID, "Carol", "", false)
	dan, _ := s.CreatePlayer(sess.ID, "Dan", "", false)

	matchID := "m_" + id
	round := domain.Round{
		ID:      "r_" + id,
		Number:  1,
		Bench:   []string{},
		Matches: []domain.Match{{ID: matchID, Court: 1, TeamA: [2]string{alice.ID, bob.ID}, TeamB: [2]string{carol.ID, dan.ID}}},
	}
	if err := s.SaveRounds(sess.ID, []domain.Round{round}); err != nil {
		t.Fatalf("SaveRounds(%s): %v", id, err)
	}
	if err := s.StartSession(sess.ID, 1, nil); err != nil {
		t.Fatalf("StartSession(%s): %v", id, err)
	}
	if scored {
		if _, err := s.UpdateScore(matchID, scoreFor, scoreAgainst); err != nil {
			t.Fatalf("UpdateScore(%s): %v", id, err)
		}
	}
	if done {
		if err := s.CompleteSession(sess.ID, false); err != nil {
			t.Fatalf("CompleteSession(%s): %v", id, err)
		}
	}
}

func approx(a, b float64) bool { return math.Abs(a-b) < 0.05 }

// Point Win % is the mean of per-Match point-share, each match weighted equally
// regardless of its points target — not a points-pooled ratio. A small blowout
// and a large squeaker in different modes must average their shares.
func TestGetCareerSummary_MeanOfShares(t *testing.T) {
	s := newTestStore(t)
	alice := createUser(t, s, "alice@example.com", "Alice")

	// Blowout at a small target: share 6/8 = 0.75.
	seedFinishedMatch(t, s, "blowout", domain.ModeAmericano, 8, alice, 6, 2, true, true)
	// Squeaker at a large target, different mode: share 20/36 = 0.5556.
	seedFinishedMatch(t, s, "squeaker", domain.ModeMexicano, 36, alice, 20, 16, true, true)

	sum, err := s.GetCareerSummary(alice)
	if err != nil {
		t.Fatalf("GetCareerSummary: %v", err)
	}
	if sum.Games != 2 {
		t.Errorf("games = %d, want 2", sum.Games)
	}
	// Mean of shares = (0.75 + 0.5556) / 2 = 0.6528 -> 65.28%.
	// A points-pooled ratio would be 26/44 = 59.1%, which this must NOT be.
	if !approx(sum.PointWinPct, 65.28) {
		t.Errorf("point_win_pct = %.2f, want ~65.28 (mean of shares, not pooled)", sum.PointWinPct)
	}
	// Both matches won.
	if !approx(sum.Winrate, 100) {
		t.Errorf("winrate = %.2f, want 100", sum.Winrate)
	}
}

// Only fully-scored matches in done sessions contribute. Unscored matches (ended
// early) and matches in still-playing sessions are excluded.
func TestGetCareerSummary_ExcludesUnscoredAndLive(t *testing.T) {
	s := newTestStore(t)
	alice := createUser(t, s, "alice@example.com", "Alice")

	seedFinishedMatch(t, s, "counted", domain.ModeAmericano, 24, alice, 16, 8, true, true)  // counts
	seedFinishedMatch(t, s, "unscored", domain.ModeAmericano, 24, alice, 0, 0, false, true) // done but unscored
	seedFinishedMatch(t, s, "live", domain.ModeAmericano, 24, alice, 12, 12, true, false)   // scored but not done

	sum, err := s.GetCareerSummary(alice)
	if err != nil {
		t.Fatalf("GetCareerSummary: %v", err)
	}
	if sum.Games != 1 {
		t.Errorf("games = %d, want 1 (only fully-scored, done match)", sum.Games)
	}
	if !approx(sum.PointWinPct, 66.67) { // 16/24
		t.Errorf("point_win_pct = %.2f, want ~66.67", sum.PointWinPct)
	}
}

// Guest teammates and opponents (no user_id) still count toward the player's games.
func TestGetCareerSummary_GuestsCounted(t *testing.T) {
	s := newTestStore(t)
	alice := createUser(t, s, "alice@example.com", "Alice")

	// All of alice's teammates/opponents are guests (userID "").
	seedFinishedMatch(t, s, "guests", domain.ModeAmericano, 24, alice, 10, 14, true, true)

	sum, err := s.GetCareerSummary(alice)
	if err != nil {
		t.Fatalf("GetCareerSummary: %v", err)
	}
	if sum.Games != 1 {
		t.Errorf("games = %d, want 1 (guest-inclusive game counted)", sum.Games)
	}
	if !approx(sum.Winrate, 0) { // 10 < 14, a loss
		t.Errorf("winrate = %.2f, want 0", sum.Winrate)
	}
}

// A brand-new user with no sessions degrades gracefully: zero games, no error,
// no jarring 0%/100% (caller hides the stat at zero games).
func TestGetCareerSummary_ZeroGames(t *testing.T) {
	s := newTestStore(t)
	alice := createUser(t, s, "alice@example.com", "Alice")

	sum, err := s.GetCareerSummary(alice)
	if err != nil {
		t.Fatalf("GetCareerSummary: %v", err)
	}
	if sum.Games != 0 {
		t.Errorf("games = %d, want 0", sum.Games)
	}
	if sum.PointWinPct != 0 || sum.Winrate != 0 {
		t.Errorf("expected zeroed summary, got point_win_pct=%.2f winrate=%.2f", sum.PointWinPct, sum.Winrate)
	}
}
