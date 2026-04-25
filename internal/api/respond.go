package api

import (
	"encoding/json"
	"net/http"
)

func respond(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}


type APIError struct {
	Code   string
	Status int
}

func respondAPIError(w http.ResponseWriter, err APIError) {
	respond(w, err.Status, map[string]string{"error": err.Code})
}

// APIError sentinels
var (
	ErrInvalidRequestBody = APIError{"invalid_request_body", http.StatusBadRequest}
	ErrInvalidCourts = APIError{"invalid_courts", http.StatusBadRequest}
	ErrInvalidPoints = APIError{"invalid_points", http.StatusBadRequest}
	ErrInvalidScheduledAt = APIError{"invalid_scheduled_at", http.StatusBadRequest}
	ErrInvalidGameMode = APIError{"invalid_game_mode", http.StatusBadRequest}
	ErrInvalidRoundsTotal = APIError{"invalid_rounds_total", http.StatusBadRequest}
	ErrInvalidCourtDuration = APIError{"invalid_court_duration", http.StatusBadRequest}
	ErrDisplayNameRequired = APIError{"display_name_required", http.StatusBadRequest}
	ErrFieldsRequired = APIError{"fields_required", http.StatusBadRequest}
	ErrNameRequired = APIError{"name_required", http.StatusBadRequest}
	ErrPasswordTooShort = APIError{"password_too_short", http.StatusBadRequest}
	ErrScoresNegative = APIError{"scores_negative", http.StatusBadRequest}
	ErrScoresInvalidSum = APIError{"scores_invalid_sum", http.StatusBadRequest}
	ErrInvalidResetLink = APIError{"invalid_reset_link", http.StatusBadRequest}

	ErrNotAuthenticated = APIError{"not_authenticated", http.StatusUnauthorized}
	ErrInvalidToken = APIError{"invalid_token", http.StatusUnauthorized}
	ErrInvalidEmailOrPassword = APIError{"invalid_email_or_password", http.StatusUnauthorized}

	ErrAdminRequired = APIError{"admin_required", http.StatusForbidden}

	ErrAlreadyInSession = APIError{"already_in_session", http.StatusConflict}
	ErrAlreadyContact = APIError{"already_contact", http.StatusConflict}
	ErrAlreadyInvited = APIError{"already_invited", http.StatusConflict}
	ErrEmailAlreadyRegistered = APIError{"email_already_registered", http.StatusConflict}
	ErrNameTaken = APIError{"name_taken", http.StatusConflict}
	ErrSessionAlreadyStarted = APIError{"session_already_started", http.StatusConflict}
	ErrSessionAlreadyEnded = APIError{"session_already_ended", http.StatusConflict}
	ErrSessionNotActive = APIError{"session_not_active", http.StatusConflict}
	ErrSessionNotStarted = APIError{"session_not_started", http.StatusConflict}
	ErrRoundNotComplete = APIError{"round_not_complete", http.StatusConflict}
	ErrRoundLimitReached = APIError{"round_limit_reached", http.StatusConflict}
	ErrTournamentExpired = APIError{"tournament_expired", http.StatusConflict}

	ErrContactNotFound = APIError{"contact_not_found", http.StatusNotFound}
	ErrInviteNotFound = APIError{"invite_not_found", http.StatusNotFound}
	ErrMatchNotFound = APIError{"match_not_found", http.StatusNotFound}
	ErrNoActiveRound = APIError{"no_active_round", http.StatusNotFound}
	ErrPlayerNotFound = APIError{"player_not_found", http.StatusNotFound}
	ErrSessionNotFound = APIError{"session_not_found", http.StatusNotFound}
	ErrUserNotFound = APIError{"user_not_found", http.StatusNotFound}

	ErrServerError = APIError{"server_error", http.StatusInternalServerError}
	ErrCouldNotCreateSession = APIError{"could_not_create_session", http.StatusInternalServerError}
)
