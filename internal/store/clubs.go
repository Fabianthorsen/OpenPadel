package store

import (
	"context"
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

// DeleteClub hard-deletes a Club and unwinds its dependents in one transaction.
// FK actions are not enforced on our SQLite connection, so the cascade is done
// explicitly here rather than left to ON DELETE clauses: past Sessions are
// detached (club_id set NULL) so they survive as ordinary Sessions, while
// membership and invite rows are removed with the Club.
func (s *Store) DeleteClub(clubID string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec("UPDATE sessions SET club_id = NULL WHERE club_id = ?", clubID); err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM club_invites WHERE club_id = ?", clubID); err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM club_members WHERE club_id = ?", clubID); err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM clubs WHERE id = ?", clubID); err != nil {
		return err
	}

	return tx.Commit()
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
