package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/fabianthorsen/openpadel/internal/domain"
	"github.com/fabianthorsen/openpadel/internal/store"
)

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email       string `json:"email"`
		DisplayName string `json:"display_name"`
		Password    string `json:"password"`
		SelfRating  int    `json:"self_rating"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondAPIError(w, ErrInvalidRequestBody)
		return
	}
	if body.Email == "" || body.DisplayName == "" || body.Password == "" {
		respondAPIError(w, ErrFieldsRequired)
		return
	}
	if len(body.Password) < 8 {
		respondAPIError(w, ErrPasswordTooShort)
		return
	}
	// self_rating is required at registration so new accounts are never unrated
	// (ADR 0006). Missing decodes to the zero value 0, which fails the range check.
	if !domain.IsValidRating(body.SelfRating) {
		respondAPIError(w, ErrInvalidRating)
		return
	}

	user, err := h.store.CreateUser(body.Email, body.DisplayName, body.Password, body.SelfRating)
	if errors.Is(err, store.ErrEmailTaken) {
		respondAPIError(w, ErrEmailAlreadyRegistered)
		return
	}
	if err != nil {
		slog.Error("register: CreateUser failed", "err", err)
		respondAPIError(w, ErrServerError)
		return
	}

	token, err := h.store.CreateAuthToken(user.ID)
	if err != nil {
		respondAPIError(w, ErrServerError)
		return
	}

	respond(w, http.StatusCreated, map[string]any{
		"token": token,
		"user":  user,
	})
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondAPIError(w, ErrInvalidRequestBody)
		return
	}

	user, err := h.store.AuthenticateUser(body.Email, body.Password)
	if errors.Is(err, store.ErrInvalidCredentials) {
		respondAPIError(w, ErrInvalidEmailOrPassword)
		return
	}
	if err != nil {
		respondAPIError(w, ErrServerError)
		return
	}

	token, err := h.store.CreateAuthToken(user.ID)
	if err != nil {
		respondAPIError(w, ErrServerError)
		return
	}

	respond(w, http.StatusOK, map[string]any{
		"token": token,
		"user":  user,
	})
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	token := extractAdminToken(r)
	if token != "" {
		h.store.DeleteAuthToken(token)
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	user := userFromContext(r)
	if user == nil {
		respondAPIError(w, ErrNotAuthenticated)
		return
	}
	respond(w, http.StatusOK, user)
}

// updateSelfRating sets the authenticated User's default self_rating. Per ADR
// 0006 this only seeds future joins — a Player.rating already snapshotted into a
// joined Session is untouched. It is the shared save path behind both the
// settings editor (#212) and the home backfill gate (#213).
func (h *Handler) updateSelfRating(w http.ResponseWriter, r *http.Request) {
	user := userFromContext(r)
	if user == nil {
		respondAPIError(w, ErrNotAuthenticated)
		return
	}
	var body struct {
		SelfRating int `json:"self_rating"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondAPIError(w, ErrInvalidRequestBody)
		return
	}
	if !domain.IsValidRating(body.SelfRating) {
		respondAPIError(w, ErrInvalidRating)
		return
	}
	if err := h.store.UpdateSelfRating(user.ID, body.SelfRating); err != nil {
		slog.Error("updateSelfRating failed", "err", err)
		respondAPIError(w, ErrServerError)
		return
	}
	updated, err := h.store.GetUserByID(user.ID)
	if err != nil {
		slog.Error("updateSelfRating: GetUserByID failed", "err", err)
		respondAPIError(w, ErrServerError)
		return
	}
	respond(w, http.StatusOK, updated)
}

