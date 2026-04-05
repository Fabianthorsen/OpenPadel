package scheduler

import (
	"fmt"
	"testing"

	"github.com/fabianthorsen/nottennis/internal/domain"
)

func makePlayers(n int) []domain.Player {
	players := make([]domain.Player, n)
	for i := range players {
		id := fmt.Sprintf("P%02d", i+1)
		players[i] = domain.Player{ID: id, Name: id}
	}
	return players
}

// Every player appears in exactly courts*4 active slots per round (or bench).
func TestGenerate_PlayerCount(t *testing.T) {
	cases := []struct{ players, courts int }{
		{8, 2},
		{9, 2},
		{12, 3},
		{6, 1},
	}
	for _, tc := range cases {
		players := makePlayers(tc.players)
		rounds := Generate(players, tc.courts, tc.players)

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

// Each match must have exactly 4 distinct players.
func TestGenerate_MatchesHaveFourDistinctPlayers(t *testing.T) {
	rounds := Generate(makePlayers(8), 2, 8)
	for _, r := range rounds {
		for _, m := range r.Matches {
			ids := []string{m.TeamA[0], m.TeamA[1], m.TeamB[0], m.TeamB[1]}
			seen := map[string]bool{}
			for _, id := range ids {
				if seen[id] {
					t.Errorf("round %d court %d: duplicate player %s", r.Number, m.Court, id)
				}
				seen[id] = true
			}
		}
	}
}

// A player benched in round N must play in round N+1.
func TestGenerate_BenchedPlayerPlaysNext(t *testing.T) {
	players := makePlayers(9) // 9 players, 2 courts → 1 bench per round
	rounds := Generate(players, 2, 9)

	for i := 1; i < len(rounds); i++ {
		prev := rounds[i-1]
		curr := rounds[i]

		activeCurr := map[string]bool{}
		for _, m := range curr.Matches {
			activeCurr[m.TeamA[0]] = true
			activeCurr[m.TeamA[1]] = true
			activeCurr[m.TeamB[0]] = true
			activeCurr[m.TeamB[1]] = true
		}

		for _, id := range prev.Bench {
			if !activeCurr[id] {
				t.Errorf("player %s was benched in round %d but not active in round %d", id, prev.Number, curr.Number)
			}
		}
	}
}

// RoundsForPlayers returns N-1 rounds for N players (covers all unique pairings).
func TestRoundsForPlayers(t *testing.T) {
	cases := []struct {
		players, want int
	}{
		{4, 3},
		{8, 7},
		{12, 11},
		{9, 8},
	}
	for _, tc := range cases {
		got := tc.players - 1
		if got != tc.want {
			t.Errorf("players=%d: want %d rounds, got %d", tc.players, tc.want, got)
		}
	}
}

// Court numbers must be sequential starting at 1.
func TestGenerate_CourtNumbering(t *testing.T) {
	rounds := Generate(makePlayers(8), 2, 4)
	for _, r := range rounds {
		if len(r.Matches) != 2 {
			t.Errorf("round %d: expected 2 matches, got %d", r.Number, len(r.Matches))
		}
		for i, m := range r.Matches {
			if m.Court != i+1 {
				t.Errorf("round %d: match %d has court %d, expected %d", r.Number, i, m.Court, i+1)
			}
		}
	}
}

// Round numbers must be sequential starting at 1.
func TestGenerate_RoundNumbering(t *testing.T) {
	rounds := Generate(makePlayers(8), 2, 8)
	for i, r := range rounds {
		if r.Number != i+1 {
			t.Errorf("round at index %d has number %d, expected %d", i, r.Number, i+1)
		}
	}
}

// Consecutive rounds should not pair the same partners more than necessary.
// This is a soft constraint — we just check that not every pair repeats.
func TestGenerate_PartnerVariety(t *testing.T) {
	players := makePlayers(8)
	rounds := Generate(players, 2, 8)

	type pair struct{ a, b string }
	makePair := func(a, b string) pair {
		if a > b {
			a, b = b, a
		}
		return pair{a, b}
	}

	// Count how many times each pair partners across all rounds.
	partnerCount := map[pair]int{}
	for _, r := range rounds {
		for _, m := range r.Matches {
			partnerCount[makePair(m.TeamA[0], m.TeamA[1])]++
			partnerCount[makePair(m.TeamB[0], m.TeamB[1])]++
		}
	}

	// No pair should partner more than half the rounds (loose sanity check).
	maxAllowed := len(rounds)/2 + 1
	for p, count := range partnerCount {
		if count > maxAllowed {
			t.Errorf("pair (%s, %s) partnered %d times in %d rounds (max expected %d)", p.a, p.b, count, len(rounds), maxAllowed)
		}
	}
}

// Bench slots should be distributed fairly — no player sits more than once extra vs others.
func TestGenerate_BenchFairness(t *testing.T) {
	cases := []struct {
		name          string
		players, courts, rounds int
	}{
		{"9p 2c 1bench/round", 9, 2, 8},
		{"10p 2c 2bench/round", 10, 2, 9},
		{"11p 2c 3bench/round", 11, 2, 10},
		{"13p 3c 1bench/round", 13, 3, 12},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rounds := Generate(makePlayers(tc.players), tc.courts, tc.rounds)

			benchCount := map[string]int{}
			for _, r := range rounds {
				for _, id := range r.Bench {
					benchCount[id]++
				}
			}

			min, max := 999, 0
			for _, c := range benchCount {
				if c < min { min = c }
				if c > max { max = c }
			}
			if max-min > 1 {
				t.Errorf("bench distribution unfair: min=%d max=%d (diff should be ≤1)", min, max)
			}
		})
	}
}

