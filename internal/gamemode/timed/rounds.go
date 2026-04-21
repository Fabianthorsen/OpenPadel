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
// T = (D*60 - (R-1)*I*60 - R*B) / R, where D is duration in minutes, I is interval in minutes, B is buffer in seconds, R is rounds
// Returns error if calculated round duration is less than 120 seconds.
func CalculateTimedRounds(playerCount, totalDurationMin, bufferSeconds, intervalBetweenRoundsMin int) (rounds, roundDurationSec int, err error) {
	if playerCount%2 == 0 {
		rounds = playerCount - 1
	} else {
		rounds = playerCount
	}

	totalSeconds := totalDurationMin * 60
	intervalSeconds := intervalBetweenRoundsMin * 60
	// Total time consumed by intervals: (rounds - 1) * intervalSeconds
	// Total time consumed by buffers: rounds * bufferSeconds
	// Remaining time for actual play: totalSeconds - (rounds - 1) * intervalSeconds - rounds * bufferSeconds
	roundDurationSec = (totalSeconds - (rounds-1)*intervalSeconds - rounds*bufferSeconds) / rounds

	if roundDurationSec < 120 {
		return 0, 0, fmt.Errorf("insufficient time: round duration would be %d seconds (minimum 120)", roundDurationSec)
	}

	return rounds, roundDurationSec, nil
}

// RecalculateRoundDuration recalculates the duration for each remaining round if tournament falls behind.
// Returns max((remainingSeconds - (remainingRounds-1)*intervalSeconds - remainingRounds*bufferSeconds) / remainingRounds, 60)
func RecalculateRoundDuration(remainingRounds, remainingSeconds, bufferSeconds, intervalBetweenRoundsMin int) int {
	intervalSeconds := intervalBetweenRoundsMin * 60
	newDuration := (remainingSeconds - (remainingRounds-1)*intervalSeconds - remainingRounds*bufferSeconds) / remainingRounds
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
