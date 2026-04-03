package scheduler

import (
	"crypto/rand"
	"math/big"
	"sort"

	"github.com/fabianthorsen/nottennis/internal/domain"
)

// Generate produces all rounds for an Americano session upfront.
// Hard constraint: a player benched in round N must play in round N+1.
// Secondary: minimise partner repeats across consecutive rounds.
func Generate(players []domain.Player, courts, totalRounds int) []domain.Round {
	ids := make([]string, len(players))
	for i, p := range players {
		ids[i] = p.ID
	}

	benchSize := len(ids) - courts*4

	lastBenchedRound := make(map[string]int) // 0 = never benched
	benchTotal := make(map[string]int)
	lastPartner := make(map[string]string)

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
			active = ids
		}

		for _, id := range bench {
			lastBenchedRound[id] = roundNum
			benchTotal[id]++
		}

		matches := assignCourts(active, courts, lastPartner, r)

		for _, m := range matches {
			lastPartner[m.TeamA[0]] = m.TeamA[1]
			lastPartner[m.TeamA[1]] = m.TeamA[0]
			lastPartner[m.TeamB[0]] = m.TeamB[1]
			lastPartner[m.TeamB[1]] = m.TeamB[0]
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

// assignCourts distributes active players across courts, trying to avoid
// recent partner pairings by scoring candidate rotations and picking the best.
func assignCourts(active []string, courts int, lastPartner map[string]string, roundIdx int) []domain.Match {
	n := len(active)
	best := rotate(active, roundIdx%n)
	bestPenalty := partnerPenalty(best, courts, lastPartner)

	for offset := 1; offset < n && offset < 8; offset++ {
		candidate := rotate(active, (roundIdx+offset)%n)
		if p := partnerPenalty(candidate, courts, lastPartner); p < bestPenalty {
			best = candidate
			bestPenalty = p
		}
	}

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

func rotate(s []string, by int) []string {
	n := len(s)
	out := make([]string, n)
	for i, v := range s {
		out[(i+by)%n] = v
	}
	return out
}

func partnerPenalty(order []string, courts int, lastPartner map[string]string) int {
	penalty := 0
	for c := 0; c < courts; c++ {
		base := c * 4
		if lastPartner[order[base]] == order[base+1] {
			penalty++
		}
		if lastPartner[order[base+2]] == order[base+3] {
			penalty++
		}
	}
	return penalty
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
