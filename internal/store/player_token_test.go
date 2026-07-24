package store_test

import (
	"testing"

	"github.com/fabianthorsen/openpadel/internal/domain"
	"github.com/fabianthorsen/openpadel/internal/store"
)

func newSession(t *testing.T, s *store.Store) *domain.Session {
	t.Helper()
	sess, err := s.CreateSession(domain.SessionInput{Courts: 2, Points: 24, GameMode: domain.ModeAmericano}, "")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	return sess
}

// CreatePlayer issues a per-player self-removal secret and stores only its hash;
// the raw secret comes back on the created Player, never the raw value in the DB
// (#241).
func TestCreatePlayer_IssuesToken(t *testing.T) {
	s := newTestStore(t)
	sess := newSession(t, s)

	p, err := s.CreatePlayer(sess.ID, "Alice", "", false)
	if err != nil {
		t.Fatalf("CreatePlayer: %v", err)
	}
	if p.Token == "" {
		t.Fatal("expected a non-empty player token")
	}

	var stored string
	if err := s.DB().QueryRow(`SELECT token_hash FROM players WHERE id = ?`, p.ID).Scan(&stored); err != nil {
		t.Fatalf("read token_hash: %v", err)
	}
	if stored == p.Token {
		t.Error("raw token stored in players.token_hash; expected a hash")
	}
	if stored == "" {
		t.Error("expected token_hash to be populated")
	}
}

func TestVerifyPlayerToken(t *testing.T) {
	s := newTestStore(t)
	sess := newSession(t, s)
	p, _ := s.CreatePlayer(sess.ID, "Alice", "", false)

	ok, err := s.VerifyPlayerToken(p.ID, p.Token)
	if err != nil {
		t.Fatalf("VerifyPlayerToken: %v", err)
	}
	if !ok {
		t.Error("expected the issued token to verify")
	}

	if ok, _ := s.VerifyPlayerToken(p.ID, "wrong"); ok {
		t.Error("wrong token should not verify")
	}
	if ok, _ := s.VerifyPlayerToken(p.ID, ""); ok {
		t.Error("empty token should not verify")
	}
	if ok, _ := s.VerifyPlayerToken("no-such-player", p.Token); ok {
		t.Error("unknown player should not verify")
	}
}

// The token must not surface in the session listing that every client reads.
func TestGetPlayers_OmitsToken(t *testing.T) {
	s := newTestStore(t)
	sess := newSession(t, s)
	_, _ = s.CreatePlayer(sess.ID, "Alice", "", false)

	got, err := s.GetSession(sess.ID)
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}
	for _, p := range got.Players {
		if p.Token != "" {
			t.Errorf("player token leaked in session listing for %s", p.Name)
		}
	}
}
