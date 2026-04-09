package store

import (
	"errors"
	"time"

	"github.com/fabianthorsen/openpadel/internal/domain"
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
	var exists int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM users WHERE id = ?`, contactUserID).Scan(&exists)
	if err != nil || exists == 0 {
		return ErrNotFound
	}

	_, err = s.db.Exec(
		`INSERT INTO contacts (user_id, contact_user_id, created_at) VALUES (?, ?, ?)`,
		userID, contactUserID, time.Now().UTC().Format(time.RFC3339),
	)
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
	rows, err := s.db.Query(`
		SELECT u.id, u.display_name, c.created_at
		FROM contacts c
		JOIN users u ON u.id = c.contact_user_id
		WHERE c.user_id = ?
		ORDER BY u.display_name ASC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contacts []domain.Contact
	for rows.Next() {
		var c domain.Contact
		var addedAt string
		if err := rows.Scan(&c.UserID, &c.DisplayName, &addedAt); err != nil {
			return nil, err
		}
		c.AddedAt, _ = time.Parse(time.RFC3339, addedAt)
		contacts = append(contacts, c)
	}
	if contacts == nil {
		contacts = []domain.Contact{}
	}
	return contacts, rows.Err()
}

// SearchUsers finds users whose display name matches the query (case-insensitive),
// excluding the searching user. Marks which results are already contacts.
func (s *Store) SearchUsers(userID, query string) ([]domain.UserSearchResult, error) {
	rows, err := s.db.Query(`
		SELECT
			u.id,
			u.display_name,
			CASE WHEN c.contact_user_id IS NOT NULL THEN 1 ELSE 0 END AS is_contact,
			u.avatar_icon,
			u.avatar_color
		FROM users u
		LEFT JOIN contacts c ON c.user_id = ? AND c.contact_user_id = u.id
		WHERE u.id != ?
		  AND u.display_name LIKE ? ESCAPE '\'
		ORDER BY u.display_name ASC
		LIMIT 20`,
		userID, userID, "%"+escapeLike(query)+"%",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []domain.UserSearchResult
	for rows.Next() {
		var r domain.UserSearchResult
		var isContact int
		if err := rows.Scan(&r.ID, &r.DisplayName, &isContact, &r.AvatarIcon, &r.AvatarColor); err != nil {
			return nil, err
		}
		r.IsContact = isContact == 1
		results = append(results, r)
	}
	if results == nil {
		results = []domain.UserSearchResult{}
	}
	return results, rows.Err()
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
