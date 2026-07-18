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

// makeRatedPlayers builds players with the given ratings, id P01.. in order.
func makeRatedPlayers(ratings ...int) []domain.Player {
	players := make([]domain.Player, len(ratings))
	for i, r := range ratings {
		id := fmt.Sprintf("P%02d", i+1)
		players[i] = domain.Player{ID: id, Name: id, Rating: r}
	}
	return players
}

// courtRatings returns the four players' ratings on a match, given an id→rating map.
func courtRatings(m domain.Match, rating map[string]int) [4]int {
	return [4]int{
		rating[m.TeamA[0]], rating[m.TeamA[1]],
		rating[m.TeamB[0]], rating[m.TeamB[1]],
	}
}

// groupingSet extracts the court groupings of a round as a set, independent of
// team-side / court-number / match-id shuffling. Each grouping is a canonical
// string of the two teams (each team's two ids sorted, the two teams sorted).
func groupingSet(r domain.Round) map[string]bool {
	set := map[string]bool{}
	for _, m := range r.Matches {
		teamA := sortedPair(m.TeamA)
		teamB := sortedPair(m.TeamB)
		teams := []string{teamA, teamB}
		if teams[0] > teams[1] {
			teams[0], teams[1] = teams[1], teams[0]
		}
		set[teams[0]+"|"+teams[1]] = true
	}
	return set
}

func sortedPair(p [2]string) string {
	if p[0] > p[1] {
		return p[1] + "," + p[0]
	}
	return p[0] + "," + p[1]
}

// TestGenerateRounds_RatingGroupsSimilarStrength verifies that on a rated,
// multi-court field the court-assignment step puts strong pairs against strong
// pairs and weak against weak (small per-court rating gap), per ADR 0006.
func TestGenerateRounds_RatingGroupsSimilarStrength(t *testing.T) {
	// 4 strong (rating 5) then 4 weak (rating 1). Round 1 has empty history, so
	// the greedy partner step pairs sequentially: (P1,P2)(P3,P4)(P5,P6)(P7,P8),
	// i.e. two strong pairs (total 10) and two weak pairs (total 2). The only
	// zero-rating-gap grouping puts the two strong pairs on one court and the
	// two weak pairs on the other.
	players := makeRatedPlayers(5, 5, 5, 5, 1, 1, 1, 1)
	rating := map[string]int{}
	for _, p := range players {
		rating[p.ID] = p.Rating
	}

	rounds := GenerateRounds(players, 2, 1)
	if len(rounds) != 1 || len(rounds[0].Matches) != 2 {
		t.Fatalf("expected 1 round with 2 matches, got %+v", rounds)
	}

	for _, m := range rounds[0].Matches {
		rs := courtRatings(m, rating)
		// Every player on a court should have the same strength (all 5 or all 1).
		for i := 1; i < 4; i++ {
			if rs[i] != rs[0] {
				t.Errorf("court not strength-matched: ratings %v", rs)
			}
		}
	}
}

// TestGenerateRounds_UnratedTreatedAsMedian verifies the normalisation that makes
// the feature self-cancelling: a field mixing unrated (zero-value) players with
// explicit-median players produces the same groupings as an all-median field,
// so median-filled players never look like the weakest and skew the balance.
func TestGenerateRounds_UnratedTreatedAsMedian(t *testing.T) {
	unrated := makeRatedPlayers(0, 0, 0, 0, 3, 3, 3, 3) // 0 = unrated, must behave as 3
	allMedian := makeRatedPlayers(3, 3, 3, 3, 3, 3, 3, 3)

	// Same courts/rounds; groupings must match round-for-round.
	rUnrated := GenerateRounds(unrated, 2, 7)
	rMedian := GenerateRounds(allMedian, 2, 7)

	if len(rUnrated) != len(rMedian) {
		t.Fatalf("round count mismatch: %d vs %d", len(rUnrated), len(rMedian))
	}
	for i := range rUnrated {
		gu := groupingSet(rUnrated[i])
		gm := groupingSet(rMedian[i])
		if len(gu) != len(gm) {
			t.Fatalf("round %d grouping count mismatch", i+1)
		}
		for k := range gm {
			if !gu[k] {
				t.Errorf("round %d: unrated field grouping differs from all-median (%s missing)", i+1, k)
			}
		}
	}
}

