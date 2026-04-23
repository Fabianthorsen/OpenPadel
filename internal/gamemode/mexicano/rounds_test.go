package mexicano

import (
	"testing"

	"github.com/fabianthorsen/openpadel/internal/domain"
)

// makeStandings builds a []Standing slice from a list of player IDs, assigned
// ranks in the order given (first ID = rank 1).
func makeStandings(ids ...string) []domain.Standing {
	out := make([]domain.Standing, len(ids))
	for i, id := range ids {
		out[i] = domain.Standing{
			Rank:     i + 1,
			PlayerID: id,
			Points:   (len(ids) - i) * 10, // descending points for clarity
		}
	}
	return out
}

func playerSet(r domain.Round) map[string]bool {
	set := make(map[string]bool)
	for _, m := range r.Matches {
		set[m.TeamA[0]] = true
		set[m.TeamA[1]] = true
		set[m.TeamB[0]] = true
		set[m.TeamB[1]] = true
	}
	return set
}

// TestMexicanoRound1_NoBench verifies round 1 with exactly courts*4 players.
func TestMexicanoRound1_NoBench(t *testing.T) {
	standings := makeStandings("p1", "p2", "p3", "p4", "p5", "p6", "p7", "p8")
	courts := 2

	r := GenerateMexicanoRound(standings, courts, 1)

	if len(r.Matches) != courts {
		t.Fatalf("expected %d matches, got %d", courts, len(r.Matches))
	}
	if len(r.Bench) != 0 {
		t.Fatalf("expected 0 bench, got %v", r.Bench)
	}

	// All 8 players should appear exactly once.
	playing := playerSet(r)
	for _, s := range standings {
		if !playing[s.PlayerID] {
			t.Errorf("player %s not in any match", s.PlayerID)
		}
	}
}

// matchHasTeams returns true if a match contains exactly the two specified player
// pairs (regardless of which is Team A or Team B, since sides are randomised).
func matchHasTeams(m domain.Match, pair1, pair2 [2]string) bool {
	sameTeam := func(got [2]string, want [2]string) bool {
		return (got[0] == want[0] && got[1] == want[1]) ||
			(got[0] == want[1] && got[1] == want[0])
	}
	return (sameTeam(m.TeamA, pair1) && sameTeam(m.TeamB, pair2)) ||
		(sameTeam(m.TeamA, pair2) && sameTeam(m.TeamB, pair1))
}

// TestMexicanoRound1_PairingCorrect verifies the canonical Mexicano pairing rule:
// rank 1+4 vs 2+3 on court 1, rank 5+8 vs 6+7 on court 2, etc.
// This balances each match by ranking (strong player paired with weak within each court).
// Team A/B sides are randomised, so we only check who plays whom.
func TestMexicanoRound1_PairingCorrect(t *testing.T) {
	ids := []string{"p1", "p2", "p3", "p4", "p5", "p6", "p7", "p8"}
	standings := makeStandings(ids...)

	r := GenerateMexicanoRound(standings, 2, 1)

	// Court 1: rank 1+4 vs rank 2+3 → {p1,p4} vs {p2,p3}
	if !matchHasTeams(r.Matches[0], [2]string{"p1", "p4"}, [2]string{"p2", "p3"}) {
		t.Errorf("court 1: want {p1,p4} vs {p2,p3}, got %v vs %v", r.Matches[0].TeamA, r.Matches[0].TeamB)
	}
	// Court 2: rank 5+8 vs rank 6+7 → {p5,p8} vs {p6,p7}
	if !matchHasTeams(r.Matches[1], [2]string{"p5", "p8"}, [2]string{"p6", "p7"}) {
		t.Errorf("court 2: want {p5,p8} vs {p6,p7}, got %v vs %v", r.Matches[1].TeamA, r.Matches[1].TeamB)
	}
}

// TestMexicanoRound_NoBench verifies that Mexicano never produces a bench —
// it requires exactly courts*4 players so everyone plays every round.
func TestMexicanoRound_NoBench(t *testing.T) {
	standings := makeStandings("p1", "p2", "p3", "p4", "p5", "p6", "p7", "p8")
	r := GenerateMexicanoRound(standings, 2, 2)
	if len(r.Bench) != 0 {
		t.Errorf("Mexicano should never have a bench, got %v", r.Bench)
	}
}

