package store

import (
	"context"
	"crypto/subtle"
	"database/sql"
	"errors"
	"time"

	"github.com/fabianthorsen/openpadel/internal/domain"
	"github.com/fabianthorsen/openpadel/internal/store/db"
)

// boolToInt64 maps a Go bool to SQLite's integer boolean (0/1).
func boolToInt64(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

func (s *Store) CreatePlayer(sessionID, name, userID string, addedByAdmin bool) (*domain.Player, error) {
	now := time.Now().UTC()
	icon := "Bot"
	color := "slate" // guests get grey Bot icon; overridden below for registered users
	// Guests (no account) have no self_rating, so they start at the neutral
	// median; a registered User seeds their Player.rating from their self_rating.
	rating := domain.MedianRating
	if userID != "" {
		// Use the user's own avatar + self_rating so their profile carries into
		// the session. A pre-rating account with a null self_rating stays median.
		if user, err := s.GetUserByID(userID); err == nil {
			icon = user.AvatarIcon
			color = user.AvatarColor
			if user.SelfRating != nil {
				rating = domain.NormalizeRating(*user.SelfRating)
			}
		}
	}
	// Per-player self-removal secret: the raw token goes back to the joining
	// client, only its hash is stored (#241).
	rawToken := randString(32)
	p := &domain.Player{
		ID:           newID(),
		SessionID:    sessionID,
		UserID:       userID,
		Name:         name,
		AvatarIcon:   icon,
		AvatarColor:  color,
		Rating:       rating,
		AddedByAdmin: addedByAdmin,
		Active:       true,
		JoinedAt:     now,
		Token:        rawToken,
	}
	var uid sql.NullString
	if userID != "" {
		uid = sql.NullString{String: userID, Valid: true}
	}
	err := s.queries.CreatePlayer(context.Background(), db.CreatePlayerParams{
		ID:           p.ID,
		SessionID:    p.SessionID,
		UserID:       uid,
		Name:         p.Name,
		AvatarIcon:   p.AvatarIcon,
		AvatarColor:  p.AvatarColor,
		Rating:       int64(p.Rating),
		AddedByAdmin: boolToInt64(addedByAdmin),
		JoinedAt:     p.JoinedAt.Format(time.RFC3339),
		TokenHash:    hashToken(rawToken),
	})
	if err != nil {
		return nil, err
	}
	return p, nil
}

// VerifyPlayerToken reports whether rawToken matches the secret issued to the
// given player at join time. An empty stored hash (pre-#241 players) never
// matches, so those players fall back to admin-only removal.
func (s *Store) VerifyPlayerToken(playerID, rawToken string) (bool, error) {
	if rawToken == "" {
		return false, nil
	}
	stored, err := s.queries.GetPlayerTokenHash(context.Background(), playerID)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if stored == "" {
		return false, nil
	}
	return subtle.ConstantTimeCompare([]byte(stored), []byte(hashToken(rawToken))) == 1, nil
}

func (s *Store) GetPlayers(sessionID string) ([]domain.Player, error) {
	rows, err := s.queries.GetPlayersBySessionID(context.Background(), sessionID)
	if err != nil {
		return nil, err
	}
	players := make([]domain.Player, 0, len(rows))
	for _, row := range rows {
		p := domain.Player{
			ID:           row.ID,
			SessionID:    row.SessionID,
			UserID:       row.UserID,
			Name:         row.Name,
			AvatarIcon:   row.AvatarIcon,
			AvatarColor:  row.AvatarColor,
			Rating:       int(row.Rating),
			AddedByAdmin: row.AddedByAdmin == 1,
			Active:       row.Active == 1,
			JoinedAt:     parseTime(row.JoinedAt),
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

// UpdatePlayerRating sets a Player's per-session rating. This is a lobby-scoped
// snapshot and does not propagate back to the owning User's self_rating.
func (s *Store) UpdatePlayerRating(playerID string, rating int) error {
	return s.queries.UpdatePlayerRating(context.Background(), db.UpdatePlayerRatingParams{
		Rating: int64(rating),
		ID:     playerID,
	})
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
