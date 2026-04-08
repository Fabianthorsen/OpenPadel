package scheduler

import (
	"github.com/fabianthorsen/openpadel/internal/domain"
)

// GenerateMexicanoRound produces the next Mexicano round from the current standings.
//
// Pairing rules (per round ≥ 2):
//   - Sort players by rank (standings[0] = leader).
//   - Bottom benchSize players sit out; but no player may bench two consecutive rounds.
//     If a forced-out (must-play) player is in the bench zone, the next eligible player
//     from above takes their bench spot.
//   - Active players are paired by position: slots [0,2] vs [1,3], [4,6] vs [5,7], etc.
//     This pits similarly-ranked players against each other.
//
// For round 1 pass randomly-shuffled standings so the first round is fair.
func GenerateMexicanoRound(
	standings []domain.Standing, // ordered slice, index 0 = rank 1 (best)
	courts int,
	benchHistory map[string]int, // total bench turns per player (read from DB)
	lastBenchedRound map[string]int, // last round number the player benched (0 = never)
	roundNum int,
) domain.Round {
	n := len(standings)
	benchSize := n - courts*4

	var bench []string
	var active []string

	if benchSize <= 0 {
		for _, s := range standings {
			active = append(active, s.PlayerID)
		}
	} else {
		// Players who benched last round must play this round.
		mustPlay := make(map[string]bool, benchSize)
		for _, s := range standings {
			if lastBenchedRound[s.PlayerID] == roundNum-1 {
				mustPlay[s.PlayerID] = true
			}
		}

		// Walk from the bottom of standings, collecting bench players.
		// Skip anyone who must play; replace them with the next eligible player.
		bench = make([]string, 0, benchSize)
		for i := n - 1; i >= 0 && len(bench) < benchSize; i-- {
			pid := standings[i].PlayerID
			if !mustPlay[pid] {
				bench = append(bench, pid)
			}
		}

		benchSet := make(map[string]bool, len(bench))
		for _, pid := range bench {
			benchSet[pid] = true
		}

		// Active players keep standings order (rank 1 first).
		for _, s := range standings {
			if !benchSet[s.PlayerID] {
				active = append(active, s.PlayerID)
			}
		}
	}

	// Pair: [0+2 vs 1+3], [4+6 vs 5+7], …
	matches := make([]domain.Match, courts)
	for c := 0; c < courts; c++ {
		base := c * 4
		matches[c] = domain.Match{
			ID:    shortID(),
			Court: c + 1,
			TeamA: [2]string{active[base], active[base+2]},
			TeamB: [2]string{active[base+1], active[base+3]},
		}
	}

	shuffleTeamSides(matches)

	return domain.Round{
		ID:      shortID(),
		Number:  roundNum,
		Bench:   bench,
		Matches: matches,
	}
}
