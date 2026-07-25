package domain

import (
	"sort"
	"time"
)

// Club leaderboard tuning — global domain constants per #129.
//
// ClubLeaderboardMinGames is how many scored games a member must have played
// (within the window) to earn a ranked row; below it they show as provisional
// ("N more to rank"). ClubLeaderboardWindow is the rolling look-back applied by
// the store when selecting qualifying games, so the board reflects current form
// rather than lifetime record.
const (
	ClubLeaderboardMinGames = 5
	ClubLeaderboardWindow   = 90 * 24 * time.Hour
)

// ClubGame is one scored game a Club member played, seen from that member's own
// perspective: their team's points (TeamPoints) versus the opposing team's
// (OppPoints), plus the game's PointsTarget for normalization and PlayedAt for
// the first-qualifying tie-break. The store emits one ClubGame per registered
// member per scored club Match (Guests produce none); the window is already
// applied upstream. Name/AvatarIcon/AvatarColor carry the member's identity for
// display and the name tie-break.
type ClubGame struct {
	UserID       string
	Name         string
	AvatarIcon   string
	AvatarColor  string
	TeamPoints   int
	OppPoints    int
	PointsTarget int
	PlayedAt     time.Time
}

// ClubLeaderboard is the ranked board plus the provisional "not yet ranked"
// list. Both slices are always non-nil so the API serializes them as []. MinGames
// echoes the qualifying threshold so the client can phrase "play N to rank"
// copy from the source of truth rather than hardcoding it.
type ClubLeaderboard struct {
	Ranked      []ClubRankEntry        `json:"ranked"`
	Provisional []ClubProvisionalEntry `json:"provisional"`
	MinGames    int                    `json:"min_games"`
}

// ClubRankEntry is one ranked member. Form is the mean normalized point-margin
// across their qualifying games, in [-1, +1]; positive means they win points by
// a wider margin than they concede. Wins/Draws/Losses are the member's per-game
// record over the same qualifying games (Wins+Draws+Losses == GamesPlayed).
type ClubRankEntry struct {
	Rank        int     `json:"rank"`
	UserID      string  `json:"user_id"`
	Name        string  `json:"name"`
	AvatarIcon  string  `json:"avatar_icon"`
	AvatarColor string  `json:"avatar_color"`
	Form        float64 `json:"form"`
	GamesPlayed int     `json:"games_played"`
	Wins        int     `json:"wins"`
	Draws       int     `json:"draws"`
	Losses      int     `json:"losses"`
}

// ClubProvisionalEntry is a member who hasn't yet played enough to rank.
// GamesToGo is how many more qualifying games they need (MinGames - played).
type ClubProvisionalEntry struct {
	UserID      string `json:"user_id"`
	Name        string `json:"name"`
	AvatarIcon  string `json:"avatar_icon"`
	AvatarColor string `json:"avatar_color"`
	GamesPlayed int    `json:"games_played"`
	GamesToGo   int    `json:"games_to_go"`
}

// clubAgg accumulates one member's games while scanning.
type clubAgg struct {
	userID      string
	name        string
	avatarIcon  string
	avatarColor string
	sumMargin   float64
	games       int
	wins        int
	draws       int
	firstPlayed time.Time
}

// RankClubLeaderboard computes the Club leaderboard from qualifying games. It is
// pure — no clock, no DB: the store has already filtered to scored club Matches
// within the window and expanded them into per-member ClubGame rows. Each
// member's form is the mean of their per-game normalized margins
// ((TeamPoints - OppPoints) / PointsTarget). Members with at least
// ClubLeaderboardMinGames games are ranked; the rest are provisional. This is
// the metric seam from #129 — all ranking logic lives here, unit-testable
// without a database.
func RankClubLeaderboard(games []ClubGame) ClubLeaderboard {
	byUser := map[string]*clubAgg{}
	for _, g := range games {
		a := byUser[g.UserID]
		if a == nil {
			a = &clubAgg{userID: g.UserID, firstPlayed: g.PlayedAt}
			byUser[g.UserID] = a
		}
		// Latest-seen identity wins (display names can change between games).
		a.name = g.Name
		a.avatarIcon = g.AvatarIcon
		a.avatarColor = g.AvatarColor

		margin := 0.0
		if g.PointsTarget != 0 {
			margin = float64(g.TeamPoints-g.OppPoints) / float64(g.PointsTarget)
		}
		a.sumMargin += margin
		a.games++
		switch {
		case g.TeamPoints > g.OppPoints:
			a.wins++
		case g.TeamPoints == g.OppPoints:
			a.draws++
		}
		if g.PlayedAt.Before(a.firstPlayed) {
			a.firstPlayed = g.PlayedAt
		}
	}

	ranked := []ClubRankEntry{}
	provisional := []ClubProvisionalEntry{}
	for _, a := range byUser {
		if a.games >= ClubLeaderboardMinGames {
			ranked = append(ranked, ClubRankEntry{
				UserID:      a.userID,
				Name:        a.name,
				AvatarIcon:  a.avatarIcon,
				AvatarColor: a.avatarColor,
				Form:        a.sumMargin / float64(a.games),
				GamesPlayed: a.games,
				Wins:        a.wins,
				Draws:       a.draws,
				Losses:      a.games - a.wins - a.draws,
			})
		} else {
			provisional = append(provisional, ClubProvisionalEntry{
				UserID:      a.userID,
				Name:        a.name,
				AvatarIcon:  a.avatarIcon,
				AvatarColor: a.avatarColor,
				GamesPlayed: a.games,
				GamesToGo:   ClubLeaderboardMinGames - a.games,
			})
		}
	}

	// Ranked tie-break chain (#129): form DESC → gamesPlayed DESC → earliest
	// first-qualifying date → name. firstPlayed lives only in the aggregate, so
	// resolve it via byUser during the sort.
	sort.SliceStable(ranked, func(i, j int) bool {
		x, y := ranked[i], ranked[j]
		if x.Form != y.Form {
			return x.Form > y.Form
		}
		if x.GamesPlayed != y.GamesPlayed {
			return x.GamesPlayed > y.GamesPlayed
		}
		fx, fy := byUser[x.UserID].firstPlayed, byUser[y.UserID].firstPlayed
		if !fx.Equal(fy) {
			return fx.Before(fy)
		}
		return x.Name < y.Name
	})
	for i := range ranked {
		ranked[i].Rank = i + 1
	}

	// Provisional: closest to qualifying first (most games), then name. Not a
	// ranked board, so no rank numbers.
	sort.SliceStable(provisional, func(i, j int) bool {
		x, y := provisional[i], provisional[j]
		if x.GamesPlayed != y.GamesPlayed {
			return x.GamesPlayed > y.GamesPlayed
		}
		return x.Name < y.Name
	})

	return ClubLeaderboard{Ranked: ranked, Provisional: provisional, MinGames: ClubLeaderboardMinGames}
}
