package domain

import "testing"

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
