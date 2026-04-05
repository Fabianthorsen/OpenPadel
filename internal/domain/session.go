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

type CareerStats struct {
	GamesPlayed  int `json:"games_played"`
	Wins         int `json:"wins"`
	Draws        int `json:"draws"`
	Losses       int `json:"losses"`
	TotalPoints  int `json:"total_points"`
	Tournaments  int `json:"tournaments"`
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
	Rank        int    `json:"rank"`
	PlayerID    string `json:"player_id"`
	Name        string `json:"name"`
	Points      int    `json:"points"`
	GamesPlayed int    `json:"games_played"`
	Wins        int    `json:"wins"`
	Draws       int    `json:"draws"`
}

type Leaderboard struct {
	SessionID    string        `json:"session_id"`
	Status       SessionStatus `json:"status"`
	CurrentRound *int          `json:"current_round"`
	TotalRounds  *int          `json:"total_rounds"`
	Standings    []Standing    `json:"standings"`
	UpdatedAt    time.Time     `json:"updated_at"`
}
