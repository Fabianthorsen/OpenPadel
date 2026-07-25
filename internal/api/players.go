package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/fabianthorsen/openpadel/internal/domain"
	"github.com/fabianthorsen/openpadel/internal/events"
	"github.com/fabianthorsen/openpadel/internal/store"
)

func (h *Handler) joinSession(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var body struct {
		Name string `json:"name"`
		// Optional guest rating. Registered users always seed from their own
		// self_rating, so this is honoured only for guests (no account).
		Rating *int `json:"rating"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondAPIError(w, ErrInvalidRequestBody)
		return
	}
	body.Name = strings.TrimSpace(body.Name)
	if body.Name == "" {
		respondAPIError(w, ErrNameRequired)
		return
	}
	if body.Rating != nil && !domain.IsValidRating(*body.Rating) {
		respondAPIError(w, ErrInvalidRating)
		return
	}

	sess, err := h.store.GetSession(id)
	if errors.Is(err, store.ErrNotFound) {
		respondAPIError(w, ErrSessionNotFound)
		return
	}
	if err != nil {
		respondAPIError(w, ErrServerError)
		return
	}
	if sess.Status != domain.StatusLobby {
		respondAPIError(w, ErrSessionAlreadyStarted)
		return
	}

	var userID string
	if u := userFromContext(r); u != nil {
		userID = u.ID
	}

	// Accept admin token from Authorization header OR X-Admin-Token header.
	adminToken := extractAdminToken(r)
	if adminToken == "" {
		adminToken = r.Header.Get("X-Admin-Token")
	}
	joinerIsAdmin := isAdmin(adminToken, sess.AdminToken)

	// A guest self-joining by link must pick a skill level (#210). Registered
	// users seed from their own self_rating, and an admin adding a guest by name
	// falls back to the median — so the rating is only required when an anonymous
	// guest joins on their own behalf.
	if body.Rating == nil && userID == "" && !joinerIsAdmin {
		respondAPIError(w, ErrRatingRequired)
		return
	}

	// A guest the admin creates by hand (admin token, no account) is marked so
	// only these players are rating-editable later (#211).
	addedByAdmin := joinerIsAdmin && userID == ""
	player, err := h.store.CreatePlayer(id, body.Name, userID, addedByAdmin)
	if err != nil {
		if isUniqueConstraintError(err) {
			respondAPIError(w, ErrNameTaken)
			return
		}
		respondAPIError(w, ErrServerError)
		return
	}

	// A guest may bring their own rating at join time; registered users keep the
	// rating already seeded from their self_rating.
	if userID == "" && body.Rating != nil {
		if err := h.store.UpdatePlayerRating(player.ID, *body.Rating); err == nil {
			player.Rating = *body.Rating
		}
	}

	// Crown the creator only when the creator themselves join as a player — a
	// registered user whose account matches the session's CreatorUserID. Never
	// auto-assign it to an admin-added guest, who isn't the creator (#211).
	if userID != "" && userID == sess.CreatorUserID && sess.CreatorPlayerID == "" {
		h.store.SetCreatorPlayer(id, player.ID) //nolint:errcheck
	}

	h.hub.Emit(id, events.Envelope{Type: events.EventSessionUpdated})
	respond(w, http.StatusCreated, player)
}

func (h *Handler) deactivatePlayer(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")
	playerID := chi.URLParam(r, "playerID")

	sess, err := h.store.GetSession(sessionID)
	if errors.Is(err, store.ErrNotFound) {
		respondAPIError(w, ErrSessionNotFound)
		return
	}
	if err != nil {
		respondAPIError(w, ErrServerError)
		return
	}
	// Allow admin OR the player removing themselves. Self-removal is proven by the
	// per-player secret issued at join (X-Player-Token), not by the target player
	// id — the id is visible to everyone in the lobby and must not grant removal
	// rights (#241).
	selfRemoval := false
	if token := r.Header.Get("X-Player-Token"); token != "" {
		ok, err := h.store.VerifyPlayerToken(playerID, token)
		if err != nil {
			respondAPIError(w, ErrServerError)
			return
		}
		selfRemoval = ok
	}
	if !isAdmin(extractAdminToken(r), sess.AdminToken) && !selfRemoval {
		respondAPIError(w, ErrAdminRequired)
		return
	}
	// The creator's own Player is the roster's anchor to its admin (admin ≠ roster
	// membership here, but the creator holds the AdminToken). Removing it — whether
	// the creator self-removes or an admin removes them — would leave a session
	// nobody on the roster administers, so it's refused; cancel the session
	// instead. CreatorPlayerID is set only once the creator joins as a Player.
	if sess.CreatorPlayerID != "" && playerID == sess.CreatorPlayerID {
		respondAPIError(w, ErrCreatorCannotLeave)
		return
	}
	if sess.Status != domain.StatusLobby {
		respondAPIError(w, ErrSessionAlreadyStarted)
		return
	}

	if err := h.store.DeactivatePlayer(playerID); errors.Is(err, store.ErrNotFound) {
		respondAPIError(w, ErrPlayerNotFound)
		return
	} else if err != nil {
		respondAPIError(w, ErrServerError)
		return
	}

	h.hub.Emit(sessionID, events.Envelope{Type: events.EventSessionUpdated})
	respond(w, http.StatusOK, map[string]any{"id": playerID, "active": false})
}

// leaveSession lets an authenticated user remove themselves from a session they
// joined. Membership is resolved server-side by matching the caller's user_id
// against the session's Player rows, so it works retroactively for any session —
// including ones joined before self-leave existed, and from a device that never
// stored a client-side player id. Guests (no account) self-remove via the
// per-player X-Player-Token secret in deactivatePlayer instead (#241).
func (h *Handler) leaveSession(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")
	user := userFromContext(r)

	sess, err := h.store.GetSession(sessionID)
	if errors.Is(err, store.ErrNotFound) {
		respondAPIError(w, ErrSessionNotFound)
		return
	}
	if err != nil {
		respondAPIError(w, ErrServerError)
		return
	}
	// Once play has begun the schedule and pairings depend on a stable roster,
	// so leaving is only allowed while still in the lobby.
	if sess.Status != domain.StatusLobby {
		respondAPIError(w, ErrSessionAlreadyStarted)
		return
	}

	var playerID string
	for i := range sess.Players {
		if sess.Players[i].UserID == user.ID && sess.Players[i].Active {
			playerID = sess.Players[i].ID
			break
		}
	}
	if playerID == "" {
		respondAPIError(w, ErrPlayerNotFound)
		return
	}
	// The creator can't leave their own session's roster — they administer it (see
	// deactivatePlayer). They'd cancel the session instead.
	if sess.CreatorPlayerID != "" && playerID == sess.CreatorPlayerID {
		respondAPIError(w, ErrCreatorCannotLeave)
		return
	}

	if err := h.store.DeactivatePlayer(playerID); err != nil {
		respondAPIError(w, ErrServerError)
		return
	}

	h.hub.Emit(sessionID, events.Envelope{Type: events.EventSessionUpdated})
	respond(w, http.StatusOK, map[string]any{"id": playerID, "active": false})
}

// updatePlayerRating lets an admin correct the per-session rating of a guest
// they added by hand (#211). It is gated by isAdmin() only — a matching
// CreatorUserID/CreatorPlayerID never grants edit rights — and further limited
// to admin-added guests, so registered users and self-joined guests keep sole
// control of their own rating.
func (h *Handler) updatePlayerRating(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")
	playerID := chi.URLParam(r, "playerID")

	var body struct {
		Rating int `json:"rating"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondAPIError(w, ErrInvalidRequestBody)
		return
	}
	if !domain.IsValidRating(body.Rating) {
		respondAPIError(w, ErrInvalidRating)
		return
	}

	sess, err := h.store.GetSession(sessionID)
	if errors.Is(err, store.ErrNotFound) {
		respondAPIError(w, ErrSessionNotFound)
		return
	}
	if err != nil {
		respondAPIError(w, ErrServerError)
		return
	}
	// Accept the admin token from the Authorization header OR the X-Admin-Token
	// header — the web client sends it via the latter (mirrors joinSession).
	adminToken := extractAdminToken(r)
	if adminToken == "" {
		adminToken = r.Header.Get("X-Admin-Token")
	}
	if !isAdmin(adminToken, sess.AdminToken) {
		respondAPIError(w, ErrAdminRequired)
		return
	}

	// The player must belong to this session.
	var target *domain.Player
	for i := range sess.Players {
		if sess.Players[i].ID == playerID {
			target = &sess.Players[i]
			break
		}
	}
	if target == nil {
		respondAPIError(w, ErrPlayerNotFound)
		return
	}
	// Only a guest the admin added by hand may be rating-edited. Registered users
	// and self-joined guests own their own rating (#211).
	if !target.AddedByAdmin {
		respondAPIError(w, ErrRatingNotEditable)
		return
	}

	if err := h.store.UpdatePlayerRating(playerID, body.Rating); err != nil {
		respondAPIError(w, ErrServerError)
		return
	}
	target.Rating = body.Rating

	h.hub.Emit(sessionID, events.Envelope{Type: events.EventSessionUpdated})
	respond(w, http.StatusOK, target)
}

func isUniqueConstraintError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}
