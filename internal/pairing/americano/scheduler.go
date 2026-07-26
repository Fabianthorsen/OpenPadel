package americano

import (
	"crypto/rand"
	"math/big"
	"sort"

	"github.com/fabianthorsen/openpadel/internal/domain"
)

// ratingWeight (W) blends skill balance into the court-assignment score:
// score = ratingGap·W + coOccurrence. It is tuned large enough that the rating
// gap usually dominates the grouping choice, while co-occurrence still breaks
// near-ties (groupings with an equal rating gap). Team rating totals span 2–10,
// so a court's rating gap is 0–8; co-occurrence deltas between candidate
// groupings are small single/low-double-digit integers, so W=100 makes any
// difference in rating gap outweigh co-occurrence, which then acts as the
// tiebreaker. See ADR 0006.
const ratingWeight = 100

// buildRatingMap builds an id→rating lookup from the players, normalising any
// out-of-range value (including the zero value of an unrated player) to the
// neutral median. This is what makes the feature self-cancelling: an all-unrated
// field maps to all-median, every court has a zero rating gap, the rating term
// vanishes, and the schedule is identical to today. See ADR 0006.
func buildRatingMap(players []domain.Player) map[string]int {
	rating := make(map[string]int, len(players))
	for _, p := range players {
		rating[p.ID] = domain.NormalizeRating(p.Rating)
	}
	return rating
}

// GenerateRounds produces all rounds for an Americano session upfront.
// Hard constraint: a player benched in round N must play in round N+1.
// Core invariant: minimise partner and matchup repeats across the full tournament.
func GenerateRounds(players []domain.Player, courts, totalRounds int) []domain.Round {
	ids := make([]string, len(players))
	for i, p := range players {
		ids[i] = p.ID
	}
	rating := buildRatingMap(players)

	benchSize := len(ids) - courts*4

	lastBenchedRound := make(map[string]int) // 0 = never benched
	benchTotal := make(map[string]int)

	// Full history — not just the previous round.
	partnerCount := map[[2]string]int{}
	courtShareCount := map[[2]string]int{}

	rounds := make([]domain.Round, totalRounds)

	for r := 0; r < totalRounds; r++ {
		roundNum := r + 1

		// Players who sat out last round must play this round.
		mustPlay := make(map[string]bool)
		if roundNum > 1 {
			for _, id := range ids {
				if lastBenchedRound[id] == roundNum-1 {
					mustPlay[id] = true
				}
			}
		}

		var forced, canBench []string
		for _, id := range ids {
			if mustPlay[id] {
				forced = append(forced, id)
			} else {
				canBench = append(canBench, id)
			}
		}

		// From canBench, those with fewest bench turns are most "due" to sit.
		sort.Slice(canBench, func(i, j int) bool {
			return benchTotal[canBench[i]] < benchTotal[canBench[j]]
		})

		var bench, active []string
		if benchSize > 0 {
			actualBenchSize := benchSize
			if actualBenchSize > len(canBench) {
				actualBenchSize = len(canBench)
			}
			bench = canBench[:actualBenchSize]
			active = append(forced, canBench[actualBenchSize:]...)
		} else {
			active = append([]string{}, ids...)
		}

		for _, id := range bench {
			lastBenchedRound[id] = roundNum
			benchTotal[id]++
		}

		matches := assignCourts(active, partnerCount, courtShareCount, rating)
		shuffleTeamSides(matches)
		shuffleCourtNumbers(matches)

		for _, m := range matches {
			partnerCount[pairKey(m.TeamA[0], m.TeamA[1])]++
			partnerCount[pairKey(m.TeamB[0], m.TeamB[1])]++
			// Track all 6 pairwise co-occurrences on this court.
			players := []string{m.TeamA[0], m.TeamA[1], m.TeamB[0], m.TeamB[1]}
			for i := 0; i < len(players); i++ {
				for j := i + 1; j < len(players); j++ {
					courtShareCount[pairKey(players[i], players[j])]++
				}
			}
		}

		benchIDs := make([]string, len(bench))
		copy(benchIDs, bench)

		rounds[r] = domain.Round{
			ID:      shortID(),
			Number:  roundNum,
			Bench:   benchIDs,
			Matches: matches,
		}
	}

	return rounds
}

