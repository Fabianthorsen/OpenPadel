package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/fabianthorsen/openpadel/internal/domain"
	"github.com/fabianthorsen/openpadel/internal/store"
)

func (h *Handler) getRounds(w http.ResponseWriter, r *http.Request) {
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
	if sess.Status == domain.StatusLobby {
		respondError(w, http.StatusConflict, "session has not started yet")
		return
	}

	rounds, err := h.store.GetRounds(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not load rounds")
		return
	}
	respond(w, http.StatusOK, map[string]any{"rounds": rounds})
}

func (h *Handler) getCurrentRound(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	round, err := h.store.GetCurrentRound(id)
	if errors.Is(err, store.ErrNotFound) {
		respondError(w, http.StatusNotFound, "no active round found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not load round")
		return
	}
	// Overlay in-memory live scores for matches that haven't been finalised yet.
	for i := range round.Matches {
		m := &round.Matches[i]
		if m.Score == nil {
			if ls, ok := h.live.Get(m.ID); ok {
				m.Live = &domain.LiveScore{A: ls.A, B: ls.B, Server: ls.Server}
			}
		}
	}
	respond(w, http.StatusOK, round)
}

func (h *Handler) submitScore(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")
	matchID := chi.URLParam(r, "matchID")

	sess, err := h.store.GetSession(sessionID)
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

	var body struct {
		ScoreA int `json:"score_a"`
		ScoreB int `json:"score_b"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.ScoreA+body.ScoreB != sess.Points {
		respondError(w, http.StatusBadRequest, "scores must sum to the session points target")
		return
	}
	if body.ScoreA < 0 || body.ScoreB < 0 {
		respondError(w, http.StatusBadRequest, "scores cannot be negative")
		return
	}

	match, err := h.store.UpdateScore(matchID, body.ScoreA, body.ScoreB)
	if errors.Is(err, store.ErrNotFound) {
		respondError(w, http.StatusNotFound, "match not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not save score")
		return
	}
	// Final score is now in the DB — clear any in-memory live score for this match.
	h.live.Clear(matchID)

	// Check if all rounds are now complete.
	done, err := h.store.AllRoundsComplete(sessionID)
	if err == nil && done {
		h.store.CompleteSession(sessionID, false)
	}

	respond(w, http.StatusOK, match)
}

func (h *Handler) updateLiveScore(w http.ResponseWriter, r *http.Request) {
	var body struct {
		A      int    `json:"a"`
		B      int    `json:"b"`
		Server string `json:"server"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.A < 0 || body.B < 0 {
		respondError(w, http.StatusBadRequest, "scores cannot be negative")
		return
	}
	matchID := chi.URLParam(r, "matchID")
	h.live.Set(matchID, body.Server, body.A, body.B)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) advanceRound(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")
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
	if sess.Status != domain.StatusActive {
		respondError(w, http.StatusConflict, "session is not active")
		return
	}
	allScored, err := h.store.CurrentRoundAllScored(sessionID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not check round status")
		return
	}
	if !allScored {
		respondError(w, http.StatusConflict, "not all matches in this round are scored")
		return
	}
	if err := h.store.AdvanceRound(sessionID); err != nil {
		respondError(w, http.StatusInternalServerError, "could not advance round")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) getLeaderboard(w http.ResponseWriter, r *http.Request) {
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

	standings, err := h.store.GetLeaderboard(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not compute leaderboard")
		return
	}

	currentRound, _ := h.store.GetCurrentRound(id)
	var currentRoundNum *int
	if currentRound != nil {
		currentRoundNum = &currentRound.Number
	}

	respond(w, http.StatusOK, domain.Leaderboard{
		SessionID:    id,
		Status:       sess.Status,
		CurrentRound: currentRoundNum,
		TotalRounds:  sess.RoundsTotal,
		Standings:    standings,
	})
}
