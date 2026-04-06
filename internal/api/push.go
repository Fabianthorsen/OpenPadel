package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	webpush "github.com/SherClockHolmes/webpush-go"

)

func (h *Handler) vapidPublicKey(w http.ResponseWriter, r *http.Request) {
	respond(w, http.StatusOK, map[string]string{"public_key": h.vapidPublic})
}

func (h *Handler) subscribePush(w http.ResponseWriter, r *http.Request) {
	u := userFromContext(r)
	if u == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var body struct {
		Endpoint string `json:"endpoint"`
		P256DH   string `json:"p256dh"`
		Auth     string `json:"auth"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Endpoint == "" {
		respondError(w, http.StatusBadRequest, "invalid subscription")
		return
	}

	if err := h.store.SavePushSubscription(u.ID, body.Endpoint, body.P256DH, body.Auth); err != nil {
		respondError(w, http.StatusInternalServerError, "could not save subscription")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) unsubscribePush(w http.ResponseWriter, r *http.Request) {
	u := userFromContext(r)
	if u == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var body struct {
		Endpoint string `json:"endpoint"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Endpoint == "" {
		respondError(w, http.StatusBadRequest, "invalid request")
		return
	}

	h.store.DeletePushSubscription(u.ID, body.Endpoint) //nolint:errcheck
	w.WriteHeader(http.StatusNoContent)
}

// sendPushToSession fans out a push notification to all subscribed players in a session.
func (h *Handler) sendPushToSession(sessionID, title, body string) {
	if h.vapidPrivate == "" || h.vapidPublic == "" {
		return
	}
	subs, err := h.store.GetPushSubscriptionsForSession(sessionID)
	if err != nil {
		log.Printf("push: get subscriptions: %v", err)
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
			log.Printf("push: send to %s: %v", sub.Endpoint, err)
			continue
		}
		resp.Body.Close()
		if resp.StatusCode == http.StatusGone {
			// Subscription expired — clean it up.
			h.store.DeleteStalePushSubscription(sub.Endpoint) //nolint:errcheck
		}
	}
}
