package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/fabianthorsen/openpadel/internal/store"
)

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email       string `json:"email"`
		DisplayName string `json:"display_name"`
		Password    string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Email == "" || body.DisplayName == "" || body.Password == "" {
		respondError(w, http.StatusBadRequest, "email, display_name and password are required")
		return
	}
	if len(body.Password) < 8 {
		respondError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}

	user, err := h.store.CreateUser(body.Email, body.DisplayName, body.Password)
	if errors.Is(err, store.ErrEmailTaken) {
		respondError(w, http.StatusConflict, "email already registered")
		return
	}
	if err != nil {
		slog.Error("register: CreateUser failed", "err", err)
		respondError(w, http.StatusInternalServerError, "could not create user")
		return
	}

	token, err := h.store.CreateAuthToken(user.ID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not create session")
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
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.store.AuthenticateUser(body.Email, body.Password)
	if errors.Is(err, store.ErrInvalidCredentials) {
		respondError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not authenticate")
		return
	}

	token, err := h.store.CreateAuthToken(user.ID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not create session")
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
		respondError(w, http.StatusUnauthorized, "not authenticated")
		return
	}
	respond(w, http.StatusOK, user)
}

func (h *Handler) updateProfile(w http.ResponseWriter, r *http.Request) {
	user := userFromContext(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "not authenticated")
		return
	}
	var body struct {
		DisplayName string `json:"display_name"`
		AvatarIcon  string `json:"avatar_icon"`
		AvatarColor string `json:"avatar_color"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.DisplayName == "" {
		respondError(w, http.StatusBadRequest, "display_name is required")
		return
	}
	updated, err := h.store.UpdateProfile(user.ID, body.DisplayName, body.AvatarIcon, body.AvatarColor)
	if err != nil {
		slog.Error("updateProfile failed", "err", err)
		respondError(w, http.StatusInternalServerError, "could not update profile")
		return
	}
	respond(w, http.StatusOK, updated)
}

func (h *Handler) profile(w http.ResponseWriter, r *http.Request) {
	user := userFromContext(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "not authenticated")
		return
	}
	stats, err := h.store.GetCareerStats(user.ID)
	if err != nil {
		slog.Error("profile: GetCareerStats failed", "err", err)
		respondError(w, http.StatusInternalServerError, "could not load stats")
		return
	}
	tennisStats, err := h.store.GetTennisCareerStats(user.ID)
	if err != nil {
		slog.Error("profile: GetTennisCareerStats failed", "err", err)
		respondError(w, http.StatusInternalServerError, "could not load stats")
		return
	}
	respond(w, http.StatusOK, map[string]any{
		"user":         user,
		"stats":        stats,
		"tennis_stats": tennisStats,
	})
}

func (h *Handler) history(w http.ResponseWriter, r *http.Request) {
	user := userFromContext(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "not authenticated")
		return
	}
	entries, err := h.store.GetTournamentHistory(user.ID)
	if err != nil {
		slog.Error("history: GetTournamentHistory failed", "err", err)
		respondError(w, http.StatusInternalServerError, "could not load history")
		return
	}
	upcoming, err := h.store.GetUpcomingTournaments(user.ID)
	if err != nil {
		slog.Error("history: GetUpcomingTournaments failed", "err", err)
		respondError(w, http.StatusInternalServerError, "could not load upcoming")
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
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Token == "" || len(body.Password) < 8 {
		respondError(w, http.StatusBadRequest, "token and password (min 8 chars) are required")
		return
	}

	if err := h.store.RedeemPasswordResetToken(body.Token, body.Password); err != nil {
		if errors.Is(err, store.ErrInvalidOrExpiredToken) {
			respondError(w, http.StatusBadRequest, "invalid or expired reset link")
			return
		}
		slog.Error("resetPassword: RedeemPasswordResetToken failed", "err", err)
		respondError(w, http.StatusInternalServerError, "could not reset password")
		return
	}
	respond(w, http.StatusOK, map[string]any{})
}

func (h *Handler) deleteAccount(w http.ResponseWriter, r *http.Request) {
	user := userFromContext(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "not authenticated")
		return
	}
	if err := h.store.DeleteUser(user.ID); err != nil {
		slog.Error("deleteAccount: DeleteUser failed", "err", err)
		respondError(w, http.StatusInternalServerError, "could not delete account")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

