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

// simulateUnlimited plays an unlimited Americano session the way the service does:
// round 1 comes from GenerateRounds(…, 1) (empty history), every later round from
// GenerateNextRound over the accumulated history.
func simulateUnlimited(t *testing.T, players []domain.Player, courts, rounds int) []domain.Round {
	t.Helper()
	history := GenerateRounds(players, courts, 1)
	for len(history) < rounds {
		nr := GenerateNextRound(history, players, courts)
		if nr == nil {
			t.Fatalf("GenerateNextRound returned nil at round %d", len(history)+1)
		}
		history = append(history, *nr)
	}
	return history
}

// partnershipOpponents maps each partnership in a round to the partnership it
// faced, keyed by canonical sorted-pair string. Both directions are recorded.
func partnershipOpponents(r domain.Round) map[string]string {
	m := map[string]string{}
	for _, match := range r.Matches {
		a := sortedPair(match.TeamA)
		b := sortedPair(match.TeamB)
		m[a] = b
		m[b] = a
	}
	return m
}

// TestGenerateNextRound_RecurringPairFacesFreshOpponent is the core guard for
// #271: when a partnership recurs, it must face a *different* opposing pair than
// it faced in its previous occurrence. The old objective went flat once partner
// counts saturated and replayed identical rounds, so a recurring pair kept facing
// the same opponent. Must hold for every random tie-break, so no seeding.
func TestGenerateNextRound_RecurringPairFacesFreshOpponent(t *testing.T) {
	cases := []struct{ players, courts int }{{8, 2}, {9, 2}}
	for _, tc := range cases {
		history := simulateUnlimited(t, makePlayers(tc.players), tc.courts, 18)
		prevOpponent := map[string]string{}
		prevRound := map[string]int{}
		for _, r := range history {
			for pair, opp := range partnershipOpponents(r) {
				if last, seen := prevOpponent[pair]; seen && last == opp {
					t.Errorf("%dp/%dc: partnership %s faced the same opponent %s in round %d and its previous occurrence (round %d)",
						tc.players, tc.courts, pair, opp, r.Number, prevRound[pair])
				}
				prevOpponent[pair] = opp
				prevRound[pair] = r.Number
			}
		}
	}
}

// TestGenerateNextRound_NoConsecutiveMatchupRepeat verifies a Match-up tuple never
// recurs in the immediately following round when an alternative grouping exists
// (it always does for ≥2 courts). Acceptance criterion for #271.
func TestGenerateNextRound_NoConsecutiveMatchupRepeat(t *testing.T) {
	cases := []struct{ players, courts int }{{8, 2}, {12, 3}}
	for _, tc := range cases {
		history := simulateUnlimited(t, makePlayers(tc.players), tc.courts, 18)
		for i := 1; i < len(history); i++ {
			prev := roundMatchups(history[i-1])
			for _, mk := range roundMatchupList(history[i]) {
				if prev[mk] {
					t.Errorf("%dp/%dc: match-up %v recurred in consecutive rounds %d→%d",
						tc.players, tc.courts, mk, history[i-1].Number, history[i].Number)
				}
			}
		}
	}
}

// TestGenerateNextRound_NoRepeatUntilExhausted verifies the achievable form of
// #271's "no Match-up repeats until all distinct tuples are exhausted": because
// the partner step fixes a round's pairs before court assignment, the guarantee
// is *local* — a round reuses a Match-up only when every grouping of that round's
// own pairs already contains a seen Match-up (no all-fresh grouping exists). This
// is exactly what the count-primary objective enforces (an all-fresh grouping has
// score sum 0, the global minimum, so it is always chosen when it exists).
func TestGenerateNextRound_NoRepeatUntilExhausted(t *testing.T) {
	cases := []struct{ players, courts int }{{8, 2}, {12, 3}}
	for _, tc := range cases {
		history := simulateUnlimited(t, makePlayers(tc.players), tc.courts, 24)
		seen := map[[4]string]bool{} // Match-ups seen strictly before the current round
		for _, r := range history {
			pairs := roundPairs(r)
			roundRepeats := false
			for _, mk := range roundMatchupList(r) {
				if seen[mk] {
					roundRepeats = true
					break
				}
			}
			if roundRepeats && hasAllFreshGrouping(pairs, seen) {
				t.Errorf("%dp/%dc round %d reused a Match-up though an all-fresh grouping of its pairs existed",
					tc.players, tc.courts, r.Number)
			}
			for _, mk := range roundMatchupList(r) {
				seen[mk] = true
			}
		}
	}
}

