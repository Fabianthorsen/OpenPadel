package api

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
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
		ClubID               string  `json:"club_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondAPIError(w, ErrInvalidRequestBody)
		return
	}

	// A club event may only be created by a Member of that Club. Owning a Club
	// role grants no authority over the resulting Session — this is just the gate
	// on who may attach a Session to a Club.
	var clubID *string
	if body.ClubID != "" {
		if !h.requireClubMember(w, r, body.ClubID) {
			return
		}
		clubID = &body.ClubID
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
		ClubID:               clubID,
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

	// A club event tells the whole Club automatically — no personal invite needed.
	// Fan out over the Club roster via user-SSE + web-push. The global home feed
	// (GetUpcomingTournaments) is deliberately left untouched.
	if clubID != nil {
		go h.notifyClubEventCreated(*clubID, sess.ID, creatorUserID)
	}

	respond(w, http.StatusCreated, sess)
}

// notifyClubEventCreated fans a new club event out to the Club's roster: a live
// user-SSE nudge so open clients refresh, plus a web-push so members hear about
// it when the app is closed. The creator (creatorUserID) is skipped — they just
// made it. Runs in its own goroutine; failures are logged, never surfaced.
func (h *Handler) notifyClubEventCreated(clubID, sessionID, creatorUserID string) {
	club, err := h.store.GetClub(clubID)
	if err != nil {
		slog.Error("notifyClubEventCreated: GetClub", "err", err)
		return
	}
	members, err := h.store.GetClubMembers(clubID)
	if err != nil {
		slog.Error("notifyClubEventCreated: GetClubMembers", "err", err)
		return
	}
	title := "New " + club.Name + " game"
	notifBody := "A new game was scheduled in " + club.Name + " — tap to join."
	for _, m := range members {
		if m.UserID == creatorUserID {
			continue
		}
		h.hub.EmitToUser(m.UserID, events.Envelope{Type: events.EventClubEventCreated})
		h.sendPushToUser(m.UserID, title, notifBody, "/s/"+sessionID)
	}
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

	// A club event carries the Club's name so the join screen can frame it as a
	// club game rather than a personal invite. This is public identity only (it's
	// on the shareable join screen a Guest sees) — never an authorization signal.
	if sess.ClubID != "" {
		if club, err := h.store.GetClub(sess.ClubID); err == nil {
			sess.ClubName = club.Name
		}
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

	var rawBody map[string]interface{}
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		respondAPIError(w, ErrInvalidRequestBody)
		return
	}
	if err := json.Unmarshal(bodyBytes, &rawBody); err != nil {
		respondAPIError(w, ErrInvalidRequestBody)
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
	if err := json.Unmarshal(bodyBytes, &body); err != nil {
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
	// Handle RoundsTotal specially: check if it was in the JSON (even if null)
	_, hasExplicitRoundsTotal := rawBody["rounds_total"]
	if hasExplicitRoundsTotal {
		input.RoundsTotal = body.RoundsTotal // This can be nil (unlimited) or a value (fixed)
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

	// Auto-default rounds_total when switching to Mexicano or Americano (only if not explicitly set in request).
	if input.RoundsTotal == nil && !hasExplicitRoundsTotal {
		if input.GameMode == domain.ModeMexicano {
			v := 7
			input.RoundsTotal = &v
		}
		// For Americano, rounds_total will be calculated at session start time based on player count
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
	// Guard: unlimited sessions (rounds_total = null) require at least 1 round with at least 1 score submitted
	if sess.RoundsTotal == nil && sess.Status == domain.StatusPlaying {
		// Check if any matches have scores (at least one round has been scored)
		hasScores, err := h.store.HasAnyScores(id)
		if err != nil || !hasScores {
			respondAPIError(w, ErrSessionTooEarlyToClose)
			return
		}
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

	// Grab pending invitees before the delete: the FK cascade removes the invite
	// rows, so we must capture who to notify while they still exist.
	pending, _ := h.store.GetSessionInvites(id)

	if err := h.store.DeleteSession(id); err != nil {
		respondAPIError(w, ErrServerError)
		return
	}
	h.hub.Emit(id, events.Envelope{Type: events.EventSessionUpdated})

	// Tell invited users their pending invite is gone so their client drops it
	// immediately instead of waiting for the next app reload.
	for _, inv := range pending {
		h.hub.EmitToUser(inv.ToUserID, events.Envelope{Type: events.EventInviteRevoked})
	}

	w.WriteHeader(http.StatusNoContent)
}
