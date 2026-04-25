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
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondAPIError(w, ErrInvalidRequestBody)
		return
	}

	var scheduledAt *time.Time
	if body.ScheduledAt != nil && *body.ScheduledAt != "" {
		t, err := time.Parse(time.RFC3339, *body.ScheduledAt)
		if err != nil {
			respondAPIError(w, ErrInvalidScheduledAt)
			return
		}
		scheduledAt = &t
	}

	gameMode := domain.GameMode(body.GameMode)
	if gameMode == "" {
		gameMode = domain.ModeAmericano
	}

	input := domain.SessionInput{
		Courts:               body.Courts,
		Points:               body.Points,
		Name:                 body.Name,
		GameMode:             gameMode,
		RoundsTotal:          body.RoundsTotal,
		ScheduledAt:          scheduledAt,
		CourtDurationMinutes: body.CourtDurationMinutes,
	}

	validationErrs := input.Validate()
	if len(validationErrs) > 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"validation_errors": validationErrs,
		}) //nolint:errcheck
		return
	}

	creatorUserID := ""
	if u := userFromContext(r); u != nil {
		creatorUserID = u.ID
	}
	sess, err := h.store.CreateSession(input, creatorUserID)
	if err != nil {
		respondAPIError(w, ErrCouldNotCreateSession)
		return
	}
	respond(w, http.StatusCreated, sess)
}

func (h *Handler) getSession(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	sess, err := h.store.GetSession(id)
	if errors.Is(err, store.ErrNotFound) {
		respondAPIError(w, ErrSessionNotFound)
		return
	}
	if err != nil {
		respondAPIError(w, ErrServerError)
		return
	}

	// Compute validation errors for the current session state.
	if sess.Status == domain.StatusLobby {
		switch sess.GameMode {
		case "mexicano":
			sess.ValidationErrors = domain.MexicanoConstraints(sess.Courts, len(activePlayers(sess.Players)))
		default: // americano
			sess.ValidationErrors = domain.AmericanoConstraints(sess.Courts, len(activePlayers(sess.Players)))
		}
		sess.CanStart = len(sess.ValidationErrors) == 0
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
		respondAPIError(w, ErrSessionNotFound)
		return
	}
	if err != nil {
		respondAPIError(w, ErrServerError)
		return
	}
	if !isAdmin(extractAdminToken(r), sess.AdminToken) {
		respondAPIError(w, ErrAdminRequired)
		return
	}
	if sess.Status != domain.StatusLobby {
		respondAPIError(w, ErrSessionAlreadyStarted)
		return
	}

	active := activePlayers(sess.Players)

	// Validate constraints before starting.
	var validationErrs []domain.ValidationError
	switch sess.GameMode {
	case "mexicano":
		validationErrs = domain.MexicanoConstraints(sess.Courts, len(active))
	default: // americano
		validationErrs = domain.AmericanoConstraints(sess.Courts, len(active))
	}
	if len(validationErrs) > 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"validation_errors": validationErrs,
		}) //nolint:errcheck
		return
	}

	// Compute ends_at from court_duration_minutes if set.
	var endsAt *time.Time
	if sess.CourtDurationMinutes != nil && *sess.CourtDurationMinutes > 0 {
		t := time.Now().UTC().Add(time.Duration(*sess.CourtDurationMinutes) * time.Minute)
		endsAt = &t
	}

	switch sess.GameMode {
	case "mexicano":
		if err := h.mexicanoSvc.Start(w, id, sess, active, endsAt); err != nil {
			return
		}
	default: // americano
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

func (h *Handler) updateSession(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	sess, err := h.store.GetSession(id)
	if errors.Is(err, store.ErrNotFound) {
		respondAPIError(w, ErrSessionNotFound)
		return
	}
	if err != nil {
		respondAPIError(w, ErrServerError)
		return
	}
	if !isAdmin(extractAdminToken(r), sess.AdminToken) {
		respondAPIError(w, ErrAdminRequired)
		return
	}
	if sess.Status != domain.StatusLobby {
		respondAPIError(w, ErrSessionAlreadyStarted)
		return
	}

	var body struct {
		Name        *string `json:"name"`
		GameMode    *string `json:"game_mode"`
		Courts      *int    `json:"courts"`
		Points      *int    `json:"points"`
		RoundsTotal *int    `json:"rounds_total"`
		ScheduledAt *string `json:"scheduled_at"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondAPIError(w, ErrInvalidRequestBody)
		return
	}

	// Apply partial update to current values.
	input := domain.SessionInput{
		Name:        sess.Name,
		GameMode:    sess.GameMode,
		Courts:      sess.Courts,
		Points:      sess.Points,
		RoundsTotal: sess.RoundsTotal,
		ScheduledAt: sess.ScheduledAt,
	}
	if body.Name != nil {
		input.Name = *body.Name
	}
	if body.GameMode != nil {
		input.GameMode = domain.GameMode(*body.GameMode)
	}
	if body.Courts != nil {
		input.Courts = *body.Courts
	}
	if body.Points != nil {
		input.Points = *body.Points
	}
	if body.RoundsTotal != nil {
		input.RoundsTotal = body.RoundsTotal
	}
	if body.ScheduledAt != nil {
		if *body.ScheduledAt == "" {
			input.ScheduledAt = nil
		} else {
			t, err := time.Parse(time.RFC3339, *body.ScheduledAt)
			if err != nil {
				respondAPIError(w, ErrInvalidScheduledAt)
				return
			}
			input.ScheduledAt = &t
		}
	}

	// Validate resulting state.
	validationErrs := input.Validate()
	if len(validationErrs) > 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"validation_errors": validationErrs,
		}) //nolint:errcheck
		return
	}

	// Auto-default rounds_total when switching to Mexicano.
	if input.GameMode == domain.ModeMexicano && input.RoundsTotal == nil {
		v := 7
		input.RoundsTotal = &v
	}

	if err := h.store.UpdateSessionConfig(id, input); err != nil {
		respondAPIError(w, ErrServerError)
		return
	}

	updated, err := h.store.GetSession(id)
	if err != nil {
		respondAPIError(w, ErrServerError)
		return
	}
	h.hub.Emit(id, events.Envelope{Type: events.EventSessionUpdated})
	respond(w, http.StatusOK, updated)
}

func (h *Handler) closeSession(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	sess, err := h.store.GetSession(id)
	if errors.Is(err, store.ErrNotFound) {
		respondAPIError(w, ErrSessionNotFound)
		return
	}
	if err != nil {
		respondAPIError(w, ErrServerError)
		return
	}
	if !isAdmin(extractAdminToken(r), sess.AdminToken) {
		respondAPIError(w, ErrAdminRequired)
		return
	}
	if sess.Status == domain.StatusDone {
		respondAPIError(w, ErrSessionAlreadyEnded)
		return
	}
	if err := h.store.CompleteSession(id, true); err != nil {
		respondAPIError(w, ErrServerError)
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
		respondAPIError(w, ErrSessionNotFound)
		return
	}
	if err != nil {
		respondAPIError(w, ErrServerError)
		return
	}
	if !isAdmin(extractAdminToken(r), sess.AdminToken) {
		respondAPIError(w, ErrAdminRequired)
		return
	}
	if err := h.store.DeleteSession(id); err != nil {
		respondAPIError(w, ErrServerError)
		return
	}
	h.hub.Emit(id, events.Envelope{Type: events.EventSessionUpdated})
	w.WriteHeader(http.StatusNoContent)
}
