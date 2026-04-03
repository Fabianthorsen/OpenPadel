package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/fabianthorsen/nottennis/internal/domain"
	"github.com/fabianthorsen/nottennis/internal/scheduler"
	"github.com/fabianthorsen/nottennis/internal/store"
)

func (h *Handler) createSession(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Courts int `json:"courts"`
		Points int `json:"points"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Courts < 1 || body.Courts > 4 {
		respondError(w, http.StatusBadRequest, "courts must be between 1 and 4")
		return
	}
	if body.Points != 16 && body.Points != 24 && body.Points != 32 {
		respondError(w, http.StatusBadRequest, "points must be 16, 24, or 32")
		return
	}

	sess, err := h.store.CreateSession(body.Courts, body.Points)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not create session")
		return
	}
	respond(w, http.StatusCreated, sess)
}

func (h *Handler) getSession(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	sess, err := h.store.GetSession(id)
	if errors.Is(err, store.ErrNotFound) {
		respondError(w, http.StatusNotFound, "session not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not load session")
		return
	}

	// Strip admin token from public responses.
	if !isAdmin(extractAdminToken(r), sess.AdminToken) {
		sess.AdminToken = ""
	}

	respond(w, http.StatusOK, sess)
}

func (h *Handler) startSession(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	sess, err := h.store.GetSession(id)
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
		respondError(w, http.StatusConflict, "session already started")
		return
	}

	active := activePlayers(sess.Players)
	minPlayers := sess.Courts*4 + 1
	if len(active) < minPlayers {
		respondError(w, http.StatusUnprocessableEntity, "not enough players to start")
		return
	}

	totalRounds := len(active)
	rounds := scheduler.Generate(active, sess.Courts, totalRounds)

	if err := h.store.SaveRounds(id, rounds); err != nil {
		respondError(w, http.StatusInternalServerError, "could not generate rounds")
		return
	}
	if err := h.store.StartSession(id, totalRounds); err != nil {
		respondError(w, http.StatusInternalServerError, "could not start session")
		return
	}

	sess, _ = h.store.GetSession(id)
	sess.AdminToken = ""
	respond(w, http.StatusOK, sess)
}

func activePlayers(players []domain.Player) []domain.Player {
	var out []domain.Player
	for _, p := range players {
		if p.Active {
			out = append(out, p)
		}
	}
	return out
}
