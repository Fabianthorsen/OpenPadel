package domain

import "time"

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	DisplayName  string    `json:"display_name"`
	AvatarIcon   string    `json:"avatar_icon"`
	AvatarColor  string    `json:"avatar_color"`
	PasswordHash string    `json:"-"`
	SelfRating   *int      `json:"self_rating"`
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

// CareerSummary is the cross-mode profile headline: the numbers that stay honest
// blended across Game Modes (ADR 0007). PointWinPct and Winrate are percentages
// in [0, 100]; both are zero when Games is zero (callers hide the stat rather
// than render a jarring 0% / 100%).
//
// The placement fields (Titles, Podiums, BestFinish, AverageFinish) come from
// the player's finishing rank in each done Session, ranked by the same
// leaderboard tiebreaker chain GetTournamentHistory uses. Unlike point-share,
// a finishing rank compares like-for-like across scoring models — 1st is 1st in
// any mode — so these are aggregated across both Game Modes here rather than
// segmented per mode. Titles counts rank-1 finishes, Podiums rank ≤ 3,
// BestFinish is the lowest (best) rank, AverageFinish the mean rank. All are
// zero when Games is zero.
type CareerSummary struct {
	Games         int     `json:"games"`
	Winrate       float64 `json:"winrate"`
	PointWinPct   float64 `json:"point_win_pct"`
	Titles        int     `json:"titles"`
	Podiums       int     `json:"podiums"`
	BestFinish    int     `json:"best_finish"`
	AverageFinish float64 `json:"average_finish"`
}

// ModeStats is the per-Game-Mode career aggregate behind the Career Stats page
// (ADR 0007). Point-share is only apples-to-apples within one scoring model, so
// Americano and Mexicano are aggregated separately and never blended here.
// PointWinPct is a percentage in [0, 100] (mean per-Match share); NetPoints is
// TotalPoints − PointsConceded and may be negative. A mode the user has no games
// in is still returned, zero-valued, so the UI can render its "no games yet"
// state. Same fully-scored / guest-inclusive / compute-on-read rules as
// CareerSummary.
type ModeStats struct {
	Mode           GameMode `json:"mode"`
	Games          int      `json:"games"`
	Wins           int      `json:"wins"`
	Draws          int      `json:"draws"`
	Losses         int      `json:"losses"`
	TotalPoints    int      `json:"total_points"`
	PointsConceded int      `json:"points_conceded"`
	NetPoints      int      `json:"net_points"`
	PointWinPct    float64  `json:"point_win_pct"`
	Tournaments    int      `json:"tournaments"`
}

// MatchOutcome is a Match's win/draw/loss result from the scoring player's side.
type MatchOutcome string

const (
	MatchResultWin  MatchOutcome = "win"
	MatchResultDraw MatchOutcome = "draw"
	MatchResultLoss MatchOutcome = "loss"
)

