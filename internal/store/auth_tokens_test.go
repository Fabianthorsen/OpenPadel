package store_test

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	"github.com/fabianthorsen/openpadel/internal/store"
)

func sha256Hex(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// readTokenExpiry returns the stored expires_at (RFC3339) for a token's hash.
func readTokenExpiry(t *testing.T, s *store.Store, rawToken string) string {
	t.Helper()
	var expiresAt string
	err := s.DB().QueryRow(`SELECT expires_at FROM auth_tokens WHERE token_hash = ?`, sha256Hex(rawToken)).Scan(&expiresAt)
	if err != nil {
		t.Fatalf("read expires_at: %v", err)
	}
	return expiresAt
}

func TestAuthToken_ValidLookup(t *testing.T) {
	s := newTestStore(t)
	u, _ := s.CreateUser("alice@example.com", "Alice", "password123", 3)

	token, err := s.CreateAuthToken(u.ID)
	if err != nil {
		t.Fatalf("CreateAuthToken: %v", err)
	}

	got, err := s.GetUserByToken(token)
	if err != nil {
		t.Fatalf("GetUserByToken: %v", err)
	}
	if got.ID != u.ID {
		t.Errorf("expected user %s, got %s", u.ID, got.ID)
	}
}

func TestAuthToken_UnknownToken(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.GetUserByToken("not-a-real-token"); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

// The raw bearer token must never be stored — only its hash — so a DB leak
// yields nothing usable (#240).
func TestAuthToken_StoredAsHashOnly(t *testing.T) {
	s := newTestStore(t)
	u, _ := s.CreateUser("alice@example.com", "Alice", "password123", 3)
	token, _ := s.CreateAuthToken(u.ID)

	var rawCount int
	if err := s.DB().QueryRow(`SELECT COUNT(*) FROM auth_tokens WHERE token_hash = ?`, token).Scan(&rawCount); err != nil {
		t.Fatalf("count raw: %v", err)
	}
	if rawCount != 0 {
		t.Error("raw token found in auth_tokens; token is not hashed at rest")
	}

	var hashCount int
	if err := s.DB().QueryRow(`SELECT COUNT(*) FROM auth_tokens WHERE token_hash = ?`, sha256Hex(token)).Scan(&hashCount); err != nil {
		t.Fatalf("count hash: %v", err)
	}
	if hashCount != 1 {
		t.Errorf("expected token stored as its SHA-256 hash, found %d rows", hashCount)
	}
}

// An expired token is rejected and purged so it cannot be reused (#240).
func TestAuthToken_Expired(t *testing.T) {
	s := newTestStore(t)
	u, _ := s.CreateUser("alice@example.com", "Alice", "password123", 3)
	raw := "expired-token-raw"
	past := time.Now().UTC().Add(-time.Hour).Format(time.RFC3339)
	_, err := s.DB().Exec(
		`INSERT INTO auth_tokens (token_hash, user_id, created_at, expires_at) VALUES (?, ?, ?, ?)`,
		sha256Hex(raw), u.ID, past, past,
	)
	if err != nil {
		t.Fatalf("seed expired token: %v", err)
	}

	if _, err := s.GetUserByToken(raw); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound for expired token, got %v", err)
	}

	var count int
	if err := s.DB().QueryRow(`SELECT COUNT(*) FROM auth_tokens WHERE token_hash = ?`, sha256Hex(raw)).Scan(&count); err != nil {
		t.Fatalf("count after expiry: %v", err)
	}
	if count != 0 {
		t.Error("expired token was not purged on lookup")
	}
}

// A token nearing the end of its window is extended on use (sliding expiry).
func TestAuthToken_SlidingExpiryExtendsOnUse(t *testing.T) {
	s := newTestStore(t)
	u, _ := s.CreateUser("alice@example.com", "Alice", "password123", 3)
	token, _ := s.CreateAuthToken(u.ID)

	// Age the token so only ~10 days remain, well past the refresh threshold.
	soon := time.Now().UTC().Add(10 * 24 * time.Hour).Format(time.RFC3339)
	if _, err := s.DB().Exec(`UPDATE auth_tokens SET expires_at = ? WHERE token_hash = ?`, soon, sha256Hex(token)); err != nil {
		t.Fatalf("age token: %v", err)
	}

	if _, err := s.GetUserByToken(token); err != nil {
		t.Fatalf("GetUserByToken: %v", err)
	}

	got, _ := time.Parse(time.RFC3339, readTokenExpiry(t, s, token))
	if remaining := time.Until(got); remaining < 29*24*time.Hour {
		t.Errorf("expected expiry extended to ~30d, only %s remaining", remaining)
	}
}

// A freshly-used token is NOT rewritten on every request — the refresh is
// throttled so busy clients don't cause a write per call.
func TestAuthToken_SlidingExpiryThrottled(t *testing.T) {
	s := newTestStore(t)
	u, _ := s.CreateUser("alice@example.com", "Alice", "password123", 3)
	token, _ := s.CreateAuthToken(u.ID)

	before := readTokenExpiry(t, s, token)
	if _, err := s.GetUserByToken(token); err != nil {
		t.Fatalf("GetUserByToken: %v", err)
	}
	after := readTokenExpiry(t, s, token)

	if before != after {
		t.Errorf("expiry was rewritten within the refresh interval: %s -> %s", before, after)
	}
}
