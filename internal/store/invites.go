package store

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/fabianthorsen/nottennis/internal/domain"
)

var ErrAlreadyInvited = errors.New("user already invited to this session")

// CreateInvite sends an invite from fromUserID to toUserID for sessionID.
func (s *Store) CreateInvite(sessionID, fromUserID, toUserID string) (*domain.Invite, error) {
	id := newInviteID()
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.Exec(
		`INSERT INTO invites (id, session_id, from_user_id, to_user_id, status, created_at)
		 VALUES (?, ?, ?, ?, 'pending', ?)`,
		id, sessionID, fromUserID, toUserID, now,
	)
	if err != nil {
		if isUniqueConstraint(err, "session_id, to_user_id") || isUniqueConstraint(err, "invites") {
			return nil, ErrAlreadyInvited
		}
		return nil, err
	}
	return s.getInviteByID(id)
}

// GetPendingInvites returns all pending invites for a user, newest first.
func (s *Store) GetPendingInvites(toUserID string) ([]domain.Invite, error) {
	rows, err := s.db.Query(`
		SELECT
			i.id, i.session_id,
			COALESCE(NULLIF(s.name, ''), 'NotTennis') AS session_name,
			i.from_user_id, u.display_name,
			i.to_user_id, i.status, i.created_at
		FROM invites i
		JOIN sessions s ON s.id = i.session_id
		JOIN users u ON u.id = i.from_user_id
		WHERE i.to_user_id = ? AND i.status = 'pending'
		ORDER BY i.created_at DESC`,
		toUserID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invites []domain.Invite
	for rows.Next() {
		inv, err := scanInvite(rows)
		if err != nil {
			return nil, err
		}
		invites = append(invites, *inv)
	}
	if invites == nil {
		invites = []domain.Invite{}
	}
	return invites, rows.Err()
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

	if _, err := tx.Exec(
		`UPDATE invites SET status = 'accepted' WHERE id = ?`, inviteID,
	); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return s.AddContactPlayer(inv.SessionID, toUserID)
}

// DeclineInvite marks the invite as declined.
func (s *Store) DeclineInvite(inviteID, toUserID string) error {
	inv, err := s.getInviteByID(inviteID)
	if err != nil {
		return err
	}
	if inv.ToUserID != toUserID {
		return ErrNotFound
	}
	if inv.Status != "pending" {
		return errors.New("invite is no longer pending")
	}
	_, err = s.db.Exec(`UPDATE invites SET status = 'declined' WHERE id = ?`, inviteID)
	return err
}

func (s *Store) getInviteByID(id string) (*domain.Invite, error) {
	row := s.db.QueryRow(`
		SELECT
			i.id, i.session_id,
			COALESCE(NULLIF(s.name, ''), 'NotTennis') AS session_name,
			i.from_user_id, u.display_name,
			i.to_user_id, i.status, i.created_at
		FROM invites i
		JOIN sessions s ON s.id = i.session_id
		JOIN users u ON u.id = i.from_user_id
		WHERE i.id = ?`, id)

	type scanner interface {
		Scan(...any) error
	}
	return scanInviteRow(row)
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
