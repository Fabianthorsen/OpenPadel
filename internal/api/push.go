package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	webpush "github.com/SherClockHolmes/webpush-go"

)

func (h *Handler) vapidPublicKey(w http.ResponseWriter, r *http.Request) {
	respond(w, http.StatusOK, map[string]string{"public_key": h.vapidPublic})
}

func (h *Handler) subscribePush(w http.ResponseWriter, r *http.Request) {
	u := userFromContext(r)
	if u == nil {
		respondAPIError(w, ErrNotAuthenticated)
		return
	}

	var body struct {
		Endpoint string `json:"endpoint"`
		P256DH   string `json:"p256dh"`
		Auth     string `json:"auth"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Endpoint == "" {
		respondAPIError(w, ErrInvalidRequestBody)
		return
	}

	if err := h.store.SavePushSubscription(u.ID, body.Endpoint, body.P256DH, body.Auth); err != nil {
		respondAPIError(w, ErrServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) unsubscribePush(w http.ResponseWriter, r *http.Request) {
	u := userFromContext(r)
	if u == nil {
		respondAPIError(w, ErrNotAuthenticated)
		return
	}

	var body struct {
		Endpoint string `json:"endpoint"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Endpoint == "" {
		respondAPIError(w, ErrInvalidRequestBody)
		return
	}

	h.store.DeletePushSubscription(u.ID, body.Endpoint) //nolint:errcheck
	w.WriteHeader(http.StatusNoContent)
}

// sendPushToUser sends a push notification to a single user's subscriptions.
func (h *Handler) sendPushToUser(userID, title, body, url string) {
	if h.vapidPrivate == "" || h.vapidPublic == "" {
		return
	}
	subs, err := h.store.GetPushSubscriptionsForUser(userID)
	if err != nil {
		slog.Error("push: get user subscriptions", "err", err)
		return
	}
	payload, _ := json.Marshal(map[string]string{"title": title, "body": body, "url": url})
	for _, sub := range subs {
		s := &webpush.Subscription{
			Endpoint: sub.Endpoint,
			Keys: webpush.Keys{
				P256dh: sub.P256DH,
				Auth:   sub.Auth,
			},
		}
		resp, err := webpush.SendNotification(payload, s, &webpush.Options{
			VAPIDPublicKey:  h.vapidPublic,
			VAPIDPrivateKey: h.vapidPrivate,
			Subscriber:      fmt.Sprintf("mailto:noreply@%s", "openpadel.app"),
			TTL:             60,
		})
		if err != nil {
			slog.Error("push: send to user", "endpoint", sub.Endpoint, "err", err)
			continue
		}
		resp.Body.Close()
		if resp.StatusCode == http.StatusGone {
			h.store.DeleteStalePushSubscription(sub.Endpoint) //nolint:errcheck
		}
	}
}

// sendPushToSession fans out a push notification to all subscribed players in a session.
func (h *Handler) sendPushToSession(sessionID, title, body string) {
	if h.vapidPrivate == "" || h.vapidPublic == "" {
		return
	}
	subs, err := h.store.GetPushSubscriptionsForSession(sessionID)
	if err != nil {
		slog.Error("push: get subscriptions", "err", err)
		return
	}

	payload, _ := json.Marshal(map[string]string{"title": title, "body": body, "url": "/s/" + sessionID})

	for _, sub := range subs {
		s := &webpush.Subscription{
			Endpoint: sub.Endpoint,
			Keys: webpush.Keys{
				P256dh: sub.P256DH,
				Auth:   sub.Auth,
			},
		}
		resp, err := webpush.SendNotification(payload, s, &webpush.Options{
			VAPIDPublicKey:  h.vapidPublic,
			VAPIDPrivateKey: h.vapidPrivate,
			Subscriber:      fmt.Sprintf("mailto:noreply@%s", "openpadel.app"),
			TTL:             60,
		})
		if err != nil {
			slog.Error("push: send notification", "endpoint", sub.Endpoint, "err", err)
			continue
		}
		resp.Body.Close()
		if resp.StatusCode == http.StatusGone {
			// Subscription expired — clean it up.
			h.store.DeleteStalePushSubscription(sub.Endpoint) //nolint:errcheck
		}
	}
}