// GenerateNextRound produces a single round given the history of previously-played rounds.
// Rebuilds constraint state from previousRounds to ensure consistent pairing strategy.
// Stateless: no session-level memoization, history read entirely from rounds slice.
func GenerateNextRound(previousRounds []domain.Round, players []domain.Player, courts int) *domain.Round {
	ids := make([]string, len(players))
	for i, p := range players {
		ids[i] = p.ID
	}
	rating := buildRatingMap(players)

	benchSize := len(ids) - courts*4

	// Rebuild state from history. Unlimited mode optimises for match-up variety
	// (see ADR 0006 amendment), so it tracks the full four-player match-up tuple —
	// how often it has occurred and the round it last occurred — rather than the
	// pairwise co-occurrence the limited/upfront path uses.
	lastBenchedRound := make(map[string]int)
	benchTotal := make(map[string]int)
	partnerCount := map[[2]string]int{}
	matchupCount := map[[4]string]int{}
	matchupLastRound := map[[4]string]int{}

	for _, prevRound := range previousRounds {
		// Track bench
		for _, id := range prevRound.Bench {
			lastBenchedRound[id] = prevRound.Number
			benchTotal[id]++
		}

		// Track partner repeats and match-up recurrence/recency.
		for _, m := range prevRound.Matches {
			partnerCount[pairKey(m.TeamA[0], m.TeamA[1])]++
			partnerCount[pairKey(m.TeamB[0], m.TeamB[1])]++
			mk := matchupKey(m.TeamA[0], m.TeamA[1], m.TeamB[0], m.TeamB[1])
			matchupCount[mk]++
			matchupLastRound[mk] = prevRound.Number
		}
	}

	// Generate next round
	roundNum := len(previousRounds) + 1

	// Players who sat out last round must play this round
	mustPlay := make(map[string]bool)
	if len(previousRounds) > 0 {
		for _, id := range ids {
			if lastBenchedRound[id] == roundNum-1 {
				mustPlay[id] = true
			}
		}
	}

	var forced, canBench []string
	for _, id := range ids {
		if mustPlay[id] {
			forced = append(forced, id)
		} else {
			canBench = append(canBench, id)
		}
	}

	// From canBench, those with fewest bench turns are most "due" to sit
	sort.Slice(canBench, func(i, j int) bool {
		return benchTotal[canBench[i]] < benchTotal[canBench[j]]
	})

	var bench, active []string
	if benchSize > 0 {
		actualBenchSize := benchSize
		if actualBenchSize > len(canBench) {
			actualBenchSize = len(canBench)
		}
		bench = canBench[:actualBenchSize]
		active = append(forced, canBench[actualBenchSize:]...)
	} else {
		active = append([]string{}, ids...)
	}

	matches := assignCourtsVariety(active, partnerCount, matchupCount, matchupLastRound, rating)
	shuffleTeamSides(matches)
	shuffleCourtNumbers(matches)

	benchIDs := make([]string, len(bench))
	copy(benchIDs, bench)

	return &domain.Round{
		ID:      shortID(),
		Number:  roundNum,
		Bench:   benchIDs,
		Matches: matches,
	}
}

// pairKey returns a canonical (sorted) key for a partnership.
func pairKey(a, b string) [2]string {
	if a > b {
		return [2]string{b, a}
	}
	return [2]string{a, b}
}

