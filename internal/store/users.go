package store

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/fabianthorsen/nottennis/internal/domain"
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

func (s *Store) GetCareerStats(userID string) (*domain.AmericanoCareerStats, error) {
	var stats domain.AmericanoCareerStats
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
		JOIN sessions s ON s.id = p.session_id AND s.status = 'complete' AND s.game_mode = 'americano'
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

// GetTennisCareerStats returns wins/losses/tournaments for tennis (2v2) sessions.
func (s *Store) GetTennisCareerStats(userID string) (*domain.TennisCareerStats, error) {
	var stats domain.TennisCareerStats
	err := s.db.QueryRow(`
		SELECT
			COUNT(DISTINCT p.session_id) AS tournaments,
			COALESCE(SUM(
				CASE WHEN json_extract(tm.state, '$.winner') = tt.team THEN 1 ELSE 0 END
			), 0) AS wins,
			COALESCE(SUM(
				CASE WHEN json_extract(tm.state, '$.winner') != '' AND json_extract(tm.state, '$.winner') != tt.team THEN 1 ELSE 0 END
			), 0) AS losses
		FROM players p
		JOIN sessions s ON s.id = p.session_id AND s.status = 'complete' AND s.game_mode = 'tennis'
		JOIN tennis_teams tt ON tt.session_id = p.session_id AND tt.player_id = p.id
		JOIN tennis_matches tm ON tm.session_id = p.session_id
		WHERE p.user_id = ? AND p.active = 1`,
		userID,
	).Scan(&stats.Tournaments, &stats.Wins, &stats.Losses)
	if err != nil {
		return nil, err
	}
	return &stats, nil
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
	s.db.Exec(`DELETE FROM password_reset_tokens WHERE user_id = ?`, user.ID)

	_, err = s.db.Exec(
		`INSERT INTO password_reset_tokens (token_hash, user_id, expires_at) VALUES (?, ?, ?)`,
		tokenHash, user.ID, expiresAt,
	)
	return raw, err
}

// RedeemPasswordResetToken validates the raw token and updates the user's password.
func (s *Store) RedeemPasswordResetToken(rawToken, newPassword string) error {
	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])

	var userID, expiresAtStr string
	err := s.db.QueryRow(
		`SELECT user_id, expires_at FROM password_reset_tokens WHERE token_hash = ?`, tokenHash,
	).Scan(&userID, &expiresAtStr)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrInvalidOrExpiredToken
	}
	if err != nil {
		return err
	}

	expiresAt, _ := time.Parse(time.RFC3339, expiresAtStr)
	if time.Now().UTC().After(expiresAt) {
		s.db.Exec(`DELETE FROM password_reset_tokens WHERE token_hash = ?`, tokenHash)
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

	if _, err := tx.Exec(`UPDATE users SET password_hash = ? WHERE id = ?`, string(newHash), userID); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM password_reset_tokens WHERE token_hash = ?`, tokenHash); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) GetTournamentHistory(userID string) ([]domain.TournamentHistoryEntry, error) {
	rows, err := s.db.Query(`
		WITH player_stats AS (
			SELECT
				p.id AS player_id,
				p.session_id,
				p.user_id,
				COALESCE(SUM(
					CASE
						WHEN (m.p1 = p.id OR m.p2 = p.id) AND m.score_a > m.score_b THEN 1
						WHEN (m.p3 = p.id OR m.p4 = p.id) AND m.score_b > m.score_a THEN 1
						ELSE 0
					END
				), 0) AS wins,
				COALESCE(SUM(
					CASE
						WHEN m.p1 = p.id OR m.p2 = p.id THEN m.score_a
						WHEN m.p3 = p.id OR m.p4 = p.id THEN m.score_b
						ELSE 0
					END
				), 0) AS points,
				COUNT(m.id) AS games_played
			FROM players p
			LEFT JOIN rounds r ON r.session_id = p.session_id
			LEFT JOIN matches m ON m.round_id = r.id
				AND (m.p1 = p.id OR m.p2 = p.id OR m.p3 = p.id OR m.p4 = p.id)
				AND m.score_a IS NOT NULL
			WHERE p.active = 1
			GROUP BY p.id
		),
		ranked AS (
			SELECT
				ps.*,
				RANK() OVER (PARTITION BY ps.session_id ORDER BY ps.points DESC, ps.wins DESC) AS rank
			FROM player_stats ps
		)
		SELECT
			s.id,
			COALESCE(NULLIF(s.name, ''), 'NotTennis'),
			s.status,
			s.created_at,
			rk.rank,
			rk.points,
			rk.games_played
		FROM ranked rk
		JOIN sessions s ON s.id = rk.session_id
		WHERE rk.user_id = ? AND s.status = 'complete'
		ORDER BY s.created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []domain.TournamentHistoryEntry
	for rows.Next() {
		var e domain.TournamentHistoryEntry
		if err := rows.Scan(&e.SessionID, &e.Name, &e.Status, &e.PlayedAt, &e.Rank, &e.Points, &e.GamesPlayed); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	if entries == nil {
		entries = []domain.TournamentHistoryEntry{}
	}
	return entries, rows.Err()
}

func (s *Store) GetUpcomingTournaments(userID string) ([]domain.UpcomingEntry, error) {
	rows, err := s.db.Query(`
		SELECT
			s.id,
			COALESCE(NULLIF(s.name, ''), 'NotTennis'),
			s.status,
			s.game_mode,
			s.courts,
			COUNT(p2.id) AS player_count,
			s.scheduled_at
		FROM players p
		JOIN sessions s ON s.id = p.session_id
		LEFT JOIN players p2 ON p2.session_id = s.id AND p2.active = 1
		WHERE p.user_id = ? AND p.active = 1 AND s.status IN ('lobby', 'active')
		GROUP BY s.id
		ORDER BY s.status DESC, COALESCE(s.scheduled_at, s.created_at) ASC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []domain.UpcomingEntry
	for rows.Next() {
		var e domain.UpcomingEntry
		var scheduledAt *string
		if err := rows.Scan(&e.SessionID, &e.Name, &e.Status, &e.GameMode, &e.Courts, &e.PlayerCount, &scheduledAt); err != nil {
			return nil, err
		}
		if scheduledAt != nil {
			t, err := time.Parse(time.RFC3339, *scheduledAt)
			if err == nil {
				e.ScheduledAt = &t
			}
		}
		entries = append(entries, e)
	}
	if entries == nil {
		entries = []domain.UpcomingEntry{}
	}
	return entries, rows.Err()
}

func (s *Store) DeleteUser(userID string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM auth_tokens WHERE user_id = ?`, userID); err != nil {
		return err
	}
	if _, err := tx.Exec(`UPDATE players SET user_id = NULL WHERE user_id = ?`, userID); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM users WHERE id = ?`, userID); err != nil {
		return err
	}
	return tx.Commit()
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
