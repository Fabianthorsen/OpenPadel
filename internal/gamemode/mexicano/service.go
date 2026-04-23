package mexicano

import (
	"encoding/json"
	"math/rand/v2"
	"net/http"
	"time"

	"github.com/fabianthorsen/openpadel/internal/domain"
	"github.com/fabianthorsen/openpadel/internal/pairing/mexicano"
)

// Store is the subset of store.Store methods used by this service.
type Store interface {
	GetLeaderboard(sessionID string) ([]domain.Standing, error)
	GetSession(id string) (*domain.Session, error)
	SaveRounds(sessionID string, rounds []domain.Round) error
	StartMexicanoSession(id string, endsAt *time.Time) error
	AdvanceMexicanoRound(sessionID string, round domain.Round) error
}

// Service orchestrates Mexicano session start and round advancement.
type Service struct {
	store Store
}

func New(store Store) *Service {
	return &Service{store: store}
}

// Start generates round 1 (random order), saves it, and activates the session.
// Returns a non-nil error only if it has already written an HTTP error response.
func (s *Service) Start(w http.ResponseWriter, sessionID string, sess *domain.Session, active []domain.Player, endsAt *time.Time) error {
	standings := make([]domain.Standing, len(active))
	for i, p := range active {
		standings[i] = domain.Standing{Rank: i + 1, PlayerID: p.ID, Name: p.Name}
	}
	rand.Shuffle(len(standings), func(i, j int) { standings[i], standings[j] = standings[j], standings[i] })
	for i := range standings {
		standings[i].Rank = i + 1
	}

	round := mexicano.GenerateRound(standings, sess.Courts, 1)
	round.SessionID = sessionID

	if err := s.store.SaveRounds(sessionID, []domain.Round{round}); err != nil {
		writeError(w, http.StatusInternalServerError, "could not save round")
		return err
	}
	if err := s.store.StartMexicanoSession(sessionID, endsAt); err != nil {
		writeError(w, http.StatusInternalServerError, "could not start session")
		return err
	}
	return nil
}

// AdvanceRound computes standings, generates the next round, and saves it.
// Returns a non-nil error only if it has already written an HTTP error response.
func (s *Service) AdvanceRound(w http.ResponseWriter, sessionID string, nextRoundNum int) error {
	standings, err := s.store.GetLeaderboard(sessionID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not compute standings")
		return err
	}
	sess, err := s.store.GetSession(sessionID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load session")
		return err
	}
	round := mexicano.GenerateRound(standings, sess.Courts, nextRoundNum)
	round.SessionID = sessionID
	if err := s.store.AdvanceMexicanoRound(sessionID, round); err != nil {
		writeError(w, http.StatusInternalServerError, "could not save next round")
		return err
	}
	return nil
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg}) //nolint:errcheck
}
