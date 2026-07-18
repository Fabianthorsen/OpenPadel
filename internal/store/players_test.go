package store_test

import (
	"errors"
	"testing"

	"github.com/fabianthorsen/openpadel/internal/store"
)

func TestCreatePlayer(t *testing.T) {
	s := newTestStore(t)
	sess := createSession(t, s)

	p, err := s.CreatePlayer(sess, "Alice", "")
	if err != nil {
		t.Fatalf("CreatePlayer: %v", err)
	}
	if p.ID == "" || p.Name != "Alice" || !p.Active {
		t.Errorf("unexpected player: %+v", p)
	}
}

func TestCreatePlayer_RatingDefaultsToMedian(t *testing.T) {
	s := newTestStore(t)
	sess := createSession(t, s)

	p, err := s.CreatePlayer(sess, "Alice", "")
	if err != nil {
		t.Fatalf("CreatePlayer: %v", err)
	}
	if p.Rating != 3 {
		t.Errorf("expected default rating 3, got %d", p.Rating)
	}

	// Read back through the store to confirm it round-trips from the column.
	players, err := s.GetPlayers(sess)
	if err != nil {
		t.Fatalf("GetPlayers: %v", err)
	}
	if len(players) != 1 || players[0].Rating != 3 {
		t.Errorf("expected persisted rating 3, got %+v", players)
	}
}

func TestUpdatePlayerRating(t *testing.T) {
	s := newTestStore(t)
	sess := createSession(t, s)

	p, _ := s.CreatePlayer(sess, "Alice", "")
	if err := s.UpdatePlayerRating(p.ID, 5); err != nil {
		t.Fatalf("UpdatePlayerRating: %v", err)
	}

	players, _ := s.GetPlayers(sess)
	if len(players) != 1 || players[0].Rating != 5 {
		t.Errorf("expected rating 5 after update, got %+v", players)
	}
}

func TestCreatePlayer_WithUserID(t *testing.T) {
	s := newTestStore(t)
	alice := createUser(t, s, "alice@example.com", "Alice")
	sess := createSession(t, s)

	p, err := s.CreatePlayer(sess, "Alice", alice)
	if err != nil {
		t.Fatalf("CreatePlayer: %v", err)
	}
	if p.UserID != alice {
		t.Errorf("expected UserID=%s, got %q", alice, p.UserID)
	}
}

func TestCreatePlayer_SeedsRatingFromSelfRating(t *testing.T) {
	s := newTestStore(t)
	user, err := s.CreateUser("expert@example.com", "Expert", "password123", 5)
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	sess := createSession(t, s)

	p, err := s.CreatePlayer(sess, "Expert", user.ID)
	if err != nil {
		t.Fatalf("CreatePlayer: %v", err)
	}
	if p.Rating != 5 {
		t.Errorf("expected rating seeded from self_rating 5, got %d", p.Rating)
	}

	// Round-trips from the column.
	players, _ := s.GetPlayers(sess)
	if len(players) != 1 || players[0].Rating != 5 {
		t.Errorf("expected persisted rating 5, got %+v", players)
	}
}

func TestGetPlayers(t *testing.T) {
	s := newTestStore(t)
	sess := createSession(t, s)

	s.CreatePlayer(sess, "Alice", "")
	s.CreatePlayer(sess, "Bob", "")

	players, err := s.GetPlayers(sess)
	if err != nil {
		t.Fatalf("GetPlayers: %v", err)
	}
	if len(players) != 2 {
		t.Errorf("expected 2 players, got %d", len(players))
	}
}

func TestGetPlayers_Empty(t *testing.T) {
	s := newTestStore(t)
	sess := createSession(t, s)

	players, err := s.GetPlayers(sess)
	if err != nil {
		t.Fatalf("GetPlayers: %v", err)
	}
	if len(players) != 0 {
		t.Errorf("expected 0 players, got %d", len(players))
	}
}

func TestDeactivatePlayer(t *testing.T) {
	s := newTestStore(t)
	sess := createSession(t, s)

	p, _ := s.CreatePlayer(sess, "Alice", "")
	if err := s.DeactivatePlayer(p.ID); err != nil {
		t.Fatalf("DeactivatePlayer: %v", err)
	}

	players, _ := s.GetPlayers(sess)
	for _, pl := range players {
		if pl.ID == p.ID && pl.Active {
			t.Error("expected player to be inactive after DeactivatePlayer")
		}
	}
}

func TestDeactivatePlayer_NotFound(t *testing.T) {
	s := newTestStore(t)

	err := s.DeactivatePlayer("p_nonexistent")
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestGetCreatorName(t *testing.T) {
	s := newTestStore(t)
	sess := createSession(t, s)

	p, _ := s.CreatePlayer(sess, "Alice", "")
	s.SetCreatorPlayer(sess, p.ID)

	name := s.GetCreatorName(sess)
	if name != "Alice" {
		t.Errorf("expected creator name 'Alice', got %q", name)
	}
}

func TestGetCreatorName_NoCreator(t *testing.T) {
	s := newTestStore(t)
	sess := createSession(t, s)

	name := s.GetCreatorName(sess)
	if name != "" {
		t.Errorf("expected empty creator name, got %q", name)
	}
}
