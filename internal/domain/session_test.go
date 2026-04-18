package domain_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/fabianthorsen/openpadel/internal/domain"
)

func TestSession_TimedAmericanoFields(t *testing.T) {
	duration := 120
	buffer := 120
	roundDuration := 960

	sess := &domain.Session{
		ID:                     "sess123",
		Status:                 domain.StatusLobby,
		GameMode:               "timed_americano",
		TotalDurationMinutes:   &duration,
		BufferSeconds:          &buffer,
		RoundDurationSeconds:   &roundDuration,
		RoundStartedAt:         nil,
	}

	if sess.TotalDurationMinutes == nil {
		t.Errorf("expected TotalDurationMinutes to be set")
	}
	if *sess.TotalDurationMinutes != 120 {
		t.Errorf("expected TotalDurationMinutes=120, got %d", *sess.TotalDurationMinutes)
	}

	if sess.BufferSeconds == nil {
		t.Errorf("expected BufferSeconds to be set")
	}
	if *sess.BufferSeconds != 120 {
		t.Errorf("expected BufferSeconds=120, got %d", *sess.BufferSeconds)
	}

	if sess.RoundDurationSeconds == nil {
		t.Errorf("expected RoundDurationSeconds to be set")
	}
	if *sess.RoundDurationSeconds != 960 {
		t.Errorf("expected RoundDurationSeconds=960, got %d", *sess.RoundDurationSeconds)
	}

	if sess.RoundStartedAt != nil {
		t.Errorf("expected RoundStartedAt to be nil initially")
	}
}

func TestSession_TimedAmericanoFieldsOptional(t *testing.T) {
	sess := &domain.Session{
		ID:                     "sess456",
		Status:                 domain.StatusLobby,
		GameMode:               "americano",
		TotalDurationMinutes:   nil,
		BufferSeconds:          nil,
		RoundDurationSeconds:   nil,
		RoundStartedAt:         nil,
	}

	if sess.TotalDurationMinutes != nil {
		t.Errorf("expected TotalDurationMinutes to be nil")
	}
	if sess.BufferSeconds != nil {
		t.Errorf("expected BufferSeconds to be nil")
	}
	if sess.RoundDurationSeconds != nil {
		t.Errorf("expected RoundDurationSeconds to be nil")
	}
	if sess.RoundStartedAt != nil {
		t.Errorf("expected RoundStartedAt to be nil")
	}
}

func TestSession_TimedAmericanoJSON_Marshal(t *testing.T) {
	now := time.Now().UTC()
	duration := 120
	buffer := 120
	roundDuration := 960

	sess := &domain.Session{
		ID:                     "sess789",
		Status:                 domain.StatusActive,
		GameMode:               "timed_americano",
		TotalDurationMinutes:   &duration,
		BufferSeconds:          &buffer,
		RoundDurationSeconds:   &roundDuration,
		RoundStartedAt:         &now,
		Players:                []domain.Player{},
		CreatedAt:              time.Now().UTC(),
		UpdatedAt:              time.Now().UTC(),
	}

	data, err := json.Marshal(sess)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if v, ok := result["total_duration_minutes"]; !ok || v != float64(120) {
		t.Errorf("expected total_duration_minutes in JSON")
	}
	if v, ok := result["buffer_seconds"]; !ok || v != float64(120) {
		t.Errorf("expected buffer_seconds in JSON")
	}
	if v, ok := result["round_duration_seconds"]; !ok || v != float64(960) {
		t.Errorf("expected round_duration_seconds in JSON")
	}
	if _, ok := result["round_started_at"]; !ok {
		t.Errorf("expected round_started_at in JSON")
	}
}

func TestSession_TimedAmericanoJSON_MarshalWithNilFields(t *testing.T) {
	sess := &domain.Session{
		ID:                     "sess999",
		Status:                 domain.StatusLobby,
		GameMode:               "americano",
		TotalDurationMinutes:   nil,
		BufferSeconds:          nil,
		RoundDurationSeconds:   nil,
		RoundStartedAt:         nil,
		Players:                []domain.Player{},
		CreatedAt:              time.Now().UTC(),
		UpdatedAt:              time.Now().UTC(),
	}

	data, err := json.Marshal(sess)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if v, ok := result["total_duration_minutes"]; ok && v != nil {
		t.Errorf("expected total_duration_minutes to be omitted or null")
	}
}
