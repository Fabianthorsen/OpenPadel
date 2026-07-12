package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/fabianthorsen/openpadel/internal/domain"
	"github.com/fabianthorsen/openpadel/internal/store"
)

func (h *Handler) createClub(w http.ResponseWriter, r *http.Request) {
	user := userFromContext(r)
	if user == nil {
		respondAPIError(w, ErrNotAuthenticated)
		return
	}

	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		AvatarIcon  string `json:"avatar_icon"`
		AvatarColor string `json:"avatar_color"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondAPIError(w, ErrInvalidRequestBody)
		return
	}

	if body.Name == "" {
		respondAPIError(w, ErrInvalidRequestBody)
		return
	}

	club, err := h.store.CreateClub(body.Name, body.Description, body.AvatarIcon, body.AvatarColor, user.ID)
	if err != nil {
		slog.Error("createClub", "err", err)
		respondAPIError(w, ErrServerError)
		return
	}

	respond(w, http.StatusCreated, club)
}

func (h *Handler) getMyClubs(w http.ResponseWriter, r *http.Request) {
	user := userFromContext(r)
	if user == nil {
		respondAPIError(w, ErrNotAuthenticated)
		return
	}

	clubs, err := h.store.GetUserClubs(user.ID)
	if err != nil {
		slog.Error("getMyClubs", "err", err)
		respondAPIError(w, ErrServerError)
		return
	}

	type ClubListItem struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		AvatarIcon  string `json:"avatar_icon"`
		AvatarColor string `json:"avatar_color"`
		MyRole      string `json:"my_role"`
		RosterCount int    `json:"roster_count"`
	}

	var result []ClubListItem
	for _, club := range clubs {
		member, err := h.store.GetClubMember(club.ID, user.ID)
		if err != nil {
			continue
		}

		count, err := h.store.GetClubMemberCount(club.ID)
		if err != nil {
			count = 0
		}

		result = append(result, ClubListItem{
			ID:          club.ID,
			Name:        club.Name,
			AvatarIcon:  club.AvatarIcon,
			AvatarColor: club.AvatarColor,
			MyRole:      member.Role,
			RosterCount: count,
		})
	}

	respond(w, http.StatusOK, result)
}

func (h *Handler) getClub(w http.ResponseWriter, r *http.Request) {
	clubID := chi.URLParam(r, "id")
	user := userFromContext(r)

	club, err := h.store.GetClub(clubID)
	if err != nil {
		if err == store.ErrNotFound {
			respondAPIError(w, ErrSessionNotFound)
		} else {
			slog.Error("getClub", "err", err)
			respondAPIError(w, ErrServerError)
		}
		return
	}

	var myRole string
	var isAdmin bool
	if user != nil {
		member, err := h.store.GetClubMember(clubID, user.ID)
		if err == nil {
			myRole = member.Role
			isAdmin = member.Role == "admin"
		}
	} else {
		respondAPIError(w, ErrNotAuthenticated)
		return
	}

	members, err := h.store.GetClubMembers(clubID)
	if err != nil {
		slog.Error("getClub: GetClubMembers", "err", err)
		respondAPIError(w, ErrServerError)
		return
	}

	detail := domain.ClubDetail{
		Club:        *club,
		Members:     members,
		IsAdmin:     isAdmin,
		MyRole:      myRole,
		RosterCount: len(members),
	}

	respond(w, http.StatusOK, detail)
}

func (h *Handler) updateClub(w http.ResponseWriter, r *http.Request) {
	clubID := chi.URLParam(r, "id")
	user := userFromContext(r)
	if user == nil {
		respondAPIError(w, ErrNotAuthenticated)
		return
	}

	isAdmin, err := h.isClubAdmin(clubID, user.ID)
	if err != nil || !isAdmin {
		respondAPIError(w, ErrAdminRequired)
		return
	}

	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		AvatarIcon  string `json:"avatar_icon"`
		AvatarColor string `json:"avatar_color"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondAPIError(w, ErrInvalidRequestBody)
		return
	}

	if body.Name == "" {
		respondAPIError(w, ErrInvalidRequestBody)
		return
	}

	err = h.store.UpdateClub(clubID, body.Name, body.Description, body.AvatarIcon, body.AvatarColor)
	if err != nil {
		slog.Error("updateClub", "err", err)
		respondAPIError(w, ErrServerError)
		return
	}

	club, err := h.store.GetClub(clubID)
	if err != nil {
		slog.Error("updateClub: GetClub", "err", err)
		respondAPIError(w, ErrServerError)
		return
	}

	respond(w, http.StatusOK, club)
}

func (h *Handler) deleteClub(w http.ResponseWriter, r *http.Request) {
	clubID := chi.URLParam(r, "id")
	user := userFromContext(r)
	if user == nil {
		respondAPIError(w, ErrNotAuthenticated)
		return
	}

	isAdmin, err := h.isClubAdmin(clubID, user.ID)
	if err != nil || !isAdmin {
		respondAPIError(w, ErrAdminRequired)
		return
	}

	err = h.store.DeleteClub(clubID)
	if err != nil {
		slog.Error("deleteClub", "err", err)
		respondAPIError(w, ErrServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) isClubAdmin(clubID, userID string) (bool, error) {
	member, err := h.store.GetClubMember(clubID, userID)
	if err != nil {
		return false, err
	}
	return member.Role == "admin", nil
}
