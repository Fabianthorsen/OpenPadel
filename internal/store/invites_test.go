package store_test

import (
	"testing"

	"github.com/fabianthorsen/openpadel/internal/store"
)

func createSession(t *testing.T, s *store.Store) string {
	t.Helper()
	sess, err := s.CreateSession(2, 24, "", "americano", 2, 6, nil, nil, nil, "")
	if err != nil {
		t.Fatalf("createSession: %v", err)
	}
	return sess.ID
}

func TestCreateInvite(t *testing.T) {
	s := newTestStore(t)
	alice := createUser(t, s, "alice@example.com", "Alice")
	bob := createUser(t, s, "bob@example.com", "Bob")
	sess := createSession(t, s)

	inv, err := s.CreateInvite(sess, alice, bob)
	if err != nil {
		t.Fatalf("CreateInvite: %v", err)
	}
	if inv.ToUserID != bob || inv.FromUserID != alice || inv.Status != "pending" {
		t.Errorf("unexpected invite: %+v", inv)
	}
}

func TestCreateInvite_Duplicate(t *testing.T) {
	s := newTestStore(t)
	alice := createUser(t, s, "alice@example.com", "Alice")
	bob := createUser(t, s, "bob@example.com", "Bob")
	sess := createSession(t, s)

	s.CreateInvite(sess, alice, bob)
	_, err := s.CreateInvite(sess, alice, bob)
	if err != store.ErrAlreadyInvited {
		t.Errorf("expected ErrAlreadyInvited, got %v", err)
	}
}

func TestGetPendingInvites(t *testing.T) {
	s := newTestStore(t)
	alice := createUser(t, s, "alice@example.com", "Alice")
	bob := createUser(t, s, "bob@example.com", "Bob")
	sess := createSession(t, s)

	s.CreateInvite(sess, alice, bob)

	invites, err := s.GetPendingInvites(bob)
	if err != nil {
		t.Fatalf("GetPendingInvites: %v", err)
	}
	if len(invites) != 1 || invites[0].FromUserID != alice {
		t.Errorf("expected 1 pending invite from alice, got %v", invites)
	}
}

func TestGetPendingInvites_Empty(t *testing.T) {
	s := newTestStore(t)
	bob := createUser(t, s, "bob@example.com", "Bob")

	invites, err := s.GetPendingInvites(bob)
	if err != nil {
		t.Fatalf("GetPendingInvites: %v", err)
	}
	if len(invites) != 0 {
		t.Errorf("expected empty, got %d", len(invites))
	}
}

func TestAcceptInvite(t *testing.T) {
	s := newTestStore(t)
	alice := createUser(t, s, "alice@example.com", "Alice")
	bob := createUser(t, s, "bob@example.com", "Bob")
	sess := createSession(t, s)

	inv, _ := s.CreateInvite(sess, alice, bob)

	player, err := s.AcceptInvite(inv.ID, bob)
	if err != nil {
		t.Fatalf("AcceptInvite: %v", err)
	}
	if player.UserID != bob {
		t.Errorf("expected player user_id=%s, got %s", bob, player.UserID)
	}

	pending, _ := s.GetPendingInvites(bob)
	if len(pending) != 0 {
		t.Errorf("expected 0 pending after accept, got %d", len(pending))
	}
}

func TestAcceptInvite_WrongUser(t *testing.T) {
	s := newTestStore(t)
	alice := createUser(t, s, "alice@example.com", "Alice")
	bob := createUser(t, s, "bob@example.com", "Bob")
	charlie := createUser(t, s, "charlie@example.com", "Charlie")
	sess := createSession(t, s)

	inv, _ := s.CreateInvite(sess, alice, bob)

	_, err := s.AcceptInvite(inv.ID, charlie)
	if err != store.ErrNotFound {
		t.Errorf("expected ErrNotFound when wrong user accepts, got %v", err)
	}
}

func TestDeclineInvite(t *testing.T) {
	s := newTestStore(t)
	alice := createUser(t, s, "alice@example.com", "Alice")
	bob := createUser(t, s, "bob@example.com", "Bob")
	sess := createSession(t, s)

	inv, _ := s.CreateInvite(sess, alice, bob)

	if err := s.DeclineInvite(inv.ID, bob); err != nil {
		t.Fatalf("DeclineInvite: %v", err)
	}

	pending, _ := s.GetPendingInvites(bob)
	if len(pending) != 0 {
		t.Errorf("expected 0 pending after decline, got %d", len(pending))
	}
}

func TestDeclineInvite_WrongUser(t *testing.T) {
	s := newTestStore(t)
	alice := createUser(t, s, "alice@example.com", "Alice")
	bob := createUser(t, s, "bob@example.com", "Bob")
	charlie := createUser(t, s, "charlie@example.com", "Charlie")
	sess := createSession(t, s)

	inv, _ := s.CreateInvite(sess, alice, bob)

	err := s.DeclineInvite(inv.ID, charlie)
	if err != store.ErrNotFound {
		t.Errorf("expected ErrNotFound when wrong user declines, got %v", err)
	}
}
