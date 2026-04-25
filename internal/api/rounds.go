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
		respondAPIError(w, ErrSessionNotFound)
		return
	}
	if err != nil {
		respondAPIError(w, ErrServerError)
		return
	}
	if sess.Status == domain.StatusLobby {
		respondAPIError(w, ErrSessionNotStarted)
		return
	}

	rounds, err := h.store.GetRounds(id)
	if err != nil {
		respondAPIError(w, ErrServerError)
		return
	}
	respond(w, http.StatusOK, map[string]any{"rounds": rounds})
}

func (h *Handler) getCurrentRound(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	round, err := h.store.GetCurrentRound(id)
	if errors.Is(err, store.ErrNotFound) {
		respondAPIError(w, ErrNoActiveRound)
		return
	}
	if err != nil {
		respondAPIError(w, ErrServerError)
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
		respondAPIError(w, ErrSessionNotFound)
		return
	}
	if err != nil {
		respondAPIError(w, ErrServerError)
		return
	}
	if sess.Status != domain.StatusPlaying {
		respondAPIError(w, ErrSessionNotActive)
		return
	}

	var body struct {
		ScoreA int `json:"score_a"`
		ScoreB int `json:"score_b"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondAPIError(w, ErrInvalidRequestBody)
		return
	}
	if body.ScoreA < 0 || body.ScoreB < 0 {
		respondAPIError(w, ErrScoresNegative)
		return
	}
	if body.ScoreA+body.ScoreB != sess.Points {
		respondAPIError(w, ErrScoresInvalidSum)
		return
	}

	match, err := h.store.UpdateScore(matchID, body.ScoreA, body.ScoreB)
	if errors.Is(err, store.ErrNotFound) {
		respondAPIError(w, ErrMatchNotFound)
		return
	}
	if err != nil {
		respondAPIError(w, ErrServerError)
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
		respondAPIError(w, ErrInvalidRequestBody)
		return
	}
	if body.A < 0 || body.B < 0 {
		respondAPIError(w, ErrScoresNegative)
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
		respondAPIError(w, ErrSessionNotFound)
		return
	}
	if err != nil {
		respondAPIError(w, ErrServerError)
		return
	}
	if !isAdmin(extractAdminToken(r), sess.AdminToken) {
		respondAPIError(w, ErrAdminRequired)
		return
	}
	if sess.Status != domain.StatusPlaying {
		respondAPIError(w, ErrSessionNotActive)
		return
	}
	if sess.EndsAt != nil && time.Now().UTC().After(*sess.EndsAt) {
		respondAPIError(w, ErrTournamentExpired)
		return
	}
	allScored, err := h.store.CurrentRoundAllScored(sessionID)
	if err != nil {
		respondAPIError(w, ErrServerError)
		return
	}
	if !allScored {
		respondAPIError(w, ErrRoundNotComplete)
		return
	}

	if sess.GameMode == "mexicano" {
		nextRound := 1
		if sess.CurrentRound != nil {
			nextRound = *sess.CurrentRound + 1
		}
		if sess.RoundsTotal != nil && nextRound > *sess.RoundsTotal {
			respondAPIError(w, ErrRoundLimitReached)
			return
		}
		if err := h.mexicanoSvc.AdvanceRound(w, sessionID, nextRound); err != nil {
			return
		}
	} else {
		if err := h.store.AdvanceRound(sessionID); err != nil {
			respondAPIError(w, ErrServerError)
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
		respondAPIError(w, ErrSessionNotFound)
		return
	}
	if err != nil {
		respondAPIError(w, ErrServerError)
		return
	}

	standings, err := h.store.GetLeaderboard(id)
	if err != nil {
		respondAPIError(w, ErrServerError)
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
