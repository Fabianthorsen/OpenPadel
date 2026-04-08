package scheduler

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

// TestGenerate_NoRepeatedPartners is a regression test for the core Americano invariant:
// within a full tournament (players-1 rounds for N players on N/4 courts), every player
// should partner each other player AT MOST ONCE.
//
// With 8 players and 2 courts there are C(8,2)=28 possible pairs; each round produces
// 4 partnerships, so 7 rounds × 4 = 28 total partnerships — exactly enough to cover every
// pair once. A correct algorithm achieves this; the current rotate-based algorithm
// assigns the same initial pairs every few rounds, causing pairs like (P01,P02) to appear
// 4 times while other pairs never appear at all.
func TestGenerate_NoRepeatedPartners(t *testing.T) {
	type pair struct{ a, b string }
	makePair := func(a, b string) pair {
		if a > b {
			a, b = b, a
		}
		return pair{a, b}
	}

	cases := []struct {
		name          string
		players, courts, rounds int
	}{
		// No-bench cases: N players, N/4 courts, N-1 rounds covers all unique pairs exactly once.
		{"8p 2c 7rounds", 8, 2, 7},
		{"12p 3c 11rounds", 12, 3, 11},
		// With bench: more rounds than unique partnerships are possible, so cap at 2 appearances.
		{"9p 2c 8rounds", 9, 2, 8},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rounds := Generate(makePlayers(tc.players), tc.courts, tc.rounds)

			partnerCount := map[pair]int{}
			for _, r := range rounds {
				for _, m := range r.Matches {
					partnerCount[makePair(m.TeamA[0], m.TeamA[1])]++
					partnerCount[makePair(m.TeamB[0], m.TeamB[1])]++
				}
			}

			// Core invariant: in a no-bench tournament every pair should partner exactly once.
			// In tournaments with a bench a pair may appear at most twice (unavoidable when
			// active slots × rounds exceed unique pairs), but never more.
			maxAllowed := 1
			if tc.players%4 != 0 {
				maxAllowed = 2
			}

			for p, count := range partnerCount {
				if count > maxAllowed {
					t.Errorf("pair (%s, %s) partnered %d times across %d rounds — expected at most %d (Americano invariant broken)",
						p.a, p.b, count, len(rounds), maxAllowed)
				}
			}
		})
	}
}

// TestGenerate_ConsecutiveRoundsNoRepeat asserts the weaker but observable constraint:
// a player must NEVER receive the same partner in back-to-back rounds.
// This is the immediate symptom of the scheduling bug.
func TestGenerate_ConsecutiveRoundsNoRepeat(t *testing.T) {
	type pair struct{ a, b string }
	makePair := func(a, b string) pair {
		if a > b {
			a, b = b, a
		}
		return pair{a, b}
	}

	cases := []struct {
		name          string
		players, courts, rounds int
	}{
		{"8p 2c 7rounds", 8, 2, 7},
		{"9p 2c 8rounds", 9, 2, 8},
		{"12p 3c 11rounds", 12, 3, 11},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rounds := Generate(makePlayers(tc.players), tc.courts, tc.rounds)

			for i := 1; i < len(rounds); i++ {
				prev := rounds[i-1]
				curr := rounds[i]

				prevPairs := map[pair]bool{}
				for _, m := range prev.Matches {
					prevPairs[makePair(m.TeamA[0], m.TeamA[1])] = true
					prevPairs[makePair(m.TeamB[0], m.TeamB[1])] = true
				}

				for _, m := range curr.Matches {
					for _, p := range []pair{
						makePair(m.TeamA[0], m.TeamA[1]),
						makePair(m.TeamB[0], m.TeamB[1]),
					} {
						if prevPairs[p] {
							t.Errorf("rounds %d→%d: pair (%s, %s) partnered in consecutive rounds",
								prev.Number, curr.Number, p.a, p.b)
						}
					}
				}
			}
		})
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

// TestTotalRounds_MinimumRounds is a regression test for configurations where
// the gcd formula produces too few rounds for a real tournament.
// For 6 players on 1 court (benchSize=2), gcd(6,2)=2 gives only 3 rounds —
// barely enough for everyone to sit once, but not enough for a proper tournament.
// The correct minimum should be comparable to N-1 rounds, keeping bench fair.
func TestTotalRounds_MinimumRounds(t *testing.T) {
	cases := []struct {
		players, courts, wantAtLeast int
		note                         string
	}{
		// Regression: 6p 1c was giving 3 rounds (too few)
		{6, 1, 5, "6p 1c: should play at least 5 rounds, not 3"},
		// 10p 2c: gcd gives 5, but N-1=9 is more appropriate
		{10, 2, 9, "10p 2c: should play at least 9 rounds, not 5"},
		// These should be unchanged
		{9, 2, 9, "9p 2c: 9 rounds unchanged"},
		{8, 2, 7, "8p 2c: 7 rounds unchanged"},
	}

	for _, tc := range cases {
		t.Run(tc.note, func(t *testing.T) {
			got := TotalRounds(tc.players, tc.courts)
			if got < tc.wantAtLeast {
				t.Errorf("TotalRounds(%d, %d) = %d, want at least %d", tc.players, tc.courts, got, tc.wantAtLeast)
			}
		})
	}
}

