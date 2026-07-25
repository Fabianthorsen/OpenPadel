package store_test

import (
	"testing"

	"github.com/fabianthorsen/openpadel/internal/store"
)

// TestCreateSessionInvitesForClub_FansOutRoster confirms the fan-out creates one
// ordinary Session invite per Member, skipping the caller themselves.
func TestCreateSessionInvitesForClub_FansOutRoster(t *testing.T) {
	s := newTestStore(t)
	alice := createUser(t, s, "alice@example.com", "Alice")
	bob := createUser(t, s, "bob@example.com", "Bob")
	carol := createUser(t, s, "carol@example.com", "Carol")
	club := createClub(t, s, alice)
	if err := s.JoinClub(club, bob); err != nil {
		t.Fatalf("JoinClub bob: %v", err)
	}
	if err := s.JoinClub(club, carol); err != nil {
		t.Fatalf("JoinClub carol: %v", err)
	}
	sess := createSession(t, s)

	created, err := s.CreateSessionInvitesForClub(sess, alice, club)
	if err != nil {
		t.Fatalf("CreateSessionInvitesForClub: %v", err)
	}
	// Bob and Carol are invited; Alice (the caller) is not invited to her own fan-out.
	if len(created) != 2 {
		t.Fatalf("expected 2 invites, got %d", len(created))
	}
	got := map[string]bool{}
	for _, inv := range created {
		if inv.SessionID != sess || inv.FromUserID != alice || inv.Status != "pending" {
			t.Errorf("unexpected invite shape: %+v", inv)
		}
		got[inv.ToUserID] = true
	}
	if !got[bob] || !got[carol] || got[alice] {
		t.Errorf("unexpected invitees: %v", got)
	}

	// Each invitee holds exactly one pending Session invite, indistinguishable
	// from a targeted one — accepting adds a Player.
	for _, u := range []string{bob, carol} {
		pending, _ := s.GetPendingInvites(u)
		if len(pending) != 1 {
			t.Fatalf("expected 1 pending invite for %s, got %d", u, len(pending))
		}
		player, err := s.AcceptInvite(pending[0].ID, u)
		if err != nil {
			t.Fatalf("AcceptInvite(%s): %v", u, err)
		}
		if player.UserID != u || player.SessionID != sess {
			t.Errorf("accepted player mismatch: %+v", player)
		}
	}
}

// TestCreateSessionInvitesForClub_SkipsJoinedAndInvited confirms the fan-out is
// tolerant: Members already joined or already holding a pending invite are
// skipped, so a repeat fan-out neither duplicates nor errors.
func TestCreateSessionInvitesForClub_SkipsJoinedAndInvited(t *testing.T) {
	s := newTestStore(t)
	alice := createUser(t, s, "alice@example.com", "Alice")
	bob := createUser(t, s, "bob@example.com", "Bob")
	carol := createUser(t, s, "carol@example.com", "Carol")
	club := createClub(t, s, alice)
	if err := s.JoinClub(club, bob); err != nil {
		t.Fatalf("JoinClub bob: %v", err)
	}
	if err := s.JoinClub(club, carol); err != nil {
		t.Fatalf("JoinClub carol: %v", err)
	}
	sess := createSession(t, s)

	// Bob is already a Player in the Session; he must not be re-invited.
	if _, err := s.AddContactPlayer(sess, bob); err != nil {
		t.Fatalf("AddContactPlayer bob: %v", err)
	}

	created, err := s.CreateSessionInvitesForClub(sess, alice, club)
	if err != nil {
		t.Fatalf("first fan-out: %v", err)
	}
	if len(created) != 1 || created[0].ToUserID != carol {
		t.Fatalf("expected only Carol invited, got %+v", created)
	}

	// A repeat fan-out is a no-op: Carol already has a pending invite.
	again, err := s.CreateSessionInvitesForClub(sess, alice, club)
	if err != nil {
		t.Fatalf("repeat fan-out: %v", err)
	}
	if len(again) != 0 {
		t.Errorf("expected 0 invites on repeat, got %d", len(again))
	}
	pending, _ := s.GetPendingInvites(carol)
	if len(pending) != 1 {
		t.Errorf("expected Carol to hold exactly 1 pending invite, got %d", len(pending))
	}
}

// TestCreateSessionInvitesForClub_UnknownSession errors when the session is gone.
func TestCreateSessionInvitesForClub_UnknownSession(t *testing.T) {
	s := newTestStore(t)
	alice := createUser(t, s, "alice@example.com", "Alice")
	club := createClub(t, s, alice)

	if _, err := s.CreateSessionInvitesForClub("nope", alice, club); err != store.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
