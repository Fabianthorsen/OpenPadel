package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/fabianthorsen/openpadel/internal/domain"
	"github.com/fabianthorsen/openpadel/internal/store/db"
)

var ErrNotFound = errors.New("not found")

func (s *Store) CreateSession(courts, points int, name, gameMode string, roundsTotal *int, scheduledAt *time.Time, courtDurationMinutes *int, creatorUserID string) (*domain.Session, error) {
	now := time.Now().UTC()
	if gameMode == "" {
		gameMode = "americano"
	}
	sess := &domain.Session{
		ID:                       newID(),
		AdminToken:               newAdminToken(),
		Status:                   domain.StatusLobby,
		Name:                     name,
		GameMode:                 domain.GameMode(gameMode),
		Courts:                   courts,
		Points:                   points,
		RoundsTotal:              roundsTotal,
		ScheduledAt:              scheduledAt,
		CourtDurationMinutes:     courtDurationMinutes,
		CreatorUserID:            creatorUserID,
		Players:                  []domain.Player{},
		CreatedAt:                now,
		UpdatedAt:                now,
	}
	var scheduledAtStr sql.NullString
	if scheduledAt != nil {
		scheduledAtStr = sql.NullString{String: scheduledAt.UTC().Format(time.RFC3339), Valid: true}
	}
	var courtDurationMinutesVal sql.NullInt64
	if courtDurationMinutes != nil {
		courtDurationMinutesVal = sql.NullInt64{Int64: int64(*courtDurationMinutes), Valid: true}
	}
	var roundsTotalVal sql.NullInt64
	if roundsTotal != nil {
		roundsTotalVal = sql.NullInt64{Int64: int64(*roundsTotal), Valid: true}
	}
	var creatorUserIDVal sql.NullString
	if creatorUserID != "" {
		creatorUserIDVal = sql.NullString{String: creatorUserID, Valid: true}
	}
	err := s.queries.CreateSession(context.Background(), db.CreateSessionParams{
		ID:                   sess.ID,
		AdminToken:           sess.AdminToken,
		Status:               string(sess.Status),
		Name:                 sess.Name,
		GameMode:             string(sess.GameMode),
		SetsToWin:            0,
		GamesPerSet:          0,
		Courts:               int64(sess.Courts),
		Points:               int64(sess.Points),
		RoundsTotal:          roundsTotalVal,
		ScheduledAt:          scheduledAtStr,
		CourtDurationMinutes: courtDurationMinutesVal,
		CreatorUserID:        creatorUserIDVal,
		CreatedAt:            sess.CreatedAt.Format(time.RFC3339),
		UpdatedAt:            sess.UpdatedAt.Format(time.RFC3339),
	})
	return sess, err
}

