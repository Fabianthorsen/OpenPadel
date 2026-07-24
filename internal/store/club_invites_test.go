package store_test

import (
	"testing"

	"github.com/fabianthorsen/openpadel/internal/store"
)

func createClub(t *testing.T, s *store.Store, creatorID string) string {
	t.Helper()
	club, err := s.CreateClub("Bouvet", "", "Star", "forest", creatorID)
	if err != nil {
		t.Fatalf("CreateClub: %v", err)
	}
	return club.ID
}

func TestCreateClubInvite(t *testing.T) {
	s := newTestStore(t)
	alice := createUser(t, s, "alice@example.com", "Alice")
	bob := createUser(t, s, "bob@example.com", "Bob")
	club := createClub(t, s, alice)

	inv, err := s.CreateClubInvite(club, alice, bob)
	if err != nil {
		t.Fatalf("CreateClubInvite: %v", err)
	}
	if inv.InviteeID != bob || inv.InviterID != alice || inv.Status != "pending" {
		t.Errorf("unexpected invite: %+v", inv)
	}
	if inv.ClubName != "Bouvet" {
		t.Errorf("expected club name, got %q", inv.ClubName)
	}
}

func TestCreateClubInvite_Duplicate(t *testing.T) {
	s := newTestStore(t)
	alice := createUser(t, s, "alice@example.com", "Alice")
	bob := createUser(t, s, "bob@example.com", "Bob")
	club := createClub(t, s, alice)

	if _, err := s.CreateClubInvite(club, alice, bob); err != nil {
		t.Fatalf("CreateClubInvite: %v", err)
	}
	_, err := s.CreateClubInvite(club, alice, bob)
	if err != store.ErrAlreadyClubInvited {
		t.Errorf("expected ErrAlreadyClubInvited, got %v", err)
	}
}

func TestCreateClubInvite_AlreadyMember(t *testing.T) {
	s := newTestStore(t)
	alice := createUser(t, s, "alice@example.com", "Alice")
	club := createClub(t, s, alice)

	// Alice is the creator/admin, so inviting her must be rejected.
	_, err := s.CreateClubInvite(club, alice, alice)
	if err != store.ErrAlreadyMember {
		t.Errorf("expected ErrAlreadyMember, got %v", err)
	}
}

func TestCreateClubInvite_ReviveDeclined(t *testing.T) {
	s := newTestStore(t)
	alice := createUser(t, s, "alice@example.com", "Alice")
	bob := createUser(t, s, "bob@example.com", "Bob")
	club := createClub(t, s, alice)

	inv, _ := s.CreateClubInvite(club, alice, bob)
	if err := s.DeclineClubInvite(inv.ID, bob); err != nil {
		t.Fatalf("DeclineClubInvite: %v", err)
	}

	// Re-inviting a declined User revives the invite to pending.
	revived, err := s.CreateClubInvite(club, alice, bob)
	if err != nil {
		t.Fatalf("re-invite after decline: %v", err)
	}
	if revived.Status != "pending" {
		t.Errorf("expected revived invite pending, got %q", revived.Status)
	}
	pending, _ := s.GetPendingClubInvites(bob)
	if len(pending) != 1 {
		t.Errorf("expected 1 pending after revive, got %d", len(pending))
	}
}

func TestGetPendingClubInvites(t *testing.T) {
	s := newTestStore(t)
	alice := createUser(t, s, "alice@example.com", "Alice")
	bob := createUser(t, s, "bob@example.com", "Bob")
	club := createClub(t, s, alice)

	if _, err := s.CreateClubInvite(club, alice, bob); err != nil {
		t.Fatalf("CreateClubInvite: %v", err)
	}

	invites, err := s.GetPendingClubInvites(bob)
	if err != nil {
		t.Fatalf("GetPendingClubInvites: %v", err)
	}
	if len(invites) != 1 || invites[0].InviterID != alice {
		t.Errorf("expected 1 pending invite from alice, got %v", invites)
	}
}

