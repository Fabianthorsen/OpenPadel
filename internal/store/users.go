package store

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/fabianthorsen/openpadel/internal/domain"
	"github.com/fabianthorsen/openpadel/internal/store/db"
)

var ErrInvalidOrExpiredToken = errors.New("invalid or expired token")

var ErrEmailTaken = errors.New("email already registered")
var ErrInvalidCredentials = errors.New("invalid email or password")

func (s *Store) CreateUser(email, displayName, password string) (*domain.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		ID:           newUserID(),
		Email:        email,
		DisplayName:  displayName,
		AvatarIcon:   randomAvatarIcon(),
		AvatarColor:  randomAvatarColor(),
		PasswordHash: string(hash),
		CreatedAt:    time.Now().UTC(),
	}

	err = s.queries.CreateUser(context.Background(), db.CreateUserParams{
		ID:           user.ID,
		Email:        user.Email,
		DisplayName:  user.DisplayName,
		AvatarIcon:   user.AvatarIcon,
		AvatarColor:  user.AvatarColor,
		PasswordHash: user.PasswordHash,
		CreatedAt:    user.CreatedAt.Format(time.RFC3339),
	})
	if err != nil {
		if isUniqueConstraint(err, "email") {
			return nil, ErrEmailTaken
		}
		return nil, err
	}
	return user, nil
}

func (s *Store) UpdateProfile(userID, displayName, avatarIcon, avatarColor string) (*domain.User, error) {
	err := s.queries.UpdateProfile(context.Background(), db.UpdateProfileParams{
		DisplayName: displayName,
		AvatarIcon:  avatarIcon,
		AvatarColor: avatarColor,
		ID:          userID,
	})
	if err != nil {
		return nil, err
	}
	// Sync avatar to all player records for this user so in-progress sessions pick it up.
	s.queries.UpdateProfileAvatarOnPlayers(context.Background(), db.UpdateProfileAvatarOnPlayersParams{
		AvatarIcon:  avatarIcon,
		AvatarColor: avatarColor,
		UserID:      sql.NullString{String: userID, Valid: true},
	})
	return s.GetUserByID(userID)
}

func (s *Store) GetUserByEmail(email string) (*domain.User, error) {
	row, err := s.queries.GetUserByEmail(context.Background(), email)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return rowToUserEmail(row), nil
}

func (s *Store) GetUserByID(id string) (*domain.User, error) {
	row, err := s.queries.GetUserByID(context.Background(), id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return rowToUserID(row), nil
}

func (s *Store) AuthenticateUser(email, password string) (*domain.User, error) {
	user, err := s.GetUserByEmail(email)
	if errors.Is(err, ErrNotFound) {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}
	return user, nil
}

func (s *Store) CreateAuthToken(userID string) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)
	err := s.queries.CreateAuthToken(context.Background(), db.CreateAuthTokenParams{
		Token:     token,
		UserID:    userID,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	})
	return token, err
}

func (s *Store) GetUserByToken(token string) (*domain.User, error) {
	userID, err := s.queries.GetUserIDByToken(context.Background(), token)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return s.GetUserByID(userID)
}

func (s *Store) DeleteAuthToken(token string) error {
	return s.queries.DeleteAuthToken(context.Background(), token)
}

func (s *Store) GetCareerStats(userID string) (*domain.AmericanoCareerStats, error) {
	row, err := s.queries.GetAmericanoCareerStats(context.Background(), sql.NullString{String: userID, Valid: true})
	if err != nil {
		return nil, err
	}
	stats := &domain.AmericanoCareerStats{
		Tournaments: int(row.Tournaments),
		GamesPlayed: int(row.GamesPlayed),
		Wins:        int(row.Wins),
		Draws:       int(row.Draws),
		TotalPoints: int(row.TotalPoints),
	}
	stats.Losses = stats.GamesPlayed - stats.Wins - stats.Draws
	return stats, nil
}

// CreatePasswordResetToken generates a secure token for the given email.
// Returns the raw token (to be emailed) and ErrNotFound if the email doesn't exist.
func (s *Store) CreatePasswordResetToken(email string) (rawToken string, err error) {
	user, err := s.GetUserByEmail(email)
	if err != nil {
		return "", err
	}

	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	raw := hex.EncodeToString(b)
	hash := sha256.Sum256([]byte(raw))
	tokenHash := hex.EncodeToString(hash[:])
	expiresAt := time.Now().UTC().Add(time.Hour).Format(time.RFC3339)

	// Delete any existing token for this user first
	s.queries.DeletePasswordResetTokensByUserID(context.Background(), user.ID)

	err = s.queries.CreatePasswordResetToken(context.Background(), db.CreatePasswordResetTokenParams{
		TokenHash: tokenHash,
		UserID:    user.ID,
		ExpiresAt: expiresAt,
	})
	return raw, err
}

