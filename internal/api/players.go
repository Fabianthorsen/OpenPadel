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

	player, err := h.store.CreatePlayer(id, body.Name, userID)
	if err != nil {
		if isUniqueConstraintError(err) {
			respondAPIError(w, ErrNameTaken)
			return
		}
		respondAPIError(w, ErrServerError)
		return
	}

	// If the joiner is the admin and no creator is set yet, mark them as creator.
	// Accept admin token from Authorization header OR X-Admin-Token header.
	adminToken := extractAdminToken(r)
	if adminToken == "" {
		adminToken = r.Header.Get("X-Admin-Token")
	}
	if isAdmin(adminToken, sess.AdminToken) && sess.CreatorPlayerID == "" {
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
	// Allow admin OR the player removing themselves (via their player token stored in localStorage key).
	// We identify self-removal by a "Player-Id" header matching the target player ID.
	selfRemoval := r.Header.Get("X-Player-Id") == playerID && playerID != ""
	if !isAdmin(extractAdminToken(r), sess.AdminToken) && !selfRemoval {
		respondAPIError(w, ErrAdminRequired)
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

func isUniqueConstraintError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}
