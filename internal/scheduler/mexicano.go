package scheduler

import (
	"github.com/fabianthorsen/openpadel/internal/domain"
)

// GenerateMexicanoRound produces the next Mexicano round from the current standings.
//
// Mexicano requires exactly courts*4 players — no bench. Every player plays every round.
//
// Pairing rule: rank 1+3 vs 2+4, rank 5+7 vs 6+8, etc.
// This pits similarly-ranked players against each other so winners face winners.
//
// For round 1 pass randomly-shuffled standings so the first round is fair.
func GenerateMexicanoRound(standings []domain.Standing, courts int, roundNum int) domain.Round {
	// Pair: [0+2 vs 1+3], [4+6 vs 5+7], …
	matches := make([]domain.Match, courts)
	for c := 0; c < courts; c++ {
		base := c * 4
		matches[c] = domain.Match{
			ID:    shortID(),
			Court: c + 1,
			TeamA: [2]string{standings[base].PlayerID, standings[base+2].PlayerID},
			TeamB: [2]string{standings[base+1].PlayerID, standings[base+3].PlayerID},
		}
	}

	shuffleTeamSides(matches)

	return domain.Round{
		ID:      shortID(),
		Number:  roundNum,
		Bench:   []string{},
		Matches: matches,
	}
}
