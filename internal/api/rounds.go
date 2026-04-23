package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/fabianthorsen/openpadel/internal/domain"
	"github.com/fabianthorsen/openpadel/internal/events"
	"github.com/fabianthorsen/openpadel/internal/store"
)

func (h *Handler) getRounds(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	sess, err := h.store.GetSession(id)
	if errors.Is(err, store.ErrNotFound) {
		respondError(w, http.StatusNotFound, "session_not_found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error")
		return
	}
	if sess.Status == domain.StatusLobby {
		respondError(w, http.StatusConflict, "session_not_started")
		return
	}

	rounds, err := h.store.GetRounds(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error")
		return
	}
	respond(w, http.StatusOK, map[string]any{"rounds": rounds})
}

func (h *Handler) getCurrentRound(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	round, err := h.store.GetCurrentRound(id)
	if errors.Is(err, store.ErrNotFound) {
		respondError(w, http.StatusNotFound, "no_active_round")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error")
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
		respondError(w, http.StatusNotFound, "session_not_found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error")
		return
	}
	if sess.Status != domain.StatusActive {
		respondError(w, http.StatusConflict, "session_not_active")
		return
	}

	var body struct {
		ScoreA int `json:"score_a"`
		ScoreB int `json:"score_b"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request_body")
		return
	}
	if body.ScoreA < 0 || body.ScoreB < 0 {
		respondError(w, http.StatusBadRequest, "scores_negative")
		return
	}
	if body.ScoreA+body.ScoreB != sess.Points {
		respondError(w, http.StatusBadRequest, "scores_invalid_sum")
		return
	}

	match, err := h.store.UpdateScore(matchID, body.ScoreA, body.ScoreB)
	if errors.Is(err, store.ErrNotFound) {
		respondError(w, http.StatusNotFound, "match_not_found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error")
		return
	}
	// Final score is now in the DB — clear any in-memory live score for this match.
	h.live.Clear(matchID)
	h.hub.Emit(sessionID, events.Envelope{Type: events.EventRoundUpdated})

	// Auto-complete logic.
	// Timer takes priority: if ends_at has passed, complete once current round is fully scored.
	// Otherwise use normal per-mode logic.
	sessionCompleted := false
	timerExpired := sess.EndsAt != nil && time.Now().UTC().After(*sess.EndsAt)
	if timerExpired {
		allScored, err := h.store.CurrentRoundAllScored(sessionID)
		if err == nil && allScored {
			h.store.CompleteSession(sessionID, false) //nolint:errcheck
			sessionCompleted = true
		}
	} else if sess.GameMode == "mexicano" {
		// Mexicano with preset rounds_total: complete when last round is fully scored.
		if sess.RoundsTotal != nil && sess.CurrentRound != nil && *sess.CurrentRound == *sess.RoundsTotal {
			allScored, err := h.store.CurrentRoundAllScored(sessionID)
			if err == nil && allScored {
				h.store.CompleteSession(sessionID, false) //nolint:errcheck
				sessionCompleted = true
			}
		}
	} else if sess.GameMode == "americano" {
		// Americano: all pre-generated rounds complete.
		done, err := h.americanoSvc.CanComplete(sessionID)
		if err == nil && done {
			h.store.CompleteSession(sessionID, false) //nolint:errcheck
			sessionCompleted = true
		}
	}
	if sessionCompleted {
		h.hub.Emit(sessionID, events.Envelope{Type: events.EventSessionUpdated})
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
		respondError(w, http.StatusBadRequest, "invalid_request_body")
		return
	}
	if body.A < 0 || body.B < 0 {
		respondError(w, http.StatusBadRequest, "scores_negative")
		return
	}
	sessionID := chi.URLParam(r, "id")
	matchID := chi.URLParam(r, "matchID")
	h.live.Set(matchID, body.Server, body.A, body.B)
	h.hub.Emit(sessionID, events.Envelope{
		Type:    events.EventLiveScore,
		Payload: map[string]any{"match_id": matchID, "a": body.A, "b": body.B, "server": body.Server},
	})
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) advanceRound(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")
	sess, err := h.store.GetSession(sessionID)
	if errors.Is(err, store.ErrNotFound) {
		respondError(w, http.StatusNotFound, "session_not_found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error")
		return
	}
	if !isAdmin(extractAdminToken(r), sess.AdminToken) {
		respondError(w, http.StatusForbidden, "admin_required")
		return
	}
	if sess.Status != domain.StatusActive {
		respondError(w, http.StatusConflict, "session_not_active")
		return
	}
	if sess.EndsAt != nil && time.Now().UTC().After(*sess.EndsAt) {
		respondError(w, http.StatusConflict, "tournament_expired")
		return
	}
	allScored, err := h.store.CurrentRoundAllScored(sessionID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error")
		return
	}
	if !allScored {
		respondError(w, http.StatusConflict, "round_not_complete")
		return
	}

	if sess.GameMode == "mexicano" {
		nextRound := 1
		if sess.CurrentRound != nil {
			nextRound = *sess.CurrentRound + 1
		}
		if sess.RoundsTotal != nil && nextRound > *sess.RoundsTotal {
			respondError(w, http.StatusConflict, "round_limit_reached")
			return
		}
		if err := h.mexicanoSvc.AdvanceRound(w, sessionID, nextRound); err != nil {
			return
		}
	} else {
		if err := h.store.AdvanceRound(sessionID); err != nil {
			respondError(w, http.StatusInternalServerError, "server_error")
			return
		}
	}
	h.hub.Emit(sessionID, events.Envelope{Type: events.EventRoundUpdated})
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) getLeaderboard(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	sess, err := h.store.GetSession(id)
	if errors.Is(err, store.ErrNotFound) {
		respondError(w, http.StatusNotFound, "session_not_found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error")
		return
	}

	standings, err := h.store.GetLeaderboard(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error")
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