func TestGetPendingClubInvites_Empty(t *testing.T) {
	s := newTestStore(t)
	bob := createUser(t, s, "bob@example.com", "Bob")

	invites, err := s.GetPendingClubInvites(bob)
	if err != nil {
		t.Fatalf("GetPendingClubInvites: %v", err)
	}
	if len(invites) != 0 {
		t.Errorf("expected empty, got %d", len(invites))
	}
}

func TestAcceptClubInvite(t *testing.T) {
	s := newTestStore(t)
	alice := createUser(t, s, "alice@example.com", "Alice")
	bob := createUser(t, s, "bob@example.com", "Bob")
	club := createClub(t, s, alice)

	inv, _ := s.CreateClubInvite(club, alice, bob)

	if _, err := s.AcceptClubInvite(inv.ID, bob); err != nil {
		t.Fatalf("AcceptClubInvite: %v", err)
	}

	// Bob is now on the roster.
	member, err := s.GetClubMember(club, bob)
	if err != nil {
		t.Fatalf("GetClubMember after accept: %v", err)
	}
	if member.Role != "member" {
		t.Errorf("expected role member, got %q", member.Role)
	}

	pending, _ := s.GetPendingClubInvites(bob)
	if len(pending) != 0 {
		t.Errorf("expected 0 pending after accept, got %d", len(pending))
	}
}

func TestAcceptClubInvite_WrongUser(t *testing.T) {
	s := newTestStore(t)
	alice := createUser(t, s, "alice@example.com", "Alice")
	bob := createUser(t, s, "bob@example.com", "Bob")
	carol := createUser(t, s, "carol@example.com", "Carol")
	club := createClub(t, s, alice)

	inv, _ := s.CreateClubInvite(club, alice, bob)

	if _, err := s.AcceptClubInvite(inv.ID, carol); err != store.ErrNotFound {
		t.Errorf("expected ErrNotFound when wrong user accepts, got %v", err)
	}
}

func TestDeclineClubInvite(t *testing.T) {
	s := newTestStore(t)
	alice := createUser(t, s, "alice@example.com", "Alice")
	bob := createUser(t, s, "bob@example.com", "Bob")
	club := createClub(t, s, alice)

	inv, _ := s.CreateClubInvite(club, alice, bob)

	if err := s.DeclineClubInvite(inv.ID, bob); err != nil {
		t.Fatalf("DeclineClubInvite: %v", err)
	}

	// Declining must not add Bob to the roster.
	if _, err := s.GetClubMember(club, bob); err != store.ErrNotFound {
		t.Errorf("expected bob not a member after decline, got %v", err)
	}
	pending, _ := s.GetPendingClubInvites(bob)
	if len(pending) != 0 {
		t.Errorf("expected 0 pending after decline, got %d", len(pending))
	}
}

func TestDeclineClubInvite_WrongUser(t *testing.T) {
	s := newTestStore(t)
	alice := createUser(t, s, "alice@example.com", "Alice")
	bob := createUser(t, s, "bob@example.com", "Bob")
	carol := createUser(t, s, "carol@example.com", "Carol")
	club := createClub(t, s, alice)

	inv, _ := s.CreateClubInvite(club, alice, bob)

	if err := s.DeclineClubInvite(inv.ID, carol); err != store.ErrNotFound {
		t.Errorf("expected ErrNotFound when wrong user declines, got %v", err)
	}
}

func TestDeleteClub_CascadesInvites(t *testing.T) {
	s := newTestStore(t)
	alice := createUser(t, s, "alice@example.com", "Alice")
	bob := createUser(t, s, "bob@example.com", "Bob")
	club := createClub(t, s, alice)

	if _, err := s.CreateClubInvite(club, alice, bob); err != nil {
		t.Fatalf("CreateClubInvite: %v", err)
	}
	if err := s.DeleteClub(club); err != nil {
		t.Fatalf("DeleteClub: %v", err)
	}

	pending, err := s.GetPendingClubInvites(bob)
	if err != nil {
		t.Fatalf("GetPendingClubInvites: %v", err)
	}
	if len(pending) != 0 {
		t.Errorf("expected invites removed after club delete, got %d", len(pending))
	}
}
