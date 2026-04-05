package store

import (
	"time"
)

type PushSubscription struct {
	Endpoint string
	P256DH   string
	Auth     string
}

func (s *Store) SavePushSubscription(userID, endpoint, p256dh, auth string) error {
	_, err := s.db.Exec(
		`INSERT INTO push_subscriptions (id, user_id, endpoint, p256dh, auth, created_at)
		 VALUES (?, ?, ?, ?, ?, ?)
		 ON CONFLICT(endpoint) DO UPDATE SET p256dh = excluded.p256dh, auth = excluded.auth`,
		newID(), userID, endpoint, p256dh, auth, time.Now().UTC().Format(time.RFC3339),
	)
	return err
}

func (s *Store) DeletePushSubscription(userID, endpoint string) error {
	_, err := s.db.Exec(
		`DELETE FROM push_subscriptions WHERE user_id = ? AND endpoint = ?`,
		userID, endpoint,
	)
	return err
}

func (s *Store) DeleteStalePushSubscription(endpoint string) error {
	_, err := s.db.Exec(`DELETE FROM push_subscriptions WHERE endpoint = ?`, endpoint)
	return err
}

// GetPushSubscriptionsForSession returns all push subscriptions for logged-in
// players in the given session.
func (s *Store) GetPushSubscriptionsForSession(sessionID string) ([]PushSubscription, error) {
	rows, err := s.db.Query(`
		SELECT ps.endpoint, ps.p256dh, ps.auth
		FROM push_subscriptions ps
		JOIN players p ON p.user_id = ps.user_id
		WHERE p.session_id = ? AND p.active = 1`,
		sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []PushSubscription
	for rows.Next() {
		var sub PushSubscription
		if err := rows.Scan(&sub.Endpoint, &sub.P256DH, &sub.Auth); err != nil {
			return nil, err
		}
		subs = append(subs, sub)
	}
	return subs, rows.Err()
}
