package americano

import (
	"fmt"
	"testing"

	"github.com/fabianthorsen/openpadel/internal/domain"
)

func makePlayers(n int) []domain.Player {
	players := make([]domain.Player, n)
	for i := range players {
		id := fmt.Sprintf("P%02d", i+1)
		players[i] = domain.Player{ID: id, Name: id}
	}
	return players
}

// TestGenerateRounds_PlayerCoverage verifies that every player appears
// in exactly one slot per round (either a match or bench).
func TestGenerateRounds_PlayerCoverage(t *testing.T) {
	cases := []struct {
		players, courts, totalRounds int
	}{
		{8, 2, 7},
		{9, 2, 9},
		{12, 3, 11},
		{6, 1, 5},
	}
	for _, tc := range cases {
		players := makePlayers(tc.players)
		rounds := GenerateRounds(players, tc.courts, tc.totalRounds)

		for _, r := range rounds {
			active := map[string]bool{}
			for _, m := range r.Matches {
				for _, id := range []string{m.TeamA[0], m.TeamA[1], m.TeamB[0], m.TeamB[1]} {
					if active[id] {
						t.Errorf("round %d: player %s appears more than once", r.Number, id)
					}
					active[id] = true
				}
			}
			for _, id := range r.Bench {
				if active[id] {
					t.Errorf("round %d: bench player %s also in a match", r.Number, id)
				}
				active[id] = true
			}
			if len(active) != tc.players {
				t.Errorf("round %d: expected %d players accounted for, got %d", r.Number, tc.players, len(active))
			}
		}
	}
}

// TestGenerateRounds_NoBench_NoConsecutiveBench verifies that when there is no
// bench (courts*4 == players), every round includes all players.
func TestGenerateRounds_NoBench(t *testing.T) {
	players := makePlayers(8)
	rounds := GenerateRounds(players, 2, 7)

	for _, r := range rounds {
		if len(r.Bench) != 0 {
			t.Errorf("round %d: expected no bench (8 players, 2 courts), got %v", r.Number, r.Bench)
		}
	}
}

// TestGenerateRounds_WithBench_NoBenchInConsecutiveRounds verifies that
// a benched player in round N must play in round N+1.
func TestGenerateRounds_WithBench_NoConsecutiveBench(t *testing.T) {
	players := makePlayers(9) // 1 bench per round (9 - 2*4)
	rounds := GenerateRounds(players, 2, 9)

	benchedLast := make(map[string]int) // playerID -> round number they were last benched

	for _, r := range rounds {
		// Check that anyone benched in the previous round is not benched now.
		for _, benchedID := range r.Bench {
			if lastBench, ok := benchedLast[benchedID]; ok && lastBench == r.Number-1 {
				t.Errorf("round %d: player %s benched in both round %d and %d", r.Number, benchedID, lastBench, r.Number)
			}
		}

		// Update the bench log.
		for _, benchedID := range r.Bench {
			benchedLast[benchedID] = r.Number
		}
	}
}

// TestGenerateRounds_RoundCount verifies the scheduler produces the requested number.
func TestGenerateRounds_RoundCount(t *testing.T) {
	players := makePlayers(8)
	totalRounds := 7
	rounds := GenerateRounds(players, 2, totalRounds)

	if len(rounds) != totalRounds {
		t.Errorf("expected %d rounds, got %d", totalRounds, len(rounds))
	}

	for i, r := range rounds {
		expected := i + 1
		if r.Number != expected {
			t.Errorf("round %d: expected number %d, got %d", i, expected, r.Number)
		}
	}
}

// TestGenerateRounds_CourtAssignment verifies that each match is assigned
// to a court in the valid range [1, courts].
func TestGenerateRounds_CourtAssignment(t *testing.T) {
	players := makePlayers(12)
	courts := 3
	rounds := GenerateRounds(players, courts, 11)

	for _, r := range rounds {
		for _, m := range r.Matches {
			if m.Court < 1 || m.Court > courts {
				t.Errorf("round %d: match %s on invalid court %d (expected 1-%d)", r.Number, m.ID, m.Court, courts)
			}
		}
	}
}

// TestGenerateRounds_MatchHasFourDistinctPlayers verifies that each match
// has exactly 4 distinct players.
func TestGenerateRounds_MatchHasFourDistinctPlayers(t *testing.T) {
	players := makePlayers(8)
	rounds := GenerateRounds(players, 2, 7)

	for _, r := range rounds {
		for _, m := range r.Matches {
			ids := map[string]bool{
				m.TeamA[0]: true,
				m.TeamA[1]: true,
				m.TeamB[0]: true,
				m.TeamB[1]: true,
			}
			if len(ids) != 4 {
				t.Errorf("round %d match %s: expected 4 distinct players, got %d", r.Number, m.ID, len(ids))
			}
		}
	}
}
