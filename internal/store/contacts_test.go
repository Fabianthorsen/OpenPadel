package store_test

import (
	"os"
	"testing"

	"github.com/fabianthorsen/openpadel/internal/store"
)

func newTestStore(t *testing.T) *store.Store {
	t.Helper()
	f, err := os.CreateTemp("", "openpadel-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	t.Cleanup(func() { os.Remove(f.Name()) })

	s, err := store.Open(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func createUser(t *testing.T, s *store.Store, email, displayName string) string {
	t.Helper()
	u, err := s.CreateUser(email, displayName, "password123")
	if err != nil {
		t.Fatalf("CreateUser(%q): %v", email, err)
	}
	return u.ID
}

func TestAddContact(t *testing.T) {
	s := newTestStore(t)
	alice := createUser(t, s, "alice@example.com", "Alice")
	bob := createUser(t, s, "bob@example.com", "Bob")

	if err := s.AddContact(alice, bob); err != nil {
		t.Fatalf("AddContact: %v", err)
	}

	contacts, err := s.GetContacts(alice)
	if err != nil {
		t.Fatalf("GetContacts: %v", err)
	}
	if len(contacts) != 1 || contacts[0].UserID != bob {
		t.Errorf("expected 1 contact (bob), got %v", contacts)
	}
}

func TestAddContact_Duplicate(t *testing.T) {
	s := newTestStore(t)
	alice := createUser(t, s, "alice@example.com", "Alice")
	bob := createUser(t, s, "bob@example.com", "Bob")

	s.AddContact(alice, bob)
	err := s.AddContact(alice, bob)
	if err != store.ErrAlreadyContact {
		t.Errorf("expected ErrAlreadyContact, got %v", err)
	}
}

func TestAddContact_Self(t *testing.T) {
	s := newTestStore(t)
	alice := createUser(t, s, "alice@example.com", "Alice")

	err := s.AddContact(alice, alice)
	if err == nil {
		t.Error("expected error when adding self as contact")
	}
}

func TestAddContact_UnknownUser(t *testing.T) {
	s := newTestStore(t)
	alice := createUser(t, s, "alice@example.com", "Alice")

	err := s.AddContact(alice, "u_nonexistent")
	if err != store.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestRemoveContact(t *testing.T) {
	s := newTestStore(t)
	alice := createUser(t, s, "alice@example.com", "Alice")
	bob := createUser(t, s, "bob@example.com", "Bob")

	s.AddContact(alice, bob)
	if err := s.RemoveContact(alice, bob); err != nil {
		t.Fatalf("RemoveContact: %v", err)
	}

	contacts, _ := s.GetContacts(alice)
	if len(contacts) != 0 {
		t.Errorf("expected 0 contacts after removal, got %d", len(contacts))
	}
}

func TestRemoveContact_NotFound(t *testing.T) {
	s := newTestStore(t)
	alice := createUser(t, s, "alice@example.com", "Alice")
	bob := createUser(t, s, "bob@example.com", "Bob")

	err := s.RemoveContact(alice, bob)
	if err != store.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestGetContacts_Empty(t *testing.T) {
	s := newTestStore(t)
	alice := createUser(t, s, "alice@example.com", "Alice")

	contacts, err := s.GetContacts(alice)
	if err != nil {
		t.Fatalf("GetContacts: %v", err)
	}
	if len(contacts) != 0 {
		t.Errorf("expected empty contacts, got %d", len(contacts))
	}
}

func TestGetContacts_NotShared(t *testing.T) {
	s := newTestStore(t)
	alice := createUser(t, s, "alice@example.com", "Alice")
	bob := createUser(t, s, "bob@example.com", "Bob")

	// Alice adds Bob but Bob has not added Alice.
	s.AddContact(alice, bob)

	bobContacts, _ := s.GetContacts(bob)
	if len(bobContacts) != 0 {
		t.Errorf("contacts are not symmetric: bob should have 0, got %d", len(bobContacts))
	}
}

func TestSearchUsers(t *testing.T) {
	s := newTestStore(t)
	alice := createUser(t, s, "alice@example.com", "Alice")
	createUser(t, s, "bob@example.com", "Bob")
	createUser(t, s, "charlie@example.com", "Charlie")

	results, err := s.SearchUsers(alice, "ob")
	if err != nil {
		t.Fatalf("SearchUsers: %v", err)
	}
	if len(results) != 1 || results[0].DisplayName != "Bob" {
		t.Errorf("expected [Bob], got %v", results)
	}
}

func TestSearchUsers_ExcludesSelf(t *testing.T) {
	s := newTestStore(t)
	alice := createUser(t, s, "alice@example.com", "Alice")

	results, err := s.SearchUsers(alice, "Alice")
	if err != nil {
		t.Fatalf("SearchUsers: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected self to be excluded, got %v", results)
	}
}

func TestSearchUsers_MarksContacts(t *testing.T) {
	s := newTestStore(t)
	alice := createUser(t, s, "alice@example.com", "Alice")
	bob := createUser(t, s, "bob@example.com", "Bob")

	s.AddContact(alice, bob)

	results, err := s.SearchUsers(alice, "Bob")
	if err != nil {
		t.Fatalf("SearchUsers: %v", err)
	}
	if len(results) != 1 || !results[0].IsContact {
		t.Errorf("expected Bob to be marked as contact, got %v", results)
	}
}

func TestIsContact(t *testing.T) {
	s := newTestStore(t)
	alice := createUser(t, s, "alice@example.com", "Alice")
	bob := createUser(t, s, "bob@example.com", "Bob")

	ok, _ := s.IsContact(alice, bob)
	if ok {
		t.Error("expected false before adding contact")
	}

	s.AddContact(alice, bob)

	ok, _ = s.IsContact(alice, bob)
	if !ok {
		t.Error("expected true after adding contact")
	}
}
