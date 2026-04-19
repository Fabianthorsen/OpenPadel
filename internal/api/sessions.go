package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/fabianthorsen/openpadel/internal/domain"
	"github.com/fabianthorsen/openpadel/internal/events"
	"github.com/fabianthorsen/openpadel/internal/store"
)

func (h *Handler) createSession(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Courts               int     `json:"courts"`
		Points               int     `json:"points"`
		Name                 string  `json:"name"`
		GameMode             string  `json:"game_mode"`
		ScheduledAt          *string `json:"scheduled_at"`
		RoundsTotal          *int    `json:"rounds_total"`
		CourtDurationMinutes *int    `json:"court_duration_minutes"`
		TotalDurationMinutes *int    `json:"total_duration_minutes"`
		BufferSeconds        *int    `json:"buffer_seconds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request_body")
		return
	}
	if body.GameMode == "" {
		body.GameMode = "americano"
	}
	if body.GameMode != "americano" && body.GameMode != "mexicano" && body.GameMode != "timed_americano" {
		respondError(w, http.StatusBadRequest, "game_mode must be 'americano', 'mexicano', or 'timed_americano'")
		return
	}

	if body.GameMode == "americano" || body.GameMode == "mexicano" || body.GameMode == "timed_americano" {
		minCourts := 1
		if body.GameMode == "mexicano" {
			minCourts = 2
		}
		if body.Courts < minCourts || body.Courts > 4 {
			respondError(w, http.StatusBadRequest, "courts must be between 1 and 4 for Americano/Timed Americano, 2 and 4 for Mexicano")
			return
		}

		// Timed Americano: no points, duration-based
		if body.GameMode == "timed_americano" {
			// Points must be 0 or omitted
			if body.Points != 0 {
				respondError(w, http.StatusBadRequest, "points must be 0 for timed_americano")
				return
			}
			// total_duration_minutes required
			if body.TotalDurationMinutes == nil || *body.TotalDurationMinutes < 15 || *body.TotalDurationMinutes > 300 {
				respondError(w, http.StatusBadRequest, "total_duration_minutes required for timed_americano, must be between 15 and 300")
				return
			}
			// buffer_seconds optional, default 120, must be 60-300 if provided
			if body.BufferSeconds == nil {
				defaultBuffer := 120
				body.BufferSeconds = &defaultBuffer
			} else if *body.BufferSeconds < 60 || *body.BufferSeconds > 300 {
				respondError(w, http.StatusBadRequest, "buffer_seconds must be between 60 and 300")
				return
			}
		} else {
			// Americano/Mexicano: require points
			if body.Points != 16 && body.Points != 24 && body.Points != 32 {
				respondError(w, http.StatusBadRequest, "points must be 16, 24, or 32")
				return
			}
		}
	}

	var scheduledAt *time.Time
	if body.ScheduledAt != nil && *body.ScheduledAt != "" {
		t, err := time.Parse(time.RFC3339, *body.ScheduledAt)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid scheduled_at format, use RFC3339")
			return
		}
		scheduledAt = &t
	}

	// Validate Mexicano preset rounds.
	if body.GameMode == "mexicano" && body.RoundsTotal != nil {
		if *body.RoundsTotal < 1 || *body.RoundsTotal > 20 {
			respondError(w, http.StatusBadRequest, "rounds_total must be between 1 and 20")
			return
		}
	}

	// Validate court duration.
	if body.CourtDurationMinutes != nil && (*body.CourtDurationMinutes < 15 || *body.CourtDurationMinutes > 300) {
		respondError(w, http.StatusBadRequest, "court_duration_minutes must be between 15 and 300")
		return
	}

	creatorUserID := ""
	if u := userFromContext(r); u != nil {
		creatorUserID = u.ID
	}
	sess, err := h.store.CreateSession(body.Courts, body.Points, body.Name, body.GameMode, body.RoundsTotal, scheduledAt, body.CourtDurationMinutes, body.TotalDurationMinutes, body.BufferSeconds, creatorUserID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not create session")
		return
	}
	respond(w, http.StatusCreated, sess)
}

func (h *Handler) getSession(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	sess, err := h.store.GetSession(id)
	if errors.Is(err, store.ErrNotFound) {
		respondError(w, http.StatusNotFound, "session_not_found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error")
		return
	}

	// Treat the logged-in creator the same as a token-holding admin.
	u := userFromContext(r)
	if u != nil && sess.CreatorUserID != "" && u.ID == sess.CreatorUserID {
		sess.IsCreator = true
	} else if !isAdmin(extractAdminToken(r), sess.AdminToken) {
		sess.AdminToken = ""
	}

	respond(w, http.StatusOK, sess)
}

func (h *Handler) startSession(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	sess, err := h.store.GetSession(id)
	if errors.Is(err, store.ErrNotFound) {
		respondError(w, http.StatusNotFound, "session_not_found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error")
		return
	}
	if !isAdmin(extractAdminToken(r), sess.AdminToken) {
		respondError(w, http.StatusForbidden, "admin_required")
		return
	}
	if sess.Status != domain.StatusLobby {
		respondError(w, http.StatusConflict, "session_already_started")
		return
	}

	active := activePlayers(sess.Players)

	// Compute ends_at from court_duration_minutes if set.
	var endsAt *time.Time
	if sess.CourtDurationMinutes != nil && *sess.CourtDurationMinutes > 0 {
		t := time.Now().UTC().Add(time.Duration(*sess.CourtDurationMinutes) * time.Minute)
		endsAt = &t
	}

	switch sess.GameMode {
	case "mexicano":
		required := sess.Courts * 4
		if len(active) != required {
			respondError(w, http.StatusUnprocessableEntity, "mexicano_player_count")
			return
		}
		if err := h.mexicanoSvc.Start(w, id, sess, active, endsAt); err != nil {
			return
		}
	case "timed_americano":
		minPlayers := sess.Courts * 4
		maxPlayers := sess.Courts * 8
		if len(active) < minPlayers {
			respondError(w, http.StatusUnprocessableEntity, "not_enough_players")
			return
		}
		if len(active) > maxPlayers {
			respondError(w, http.StatusUnprocessableEntity, "too_many_players")
			return
		}
		if err := h.timedSvc.Start(w, id, sess, active); err != nil {
			return
		}
	default: // americano
		minPlayers := sess.Courts * 4
		if len(active) < minPlayers {
			respondError(w, http.StatusUnprocessableEntity, "not_enough_players")
			return
		}
		if err := h.americanoSvc.Start(w, id, sess, active, endsAt); err != nil {
			return
		}
	}

	sess, _ = h.store.GetSession(id)

	// Fan out push notifications to all subscribed players in the session.
	adminName := ""
	for _, p := range sess.Players {
		if p.ID == sess.CreatorPlayerID {
			adminName = p.Name
			break
		}
	}
	if adminName == "" {
		adminName = "Admin"
	}
	name := playerShortName(adminName)
	tournamentName := sess.Name
	var notifBody string
	if tournamentName != "" {
		notifBody = name + " just started \"" + tournamentName + "\", tap to watch scores!"
	} else {
		notifBody = name + " just started the tournament, tap to watch scores!"
	}
	go h.sendPushToSession(id, "Tournament started!", notifBody)

	sess.AdminToken = ""
	h.hub.Emit(id, events.Envelope{Type: events.EventSessionUpdated})
	respond(w, http.StatusOK, sess)
}

// playerShortName returns "Firstname L." for multi-word names, or the name as-is.
func playerShortName(name string) string {
	words := strings.Fields(name)
	if len(words) <= 1 {
		return name
	}
	last := words[len(words)-1]
	return words[0] + " " + strings.ToUpper(string([]rune(last)[0])) + "."
}

func activePlayers(players []domain.Player) []domain.Player {
	var out []domain.Player
	for _, p := range players {
		if p.Active {
			out = append(out, p)
		}
	}
	return out
}

func (h *Handler) closeSession(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	sess, err := h.store.GetSession(id)
	if errors.Is(err, store.ErrNotFound) {
		respondError(w, http.StatusNotFound, "session_not_found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error")
		return
	}
	if !isAdmin(extractAdminToken(r), sess.AdminToken) {
		respondError(w, http.StatusForbidden, "admin_required")
		return
	}
	if sess.Status == domain.StatusComplete {
		respondError(w, http.StatusConflict, "session_already_ended")
		return
	}
	if err := h.store.CompleteSession(id, true); err != nil {
		respondError(w, http.StatusInternalServerError, "server_error")
		return
	}

	standings, err := h.store.GetLeaderboard(id)
	if err == nil && len(standings) > 0 && standings[0].UserID != nil {
		h.store.IncrementTournamentWinCount(*standings[0].UserID)
	}

	h.hub.Emit(id, events.Envelope{Type: events.EventSessionUpdated})
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) cancelSession(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	sess, err := h.store.GetSession(id)
	if errors.Is(err, store.ErrNotFound) {
		respondError(w, http.StatusNotFound, "session_not_found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error")
		return
	}
	if !isAdmin(extractAdminToken(r), sess.AdminToken) {
		respondError(w, http.StatusForbidden, "admin_required")
		return
	}
	if err := h.store.DeleteSession(id); err != nil {
		respondError(w, http.StatusInternalServerError, "server_error")
		return
	}
	h.hub.Emit(id, events.Envelope{Type: events.EventSessionUpdated})
	w.WriteHeader(http.StatusNoContent)
}
