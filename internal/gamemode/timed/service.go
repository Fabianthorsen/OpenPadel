package timed

import (
	"encoding/json"
	"math/rand/v2"
	"net/http"
	"time"

	"github.com/fabianthorsen/openpadel/internal/domain"
	"github.com/fabianthorsen/openpadel/internal/events"
)

// Store is the subset of store.Store methods used by this service.
type Store interface {
	SaveRounds(sessionID string, rounds []domain.Round) error
	StartTimedAmericanoSession(id, status string, roundsTotal int, totalDurationMin, roundDurationSec *int, endsAt *time.Time) error
	SetRoundStartedAt(id string, roundStartedAt *time.Time) error
	UpdateRoundDuration(id string, roundDurationSec *int) error
	AdvanceRound(id string) error
}

// Service orchestrates Timed Americano session start and round advancement.
type Service struct {
	store Store
	hub   *events.Hub
}

func New(store Store, hub *events.Hub) *Service {
	return &Service{store: store, hub: hub}
}

// Start shuffles players, calculates round count/duration, generates all rounds, and activates the session.
// Returns a non-nil error only if it has already written an HTTP error response.
func (s *Service) Start(w http.ResponseWriter, sessionID string, sess *domain.Session, active []domain.Player) error {
	rand.Shuffle(len(active), func(i, j int) { active[i], active[j] = active[j], active[i] })

	roundCount, roundDurationSec, err := CalculateTimedRounds(len(active), *sess.TotalDurationMinutes)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return err
	}

	rounds, err := GenerateTimedAmericano(active, sess.Courts, roundCount)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "server_error")
		return err
	}

	if err := s.store.SaveRounds(sessionID, rounds); err != nil {
		writeError(w, http.StatusInternalServerError, "server_error")
		return err
	}

	now := time.Now().UTC()
	endsAt := now.Add(time.Duration(*sess.TotalDurationMinutes) * time.Minute)
	if err := s.store.StartTimedAmericanoSession(sessionID, string(domain.StatusActive), roundCount, sess.TotalDurationMinutes, &roundDurationSec, &endsAt); err != nil {
		writeError(w, http.StatusInternalServerError, "server_error")
		return err
	}

	if err := s.store.SetRoundStartedAt(sessionID, &now); err != nil {
		writeError(w, http.StatusInternalServerError, "server_error")
		return err
	}
	return nil
}

// AdvanceRound recalculates the round duration for remaining rounds, updates the DB, and advances.
// Returns a non-nil error only if it has already written an HTTP error response.
func (s *Service) AdvanceRound(w http.ResponseWriter, sessionID string, sess *domain.Session) error {
	now := time.Now().UTC()
	remainingSeconds := int(sess.EndsAt.Sub(now).Seconds())
	if remainingSeconds < 0 {
		remainingSeconds = 0
	}
	remainingRounds := *sess.RoundsTotal - *sess.CurrentRound

	newDurationSec := RecalculateRoundDuration(remainingRounds, remainingSeconds)

	if err := s.store.UpdateRoundDuration(sessionID, &newDurationSec); err != nil {
		writeError(w, http.StatusInternalServerError, "server_error")
		return err
	}
	if err := s.store.SetRoundStartedAt(sessionID, &now); err != nil {
		writeError(w, http.StatusInternalServerError, "server_error")
		return err
	}

	s.hub.Emit(sessionID, events.Envelope{
		Type: events.EventTimerSync,
		Payload: map[string]any{
			"round_duration_seconds": newDurationSec,
			"round_started_at":       now.Format(time.RFC3339),
			"remaining_rounds":       remainingRounds,
		},
	})

	if err := s.store.AdvanceRound(sessionID); err != nil {
		writeError(w, http.StatusInternalServerError, "server_error")
		return err
	}
	return nil
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg}) //nolint:errcheck
}
