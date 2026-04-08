package api

import (
	"math/rand/v2"
	"net/http"
	"time"

	"github.com/fabianthorsen/openpadel/internal/domain"
	"github.com/fabianthorsen/openpadel/internal/scheduler"
)

// startMexicanoSession generates round 1 (random order), saves it, and activates the session.
// Returns a non-nil error only if it has already written an error response.
func (h *Handler) startMexicanoSession(w http.ResponseWriter, sessionID string, sess *domain.Session, active []domain.Player, endsAt *time.Time) error {
	// Round 1 uses random standings (points all zero).
	standings := make([]domain.Standing, len(active))
	for i, p := range active {
		standings[i] = domain.Standing{
			Rank:     i + 1,
			PlayerID: p.ID,
			Name:     p.Name,
		}
	}
	rand.Shuffle(len(standings), func(i, j int) { standings[i], standings[j] = standings[j], standings[i] })
	// Re-assign ranks after shuffle so GenerateMexicanoRound uses correct ordering.
	for i := range standings {
		standings[i].Rank = i + 1
	}

	round := scheduler.GenerateMexicanoRound(standings, sess.Courts, 1)
	round.SessionID = sessionID

	if err := h.store.SaveRounds(sessionID, []domain.Round{round}); err != nil {
		respondError(w, http.StatusInternalServerError, "could not save round")
		return err
	}
	if err := h.store.StartMexicanoSession(sessionID, endsAt); err != nil {
		respondError(w, http.StatusInternalServerError, "could not start session")
		return err
	}
	return nil
}

// advanceMexicanoRound computes standings, generates the next round, and saves it.
// Returns a non-nil error only if it has already written an error response.
func (h *Handler) advanceMexicanoRound(w http.ResponseWriter, sessionID string, nextRoundNum int) error {
	standings, err := h.store.GetLeaderboard(sessionID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not compute standings")
		return err
	}

	sess, err := h.store.GetSession(sessionID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not load session")
		return err
	}

	round := scheduler.GenerateMexicanoRound(standings, sess.Courts, nextRoundNum)
	round.SessionID = sessionID

	if err := h.store.AdvanceMexicanoRound(sessionID, round); err != nil {
		respondError(w, http.StatusInternalServerError, "could not save next round")
		return err
	}
	return nil
}
