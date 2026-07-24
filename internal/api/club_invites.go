package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/fabianthorsen/openpadel/internal/events"
	"github.com/fabianthorsen/openpadel/internal/store"
)

// sendClubInvite lets any Club Member invite a registered User to the Club.
func (h *Handler) sendClubInvite(w http.ResponseWriter, r *http.Request) {
	clubID := chi.URLParam(r, "id")
	user := userFromContext(r)
	if user == nil {
		respondAPIError(w, ErrNotAuthenticated)
		return
	}

	var body struct {
		ToUserID string `json:"to_user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ToUserID == "" {
		respondAPIError(w, ErrInvalidRequestBody)
		return
	}

	// Any Member may invite — membership is the only gate (not admin).
	if _, err := h.store.GetClubMember(clubID, user.ID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			respondAPIError(w, ErrNotClubMember)
			return
		}
		slog.Error("sendClubInvite: GetClubMember", "err", err)
		respondAPIError(w, ErrServerError)
		return
	}

	if _, err := h.store.GetUserByID(body.ToUserID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			respondAPIError(w, ErrUserNotFound)
			return
		}
		slog.Error("sendClubInvite: GetUserByID", "err", err)
		respondAPIError(w, ErrServerError)
		return
	}

	inv, err := h.store.CreateClubInvite(clubID, user.ID, body.ToUserID)
	if errors.Is(err, store.ErrAlreadyMember) {
		respondAPIError(w, ErrAlreadyClubMember)
		return
	}
	if errors.Is(err, store.ErrAlreadyClubInvited) {
		respondAPIError(w, ErrAlreadyClubInvited)
		return
	}
	if err != nil {
		slog.Error("sendClubInvite", "err", err)
		respondAPIError(w, ErrServerError)
		return
	}

	h.hub.EmitToUser(body.ToUserID, events.Envelope{Type: events.EventClubInviteReceived})
	notifBody := inv.InviterDisplayName + " invited you to join " + inv.ClubName
	go h.sendPushToUser(body.ToUserID, "Club invite", notifBody, "/profile")
	respond(w, http.StatusCreated, inv)
}

// getMyClubInvites returns the caller's pending Club invites.
func (h *Handler) getMyClubInvites(w http.ResponseWriter, r *http.Request) {
	user := userFromContext(r)
	if user == nil {
		respondAPIError(w, ErrNotAuthenticated)
		return
	}
	invites, err := h.store.GetPendingClubInvites(user.ID)
	if err != nil {
		slog.Error("getMyClubInvites", "err", err)
		respondAPIError(w, ErrServerError)
		return
	}
	respond(w, http.StatusOK, invites)
}

// acceptClubInvite makes the caller a Member of the invited Club.
func (h *Handler) acceptClubInvite(w http.ResponseWriter, r *http.Request) {
	user := userFromContext(r)
	if user == nil {
		respondAPIError(w, ErrNotAuthenticated)
		return
	}
	inviteID := chi.URLParam(r, "inviteID")

	inv, err := h.store.AcceptClubInvite(inviteID, user.ID)
	if errors.Is(err, store.ErrNotFound) {
		respondAPIError(w, ErrInviteNotFound)
		return
	}
	if errors.Is(err, store.ErrInviteNotPending) {
		respondAPIError(w, ErrAlreadyClubMember)
		return
	}
	if err != nil {
		slog.Error("acceptClubInvite", "err", err)
		respondAPIError(w, ErrServerError)
		return
	}
	respond(w, http.StatusOK, map[string]string{"id": inv.ClubID})
}

// declineClubInvite marks the caller's pending Club invite declined.
func (h *Handler) declineClubInvite(w http.ResponseWriter, r *http.Request) {
	user := userFromContext(r)
	if user == nil {
		respondAPIError(w, ErrNotAuthenticated)
		return
	}
	inviteID := chi.URLParam(r, "inviteID")

	err := h.store.DeclineClubInvite(inviteID, user.ID)
	if errors.Is(err, store.ErrNotFound) {
		respondAPIError(w, ErrInviteNotFound)
		return
	}
	if errors.Is(err, store.ErrInviteNotPending) {
		respondAPIError(w, ErrInviteNotFound)
		return
	}
	if err != nil {
		slog.Error("declineClubInvite", "err", err)
		respondAPIError(w, ErrServerError)
		return
	}
	respond(w, http.StatusOK, map[string]string{"status": "declined"})
}
