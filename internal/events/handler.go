package events

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

// ServeSSE returns an http.HandlerFunc that streams SSE events for a session.
// The endpoint is unauthenticated (session data is already public).
func (h *Hub) ServeSSE() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := chi.URLParam(r, "id")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache, no-transform")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no")

		c := &client{ch: make(chan Envelope, 16)}
		h.add(sessionID, c)
		defer h.remove(sessionID, c)

		// Prime past any buffering proxy.
		fmt.Fprint(w, ": ok\n\n")
		flusher.Flush()

		tick := time.NewTicker(20 * time.Second)
		defer tick.Stop()

		for {
			select {
			case <-r.Context().Done():
				return
			case <-tick.C:
				fmt.Fprint(w, ": ping\n\n")
				flusher.Flush()
			case evt, open := <-c.ch:
				if !open {
					return
				}
				b, _ := json.Marshal(evt)
				fmt.Fprintf(w, "event: %s\ndata: %s\n\n", evt.Type, b)
				flusher.Flush()
			}
		}
	}
}
