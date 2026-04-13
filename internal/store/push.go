package store

import (
	"context"
	"time"

	"github.com/fabianthorsen/openpadel/internal/store/db"
)

type PushSubscription struct {
	Endpoint string
	P256DH   string
	Auth     string
}

func (s *Store) SavePushSubscription(userID, endpoint, p256dh, auth string) error {
	return s.queries.SavePushSubscription(context.Background(), db.SavePushSubscriptionParams{
		ID:        newID(),
		UserID:    userID,
		Endpoint:  endpoint,
		P256dh:    p256dh,
		Auth:      auth,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	})
}

func (s *Store) DeletePushSubscription(userID, endpoint string) error {
	return s.queries.DeletePushSubscription(context.Background(), db.DeletePushSubscriptionParams{
		UserID:   userID,
		Endpoint: endpoint,
	})
}

func (s *Store) DeleteStalePushSubscription(endpoint string) error {
	return s.queries.DeleteStalePushSubscription(context.Background(), endpoint)
}

// GetPushSubscriptionsForSession returns all push subscriptions for logged-in
// players in the given session.
func (s *Store) GetPushSubscriptionsForSession(sessionID string) ([]PushSubscription, error) {
	rows, err := s.queries.GetPushSubscriptionsForSession(context.Background(), sessionID)
	if err != nil {
		return nil, err
	}

	var subs []PushSubscription
	for _, row := range rows {
		sub := PushSubscription{
			Endpoint: row.Endpoint,
			P256DH:   row.P256dh,
			Auth:     row.Auth,
		}
		subs = append(subs, sub)
	}
	return subs, nil
}
