package store_test

import (
	"errors"
	"testing"

	"github.com/fabianthorsen/openpadel/internal/store"
)

func TestCreateUser(t *testing.T) {
	s := newTestStore(t)
	u, err := s.CreateUser("alice@example.com", "Alice", "password123")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if u.ID == "" || u.Email != "alice@example.com" || u.DisplayName != "Alice" {
		t.Errorf("unexpected user: %+v", u)
	}
}

func TestCreateUser_DuplicateEmail(t *testing.T) {
	s := newTestStore(t)
	s.CreateUser("alice@example.com", "Alice", "password123")
	_, err := s.CreateUser("alice@example.com", "Other", "password123")
	if !errors.Is(err, store.ErrEmailTaken) {
		t.Errorf("expected ErrEmailTaken, got %v", err)
	}
}

func TestAuthenticateUser(t *testing.T) {
	s := newTestStore(t)
	s.CreateUser("alice@example.com", "Alice", "password123")

	u, err := s.AuthenticateUser("alice@example.com", "password123")
	if err != nil {
		t.Fatalf("AuthenticateUser: %v", err)
	}
	if u.Email != "alice@example.com" {
		t.Errorf("expected email alice@example.com, got %q", u.Email)
	}
}

func TestAuthenticateUser_WrongPassword(t *testing.T) {
	s := newTestStore(t)
	s.CreateUser("alice@example.com", "Alice", "password123")

	_, err := s.AuthenticateUser("alice@example.com", "wrongpassword")
	if !errors.Is(err, store.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthenticateUser_UnknownEmail(t *testing.T) {
	s := newTestStore(t)

	_, err := s.AuthenticateUser("nobody@example.com", "password123")
	if !errors.Is(err, store.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestCreateAuthToken_GetUserByToken(t *testing.T) {
	s := newTestStore(t)
	u, _ := s.CreateUser("alice@example.com", "Alice", "password123")

	token, err := s.CreateAuthToken(u.ID)
	if err != nil {
		t.Fatalf("CreateAuthToken: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	found, err := s.GetUserByToken(token)
	if err != nil {
		t.Fatalf("GetUserByToken: %v", err)
	}
	if found.ID != u.ID {
		t.Errorf("expected user %s, got %s", u.ID, found.ID)
	}
}

func TestGetUserByToken_NotFound(t *testing.T) {
	s := newTestStore(t)

	_, err := s.GetUserByToken("nonexistent-token")
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestDeleteAuthToken(t *testing.T) {
	s := newTestStore(t)
	u, _ := s.CreateUser("alice@example.com", "Alice", "password123")
	token, _ := s.CreateAuthToken(u.ID)

	if err := s.DeleteAuthToken(token); err != nil {
		t.Fatalf("DeleteAuthToken: %v", err)
	}

	_, err := s.GetUserByToken(token)
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound after deletion, got %v", err)
	}
}

func TestUpdateProfile(t *testing.T) {
	s := newTestStore(t)
	u, _ := s.CreateUser("alice@example.com", "Alice", "password123")

	updated, err := s.UpdateProfile(u.ID, "Alice Updated", "Star", "blue")
	if err != nil {
		t.Fatalf("UpdateProfile: %v", err)
	}
	if updated.DisplayName != "Alice Updated" {
		t.Errorf("expected display name 'Alice Updated', got %q", updated.DisplayName)
	}
	if updated.AvatarIcon != "Star" {
		t.Errorf("expected avatar icon 'Star', got %q", updated.AvatarIcon)
	}
}

func TestPasswordResetFlow(t *testing.T) {
	s := newTestStore(t)
	s.CreateUser("alice@example.com", "Alice", "password123")

	rawToken, err := s.CreatePasswordResetToken("alice@example.com")
	if err != nil {
		t.Fatalf("CreatePasswordResetToken: %v", err)
	}
	if rawToken == "" {
		t.Fatal("expected non-empty reset token")
	}

	if err := s.RedeemPasswordResetToken(rawToken, "newpassword456"); err != nil {
		t.Fatalf("RedeemPasswordResetToken: %v", err)
	}

	_, err = s.AuthenticateUser("alice@example.com", "newpassword456")
	if err != nil {
		t.Fatalf("expected login with new password to succeed: %v", err)
	}

	_, err = s.AuthenticateUser("alice@example.com", "password123")
	if !errors.Is(err, store.ErrInvalidCredentials) {
		t.Errorf("expected old password to be rejected, got %v", err)
	}
}

func TestPasswordResetToken_Invalid(t *testing.T) {
	s := newTestStore(t)

	err := s.RedeemPasswordResetToken("bad-token", "newpassword456")
	if !errors.Is(err, store.ErrInvalidOrExpiredToken) {
		t.Errorf("expected ErrInvalidOrExpiredToken, got %v", err)
	}
}

func TestPasswordResetToken_UnknownEmail(t *testing.T) {
	s := newTestStore(t)

	_, err := s.CreatePasswordResetToken("nobody@example.com")
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestDeleteUser(t *testing.T) {
	s := newTestStore(t)
	u, _ := s.CreateUser("alice@example.com", "Alice", "password123")
	token, _ := s.CreateAuthToken(u.ID)

	if err := s.DeleteUser(u.ID); err != nil {
		t.Fatalf("DeleteUser: %v", err)
	}

	_, err := s.GetUserByToken(token)
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected token to be invalidated after DeleteUser, got %v", err)
	}
}
