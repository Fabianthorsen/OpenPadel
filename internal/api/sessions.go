package api

import (
	"encoding/json"
	"errors"
	"math/rand/v2"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/fabianthorsen/openpadel/internal/domain"
	"github.com/fabianthorsen/openpadel/internal/events"
	"github.com/fabianthorsen/openpadel/internal/scheduler"
	"github.com/fabianthorsen/openpadel/internal/store"
)

func (h *Handler) createSession(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Courts               int     `json:"courts"`
		Points               int     `json:"points"`
		Name                 string  `json:"name"`
		GameMode             string  `json:"game_mode"`
		SetsToWin            int     `json:"sets_to_win"`
		GamesPerSet          int     `json:"games_per_set"`
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
	if body.GameMode != "americano" && body.GameMode != "mexicano" && body.GameMode != "timed_americano" && body.GameMode != "tennis" {
		respondError(w, http.StatusBadRequest, "game_mode must be 'americano', 'mexicano', 'timed_americano', or 'tennis'")
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
	} else {
		// Tennis: fixed 1 court, no points concept.
		body.Courts = 1
		body.Points = 0
		if body.SetsToWin != 2 && body.SetsToWin != 3 {
			body.SetsToWin = 2 // default best of 3
		}
		if body.GamesPerSet != 4 && body.GamesPerSet != 6 {
			body.GamesPerSet = 6 // default
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
	sess, err := h.store.CreateSession(body.Courts, body.Points, body.Name, body.GameMode, body.SetsToWin, body.GamesPerSet, body.RoundsTotal, scheduledAt, body.CourtDurationMinutes, body.TotalDurationMinutes, body.BufferSeconds, creatorUserID)
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
	case "tennis":
		if err := h.startTennisSession(w, id, active); err != nil {
			return
		}
	case "mexicano":
		required := sess.Courts * 4
		if len(active) != required {
			respondError(w, http.StatusUnprocessableEntity, "mexicano_player_count")
			return
		}
		if err := h.startMexicanoSession(w, id, sess, active, endsAt); err != nil {
			return
		}
	case "timed_americano":
		minPlayers := sess.Courts * 4
		if len(active) < minPlayers {
			respondError(w, http.StatusUnprocessableEntity, "not_enough_players")
			return
		}
		// Shuffle players
		rand.Shuffle(len(active), func(i, j int) {
			active[i], active[j] = active[j], active[i]
		})
		// Calculate rounds
		roundCount, roundDurationSec, err := scheduler.CalculateTimedRounds(len(active), *sess.TotalDurationMinutes, *sess.BufferSeconds)
		if err != nil {
			respondError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		// Generate rounds
		rounds, err := scheduler.GenerateTimedAmericano(active, sess.Courts, roundCount)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "server_error")
			return
		}
		// Save rounds
		if err := h.store.SaveRounds(id, rounds); err != nil {
			respondError(w, http.StatusInternalServerError, "server_error")
			return
		}
		// Calculate end time
		now := time.Now().UTC()
		endsAt := now.Add(time.Duration(*sess.TotalDurationMinutes) * time.Minute)
		// Start session
		if err := h.store.StartTimedAmericanoSession(id, string(domain.StatusActive), roundCount, sess.TotalDurationMinutes, sess.BufferSeconds, &roundDurationSec, &endsAt); err != nil {
			respondError(w, http.StatusInternalServerError, "server_error")
			return
		}
		// Set round started time
		if err := h.store.SetRoundStartedAt(id, &now); err != nil {
			respondError(w, http.StatusInternalServerError, "server_error")
			return
		}
	default: // americano
		minPlayers := sess.Courts * 4
		if len(active) < minPlayers {
			respondError(w, http.StatusUnprocessableEntity, "not_enough_players")
			return
		}
		rand.Shuffle(len(active), func(i, j int) { active[i], active[j] = active[j], active[i] })
		totalRounds := scheduler.TotalRounds(len(active), sess.Courts)
		rounds := scheduler.Generate(active, sess.Courts, totalRounds)
		if err := h.store.SaveRounds(id, rounds); err != nil {
			respondError(w, http.StatusInternalServerError, "server_error")
			return
		}
		if err := h.store.StartSession(id, totalRounds, endsAt); err != nil {
			respondError(w, http.StatusInternalServerError, "server_error")
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