// TestMexicanoRound_CourtNumbers verifies courts are numbered 1..N.
func TestMexicanoRound_CourtNumbers(t *testing.T) {
	standings := makeStandings("p1", "p2", "p3", "p4", "p5", "p6", "p7", "p8",
		"p9", "p10", "p11", "p12")
	r := GenerateMexicanoRound(standings, 3, 1)

	for i, m := range r.Matches {
		if m.Court != i+1 {
			t.Errorf("match %d: expected court %d, got %d", i, i+1, m.Court)
		}
	}
}

// TestMexicanoRound_RoundNumberSet checks that the returned round has the correct number.
func TestMexicanoRound_RoundNumberSet(t *testing.T) {
	standings := makeStandings("p1", "p2", "p3", "p4", "p5", "p6", "p7", "p8")
	r := GenerateMexicanoRound(standings, 2, 5)
	if r.Number != 5 {
		t.Errorf("expected round number 5, got %d", r.Number)
	}
}

// ---------------------------------------------------------------------------
// Winners-play-winners / losers-play-losers tests
// ---------------------------------------------------------------------------

// courtPlayers returns the set of all player IDs on a given court (0-indexed).
func courtPlayers(r domain.Round, courtIdx int) map[string]bool {
	m := r.Matches[courtIdx]
	return map[string]bool{
		m.TeamA[0]: true, m.TeamA[1]: true,
		m.TeamB[0]: true, m.TeamB[1]: true,
	}
}

// TestMexicanoWinnersPlayWinners verifies that after round 1 the top-ranked
// players (winners) face each other on court 1, and the bottom-ranked players
// (losers) face each other on court 2.
func TestMexicanoWinnersPlayWinners(t *testing.T) {
	// Simulate round-2 standings: top-4 won round 1 (24 pts each),
	// bottom-4 lost (0 pts each).
	standings := []domain.Standing{
		{Rank: 1, PlayerID: "p1", Points: 24},
		{Rank: 2, PlayerID: "p3", Points: 24},
		{Rank: 3, PlayerID: "p5", Points: 24},
		{Rank: 4, PlayerID: "p7", Points: 24},
		{Rank: 5, PlayerID: "p2", Points: 0},
		{Rank: 6, PlayerID: "p4", Points: 0},
		{Rank: 7, PlayerID: "p6", Points: 0},
		{Rank: 8, PlayerID: "p8", Points: 0},
	}

	r := GenerateMexicanoRound(standings, 2, 2)

	winners := map[string]bool{"p1": true, "p3": true, "p5": true, "p7": true}
	losers := map[string]bool{"p2": true, "p4": true, "p6": true, "p8": true}

	// Court 1 must only contain winners.
	for pid := range courtPlayers(r, 0) {
		if !winners[pid] {
			t.Errorf("court 1 (winners' court) contains non-winner %s", pid)
		}
	}

	// Court 2 must only contain losers.
	for pid := range courtPlayers(r, 1) {
		if !losers[pid] {
			t.Errorf("court 2 (losers' court) contains non-loser %s", pid)
		}
	}
}

// TestMexicanoWinnersPlayWinners_3Courts verifies the three-court case:
// court 1 = ranks 1-4, court 2 = ranks 5-8, court 3 = ranks 9-12.
func TestMexicanoWinnersPlayWinners_3Courts(t *testing.T) {
	standings := []domain.Standing{
		{Rank: 1, PlayerID: "p1", Points: 48},
		{Rank: 2, PlayerID: "p2", Points: 42},
		{Rank: 3, PlayerID: "p3", Points: 36},
		{Rank: 4, PlayerID: "p4", Points: 30},
		{Rank: 5, PlayerID: "p5", Points: 24},
		{Rank: 6, PlayerID: "p6", Points: 20},
		{Rank: 7, PlayerID: "p7", Points: 16},
		{Rank: 8, PlayerID: "p8", Points: 12},
		{Rank: 9, PlayerID: "p9", Points: 8},
		{Rank: 10, PlayerID: "p10", Points: 4},
		{Rank: 11, PlayerID: "p11", Points: 2},
		{Rank: 12, PlayerID: "p12", Points: 0},
	}
	tiers := []map[string]bool{
		{"p1": true, "p2": true, "p3": true, "p4": true},
		{"p5": true, "p6": true, "p7": true, "p8": true},
		{"p9": true, "p10": true, "p11": true, "p12": true},
	}

	r := GenerateMexicanoRound(standings, 3, 2)

	for i, tier := range tiers {
		for pid := range courtPlayers(r, i) {
			if !tier[pid] {
				t.Errorf("court %d contains out-of-tier player %s", i+1, pid)
			}
		}
	}
}

