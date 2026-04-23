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
	AllRoundsComplete(sessionID string) (bool, error)
}

// Service orchestrates Americano session start and round completion checks.
type Service struct {
	store Store
}

func New(store Store) *Service {
	return &Service{store: store}
}

// Start shuffles active players, generates all rounds, saves them, and activates the session.
// Returns a non-nil error only if it has already written an HTTP error response.
func (s *Service) Start(w http.ResponseWriter, sessionID string, sess *domain.Session, active []domain.Player, endsAt *time.Time) error {
	rand.Shuffle(len(active), func(i, j int) { active[i], active[j] = active[j], active[i] })
	totalRounds := americano.TotalRounds(len(active), sess.Courts)
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

// CanComplete returns true when all pre-generated rounds have been fully scored.
func (s *Service) CanComplete(sessionID string) (bool, error) {
	return s.store.AllRoundsComplete(sessionID)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg}) //nolint:errcheck
}