// roundPairs returns the partnerships (as pairs) that played in a round.
func roundPairs(r domain.Round) [][2]string {
	pairs := make([][2]string, 0, len(r.Matches)*2)
	for _, m := range r.Matches {
		pairs = append(pairs, m.TeamA, m.TeamB)
	}
	return pairs
}

// hasAllFreshGrouping reports whether the pairs can be grouped into courts such
// that no resulting Match-up tuple is in seen. It enumerates every grouping the
// court-assignment search itself considers.
func hasAllFreshGrouping(pairs [][2]string, seen map[[4]string]bool) bool {
	used := make([]bool, len(pairs))
	var rec func() bool
	rec = func() bool {
		first := -1
		for i := range pairs {
			if !used[i] {
				first = i
				break
			}
		}
		if first == -1 {
			return true // all pairs placed, none hit a seen Match-up
		}
		used[first] = true
		defer func() { used[first] = false }()
		for second := first + 1; second < len(pairs); second++ {
			if used[second] {
				continue
			}
			mk := matchupKey(pairs[first][0], pairs[first][1], pairs[second][0], pairs[second][1])
			if seen[mk] {
				continue
			}
			used[second] = true
			if rec() {
				used[second] = false
				return true
			}
			used[second] = false
		}
		return false
	}
	return rec()
}

// roundMatchups returns the set of canonical Match-up tuples in a round.
func roundMatchups(r domain.Round) map[[4]string]bool {
	set := map[[4]string]bool{}
	for _, mk := range roundMatchupList(r) {
		set[mk] = true
	}
	return set
}

func roundMatchupList(r domain.Round) [][4]string {
	list := make([][4]string, 0, len(r.Matches))
	for _, m := range r.Matches {
		list = append(list, matchupKey(m.TeamA[0], m.TeamA[1], m.TeamB[0], m.TeamB[1]))
	}
	return list
}

// TestGenerateNextRound_1CourtBenchVarietyIsOptimal is the #274 guard. On 1 court
// the court-assignment step has no freedom, so opponent variety can only come from
// which player is benched and how the four active players are paired. This test
// proves the scheduler picks the lexicographic optimum — partnerRepeats → matchupCount
// → recency — over *every fair bench choice and every pairing*. Optimality is the
// exact statement of "a recurring pair faces a fresh opponent whenever a fair bench
// choice allows it": if a fresher round were reachable at no greater partner-repeat
// cost, the optimum would take it. Holds for every random tie-break, so no seeding.
func TestGenerateNextRound_1CourtBenchVarietyIsOptimal(t *testing.T) {
	for _, players := range []int{5, 6} {
		history := GenerateRounds(makePlayers(players), 1, 1)
		for len(history) < 24 {
			pc, ms, benchTotal, lastBench := replayState(history)
			roundNum := len(history) + 1
			canBench := eligibleBench(makePlayers(players), lastBench, roundNum)

			want := brute1CourtOptimum(makePlayers(players), canBench, benchTotal, pc, ms)

			nr := GenerateNextRound(history, makePlayers(players), 1)
			m := nr.Matches[0]
			got := courtKey3(m.TeamA, m.TeamB, pc, ms)

			if lessKey(want, got) { // got strictly worse than the achievable optimum
				t.Fatalf("%dp/1c round %d: scheduler key %v is worse than optimum %v",
					players, roundNum, got, want)
			}
			history = append(history, *nr)
		}
	}
}

