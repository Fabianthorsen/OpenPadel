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

func (s *Store) RemoveClubMember(clubID, userID string) error {
	return s.queries.DeleteClubMember(context.Background(), db.DeleteClubMemberParams{
		ClubID: clubID,
		UserID: userID,
	})
}

func (s *Store) UpdateClubMemberRole(clubID, userID, role string) error {
	return s.queries.UpdateClubMemberRole(context.Background(), db.UpdateClubMemberRoleParams{
		ClubID: clubID,
		UserID: userID,
		Role:   role,
	})
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
