package timed

import (
	"fmt"

	"github.com/fabianthorsen/openpadel/internal/domain"
	"github.com/fabianthorsen/openpadel/internal/gamemode/americano"
)

type TimedAmericanoConfig struct {
	Players           []domain.Player
	Courts            int
	TotalDurationMin  int
	BufferSeconds     int
	PlayerCount       int
}

// CalculateTimedRounds computes the number of rounds and per-round duration for a timed americano tournament.
// R = P-1 (even) or P (odd) where P is player count
// T = (D*60 - R*B) / R, where D is duration in minutes, B is buffer in seconds
// Returns error if calculated round duration is less than 120 seconds.
func CalculateTimedRounds(playerCount, totalDurationMin, bufferSeconds int) (rounds, roundDurationSec int, err error) {
	if playerCount%2 == 0 {
		rounds = playerCount - 1
	} else {
		rounds = playerCount
	}

	totalSeconds := totalDurationMin * 60
	roundDurationSec = (totalSeconds - rounds*bufferSeconds) / rounds

	if roundDurationSec < 120 {
		return 0, 0, fmt.Errorf("insufficient time: round duration would be %d seconds (minimum 120)", roundDurationSec)
	}

	return rounds, roundDurationSec, nil
}

// RecalculateRoundDuration recalculates the duration for each remaining round if tournament falls behind.
// Returns max((remainingSeconds - remainingRounds*bufferSeconds) / remainingRounds, 60)
func RecalculateRoundDuration(remainingRounds, remainingSeconds, bufferSeconds int) int {
	newDuration := (remainingSeconds - remainingRounds*bufferSeconds) / remainingRounds
	if newDuration < 60 {
		newDuration = 60
	}
	return newDuration
}

// GenerateTimedAmericano generates all rounds for a timed americano session.
// It reuses the existing Americano scheduler to ensure fair rotation constraints.
func GenerateTimedAmericano(players []domain.Player, courts, totalRounds int) ([]domain.Round, error) {
	rounds := americano.Generate(players, courts, totalRounds)
	return rounds, nil
}