// matchupKey returns a canonical key for a court match-up — the full four-player
// opposition {A+B vs C+D}. Each team's two ids are sorted, and the two teams are
// then sorted against each other, so {A+B vs C+D} == {C+D vs A+B}. This is the
// unit the unlimited scheduler keeps fresh (see ADR 0006 amendment): partnerships
// may recur, but a recurring pair should meet a different opposing pair.
func matchupKey(a0, a1, b0, b1 string) [4]string {
	pa := pairKey(a0, a1)
	pb := pairKey(b0, b1)
	if pa[0] > pb[0] || (pa[0] == pb[0] && pa[1] > pb[1]) {
		pa, pb = pb, pa
	}
	return [4]string{pa[0], pa[1], pb[0], pb[1]}
}

// bestPartnerMatching finds the partner pairing of players that minimises total
// partner-repeat count using backtracking with pruning. For up to 16 players the
// search tree is small enough to be exhaustive (worst case: 10,395 nodes for 12p).
func bestPartnerMatching(players []string, partnerCount map[[2]string]int) [][2]string {
	n := len(players)
	used := make([]bool, n)
	scratch := make([][2]string, n/2)
	best := make([][2]string, n/2)
	bestScore := int(^uint(0) >> 1)

	var bt func(pairIdx, score int)
	bt = func(pairIdx, score int) {
		if score >= bestScore {
			return // prune: can't beat current best
		}
		if pairIdx == n/2 {
			bestScore = score
			copy(best, scratch)
			return
		}
		// Always pair the first unused player — reduces branching factor.
		first := -1
		for i := range players {
			if !used[i] {
				first = i
				break
			}
		}
		used[first] = true
		for second := first + 1; second < n; second++ {
			if !used[second] {
				used[second] = true
				scratch[pairIdx] = [2]string{players[first], players[second]}
				bt(pairIdx+1, score+partnerCount[pairKey(players[first], players[second])]*1000)
				used[second] = false
			}
		}
		used[first] = false
	}

	bt(0, 0)
	return best
}

// bestCourtAssignment groups the given partner pairs into courts, minimising a
// weighted blend of the per-court rating gap and court co-occurrence:
//
//	score = ratingGap·W + coOccurrence
//
// where ratingGap = |teamA_total − teamB_total| pairs a strong team against
// another strong team (and weak against weak), and co-occurrence (how often any
// two players have shared a court) breaks ties. With an all-median (or unrated)
// field every rating gap is zero, so this reduces exactly to the previous
// co-occurrence-only objective. See ADR 0006.
// Number of groupings: for 4 pairs→3, 6 pairs→15, 8 pairs→105.
func bestCourtAssignment(pairs [][2]string, courtShareCount map[[2]string]int, rating map[string]int) []domain.Match {
	numPairs := len(pairs)
	courts := numPairs / 2
	usedPair := make([]bool, numPairs)
	scratch := make([]domain.Match, courts)
	best := make([]domain.Match, courts)
	bestScore := int(^uint(0) >> 1)

	var bt func(courtIdx, score int)
	bt = func(courtIdx, score int) {
		if score >= bestScore {
			return
		}
		if courtIdx == courts {
			bestScore = score
			copy(best, scratch)
			return
		}
		// Fix the first unused pair as TeamA of this court.
		first := -1
		for i := range pairs {
			if !usedPair[i] {
				first = i
				break
			}
		}
		usedPair[first] = true
		for second := first + 1; second < numPairs; second++ {
			if !usedPair[second] {
				usedPair[second] = true
				// Score: rating gap between the two teams (weighted) plus the
				// sum of pairwise co-occurrences for all 6 pairs on this court.
				players := []string{pairs[first][0], pairs[first][1], pairs[second][0], pairs[second][1]}
				courtScore := 0
				for i := 0; i < 4; i++ {
					for j := i + 1; j < 4; j++ {
						courtScore += courtShareCount[pairKey(players[i], players[j])]
					}
				}
				teamATotal := rating[pairs[first][0]] + rating[pairs[first][1]]
				teamBTotal := rating[pairs[second][0]] + rating[pairs[second][1]]
				ratingGap := teamATotal - teamBTotal
				if ratingGap < 0 {
					ratingGap = -ratingGap
				}
				courtScore += ratingGap * ratingWeight
				scratch[courtIdx] = domain.Match{
					ID:    shortID(),
					Court: courtIdx + 1,
					TeamA: [2]string{pairs[first][0], pairs[first][1]},
					TeamB: [2]string{pairs[second][0], pairs[second][1]},
				}
				bt(courtIdx+1, score+courtScore)
				usedPair[second] = false
			}
		}
		usedPair[first] = false
	}

	bt(0, 0)
	return best
}

