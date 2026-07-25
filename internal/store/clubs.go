package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/fabianthorsen/openpadel/internal/domain"
	"github.com/fabianthorsen/openpadel/internal/store/db"
)

func (s *Store) CreateClub(name, description, avatarIcon, avatarColor, creatorUserID string) (*domain.Club, error) {
	clubID := newID()
	joinCode := randString(32)
	now := time.Now().UTC().Format(time.RFC3339)

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	qtx := s.queries.WithTx(tx)

	if err := qtx.CreateClub(context.Background(), db.CreateClubParams{
		ID:          clubID,
		Name:        name,
		Description: description,
		AvatarIcon:  avatarIcon,
		AvatarColor: avatarColor,
		JoinCode:    joinCode,
		CreatedBy:   creatorUserID,
		CreatedAt:   now,
	}); err != nil {
		return nil, err
	}

	if err := qtx.InsertClubMember(context.Background(), db.InsertClubMemberParams{
		ClubID:   clubID,
		UserID:   creatorUserID,
		Role:     "admin",
		JoinedAt: now,
	}); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	club, err := s.getClubByID(clubID)
	if err != nil {
		return nil, err
	}

	return club, nil
}

func (s *Store) GetClub(clubID string) (*domain.Club, error) {
	return s.getClubByID(clubID)
}

func (s *Store) GetClubByJoinCode(joinCode string) (*domain.Club, error) {
	row, err := s.queries.GetClubByJoinCode(context.Background(), joinCode)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &domain.Club{
		ID:          row.ID,
		Name:        row.Name,
		Description: row.Description,
		AvatarIcon:  row.AvatarIcon,
		AvatarColor: row.AvatarColor,
		JoinCode:    row.JoinCode,
		CreatedBy:   row.CreatedBy,
		CreatedAt:   parseTime(row.CreatedAt),
	}, nil
}

func (s *Store) GetUserClubs(userID string) ([]domain.Club, error) {
	rows, err := s.queries.GetUserClubs(context.Background(), userID)
	if err != nil {
		return nil, err
	}

	var clubs []domain.Club
	for _, row := range rows {
		clubs = append(clubs, domain.Club{
			ID:          row.ID,
			Name:        row.Name,
			Description: row.Description,
			AvatarIcon:  row.AvatarIcon,
			AvatarColor: row.AvatarColor,
			JoinCode:    row.JoinCode,
			CreatedBy:   row.CreatedBy,
			CreatedAt:   parseTime(row.CreatedAt),
		})
	}

	if clubs == nil {
		clubs = []domain.Club{}
	}

	return clubs, nil
}

func (s *Store) GetClubMembers(clubID string) ([]domain.ClubMember, error) {
	rows, err := s.queries.GetClubMembers(context.Background(), clubID)
	if err != nil {
		return nil, err
	}

	var members []domain.ClubMember
	for _, row := range rows {
		members = append(members, domain.ClubMember{
			UserID:      row.UserID,
			DisplayName: row.DisplayName,
			Role:        row.Role,
			AvatarIcon:  row.AvatarIcon,
			AvatarColor: row.AvatarColor,
			JoinedAt:    parseTime(row.JoinedAt),
		})
	}

	if members == nil {
		members = []domain.ClubMember{}
	}

	return members, nil
}

