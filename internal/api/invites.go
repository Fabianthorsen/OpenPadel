package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/fabianthorsen/nottennis/internal/domain"
	"github.com/fabianthorsen/nottennis/internal/store"
)

// sendInvite lets the session admin invite a contact by user_id.
func (h *Handler) sendInvite(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")

	var body struct {
		ToUserID string `json:"to_user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ToUserID == "" {
		respondError(w, http.StatusBadRequest, "to_user_id is required")
		return
	}

	sess, err := h.store.GetSession(sessionID)
	if errors.Is(err, store.ErrNotFound) {
		respondError(w, http.StatusNotFound, "session not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not load session")
		return
	}
	if !isAdmin(extractAdminToken(r), sess.AdminToken) {
		respondError(w, http.StatusForbidden, "admin access required")
		return
	}
	if sess.Status != domain.StatusLobby {
		respondError(w, http.StatusConflict, "session has already started")
		return
	}

	fromUser := userFromContext(r)
	if fromUser == nil {
		respondError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	inv, err := h.store.CreateInvite(sessionID, fromUser.ID, body.ToUserID)
	if errors.Is(err, store.ErrNotFound) {
		respondError(w, http.StatusNotFound, "user not found")
		return
	}
	if errors.Is(err, store.ErrAlreadyInvited) {
		respondError(w, http.StatusConflict, "user already invited")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not send invite")
		return
	}

	respond(w, http.StatusCreated, inv)
}

// getMyInvites returns all pending invites for the authenticated user.
func (h *Handler) getMyInvites(w http.ResponseWriter, r *http.Request) {
	user := userFromContext(r)
	invites, err := h.store.GetPendingInvites(user.ID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not fetch invites")
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
		respondError(w, http.StatusNotFound, "invite not found")
		return
	}
	if err != nil {
		if isUniqueConstraintError(err) {
			respondError(w, http.StatusConflict, "already in this session")
			return
		}
		respondError(w, http.StatusInternalServerError, "could not accept invite")
		return
	}
	respond(w, http.StatusOK, player)
}

// declineInvite declines a pending invite.
func (h *Handler) declineInvite(w http.ResponseWriter, r *http.Request) {
	user := userFromContext(r)
	inviteID := chi.URLParam(r, "inviteID")

	err := h.store.DeclineInvite(inviteID, user.ID)
	if errors.Is(err, store.ErrNotFound) {
		respondError(w, http.StatusNotFound, "invite not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not decline invite")
		return
	}
	respond(w, http.StatusOK, map[string]string{"status": "declined"})
}