func (s *Store) GetSession(id string) (*domain.Session, error) {
	row, err := s.queries.GetSession(context.Background(), id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	sess := rowToSession(row)
	players, err := s.GetPlayers(id)
	if err != nil {
		return nil, err
	}
	sess.Players = players
	return sess, nil
}

func (s *Store) SetCreatorPlayer(sessionID, playerID string) error {
	return s.queries.SetCreatorPlayer(context.Background(), db.SetCreatorPlayerParams{
		CreatorPlayerID: sql.NullString{String: playerID, Valid: true},
		UpdatedAt:       time.Now().UTC().Format(time.RFC3339),
		ID:              sessionID,
	})
}

func (s *Store) StartSession(id string, roundsTotal int, endsAt *time.Time) error {
	var endsAtStr sql.NullString
	if endsAt != nil {
		endsAtStr = sql.NullString{String: endsAt.UTC().Format(time.RFC3339), Valid: true}
	}
	return s.queries.StartSession(context.Background(), db.StartSessionParams{
		Status:      string(domain.StatusPlaying),
		RoundsTotal: sql.NullInt64{Int64: int64(roundsTotal), Valid: true},
		EndsAt:      endsAtStr,
		UpdatedAt:   time.Now().UTC().Format(time.RFC3339),
		ID:          id,
	})
}

// StartMexicanoSession activates the session with current_round=1.
// rounds_total is preserved from creation time (null = open-ended, N = preset).
func (s *Store) StartMexicanoSession(id string, endsAt *time.Time) error {
	var endsAtStr sql.NullString
	if endsAt != nil {
		endsAtStr = sql.NullString{String: endsAt.UTC().Format(time.RFC3339), Valid: true}
	}
	return s.queries.StartMexicanoSession(context.Background(), db.StartMexicanoSessionParams{
		Status:    string(domain.StatusPlaying),
		EndsAt:    endsAtStr,
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
		ID:        id,
	})
}

func (s *Store) AdvanceRound(id string) error {
	return s.queries.AdvanceRound(context.Background(), db.AdvanceRoundParams{
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
		ID:        id,
	})
}

func (s *Store) CurrentRoundAllScored(sessionID string) (bool, error) {
	unscored, err := s.queries.CurrentRoundAllScored(context.Background(), sessionID)
	return unscored == 0, err
}

func (s *Store) IncrementTournamentWinCount(userID string) error {
	return s.queries.IncrementTournamentWinCount(context.Background(), userID)
}

func (s *Store) CompleteSession(id string, endedEarly bool) error {
	endedEarlyInt := int64(0)
	if endedEarly {
		endedEarlyInt = 1
	}
	return s.queries.CompleteSession(context.Background(), db.CompleteSessionParams{
		Status:     string(domain.StatusDone),
		EndedEarly: endedEarlyInt,
		UpdatedAt:  time.Now().UTC().Format(time.RFC3339),
		ID:         id,
	})
}

func rowToSession(row db.GetSessionRow) *domain.Session {
	sess := &domain.Session{
		ID:           row.ID,
		AdminToken:   row.AdminToken,
		Status:       domain.SessionStatus(row.Status),
		Name:         row.Name,
		GameMode:     domain.GameMode(row.GameMode),
		Courts:       int(row.Courts),
		Points:       int(row.Points),
		CurrentRound: intPtr(row.CurrentRound),
		CreatedAt:    parseTime(row.CreatedAt),
		UpdatedAt:    parseTime(row.UpdatedAt),
		Players:      []domain.Player{},
	}

	if row.RoundsTotal.Valid {
		v := int(row.RoundsTotal.Int64)
		sess.RoundsTotal = &v
	}
	if row.CreatorPlayerID.Valid {
		sess.CreatorPlayerID = row.CreatorPlayerID.String
	}
	if row.CreatorUserID.Valid {
		sess.CreatorUserID = row.CreatorUserID.String
	}
	if row.ScheduledAt.Valid {
		sess.ScheduledAt = parseTimePtr(row.ScheduledAt.String)
	}
	if row.CourtDurationMinutes.Valid {
		v := int(row.CourtDurationMinutes.Int64)
		sess.CourtDurationMinutes = &v
	}
	if row.EndsAt.Valid {
		sess.EndsAt = parseTimePtr(row.EndsAt.String)
	}
	return sess
}

func intPtr(v int64) *int {
	if v == 0 {
		return nil
	}
	i := int(v)
	return &i
}

func parseTime(s string) time.Time {
	t, _ := time.Parse(time.RFC3339, s)
	return t
}

func parseTimePtr(s string) *time.Time {
	t, _ := time.Parse(time.RFC3339, s)
	return &t
}

type SessionPatch struct {
	Name        string
	GameMode    string
	Courts      int
	Points      int
	RoundsTotal *int
	ScheduledAt *time.Time
}

func (s *Store) UpdateSessionConfig(id string, p SessionPatch) error {
	var roundsTotalVal sql.NullInt64
	if p.RoundsTotal != nil {
		roundsTotalVal = sql.NullInt64{Int64: int64(*p.RoundsTotal), Valid: true}
	}
	var scheduledAtStr sql.NullString
	if p.ScheduledAt != nil {
		scheduledAtStr = sql.NullString{String: p.ScheduledAt.UTC().Format(time.RFC3339), Valid: true}
	}
	return s.queries.UpdateSessionConfig(context.Background(), db.UpdateSessionConfigParams{
		Name:        p.Name,
		GameMode:    p.GameMode,
		Courts:      int64(p.Courts),
		Points:      int64(p.Points),
		RoundsTotal: roundsTotalVal,
		ScheduledAt: scheduledAtStr,
		UpdatedAt:   time.Now().UTC().Format(time.RFC3339),
		ID:          id,
	})
}

func (s *Store) DeleteSession(id string) error {
	// Delete in dependency order due to foreign keys.
	s.queries.DeleteBench(context.Background(), id)
	s.queries.DeleteMatches(context.Background(), id)
	s.queries.DeleteRounds(context.Background(), id)
	s.queries.DeletePlayers(context.Background(), id)
	return s.queries.DeleteSession(context.Background(), id)
}
