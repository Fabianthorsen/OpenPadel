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

func (s *Store) CreateSession(courts, points int, name, gameMode string, roundsTotal *int, scheduledAt *time.Time, courtDurationMinutes *int, totalDurationMinutes *int, bufferSeconds *int, intervalBetweenRoundsMin *int, creatorUserID string) (*domain.Session, error) {
	now := time.Now().UTC()
	if gameMode == "" {
		gameMode = "americano"
	}
	sess := &domain.Session{
		ID:                       newID(),
		AdminToken:               newAdminToken(),
		Status:                   domain.StatusLobby,
		Name:                     name,
		GameMode:                 gameMode,
		Courts:                   courts,
		Points:                   points,
		RoundsTotal:              roundsTotal,
		ScheduledAt:              scheduledAt,
		CourtDurationMinutes:     courtDurationMinutes,
		TotalDurationMinutes:     totalDurationMinutes,
		BufferSeconds:            bufferSeconds,
		IntervalBetweenRoundsMin: intervalBetweenRoundsMin,
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
	var totalDurationMinutesVal sql.NullInt64
	if totalDurationMinutes != nil {
		totalDurationMinutesVal = sql.NullInt64{Int64: int64(*totalDurationMinutes), Valid: true}
	}
	var bufferSecondsVal sql.NullInt64
	if bufferSeconds != nil {
		bufferSecondsVal = sql.NullInt64{Int64: int64(*bufferSeconds), Valid: true}
	}
	var intervalBetweenRoundsMinVal sql.NullInt64
	if intervalBetweenRoundsMin != nil {
		intervalBetweenRoundsMinVal = sql.NullInt64{Int64: int64(*intervalBetweenRoundsMin), Valid: true}
	}
	var creatorUserIDVal sql.NullString
	if creatorUserID != "" {
		creatorUserIDVal = sql.NullString{String: creatorUserID, Valid: true}
	}
	err := s.queries.CreateSession(context.Background(), db.CreateSessionParams{
		ID:                           sess.ID,
		AdminToken:                   sess.AdminToken,
		Status:                       string(sess.Status),
		Name:                         sess.Name,
		GameMode:                     sess.GameMode,
		SetsToWin:                    0,
		GamesPerSet:                  0,
		Courts:                       int64(sess.Courts),
		Points:                       int64(sess.Points),
		RoundsTotal:                  roundsTotalVal,
		ScheduledAt:                  scheduledAtStr,
		CourtDurationMinutes:         courtDurationMinutesVal,
		TotalDurationMinutes:         totalDurationMinutesVal,
		BufferSeconds:                bufferSecondsVal,
		IntervalBetweenRoundsMinutes: intervalBetweenRoundsMinVal,
		CreatorUserID:                creatorUserIDVal,
		CreatedAt:                    sess.CreatedAt.Format(time.RFC3339),
		UpdatedAt:                    sess.UpdatedAt.Format(time.RFC3339),
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
		Status:      string(domain.StatusActive),
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
		Status:    string(domain.StatusActive),
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
		Status:     string(domain.StatusComplete),
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
		GameMode:     row.GameMode,
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
	if row.TotalDurationMinutes.Valid {
		v := int(row.TotalDurationMinutes.Int64)
		sess.TotalDurationMinutes = &v
	}
	if row.BufferSeconds.Valid {
		v := int(row.BufferSeconds.Int64)
		sess.BufferSeconds = &v
	}
	if row.IntervalBetweenRoundsMinutes.Valid {
		v := int(row.IntervalBetweenRoundsMinutes.Int64)
		sess.IntervalBetweenRoundsMin = &v
	}
	if row.RoundDurationSeconds.Valid {
		v := int(row.RoundDurationSeconds.Int64)
		sess.RoundDurationSeconds = &v
	}
	if row.RoundStartedAt.Valid {
		sess.RoundStartedAt = parseTimePtr(row.RoundStartedAt.String)
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

func (s *Store) StartTimedAmericanoSession(id, status string, roundsTotal int, totalDurationMin, bufferSec, intervalBetweenRoundsMin, roundDurationSec *int, endsAt *time.Time) error {
	var roundsTotalVal sql.NullInt64
	if roundsTotal > 0 {
		roundsTotalVal = sql.NullInt64{Int64: int64(roundsTotal), Valid: true}
	}

	var totalDurationMinVal sql.NullInt64
	if totalDurationMin != nil {
		totalDurationMinVal = sql.NullInt64{Int64: int64(*totalDurationMin), Valid: true}
	}

	var bufferSecVal sql.NullInt64
	if bufferSec != nil {
		bufferSecVal = sql.NullInt64{Int64: int64(*bufferSec), Valid: true}
	}

	var intervalBetweenRoundsMinVal sql.NullInt64
	if intervalBetweenRoundsMin != nil {
		intervalBetweenRoundsMinVal = sql.NullInt64{Int64: int64(*intervalBetweenRoundsMin), Valid: true}
	}

	var roundDurationSecVal sql.NullInt64
	if roundDurationSec != nil {
		roundDurationSecVal = sql.NullInt64{Int64: int64(*roundDurationSec), Valid: true}
	}

	var endsAtStr sql.NullString
	if endsAt != nil {
		endsAtStr = sql.NullString{String: endsAt.UTC().Format(time.RFC3339), Valid: true}
	}

	return s.queries.StartTimedAmericanoSession(context.Background(), db.StartTimedAmericanoSessionParams{
		Status:                       status,
		RoundsTotal:                  roundsTotalVal,
		TotalDurationMinutes:         totalDurationMinVal,
		BufferSeconds:                bufferSecVal,
		IntervalBetweenRoundsMinutes: intervalBetweenRoundsMinVal,
		RoundDurationSeconds:         roundDurationSecVal,
		EndsAt:                       endsAtStr,
		UpdatedAt:                    time.Now().UTC().Format(time.RFC3339),
		ID:                           id,
	})
}

func (s *Store) SetRoundStartedAt(id string, roundStartedAt *time.Time) error {
	var roundStartedAtStr sql.NullString
	if roundStartedAt != nil {
		roundStartedAtStr = sql.NullString{String: roundStartedAt.UTC().Format(time.RFC3339), Valid: true}
	}

	return s.queries.SetRoundStartedAt(context.Background(), db.SetRoundStartedAtParams{
		RoundStartedAt: roundStartedAtStr,
		UpdatedAt:      time.Now().UTC().Format(time.RFC3339),
		ID:             id,
	})
}

func (s *Store) UpdateRoundDuration(id string, roundDurationSec *int) error {
	var roundDurationSecVal sql.NullInt64
	if roundDurationSec != nil {
		roundDurationSecVal = sql.NullInt64{Int64: int64(*roundDurationSec), Valid: true}
	}

	return s.queries.UpdateRoundDuration(context.Background(), db.UpdateRoundDurationParams{
		RoundDurationSeconds: roundDurationSecVal,
		UpdatedAt:            time.Now().UTC().Format(time.RFC3339),
		ID:                   id,
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
