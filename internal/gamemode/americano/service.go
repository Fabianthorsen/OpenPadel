package americano

import (
	"encoding/json"
	"math/rand/v2"
	"net/http"
	"time"

	"github.com/fabianthorsen/openpadel/internal/domain"
	"github.com/fabianthorsen/openpadel/internal/pairing/americano"
)

// Store is the subset of store.Store methods used by this service.
type Store interface {
	SaveRounds(sessionID string, rounds []domain.Round) error
	StartSession(id string, roundsTotal int, endsAt *time.Time) error
	StartAmericanoSession(id string, endsAt *time.Time) error
	AllRoundsComplete(sessionID string) (bool, error)
	AdvanceAmericanoRound(sessionID string, round domain.Round) error
	GetRounds(sessionID string) ([]domain.Round, error)
}

// Service orchestrates Americano session start and round completion checks.
type Service struct {
	store Store
}

func New(store Store) *Service {
	return &Service{store: store}
}

// Start shuffles active players and generates rounds based on session mode.
// Fixed mode: generates all N rounds upfront, saves them, activates with N.
// Unlimited mode: generates only round 1, saves it, activates while preserving null rounds_total.
// If RoundsTotal is nil (not explicitly set during PATCH), calculates default from player count.
// Returns a non-nil error only if it has already written an HTTP error response.
func (s *Service) Start(w http.ResponseWriter, sessionID string, sess *domain.Session, active []domain.Player, endsAt *time.Time) error {
	rand.Shuffle(len(active), func(i, j int) { active[i], active[j] = active[j], active[i] })

	// If RoundsTotal is nil, calculate it from player count (treats as fixed mode by default)
	roundsTotal := sess.RoundsTotal
	if roundsTotal == nil {
		calculated := americano.TotalRounds(len(active), sess.Courts)
		roundsTotal = &calculated
	}

	// Fixed mode: generate all rounds upfront
	if roundsTotal != nil {
		totalRounds := *roundsTotal
		rounds := americano.GenerateRounds(active, sess.Courts, totalRounds)
		if err := s.store.SaveRounds(sessionID, rounds); err != nil {
			writeError(w, http.StatusInternalServerError, "server_error")
			return err
		}
		if err := s.store.StartSession(sessionID, totalRounds, endsAt); err != nil {
			writeError(w, http.StatusInternalServerError, "server_error")
			return err
		}
		return nil
	}

	// Unlimited mode: generate only round 1, preserve null rounds_total
	// (This branch shouldn't be reached now, but keeping for clarity)
	rounds := americano.GenerateRounds(active, sess.Courts, 1)
	if err := s.store.SaveRounds(sessionID, rounds); err != nil {
		writeError(w, http.StatusInternalServerError, "server_error")
		return err
	}
	if err := s.store.StartAmericanoSession(sessionID, endsAt); err != nil {
		writeError(w, http.StatusInternalServerError, "server_error")
		return err
	}
	return nil
}

// CanComplete returns true when all pre-generated rounds have been fully scored.
func (s *Service) CanComplete(sessionID string) (bool, error) {
	return s.store.AllRoundsComplete(sessionID)
}

// AdvanceAmericanoRound generates the next round for unlimited Americano sessions.
// Fetches previously-played rounds from DB, generates the next round using streaming
// generator, and saves it atomically with session state update.
func (s *Service) AdvanceAmericanoRound(sessionID string, previousRounds []domain.Round, players []domain.Player, courts int) error {
	// Generate next round from previous rounds using streaming generator
	nextRound := americano.GenerateNextRound(previousRounds, players, courts)

	// Save atomically with current_round update
	return s.store.AdvanceAmericanoRound(sessionID, *nextRound)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg}) //nolint:errcheck
}
