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

func (s *Store) CreateUser(email, displayName, password string, selfRating int) (*domain.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	rating := selfRating
	user := &domain.User{
		ID:           newUserID(),
		Email:        email,
		DisplayName:  displayName,
		AvatarIcon:   randomAvatarIcon(),
		AvatarColor:  randomAvatarColor(),
		PasswordHash: string(hash),
		SelfRating:   &rating,
		CreatedAt:    time.Now().UTC(),
	}

	err = s.queries.CreateUser(context.Background(), db.CreateUserParams{
		ID:           user.ID,
		Email:        user.Email,
		DisplayName:  user.DisplayName,
		AvatarIcon:   user.AvatarIcon,
		AvatarColor:  user.AvatarColor,
		PasswordHash: user.PasswordHash,
		SelfRating:   nullInt64FromPtr(user.SelfRating),
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

// authTokenTTL is how long an auth token stays valid without use. Expiry is
// sliding (extended on use, see GetUserByToken), so an actively-used token never
// expires; an idle one dies after this window. authTokenRefreshInterval throttles
// how often the sliding expiry is rewritten, so a busy client (SSE/polling)
// doesn't trigger a DB write on every request (#240).
const (
	authTokenTTL             = 30 * 24 * time.Hour
	authTokenRefreshInterval = 24 * time.Hour
)

// hashToken derives the at-rest identifier for a bearer/reset token. Only the
// hash is stored, so a database leak never exposes usable tokens.
func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func (s *Store) CreateAuthToken(userID string) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	raw := hex.EncodeToString(b)
	now := time.Now().UTC()
	err := s.queries.CreateAuthToken(context.Background(), db.CreateAuthTokenParams{
		TokenHash: hashToken(raw),
		UserID:    userID,
		CreatedAt: now.Format(time.RFC3339),
		ExpiresAt: now.Add(authTokenTTL).Format(time.RFC3339),
	})
	return raw, err
}

func (s *Store) GetUserByToken(token string) (*domain.User, error) {
	tokenHash := hashToken(token)
	row, err := s.queries.GetAuthTokenByHash(context.Background(), tokenHash)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	expiresAt, err := time.Parse(time.RFC3339, row.ExpiresAt)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	if now.After(expiresAt) {
		// Expired: drop it so it can't be reused and reads as an unknown token.
		s.queries.DeleteAuthToken(context.Background(), tokenHash) //nolint:errcheck
		return nil, ErrNotFound
	}

	// Sliding expiry: extend on use, but only once the token has drifted more
	// than authTokenRefreshInterval from a full TTL, to avoid a write per request.
	if time.Until(expiresAt) < authTokenTTL-authTokenRefreshInterval {
		s.queries.UpdateAuthTokenExpiry(context.Background(), db.UpdateAuthTokenExpiryParams{ //nolint:errcheck
			ExpiresAt: now.Add(authTokenTTL).Format(time.RFC3339),
			TokenHash: tokenHash,
		})
	}

	return s.GetUserByID(row.UserID)
}

func (s *Store) DeleteAuthToken(token string) error {
	return s.queries.DeleteAuthToken(context.Background(), hashToken(token))
}

// GetCareerSummary returns the cross-mode profile headline (Point Win %, winrate,
// games) for a user, computed on read from every fully-scored Match across both
// Game Modes. See domain.CareerSummary and ADR 0007.
func (s *Store) GetCareerSummary(userID string) (*domain.CareerSummary, error) {
	row, err := s.queries.GetCareerSummary(context.Background(), sql.NullString{String: userID, Valid: true})
	if err != nil {
		return nil, err
	}
	summary := &domain.CareerSummary{
		Games:       int(row.Games),
		PointWinPct: row.PointShare * 100,
	}
	if summary.Games > 0 {
		summary.Winrate = float64(row.Wins) / float64(summary.Games) * 100
	}

	// Placement stats (titles / podiums / best / average finish) are not a SQL
	// aggregate: the finishing rank comes from the leaderboard tiebreaker chain,
	// so we rank each done Session in Go. Ranks compare across scoring models, so
	// these blend both Game Modes (unlike point-share).
	placement, err := s.getPlacementStats(userID)
	if err != nil {
		return nil, err
	}
	summary.Titles = placement.titles
	summary.Podiums = placement.podiums
	summary.BestFinish = placement.bestFinish
	if placement.count > 0 {
		summary.AverageFinish = float64(placement.rankSum) / float64(placement.count)
	}
	return summary, nil
}

// GetModeStats returns the per-Game-Mode career aggregates behind the Career
// Stats page (ADR 0007), computed on read. Both Game Modes are always returned
// in canonical order (Americano, then Mexicano); a mode the user has no history
// in comes back zero-valued so the UI can render its "no games yet" state.
func (s *Store) GetModeStats(userID string) ([]domain.ModeStats, error) {
	rows, err := s.queries.GetModeStats(context.Background(), sql.NullString{String: userID, Valid: true})
	if err != nil {
		return nil, err
	}
	byMode := make(map[domain.GameMode]domain.ModeStats, len(rows))
	for _, row := range rows {
		mode := domain.GameMode(row.Mode)
		ms := domain.ModeStats{
			Mode:           mode,
			Tournaments:    int(row.Tournaments),
			Games:          int(row.Games),
			Wins:           int(row.Wins),
			Draws:          int(row.Draws),
			TotalPoints:    int(row.TotalPoints),
			PointsConceded: int(row.PointsConceded),
			PointWinPct:    row.PointShare * 100,
		}
		ms.Losses = ms.Games - ms.Wins - ms.Draws
		ms.NetPoints = ms.TotalPoints - ms.PointsConceded
		byMode[mode] = ms
	}
	modes := domain.ModeAmericano.Values()
	out := make([]domain.ModeStats, 0, len(modes))
	for _, mode := range modes {
		if ms, ok := byMode[mode]; ok {
			out = append(out, ms)
		} else {
			out = append(out, domain.ModeStats{Mode: mode})
		}
	}
	return out, nil
}

// GetMatchResultsSeries returns the per-Match results series behind the Career
// Stats page's recent-form curve (ADR 0007): one row per fully-scored Match the
// user played in a done Session, oldest-first (by Session date, then round, then
// court). Points/conceded and the mode/date come straight from the query; the
// win/draw/loss outcome is the sign of the point differential. See domain.MatchResult
// and ADR 0007 for why the series is per-Match.
func (s *Store) GetMatchResultsSeries(userID string) ([]domain.MatchResult, error) {
	rows, err := s.queries.GetMatchResultsSeries(context.Background(), sql.NullString{String: userID, Valid: true})
	if err != nil {
		return nil, err
	}

	series := make([]domain.MatchResult, 0, len(rows))
	for _, row := range rows {
		points := int(row.Points)
		conceded := int(row.Conceded)
		result := domain.MatchResultDraw
		switch {
		case points > conceded:
			result = domain.MatchResultWin
		case points < conceded:
			result = domain.MatchResultLoss
		}
		series = append(series, domain.MatchResult{
			MatchID:  row.MatchID,
			Mode:     domain.GameMode(row.Mode),
			Date:     row.Date,
			Points:   points,
			Conceded: conceded,
			Result:   result,
		})
	}
	return series, nil
}

// placementStats accumulates the user's finishing ranks across their done
// Sessions: titles (rank 1), podiums (rank ≤ 3), the best (lowest) rank, and the
// running sum/count behind average finish.
type placementStats struct {
	titles     int
	podiums    int
	bestFinish int
	rankSum    int
	count      int
}

// getPlacementStats computes the cross-mode placement aggregate from the user's
// finishing rank in each done Session they played a scored match in, reusing the
// same ranked-Sessions walk as the tournament history timeline. Sessions the
// user has no scored Match in (ended early before finishing a game) and Sessions
// with no ranked standing (rank 0) are skipped rather than counted as finishes.
// Both Game Modes fold together: a finishing rank is comparable across scoring
// models in a way point-share is not.
func (s *Store) getPlacementStats(userID string) (placementStats, error) {
	ranked, err := s.rankedTournaments(userID)
	if err != nil {
		return placementStats{}, err
	}
	var p placementStats
	for _, rt := range ranked {
		rank := rt.entry.Rank
		if !rt.scored || rank == 0 {
			continue
		}
		if rank == 1 {
			p.titles++
		}
		if rank <= 3 {
			p.podiums++
		}
		if p.bestFinish == 0 || rank < p.bestFinish {
			p.bestFinish = rank
		}
		p.rankSum += rank
		p.count++
	}
	return p, nil
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
	tokenHash := hashToken(raw)
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
	tokenHash := hashToken(rawToken)

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

// rankedTournament is one of the user's done Sessions with their finishing rank
// resolved from the leaderboard and whether they played a fully-scored Match in
// it. It backs both the tournament history timeline and the placement stats, so
// each done Session is ranked exactly once per read.
type rankedTournament struct {
	entry  domain.TournamentHistoryEntry
	scored bool
}

// rankedTournaments walks the user's done Sessions (newest first) and resolves
// each one's finishing rank/points/games from GetLeaderboard — the same sort +
// tiebreaker chain the rest of the app uses. Leaderboard errors on a single
// Session are tolerated (that Session simply carries a zero rank), matching the
// historical behaviour of the tournament history endpoint.
func (s *Store) rankedTournaments(userID string) ([]rankedTournament, error) {
	sessions, err := s.queries.GetTournamentHistorySessions(context.Background(), sql.NullString{String: userID, Valid: true})
	if err != nil {
		return nil, err
	}
	out := make([]rankedTournament, 0, len(sessions))
	for _, sess := range sessions {
		rt := rankedTournament{
			entry: domain.TournamentHistoryEntry{
				SessionID:  sess.ID,
				Name:       sess.Name,
				Status:     sess.Status,
				PlayedAt:   sess.CreatedAt,
				EndedEarly: sess.EndedEarly == 1,
			},
			scored: sess.Scored == 1,
		}
		standings, err := s.GetLeaderboard(sess.ID)
		if err == nil {
			for _, st := range standings {
				if st.UserID != nil && *st.UserID == userID {
					rt.entry.Rank = st.Rank
					rt.entry.Points = st.Points
					rt.entry.GamesPlayed = st.GamesPlayed
					break
				}
			}
		}
		out = append(out, rt)
	}
	return out, nil
}

func (s *Store) GetTournamentHistory(userID string) ([]domain.TournamentHistoryEntry, error) {
	ranked, err := s.rankedTournaments(userID)
	if err != nil {
		return nil, err
	}
	entries := make([]domain.TournamentHistoryEntry, 0, len(ranked))
	for _, rt := range ranked {
		entries = append(entries, rt.entry)
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
			GameMode:    domain.GameMode(row.GameMode),
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
		SelfRating:   intPtrFromNull(row.SelfRating),
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
		SelfRating:   intPtrFromNull(row.SelfRating),
	}
	u.CreatedAt, _ = time.Parse(time.RFC3339, row.CreatedAt)
	return u
}

// UpdateSelfRating sets the authenticated User's default self_rating. Per ADR
// 0006 this only seeds future sessions; it never rewrites Player.rating in
// sessions already joined.
func (s *Store) UpdateSelfRating(userID string, rating int) error {
	return s.queries.UpdateUserSelfRating(context.Background(), db.UpdateUserSelfRatingParams{
		SelfRating: sql.NullInt64{Int64: int64(rating), Valid: true},
		ID:         userID,
	})
}

func nullInt64FromPtr(p *int) sql.NullInt64 {
	if p == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: int64(*p), Valid: true}
}

func intPtrFromNull(n sql.NullInt64) *int {
	if !n.Valid {
		return nil
	}
	v := int(n.Int64)
	return &v
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