// TestMexicanoWinnersPlayWinners_ExactPairing checks that within a court the
// partner pairing follows the canonical rule: rank N+rank N+3 vs rank N+1+rank N+2.
// Team A/B sides are randomised, so we only check who plays whom.
func TestMexicanoWinnersPlayWinners_ExactPairing(t *testing.T) {
	standings := []domain.Standing{
		{Rank: 1, PlayerID: "a", Points: 32},
		{Rank: 2, PlayerID: "b", Points: 24},
		{Rank: 3, PlayerID: "c", Points: 16},
		{Rank: 4, PlayerID: "d", Points: 8},
		{Rank: 5, PlayerID: "e", Points: 4},
		{Rank: 6, PlayerID: "f", Points: 2},
		{Rank: 7, PlayerID: "g", Points: 1},
		{Rank: 8, PlayerID: "h", Points: 0},
	}

	r := GenerateMexicanoRound(standings, 2, 2)

	// Court 1: rank1+rank4 vs rank2+rank3 → {a,d} vs {b,c}
	if !matchHasTeams(r.Matches[0], [2]string{"a", "d"}, [2]string{"b", "c"}) {
		t.Errorf("court 1: want {a,d} vs {b,c}, got %v vs %v", r.Matches[0].TeamA, r.Matches[0].TeamB)
	}

	// Court 2: rank5+rank8 vs rank6+rank7 → {e,h} vs {f,g}
	if !matchHasTeams(r.Matches[1], [2]string{"e", "h"}, [2]string{"f", "g"}) {
		t.Errorf("court 2: want {e,h} vs {f,g}, got %v vs %v", r.Matches[1].TeamA, r.Matches[1].TeamB)
	}
}

// TestMexicanoProgression simulates multiple rounds and verifies pairings
// shift as the standings change between rounds.
//
// GenerateMexicanoRound does not sort — it uses the standings slice as-is,
// exactly as GetLeaderboard provides it (already sorted by points desc).
func TestMexicanoProgression(t *testing.T) {
	// After round 1: p1/p2/p5/p6 won (24 pts each), p3/p4/p7/p8 lost (0 pts).
	// GetLeaderboard returns them sorted: all 24-pt players first.
	round2Standings := []domain.Standing{
		{Rank: 1, PlayerID: "p1", Points: 24},
		{Rank: 2, PlayerID: "p2", Points: 24},
		{Rank: 3, PlayerID: "p5", Points: 24},
		{Rank: 4, PlayerID: "p6", Points: 24},
		{Rank: 5, PlayerID: "p3", Points: 0},
		{Rank: 6, PlayerID: "p4", Points: 0},
		{Rank: 7, PlayerID: "p7", Points: 0},
		{Rank: 8, PlayerID: "p8", Points: 0},
	}

	r2 := GenerateMexicanoRound(round2Standings, 2, 2)

	// Court 1 must contain only the top-4 (24 pts each).
	topFour := map[string]bool{"p1": true, "p2": true, "p5": true, "p6": true}
	for pid := range courtPlayers(r2, 0) {
		if !topFour[pid] {
			t.Errorf("round 2 court 1 should have top-4 (24 pts), got %s", pid)
		}
	}
	// Court 2 must contain only the bottom-4 (0 pts).
	bottomFour := map[string]bool{"p3": true, "p4": true, "p7": true, "p8": true}
	for pid := range courtPlayers(r2, 1) {
		if !bottomFour[pid] {
			t.Errorf("round 2 court 2 should have bottom-4 (0 pts), got %s", pid)
		}
	}

	// After round 2: p1/p5 win court 1 (now 48 pts each), p2/p6 lose (stay 24 pts).
	// p3/p7 win court 2 (now 24 pts each), p4/p8 lose (stay 0 pts).
	// Sorted standings: p1=48, p5=48, p2=24, p6=24, p3=24, p7=24, p4=0, p8=0.
	round3Standings := []domain.Standing{
		{Rank: 1, PlayerID: "p1", Points: 48},
		{Rank: 2, PlayerID: "p5", Points: 48},
		{Rank: 3, PlayerID: "p2", Points: 24},
		{Rank: 4, PlayerID: "p6", Points: 24},
		{Rank: 5, PlayerID: "p3", Points: 24},
		{Rank: 6, PlayerID: "p7", Points: 24},
		{Rank: 7, PlayerID: "p4", Points: 0},
		{Rank: 8, PlayerID: "p8", Points: 0},
	}

	r3 := GenerateMexicanoRound(round3Standings, 2, 3)

	// Court 1 must contain the new top-4: p1, p5 (48 pts) + p2, p6 (24 pts).
	newTopFour := map[string]bool{"p1": true, "p5": true, "p2": true, "p6": true}
	for pid := range courtPlayers(r3, 0) {
		if !newTopFour[pid] {
			t.Errorf("round 3 court 1 should have new top-4 (p1/p5/p2/p6), got %s", pid)
		}
	}
	// Court 2 must contain the new bottom-4: p3, p7 (24 pts) + p4, p8 (0 pts).
	newBottomFour := map[string]bool{"p3": true, "p7": true, "p4": true, "p8": true}
	for pid := range courtPlayers(r3, 1) {
		if !newBottomFour[pid] {
			t.Errorf("round 3 court 2 should have new bottom-4 (p3/p7/p4/p8), got %s", pid)
		}
	}
}