// assignCourts finds the optimal assignment of active players to courts using
// exact backtracking search: first minimise partner repeats (primary constraint),
// then minimise court co-occurrence (secondary). Guaranteed optimal for up to 16 players.
func assignCourts(active []string, partnerCount map[[2]string]int, courtShareCount map[[2]string]int, rating map[string]int) []domain.Match {
	pairs := bestPartnerMatching(active, partnerCount)
	return bestCourtAssignment(pairs, courtShareCount, rating)
}

// assignCourtsVariety is the unlimited-mode counterpart of assignCourts. It keeps
// the same partner step (minimise partner repeats) but assigns courts with the
// variety-first objective (see bestCourtAssignmentVariety and the ADR 0006
// amendment): partnerships may recur, but a recurring pair meets a fresh opponent.
func assignCourtsVariety(active []string, partnerCount map[[2]string]int, matchupCount, matchupLastRound map[[4]string]int, rating map[string]int) []domain.Match {
	pairs := bestPartnerMatching(active, partnerCount)
	return bestCourtAssignmentVariety(pairs, matchupCount, matchupLastRound, rating)
}

// bestCourtAssignmentVariety groups the given partner pairs into courts with a
// variety-first lexicographic objective, minimised over the whole round:
//
//	key = (matchupCount, recency, ratingGap)
//
// where, per court:
//   - matchupCount = how many times this exact {teamA vs teamB} match-up has
//     occurred → use every distinct match-up before repeating any;
//   - recency = the round number this match-up last occurred (0 if never), so
//     when a repeat is unavoidable the *stalest* match-up is reused first;
//   - ratingGap = |teamA_total − teamB_total|, the ADR 0006 balance term, demoted
//     here to a final tie-breaker (self-cancelling when the field is unrated).
//
// All three components are non-negative, so their per-round sums are monotonic in
// the number of courts placed and the lexicographic prune below is exact. Genuine
// ties are broken randomly by shuffling the pair order before the search, so equal
// rounds are never emitted identically. See ADR 0006 amendment.
func bestCourtAssignmentVariety(pairs [][2]string, matchupCount, matchupLastRound map[[4]string]int, rating map[string]int) []domain.Match {
	shufflePairs(pairs)

	numPairs := len(pairs)
	courts := numPairs / 2
	usedPair := make([]bool, numPairs)
	scratch := make([]domain.Match, courts)
	best := make([]domain.Match, courts)
	var bestScore [3]int
	haveBest := false

	var bt func(courtIdx int, score [3]int)
	bt = func(courtIdx int, score [3]int) {
		if haveBest && !lessKey(score, bestScore) {
			return // prune: partial score already ≥ best (all components monotonic)
		}
		if courtIdx == courts {
			bestScore = score
			haveBest = true
			copy(best, scratch)
			return
		}
		// Fix the first unused pair as TeamA of this court.
		first := -1
		for i := range pairs {
			if !usedPair[i] {
				first = i
				break
			}
		}
		usedPair[first] = true
		for second := first + 1; second < numPairs; second++ {
			if usedPair[second] {
				continue
			}
			usedPair[second] = true
			scratch[courtIdx] = domain.Match{
				ID:    shortID(),
				Court: courtIdx + 1,
				TeamA: [2]string{pairs[first][0], pairs[first][1]},
				TeamB: [2]string{pairs[second][0], pairs[second][1]},
			}
			key := courtVarietyKey(pairs[first], pairs[second], matchupCount, matchupLastRound, rating)
			bt(courtIdx+1, addKey(score, key))
			usedPair[second] = false
		}
		usedPair[first] = false
	}

	bt(0, [3]int{})
	return best
}

