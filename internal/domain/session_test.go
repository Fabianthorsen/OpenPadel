package domain

import (
	"fmt"
	"testing"
)

func TestGameMode_IsValid(t *testing.T) {
	tests := []struct {
		mode  GameMode
		valid bool
	}{
		{ModeAmericano, true},
		{ModeMexicano, true},
		{GameMode("invalid"), false},
		{GameMode(""), false},
	}

	for _, tt := range tests {
		if got := tt.mode.IsValid(); got != tt.valid {
			t.Errorf("GameMode(%q).IsValid() = %v, want %v", string(tt.mode), got, tt.valid)
		}
	}
}

func TestGameMode_Values(t *testing.T) {
	vals := GameMode("").Values()
	if len(vals) != 2 {
		t.Errorf("GameMode.Values() returned %d values, want 2", len(vals))
	}
	if vals[0] != ModeAmericano || vals[1] != ModeMexicano {
		t.Errorf("GameMode.Values() = %v, want [%q %q]", vals, ModeAmericano, ModeMexicano)
	}
}

func TestSessionStatus_IsValid(t *testing.T) {
	tests := []struct {
		status SessionStatus
		valid  bool
	}{
		{StatusLobby, true},
		{StatusPlaying, true},
		{StatusDone, true},
		{SessionStatus("invalid"), false},
		{SessionStatus(""), false},
	}

	for _, tt := range tests {
		if got := tt.status.IsValid(); got != tt.valid {
			t.Errorf("SessionStatus(%q).IsValid() = %v, want %v", string(tt.status), got, tt.valid)
		}
	}
}

func TestInviteStatus_IsValid(t *testing.T) {
	tests := []struct {
		status InviteStatus
		valid  bool
	}{
		{InvitePending, true},
		{InviteAccepted, true},
		{InviteDeclined, true},
		{InviteStatus("invalid"), false},
		{InviteStatus(""), false},
	}

	for _, tt := range tests {
		if got := tt.status.IsValid(); got != tt.valid {
			t.Errorf("InviteStatus(%q).IsValid() = %v, want %v", string(tt.status), got, tt.valid)
		}
	}
}

func TestSessionInput_Validate_Valid(t *testing.T) {
	input := SessionInput{
		Courts:   2,
		Points:   24,
		Name:     "Test Session",
		GameMode: ModeAmericano,
	}
	errs := input.Validate()
	if len(errs) != 0 {
		t.Errorf("Valid input returned %d errors, want 0: %v", len(errs), errs)
	}
}

func TestSessionInput_Validate_InvalidGameMode(t *testing.T) {
	input := SessionInput{
		Courts:   2,
		Points:   24,
		GameMode: GameMode("invalid"),
	}
	errs := input.Validate()
	if len(errs) == 0 {
		t.Error("Invalid game mode should return error")
	}
	found := false
	for _, e := range errs {
		if e.Code == "invalid_game_mode" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected invalid_game_mode error, got %v", errs)
	}
}

func TestSessionInput_Validate_InvalidCourts(t *testing.T) {
	tests := []struct {
		name     string
		gameMode GameMode
		courts   int
	}{
		{"Americano below min", ModeAmericano, 0},
		{"Americano above max", ModeAmericano, 5},
		{"Mexicano below min", ModeMexicano, 1},
		{"Mexicano above max", ModeMexicano, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := SessionInput{
				Courts:   tt.courts,
				Points:   24,
				GameMode: tt.gameMode,
			}
			errs := input.Validate()
			found := false
			for _, e := range errs {
				if e.Code == "invalid_courts" {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected invalid_courts error, got %v", errs)
			}
		})
	}
}

func TestSessionInput_Validate_InvalidPoints(t *testing.T) {
	tests := []int{15, 17, 25, 31, 33}
	for _, points := range tests {
		t.Run(fmt.Sprintf("points=%d", points), func(t *testing.T) {
			input := SessionInput{
				Courts:   2,
				Points:   points,
				GameMode: ModeAmericano,
			}
			errs := input.Validate()
			found := false
			for _, e := range errs {
				if e.Code == "invalid_points" {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected invalid_points error for points=%d, got %v", points, errs)
			}
		})
	}
}

func TestSessionInput_Validate_InvalidRoundsTotal(t *testing.T) {
	tests := []struct {
		rounds *int
		valid  bool
	}{
		{nil, true},
		{intPtr(0), false},
		{intPtr(1), true},
		{intPtr(20), true},
		{intPtr(21), false},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("rounds=%v", tt.rounds), func(t *testing.T) {
			input := SessionInput{
				Courts:      2,
				Points:      24,
				GameMode:    ModeMexicano,
				RoundsTotal: tt.rounds,
			}
			errs := input.Validate()
			found := false
			for _, e := range errs {
				if e.Code == "invalid_rounds_total" {
					found = true
					break
				}
			}
			if found != !tt.valid {
				t.Errorf("Test %d: rounds=%v, expected found=%v, got %v", i, tt.rounds, !tt.valid, errs)
			}
		})
	}
}

func TestSessionInput_Validate_InvalidCourtDuration(t *testing.T) {
	tests := []struct {
		duration *int
		valid    bool
	}{
		{nil, true},
		{intPtr(14), false},
		{intPtr(15), true},
		{intPtr(300), true},
		{intPtr(301), false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("duration=%v", tt.duration), func(t *testing.T) {
			input := SessionInput{
				Courts:               2,
				Points:               24,
				GameMode:             ModeAmericano,
				CourtDurationMinutes: tt.duration,
			}
			errs := input.Validate()
			found := false
			for _, e := range errs {
				if e.Code == "invalid_court_duration" {
					found = true
					break
				}
			}
			if found != !tt.valid {
				t.Errorf("duration=%v, expected found=%v, got %v", tt.duration, !tt.valid, errs)
			}
		})
	}
}

func intPtr(v int) *int {
	return &v
}