func (h *Handler) updateProfile(w http.ResponseWriter, r *http.Request) {
	user := userFromContext(r)
	if user == nil {
		respondAPIError(w, ErrNotAuthenticated)
		return
	}
	var body struct {
		DisplayName string `json:"display_name"`
		AvatarIcon  string `json:"avatar_icon"`
		AvatarColor string `json:"avatar_color"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondAPIError(w, ErrInvalidRequestBody)
		return
	}
	if body.DisplayName == "" {
		respondAPIError(w, ErrDisplayNameRequired)
		return
	}
	updated, err := h.store.UpdateProfile(user.ID, body.DisplayName, body.AvatarIcon, body.AvatarColor)
	if err != nil {
		slog.Error("updateProfile failed", "err", err)
		respondAPIError(w, ErrServerError)
		return
	}
	respond(w, http.StatusOK, updated)
}

func (h *Handler) profile(w http.ResponseWriter, r *http.Request) {
	user := userFromContext(r)
	if user == nil {
		respondAPIError(w, ErrNotAuthenticated)
		return
	}
	stats, err := h.store.GetCareerSummary(user.ID)
	if err != nil {
		slog.Error("profile: GetCareerSummary failed", "err", err)
		respondAPIError(w, ErrServerError)
		return
	}
	respond(w, http.StatusOK, map[string]any{
		"user":  user,
		"stats": stats,
	})
}

// stats returns the per-Game-Mode career aggregates behind the Career Stats page
// (ADR 0007). Auth-gated; both modes are always present (zero-valued when the
// user has no games in one). See domain.ModeStats.
func (h *Handler) stats(w http.ResponseWriter, r *http.Request) {
	user := userFromContext(r)
	if user == nil {
		respondAPIError(w, ErrNotAuthenticated)
		return
	}
	modes, err := h.store.GetModeStats(user.ID)
	if err != nil {
		slog.Error("stats: GetModeStats failed", "err", err)
		respondAPIError(w, ErrServerError)
		return
	}
	// The per-Match results series backs the cross-mode recent-form curve derived
	// client-side, so no new endpoint is needed per stat (ADR 0007).
	series, err := h.store.GetMatchResultsSeries(user.ID)
	if err != nil {
		slog.Error("stats: GetMatchResultsSeries failed", "err", err)
		respondAPIError(w, ErrServerError)
		return
	}
	respond(w, http.StatusOK, map[string]any{"modes": modes, "series": series})
}

func (h *Handler) history(w http.ResponseWriter, r *http.Request) {
	user := userFromContext(r)
	if user == nil {
		respondAPIError(w, ErrNotAuthenticated)
		return
	}
	entries, err := h.store.GetTournamentHistory(user.ID)
	if err != nil {
		slog.Error("history: GetTournamentHistory failed", "err", err)
		respondAPIError(w, ErrServerError)
		return
	}
	upcoming, err := h.store.GetUpcomingTournaments(user.ID)
	if err != nil {
		slog.Error("history: GetUpcomingTournaments failed", "err", err)
		respondAPIError(w, ErrServerError)
		return
	}
	respond(w, http.StatusOK, map[string]any{"tournaments": entries, "upcoming": upcoming})
}

func (h *Handler) forgotPassword(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Email == "" {
		respond(w, http.StatusOK, map[string]any{})
		return
	}

	rawToken, err := h.store.CreatePasswordResetToken(body.Email)
	if err != nil {
		// Swallow ErrNotFound silently — don't reveal whether email exists
		if !errors.Is(err, store.ErrNotFound) {
			slog.Error("forgotPassword: CreatePasswordResetToken failed", "err", err)
		}
		respond(w, http.StatusOK, map[string]any{})
		return
	}

	resetURL := h.appURL + "/reset?token=" + rawToken
	user, _ := h.store.GetUserByEmail(body.Email)
	if err := h.email.SendPasswordReset(body.Email, user.DisplayName, resetURL); err != nil {
		slog.Error("forgotPassword: SendPasswordReset failed", "err", err)
	}
	respond(w, http.StatusOK, map[string]any{})
}

func (h *Handler) resetPassword(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondAPIError(w, ErrInvalidRequestBody)
		return
	}
	if body.Token == "" || len(body.Password) < 8 {
		respondAPIError(w, ErrFieldsRequired)
		return
	}

	if err := h.store.RedeemPasswordResetToken(body.Token, body.Password); err != nil {
		if errors.Is(err, store.ErrInvalidOrExpiredToken) {
			respondAPIError(w, ErrInvalidResetLink)
			return
		}
		slog.Error("resetPassword: RedeemPasswordResetToken failed", "err", err)
		respondAPIError(w, ErrServerError)
		return
	}
	respond(w, http.StatusOK, map[string]any{})
}

func (h *Handler) deleteAccount(w http.ResponseWriter, r *http.Request) {
	user := userFromContext(r)
	if user == nil {
		respondAPIError(w, ErrNotAuthenticated)
		return
	}
	if err := h.store.DeleteUser(user.ID); err != nil {
		slog.Error("deleteAccount: DeleteUser failed", "err", err)
		respondAPIError(w, ErrServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) getUserSessions(w http.ResponseWriter, r *http.Request) {
	user := userFromContext(r)
	if user == nil {
		respondAPIError(w, ErrNotAuthenticated)
		return
	}
	sessions, err := h.store.GetUserSessions(user.ID)
	if err != nil {
		slog.Error("getUserSessions: GetUserSessions failed", "err", err)
		respondAPIError(w, ErrServerError)
		return
	}
	respond(w, http.StatusOK, map[string]any{
		"sessions": sessions,
	})
}