// MatchResult is one row of the per-Match results series behind the Career Stats
// page's recent-form curve (ADR 0007). One row per fully-scored Match the player
// took part in within a done Session, ordered oldest-first so the client can read
// it as a time series and derive the form stats client-side (no new endpoint per
// stat). Per-Match rather than per-Session because the Match is ADR 0007's atomic
// Point Win % unit — see the ADR for why.
//
// Points/Conceded are the player's own-team and opponent-team score for that
// Match; Result is the outcome from that differential (win when Points >
// Conceded, draw when equal, loss otherwise). Date is the Session's date (Matches
// carry no timestamp of their own). Point Win % per match is derived client-side
// from Points/Conceded, so it isn't carried here.
type MatchResult struct {
	MatchID  string       `json:"match_id"`
	Mode     GameMode     `json:"mode"`
	Date     string       `json:"date"`
	Points   int          `json:"points"`
	Conceded int          `json:"conceded"`
	Result   MatchOutcome `json:"result"`
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
	StatusLobby   SessionStatus = "lobby"
	StatusPlaying SessionStatus = "playing"
	StatusDone    SessionStatus = "done"
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
	// ClubID, when set, marks the Session as a club event owned by that Club. It
	// is identity/ownership metadata only — it grants no authority over the
	// Session, which stays governed by the AdminToken like any other.
	ClubID *string
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
	ID                   string            `json:"id"`
	AdminToken           string            `json:"admin_token,omitempty"`
	Status               SessionStatus     `json:"status"`
	Name                 string            `json:"name,omitempty"`
	GameMode             GameMode          `json:"game_mode"`
	Courts               int               `json:"courts"`
	Points               int               `json:"points"`
	RoundsTotal          *int              `json:"rounds_total,omitempty"`
	CurrentRound         *int              `json:"current_round,omitempty"`
	CreatorPlayerID      string            `json:"creator_player_id,omitempty"`
	CreatorUserID        string            `json:"-"`
	ClubID               string            `json:"club_id,omitempty"`
	IsCreator            bool              `json:"is_creator,omitempty"`
	ScheduledAt          *time.Time        `json:"scheduled_at,omitempty"`
	CourtDurationMinutes *int              `json:"court_duration_minutes,omitempty"`
	EndsAt               *time.Time        `json:"ends_at,omitempty"`
	Players              []Player          `json:"players"`
	ValidationErrors     []ValidationError `json:"validation_errors,omitempty"`
	CanStart             bool              `json:"can_start"`
	CreatedAt            time.Time         `json:"created_at"`
	UpdatedAt            time.Time         `json:"updated_at"`
}

type Player struct {
	ID          string `json:"id"`
	SessionID   string `json:"session_id"`
	UserID      string `json:"user_id,omitempty"`
	Name        string `json:"name"`
	AvatarIcon  string `json:"avatar_icon"`
	AvatarColor string `json:"avatar_color"`
	Rating      int    `json:"rating"`
	// AddedByAdmin marks a guest the admin created by hand (via an admin token).
	// Only these players may have their rating inline-edited by the admin (#211).
	AddedByAdmin bool      `json:"added_by_admin"`
	Active       bool      `json:"active"`
	JoinedAt     time.Time `json:"joined_at"`
	// Token is the per-player secret returned only to the joining client (in the
	// join response), required to self-remove a guest (#241). It is never loaded
	// into the shared session listing, so it stays empty (and omitted) there.
	Token string `json:"player_token,omitempty"`
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

type Club struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	AvatarIcon  string    `json:"avatar_icon"`
	AvatarColor string    `json:"avatar_color"`
	JoinCode    string    `json:"join_code"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
}

type ClubMember struct {
	UserID      string    `json:"user_id"`
	DisplayName string    `json:"display_name"`
	Role        string    `json:"role"`
	AvatarIcon  string    `json:"avatar_icon"`
	AvatarColor string    `json:"avatar_color"`
	JoinedAt    time.Time `json:"joined_at"`
}

// ClubInvite is a pending request for a specific registered User to join a Club.
// It is distinct from an Invite, which targets a Session.
type ClubInvite struct {
	ID                 string       `json:"id"`
	ClubID             string       `json:"club_id"`
	ClubName           string       `json:"club_name"`
	ClubAvatarIcon     string       `json:"club_avatar_icon"`
	ClubAvatarColor    string       `json:"club_avatar_color"`
	InviterID          string       `json:"inviter_id"`
	InviterDisplayName string       `json:"inviter_display_name"`
	InviteeID          string       `json:"invitee_id"`
	Status             InviteStatus `json:"status"`
	CreatedAt          time.Time    `json:"created_at"`
}

type ClubDetail struct {
	Club        Club         `json:"club"`
	Members     []ClubMember `json:"members"`
	IsAdmin     bool         `json:"is_admin"`
	MyRole      string       `json:"my_role,omitempty"`
	RosterCount int          `json:"roster_count"`
}