// RedeemPasswordResetToken validates the raw token and updates the user's password.
func (s *Store) RedeemPasswordResetToken(rawToken, newPassword string) error {
	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])

	row, err := s.queries.GetPasswordResetToken(context.Background(), tokenHash)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrInvalidOrExpiredToken
	}
	if err != nil {
		return err
	}

	expiresAt, _ := time.Parse(time.RFC3339, row.ExpiresAt)
	if time.Now().UTC().After(expiresAt) {
		s.queries.DeletePasswordResetToken(context.Background(), tokenHash)
		return ErrInvalidOrExpiredToken
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	qtx := s.queries.WithTx(tx)
	if err := qtx.UpdateUserPassword(context.Background(), db.UpdateUserPasswordParams{
		PasswordHash: string(newHash),
		ID:           row.UserID,
	}); err != nil {
		return err
	}
	if err := qtx.DeletePasswordResetToken(context.Background(), tokenHash); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) GetTournamentHistory(userID string) ([]domain.TournamentHistoryEntry, error) {
	sessions, err := s.queries.GetTournamentHistorySessions(context.Background(), sql.NullString{String: userID, Valid: true})
	if err != nil {
		return nil, err
	}

	var entries []domain.TournamentHistoryEntry
	for _, sess := range sessions {
		e := domain.TournamentHistoryEntry{
			SessionID:  sess.ID,
			Name:       sess.Name,
			Status:     sess.Status,
			PlayedAt:   sess.CreatedAt,
			EndedEarly: sess.EndedEarly == 1,
		}

		// Compute rank/points/games from the leaderboard (reuses existing sort + tiebreaker logic).
		standings, err := s.GetLeaderboard(sess.ID)
		if err == nil {
			for _, st := range standings {
				if st.UserID != nil && *st.UserID == userID {
					e.Rank = st.Rank
					e.Points = st.Points
					e.GamesPlayed = st.GamesPlayed
					break
				}
			}
		}

		entries = append(entries, e)
	}
	if entries == nil {
		entries = []domain.TournamentHistoryEntry{}
	}
	return entries, nil
}

func (s *Store) GetUpcomingTournaments(userID string) ([]domain.UpcomingEntry, error) {
	rows, err := s.queries.GetUpcomingTournaments(context.Background(), sql.NullString{String: userID, Valid: true})
	if err != nil {
		return nil, err
	}

	var entries []domain.UpcomingEntry
	for _, row := range rows {
		e := domain.UpcomingEntry{
			SessionID:   row.ID,
			Name:        row.Name,
			Status:      row.Status,
			GameMode:    row.GameMode,
			Courts:      int(row.Courts),
			PlayerCount: int(row.PlayerCount),
		}
		// Handle scheduled_at which is nullable
		if row.ScheduledAt.Valid {
			t, err := time.Parse(time.RFC3339, row.ScheduledAt.String)
			if err == nil {
				e.ScheduledAt = &t
			}
		}
		entries = append(entries, e)
	}
	if entries == nil {
		entries = []domain.UpcomingEntry{}
	}
	return entries, nil
}

func (s *Store) DeleteUser(userID string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	qtx := s.queries.WithTx(tx)
	if err := qtx.DeleteAuthTokensByUserID(context.Background(), userID); err != nil {
		return err
	}
	if err := qtx.UpdatePlayerUserIDToNull(context.Background(), sql.NullString{String: userID, Valid: true}); err != nil {
		return err
	}
	if err := qtx.DeleteUser(context.Background(), userID); err != nil {
		return err
	}
	return tx.Commit()
}

func rowToUserEmail(row db.GetUserByEmailRow) *domain.User {
	u := &domain.User{
		ID:           row.ID,
		Email:        row.Email,
		DisplayName:  row.DisplayName,
		AvatarIcon:   row.AvatarIcon,
		AvatarColor:  row.AvatarColor,
		PasswordHash: row.PasswordHash,
	}
	u.CreatedAt, _ = time.Parse(time.RFC3339, row.CreatedAt)
	return u
}

func rowToUserID(row db.GetUserByIDRow) *domain.User {
	u := &domain.User{
		ID:           row.ID,
		Email:        row.Email,
		DisplayName:  row.DisplayName,
		AvatarIcon:   row.AvatarIcon,
		AvatarColor:  row.AvatarColor,
		PasswordHash: row.PasswordHash,
	}
	u.CreatedAt, _ = time.Parse(time.RFC3339, row.CreatedAt)
	return u
}

func newUserID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return "u" + hex.EncodeToString(b)
}

// isUniqueConstraint checks if a SQLite error is a UNIQUE constraint on a specific column.
func isUniqueConstraint(err error, column string) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return len(s) > 0 && containsAll(s, "UNIQUE constraint failed", column)
}

func containsAll(s string, subs ...string) bool {
	for _, sub := range subs {
		found := false
		for i := 0; i <= len(s)-len(sub); i++ {
			if s[i:i+len(sub)] == sub {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
