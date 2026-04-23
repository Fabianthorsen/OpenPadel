package mexicano

import (
	"crypto/rand"
	"math/big"

	"github.com/fabianthorsen/openpadel/internal/domain"
)

// GenerateMexicanoRound produces the next Mexicano round from the current standings.
//
// Mexicano requires exactly courts*4 players — no bench. Every player plays every round.
//
// Pairing rule: rank 1+4 vs 2+3, rank 5+8 vs 6+7, etc.
// This balances each match by ranking (strongest paired with weakest, 2nd with 3rd).
//
// For round 1 pass randomly-shuffled standings so the first round is fair.
func GenerateMexicanoRound(standings []domain.Standing, courts int, roundNum int) domain.Round {
	// Pair: [0+3 vs 1+2], [4+7 vs 5+6], …
	matches := make([]domain.Match, courts)
	for c := 0; c < courts; c++ {
		base := c * 4
		matches[c] = domain.Match{
			ID:    shortID(),
			Court: c + 1,
			TeamA: [2]string{standings[base].PlayerID, standings[base+3].PlayerID},
			TeamB: [2]string{standings[base+1].PlayerID, standings[base+2].PlayerID},
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

const idAlphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

func shortID() string {
	b := make([]byte, 6)
	for i := range b {
		idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(idAlphabet))))
		b[i] = idAlphabet[idx.Int64()]
	}
	return string(b)
}

func shuffleTeamSides(matches []domain.Match) {
	for i := range matches {
		n, _ := rand.Int(rand.Reader, big.NewInt(2))
		if n.Int64() == 1 {
			matches[i].TeamA, matches[i].TeamB = matches[i].TeamB, matches[i].TeamA
		}
	}
}
