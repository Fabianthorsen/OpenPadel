package domain

import (
	"testing"
	"time"
)

// day is a helper: a fixed base date plus n days, used to make the
// first-qualifying-date tie-break explicit and deterministic.
func day(n int) time.Time {
	return time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC).AddDate(0, 0, n)
}

// game is a compact ClubGame builder for tests.
func game(userID, name string, team, opp, target, dayN int) ClubGame {
	return ClubGame{
		UserID:       userID,
		Name:         name,
		TeamPoints:   team,
		OppPoints:    opp,
		PointsTarget: target,
		PlayedAt:     day(dayN),
	}
}

// nGames returns count identical games for a user (a fast way to clear the
// MinGames gate) with a fixed margin.
func nGames(userID, name string, count, team, opp, target int) []ClubGame {
	out := make([]ClubGame, 0, count)
	for i := 0; i < count; i++ {
		out = append(out, game(userID, name, team, opp, target, i))
	}
	return out
}

// TestRankClubLeaderboard_MarginNormalization checks that the per-game margin is
// normalized by each game's own points target, so a +8 margin is worth more at a
// 16-point target than at 32. The three users each win every game by the same
// raw margin (+8) but at different targets; form must order them 16 > 24 > 32.
func TestRankClubLeaderboard_MarginNormalization(t *testing.T) {
	var games []ClubGame
	games = append(games, nGames("u16", "Sixteen", ClubLeaderboardMinGames, 12, 4, 16)...)    // +8/16 = 0.5
	games = append(games, nGames("u24", "TwentyFour", ClubLeaderboardMinGames, 16, 8, 24)...) // +8/24 = 0.333
	games = append(games, nGames("u32", "ThirtyTwo", ClubLeaderboardMinGames, 20, 12, 32)...) // +8/32 = 0.25

	board := RankClubLeaderboard(games)

	if len(board.Ranked) != 3 {
		t.Fatalf("expected 3 ranked, got %d", len(board.Ranked))
	}
	if got := board.Ranked[0].UserID; got != "u16" {
		t.Errorf("rank 1: expected u16, got %s", got)
	}
	if got := board.Ranked[1].UserID; got != "u24" {
		t.Errorf("rank 2: expected u24, got %s", got)
	}
	if got := board.Ranked[2].UserID; got != "u32" {
		t.Errorf("rank 3: expected u32, got %s", got)
	}
	if got := board.Ranked[0].Form; !approx(got, 0.5) {
		t.Errorf("u16 form: expected 0.5, got %v", got)
	}
	// Ranks are 1-indexed and dense.
	for i, e := range board.Ranked {
		if e.Rank != i+1 {
			t.Errorf("entry %d has rank %d, want %d", i, e.Rank, i+1)
		}
	}
}

// TestRankClubLeaderboard_MinGamesGate verifies the ranked/provisional split at
// and just below the threshold: exactly MinGames games ranks; MinGames-1 is
// provisional with gamesToGo == 1.
func TestRankClubLeaderboard_MinGamesGate(t *testing.T) {
	var games []ClubGame
	games = append(games, nGames("ranked", "Ranked", ClubLeaderboardMinGames, 12, 4, 16)...)
	games = append(games, nGames("prov", "Provisional", ClubLeaderboardMinGames-1, 12, 4, 16)...)

	board := RankClubLeaderboard(games)

	if len(board.Ranked) != 1 {
		t.Fatalf("expected 1 ranked, got %d", len(board.Ranked))
	}
	if board.Ranked[0].UserID != "ranked" {
		t.Errorf("expected 'ranked' user ranked, got %s", board.Ranked[0].UserID)
	}
	if board.Ranked[0].GamesPlayed != ClubLeaderboardMinGames {
		t.Errorf("ranked gamesPlayed: expected %d, got %d", ClubLeaderboardMinGames, board.Ranked[0].GamesPlayed)
	}

	if len(board.Provisional) != 1 {
		t.Fatalf("expected 1 provisional, got %d", len(board.Provisional))
	}
	p := board.Provisional[0]
	if p.UserID != "prov" {
		t.Errorf("expected 'prov' provisional, got %s", p.UserID)
	}
	if p.GamesPlayed != ClubLeaderboardMinGames-1 {
		t.Errorf("provisional gamesPlayed: expected %d, got %d", ClubLeaderboardMinGames-1, p.GamesPlayed)
	}
	if p.GamesToGo != 1 {
		t.Errorf("provisional gamesToGo: expected 1, got %d", p.GamesToGo)
	}
}

// TestRankClubLeaderboard_TiebreakForm covers the first tie-break: higher form
// wins. Both users have the same game count; A has the better margin.
func TestRankClubLeaderboard_TiebreakForm(t *testing.T) {
	var games []ClubGame
	games = append(games, nGames("a", "A", ClubLeaderboardMinGames, 14, 2, 16)...) // 0.75
	games = append(games, nGames("b", "B", ClubLeaderboardMinGames, 10, 6, 16)...) // 0.25
	board := RankClubLeaderboard(games)
	if board.Ranked[0].UserID != "a" || board.Ranked[1].UserID != "b" {
		t.Errorf("form tie-break: expected [a,b], got [%s,%s]", board.Ranked[0].UserID, board.Ranked[1].UserID)
	}
}

