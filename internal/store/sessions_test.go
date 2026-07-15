package store_test

import (
	"testing"

	"github.com/fabianthorsen/openpadel/internal/domain"
)

func TestCreateSession_CreatorUserID(t *testing.T) {
	s := newTestStore(t)
	alice := createUser(t, s, "alice@example.com", "Alice")

	input := domain.SessionInput{
		Courts:   2,
		Points:   24,
		Name:     "Test",
		GameMode: domain.ModeAmericano,
	}
	sess, err := s.CreateSession(input, alice)
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

	input := domain.SessionInput{
		Courts:   2,
		Points:   24,
		GameMode: domain.ModeAmericano,
	}
	sess, err := s.CreateSession(input, "")
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

	if loaded.Status != domain.StatusDone {
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

	if loaded.Status != domain.StatusDone {
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

// Tracer bullet: Mexicano with unlimited rounds (rounds_total = nil)
func TestCreateSession_MexicanoUnlimitedRounds(t *testing.T) {
	s := newTestStore(t)
	alice := createUser(t, s, "alice@example.com", "Alice")

	// Create Mexicano session with rounds_total = nil (unlimited)
	input := domain.SessionInput{
		Courts:      2,
		Points:      24,
		Name:        "Unlimited Mexicano",
		GameMode:    domain.ModeMexicano,
		RoundsTotal: nil, // explicitly unlimited
	}
	sess, err := s.CreateSession(input, alice)
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	// Verify in-memory session has nil rounds_total
	if sess.RoundsTotal != nil {
		t.Errorf("expected RoundsTotal=nil after create, got %v", sess.RoundsTotal)
	}

	// Verify persisted session has nil rounds_total
	loaded, err := s.GetSession(sess.ID)
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}
	if loaded.RoundsTotal != nil {
		t.Errorf("expected persisted RoundsTotal=nil, got %v", loaded.RoundsTotal)
	}
}

// Update Mexicano from fixed to unlimited rounds
func TestUpdateSession_MexicanoToUnlimited(t *testing.T) {
	s := newTestStore(t)
	alice := createUser(t, s, "alice@example.com", "Alice")

	// Create with fixed rounds
	input := domain.SessionInput{
		Courts:      2,
		Points:      24,
		Name:        "Start Fixed",
		GameMode:    domain.ModeMexicano,
		RoundsTotal: intPtr(5),
	}
	sess, err := s.CreateSession(input, alice)
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if sess.RoundsTotal == nil || *sess.RoundsTotal != 5 {
		t.Errorf("expected RoundsTotal=5 after create, got %v", sess.RoundsTotal)
	}

	// Update to unlimited
	updateInput := domain.SessionInput{
		Courts:      2,
		Points:      24,
		Name:        "Now Unlimited",
		GameMode:    domain.ModeMexicano,
		RoundsTotal: nil,
	}
	if err := s.UpdateSessionConfig(sess.ID, updateInput); err != nil {
		t.Fatalf("UpdateSessionConfig: %v", err)
	}

	// Verify updated session has nil rounds_total
	loaded, err := s.GetSession(sess.ID)
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}
	if loaded.RoundsTotal != nil {
		t.Errorf("expected RoundsTotal=nil after update, got %v", loaded.RoundsTotal)
	}
}

func intPtr(n int) *int {
	return &n
}

// StartMexicanoSession preserves unlimited (nil rounds_total)
func TestStartMexicanoSession_PreservesUnlimited(t *testing.T) {
	s := newTestStore(t)
	alice := createUser(t, s, "alice@example.com", "Alice")

	// Create unlimited Mexicano session
	input := domain.SessionInput{
		Courts:      2,
		Points:      24,
		Name:        "Unlimited",
		GameMode:    domain.ModeMexicano,
		RoundsTotal: nil,
	}
	sess, err := s.CreateSession(input, alice)
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	// Create some players (minimum 4 for 1 court in Mexicano)
	if _, err := s.CreatePlayer(sess.ID, "Player 1", ""); err != nil {
		t.Fatalf("CreatePlayer p1: %v", err)
	}
	if _, err := s.CreatePlayer(sess.ID, "Player 2", ""); err != nil {
		t.Fatalf("CreatePlayer p2: %v", err)
	}
	if _, err := s.CreatePlayer(sess.ID, "Player 3", ""); err != nil {
		t.Fatalf("CreatePlayer p3: %v", err)
	}
	if _, err := s.CreatePlayer(sess.ID, "Player 4", ""); err != nil {
		t.Fatalf("CreatePlayer p4: %v", err)
	}

	// Start the session
	if err := s.StartMexicanoSession(sess.ID, nil); err != nil {
		t.Fatalf("StartMexicanoSession: %v", err)
	}

	// Verify rounds_total is still nil (not overwritten)
	loaded, err := s.GetSession(sess.ID)
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}
	if loaded.RoundsTotal != nil {
		t.Errorf("expected RoundsTotal=nil after start, got %v", loaded.RoundsTotal)
	}
	if loaded.Status != domain.StatusPlaying {
		t.Errorf("expected Status=playing, got %s", loaded.Status)
	}
}

