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
	if user == nil {
		respondAPIError(w, ErrNotAuthenticated)
		return
	}

	club, err := h.store.GetClub(clubID)
	if err != nil {
		if err == store.ErrNotFound {
			respondAPIError(w, ErrClubNotFound)
		} else {
			slog.Error("getClub", "err", err)
			respondAPIError(w, ErrServerError)
		}
		return
	}

	member, err := h.store.GetClubMember(clubID, user.ID)
	if err != nil {
		if err == store.ErrNotFound {
			respondAPIError(w, ErrNotClubMember)
			return
		}
		slog.Error("getClub: GetClubMember", "err", err)
		respondAPIError(w, ErrServerError)
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
		IsAdmin:     member.Role == "admin",
		MyRole:      member.Role,
		RosterCount: len(members),
	}

	respond(w, http.StatusOK, detail)
}

// previewClubJoin returns a no-auth preview of a Club behind a join link so a
// visitor can see what they're joining before logging in. An unknown code 404s.
// The join_code is deliberately not echoed back — the caller already has it.
func (h *Handler) previewClubJoin(w http.ResponseWriter, r *http.Request) {
	joinCode := chi.URLParam(r, "join_code")

	club, err := h.store.GetClubByJoinCode(joinCode)
	if err != nil {
		if err == store.ErrNotFound {
			respondAPIError(w, ErrClubNotFound)
			return
		}
		slog.Error("previewClubJoin", "err", err)
		respondAPIError(w, ErrServerError)
		return
	}

	count, err := h.store.GetClubMemberCount(club.ID)
	if err != nil {
		slog.Error("previewClubJoin: GetClubMemberCount", "err", err)
		respondAPIError(w, ErrServerError)
		return
	}

	respond(w, http.StatusOK, struct {
		Name        string `json:"name"`
		AvatarIcon  string `json:"avatar_icon"`
		AvatarColor string `json:"avatar_color"`
		MemberCount int    `json:"member_count"`
	}{
		Name:        club.Name,
		AvatarIcon:  club.AvatarIcon,
		AvatarColor: club.AvatarColor,
		MemberCount: count,
	})
}

// joinClub adds the authenticated caller to the Club behind the join code. It is
// idempotent: an already-a-member caller succeeds without creating a duplicate
// row. Returns the Club id so the client can navigate to the Club page.
func (h *Handler) joinClub(w http.ResponseWriter, r *http.Request) {
	user := userFromContext(r)
	if user == nil {
		respondAPIError(w, ErrNotAuthenticated)
		return
	}

	var body struct {
		JoinCode string `json:"join_code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondAPIError(w, ErrInvalidRequestBody)
		return
	}
	if body.JoinCode == "" {
		respondAPIError(w, ErrInvalidRequestBody)
		return
	}

	club, err := h.store.GetClubByJoinCode(body.JoinCode)
	if err != nil {
		if err == store.ErrNotFound {
			respondAPIError(w, ErrClubNotFound)
			return
		}
		slog.Error("joinClub: GetClubByJoinCode", "err", err)
		respondAPIError(w, ErrServerError)
		return
	}

	if err := h.store.JoinClub(club.ID, user.ID); err != nil && err != store.ErrAlreadyMember {
		slog.Error("joinClub", "err", err)
		respondAPIError(w, ErrServerError)
		return
	}

	respond(w, http.StatusOK, map[string]string{"id": club.ID})
}

// rotateClubJoinCode revokes the Club's current join code and issues a new one,
// invalidating any previously shared link. Club Admins only.
func (h *Handler) rotateClubJoinCode(w http.ResponseWriter, r *http.Request) {
	clubID := chi.URLParam(r, "id")
	user := userFromContext(r)
	if user == nil {
		respondAPIError(w, ErrNotAuthenticated)
		return
	}

	member, err := h.store.GetClubMember(clubID, user.ID)
	if err != nil {
		if err == store.ErrNotFound {
			respondAPIError(w, ErrNotClubMember)
			return
		}
		slog.Error("rotateClubJoinCode: GetClubMember", "err", err)
		respondAPIError(w, ErrServerError)
		return
	}
	if member.Role != "admin" {
		respondAPIError(w, ErrAdminRequired)
		return
	}

	newCode, err := h.store.UpdateClubJoinCode(clubID)
	if err != nil {
		slog.Error("rotateClubJoinCode", "err", err)
		respondAPIError(w, ErrServerError)
		return
	}

	respond(w, http.StatusOK, map[string]string{"join_code": newCode})
}
