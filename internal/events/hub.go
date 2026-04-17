package events

import (
	"log/slog"
	"sync"
)

type client struct {
	ch chan Envelope
}

// Hub manages per-session SSE client subscriptions.
type Hub struct {
	mu   sync.RWMutex
	subs map[string]map[*client]struct{}
}

func NewHub() *Hub {
	return &Hub{subs: make(map[string]map[*client]struct{})}
}

func (h *Hub) add(sessionID string, c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.subs[sessionID] == nil {
		h.subs[sessionID] = make(map[*client]struct{})
	}
	h.subs[sessionID][c] = struct{}{}
}

func (h *Hub) remove(sessionID string, c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if m := h.subs[sessionID]; m != nil {
		delete(m, c)
		if len(m) == 0 {
			delete(h.subs, sessionID)
		}
	}
	close(c.ch)
}

// Emit broadcasts an event to all clients subscribed to sessionID.
// It never blocks: slow clients drop the event.
func (h *Hub) Emit(sessionID string, evt Envelope) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.subs[sessionID] {
		select {
		case c.ch <- evt:
		default:
			slog.Warn("sse drop", "session", sessionID, "type", evt.Type)
		}
	}
}
