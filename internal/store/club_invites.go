package store

import (
	"context"
	"errors"
	"time"

	"github.com/fabianthorsen/openpadel/internal/domain"
	"github.com/fabianthorsen/openpadel/internal/store/db"
)

var (
	ErrAlreadyClubInvited = errors.New("user already invited to this club")
	ErrInviteNotPending   = errors.New("club invite is no longer pending")
)

// CreateClubInvite records a pending invite for inviteeID to join clubID, sent by
// inviterID. If a prior invite to the same User was declined it is reset to pending
// so a Member can re-invite. A still-pending duplicate returns ErrAlreadyClubInvited;
// an already-a-Member invitee returns ErrAlreadyMember.
func (s *Store) CreateClubInvite(clubID, inviterID, inviteeID string) (*domain.ClubInvite, error) {
	if existing, err := s.GetClubMember(clubID, inviteeID); err == nil && existing != nil {
		return nil, ErrAlreadyMember
	} else if err != nil && err != ErrNotFound {
		return nil, err
	}

	id := newClubInviteID()
	now := time.Now().UTC().Format(time.RFC3339)
	err := s.queries.CreateClubInvite(context.Background(), db.CreateClubInviteParams{
		ID:        id,
		ClubID:    clubID,
		InviterID: inviterID,
		InviteeID: inviteeID,
		CreatedAt: now,
	})
	if err != nil {
		if isUniqueConstraint(err, "club_id") || isUniqueConstraint(err, "club_invites") {
			// A declined invite is revived to pending so the same User can be
			// re-invited; a still-pending one is a graceful no-op error.
			row, getErr := s.queries.GetClubInviteByClubAndInvitee(context.Background(), db.GetClubInviteByClubAndInviteeParams{
				ClubID:    clubID,
				InviteeID: inviteeID,
			})
			if getErr != nil {
				return nil, getErr
			}
			if row.Status == "declined" {
				if resetErr := s.queries.ResetClubInvite(context.Background(), db.ResetClubInviteParams{
					InviterID: inviterID,
					ID:        row.ID,
				}); resetErr != nil {
					return nil, resetErr
				}
				return s.getClubInviteByID(row.ID)
			}
			return nil, ErrAlreadyClubInvited
		}
		return nil, err
	}
	return s.getClubInviteByID(id)
}

// GetPendingClubInvites returns all pending Club invites for a User, newest first.
func (s *Store) GetPendingClubInvites(inviteeID string) ([]domain.ClubInvite, error) {
	rows, err := s.queries.GetClubInvitesByInviteeID(context.Background(), inviteeID)
	if err != nil {
		return nil, err
	}

	invites := make([]domain.ClubInvite, 0, len(rows))
	for _, row := range rows {
		invites = append(invites, domain.ClubInvite{
			ID:                 row.ID,
			ClubID:             row.ClubID,
			ClubName:           row.ClubName,
			ClubAvatarIcon:     row.ClubAvatarIcon,
			ClubAvatarColor:    row.ClubAvatarColor,
			InviterID:          row.InviterID,
			InviterDisplayName: row.InviterDisplayName,
			InviteeID:          inviteeID,
			Status:             domain.InviteStatus(row.Status),
			CreatedAt:          parseTime(row.CreatedAt),
		})
	}
	return invites, nil
}

// AcceptClubInvite marks the invite accepted and adds the invitee to the Club roster
// as a plain member, in one transaction. Only the invitee may accept.
func (s *Store) AcceptClubInvite(inviteID, inviteeID string) (*domain.ClubInvite, error) {
	inv, err := s.getClubInviteByID(inviteID)
	if err != nil {
		return nil, err
	}
	if inv.InviteeID != inviteeID {
		return nil, ErrNotFound
	}
	if inv.Status != "pending" {
		return nil, ErrInviteNotPending
	}

	now := time.Now().UTC().Format(time.RFC3339)

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	qtx := s.queries.WithTx(tx)
	if err := qtx.UpdateClubInviteStatus(context.Background(), db.UpdateClubInviteStatusParams{
		Status: "accepted",
		ID:     inviteID,
	}); err != nil {
		return nil, err
	}
	if err := qtx.InsertClubMember(context.Background(), db.InsertClubMemberParams{
		ClubID:   inv.ClubID,
		UserID:   inviteeID,
		Role:     "member",
		JoinedAt: now,
	}); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return inv, nil
}

// DeclineClubInvite marks the invite declined. Only the invitee may decline.
func (s *Store) DeclineClubInvite(inviteID, inviteeID string) error {
	inv, err := s.getClubInviteByID(inviteID)
	if err != nil {
		return err
	}
	if inv.InviteeID != inviteeID {
		return ErrNotFound
	}
	if inv.Status != "pending" {
		return ErrInviteNotPending
	}
	return s.queries.UpdateClubInviteStatus(context.Background(), db.UpdateClubInviteStatusParams{
		Status: "declined",
		ID:     inviteID,
	})
}

func (s *Store) getClubInviteByID(id string) (*domain.ClubInvite, error) {
	row, err := s.queries.GetClubInvite(context.Background(), id)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &domain.ClubInvite{
		ID:                 row.ID,
		ClubID:             row.ClubID,
		ClubName:           row.ClubName,
		ClubAvatarIcon:     row.ClubAvatarIcon,
		ClubAvatarColor:    row.ClubAvatarColor,
		InviterID:          row.InviterID,
		InviterDisplayName: row.InviterDisplayName,
		InviteeID:          row.InviteeID,
		Status:             domain.InviteStatus(row.Status),
		CreatedAt:          parseTime(row.CreatedAt),
	}, nil
}

func newClubInviteID() string {
	return "cinv_" + randString(16)
}
