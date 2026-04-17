package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/fabianthorsen/openpadel/internal/events"
	"github.com/fabianthorsen/openpadel/internal/store"
)

// sendInvite lets the session admin invite a contact by user_id.
func (h *Handler) sendInvite(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")

	var body struct {
		ToUserID string `json:"to_user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ToUserID == "" {
		respondError(w, http.StatusBadRequest, "invalid_request_body")
		return
	}

	_, err := h.store.GetSession(sessionID)
	if errors.Is(err, store.ErrNotFound) {
		respondError(w, http.StatusNotFound, "session_not_found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error")
		return
	}

	fromUser := userFromContext(r)

	inv, err := h.store.CreateInvite(sessionID, fromUser.ID, body.ToUserID)
	if errors.Is(err, store.ErrNotFound) {
		respondError(w, http.StatusNotFound, "user_not_found")
		return
	}
	if errors.Is(err, store.ErrAlreadyInvited) {
		respondError(w, http.StatusConflict, "already_invited")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error")
		return
	}

	h.hub.Emit(sessionID, events.Envelope{Type: events.EventSessionUpdated})
	notifBody := inv.FromDisplayName + " invited you to join a Padel tournament!"
	go h.sendPushToUser(body.ToUserID, "You've been invited!", notifBody, "/s/"+sessionID)
	respond(w, http.StatusCreated, inv)
}

// getSessionInvites returns all pending invites for a session.
func (h *Handler) getSessionInvites(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")
	invites, err := h.store.GetSessionInvites(sessionID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error")
		return
	}
	respond(w, http.StatusOK, invites)
}

// getMyInvites returns all pending invites for the authenticated user.
func (h *Handler) getMyInvites(w http.ResponseWriter, r *http.Request) {
	user := userFromContext(r)
	invites, err := h.store.GetPendingInvites(user.ID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error")
		return
	}
	respond(w, http.StatusOK, invites)
}

// acceptInvite accepts a pending invite and adds the user as a player.
func (h *Handler) acceptInvite(w http.ResponseWriter, r *http.Request) {
	user := userFromContext(r)
	inviteID := chi.URLParam(r, "inviteID")

	player, err := h.store.AcceptInvite(inviteID, user.ID)
	if errors.Is(err, store.ErrNotFound) {
		respondError(w, http.StatusNotFound, "invite_not_found")
		return
	}
	if err != nil {
		if isUniqueConstraintError(err) {
			respondError(w, http.StatusConflict, "already_in_session")
			return
		}
		respondError(w, http.StatusInternalServerError, "server_error")
		return
	}
	h.hub.Emit(player.SessionID, events.Envelope{Type: events.EventSessionUpdated})
	respond(w, http.StatusOK, player)
}

// declineInvite declines a pending invite.
func (h *Handler) declineInvite(w http.ResponseWriter, r *http.Request) {
	user := userFromContext(r)
	inviteID := chi.URLParam(r, "inviteID")

	sessionID, err := h.store.DeclineInvite(inviteID, user.ID)
	if errors.Is(err, store.ErrNotFound) {
		respondError(w, http.StatusNotFound, "invite_not_found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error")
		return
	}
	h.hub.Emit(sessionID, events.Envelope{Type: events.EventSessionUpdated})
	respond(w, http.StatusOK, map[string]string{"status": "declined"})
}