// TestMexicanoUnlikeAmericano_NoPrecomputedRounds confirms that Mexicano does
// NOT use TotalRounds — it has no fixed number of rounds.
// Specifically: calling GenerateMexicanoRound with any round number should work,
// meaning there is no built-in limit on how many rounds can be played.
func TestMexicanoUnlikeAmericano_NoPrecomputedRounds(t *testing.T) {
	standings := makeStandings("p1", "p2", "p3", "p4", "p5", "p6", "p7", "p8")

	for roundNum := 1; roundNum <= 20; roundNum++ {
		r := GenerateMexicanoRound(standings, 2, roundNum)
		if r.Number != roundNum {
			t.Errorf("round %d: got number %d", roundNum, r.Number)
		}
		if len(r.Matches) != 2 {
			t.Errorf("round %d: expected 2 matches, got %d", roundNum, len(r.Matches))
		}
	}
}

// TestMexicanoUnlikeAmericano_PairingsDeterministic verifies that unlike
// Americano (which uses backtracking to minimise repeats), Mexicano pairings
// are purely positional and therefore deterministic: same standings → same
// player pairs face each other (Team A/B sides are still randomised).
func TestMexicanoUnlikeAmericano_PairingsDeterministic(t *testing.T) {
	standings := makeStandings("p1", "p2", "p3", "p4", "p5", "p6", "p7", "p8")

	// Run 20 times and verify the same two teams always face each other on each court.
	// The first call establishes the expected pairings.
	ref := GenerateMexicanoRound(standings, 2, 2)
	for run := 0; run < 20; run++ {
		r := GenerateMexicanoRound(standings, 2, 2)
		for i := range ref.Matches {
			ra, rb := ref.Matches[i].TeamA, ref.Matches[i].TeamB
			if !matchHasTeams(r.Matches[i], ra, rb) {
				t.Errorf("run %d match %d: pairings changed — want {%v,%v}, got {%v,%v}",
					run, i, ra, rb, r.Matches[i].TeamA, r.Matches[i].TeamB)
			}
		}
	}
}

// TestMexicanoTeamSidesRandomised verifies that over many rounds the same player
// appears on both Team A and Team B — i.e. the side shuffle is working.
func TestMexicanoTeamSidesRandomised(t *testing.T) {
	standings := makeStandings("p1", "p2", "p3", "p4", "p5", "p6", "p7", "p8")

	seenA := false
	seenB := false
	for round := 1; round <= 50; round++ {
		r := GenerateMexicanoRound(standings, 2, round)
		for _, m := range r.Matches {
			for _, pid := range m.TeamA {
				if pid == "p1" {
					seenA = true
				}
			}
			for _, pid := range m.TeamB {
				if pid == "p1" {
					seenB = true
				}
			}
		}
		if seenA && seenB {
			return // pass
		}
	}
	if !seenA {
		t.Error("p1 was never on Team A across 50 rounds")
	}
	if !seenB {
		t.Error("p1 was never on Team B across 50 rounds")
	}
}