func (s *Store) GetClubMemberCount(clubID string) (int, error) {
	count, err := s.queries.GetClubMemberCount(context.Background(), clubID)
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (s *Store) GetClubMember(clubID, userID string) (*domain.ClubMember, error) {
	row, err := s.queries.GetClubMember(context.Background(), db.GetClubMemberParams{
		ClubID: clubID,
		UserID: userID,
	})
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &domain.ClubMember{
		UserID:   row.UserID,
		Role:     row.Role,
		JoinedAt: parseTime(row.JoinedAt),
	}, nil
}

func (s *Store) UpdateClub(clubID, name, description, avatarIcon, avatarColor string) error {
	return s.queries.UpdateClub(context.Background(), db.UpdateClubParams{
		ID:          clubID,
		Name:        name,
		Description: description,
		AvatarIcon:  avatarIcon,
		AvatarColor: avatarColor,
	})
}

func (s *Store) UpdateClubJoinCode(clubID string) (string, error) {
	newCode := randString(32)
	err := s.queries.UpdateJoinCode(context.Background(), db.UpdateJoinCodeParams{
		ID:       clubID,
		JoinCode: newCode,
	})
	if err != nil {
		return "", err
	}
	return newCode, nil
}

// DeleteClub hard-deletes a Club and lets the database unwind its dependents via
// the declared ON DELETE actions (now that FK enforcement is on, see #249):
// club_members and club_invites CASCADE away, while past Sessions are detached
// (sessions.club_id ON DELETE SET NULL) so they survive as ordinary Sessions.
func (s *Store) DeleteClub(clubID string) error {
	return s.queries.DeleteClub(context.Background(), clubID)
}

func (s *Store) JoinClub(clubID, userID string) error {
	now := time.Now().UTC().Format(time.RFC3339)

	existingMember, err := s.GetClubMember(clubID, userID)
	if existingMember != nil {
		return ErrAlreadyMember
	}
	if err != ErrNotFound && err != nil {
		return err
	}

	return s.queries.InsertClubMember(context.Background(), db.InsertClubMemberParams{
		ClubID:   clubID,
		UserID:   userID,
		Role:     "member",
		JoinedAt: now,
	})
}

// RemoveClubMember removes a member from the Club. Removing the sole Admin is
// refused with ErrLastAdmin so the Club never drops to zero Admins; a non-member
// target yields ErrNotFound.
func (s *Store) RemoveClubMember(clubID, userID string) error {
	member, err := s.GetClubMember(clubID, userID)
	if err != nil {
		return err
	}
	// Leaving would drop the last Admin.
	if err := s.guardLastAdmin(clubID, member.Role, "member"); err != nil {
		return err
	}
	return s.queries.DeleteClubMember(context.Background(), db.DeleteClubMemberParams{
		ClubID: clubID,
		UserID: userID,
	})
}

// UpdateClubMemberRole promotes or demotes a member. Demoting the sole Admin is
// refused with ErrLastAdmin; a non-member target yields ErrNotFound.
func (s *Store) UpdateClubMemberRole(clubID, userID, role string) error {
	member, err := s.GetClubMember(clubID, userID)
	if err != nil {
		return err
	}
	if err := s.guardLastAdmin(clubID, member.Role, role); err != nil {
		return err
	}
	return s.queries.UpdateClubMemberRole(context.Background(), db.UpdateClubMemberRoleParams{
		ClubID: clubID,
		UserID: userID,
		Role:   role,
	})
}

// guardLastAdmin returns ErrLastAdmin when a member's role transition from
// currentRole to newRole would strip the Club of its final Admin — i.e. an Admin
// being demoted or removed (newRole != "admin") while no other Admin remains.
func (s *Store) guardLastAdmin(clubID, currentRole, newRole string) error {
	if currentRole != "admin" || newRole == "admin" {
		return nil
	}
	count, err := s.GetClubAdminCount(clubID)
	if err != nil {
		return err
	}
	if count <= 1 {
		return ErrLastAdmin
	}
	return nil
}

func (s *Store) GetClubAdminCount(clubID string) (int, error) {
	count, err := s.queries.GetClubAdminCount(context.Background(), clubID)
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

// GetClubEvents returns the Club's upcoming events — Sessions owned by the Club
// that are still in lobby or playing — newest-relevant first (playing before
// lobby, then by scheduled/created time). Done Sessions are excluded; this is the
// "what's coming up" feed for the Club home, not a history.
func (s *Store) GetClubEvents(clubID string) ([]domain.UpcomingEntry, error) {
	rows, err := s.queries.GetClubEvents(context.Background(), sql.NullString{String: clubID, Valid: true})
	if err != nil {
		return nil, err
	}
	events := make([]domain.UpcomingEntry, 0, len(rows))
	for _, row := range rows {
		e := domain.UpcomingEntry{
			SessionID:   row.ID,
			Name:        row.Name,
			Status:      row.Status,
			GameMode:    domain.GameMode(row.GameMode),
			Courts:      int(row.Courts),
			PlayerCount: int(row.PlayerCount),
		}
		if row.ScheduledAt.Valid {
			e.ScheduledAt = parseTimePtr(row.ScheduledAt.String)
		}
		events = append(events, e)
	}
	return events, nil
}

// GetClubLeaderboard selects the Club's qualifying games and ranks them via the
// pure domain function. Qualifying rows are: this Club's completed ('done')
// Sessions, scored Matches only (score_a IS NOT NULL), within the rolling window
// by COALESCE(scheduled_at, created_at). Restricting to done Sessions keeps the
// board's game counts reconciled with career stats (which are also done-only) —
// a live event contributes once it's finished. One row is emitted per registered member
// per Match, from that member's perspective (team vs opponent points); Guests
// (players.user_id NULL) fall out at the users join, so they contribute
// partner/opponent points to the team totals but never a leaderboard row of
// their own. Nothing is materialized — the board recomputes on every read, so a
// newly scored Match shows up on the next call.
func (s *Store) GetClubLeaderboard(clubID string) (domain.ClubLeaderboard, error) {
	cutoff := time.Now().UTC().Add(-domain.ClubLeaderboardWindow).Format(time.RFC3339)
	rows, err := s.db.Query(`
		SELECT
			pl.user_id,
			u.display_name,
			u.avatar_icon,
			u.avatar_color,
			CASE WHEN pl.id IN (m.p1, m.p2) THEN m.score_a ELSE m.score_b END AS team_points,
			CASE WHEN pl.id IN (m.p1, m.p2) THEN m.score_b ELSE m.score_a END AS opp_points,
			s.points AS points_target,
			COALESCE(s.scheduled_at, s.created_at) AS played_at
		FROM matches m
		JOIN rounds r ON r.id = m.round_id
		JOIN sessions s ON s.id = r.session_id
		JOIN players pl ON pl.id IN (m.p1, m.p2, m.p3, m.p4)
		JOIN users u ON u.id = pl.user_id
		WHERE s.club_id = ?
		  AND s.status = 'done'
		  AND m.score_a IS NOT NULL
		  AND COALESCE(s.scheduled_at, s.created_at) >= ?`,
		clubID, cutoff,
	)
	if err != nil {
		return domain.ClubLeaderboard{}, err
	}
	defer func() { _ = rows.Close() }()

	var games []domain.ClubGame
	for rows.Next() {
		var g domain.ClubGame
		var playedAt string
		if err := rows.Scan(
			&g.UserID, &g.Name, &g.AvatarIcon, &g.AvatarColor,
			&g.TeamPoints, &g.OppPoints, &g.PointsTarget, &playedAt,
		); err != nil {
			return domain.ClubLeaderboard{}, err
		}
		g.PlayedAt = parseTime(playedAt)
		games = append(games, g)
	}
	if err := rows.Err(); err != nil {
		return domain.ClubLeaderboard{}, err
	}

	return domain.RankClubLeaderboard(games), nil
}

func (s *Store) getClubByID(clubID string) (*domain.Club, error) {
	row, err := s.queries.GetClub(context.Background(), clubID)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &domain.Club{
		ID:          row.ID,
		Name:        row.Name,
		Description: row.Description,
		AvatarIcon:  row.AvatarIcon,
		AvatarColor: row.AvatarColor,
		JoinCode:    row.JoinCode,
		CreatedBy:   row.CreatedBy,
		CreatedAt:   parseTime(row.CreatedAt),
	}, nil
}
