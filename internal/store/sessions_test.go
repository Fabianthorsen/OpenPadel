package store_test

import (
	"testing"
	"time"

	"github.com/fabianthorsen/openpadel/internal/domain"
)

func TestCreateSession_WithTimedAmericanoParams(t *testing.T) {
	s := newTestStore(t)
	totalDurationMin := 120

	sess, err := s.CreateSession(1, 0, "Timed Americano Test", "timed_americano", nil, nil, nil, &totalDurationMin, "")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	loaded, err := s.GetSession(sess.ID)
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}

	if loaded.GameMode != "timed_americano" {
		t.Errorf("expected GameMode='timed_americano', got %q", loaded.GameMode)
	}

	if loaded.TotalDurationMinutes == nil || *loaded.TotalDurationMinutes != 120 {
		t.Errorf("expected TotalDurationMinutes=120, got %v", loaded.TotalDurationMinutes)
	}

	if loaded.Points != 0 {
		t.Errorf("expected Points=0 for timed_americano, got %d", loaded.Points)
	}
}

func TestCreateSession_CreatorUserID(t *testing.T) {
	s := newTestStore(t)
	alice := createUser(t, s, "alice@example.com", "Alice")

	sess, err := s.CreateSession(2, 24, "Test", "americano", nil, nil, nil, nil, alice)
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if sess.CreatorUserID != alice {
		t.Errorf("expected CreatorUserID=%s, got %q", alice, sess.CreatorUserID)
	}

	loaded, err := s.GetSession(sess.ID)
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}
	if loaded.CreatorUserID != alice {
		t.Errorf("expected persisted CreatorUserID=%s, got %q", alice, loaded.CreatorUserID)
	}
}

func TestCreateSession_NoCreatorUserID(t *testing.T) {
	s := newTestStore(t)

	sess, err := s.CreateSession(2, 24, "", "americano", nil, nil, nil, nil, "")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if sess.CreatorUserID != "" {
		t.Errorf("expected empty CreatorUserID, got %q", sess.CreatorUserID)
	}
}

func TestCompleteSession_EndedEarly(t *testing.T) {
	s := newTestStore(t)
	sess := createSession(t, s)

	if err := s.CompleteSession(sess, true); err != nil {
		t.Fatalf("CompleteSession(endedEarly=true): %v", err)
	}

	loaded, err := s.GetSession(sess)
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}

	if loaded.Status != domain.StatusComplete {
		t.Errorf("expected status 'complete', got %q", loaded.Status)
	}
}

func TestCompleteSession_NaturalCompletion(t *testing.T) {
	s := newTestStore(t)
	sess := createSession(t, s)

	if err := s.CompleteSession(sess, false); err != nil {
		t.Fatalf("CompleteSession(endedEarly=false): %v", err)
	}

	loaded, err := s.GetSession(sess)
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}

	if loaded.Status != domain.StatusComplete {
		t.Errorf("expected status 'complete', got %q", loaded.Status)
	}
}

func TestGetTournamentHistory_EndedEarly(t *testing.T) {
	s := newTestStore(t)
	alice := createUser(t, s, "alice@example.com", "Alice")
	sess := createSession(t, s)

	// Join as Alice
	_, err := s.CreatePlayer(sess, "Alice", alice)
	if err != nil {
		t.Fatalf("CreatePlayer: %v", err)
	}

	// Start and complete with ended_early = true
	if err := s.StartSession(sess, 1, nil); err != nil {
		t.Fatalf("StartSession: %v", err)
	}
	if err := s.CompleteSession(sess, true); err != nil {
		t.Fatalf("CompleteSession: %v", err)
	}

	// Get history
	history, err := s.GetTournamentHistory(alice)
	if err != nil {
		t.Fatalf("GetTournamentHistory: %v", err)
	}

	if len(history) != 1 {
		t.Fatalf("expected 1 tournament, got %d", len(history))
	}

	if !history[0].EndedEarly {
		t.Errorf("expected EndedEarly=true, got false")
	}
}

