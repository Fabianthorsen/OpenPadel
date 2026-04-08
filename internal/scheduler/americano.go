package scheduler

import (
	"crypto/rand"
	"math/big"
	"sort"

	"github.com/fabianthorsen/openpadel/internal/domain"
)

// Generate produces all rounds for an Americano session upfront.
// Hard constraint: a player benched in round N must play in round N+1.
// Core invariant: minimise partner and matchup repeats across the full tournament.
func Generate(players []domain.Player, courts, totalRounds int) []domain.Round {
	ids := make([]string, len(players))
	for i, p := range players {
		ids[i] = p.ID
	}

	benchSize := len(ids) - courts*4

	lastBenchedRound := make(map[string]int) // 0 = never benched
	benchTotal := make(map[string]int)

	// Full history — not just the previous round.
	partnerCount := map[[2]string]int{}
	matchupCount := map[[4]string]int{}

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
			bench = canBench[:benchSize]
			active = append(forced, canBench[benchSize:]...)
		} else {
			active = append([]string{}, ids...)
		}

		for _, id := range bench {
			lastBenchedRound[id] = roundNum
			benchTotal[id]++
		}

		matches := assignCourts(active, courts, partnerCount, matchupCount)

		for _, m := range matches {
			partnerCount[pairKey(m.TeamA[0], m.TeamA[1])]++
			partnerCount[pairKey(m.TeamB[0], m.TeamB[1])]++
			matchupCount[matchupKey(m.TeamA[0], m.TeamA[1], m.TeamB[0], m.TeamB[1])]++
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

// pairKey returns a canonical (sorted) key for a partnership.
func pairKey(a, b string) [2]string {
	if a > b {
		return [2]string{b, a}
	}
	return [2]string{a, b}
}

// matchupKey returns a canonical key for a court matchup.
// {A+B vs C+D} == {C+D vs A+B}, so both teams are sorted.
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

// bestCourtAssignment groups the given partner pairs into courts, minimising
// matchup repeats. Number of groupings: for 4 pairs→3, 6 pairs→15, 8 pairs→105.
func bestCourtAssignment(pairs [][2]string, matchupCount map[[4]string]int) []domain.Match {
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
				a0, a1 := pairs[first][0], pairs[first][1]
				b0, b1 := pairs[second][0], pairs[second][1]
				scratch[courtIdx] = domain.Match{
					ID:    shortID(),
					Court: courtIdx + 1,
					TeamA: [2]string{a0, a1},
					TeamB: [2]string{b0, b1},
				}
				bt(courtIdx+1, score+matchupCount[matchupKey(a0, a1, b0, b1)]*10)
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
// then minimise matchup repeats (secondary). Guaranteed optimal for up to 16 players.
func assignCourts(active []string, courts int, partnerCount map[[2]string]int, matchupCount map[[4]string]int) []domain.Match {
	pairs := bestPartnerMatching(active, partnerCount)
	return bestCourtAssignment(pairs, matchupCount)
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
	target := players - 1                       // minimum meaningful rounds
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

const idAlphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

func shortID() string {
	b := make([]byte, 6)
	for i := range b {
		idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(idAlphabet))))
		b[i] = idAlphabet[idx.Int64()]
	}
	return string(b)
}
