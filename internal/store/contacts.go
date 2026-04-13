package store

import (
	"context"
	"errors"
	"time"

	"github.com/fabianthorsen/openpadel/internal/domain"
	"github.com/fabianthorsen/openpadel/internal/store/db"
)

// AddContact adds contact_user_id as a contact of user_id.
// Returns ErrNotFound if the contact user doesn't exist.
// Returns ErrAlreadyContact if the relationship already exists.
var ErrAlreadyContact = errors.New("already a contact")

func (s *Store) AddContact(userID, contactUserID string) error {
	if userID == contactUserID {
		return errors.New("cannot add yourself as a contact")
	}

	// Verify the target user exists.
	exists, err := s.queries.UserExists(context.Background(), contactUserID)
	if err != nil || exists == 0 {
		return ErrNotFound
	}

	err = s.queries.AddContact(context.Background(), db.AddContactParams{
		UserID:        userID,
		ContactUserID: contactUserID,
		CreatedAt:     time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		if isUniqueConstraint(err, "contacts") {
			return ErrAlreadyContact
		}
		return err
	}
	return nil
}

// RemoveContact removes contact_user_id from user_id's contacts.
func (s *Store) RemoveContact(userID, contactUserID string) error {
	res, err := s.db.Exec(
		`DELETE FROM contacts WHERE user_id = ? AND contact_user_id = ?`,
		userID, contactUserID,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// GetContacts returns all contacts for a user, ordered by display name.
func (s *Store) GetContacts(userID string) ([]domain.Contact, error) {
	rows, err := s.queries.GetContacts(context.Background(), userID)
	if err != nil {
		return nil, err
	}

	var contacts []domain.Contact
	for _, row := range rows {
		c := domain.Contact{
			UserID:      row.ContactUserID,
			DisplayName: row.DisplayName,
			AddedAt:     parseTime(row.CreatedAt),
		}
		contacts = append(contacts, c)
	}
	if contacts == nil {
		contacts = []domain.Contact{}
	}
	return contacts, nil
}

// SearchUsers finds users whose display name matches the query (case-insensitive),
// excluding the searching user. Marks which results are already contacts.
func (s *Store) SearchUsers(userID, query string) ([]domain.UserSearchResult, error) {
	rows, err := s.queries.SearchUsers(context.Background(), db.SearchUsersParams{
		UserID:      userID,
		DisplayName: "%" + escapeLike(query) + "%",
		ID:          userID,
		Limit:       20,
	})
	if err != nil {
		return nil, err
	}

	var results []domain.UserSearchResult
	for _, row := range rows {
		r := domain.UserSearchResult{
			ID:          row.ID,
			DisplayName: row.DisplayName,
			AvatarIcon:  row.AvatarIcon,
			AvatarColor: row.AvatarColor,
			IsContact:   row.IsContact == 1,
		}
		results = append(results, r)
	}
	if results == nil {
		results = []domain.UserSearchResult{}
	}
	return results, nil
}


// AddContactPlayer looks up the contact user's display name and creates a player
// record in the session linked to their user account. Returns ErrNotFound if the
// user doesn't exist.
func (s *Store) AddContactPlayer(sessionID, contactUserID string) (*domain.Player, error) {
	user, err := s.GetUserByID(contactUserID)
	if err != nil {
		return nil, err
	}
	return s.CreatePlayer(sessionID, user.DisplayName, user.ID)
}

// escapeLike escapes special LIKE characters in a query string.
func escapeLike(s string) string {
	out := make([]byte, 0, len(s)*2)
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '%' || c == '_' || c == '\\' {
			out = append(out, '\\')
		}
		out = append(out, c)
	}
	return string(out)
}
