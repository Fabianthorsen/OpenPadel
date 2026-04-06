package domain

import "time"

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	DisplayName  string    `json:"display_name"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

type AuthToken struct {
	Token  string `json:"token"`
	UserID string `json:"user_id"`
}

type Contact struct {
	UserID      string    `json:"user_id"`
	DisplayName string    `json:"display_name"`
	AddedAt     time.Time `json:"added_at"`
}

type Invite struct {
	ID              string    `json:"id"`
	SessionID       string    `json:"session_id"`
	SessionName     string    `json:"session_name"`
	FromUserID      string    `json:"from_user_id"`
	FromDisplayName string    `json:"from_display_name"`
	ToUserID        string    `json:"to_user_id"`
	Status          string    `json:"status"` // pending | accepted | declined
	CreatedAt       time.Time `json:"created_at"`
}

type UserSearchResult struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	IsContact   bool   `json:"is_contact"`
}

type AmericanoCareerStats struct {
	GamesPlayed  int `json:"games_played"`
	Wins         int `json:"wins"`
	Draws        int `json:"draws"`
	Losses       int `json:"losses"`
	TotalPoints  int `json:"total_points"`
	Tournaments  int `json:"tournaments"`
}

type TennisCareerStats struct {
	Tournaments int `json:"tournaments"`
	Wins        int `json:"wins"`
	Losses      int `json:"losses"`
}

type TournamentHistoryEntry struct {
	SessionID   string `json:"session_id"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	PlayedAt    string `json:"played_at"`
	Rank        int    `json:"rank"`
	Points      int    `json:"points"`
	GamesPlayed int    `json:"games_played"`
}

type UpcomingEntry struct {
	SessionID   string     `json:"session_id"`
	Name        string     `json:"name"`
	Status      string     `json:"status"`
	GameMode    string     `json:"game_mode"`
	Courts      int        `json:"courts"`
	PlayerCount int        `json:"player_count"`
	ScheduledAt *time.Time `json:"scheduled_at,omitempty"`
}

type SessionStatus string

const (
	StatusLobby    SessionStatus = "lobby"
	StatusActive   SessionStatus = "active"
	StatusComplete SessionStatus = "complete"
)

type Session struct {
	ID              string        `json:"id"`
	AdminToken      string        `json:"admin_token,omitempty"`
	Status          SessionStatus `json:"status"`
	Name            string        `json:"name,omitempty"`
	GameMode        string        `json:"game_mode"`
	SetsToWin       int           `json:"sets_to_win"`
	GamesPerSet     int           `json:"games_per_set"`
	Courts          int           `json:"courts"`
	Points          int           `json:"points"`
	RoundsTotal     *int          `json:"rounds_total,omitempty"`
	CurrentRound    *int          `json:"current_round,omitempty"`
	CreatorPlayerID string        `json:"creator_player_id,omitempty"`
	ScheduledAt     *time.Time    `json:"scheduled_at,omitempty"`
	Players         []Player      `json:"players"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
}

// Tennis types

type TennisTeam struct {
	PlayerID string `json:"player_id"`
	Name     string `json:"name"`
	Team     string `json:"team"` // "a" or "b"
}

type TennisState struct {
	Sets       [][2]int `json:"sets"`
	GamesA     int      `json:"games_a"`
	GamesB     int      `json:"games_b"`
	PointsA    int      `json:"points_a"`
	PointsB    int      `json:"points_b"`
	InTiebreak bool     `json:"in_tiebreak"`
	TiebreakA  int      `json:"tiebreak_a"`
	TiebreakB  int      `json:"tiebreak_b"`
	Server     string   `json:"server"` // "a" or "b"
	Winner     string   `json:"winner"` // "a", "b", or ""
}

type TennisMatch struct {
	ID        string      `json:"id"`
	SessionID string      `json:"session_id"`
	State     TennisState `json:"state"`
	Teams     TennisTeams `json:"teams"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

type TennisTeams struct {
	A []TennisTeam `json:"a"`
	B []TennisTeam `json:"b"`
}

type Player struct {
	ID        string    `json:"id"`
	SessionID string    `json:"session_id"`
	UserID    string    `json:"user_id,omitempty"`
	Name      string    `json:"name"`
	Active    bool      `json:"active"`
	JoinedAt  time.Time `json:"joined_at"`
}

type Round struct {
	ID        string   `json:"id"`
	SessionID string   `json:"session_id"`
	Number    int      `json:"number"`
	Bench     []string `json:"bench"`
	Matches   []Match  `json:"matches"`
}

type Match struct {
	ID      string     `json:"id"`
	RoundID string     `json:"round_id"`
	Court   int        `json:"court"`
	TeamA   [2]string  `json:"team_a"`
	TeamB   [2]string  `json:"team_b"`
	Score   *Score     `json:"score"`
	Live    *LiveScore `json:"live,omitempty"`
}

type Score struct {
	A int `json:"a"`
	B int `json:"b"`
}

type LiveScore struct {
	A      int    `json:"a"`
	B      int    `json:"b"`
	Server string `json:"server,omitempty"`
}

type Standing struct {
	Rank        int     `json:"rank"`
	PlayerID    string  `json:"player_id"`
	UserID      *string `json:"user_id,omitempty"`
	Name        string  `json:"name"`
	Points      int     `json:"points"`
	GamesPlayed int     `json:"games_played"`
	Wins        int     `json:"wins"`
	Draws       int     `json:"draws"`
}

type Leaderboard struct {
	SessionID    string        `json:"session_id"`
	Status       SessionStatus `json:"status"`
	CurrentRound *int          `json:"current_round"`
	TotalRounds  *int          `json:"total_rounds"`
	Standings    []Standing    `json:"standings"`
	UpdatedAt    time.Time     `json:"updated_at"`
}
