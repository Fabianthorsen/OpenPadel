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
	"github.com/fabianthorsen/openpadel/internal/scheduler"
	"github.com/fabianthorsen/openpadel/internal/store"
)

func (h *Handler) createSession(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Courts      int     `json:"courts"`
		Points      int     `json:"points"`
		Name        string  `json:"name"`
		GameMode    string  `json:"game_mode"`
		SetsToWin   int     `json:"sets_to_win"`
		GamesPerSet int     `json:"games_per_set"`
		ScheduledAt *string `json:"scheduled_at"`
		RoundsTotal *int    `json:"rounds_total"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.GameMode == "" {
		body.GameMode = "americano"
	}
	if body.GameMode != "americano" && body.GameMode != "mexicano" && body.GameMode != "tennis" {
		respondError(w, http.StatusBadRequest, "game_mode must be 'americano', 'mexicano', or 'tennis'")
		return
	}

	if body.GameMode == "americano" || body.GameMode == "mexicano" {
		if body.Courts < 1 || body.Courts > 4 {
			respondError(w, http.StatusBadRequest, "courts must be between 1 and 4")
			return
		}
		if body.Points != 16 && body.Points != 24 && body.Points != 32 {
			respondError(w, http.StatusBadRequest, "points must be 16, 24, or 32")
			return
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

	sess, err := h.store.CreateSession(body.Courts, body.Points, body.Name, body.GameMode, body.SetsToWin, body.GamesPerSet, body.RoundsTotal, scheduledAt)
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
		respondError(w, http.StatusNotFound, "session not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not load session")
		return
	}

	// Strip admin token from public responses.
	if !isAdmin(extractAdminToken(r), sess.AdminToken) {
		sess.AdminToken = ""
	}

	respond(w, http.StatusOK, sess)
}

func (h *Handler) startSession(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	sess, err := h.store.GetSession(id)
	if errors.Is(err, store.ErrNotFound) {
		respondError(w, http.StatusNotFound, "session not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not load session")
		return
	}
	if !isAdmin(extractAdminToken(r), sess.AdminToken) {
		respondError(w, http.StatusForbidden, "admin access required")
		return
	}
	if sess.Status != domain.StatusLobby {
		respondError(w, http.StatusConflict, "session already started")
		return
	}

	active := activePlayers(sess.Players)

	switch sess.GameMode {
	case "tennis":
		if err := h.startTennisSession(w, id, active); err != nil {
			return
		}
	case "mexicano":
		minPlayers := sess.Courts * 4
		if len(active) < minPlayers || len(active) < 8 {
			respondError(w, http.StatusUnprocessableEntity, "not enough players to start")
			return
		}
		if err := h.startMexicanoSession(w, id, sess, active); err != nil {
			return
		}
	default: // americano
		minPlayers := sess.Courts * 4
		if len(active) < minPlayers {
			respondError(w, http.StatusUnprocessableEntity, "not enough players to start")
			return
		}
		rand.Shuffle(len(active), func(i, j int) { active[i], active[j] = active[j], active[i] })
		totalRounds := scheduler.TotalRounds(len(active), sess.Courts)
		rounds := scheduler.Generate(active, sess.Courts, totalRounds)
		if err := h.store.SaveRounds(id, rounds); err != nil {
			respondError(w, http.StatusInternalServerError, "could not generate rounds")
			return
		}
		if err := h.store.StartSession(id, totalRounds); err != nil {
			respondError(w, http.StatusInternalServerError, "could not start session")
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
		respondError(w, http.StatusNotFound, "session not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not load session")
		return
	}
	if !isAdmin(extractAdminToken(r), sess.AdminToken) {
		respondError(w, http.StatusForbidden, "admin access required")
		return
	}
	if sess.Status == domain.StatusComplete {
		respondError(w, http.StatusConflict, "session already ended")
		return
	}
	if err := h.store.CompleteSession(id, true); err != nil {
		respondError(w, http.StatusInternalServerError, "could not close session")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) cancelSession(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	sess, err := h.store.GetSession(id)
	if errors.Is(err, store.ErrNotFound) {
		respondError(w, http.StatusNotFound, "session not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not load session")
		return
	}
	if !isAdmin(extractAdminToken(r), sess.AdminToken) {
		respondError(w, http.StatusForbidden, "admin access required")
		return
	}
	if err := h.store.DeleteSession(id); err != nil {
		respondError(w, http.StatusInternalServerError, "could not cancel session")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
