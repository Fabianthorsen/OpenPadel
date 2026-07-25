package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/fabianthorsen/openpadel/internal/domain"
	"github.com/fabianthorsen/openpadel/internal/events"
	"github.com/fabianthorsen/openpadel/internal/store"
)

// notifyInviteReceived fans an accepted Session invite out to its recipient over
// both channels: an SSE nudge for a live client, and a push for a backgrounded
// one. Shared by the targeted and whole-club invite paths so a Session invite
// notifies identically however it was sent.
func (h *Handler) notifyInviteReceived(inv domain.Invite) {
	h.hub.EmitToUser(inv.ToUserID, events.Envelope{Type: events.EventInviteReceived})
	notifBody := inv.FromDisplayName + " invited you to join a Padel game!"
	go h.sendPushToUser(inv.ToUserID, "You've been invited!", notifBody, "/s/"+inv.SessionID)
}

// sendInvite lets the session admin invite a contact by user_id.
func (h *Handler) sendInvite(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")

	var body struct {
		ToUserID string `json:"to_user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ToUserID == "" {
		respondAPIError(w, ErrInvalidRequestBody)
		return
	}

	_, err := h.store.GetSession(sessionID)
	if errors.Is(err, store.ErrNotFound) {
		respondAPIError(w, ErrSessionNotFound)
		return
	}
	if err != nil {
		respondAPIError(w, ErrServerError)
		return
	}

	fromUser := userFromContext(r)

	inv, err := h.store.CreateInvite(sessionID, fromUser.ID, body.ToUserID)
	if errors.Is(err, store.ErrNotFound) {
		respondAPIError(w, ErrUserNotFound)
		return
	}
	if errors.Is(err, store.ErrAlreadyInvited) {
		respondAPIError(w, ErrAlreadyInvited)
		return
	}
	if err != nil {
		respondAPIError(w, ErrServerError)
		return
	}

	h.notifyInviteReceived(*inv)
	respond(w, http.StatusCreated, inv)
}

// sendClubInvites fans one Club's roster out into ordinary Session invites in a
// single call — "invite my whole club". It reuses the Session Invite model
// unchanged; each Member receives an invite indistinguishable from a targeted
// one. The gate is Club membership (any role), mirroring targeted club-invite
// authz — never the Session AdminToken. Members already joined or already
// invited are skipped silently, so the fan-out never partially fails.
func (h *Handler) sendClubInvites(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")

	var body struct {
		ClubID string `json:"club_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ClubID == "" {
		respondAPIError(w, ErrInvalidRequestBody)
		return
	}

	if _, err := h.store.GetSession(sessionID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			respondAPIError(w, ErrSessionNotFound)
			return
		}
		respondAPIError(w, ErrServerError)
		return
	}

	fromUser := userFromContext(r)

	// Only a Member of the named Club may fan its roster out.
	if _, err := h.store.GetClubMember(body.ClubID, fromUser.ID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			respondAPIError(w, ErrNotClubMember)
			return
		}
		slog.Error("sendClubInvites: GetClubMember", "err", err)
		respondAPIError(w, ErrServerError)
		return
	}

	invites, err := h.store.CreateSessionInvitesForClub(sessionID, fromUser.ID, body.ClubID)
	if err != nil {
		slog.Error("sendClubInvites", "err", err)
		respondAPIError(w, ErrServerError)
		return
	}

	for _, inv := range invites {
		h.notifyInviteReceived(inv)
	}
	respond(w, http.StatusCreated, invites)
}

// getSessionInvites returns all pending invites for a session.
func (h *Handler) getSessionInvites(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")
	invites, err := h.store.GetSessionInvites(sessionID)
	if err != nil {
		respondAPIError(w, ErrServerError)
		return
	}
	respond(w, http.StatusOK, invites)
}

// getMyInvites returns all pending invites for the authenticated user.
func (h *Handler) getMyInvites(w http.ResponseWriter, r *http.Request) {
	user := userFromContext(r)
	invites, err := h.store.GetPendingInvites(user.ID)
	if err != nil {
		respondAPIError(w, ErrServerError)
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
		respondAPIError(w, ErrInviteNotFound)
		return
	}
	if err != nil {
		if isUniqueConstraintError(err) {
			respondAPIError(w, ErrAlreadyInSession)
			return
		}
		respondAPIError(w, ErrServerError)
		return
	}
	h.hub.Emit(player.SessionID, events.Envelope{Type: events.EventSessionUpdated})
	respond(w, http.StatusOK, player)
}

// handleUserEvents streams SSE events for the authenticated user.
// Accepts a bearer token via Authorization header or ?token= query param
// because EventSource cannot set custom headers.
func (h *Handler) handleUserEvents(w http.ResponseWriter, r *http.Request) {
	token := extractAdminToken(r)
	if token == "" {
		token = extractTokenFromQuery(r)
	}
	if token == "" {
		respondAPIError(w, ErrNotAuthenticated)
		return
	}
	user, err := h.store.GetUserByToken(token)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			respondAPIError(w, ErrInvalidToken)
			return
		}
		respondAPIError(w, ErrServerError)
		return
	}
	h.hub.ServeUserSSE(user.ID)(w, r)
}

// declineInvite declines a pending invite.
func (h *Handler) declineInvite(w http.ResponseWriter, r *http.Request) {
	user := userFromContext(r)
	inviteID := chi.URLParam(r, "inviteID")

	sessionID, err := h.store.DeclineInvite(inviteID, user.ID)
	if errors.Is(err, store.ErrNotFound) {
		respondAPIError(w, ErrInviteNotFound)
		return
	}
	if err != nil {
		respondAPIError(w, ErrServerError)
		return
	}
	h.hub.Emit(sessionID, events.Envelope{Type: events.EventSessionUpdated})
	respond(w, http.StatusOK, map[string]string{"status": "declined"})
}
