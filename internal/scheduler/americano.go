package scheduler

import (
	"crypto/rand"
	"math/big"
	mrand "math/rand"
	"sort"

	"github.com/fabianthorsen/nottennis/internal/domain"
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

// scorePermutation returns a penalty score for a given player ordering.
// Partner repeats are weighted heavily; matchup repeats less so.
func scorePermutation(perm []string, courts int, partnerCount map[[2]string]int, matchupCount map[[4]string]int) int {
	score := 0
	for c := 0; c < courts; c++ {
		base := c * 4
		a0, a1, b0, b1 := perm[base], perm[base+1], perm[base+2], perm[base+3]
		score += partnerCount[pairKey(a0, a1)] * 1000
		score += partnerCount[pairKey(b0, b1)] * 1000
		score += matchupCount[matchupKey(a0, a1, b0, b1)] * 10
	}
	return score
}

// optimizeTeamSplits tries all 3 pairings within each group of 4 and keeps the best.
// Given players W,X,Y,Z: W+X vs Y+Z | W+Y vs X+Z | W+Z vs X+Y
func optimizeTeamSplits(perm []string, courts int, partnerCount map[[2]string]int, matchupCount map[[4]string]int) []string {
	result := make([]string, len(perm))
	copy(result, perm)

	for c := 0; c < courts; c++ {
		base := c * 4
		w, x, y, z := perm[base], perm[base+1], perm[base+2], perm[base+3]

		splits := [][4]string{
			{w, x, y, z},
			{w, y, x, z},
			{w, z, x, y},
		}

		bestSplit := splits[0]
		bestScore := scorePermutation([]string{splits[0][0], splits[0][1], splits[0][2], splits[0][3]}, 1, partnerCount, matchupCount)

		for _, split := range splits[1:] {
			s := scorePermutation([]string{split[0], split[1], split[2], split[3]}, 1, partnerCount, matchupCount)
			if s < bestScore {
				bestSplit = split
				bestScore = s
			}
		}

		result[base] = bestSplit[0]
		result[base+1] = bestSplit[1]
		result[base+2] = bestSplit[2]
		result[base+3] = bestSplit[3]
	}

	return result
}

// assignCourts finds the best assignment of active players to courts by random
// search over the full partner+matchup history. For tournament sizes up to 16
// players, 2000 random shuffles reliably finds an optimal (zero-penalty) assignment
// in early rounds and a near-optimal one in later rounds.
func assignCourts(active []string, courts int, partnerCount map[[2]string]int, matchupCount map[[4]string]int) []domain.Match {
	candidate := make([]string, len(active))
	copy(candidate, active)

	best := make([]string, len(active))
	copy(best, candidate)
	bestScore := scorePermutation(candidate, courts, partnerCount, matchupCount)

	for i := 0; i < 2000 && bestScore > 0; i++ {
		mrand.Shuffle(len(candidate), func(a, b int) {
			candidate[a], candidate[b] = candidate[b], candidate[a]
		})
		copy(candidate, active)
		mrand.Shuffle(len(candidate), func(a, b int) {
			candidate[a], candidate[b] = candidate[b], candidate[a]
		})
		if s := scorePermutation(candidate, courts, partnerCount, matchupCount); s < bestScore {
			copy(best, candidate)
			bestScore = s
		}
	}

	// Final pass: try all 3 team-split options per court to minimise matchup repeats.
	best = optimizeTeamSplits(best, courts, partnerCount, matchupCount)

	matches := make([]domain.Match, courts)
	for c := 0; c < courts; c++ {
		base := c * 4
		matches[c] = domain.Match{
			ID:    shortID(),
			Court: c + 1,
			TeamA: [2]string{best[base], best[base+1]},
			TeamB: [2]string{best[base+2], best[base+3]},
		}
	}
	return matches
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