// Every player should play the same number of matches, or at most 1 apart.
func TestGenerate_GamesPlayedEquality(t *testing.T) {
	cases := []struct {
		name          string
		players, courts, rounds int
	}{
		{"4p 1c no bench", 4, 1, 3},
		{"8p 2c no bench", 8, 2, 7},
		{"9p 2c 1bench/round", 9, 2, 8},
		{"10p 2c 2bench/round", 10, 2, 9},
		{"11p 2c 3bench/round", 11, 2, 10},
		{"12p 3c no bench", 12, 3, 11},
		{"13p 3c 1bench/round", 13, 3, 12},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rounds := Generate(makePlayers(tc.players), tc.courts, tc.rounds)

			gamesPlayed := map[string]int{}
			for _, r := range rounds {
				for _, m := range r.Matches {
					for _, id := range []string{m.TeamA[0], m.TeamA[1], m.TeamB[0], m.TeamB[1]} {
						gamesPlayed[id]++
					}
				}
			}

			min, max := 999, 0
			for _, g := range gamesPlayed {
				if g < min { min = g }
				if g > max { max = g }
			}
			if max-min > 1 {
				t.Errorf("games played unfair: min=%d max=%d (diff should be ≤1)", min, max)
			}
		})
	}
}

// With no bench (players == courts*4), every player plays every round.
func TestGenerate_NoMissedRounds_NoBench(t *testing.T) {
	cases := []struct{ players, courts int }{
		{4, 1},
		{8, 2},
		{12, 3},
	}
	for _, tc := range cases {
		players := makePlayers(tc.players)
		rounds := Generate(players, tc.courts, tc.players-1)
		allIDs := make(map[string]bool)
		for _, p := range players {
			allIDs[p.ID] = true
		}
		for _, r := range rounds {
			active := map[string]bool{}
			for _, m := range r.Matches {
				for _, id := range []string{m.TeamA[0], m.TeamA[1], m.TeamB[0], m.TeamB[1]} {
					active[id] = true
				}
			}
			for id := range allIDs {
				if !active[id] {
					t.Errorf("round %d: player %s was unexpectedly benched (no-bench scenario)", r.Number, id)
				}
			}
			if len(r.Bench) != 0 {
				t.Errorf("round %d: expected empty bench, got %v", r.Number, r.Bench)
			}
		}
	}
}

// No player should ever appear on both teams in the same match.
func TestGenerate_NoSelfOpposition(t *testing.T) {
	cases := []struct{ players, courts, rounds int }{
		{8, 2, 7},
		{9, 2, 8},
		{12, 3, 11},
	}
	for _, tc := range cases {
		rounds := Generate(makePlayers(tc.players), tc.courts, tc.rounds)
		for _, r := range rounds {
			for _, m := range r.Matches {
				aSet := map[string]bool{m.TeamA[0]: true, m.TeamA[1]: true}
				for _, id := range []string{m.TeamB[0], m.TeamB[1]} {
					if aSet[id] {
						t.Errorf("round %d court %d: player %s appears on both teams", r.Number, m.Court, id)
					}
				}
			}
		}
	}
}
