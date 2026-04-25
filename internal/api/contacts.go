package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/fabianthorsen/openpadel/internal/store"
)

func (h *Handler) getContacts(w http.ResponseWriter, r *http.Request) {
	user := userFromContext(r)
	contacts, err := h.store.GetContacts(user.ID)
	if err != nil {
		respondAPIError(w, ErrServerError)
		return
	}
	respond(w, http.StatusOK, contacts)
}

func (h *Handler) addContact(w http.ResponseWriter, r *http.Request) {
	user := userFromContext(r)
	var body struct {
		ContactUserID string `json:"contact_user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ContactUserID == "" {
		respondAPIError(w, ErrInvalidRequestBody)
		return
	}

	err := h.store.AddContact(user.ID, body.ContactUserID)
	if errors.Is(err, store.ErrNotFound) {
		respondAPIError(w, ErrUserNotFound)
		return
	}
	if errors.Is(err, store.ErrAlreadyContact) {
		respondAPIError(w, ErrAlreadyContact)
		return
	}
	if err != nil {
		respondAPIError(w, ErrServerError)
		return
	}
	respond(w, http.StatusCreated, map[string]string{"status": "ok"})
}

func (h *Handler) removeContact(w http.ResponseWriter, r *http.Request) {
	user := userFromContext(r)
	contactUserID := chi.URLParam(r, "contactID")

	err := h.store.RemoveContact(user.ID, contactUserID)
	if errors.Is(err, store.ErrNotFound) {
		respondAPIError(w, ErrContactNotFound)
		return
	}
	if err != nil {
		respondAPIError(w, ErrServerError)
		return
	}
	respond(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) searchUsers(w http.ResponseWriter, r *http.Request) {
	user := userFromContext(r)
	q := r.URL.Query().Get("q")
	if len(q) < 2 {
		respondAPIError(w, ErrInvalidRequestBody)
		return
	}

	results, err := h.store.SearchUsers(user.ID, q)
	if err != nil {
		respondAPIError(w, ErrServerError)
		return
	}
	respond(w, http.StatusOK, results)
}
