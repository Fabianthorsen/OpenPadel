package store_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/fabianthorsen/openpadel/internal/domain"
)

// TestGetClubLeaderboard_RowSelection exercises the store's qualifying-row
// selection and windowing: a Session dated outside the window is excluded, a
// non-club Session never accrues, and Guests contribute team points but produce
// no leaderboard row. Ranking itself is covered by the pure domain tests.
func TestGetClubLeaderboard_RowSelection(t *testing.T) {
	s := newTestStore(t)
	alice := createUser(t, s, "alice@example.com", "Alice")
	bob := createUser(t, s, "bob@example.com", "Bob")
	carol := createUser(t, s, "carol@example.com", "Carol")
	dave := createUser(t, s, "dave@example.com", "Dave")
	clubID := createClub(t, s, alice)

	var matchSeq int
	// mkPlayer adds a player to a session; userID "" makes a Guest.
	mkPlayer := func(sessionID, name, userID string) string {
		p, err := s.CreatePlayer(sessionID, name, userID, false)
		if err != nil {
			t.Fatalf("CreatePlayer: %v", err)
		}
		return p.ID
	}
	// scoreMatch saves a single 2v2 round (teamA = p1,p2) and records its score.
	scoreMatch := func(sessionID, p1, p2, p3, p4 string, a, b int) {
		matchSeq++
		round := domain.Round{
			ID:     fmt.Sprintf("r-%d", matchSeq),
			Number: 1,
			Bench:  []string{},
			Matches: []domain.Match{
				{ID: fmt.Sprintf("m-%d", matchSeq), Court: 1, TeamA: [2]string{p1, p2}, TeamB: [2]string{p3, p4}},
			},
		}
		if err := s.SaveRounds(sessionID, []domain.Round{round}); err != nil {
			t.Fatalf("SaveRounds: %v", err)
		}
		if _, err := s.UpdateScore(fmt.Sprintf("m-%d", matchSeq), a, b); err != nil {
			t.Fatalf("UpdateScore: %v", err)
		}
	}
	linkClub := func(sessionID string) {
		if _, err := s.DB().Exec("UPDATE sessions SET club_id = ? WHERE id = ?", clubID, sessionID); err != nil {
			t.Fatalf("link session to club: %v", err)
		}
	}
	markDone := func(sessionID string) {
		if _, err := s.DB().Exec("UPDATE sessions SET status = 'done' WHERE id = ?", sessionID); err != nil {
			t.Fatalf("mark session done: %v", err)
		}
	}

	// S1: completed in-window club Session — alice&bob (16) beat carol&dave (8).
	s1 := createSession(t, s)
	linkClub(s1)
	scoreMatch(s1, mkPlayer(s1, "Alice", alice), mkPlayer(s1, "Bob", bob), mkPlayer(s1, "Carol", carol), mkPlayer(s1, "Dave", dave), 16, 8)
	markDone(s1)

	// S2: completed club Session dated 200 days ago — outside the 90-day window, excluded.
	s2 := createSession(t, s)
	linkClub(s2)
	scoreMatch(s2, mkPlayer(s2, "Alice", alice), mkPlayer(s2, "Bob", bob), mkPlayer(s2, "Carol", carol), mkPlayer(s2, "Dave", dave), 16, 8)
	markDone(s2)
	old := time.Now().UTC().AddDate(0, 0, -200).Format(time.RFC3339)
	if _, err := s.DB().Exec("UPDATE sessions SET created_at = ? WHERE id = ?", old, s2); err != nil {
		t.Fatalf("age session: %v", err)
	}

	// S3: completed in-window Session NOT owned by the Club — never accrues.
	s3 := createSession(t, s)
	scoreMatch(s3, mkPlayer(s3, "Alice", alice), mkPlayer(s3, "Bob", bob), mkPlayer(s3, "Carol", carol), mkPlayer(s3, "Dave", dave), 16, 8)
	markDone(s3)

	// S5: in-progress club Session (still 'playing', not done) with a scored
	// Match — must be excluded until the event completes.
	s5 := createSession(t, s)
	linkClub(s5)
	scoreMatch(s5, mkPlayer(s5, "Alice", alice), mkPlayer(s5, "Bob", bob), mkPlayer(s5, "Carol", carol), mkPlayer(s5, "Dave", dave), 16, 8)
	if _, err := s.DB().Exec("UPDATE sessions SET status = 'playing' WHERE id = ?", s5); err != nil {
		t.Fatalf("mark session playing: %v", err)
	}

	board, err := s.GetClubLeaderboard(clubID)
	if err != nil {
		t.Fatalf("GetClubLeaderboard: %v", err)
	}
	// Only S1 counts: each of the four members has exactly one game.
	if len(board.Ranked) != 0 {
		t.Errorf("expected 0 ranked (all below MinGames), got %d", len(board.Ranked))
	}
	counts := provisionalCounts(board)
	if len(counts) != 4 {
		t.Fatalf("expected 4 provisional members, got %d: %+v", len(counts), counts)
	}
	for _, u := range []string{alice, bob, carol, dave} {
		if counts[u] != 1 {
			t.Errorf("user %s: expected 1 game (S2 out-of-window + S3 non-club + S5 not-done excluded), got %d", u, counts[u])
		}
	}
	if go2 := gamesToGo(board, alice); go2 != domain.ClubLeaderboardMinGames-1 {
		t.Errorf("alice gamesToGo: expected %d, got %d", domain.ClubLeaderboardMinGames-1, go2)
	}

	// S4: completed in-window club Session where a Guest fills a slot. The Guest
	// scores team points but must not appear as a leaderboard row.
	s4 := createSession(t, s)
	linkClub(s4)
	scoreMatch(s4, mkPlayer(s4, "Alice", alice), mkPlayer(s4, "Bob", bob), mkPlayer(s4, "Carol", carol), mkPlayer(s4, "Gus", ""), 16, 8)
	markDone(s4)

	board, err = s.GetClubLeaderboard(clubID)
	if err != nil {
		t.Fatalf("GetClubLeaderboard after guest match: %v", err)
	}
	counts = provisionalCounts(board)
	if len(counts) != 4 {
		t.Fatalf("guest must not create a row; expected 4 members, got %d: %+v", len(counts), counts)
	}
	if counts[alice] != 2 {
		t.Errorf("alice: expected 2 games after S4, got %d", counts[alice])
	}
	for name := range counts {
		if name == "Gus" {
			t.Error("Guest appeared on the leaderboard")
		}
	}
}

func provisionalCounts(b domain.ClubLeaderboard) map[string]int {
	m := map[string]int{}
	for _, p := range b.Provisional {
		m[p.UserID] = p.GamesPlayed
	}
	for _, r := range b.Ranked {
		m[r.UserID] = r.GamesPlayed
	}
	return m
}

func gamesToGo(b domain.ClubLeaderboard, userID string) int {
	for _, p := range b.Provisional {
		if p.UserID == userID {
			return p.GamesToGo
		}
	}
	return -1
}
