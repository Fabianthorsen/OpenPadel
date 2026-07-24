package store_test

import (
	"database/sql"
	"testing"

	"github.com/fabianthorsen/openpadel/internal/domain"
	"github.com/fabianthorsen/openpadel/internal/store"
)

// TestDeleteClub_UnlinksSessions verifies the ON DELETE SET NULL on
// sessions.club_id: deleting a Club keeps its past Sessions, only orphaning them
// from the Club so they survive as ordinary Sessions (hard delete, no soft-delete).
func TestDeleteClub_UnlinksSessions(t *testing.T) {
	s := newTestStore(t)
	alice := createUser(t, s, "alice@example.com", "Alice")
	bob := createUser(t, s, "bob@example.com", "Bob")
	carol := createUser(t, s, "carol@example.com", "Carol")
	clubID := createClub(t, s, alice)

	// A second member and a pending invite so we can assert both cascade away.
	if err := s.JoinClub(clubID, bob); err != nil {
		t.Fatalf("JoinClub: %v", err)
	}
	if _, err := s.CreateClubInvite(clubID, alice, carol); err != nil {
		t.Fatalf("CreateClubInvite: %v", err)
	}

	sess, err := s.CreateSession(domain.SessionInput{Courts: 1, Points: 24, GameMode: domain.ModeAmericano}, alice)
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	// Link the Session to the Club directly — no API path attaches a Session to a
	// Club yet, but the column and its SET NULL semantics exist for future
	// club-owned games, and #126 asks that they survive a Club deletion.
	if _, err := s.DB().Exec("UPDATE sessions SET club_id = ? WHERE id = ?", clubID, sess.ID); err != nil {
		t.Fatalf("link session to club: %v", err)
	}

	if err := s.DeleteClub(clubID); err != nil {
		t.Fatalf("DeleteClub: %v", err)
	}

	// The Session still exists...
	if _, err := s.GetSession(sess.ID); err != nil {
		t.Fatalf("GetSession after club delete: %v", err)
	}
	// ...with its Club link cleared rather than cascade-deleted.
	var clubIDCol sql.NullString
	if err := s.DB().QueryRow("SELECT club_id FROM sessions WHERE id = ?", sess.ID).Scan(&clubIDCol); err != nil {
		t.Fatalf("read club_id: %v", err)
	}
	if clubIDCol.Valid {
		t.Errorf("expected club_id NULL after club delete, got %q", clubIDCol.String)
	}

	// Membership and invite rows are gone.
	if n := countRows(t, s, "club_members", clubID); n != 0 {
		t.Errorf("expected 0 club_members after delete, got %d", n)
	}
	if n := countRows(t, s, "club_invites", clubID); n != 0 {
		t.Errorf("expected 0 club_invites after delete, got %d", n)
	}
}

func countRows(t *testing.T, s *store.Store, table, clubID string) int {
	t.Helper()
	var n int
	// #nosec G202 -- table is a test-only constant, never user input.
	if err := s.DB().QueryRow("SELECT COUNT(*) FROM "+table+" WHERE club_id = ?", clubID).Scan(&n); err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	return n
}