// TestGenerateNextRound_BenchStaysFair verifies #274's hard constraint: variety-aware
// bench selection never sacrifices fairness. Over a long unlimited session the bench
// counts stay within 1 of each other, and no one benched last round is benched again.
func TestGenerateNextRound_BenchStaysFair(t *testing.T) {
	cases := []struct{ players, courts int }{{5, 1}, {6, 1}, {9, 2}, {11, 2}}
	for _, tc := range cases {
		history := simulateUnlimited(t, makePlayers(tc.players), tc.courts, 30)
		benchCount := map[string]int{}
		for i, r := range history {
			if i > 0 {
				prev := map[string]bool{}
				for _, id := range history[i-1].Bench {
					prev[id] = true
				}
				for _, id := range r.Bench {
					if prev[id] {
						t.Errorf("%dp/%dc round %d: %s benched two rounds running", tc.players, tc.courts, r.Number, id)
					}
				}
			}
			for _, id := range r.Bench {
				benchCount[id]++
			}
		}
		mn, mx := 1<<30, -1
		for _, p := range makePlayers(tc.players) {
			c := benchCount[p.ID]
			if c < mn {
				mn = c
			}
			if c > mx {
				mx = c
			}
		}
		if mx-mn > 1 {
			t.Errorf("%dp/%dc: bench counts uneven (spread %d..%d)", tc.players, tc.courts, mn, mx)
		}
	}
}

// replayState rebuilds the scheduler's history-derived state the way GenerateNextRound
// does, for use as an independent oracle in tests.
func replayState(history []domain.Round) (pc map[[2]string]int, ms map[[4]string]matchupStat, benchTotal, lastBench map[string]int) {
	pc = map[[2]string]int{}
	ms = map[[4]string]matchupStat{}
	benchTotal = map[string]int{}
	lastBench = map[string]int{}
	for _, r := range history {
		for _, id := range r.Bench {
			benchTotal[id]++
			lastBench[id] = r.Number
		}
		for _, m := range r.Matches {
			pc[pairKey(m.TeamA[0], m.TeamA[1])]++
			pc[pairKey(m.TeamB[0], m.TeamB[1])]++
			mk := matchupKey(m.TeamA[0], m.TeamA[1], m.TeamB[0], m.TeamB[1])
			st := ms[mk]
			st.count++
			st.lastRound = r.Number
			ms[mk] = st
		}
	}
	return pc, ms, benchTotal, lastBench
}

// eligibleBench returns the players allowed to be benched this round (everyone not
// benched in the immediately preceding round).
func eligibleBench(players []domain.Player, lastBench map[string]int, roundNum int) []string {
	var canBench []string
	for _, p := range players {
		if lastBench[p.ID] != roundNum-1 {
			canBench = append(canBench, p.ID)
		}
	}
	return canBench
}

// brute1CourtOptimum returns the best achievable partnerRepeats → matchupCount →
// recency key over every fair bench choice and every pairing of the four remaining
// players, for a single-court round.
func brute1CourtOptimum(players []domain.Player, canBench []string, benchTotal map[string]int, pc map[[2]string]int, ms map[[4]string]matchupStat) [3]int {
	benchSize := len(players) - 4 // single court
	best := [3]int{1 << 30, 0, 0}
	for _, cand := range fairBenchChoices(canBench, benchTotal, benchSize) {
		benched := map[string]bool{}
		for _, id := range cand {
			benched[id] = true
		}
		var act []string
		for _, p := range players {
			if !benched[p.ID] {
				act = append(act, p.ID)
			}
		}
		splits := [3][2][2]string{
			{{act[0], act[1]}, {act[2], act[3]}},
			{{act[0], act[2]}, {act[1], act[3]}},
			{{act[0], act[3]}, {act[1], act[2]}},
		}
		for _, s := range splits {
			if k := courtKey3(s[0], s[1], pc, ms); lessKey(k, best) {
				best = k
			}
		}
	}
	return best
}

// courtKey3 is the partnerRepeats → matchupCount → recency key for a single court,
// the test-side mirror of the scheduler's objective (rating omitted: tests here use
// an unrated field, so the rating term is zero).
func courtKey3(a, b [2]string, pc map[[2]string]int, ms map[[4]string]matchupStat) [3]int {
	k := [3]int{pc[pairKey(a[0], a[1])] + pc[pairKey(b[0], b[1])], 0, 0}
	st := ms[matchupKey(a[0], a[1], b[0], b[1])]
	k[1] = st.count
	if st.count > 0 {
		k[2] = st.lastRound
	}
	return k
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
