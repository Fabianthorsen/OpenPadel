package store

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/fabianthorsen/openpadel/internal/domain"
	"github.com/fabianthorsen/openpadel/internal/store/db"
)

var ErrAlreadyInvited = errors.New("user already invited to this session")

// CreateInvite sends an invite from fromUserID to toUserID for sessionID.
func (s *Store) CreateInvite(sessionID, fromUserID, toUserID string) (*domain.Invite, error) {
	id := newInviteID()
	now := time.Now().UTC().Format(time.RFC3339)
	err := s.queries.CreateInvite(context.Background(), db.CreateInviteParams{
		ID:         id,
		SessionID:  sessionID,
		FromUserID: fromUserID,
		ToUserID:   toUserID,
		CreatedAt:  now,
	})
	if err != nil {
		if isUniqueConstraint(err, "session_id, to_user_id") || isUniqueConstraint(err, "invites") {
			// If the existing invite was declined, reset it to pending so the
			// admin can re-invite the same user.
			var existingID, existingStatus string
			row := s.db.QueryRowContext(context.Background(),
				"SELECT id, status FROM invites WHERE session_id = ? AND to_user_id = ?",
				sessionID, toUserID)
			if scanErr := row.Scan(&existingID, &existingStatus); scanErr == nil && existingStatus == "declined" {
				if _, execErr := s.db.ExecContext(context.Background(),
					"UPDATE invites SET status = 'pending', from_user_id = ? WHERE id = ?",
					fromUserID, existingID); execErr != nil {
					return nil, execErr
				}
				return s.getInviteByID(existingID)
			}
			return nil, ErrAlreadyInvited
		}
		return nil, err
	}
	return s.getInviteByID(id)
}

// CreateSessionInvitesForClub fans a Club's roster out into ordinary Session
// invites — one per Member, each identical in shape to a targeted invite, so
// accepting one creates a Player exactly as today. It is a fan-out, not a new
// invite type. The operation is idempotent and tolerant: the caller
// (fromUserID), Members already joined to the Session, and Members already
// holding a pending invite are skipped rather than erroring, so a repeat
// fan-out is safe. A previously declined Member is re-invited (reset to
// pending) by the underlying CreateInvite. Returns the invites actually created.
func (s *Store) CreateSessionInvitesForClub(sessionID, fromUserID, clubID string) ([]domain.Invite, error) {
	members, err := s.GetClubMembers(clubID)
	if err != nil {
		return nil, err
	}

	sess, err := s.GetSession(sessionID)
	if err != nil {
		return nil, err
	}
	joined := make(map[string]bool, len(sess.Players))
	for _, p := range sess.Players {
		if p.UserID != "" {
			joined[p.UserID] = true
		}
	}

	created := []domain.Invite{}
	for _, m := range members {
		if m.UserID == fromUserID || joined[m.UserID] {
			continue
		}
		inv, err := s.CreateInvite(sessionID, fromUserID, m.UserID)
		if errors.Is(err, ErrAlreadyInvited) {
			continue
		}
		if err != nil {
			return nil, err
		}
		created = append(created, *inv)
	}
	return created, nil
}

// GetPendingInvites returns all pending invites for a user, newest first.
func (s *Store) GetPendingInvites(toUserID string) ([]domain.Invite, error) {
	rows, err := s.queries.GetInvitesByUserID(context.Background(), toUserID)
	if err != nil {
		return nil, err
	}

	var invites []domain.Invite
	for _, row := range rows {
		inv := domain.Invite{
			ID:              row.ID,
			SessionID:       row.SessionID,
			SessionName:     row.SessionName,
			FromUserID:      row.FromUserID,
			FromDisplayName: row.FromDisplayName,
			ToUserID:        toUserID,
			Status:          domain.InviteStatus(row.Status),
			CreatedAt:       parseTime(row.CreatedAt),
		}
		invites = append(invites, inv)
	}
	if invites == nil {
		invites = []domain.Invite{}
	}
	return invites, nil
}

// GetSessionInvites returns all pending invites for a session, including the invited user's name.
func (s *Store) GetSessionInvites(sessionID string) ([]domain.Invite, error) {
	rows, err := s.queries.GetPendingInvitesBySessionID(context.Background(), sessionID)
	if err != nil {
		return nil, err
	}

	var invites []domain.Invite
	for _, row := range rows {
		inv := domain.Invite{
			ID:              row.ID,
			SessionID:       sessionID,
			FromUserID:      row.FromUserID,
			FromDisplayName: row.FromDisplayName,
			ToUserID:        row.ToUserID,
			ToDisplayName:   row.ToDisplayName,
			Status:          domain.InviteStatus(row.Status),
			CreatedAt:       parseTime(row.CreatedAt),
		}
		invites = append(invites, inv)
	}
	if invites == nil {
		invites = []domain.Invite{}
	}
	return invites, nil
}

// AcceptInvite marks the invite as accepted and adds the user as a player.
// Returns the created player record.
func (s *Store) AcceptInvite(inviteID, toUserID string) (*domain.Player, error) {
	inv, err := s.getInviteByID(inviteID)
	if err != nil {
		return nil, err
	}
	if inv.ToUserID != toUserID {
		return nil, ErrNotFound
	}
	if inv.Status != "pending" {
		return nil, errors.New("invite is no longer pending")
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	qtx := s.queries.WithTx(tx)
	if err := qtx.UpdateInviteStatus(context.Background(), db.UpdateInviteStatusParams{
		Status: "accepted",
		ID:     inviteID,
	}); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return s.AddContactPlayer(inv.SessionID, toUserID)
}

// DeclineInvite marks the invite as declined. Returns the session_id so the
// caller can emit SSE events without a second lookup.
func (s *Store) DeclineInvite(inviteID, toUserID string) (string, error) {
	inv, err := s.getInviteByID(inviteID)
	if err != nil {
		return "", err
	}
	if inv.ToUserID != toUserID {
		return "", ErrNotFound
	}
	if inv.Status != "pending" {
		return "", errors.New("invite is no longer pending")
	}
	return inv.SessionID, s.queries.UpdateInviteStatus(context.Background(), db.UpdateInviteStatusParams{
		Status: "declined",
		ID:     inviteID,
	})
}

func (s *Store) getInviteByID(id string) (*domain.Invite, error) {
	row, err := s.queries.GetInvite(context.Background(), id)
	if err != nil {
		return nil, err
	}
	return &domain.Invite{
		ID:              row.ID,
		SessionID:       row.SessionID,
		SessionName:     row.SessionName,
		FromUserID:      row.FromUserID,
		FromDisplayName: row.FromDisplayName,
		ToUserID:        row.ToUserID,
		Status:          domain.InviteStatus(row.Status),
		CreatedAt:       parseTime(row.CreatedAt),
	}, nil
}

func scanInvite(row interface{ Scan(...any) error }) (*domain.Invite, error) {
	var inv domain.Invite
	var createdAt string
	if err := row.Scan(
		&inv.ID, &inv.SessionID, &inv.SessionName,
		&inv.FromUserID, &inv.FromDisplayName,
		&inv.ToUserID, &inv.Status, &createdAt,
	); err != nil {
		return nil, err
	}
	inv.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	return &inv, nil
}

func scanInviteRow(row interface{ Scan(...any) error }) (*domain.Invite, error) {
	return scanInvite(row)
}

func newInviteID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return "inv_" + hex.EncodeToString(b)
}