func TestGetTournamentHistory_NaturalCompletion(t *testing.T) {
	s := newTestStore(t)
	alice := createUser(t, s, "alice@example.com", "Alice")
	sess := createSession(t, s)

	// Join as Alice
	_, err := s.CreatePlayer(sess, "Alice", alice)
	if err != nil {
		t.Fatalf("CreatePlayer: %v", err)
	}

	// Start and complete with ended_early = false
	if err := s.StartSession(sess, 1, nil); err != nil {
		t.Fatalf("StartSession: %v", err)
	}
	if err := s.CompleteSession(sess, false); err != nil {
		t.Fatalf("CompleteSession: %v", err)
	}

	// Get history
	history, err := s.GetTournamentHistory(alice)
	if err != nil {
		t.Fatalf("GetTournamentHistory: %v", err)
	}

	if len(history) != 1 {
		t.Fatalf("expected 1 tournament, got %d", len(history))
	}

	if history[0].EndedEarly {
		t.Errorf("expected EndedEarly=false, got true")
	}
}

func TestStartTimedAmericanoSession(t *testing.T) {
	s := newTestStore(t)
	sess := createSession(t, s)

	duration := 120
	roundDuration := 960

	if err := s.StartTimedAmericanoSession(sess, "active", 10, &duration, &roundDuration, nil); err != nil {
		t.Fatalf("StartTimedAmericanoSession: %v", err)
	}

	loaded, err := s.GetSession(sess)
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}

	if loaded.Status != domain.StatusActive {
		t.Errorf("expected status 'active', got %q", loaded.Status)
	}

	if loaded.RoundsTotal == nil || *loaded.RoundsTotal != 10 {
		t.Errorf("expected RoundsTotal=10, got %v", loaded.RoundsTotal)
	}

	if loaded.TotalDurationMinutes == nil || *loaded.TotalDurationMinutes != 120 {
		t.Errorf("expected TotalDurationMinutes=120, got %v", loaded.TotalDurationMinutes)
	}

	if loaded.RoundDurationSeconds == nil || *loaded.RoundDurationSeconds != 960 {
		t.Errorf("expected RoundDurationSeconds=960, got %v", loaded.RoundDurationSeconds)
	}
}

func TestSetRoundStartedAt(t *testing.T) {
	s := newTestStore(t)
	sess := createSession(t, s)

	duration := 120
	roundDuration := 960

	if err := s.StartTimedAmericanoSession(sess, "active", 10, &duration, &roundDuration, nil); err != nil {
		t.Fatalf("StartTimedAmericanoSession: %v", err)
	}

	now := time.Now().UTC()
	if err := s.SetRoundStartedAt(sess, &now); err != nil {
		t.Fatalf("SetRoundStartedAt: %v", err)
	}

	loaded, err := s.GetSession(sess)
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}

	if loaded.RoundStartedAt == nil {
		t.Errorf("expected RoundStartedAt to be set")
	}
}

func TestUpdateRoundDuration(t *testing.T) {
	s := newTestStore(t)
	sess := createSession(t, s)

	duration := 120
	roundDuration := 960

	if err := s.StartTimedAmericanoSession(sess, "active", 10, &duration, &roundDuration, nil); err != nil {
		t.Fatalf("StartTimedAmericanoSession: %v", err)
	}

	newDuration := 600
	if err := s.UpdateRoundDuration(sess, &newDuration); err != nil {
		t.Fatalf("UpdateRoundDuration: %v", err)
	}

	loaded, err := s.GetSession(sess)
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}

	if loaded.RoundDurationSeconds == nil || *loaded.RoundDurationSeconds != 600 {
		t.Errorf("expected RoundDurationSeconds=600, got %v", loaded.RoundDurationSeconds)
	}
}