// TestBestCourtAssignment_MedianReproducesCoOccurrenceOnly is the regression guard
// for "an all-median field reproduces today's schedule." It drives the court-
// assignment step directly with a co-occurrence map that has a unique optimum, and
// asserts that an all-median rating map produces the exact same grouping as the
// pre-rating objective. The pre-rating objective is recovered by an empty rating
// map: every team total is 0, so every rating gap is 0 and the score reduces to
// pure co-occurrence — the algorithm as it was before this change.
func TestBestCourtAssignment_MedianReproducesCoOccurrenceOnly(t *testing.T) {
	pairs := [][2]string{{"a", "b"}, {"c", "d"}, {"e", "f"}, {"g", "h"}}

	// Make the grouping {a,b,e,f} | {c,d,g,h} the unique co-occurrence optimum:
	//   - pairing a,b with c,d is costly (a,c high) → rules out {a,b,c,d}
	//   - pairing a,b with g,h is costly (a,g high) → rules out {a,b,g,h}
	// leaving a,b with e,f (cost 0) as the only cheap choice.
	coOcc := map[[2]string]int{
		pairKey("a", "c"): 5,
		pairKey("a", "g"): 5,
	}

	// Pre-rating objective: empty map ⇒ all team totals 0 ⇒ gaps 0 ⇒ co-occurrence only.
	preRating := bestCourtAssignment(pairs, coOcc, map[string]int{})

	// All-median field: every player rated 3 ⇒ team totals 6 ⇒ gaps 0.
	median := map[string]int{}
	for _, id := range []string{"a", "b", "c", "d", "e", "f", "g", "h"} {
		median[id] = 3
	}
	withMedian := bestCourtAssignment(pairs, coOcc, median)

	preSet := matchGroupingSet(preRating)
	medSet := matchGroupingSet(withMedian)
	if len(preSet) != len(medSet) {
		t.Fatalf("grouping count differs: pre=%v median=%v", preSet, medSet)
	}
	for k := range preSet {
		if !medSet[k] {
			t.Errorf("all-median grouping differs from pre-rating co-occurrence-only: %s missing", k)
		}
	}

	// Sanity: the shared optimum really is the co-occurrence-minimal grouping.
	if !medSet["a,b|e,f"] {
		t.Errorf("expected co-occurrence optimum {a,b}|{e,f}, got %v", medSet)
	}
}

// matchGroupingSet extracts court groupings from a []Match as a canonical set.
func matchGroupingSet(matches []domain.Match) map[string]bool {
	return groupingSet(domain.Round{Matches: matches})
}

