package store

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/fabianthorsen/nottennis/internal/domain"
)

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
		PasswordHash: string(hash),
		CreatedAt:    time.Now().UTC(),
	}

	_, err = s.db.Exec(
		`INSERT INTO users (id, email, display_name, password_hash, created_at)
		 VALUES (?, ?, ?, ?, ?)`,
		user.ID, user.Email, user.DisplayName, user.PasswordHash,
		user.CreatedAt.Format(time.RFC3339),
	)
	if err != nil {
		if isUniqueConstraint(err, "email") {
			return nil, ErrEmailTaken
		}
		return nil, err
	}
	return user, nil
}

func (s *Store) GetUserByEmail(email string) (*domain.User, error) {
	return scanUser(s.db.QueryRow(
		`SELECT id, email, display_name, password_hash, created_at FROM users WHERE email = ?`, email,
	))
}

func (s *Store) GetUserByID(id string) (*domain.User, error) {
	return scanUser(s.db.QueryRow(
		`SELECT id, email, display_name, password_hash, created_at FROM users WHERE id = ?`, id,
	))
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
	_, err := s.db.Exec(
		`INSERT INTO auth_tokens (token, user_id, created_at) VALUES (?, ?, ?)`,
		token, userID, time.Now().UTC().Format(time.RFC3339),
	)
	return token, err
}

func (s *Store) GetUserByToken(token string) (*domain.User, error) {
	var userID string
	err := s.db.QueryRow(
		`SELECT user_id FROM auth_tokens WHERE token = ?`, token,
	).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return s.GetUserByID(userID)
}

func (s *Store) DeleteAuthToken(token string) error {
	_, err := s.db.Exec(`DELETE FROM auth_tokens WHERE token = ?`, token)
	return err
}

func (s *Store) GetCareerStats(userID string) (*domain.CareerStats, error) {
	var stats domain.CareerStats
	err := s.db.QueryRow(`
		SELECT
			COUNT(DISTINCT p.session_id) AS tournaments,
			COUNT(m.id) AS games_played,
			COALESCE(SUM(
				CASE
					WHEN (m.p1 = p.id OR m.p2 = p.id) AND m.score_a > m.score_b THEN 1
					WHEN (m.p3 = p.id OR m.p4 = p.id) AND m.score_b > m.score_a THEN 1
					ELSE 0
				END
			), 0) AS wins,
			COALESCE(SUM(
				CASE WHEN m.score_a = m.score_b THEN 1 ELSE 0 END
			), 0) AS draws,
			COALESCE(SUM(
				CASE
					WHEN m.p1 = p.id OR m.p2 = p.id THEN m.score_a
					WHEN m.p3 = p.id OR m.p4 = p.id THEN m.score_b
					ELSE 0
				END
			), 0) AS total_points
		FROM players p
		LEFT JOIN rounds r ON r.session_id = p.session_id
		LEFT JOIN matches m ON m.round_id = r.id
			AND (m.p1 = p.id OR m.p2 = p.id OR m.p3 = p.id OR m.p4 = p.id)
			AND m.score_a IS NOT NULL
		WHERE p.user_id = ? AND p.active = 1`,
		userID,
	).Scan(&stats.Tournaments, &stats.GamesPlayed, &stats.Wins, &stats.Draws, &stats.TotalPoints)
	if err != nil {
		return nil, err
	}
	stats.Losses = stats.GamesPlayed - stats.Wins - stats.Draws
	return &stats, nil
}

func scanUser(row *sql.Row) (*domain.User, error) {
	var u domain.User
	var createdAt string
	err := row.Scan(&u.ID, &u.Email, &u.DisplayName, &u.PasswordHash, &createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	u.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	return &u, nil
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
