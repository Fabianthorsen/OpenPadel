package domain

import "time"

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	DisplayName  string    `json:"display_name"`
	AvatarIcon   string    `json:"avatar_icon"`
	AvatarColor  string    `json:"avatar_color"`
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
	ID              string       `json:"id"`
	SessionID       string       `json:"session_id"`
	SessionName     string       `json:"session_name"`
	FromUserID      string       `json:"from_user_id"`
	FromDisplayName string       `json:"from_display_name"`
	ToUserID        string       `json:"to_user_id"`
	ToDisplayName   string       `json:"to_display_name,omitempty"`
	Status          InviteStatus `json:"status"`
	CreatedAt       time.Time    `json:"created_at"`
}

type UserSearchResult struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	IsContact   bool   `json:"is_contact"`
	AvatarIcon  string `json:"avatar_icon"`
	AvatarColor string `json:"avatar_color"`
}

type AmericanoCareerStats struct {
	GamesPlayed int `json:"games_played"`
	Wins        int `json:"wins"`
	Draws       int `json:"draws"`
	Losses      int `json:"losses"`
	TotalPoints int `json:"total_points"`
	Tournaments int `json:"tournaments"`
}

type TournamentHistoryEntry struct {
	SessionID   string `json:"session_id"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	PlayedAt    string `json:"played_at"`
	Rank        int    `json:"rank"`
	Points      int    `json:"points"`
	GamesPlayed int    `json:"games_played"`
	EndedEarly  bool   `json:"ended_early"`
}

type UpcomingEntry struct {
	SessionID   string     `json:"session_id"`
	Name        string     `json:"name"`
	Status      string     `json:"status"`
	GameMode    GameMode   `json:"game_mode"`
	Courts      int        `json:"courts"`
	PlayerCount int        `json:"player_count"`
	ScheduledAt *time.Time `json:"scheduled_at,omitempty"`
}

type SessionStatus string

const (
	StatusLobby    SessionStatus = "lobby"
	StatusPlaying  SessionStatus = "playing"
	StatusDone     SessionStatus = "done"
)

func (s SessionStatus) IsValid() bool {
	return s == StatusLobby || s == StatusPlaying || s == StatusDone
}

func (s SessionStatus) Values() []SessionStatus {
	return []SessionStatus{StatusLobby, StatusPlaying, StatusDone}
}

type GameMode string

const MaxCourts = 4

const (
	ModeAmericano GameMode = "americano"
	ModeMexicano  GameMode = "mexicano"
)

func (g GameMode) IsValid() bool {
	return g == ModeAmericano || g == ModeMexicano
}

func (g GameMode) Values() []GameMode {
	return []GameMode{ModeAmericano, ModeMexicano}
}

type InviteStatus string

const (
	InvitePending  InviteStatus = "pending"
	InviteAccepted InviteStatus = "accepted"
	InviteDeclined InviteStatus = "declined"
)

func (s InviteStatus) IsValid() bool {
	return s == InvitePending || s == InviteAccepted || s == InviteDeclined
}

type SessionInput struct {
	Courts               int
	Points               int
	Name                 string
	GameMode             GameMode
	RoundsTotal          *int
	ScheduledAt          *time.Time
	CourtDurationMinutes *int
}

func (si SessionInput) Validate() []ValidationError {
	var errs []ValidationError

	if !si.GameMode.IsValid() {
		errs = append(errs, ValidationError{
			Code: "invalid_game_mode",
		})
	}

	minCourts := 1
	if si.GameMode == ModeMexicano {
		minCourts = 2
	}
	if si.Courts < minCourts || si.Courts > 4 {
		errs = append(errs, ValidationError{
			Code: "invalid_courts",
		})
	}

	if si.Points != 16 && si.Points != 24 && si.Points != 32 {
		errs = append(errs, ValidationError{
			Code: "invalid_points",
		})
	}

	if si.GameMode == ModeMexicano && si.RoundsTotal != nil {
		if *si.RoundsTotal < 1 || *si.RoundsTotal > 20 {
			errs = append(errs, ValidationError{
				Code: "invalid_rounds_total",
			})
		}
	}

	if si.CourtDurationMinutes != nil {
		if *si.CourtDurationMinutes < 15 || *si.CourtDurationMinutes > 300 {
			errs = append(errs, ValidationError{
				Code: "invalid_court_duration",
			})
		}
	}

	return errs
}

type Session struct {
	ID                       string        `json:"id"`
	AdminToken               string        `json:"admin_token,omitempty"`
	Status                   SessionStatus `json:"status"`
	Name                     string        `json:"name,omitempty"`
	GameMode                 GameMode      `json:"game_mode"`
	Courts                   int           `json:"courts"`
	Points                   int           `json:"points"`
	RoundsTotal              *int          `json:"rounds_total,omitempty"`
	CurrentRound             *int          `json:"current_round,omitempty"`
	CreatorPlayerID          string        `json:"creator_player_id,omitempty"`
	CreatorUserID            string        `json:"-"`
	IsCreator                bool          `json:"is_creator,omitempty"`
	ScheduledAt              *time.Time    `json:"scheduled_at,omitempty"`
	CourtDurationMinutes     *int          `json:"court_duration_minutes,omitempty"`
	EndsAt                   *time.Time    `json:"ends_at,omitempty"`
	Players                  []Player         `json:"players"`
	ValidationErrors         []ValidationError `json:"validation_errors,omitempty"`
	CanStart                 bool             `json:"can_start"`
	CreatedAt                time.Time        `json:"created_at"`
	UpdatedAt                time.Time        `json:"updated_at"`
}

type Player struct {
	ID          string    `json:"id"`
	SessionID   string    `json:"session_id"`
	UserID      string    `json:"user_id,omitempty"`
	Name        string    `json:"name"`
	AvatarIcon  string    `json:"avatar_icon"`
	AvatarColor string    `json:"avatar_color"`
	Active      bool      `json:"active"`
	JoinedAt    time.Time `json:"joined_at"`
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
	Rank           int     `json:"rank"`
	PlayerID       string  `json:"player_id"`
	UserID         *string `json:"user_id,omitempty"`
	Name           string  `json:"name"`
	Points         int     `json:"points"`
	PointsConceded int     `json:"points_conceded"`
	GamesPlayed    int     `json:"games_played"`
	Wins           int     `json:"wins"`
	Draws          int     `json:"draws"`
	AvatarIcon     string  `json:"avatar_icon"`
	AvatarColor    string  `json:"avatar_color"`
}

type Leaderboard struct {
	SessionID    string        `json:"session_id"`
	Status       SessionStatus `json:"status"`
	CurrentRound *int          `json:"current_round"`
	TotalRounds  *int          `json:"total_rounds"`
	Standings    []Standing    `json:"standings"`
	UpdatedAt    time.Time     `json:"updated_at"`
}
