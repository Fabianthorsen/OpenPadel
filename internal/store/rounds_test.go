package store_test

import (
	"testing"

	"github.com/fabianthorsen/openpadel/internal/domain"
)

func TestSaveRounds_GetRounds(t *testing.T) {
	s := newTestStore(t)
	sess := createSession(t, s)

	p1, _ := s.CreatePlayer(sess, "Alice", "")
	p2, _ := s.CreatePlayer(sess, "Bob", "")
	p3, _ := s.CreatePlayer(sess, "Charlie", "")
	p4, _ := s.CreatePlayer(sess, "Diana", "")

	round := domain.Round{
		ID:     "r001",
		Number: 1,
		Bench:  []string{},
		Matches: []domain.Match{
			{ID: "m001", Court: 1, TeamA: [2]string{p1.ID, p2.ID}, TeamB: [2]string{p3.ID, p4.ID}},
		},
	}
	if err := s.SaveRounds(sess, []domain.Round{round}); err != nil {
		t.Fatalf("SaveRounds: %v", err)
	}

	rounds, err := s.GetRounds(sess)
	if err != nil {
		t.Fatalf("GetRounds: %v", err)
	}
	if len(rounds) != 1 {
		t.Fatalf("expected 1 round, got %d", len(rounds))
	}
	if len(rounds[0].Matches) != 1 {
		t.Errorf("expected 1 match, got %d", len(rounds[0].Matches))
	}
}

func TestGetCurrentRound(t *testing.T) {
	s := newTestStore(t)
	sess := createSession(t, s)

	p1, _ := s.CreatePlayer(sess, "Alice", "")
	p2, _ := s.CreatePlayer(sess, "Bob", "")
	p3, _ := s.CreatePlayer(sess, "Charlie", "")
	p4, _ := s.CreatePlayer(sess, "Diana", "")

	round := domain.Round{
		ID:     "r001",
		Number: 1,
		Bench:  []string{},
		Matches: []domain.Match{
			{ID: "m001", Court: 1, TeamA: [2]string{p1.ID, p2.ID}, TeamB: [2]string{p3.ID, p4.ID}},
		},
	}
	s.SaveRounds(sess, []domain.Round{round})
	s.StartSession(sess, 1, nil)

	current, err := s.GetCurrentRound(sess)
	if err != nil {
		t.Fatalf("GetCurrentRound: %v", err)
	}
	if current.Number != 1 {
		t.Errorf("expected round number 1, got %d", current.Number)
	}
	if len(current.Matches) != 1 {
		t.Errorf("expected 1 match, got %d", len(current.Matches))
	}
}

func TestUpdateScore(t *testing.T) {
	s := newTestStore(t)
	sess := createSession(t, s)

	p1, _ := s.CreatePlayer(sess, "Alice", "")
	p2, _ := s.CreatePlayer(sess, "Bob", "")
	p3, _ := s.CreatePlayer(sess, "Charlie", "")
	p4, _ := s.CreatePlayer(sess, "Diana", "")

	round := domain.Round{
		ID:     "r001",
		Number: 1,
		Bench:  []string{},
		Matches: []domain.Match{
			{ID: "m001", Court: 1, TeamA: [2]string{p1.ID, p2.ID}, TeamB: [2]string{p3.ID, p4.ID}},
		},
	}
	s.SaveRounds(sess, []domain.Round{round})
	s.StartSession(sess, 1, nil)

	match, err := s.UpdateScore("m001", 16, 8)
	if err != nil {
		t.Fatalf("UpdateScore: %v", err)
	}
	if match.Score == nil {
		t.Fatal("expected score to be set")
	}
	if match.Score.A != 16 || match.Score.B != 8 {
		t.Errorf("expected score 16-8, got %d-%d", match.Score.A, match.Score.B)
	}
}

func TestGetLeaderboard(t *testing.T) {
	s := newTestStore(t)
	sess := createSession(t, s)

	p1, _ := s.CreatePlayer(sess, "Alice", "")
	p2, _ := s.CreatePlayer(sess, "Bob", "")
	p3, _ := s.CreatePlayer(sess, "Charlie", "")
	p4, _ := s.CreatePlayer(sess, "Diana", "")

	round := domain.Round{
		ID:     "r001",
		Number: 1,
		Bench:  []string{},
		Matches: []domain.Match{
			{ID: "m001", Court: 1, TeamA: [2]string{p1.ID, p2.ID}, TeamB: [2]string{p3.ID, p4.ID}},
		},
	}
	s.SaveRounds(sess, []domain.Round{round})
	s.StartSession(sess, 1, nil)
	s.UpdateScore("m001", 16, 8)

	standings, err := s.GetLeaderboard(sess)
	if err != nil {
		t.Fatalf("GetLeaderboard: %v", err)
	}
	if len(standings) != 4 {
		t.Fatalf("expected 4 standings, got %d", len(standings))
	}
	if standings[0].Points < standings[len(standings)-1].Points {
		t.Error("standings should be sorted descending by points")
	}
	for i, st := range standings {
		if st.Rank != i+1 {
			t.Errorf("standing[%d] has rank %d, expected %d", i, st.Rank, i+1)
		}
	}
}

func TestAllRoundsComplete(t *testing.T) {
	s := newTestStore(t)
	sess := createSession(t, s)

	p1, _ := s.CreatePlayer(sess, "Alice", "")
	p2, _ := s.CreatePlayer(sess, "Bob", "")
	p3, _ := s.CreatePlayer(sess, "Charlie", "")
	p4, _ := s.CreatePlayer(sess, "Diana", "")

	round := domain.Round{
		ID:     "r001",
		Number: 1,
		Bench:  []string{},
		Matches: []domain.Match{
			{ID: "m001", Court: 1, TeamA: [2]string{p1.ID, p2.ID}, TeamB: [2]string{p3.ID, p4.ID}},
		},
	}
	s.SaveRounds(sess, []domain.Round{round})
	s.StartSession(sess, 1, nil)

	done, err := s.AllRoundsComplete(sess)
	if err != nil {
		t.Fatalf("AllRoundsComplete: %v", err)
	}
	if done {
		t.Error("expected rounds to be incomplete before scoring")
	}

	s.UpdateScore("m001", 16, 8)
	done, err = s.AllRoundsComplete(sess)
	if err != nil {
		t.Fatalf("AllRoundsComplete after score: %v", err)
	}
	if !done {
		t.Error("expected all rounds complete after scoring the only match")
	}
}
