package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/fabianthorsen/openpadel/internal/domain"
	"github.com/fabianthorsen/openpadel/internal/store"
	"github.com/fabianthorsen/openpadel/internal/tennis"
)

// startTennisSession validates teams and creates the match record. Returns non-nil error
// if it has already written an error response (so the caller can bail out).
func (h *Handler) startTennisSession(w http.ResponseWriter, sessionID string, active []domain.Player) error {
	teams, err := h.store.GetTennisTeams(sessionID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not load teams")
		return err
	}

	var teamA, teamB int
	for _, t := range teams {
		if t.Team == "a" {
			teamA++
		} else {
			teamB++
		}
	}
	if teamA != 2 || teamB != 2 {
		respondError(w, http.StatusUnprocessableEntity, "each team must have exactly 2 players")
		return errors.New("invalid teams")
	}

	if _, err := h.store.CreateTennisMatch(sessionID); err != nil {
		respondError(w, http.StatusInternalServerError, "could not create match")
		return err
	}
	if err := h.store.StartSession(sessionID, 1, nil); err != nil {
		respondError(w, http.StatusInternalServerError, "could not start session")
		return err
	}
	return nil
}

// setTennisTeams replaces team assignments for a session (admin, lobby only).
func (h *Handler) setTennisTeams(w http.ResponseWriter, r *http.Request) {
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
	if sess.GameMode != "tennis" {
		respondError(w, http.StatusBadRequest, "session is not a tennis match")
		return
	}

	var body struct {
		Teams []domain.TennisTeam `json:"teams"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	for _, t := range body.Teams {
		if t.Team != "a" && t.Team != "b" {
			respondError(w, http.StatusBadRequest, "team must be 'a' or 'b'")
			return
		}
	}

	if err := h.store.SaveTennisTeams(id, body.Teams); err != nil {
		respondError(w, http.StatusInternalServerError, "could not save teams")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// getTennisMatch returns the current match state and team assignments.
func (h *Handler) getTennisMatch(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	match, err := h.store.GetTennisMatch(id)
	if errors.Is(err, store.ErrNotFound) {
		respondError(w, http.StatusNotFound, "match not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not load match")
		return
	}
	respond(w, http.StatusOK, match)
}

// setTennisServer sets which team is currently serving.
func (h *Handler) setTennisServer(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	team := chi.URLParam(r, "team")
	if team != "a" && team != "b" {
		respondError(w, http.StatusBadRequest, "team must be 'a' or 'b'")
		return
	}
	match, err := h.store.GetTennisMatch(id)
	if errors.Is(err, store.ErrNotFound) {
		respondError(w, http.StatusNotFound, "match not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not load match")
		return
	}
	match.State.Server = team
	if err := h.store.SaveTennisState(match.ID, match.State); err != nil {
		respondError(w, http.StatusInternalServerError, "could not save state")
		return
	}
	respond(w, http.StatusOK, match)
}

// addTennisPoint awards a point to team "a" or "b".
func (h *Handler) addTennisPoint(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	team := chi.URLParam(r, "team")
	if team != "a" && team != "b" {
		respondError(w, http.StatusBadRequest, "team must be 'a' or 'b'")
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
	if sess.Status != domain.StatusActive {
		respondError(w, http.StatusConflict, "session is not active")
		return
	}

	match, err := h.store.GetTennisMatch(id)
	if errors.Is(err, store.ErrNotFound) {
		respondError(w, http.StatusNotFound, "match not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not load match")
		return
	}

	newState := tennis.AddPoint(match.State, team, sess.SetsToWin, sess.GamesPerSet)
	if err := h.store.SaveTennisState(match.ID, newState); err != nil {
		respondError(w, http.StatusInternalServerError, "could not save state")
		return
	}

	// If match just finished, mark session complete.
	if newState.Winner != "" && sess.Status == domain.StatusActive {
		h.store.CompleteSession(id, false) //nolint:errcheck
	}

	match.State = newState
	respond(w, http.StatusOK, match)
}
