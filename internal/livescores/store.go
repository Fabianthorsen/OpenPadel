package livescores

import "sync"

// Score holds the in-progress live score for a single match.
type Score struct {
	A      int
	B      int
	Server string // "a" or "b"
}

// Store is a concurrency-safe in-memory map of matchID → live score.
// Multiple goroutines may read simultaneously (RLock); writes are exclusive (Lock).
// Nothing is persisted — scores reset if the server restarts, which is fine since
// only in-progress tap state lives here; finalised scores are in the database.
type Store struct {
	mu     sync.RWMutex
	scores map[string]Score
}

func New() *Store {
	return &Store{scores: make(map[string]Score)}
}

func (s *Store) Set(matchID, server string, a, b int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.scores[matchID] = Score{A: a, B: b, Server: server}
}

func (s *Store) Get(matchID string) (Score, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sc, ok := s.scores[matchID]
	return sc, ok
}

func (s *Store) Clear(matchID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.scores, matchID)
}