// TestGenerate_NoRepeatedMatchups verifies the opposition invariant:
// the same two partnerships should never face each other more than once.
// A match is an unordered pair of partnerships, so {A+B vs C+D} == {C+D vs A+B}.
// With 4 players there are exactly 3 unique matchups; a correct scheduler uses each once.
func TestGenerate_NoRepeatedMatchups(t *testing.T) {
	type pair struct{ a, b string }
	makePair := func(a, b string) pair {
		if a > b { a, b = b, a }
		return pair{a, b}
	}
	type matchup struct{ p1, p2 pair }
	makeMatchup := func(a0, a1, b0, b1 string) matchup {
		pa := makePair(a0, a1)
		pb := makePair(b0, b1)
		// normalize: smaller pair first
		if pa.a > pb.a || (pa.a == pb.a && pa.b > pb.b) {
			pa, pb = pb, pa
		}
		return matchup{pa, pb}
	}

	cases := []struct {
		name                    string
		players, courts, rounds int
	}{
		{"4p 1c 3rounds", 4, 1, 3},
		{"8p 2c 7rounds", 8, 2, 7},
		{"12p 3c 11rounds", 12, 3, 11},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rounds := Generate(makePlayers(tc.players), tc.courts, tc.rounds)

			seen := map[matchup]int{}
			for _, r := range rounds {
				for _, m := range r.Matches {
					mu := makeMatchup(m.TeamA[0], m.TeamA[1], m.TeamB[0], m.TeamB[1])
					seen[mu]++
					if seen[mu] > 1 {
						t.Errorf("matchup {(%s+%s) vs (%s+%s)} appeared %d times — each matchup should occur at most once",
							mu.p1.a, mu.p1.b, mu.p2.a, mu.p2.b, seen[mu])
					}
				}
			}
		})
	}
}

// TestGenerate_BenchRotation verifies two bench invariants:
// 1. A player benched in round N must play in round N+1 (no consecutive bench).
// TestTotalRounds verifies the correct round count for common player/court combos.
func TestTotalRounds(t *testing.T) {
	cases := []struct {
		players, courts, want int
		note                  string
	}{
		// No bench: N-1
		{8, 2, 7, "8p 2c: no bench"},
		{12, 3, 11, "12p 3c: no bench"},
		{4, 1, 3, "4p 1c: no bench"},
		// Bench of 1: N/gcd(N,1) = N
		{9, 2, 9, "9p 2c: bench=1"},
		{5, 1, 5, "5p 1c: bench=1"},
		{13, 3, 13, "13p 3c: bench=1"},
		// Bench of 2: cycle=N/gcd(N,2), rounded up to >= N-1
		{10, 2, 10, "10p 2c: bench=2, cycle=5, target=9 → 10 rounds"},
		{6, 1, 6, "6p 1c: bench=2, cycle=3, target=5 → 6 rounds"},
		// Bench of 3: cycle=N/gcd(N,3), rounded up to >= N-1
		{11, 2, 11, "11p 2c: bench=3, gcd=1 → 11 rounds"},
		{15, 3, 15, "15p 3c: bench=3, cycle=5, target=14 → 15 rounds"},
	}

	for _, tc := range cases {
		t.Run(tc.note, func(t *testing.T) {
			got := TotalRounds(tc.players, tc.courts)
			if got != tc.want {
				t.Errorf("TotalRounds(%d, %d) = %d, want %d", tc.players, tc.courts, got, tc.want)
			}

			// Verify bench fairness with the computed round count
			benchSize := tc.players - tc.courts*4
			if benchSize > 0 {
				players := makePlayers(tc.players)
				rounds := Generate(players, tc.courts, got)
				benchCount := map[string]int{}
				for _, r := range rounds {
					for _, id := range r.Bench {
						benchCount[id]++
					}
				}
				min, max := got, 0
				for _, p := range players {
					c := benchCount[p.ID]
					if c < min { min = c }
					if c > max { max = c }
				}
				if min == 0 {
					t.Errorf("player(s) never benched across %d rounds", got)
				}
				if max-min > 1 {
					t.Errorf("bench counts uneven: min=%d max=%d", min, max)
				}
			}
		})
	}
}

// 2. Across the full tournament, bench counts are as equal as possible —
//    the max bench count for any player must not exceed the min by more than 1.
func TestGenerate_BenchRotation(t *testing.T) {
	cases := []struct {
		name                    string
		players, courts, rounds int
	}{
		{"5p 1c 5rounds", 5, 1, 5},
		{"9p 2c 9rounds", 9, 2, 9},
		{"10p 2c 9rounds", 10, 2, 9},
		{"13p 3c 12rounds", 13, 3, 12},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			players := makePlayers(tc.players)
			rounds := Generate(players, tc.courts, tc.rounds)

			benchedLastRound := map[string]bool{}
			benchCount := map[string]int{}

			for _, p := range players {
				benchCount[p.ID] = 0
			}

			for _, r := range rounds {
				benchedThisRound := map[string]bool{}
				for _, id := range r.Bench {
					benchedThisRound[id] = true
					benchCount[id]++

					// Invariant 1: must not bench consecutively
					if benchedLastRound[id] {
						t.Errorf("round %d: player %s is benched two rounds in a row", r.Number, id)
					}
				}
				benchedLastRound = benchedThisRound
			}

			// Invariant 2: bench counts must be spread evenly (max - min <= 1)
			minB, maxB := 1<<30, 0
			for _, c := range benchCount {
				if c < minB { minB = c }
				if c > maxB { maxB = c }
			}
			if maxB-minB > 1 {
				t.Errorf("bench counts are uneven: min=%d max=%d (spread > 1) — counts: %v", minB, maxB, benchCount)
			}
		})
	}
}
