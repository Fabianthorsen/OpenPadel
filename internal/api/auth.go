package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/fabianthorsen/nottennis/internal/store"
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
		log.Printf("register: CreateUser failed: %v", err)
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

func (h *Handler) profile(w http.ResponseWriter, r *http.Request) {
	user := userFromContext(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "not authenticated")
		return
	}
	stats, err := h.store.GetCareerStats(user.ID)
	if err != nil {
		log.Printf("profile: GetCareerStats failed: %v", err)
		respondError(w, http.StatusInternalServerError, "could not load stats")
		return
	}
	respond(w, http.StatusOK, map[string]any{
		"user":  user,
		"stats": stats,
	})
}

