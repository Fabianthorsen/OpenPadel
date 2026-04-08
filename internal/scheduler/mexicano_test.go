package scheduler

import (
	"testing"

	"github.com/fabianthorsen/openpadel/internal/domain"
)

// makeStandings builds a []Standing slice from a list of player IDs, assigned
// ranks in the order given (first ID = rank 1).
func makeStandings(ids ...string) []domain.Standing {
	out := make([]domain.Standing, len(ids))
	for i, id := range ids {
		out[i] = domain.Standing{
			Rank:     i + 1,
			PlayerID: id,
			Points:   (len(ids) - i) * 10, // descending points for clarity
		}
	}
	return out
}

func playerSet(r domain.Round) map[string]bool {
	set := make(map[string]bool)
	for _, m := range r.Matches {
		set[m.TeamA[0]] = true
		set[m.TeamA[1]] = true
		set[m.TeamB[0]] = true
		set[m.TeamB[1]] = true
	}
	return set
}

// TestMexicanoRound1_NoBench verifies round 1 with exactly courts*4 players.
func TestMexicanoRound1_NoBench(t *testing.T) {
	standings := makeStandings("p1", "p2", "p3", "p4", "p5", "p6", "p7", "p8")
	courts := 2

	r := GenerateMexicanoRound(standings, courts, nil, nil, 1)

	if len(r.Matches) != courts {
		t.Fatalf("expected %d matches, got %d", courts, len(r.Matches))
	}
	if len(r.Bench) != 0 {
		t.Fatalf("expected 0 bench, got %v", r.Bench)
	}

	// All 8 players should appear exactly once.
	playing := playerSet(r)
	for _, s := range standings {
		if !playing[s.PlayerID] {
			t.Errorf("player %s not in any match", s.PlayerID)
		}
	}
}

// TestMexicanoRound1_Pairing verifies the specific pairing rule.
// Slots: [0+2 vs 1+3] on court 1, [4+6 vs 5+7] on court 2.
func TestMexicanoRound1_Pairing(t *testing.T) {
	ids := []string{"p1", "p2", "p3", "p4", "p5", "p6", "p7", "p8"}
	standings := makeStandings(ids...)
	courts := 2

	r := GenerateMexicanoRound(standings, courts, nil, nil, 1)

	c1 := r.Matches[0]
	if c1.TeamA[0] != "p1" || c1.TeamA[1] != "p3" {
		t.Errorf("court 1 team A: want [p1 p3], got %v", c1.TeamA)
	}
	if c1.TeamB[0] != "p2" || c1.TeamB[1] != "p4" {
		t.Errorf("court 1 team B: want [p2 p4], got %v", c1.TeamB)
	}

	c2 := r.Matches[1]
	if c2.TeamA[0] != "p5" || c2.TeamA[1] != "p7" {
		t.Errorf("court 2 team A: want [p5 p7], got %v", c2.TeamA)
	}
	if c2.TeamB[0] != "p6" || c2.TeamB[1] != "p8" {
		t.Errorf("court 2 team B: want [p6 p8], got %v", c2.TeamB)
	}
}

// TestMexicanoRound_Bench verifies that bottom-ranked players bench.
func TestMexicanoRound_Bench(t *testing.T) {
	// 9 players, 2 courts → 1 on bench (bottom rank = p9)
	ids := []string{"p1", "p2", "p3", "p4", "p5", "p6", "p7", "p8", "p9"}
	standings := makeStandings(ids...)

	r := GenerateMexicanoRound(standings, 2, nil, nil, 2)

	if len(r.Bench) != 1 {
		t.Fatalf("expected 1 bench player, got %v", r.Bench)
	}
	if r.Bench[0] != "p9" {
		t.Errorf("expected p9 (lowest rank) to bench, got %s", r.Bench[0])
	}

	playing := playerSet(r)
	if playing["p9"] {
		t.Error("p9 should not be playing")
	}
}

// TestMexicanoRound_NoBenchTwiceInARow verifies that a player who benched last
// round is forced to play, and the next eligible low-ranked player benches instead.
func TestMexicanoRound_NoBenchTwiceInARow(t *testing.T) {
	// 9 players, 2 courts → 1 bench slot
	ids := []string{"p1", "p2", "p3", "p4", "p5", "p6", "p7", "p8", "p9"}
	standings := makeStandings(ids...)

	// p9 benched last round (round 2), so in round 3 they must play.
	lastBenched := map[string]int{"p9": 2}

	r := GenerateMexicanoRound(standings, 2, nil, lastBenched, 3)

	if len(r.Bench) != 1 {
		t.Fatalf("expected 1 bench player, got %v", r.Bench)
	}
	if r.Bench[0] == "p9" {
		t.Error("p9 benched last round and must play this round")
	}
	// Next eligible low-ranked player is p8
	if r.Bench[0] != "p8" {
		t.Errorf("expected p8 to bench (next lowest eligible), got %s", r.Bench[0])
	}
}

// TestMexicanoRound_MultipleBench verifies bench with 2 bench slots.
func TestMexicanoRound_MultipleBench(t *testing.T) {
	// 10 players, 2 courts → 2 bench slots
	ids := []string{"p1", "p2", "p3", "p4", "p5", "p6", "p7", "p8", "p9", "p10"}
	standings := makeStandings(ids...)

	r := GenerateMexicanoRound(standings, 2, nil, nil, 2)

	if len(r.Bench) != 2 {
		t.Fatalf("expected 2 bench players, got %v", r.Bench)
	}

	benchSet := map[string]bool{}
	for _, pid := range r.Bench {
		benchSet[pid] = true
	}
	// Bottom 2 (p9, p10) should bench
	if !benchSet["p9"] || !benchSet["p10"] {
		t.Errorf("expected p9 and p10 to bench, got %v", r.Bench)
	}
}

// TestMexicanoRound_CourtNumbers verifies courts are numbered 1..N.
func TestMexicanoRound_CourtNumbers(t *testing.T) {
	standings := makeStandings("p1", "p2", "p3", "p4", "p5", "p6", "p7", "p8",
		"p9", "p10", "p11", "p12")
	r := GenerateMexicanoRound(standings, 3, nil, nil, 1)

	for i, m := range r.Matches {
		if m.Court != i+1 {
			t.Errorf("match %d: expected court %d, got %d", i, i+1, m.Court)
		}
	}
}

// TestMexicanoRound_RoundNumberSet checks that the returned round has the correct number.
func TestMexicanoRound_RoundNumberSet(t *testing.T) {
	standings := makeStandings("p1", "p2", "p3", "p4", "p5", "p6", "p7", "p8")
	r := GenerateMexicanoRound(standings, 2, nil, nil, 5)
	if r.Number != 5 {
		t.Errorf("expected round number 5, got %d", r.Number)
	}
}
