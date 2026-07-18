package mexicano

import (
	"encoding/json"
	"net/http"
	"sort"
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

// Start generates round 1 and activates the session. Round 1 is seeded by rating
// (descending) rather than randomly: mexicano.GenerateRound pairs rank 1+4 vs 2+3,
// so a rating-descending seed makes the opening round balanced by skill instead of
// a coin-flip. Rounds 2+ are driven by live standings (see AdvanceRound) and are
// unaffected. Per ADR 0006.
// Returns a non-nil error only if it has already written an HTTP error response.
func (s *Service) Start(w http.ResponseWriter, sessionID string, sess *domain.Session, active []domain.Player, endsAt *time.Time) error {
	seeded := make([]domain.Player, len(active))
	copy(seeded, active)
	// Stable sort by normalised rating descending; an all-equal (e.g. all-median
	// or unrated) field keeps its incoming order, so this is a no-op there.
	sort.SliceStable(seeded, func(i, j int) bool {
		return domain.NormalizeRating(seeded[i].Rating) > domain.NormalizeRating(seeded[j].Rating)
	})
	standings := make([]domain.Standing, len(seeded))
	for i, p := range seeded {
		standings[i] = domain.Standing{Rank: i + 1, PlayerID: p.ID, Name: p.Name}
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
