package store_test

import (
	"testing"

	"github.com/fabianthorsen/openpadel/internal/domain"
)

func TestCreateSession_CreatorUserID(t *testing.T) {
	s := newTestStore(t)
	alice := createUser(t, s, "alice@example.com", "Alice")

	sess, err := s.CreateSession(2, 24, "Test", "americano", nil, nil, nil, alice)
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

	sess, err := s.CreateSession(2, 24, "", "americano", nil, nil, nil, "")
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

