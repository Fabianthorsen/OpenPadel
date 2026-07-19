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

// modeStatsFor returns the ModeStats for the given mode from a GetModeStats
// result. GetModeStats always returns both modes, so this never fails to find.
func modeStatsFor(t *testing.T, all []domain.ModeStats, mode domain.GameMode) domain.ModeStats {
	t.Helper()
	for _, ms := range all {
		if ms.Mode == mode {
			return ms
		}
	}
	t.Fatalf("mode %s not present in ModeStats result %+v", mode, all)
	return domain.ModeStats{}
}

// Americano and Mexicano are aggregated into separate sections; a match in one
// mode never bleeds into the other's numbers, and point-share is a per-mode mean.
func TestGetModeStats_SeparatesModes(t *testing.T) {
	s := newTestStore(t)
	alice := createUser(t, s, "alice@example.com", "Alice")

	// Americano: two matches — a win (16–8) and a loss (10–14).
	seedFinishedMatch(t, s, "a1", domain.ModeAmericano, 24, alice, 16, 8, true, true)
	seedFinishedMatch(t, s, "a2", domain.ModeAmericano, 24, alice, 10, 14, true, true)
	// Mexicano: one win (20–16).
	seedFinishedMatch(t, s, "m1", domain.ModeMexicano, 36, alice, 20, 16, true, true)

	all, err := s.GetModeStats(alice)
	if err != nil {
		t.Fatalf("GetModeStats: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("expected both modes, got %d rows", len(all))
	}
	// Canonical order: Americano first, then Mexicano.
	if all[0].Mode != domain.ModeAmericano || all[1].Mode != domain.ModeMexicano {
		t.Errorf("modes out of canonical order: %s, %s", all[0].Mode, all[1].Mode)
	}

	am := modeStatsFor(t, all, domain.ModeAmericano)
	if am.Games != 2 || am.Wins != 1 || am.Losses != 1 || am.Draws != 0 {
		t.Errorf("americano record: games=%d wins=%d losses=%d draws=%d, want 2/1/1/0", am.Games, am.Wins, am.Losses, am.Draws)
	}
	if am.TotalPoints != 26 || am.PointsConceded != 22 || am.NetPoints != 4 {
		t.Errorf("americano points: total=%d conceded=%d net=%d, want 26/22/4", am.TotalPoints, am.PointsConceded, am.NetPoints)
	}
	if am.Tournaments != 2 {
		t.Errorf("americano tournaments = %d, want 2", am.Tournaments)
	}
	// Mean of shares (16/24, 10/24) = (0.6667 + 0.4167)/2 = 0.5417 -> 54.17%.
	if !approx(am.PointWinPct, 54.17) {
		t.Errorf("americano point_win_pct = %.2f, want ~54.17", am.PointWinPct)
	}

	mx := modeStatsFor(t, all, domain.ModeMexicano)
	if mx.Games != 1 || mx.Wins != 1 || mx.NetPoints != 4 {
		t.Errorf("mexicano: games=%d wins=%d net=%d, want 1/1/4", mx.Games, mx.Wins, mx.NetPoints)
	}
	if mx.TotalPoints != 20 || mx.PointsConceded != 16 {
		t.Errorf("mexicano points: total=%d conceded=%d, want 20/16", mx.TotalPoints, mx.PointsConceded)
	}
}

// Net point differential is scored − conceded, and is negative for a losing net.
func TestGetModeStats_NetDifferentialNegative(t *testing.T) {
	s := newTestStore(t)
	alice := createUser(t, s, "alice@example.com", "Alice")

	seedFinishedMatch(t, s, "loss", domain.ModeAmericano, 24, alice, 8, 16, true, true)

	am := modeStatsFor(t, mustModeStats(t, s, alice), domain.ModeAmericano)
	if am.TotalPoints != 8 || am.PointsConceded != 16 {
		t.Errorf("points: total=%d conceded=%d, want 8/16", am.TotalPoints, am.PointsConceded)
	}
	if am.NetPoints != -8 {
		t.Errorf("net = %d, want -8", am.NetPoints)
	}
}

// A mode with no games is still returned, zero-valued, so the UI renders its
// "no games yet" state rather than dropping the section.
func TestGetModeStats_ModeWithNoGames(t *testing.T) {
	s := newTestStore(t)
	alice := createUser(t, s, "alice@example.com", "Alice")

	seedFinishedMatch(t, s, "only", domain.ModeAmericano, 24, alice, 16, 8, true, true)

	all := mustModeStats(t, s, alice)
	mx := modeStatsFor(t, all, domain.ModeMexicano)
	if mx.Games != 0 || mx.Wins != 0 || mx.TotalPoints != 0 || mx.NetPoints != 0 || mx.PointWinPct != 0 || mx.Tournaments != 0 {
		t.Errorf("expected zeroed mexicano section, got %+v", mx)
	}
}

// Only fully-scored matches in done sessions contribute; unscored and still-live
// matches are excluded, exactly as the cross-mode summary excludes them.
func TestGetModeStats_ExcludesUnscoredAndLive(t *testing.T) {
	s := newTestStore(t)
	alice := createUser(t, s, "alice@example.com", "Alice")

	seedFinishedMatch(t, s, "counted", domain.ModeAmericano, 24, alice, 16, 8, true, true)  // counts
	seedFinishedMatch(t, s, "unscored", domain.ModeAmericano, 24, alice, 0, 0, false, true) // done but unscored
	seedFinishedMatch(t, s, "live", domain.ModeAmericano, 24, alice, 12, 12, true, false)   // scored but not done

	am := modeStatsFor(t, mustModeStats(t, s, alice), domain.ModeAmericano)
	if am.Games != 1 {
		t.Errorf("games = %d, want 1 (only fully-scored, done match)", am.Games)
	}
	if am.TotalPoints != 16 || am.PointsConceded != 8 {
		t.Errorf("points: total=%d conceded=%d, want 16/8", am.TotalPoints, am.PointsConceded)
	}
}

// Guest teammates/opponents (no user_id) still count toward the mode's games.
func TestGetModeStats_GuestsCounted(t *testing.T) {
	s := newTestStore(t)
	alice := createUser(t, s, "alice@example.com", "Alice")

	seedFinishedMatch(t, s, "guests", domain.ModeAmericano, 24, alice, 10, 14, true, true)

	am := modeStatsFor(t, mustModeStats(t, s, alice), domain.ModeAmericano)
	if am.Games != 1 {
		t.Errorf("games = %d, want 1 (guest-inclusive game counted)", am.Games)
	}
}

// A brand-new user with no sessions gets both modes back, fully zeroed.
func TestGetModeStats_ZeroGames(t *testing.T) {
	s := newTestStore(t)
	alice := createUser(t, s, "alice@example.com", "Alice")

	all := mustModeStats(t, s, alice)
	if len(all) != 2 {
		t.Fatalf("expected both modes, got %d", len(all))
	}
	for _, ms := range all {
		if ms.Games != 0 || ms.PointWinPct != 0 || ms.Tournaments != 0 {
			t.Errorf("expected zeroed %s section, got %+v", ms.Mode, ms)
		}
	}
}

// Placement stats come from the player's finishing rank in each done Session
// (ADR 0007), blended across both Game Modes since a rank compares like-for-like
// where point-share does not. In a single-match session the winning team (alice)
// finishes rank 1, the losing team rank 3, so win/loss controls her placement.
// Titles = rank-1 finishes, podiums = rank ≤ 3, best finish = lowest rank,
// average finish = mean rank.
func TestGetCareerSummary_Placement(t *testing.T) {
	s := newTestStore(t)
	alice := createUser(t, s, "alice@example.com", "Alice")

	// Americano: two firsts (wins → rank 1) and one third (loss → rank 3).
	seedFinishedMatch(t, s, "p1", domain.ModeAmericano, 24, alice, 16, 8, true, true)  // rank 1
	seedFinishedMatch(t, s, "p2", domain.ModeAmericano, 24, alice, 21, 3, true, true)  // rank 1
	seedFinishedMatch(t, s, "p3", domain.ModeAmericano, 24, alice, 10, 14, true, true) // rank 3
	// Mexicano: a single third-place finish (loss → rank 3), folded into the same aggregate.
	seedFinishedMatch(t, s, "p4", domain.ModeMexicano, 36, alice, 15, 21, true, true) // rank 3

	sum, err := s.GetCareerSummary(alice)
	if err != nil {
		t.Fatalf("GetCareerSummary: %v", err)
	}
	if sum.Titles != 2 {
		t.Errorf("titles = %d, want 2 (two rank-1 finishes)", sum.Titles)
	}
	if sum.Podiums != 4 {
		t.Errorf("podiums = %d, want 4 (all four finishes are rank ≤ 3)", sum.Podiums)
	}
	if sum.BestFinish != 1 {
		t.Errorf("best finish = %d, want 1", sum.BestFinish)
	}
	// Mean of ranks (1, 1, 3, 3) = 8/4 = 2.
	if !approx(sum.AverageFinish, 2) {
		t.Errorf("average finish = %.3f, want ~2", sum.AverageFinish)
	}
}

// A user with no games has no placements: every placement field is zero (and the
// UI hides the tiles off Games == 0, same as the rest of the headline).
func TestGetCareerSummary_PlacementNoGames(t *testing.T) {
	s := newTestStore(t)
	alice := createUser(t, s, "alice@example.com", "Alice")

	sum, err := s.GetCareerSummary(alice)
	if err != nil {
		t.Fatalf("GetCareerSummary: %v", err)
	}
	if sum.Titles != 0 || sum.Podiums != 0 || sum.BestFinish != 0 || sum.AverageFinish != 0 {
		t.Errorf("expected no placements, got titles=%d podiums=%d best=%d avg=%.2f",
			sum.Titles, sum.Podiums, sum.BestFinish, sum.AverageFinish)
	}
}

// Only fully-scored, done sessions yield a finish: an unscored (ended-early) or
// still-live session gives the player no ranked standing to count.
func TestGetCareerSummary_PlacementExcludesUnscoredAndLive(t *testing.T) {
	s := newTestStore(t)
	alice := createUser(t, s, "alice@example.com", "Alice")

	seedFinishedMatch(t, s, "counted", domain.ModeAmericano, 24, alice, 16, 8, true, true)  // rank 1, counts
	seedFinishedMatch(t, s, "unscored", domain.ModeAmericano, 24, alice, 0, 0, false, true) // done but unscored
	seedFinishedMatch(t, s, "live", domain.ModeAmericano, 24, alice, 12, 12, true, false)   // scored but not done

	sum, err := s.GetCareerSummary(alice)
	if err != nil {
		t.Fatalf("GetCareerSummary: %v", err)
	}
	if sum.Titles != 1 || sum.Podiums != 1 || sum.BestFinish != 1 || !approx(sum.AverageFinish, 1) {
		t.Errorf("expected exactly one counted first place, got titles=%d podiums=%d best=%d avg=%.2f",
			sum.Titles, sum.Podiums, sum.BestFinish, sum.AverageFinish)
	}
}

func mustModeStats(t *testing.T, s *store.Store, userID string) []domain.ModeStats {
	t.Helper()
	all, err := s.GetModeStats(userID)
	if err != nil {
		t.Fatalf("GetModeStats: %v", err)
	}
	return all
}