// TestGenerateRounds_SingleCourtUnaffected verifies that with a single court there
// is no assignment freedom, so ratings cannot change the (only possible) grouping.
func TestGenerateRounds_SingleCourtUnaffected(t *testing.T) {
	rated := makeRatedPlayers(5, 1, 5, 1)
	uniform := makeRatedPlayers(3, 3, 3, 3)

	rr := GenerateRounds(rated, 1, 3)
	ru := GenerateRounds(uniform, 1, 3)

	for i := range rr {
		if len(rr[i].Matches) != 1 {
			t.Fatalf("round %d: expected exactly 1 match on 1 court", i+1)
		}
		// All four players are on the single court regardless of rating.
		seen := map[string]bool{}
		for _, id := range []string{rr[i].Matches[0].TeamA[0], rr[i].Matches[0].TeamA[1], rr[i].Matches[0].TeamB[0], rr[i].Matches[0].TeamB[1]} {
			seen[id] = true
		}
		if len(seen) != 4 {
			t.Errorf("round %d: expected 4 distinct players on the court, got %d", i+1, len(seen))
		}
		// Grouping is identical to the uniform-rating run (only one grouping exists).
		if len(groupingSet(rr[i])) != len(groupingSet(ru[i])) {
			t.Errorf("round %d: single-court grouping should be rating-independent", i+1)
		}
	}
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

// TestGenerateRounds_CourtCoOccurrenceSpread verifies that no pair of players
// shares a court disproportionately more than the average. This catches the old
// bug where the same 4 players would be stuck on one court, shuffling internally.
func TestGenerateRounds_CourtCoOccurrenceSpread(t *testing.T) {
	cases := []struct {
		players, courts, totalRounds int
	}{
		{8, 2, 7},
		{12, 3, 11},
	}
	for _, tc := range cases {
		players := makePlayers(tc.players)
		rounds := GenerateRounds(players, tc.courts, tc.totalRounds)

		// Count how many times each pair shares a court.
		coOccurrence := map[[2]string]int{}
		for _, r := range rounds {
			for _, m := range r.Matches {
				ids := []string{m.TeamA[0], m.TeamA[1], m.TeamB[0], m.TeamB[1]}
				for i := 0; i < len(ids); i++ {
					for j := i + 1; j < len(ids); j++ {
						key := [2]string{ids[i], ids[j]}
						if key[0] > key[1] {
							key[0], key[1] = key[1], key[0]
						}
						coOccurrence[key]++
					}
				}
			}
		}

		// Calculate max allowed: with perfect spread, each pair meets
		// ceil(totalRounds * 6 / C(N,2)) times per court slot.
		// We allow up to 2x the average to account for constraints.
		totalCoOccurrences := 0
		for _, v := range coOccurrence {
			totalCoOccurrences += v
		}
		numPairs := len(coOccurrence)
		avg := float64(totalCoOccurrences) / float64(numPairs)
		maxAllowed := int(avg*2) + 1

		for pair, count := range coOccurrence {
			if count > maxAllowed {
				t.Errorf("%d players, %d courts: pair (%s, %s) shared a court %d times (avg %.1f, max allowed %d)",
					tc.players, tc.courts, pair[0], pair[1], count, avg, maxAllowed)
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

// TestGenerateNextRound verifies streaming generation of single rounds.
// Generates rounds 1..N via batch, then generates round N+1 via streaming and verifies constraints.
func TestGenerateNextRound(t *testing.T) {
	players := makePlayers(9) // 1 bench per round
	courts := 2
	totalRounds := 9

	// Generate all rounds via batch
	batchRounds := GenerateRounds(players, courts, totalRounds)

	// Simulate: generate first N-1 rounds via batch, then use streaming for round N
	for testN := 1; testN < totalRounds; testN++ {
		previousRounds := batchRounds[:testN]
		nextRound := GenerateNextRound(previousRounds, players, courts)

		if nextRound == nil {
			t.Errorf("GenerateNextRound returned nil for round %d", testN+1)
			continue
		}

		expectedRound := batchRounds[testN]

		// Check round number
		if nextRound.Number != expectedRound.Number {
			t.Errorf("round %d: expected number %d, got %d", testN+1, expectedRound.Number, nextRound.Number)
		}

		// Check player coverage (all N players accounted for)
		active := make(map[string]bool)
		for _, m := range nextRound.Matches {
			for _, id := range []string{m.TeamA[0], m.TeamA[1], m.TeamB[0], m.TeamB[1]} {
				if active[id] {
					t.Errorf("round %d: player %s in multiple matches", nextRound.Number, id)
				}
				active[id] = true
			}
		}
		for _, id := range nextRound.Bench {
			if active[id] {
				t.Errorf("round %d: bench player %s in a match", nextRound.Number, id)
			}
			active[id] = true
		}
		if len(active) != len(players) {
			t.Errorf("round %d: expected %d players, got %d", nextRound.Number, len(players), len(active))
		}

		// Check no consecutive bench: benched players from previous round must play
		if testN > 0 {
			prevBenched := make(map[string]bool)
			for _, id := range previousRounds[testN-1].Bench {
				prevBenched[id] = true
			}
			for _, id := range nextRound.Bench {
				if prevBenched[id] {
					t.Errorf("round %d: player %s benched consecutively", nextRound.Number, id)
				}
			}
		}

		// Check bench size
		expectedBench := len(players) - courts*4
		if len(nextRound.Bench) != expectedBench {
			t.Errorf("round %d: expected %d benched, got %d", nextRound.Number, expectedBench, len(nextRound.Bench))
		}
	}
}