// courtVarietyKey scores a single court under the variety-first objective. recency
// is the round number the match-up last occurred (smaller = staler = preferred);
// it is 0 for a never-seen match-up, which is irrelevant there because count is 0
// and dominates. All components are non-negative.
func courtVarietyKey(a, b [2]string, matchupCount, matchupLastRound map[[4]string]int, rating map[string]int) [3]int {
	mk := matchupKey(a[0], a[1], b[0], b[1])
	count := matchupCount[mk]
	recency := 0
	if count > 0 {
		recency = matchupLastRound[mk]
	}
	gap := (rating[a[0]] + rating[a[1]]) - (rating[b[0]] + rating[b[1]])
	if gap < 0 {
		gap = -gap
	}
	return [3]int{count, recency, gap}
}

// lessKey reports whether a is lexicographically smaller than b.
func lessKey(a, b [3]int) bool {
	for i := range a {
		if a[i] != b[i] {
			return a[i] < b[i]
		}
	}
	return false
}

// addKey sums two lexicographic keys component-wise.
func addKey(a, b [3]int) [3]int {
	return [3]int{a[0] + b[0], a[1] + b[1], a[2] + b[2]}
}

// shufflePairs randomises the order of pairs in place (Fisher-Yates) so that the
// variety search, which fixes the first unused pair first, returns a random choice
// among equally-optimal groupings rather than a deterministic one.
func shufflePairs(pairs [][2]string) {
	for i := len(pairs) - 1; i > 0; i-- {
		j, _ := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		pairs[i], pairs[j.Int64()] = pairs[j.Int64()], pairs[i]
	}
}

// TotalRounds returns the correct number of rounds for a fair Americano tournament.
// For no-bench configs (players == courts*4): N-1 rounds covers all unique pairs.
// For bench configs: smallest multiple of N/gcd(N,benchSize) that is >= N-1,
// ensuring everyone sits out equally AND there are enough rounds to be meaningful.
func TotalRounds(players, courts int) int {
	benchSize := players - courts*4
	if benchSize <= 0 {
		return players - 1
	}
	cycle := players / gcd(players, benchSize) // rounds per full bench rotation
	target := players - 1                      // minimum meaningful rounds
	// Round up to the nearest full cycle >= target
	n := (target + cycle - 1) / cycle
	return n * cycle
}

func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

// shuffleTeamSides randomly swaps Team A and Team B on each match so that the
// "Team A" label doesn't persistently attach to any particular player across rounds.
func shuffleTeamSides(matches []domain.Match) {
	for i := range matches {
		n, _ := rand.Int(rand.Reader, big.NewInt(2))
		if n.Int64() == 1 {
			matches[i].TeamA, matches[i].TeamB = matches[i].TeamB, matches[i].TeamA
		}
	}
}

// shuffleCourtNumbers randomly assigns court numbers 1..N to the N matches,
// breaking the deterministic court-to-group mapping from the backtracking order.
func shuffleCourtNumbers(matches []domain.Match) {
	// Fisher-Yates shuffle on court numbers.
	n := len(matches)
	for i := n - 1; i > 0; i-- {
		j, _ := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		matches[i].Court, matches[j.Int64()].Court = matches[j.Int64()].Court, matches[i].Court
	}
}

const idAlphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

func shortID() string {
	b := make([]byte, 6)
	for i := range b {
		idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(idAlphabet))))
		b[i] = idAlphabet[idx.Int64()]
	}
	return string(b)
}