// TestRankClubLeaderboard_TiebreakGamesPlayed covers the second tie-break: equal
// form, more games played ranks higher.
func TestRankClubLeaderboard_TiebreakGamesPlayed(t *testing.T) {
	var games []ClubGame
	// Same 0.5 form, but A has one extra game.
	games = append(games, nGames("a", "A", ClubLeaderboardMinGames+1, 12, 4, 16)...)
	games = append(games, nGames("b", "B", ClubLeaderboardMinGames, 12, 4, 16)...)
	board := RankClubLeaderboard(games)
	if board.Ranked[0].UserID != "a" || board.Ranked[1].UserID != "b" {
		t.Errorf("gamesPlayed tie-break: expected [a,b], got [%s,%s]", board.Ranked[0].UserID, board.Ranked[1].UserID)
	}
}

// TestRankClubLeaderboard_TiebreakFirstQualifying covers the third tie-break:
// equal form and games, the member who first qualified (earliest first game)
// ranks higher.
func TestRankClubLeaderboard_TiebreakFirstQualifying(t *testing.T) {
	// Both play MinGames games of identical margin, but B's earliest game is
	// earlier than A's.
	var games []ClubGame
	for i := 0; i < ClubLeaderboardMinGames; i++ {
		games = append(games, game("a", "A", 12, 4, 16, 10+i)) // days 10..
		games = append(games, game("b", "B", 12, 4, 16, 5+i))  // days 5.. (earlier)
	}
	board := RankClubLeaderboard(games)
	if board.Ranked[0].UserID != "b" || board.Ranked[1].UserID != "a" {
		t.Errorf("firstQualifying tie-break: expected [b,a], got [%s,%s]", board.Ranked[0].UserID, board.Ranked[1].UserID)
	}
}

// TestRankClubLeaderboard_TiebreakName covers the final tie-break: everything
// equal, order by name.
func TestRankClubLeaderboard_TiebreakName(t *testing.T) {
	// Identical form, games, and dates — only the name differs.
	var games []ClubGame
	for i := 0; i < ClubLeaderboardMinGames; i++ {
		games = append(games, game("z", "Zoe", 12, 4, 16, i))
		games = append(games, game("a", "Aaron", 12, 4, 16, i))
	}
	board := RankClubLeaderboard(games)
	if board.Ranked[0].Name != "Aaron" || board.Ranked[1].Name != "Zoe" {
		t.Errorf("name tie-break: expected [Aaron,Zoe], got [%s,%s]", board.Ranked[0].Name, board.Ranked[1].Name)
	}
}

// TestRankClubLeaderboard_WinDrawLoss checks the per-game record: a win when the
// member's team outscores the opponent, a draw when equal, a loss otherwise, and
// Wins+Draws+Losses == GamesPlayed.
func TestRankClubLeaderboard_WinDrawLoss(t *testing.T) {
	var games []ClubGame
	// 5 games (clears the gate): 3 wins, 1 draw, 1 loss.
	games = append(games, game("u", "U", 16, 8, 24, 0))  // win
	games = append(games, game("u", "U", 20, 4, 24, 1))  // win
	games = append(games, game("u", "U", 13, 11, 24, 2)) // win
	games = append(games, game("u", "U", 12, 12, 24, 3)) // draw
	games = append(games, game("u", "U", 5, 19, 24, 4))  // loss

	board := RankClubLeaderboard(games)
	if len(board.Ranked) != 1 {
		t.Fatalf("expected 1 ranked, got %d", len(board.Ranked))
	}
	e := board.Ranked[0]
	if e.Wins != 3 || e.Draws != 1 || e.Losses != 1 {
		t.Errorf("record: got %d-%d-%d, want 3-1-1", e.Wins, e.Draws, e.Losses)
	}
	if e.Wins+e.Draws+e.Losses != e.GamesPlayed {
		t.Errorf("W+D+L (%d) must equal GamesPlayed (%d)", e.Wins+e.Draws+e.Losses, e.GamesPlayed)
	}
}

// TestRankClubLeaderboard_Empty returns empty, non-nil slices so the API always
// serializes to [] rather than null, and echoes the MinGames threshold.
func TestRankClubLeaderboard_Empty(t *testing.T) {
	board := RankClubLeaderboard(nil)
	if board.Ranked == nil {
		t.Error("Ranked should be non-nil empty slice")
	}
	if board.Provisional == nil {
		t.Error("Provisional should be non-nil empty slice")
	}
	if board.MinGames != ClubLeaderboardMinGames {
		t.Errorf("MinGames = %d, want %d", board.MinGames, ClubLeaderboardMinGames)
	}
}

func approx(a, b float64) bool {
	const eps = 1e-9
	d := a - b
	return d < eps && d > -eps
}
