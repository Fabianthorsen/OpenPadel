package store_test

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/fabianthorsen/openpadel/internal/domain"
	"github.com/fabianthorsen/openpadel/internal/store"
)

// TestForeignKeysEnforced pins the #249 fix: the store connection must come up
// with FK enforcement on and WAL journaling, both of which the old mattn-style
// DSN silently failed to apply on the modernc.org/sqlite driver.
func TestForeignKeysEnforced(t *testing.T) {
	s := newTestStore(t)

	var fk int
	if err := s.DB().QueryRow("PRAGMA foreign_keys").Scan(&fk); err != nil {
		t.Fatalf("read foreign_keys: %v", err)
	}
	if fk != 1 {
		t.Errorf("expected foreign_keys=1, got %d", fk)
	}

	var journal string
	if err := s.DB().QueryRow("PRAGMA journal_mode").Scan(&journal); err != nil {
		t.Fatalf("read journal_mode: %v", err)
	}
	if journal != "wal" {
		t.Errorf("expected journal_mode=wal, got %q", journal)
	}

	// A row referencing a non-existent parent must be rejected.
	_, err := s.DB().Exec(`INSERT INTO players (id, session_id, name, joined_at) VALUES ('p1','ghost','n','t')`)
	if err == nil {
		t.Fatal("expected FK violation inserting player with unknown session_id, got nil")
	}
}

// TestDeleteUser_DetachesHistory verifies account deletion clears the creator
// label from Sessions (rather than deleting them) and removes reset tokens that
// used to orphan while FK enforcement was off.
func TestDeleteUser_DetachesHistory(t *testing.T) {
	s := newTestStore(t)
	alice := createUser(t, s, "alice@example.com", "Alice")

	sess, err := s.CreateSession(domain.SessionInput{Courts: 1, Points: 24, GameMode: domain.ModeAmericano}, alice)
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if _, err := s.CreatePasswordResetToken("alice@example.com"); err != nil {
		t.Fatalf("CreatePasswordResetToken: %v", err)
	}

	if err := s.DeleteUser(alice); err != nil {
		t.Fatalf("DeleteUser: %v", err)
	}

	// Session survives with its creator cleared.
	if _, err := s.GetSession(sess.ID); err != nil {
		t.Fatalf("GetSession after user delete: %v", err)
	}
	var creator sql.NullString
	if err := s.DB().QueryRow("SELECT creator_user_id FROM sessions WHERE id = ?", sess.ID).Scan(&creator); err != nil {
		t.Fatalf("read creator_user_id: %v", err)
	}
	if creator.Valid {
		t.Errorf("expected creator_user_id NULL, got %q", creator.String)
	}

	// Reset tokens are gone rather than orphaned.
	var n int
	if err := s.DB().QueryRow("SELECT COUNT(*) FROM password_reset_tokens WHERE user_id = ?", alice).Scan(&n); err != nil {
		t.Fatalf("count reset tokens: %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0 reset tokens after delete, got %d", n)
	}
}

// TestDeleteUser_DeletesSoleOwnedClub: when the departing User is a Club's only
// member, the Club is dissolved rather than left ownerless.
func TestDeleteUser_DeletesSoleOwnedClub(t *testing.T) {
	s := newTestStore(t)
	alice := createUser(t, s, "alice@example.com", "Alice")
	clubID := createClub(t, s, alice)

	if err := s.DeleteUser(alice); err != nil {
		t.Fatalf("DeleteUser: %v", err)
	}

	if _, err := s.GetClub(clubID); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected sole-owned club to be deleted, got %v", err)
	}
}

// TestDeleteUser_ReassignsClubOwnership: when other members remain, the Club
// survives — ownership and the Admin role pass to a remaining member.
func TestDeleteUser_ReassignsClubOwnership(t *testing.T) {
	s := newTestStore(t)
	alice := createUser(t, s, "alice@example.com", "Alice")
	bob := createUser(t, s, "bob@example.com", "Bob")
	carol := createUser(t, s, "carol@example.com", "Carol")
	clubID := createClub(t, s, alice)

	if err := s.JoinClub(clubID, bob); err != nil {
		t.Fatalf("JoinClub bob: %v", err)
	}
	// A pending invite alice sent must not block her deletion.
	if _, err := s.CreateClubInvite(clubID, alice, carol); err != nil {
		t.Fatalf("CreateClubInvite: %v", err)
	}

	if err := s.DeleteUser(alice); err != nil {
		t.Fatalf("DeleteUser: %v", err)
	}

	club, err := s.GetClub(clubID)
	if err != nil {
		t.Fatalf("GetClub after owner delete: %v", err)
	}
	if club.CreatedBy != bob {
		t.Errorf("expected created_by reassigned to bob, got %q", club.CreatedBy)
	}

	members, err := s.GetClubMembers(clubID)
	if err != nil {
		t.Fatalf("GetClubMembers: %v", err)
	}
	if len(members) != 1 {
		t.Fatalf("expected 1 remaining member, got %d", len(members))
	}
	if members[0].UserID != bob {
		t.Errorf("expected bob to remain, got %q", members[0].UserID)
	}
	if members[0].Role != "admin" {
		t.Errorf("expected bob promoted to admin, got %q", members[0].Role)
	}
}
