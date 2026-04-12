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
		respondError(w, http.StatusInternalServerError, "server_error")
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
		respondError(w, http.StatusBadRequest, "invalid_request_body")
		return
	}

	err := h.store.AddContact(user.ID, body.ContactUserID)
	if errors.Is(err, store.ErrNotFound) {
		respondError(w, http.StatusNotFound, "user_not_found")
		return
	}
	if errors.Is(err, store.ErrAlreadyContact) {
		respondError(w, http.StatusConflict, "already_contact")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error")
		return
	}
	respond(w, http.StatusCreated, map[string]string{"status": "ok"})
}

func (h *Handler) removeContact(w http.ResponseWriter, r *http.Request) {
	user := userFromContext(r)
	contactUserID := chi.URLParam(r, "contactID")

	err := h.store.RemoveContact(user.ID, contactUserID)
	if errors.Is(err, store.ErrNotFound) {
		respondError(w, http.StatusNotFound, "contact_not_found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error")
		return
	}
	respond(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) searchUsers(w http.ResponseWriter, r *http.Request) {
	user := userFromContext(r)
	q := r.URL.Query().Get("q")
	if len(q) < 2 {
		respondError(w, http.StatusBadRequest, "invalid_request_body")
		return
	}

	results, err := h.store.SearchUsers(user.ID, q)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error")
		return
	}
	respond(w, http.StatusOK, results)
}
