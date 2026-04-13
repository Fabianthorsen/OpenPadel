package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/fabianthorsen/openpadel/internal/domain"
	"github.com/fabianthorsen/openpadel/internal/store/db"
)

func (s *Store) CreatePlayer(sessionID, name, userID string) (*domain.Player, error) {
	now := time.Now().UTC()
	icon := "Bot"
	color := "slate" // guests get grey Bot icon; overridden below for registered users
	if userID != "" {
		// Use the user's own avatar so their profile icon carries into sessions.
		avatar, err := s.queries.GetUserAvatarByUserID(context.Background(), userID)
		if err == nil {
			icon = avatar.AvatarIcon
			color = avatar.AvatarColor
		}
	}
	p := &domain.Player{
		ID:          newID(),
		SessionID:   sessionID,
		UserID:      userID,
		Name:        name,
		AvatarIcon:  icon,
		AvatarColor: color,
		Active:      true,
		JoinedAt:    now,
	}
	var uid sql.NullString
	if userID != "" {
		uid = sql.NullString{String: userID, Valid: true}
	}
	err := s.queries.CreatePlayer(context.Background(), db.CreatePlayerParams{
		ID:          p.ID,
		SessionID:   p.SessionID,
		UserID:      uid,
		Name:        p.Name,
		AvatarIcon:  p.AvatarIcon,
		AvatarColor: p.AvatarColor,
		JoinedAt:    p.JoinedAt.Format(time.RFC3339),
	})
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Store) GetPlayers(sessionID string) ([]domain.Player, error) {
	rows, err := s.queries.GetPlayersBySessionID(context.Background(), sessionID)
	if err != nil {
		return nil, err
	}
	players := make([]domain.Player, 0, len(rows))
	for _, row := range rows {
		p := domain.Player{
			ID:          row.ID,
			SessionID:   row.SessionID,
			UserID:      row.UserID,
			Name:        row.Name,
			AvatarIcon:  row.AvatarIcon,
			AvatarColor: row.AvatarColor,
			Active:      row.Active == 1,
			JoinedAt:    parseTime(row.JoinedAt),
		}
		players = append(players, p)
	}
	return players, nil
}

// GetCreatorName returns the name of the creator player for the given session,
// or an empty string if the creator hasn't joined yet.
func (s *Store) GetCreatorName(sessionID string) string {
	name, err := s.queries.GetCreatorName(context.Background(), sessionID)
	if err != nil {
		return ""
	}
	return name
}

func (s *Store) DeactivatePlayer(playerID string) error {
	res, err := s.db.Exec(`UPDATE players SET active = 0 WHERE id = ?`, playerID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}


