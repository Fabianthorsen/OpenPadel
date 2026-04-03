package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/fabianthorsen/nottennis/internal/domain"
	"github.com/fabianthorsen/nottennis/internal/store"
)

func (h *Handler) joinSession(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	body.Name = strings.TrimSpace(body.Name)
	if body.Name == "" {
		respondError(w, http.StatusBadRequest, "name is required")
		return
	}

	sess, err := h.store.GetSession(id)
	if errors.Is(err, store.ErrNotFound) {
		respondError(w, http.StatusNotFound, "session not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not load session")
		return
	}
	if sess.Status != domain.StatusLobby {
		respondError(w, http.StatusConflict, "session has already started")
		return
	}

	player, err := h.store.CreatePlayer(id, body.Name)
	if err != nil {
		if isUniqueConstraintError(err) {
			respondError(w, http.StatusConflict, "Oops, somebody already took that name")
			return
		}
		respondError(w, http.StatusInternalServerError, "could not join session")
		return
	}

	// If the joiner is the admin and no creator is set yet, mark them as creator.
	if isAdmin(extractAdminToken(r), sess.AdminToken) && sess.CreatorPlayerID == "" {
		h.store.SetCreatorPlayer(id, player.ID) //nolint:errcheck
	}

	respond(w, http.StatusCreated, player)
}

func (h *Handler) deactivatePlayer(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")
	playerID := chi.URLParam(r, "playerID")

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
		respondError(w, http.StatusConflict, "cannot remove players after session has started")
		return
	}

	if err := h.store.DeactivatePlayer(playerID); errors.Is(err, store.ErrNotFound) {
		respondError(w, http.StatusNotFound, "player not found")
		return
	} else if err != nil {
		respondError(w, http.StatusInternalServerError, "could not remove player")
		return
	}

	respond(w, http.StatusOK, map[string]any{"id": playerID, "active": false})
}

func isUniqueConstraintError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}
