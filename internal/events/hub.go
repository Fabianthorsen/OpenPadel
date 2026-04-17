package events

import (
	"log/slog"
	"sync"
)

const (
	sessionKeyPrefix = "s:"
	userKeyPrefix    = "u:"
)

type client struct {
	ch chan Envelope
}

// Hub manages SSE client subscriptions keyed by session or user.
type Hub struct {
	mu   sync.RWMutex
	subs map[string]map[*client]struct{}
}

func NewHub() *Hub {
	return &Hub{subs: make(map[string]map[*client]struct{})}
}

func (h *Hub) add(key string, c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.subs[key] == nil {
		h.subs[key] = make(map[*client]struct{})
	}
	h.subs[key][c] = struct{}{}
}

func (h *Hub) remove(key string, c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if m := h.subs[key]; m != nil {
		delete(m, c)
		if len(m) == 0 {
			delete(h.subs, key)
		}
	}
	close(c.ch)
}

func (h *Hub) emit(key string, evt Envelope) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.subs[key] {
		select {
		case c.ch <- evt:
		default:
			slog.Warn("sse drop", "key", key, "type", evt.Type)
		}
	}
}

// Emit broadcasts an event to all clients subscribed to sessionID.
// It never blocks: slow clients drop the event.
func (h *Hub) Emit(sessionID string, evt Envelope) {
	h.emit(sessionKeyPrefix+sessionID, evt)
}

// EmitToUser broadcasts an event to all SSE clients for a given user ID.
// It never blocks: slow clients drop the event.
func (h *Hub) EmitToUser(userID string, evt Envelope) {
	h.emit(userKeyPrefix+userID, evt)
}