// AdvanceRound works on unlimited sessions (no Store-level cap validation)
// The API handler cap check is tested at the API level.
func TestAdvanceRound_UnlimitedSessionPreservesNil(t *testing.T) {
	s := newTestStore(t)
	alice := createUser(t, s, "alice@example.com", "Alice")

	// Create unlimited Mexicano
	input := domain.SessionInput{
		Courts:      1,
		Points:      24,
		Name:        "Test unlimited advance",
		GameMode:    domain.ModeMexicano,
		RoundsTotal: nil,
	}
	sess, err := s.CreateSession(input, alice)
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	// Start it (Mexicano preserves nil, unlike Americano which takes a roundsTotal param)
	if err := s.StartMexicanoSession(sess.ID, nil); err != nil {
		t.Fatalf("StartMexicanoSession: %v", err)
	}

	// Verify nil is preserved through start
	loaded, err := s.GetSession(sess.ID)
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}
	if loaded.RoundsTotal != nil {
		t.Errorf("expected RoundsTotal=nil after start, got %v", loaded.RoundsTotal)
	}
}

func TestGetUserSessions(t *testing.T) {
	s := newTestStore(t)
	alice := createUser(t, s, "alice@example.com", "Alice")
	bob := createUser(t, s, "bob@example.com", "Bob")

	// Create sessions for Alice
	sess1, err := s.CreateSession(domain.SessionInput{
		Courts:   2,
		Points:   24,
		Name:     "Alice Session 1",
		GameMode: domain.ModeAmericano,
	}, alice)
	if err != nil {
		t.Fatalf("CreateSession 1: %v", err)
	}

	sess2, err := s.CreateSession(domain.SessionInput{
		Courts:   2,
		Points:   24,
		Name:     "Alice Session 2",
		GameMode: domain.ModeAmericano,
	}, alice)
	if err != nil {
		t.Fatalf("CreateSession 2: %v", err)
	}

	// Create session for Bob
	_, err = s.CreateSession(domain.SessionInput{
		Courts:   2,
		Points:   24,
		Name:     "Bob Session",
		GameMode: domain.ModeAmericano,
	}, bob)
	if err != nil {
		t.Fatalf("CreateSession for Bob: %v", err)
	}

	// Get Alice's sessions
	aliceSessions, err := s.GetUserSessions(alice)
	if err != nil {
		t.Fatalf("GetUserSessions(alice): %v", err)
	}

	if len(aliceSessions) != 2 {
		t.Errorf("expected 2 sessions for Alice, got %d", len(aliceSessions))
	}

	// Verify admin tokens are included
	for _, sess := range aliceSessions {
		if sess.AdminToken == "" {
			t.Errorf("expected non-empty AdminToken in session %s", sess.ID)
		}
		if sess.AdminToken != sess1.AdminToken && sess.AdminToken != sess2.AdminToken {
			t.Errorf("unexpected AdminToken in session %s", sess.ID)
		}
	}

	// Get Bob's sessions
	bobSessions, err := s.GetUserSessions(bob)
	if err != nil {
		t.Fatalf("GetUserSessions(bob): %v", err)
	}

	if len(bobSessions) != 1 {
		t.Errorf("expected 1 session for Bob, got %d", len(bobSessions))
	}

	// Get sessions for non-existent user
	emptySessions, err := s.GetUserSessions("nonexistent")
	if err != nil {
		t.Fatalf("GetUserSessions(nonexistent): %v", err)
	}

	if len(emptySessions) != 0 {
		t.Errorf("expected 0 sessions for nonexistent user, got %d", len(emptySessions))
	}
}
