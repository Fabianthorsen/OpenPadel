package timed_test

import (
	"testing"

	"github.com/fabianthorsen/openpadel/internal/domain"
	"github.com/fabianthorsen/openpadel/internal/gamemode/timed"
)

func TestCalculateTimedRounds_EvenPlayers_8Players(t *testing.T) {
	// 8 players, 120 min total
	// R = P - 1 = 7 (even)
	// T = (120 * 60) / 7 ≈ 1028 seconds
	rounds, duration, err := timed.CalculateTimedRounds(8, 120)
	if err != nil {
		t.Fatalf("CalculateTimedRounds: %v", err)
	}

	if rounds != 7 {
		t.Errorf("expected 7 rounds, got %d", rounds)
	}

	expectedDuration := (120 * 60) / 7
	if duration != expectedDuration {
		t.Errorf("expected duration %d seconds, got %d", expectedDuration, duration)
	}
	if duration < 120 {
		t.Errorf("duration should be at least 120 seconds")
	}
}

func TestCalculateTimedRounds_OddPlayers_9Players(t *testing.T) {
	// 9 players, 120 min total
	// R = P = 9 (odd)
	// T = (120 * 60) / 9 = 800 seconds
	rounds, duration, err := timed.CalculateTimedRounds(9, 120)
	if err != nil {
		t.Fatalf("CalculateTimedRounds: %v", err)
	}

	if rounds != 9 {
		t.Errorf("expected 9 rounds, got %d", rounds)
	}

	expectedDuration := (120 * 60) / 9
	if duration != expectedDuration {
		t.Errorf("expected duration %d seconds, got %d", expectedDuration, duration)
	}
	if duration < 120 {
		t.Errorf("duration should be at least 120 seconds")
	}
}

func TestCalculateTimedRounds_MinimumValidation(t *testing.T) {
	// Not enough time: would result in T < 120
	// 4 players, 10 min total
	// R = 3, T = 600 / 3 = 200 seconds (should now be valid since formula is simplified)
	_, duration, err := timed.CalculateTimedRounds(4, 10)
	if err != nil {
		t.Fatalf("CalculateTimedRounds: %v", err)
	}
	if duration < 120 {
		t.Errorf("expected duration >= 120 seconds, got %d", duration)
	}
}

func TestCalculateTimedRounds_LargeGroup(t *testing.T) {
	// 16 players, 180 min total
	// R = P - 1 = 15 (even)
	// T = (180 * 60) / 15 = 720 seconds
	rounds, duration, err := timed.CalculateTimedRounds(16, 180)
	if err != nil {
		t.Fatalf("CalculateTimedRounds: %v", err)
	}

	if rounds != 15 {
		t.Errorf("expected 15 rounds, got %d", rounds)
	}

	expectedDuration := (180 * 60) / 15
	if duration != expectedDuration {
		t.Errorf("expected duration %d seconds, got %d", expectedDuration, duration)
	}
}

func TestRecalculateRoundDuration_MidTournamentDrift(t *testing.T) {
	// 8 rounds remaining, 3600 seconds (60 min) remaining
	// T_new = 3600 / 8 = 450 seconds
	newDuration := timed.RecalculateRoundDuration(8, 3600)
	expectedDuration := 3600 / 8
	if newDuration != expectedDuration {
		t.Errorf("expected duration %d seconds, got %d", expectedDuration, newDuration)
	}
	if newDuration < 60 {
		t.Errorf("duration should enforce minimum of 60 seconds")
	}
}

func TestRecalculateRoundDuration_EnforcesMinimum(t *testing.T) {
	// Very tight: 4 rounds, 300 seconds (5 min) total
	// T_new = 300 / 4 = 75 seconds (> 60, so valid)
	newDuration := timed.RecalculateRoundDuration(4, 300)
	if newDuration < 60 {
		t.Errorf("expected duration >= 60 seconds (minimum), got %d", newDuration)
	}
}

func TestCalculateTimedRounds_DurationBoundaries(t *testing.T) {
	tests := []struct {
		name      string
		players   int
		duration  int
		wantError bool
	}{
		{
			name:      "standard duration valid",
			players:   8,
			duration:  120,
			wantError: false,
		},
		{
			name:      "short duration valid",
			players:   8,
			duration:  60,
			wantError: false,
		},
		{
			name:      "extremely short duration invalid",
			players:   20,
			duration:  15,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := timed.CalculateTimedRounds(tt.players, tt.duration)
			if (err != nil) != tt.wantError {
				t.Errorf("CalculateTimedRounds: got error %v, want error %v", err, tt.wantError)
			}
		})
	}
}

func TestGenerateTimedAmericano_SameRotationConstraints(t *testing.T) {
	players := makePlayers(8)

	rounds, err := timed.GenerateTimedAmericano(players, 2, 7)
	if err != nil {
		t.Fatalf("GenerateTimedAmericano: %v", err)
	}

	if len(rounds) != 7 {
		t.Errorf("expected 7 rounds, got %d", len(rounds))
	}

	for i, r := range rounds {
		if r.Number != i+1 {
			t.Errorf("round %d: expected number %d, got %d", i, i+1, r.Number)
		}

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
		if len(active) != len(players) {
			t.Errorf("round %d: expected all %d players accounted for, got %d", r.Number, len(players), len(active))
		}
	}
}

func makePlayers(n int) []domain.Player {
	players := make([]domain.Player, n)
	for i := range players {
		players[i] = domain.Player{
			ID: mustShortID(i + 1),
		}
	}
	return players
}

func mustShortID(n int) string {
	// Simple deterministic ID for testing
	const idAlphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	id := make([]byte, 3)
	val := n
	for i := len(id) - 1; i >= 0; i-- {
		id[i] = idAlphabet[val%len(idAlphabet)]
		val /= len(idAlphabet)
	}
	return string(id)
}
