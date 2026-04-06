// Package tennis implements the scoring logic for Regular 2v2 tennis matches.
// Rules: standard sets (first to gamesPerSet, win by 2; tiebreak at gamesPerSet-gamesPerSet),
// best of N sets, golden point at deuce (40:40 → next point wins).
package tennis

import "github.com/fabianthorsen/openpadel/internal/domain"

// AddPoint applies a point to team "a" or "b" and returns the updated state.
// If the match already has a winner, state is returned unchanged.
func AddPoint(state domain.TennisState, team string, setsToWin, gamesPerSet int) domain.TennisState {
	if state.Winner != "" {
		return state
	}
	if state.Sets == nil {
		state.Sets = [][2]int{}
	}
	if gamesPerSet != 4 && gamesPerSet != 6 {
		gamesPerSet = 6
	}

	// Rotate server: changes every game (simplified — server starts with whoever state.Server is).
	// We track server in state; it flips on each winGame call.

	if state.InTiebreak {
		return applyTiebreakPoint(state, team, setsToWin, gamesPerSet)
	}
	return applyGamePoint(state, team, setsToWin, gamesPerSet)
}

func applyGamePoint(state domain.TennisState, team string, setsToWin, gamesPerSet int) domain.TennisState {
	// Golden point: at 3-3 (40:40), the next point wins the game.
	if state.PointsA == 3 && state.PointsB == 3 {
		return winGame(state, team, setsToWin, gamesPerSet)
	}

	if team == "a" {
		state.PointsA++
	} else {
		state.PointsB++
	}

	if state.PointsA >= 4 {
		return winGame(state, "a", setsToWin, gamesPerSet)
	}
	if state.PointsB >= 4 {
		return winGame(state, "b", setsToWin, gamesPerSet)
	}
	return state
}

func winGame(state domain.TennisState, team string, setsToWin, gamesPerSet int) domain.TennisState {
	state.PointsA = 0
	state.PointsB = 0
	if team == "a" {
		state.GamesA++
	} else {
		state.GamesB++
	}
	// Flip server on each game change (only when server has been designated).
	if state.Server == "a" {
		state.Server = "b"
	} else if state.Server == "b" {
		state.Server = "a"
	}
	return checkSetWin(state, setsToWin, gamesPerSet)
}

func checkSetWin(state domain.TennisState, setsToWin, gamesPerSet int) domain.TennisState {
	g := gamesPerSet
	// Tiebreak at g-g.
	if state.GamesA == g && state.GamesB == g {
		state.InTiebreak = true
		state.TiebreakA = 0
		state.TiebreakB = 0
		return state
	}
	// Win set: first to g, win by 2. Can also reach g+1 - g-1.
	if (state.GamesA >= g && state.GamesA-state.GamesB >= 2) || state.GamesA == g+1 {
		return winSet(state, "a", setsToWin)
	}
	if (state.GamesB >= g && state.GamesB-state.GamesA >= 2) || state.GamesB == g+1 {
		return winSet(state, "b", setsToWin)
	}
	return state
}

func winSet(state domain.TennisState, team string, setsToWin int) domain.TennisState {
	state.Sets = append(state.Sets, [2]int{state.GamesA, state.GamesB})
	state.GamesA = 0
	state.GamesB = 0
	state.InTiebreak = false
	state.TiebreakA = 0
	state.TiebreakB = 0

	setsA, setsB := countSets(state.Sets)
	if setsA == setsToWin {
		state.Winner = "a"
	} else if setsB == setsToWin {
		state.Winner = "b"
	}
	return state
}

func applyTiebreakPoint(state domain.TennisState, team string, setsToWin, gamesPerSet int) domain.TennisState {
	if team == "a" {
		state.TiebreakA++
	} else {
		state.TiebreakB++
	}
	// Tiebreak: first to 7, win by 2.
	if state.TiebreakA >= 7 && state.TiebreakA-state.TiebreakB >= 2 {
		state.GamesA = gamesPerSet + 1
		state.GamesB = gamesPerSet
		return winSet(state, "a", setsToWin)
	}
	if state.TiebreakB >= 7 && state.TiebreakB-state.TiebreakA >= 2 {
		state.GamesA = gamesPerSet
		state.GamesB = gamesPerSet + 1
		return winSet(state, "b", setsToWin)
	}
	return state
}

func countSets(sets [][2]int) (a, b int) {
	for _, s := range sets {
		if s[0] > s[1] {
			a++
		} else if s[1] > s[0] {
			b++
		}
	}
	return
}
