package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/fabianthorsen/openpadel/internal/domain"
	"github.com/fabianthorsen/openpadel/internal/store"
)

// requireClubAdmin resolves the caller's Club membership and confirms they hold
// the admin role. On failure it writes the appropriate error (401 when not
// authenticated, 403 not_club_member when not a member, 403 admin_required
// otherwise) and returns ok=false. This gate is account-scoped Club authz and is
// deliberately separate from Session AdminToken authz — the two never bridge.
func (h *Handler) requireClubAdmin(w http.ResponseWriter, r *http.Request, clubID string) bool {
	user := userFromContext(r)
	if user == nil {
		respondAPIError(w, ErrNotAuthenticated)
		return false
	}

	member, err := h.store.GetClubMember(clubID, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			respondAPIError(w, ErrNotClubMember)
			return false
		}
		slog.Error("requireClubAdmin: GetClubMember", "err", err)
		respondAPIError(w, ErrServerError)
		return false
	}
	if member.Role != "admin" {
		respondAPIError(w, ErrAdminRequired)
		return false
	}
	return true
}

// requireClubMember resolves the caller's Club membership (any role). On failure
// it writes the appropriate error (401 when not authenticated, 403
// not_club_member otherwise) and returns ok=false. Like requireClubAdmin this is
// account-scoped Club authz, never a Session AdminToken check.
func (h *Handler) requireClubMember(w http.ResponseWriter, r *http.Request, clubID string) bool {
	user := userFromContext(r)
	if user == nil {
		respondAPIError(w, ErrNotAuthenticated)
		return false
	}
	if _, err := h.store.GetClubMember(clubID, user.ID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			respondAPIError(w, ErrNotClubMember)
			return false
		}
		slog.Error("requireClubMember: GetClubMember", "err", err)
		respondAPIError(w, ErrServerError)
		return false
	}
	return true
}

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

// getClubEvents returns the Club's upcoming events (lobby|playing), ordered so
// the soonest/most-active is first. Members only — the events feed is part of the
// Club home, which is member-gated. Guests reach an individual event via its
// public Session join link, not through this list.
func (h *Handler) getClubEvents(w http.ResponseWriter, r *http.Request) {
	clubID := chi.URLParam(r, "id")
	if !h.requireClubMember(w, r, clubID) {
		return
	}

	events, err := h.store.GetClubEvents(clubID)
	if err != nil {
		slog.Error("getClubEvents", "err", err)
		respondAPIError(w, ErrServerError)
		return
	}
	respond(w, http.StatusOK, events)
}

// getClubLeaderboard returns the Club's current-form leaderboard — the ranked
// board plus the provisional "not yet ranked" list. Members only, like the rest
// of the Club home. Recomputed on every read (no materialized state), so a
// newly scored Match shows up on the next call.
func (h *Handler) getClubLeaderboard(w http.ResponseWriter, r *http.Request) {
	clubID := chi.URLParam(r, "id")
	if !h.requireClubMember(w, r, clubID) {
		return
	}

	board, err := h.store.GetClubLeaderboard(clubID)
	if err != nil {
		slog.Error("getClubLeaderboard", "err", err)
		respondAPIError(w, ErrServerError)
		return
	}
	respond(w, http.StatusOK, board)
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
	if !h.requireClubAdmin(w, r, clubID) {
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

// updateClub edits the Club's name/description/avatar. Club Admins only.
func (h *Handler) updateClub(w http.ResponseWriter, r *http.Request) {
	clubID := chi.URLParam(r, "id")
	if !h.requireClubAdmin(w, r, clubID) {
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

	if err := h.store.UpdateClub(clubID, body.Name, body.Description, body.AvatarIcon, body.AvatarColor); err != nil {
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

// deleteClub hard-deletes the Club (Club Admins only). club_members and
// club_invites cascade; sessions.club_id is unset via ON DELETE SET NULL so past
// games survive as ordinary Sessions.
func (h *Handler) deleteClub(w http.ResponseWriter, r *http.Request) {
	clubID := chi.URLParam(r, "id")
	if !h.requireClubAdmin(w, r, clubID) {
		return
	}

	if err := h.store.DeleteClub(clubID); err != nil {
		slog.Error("deleteClub", "err", err)
		respondAPIError(w, ErrServerError)
		return
	}
	respond(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// removeClubMember lets a Member leave (self-removal) or a Club Admin remove
// another member. The sole Admin is blocked from leaving — a Club always keeps
// ≥1 Admin while it exists.
func (h *Handler) removeClubMember(w http.ResponseWriter, r *http.Request) {
	clubID := chi.URLParam(r, "id")
	targetID := chi.URLParam(r, "userID")
	user := userFromContext(r)
	if user == nil {
		respondAPIError(w, ErrNotAuthenticated)
		return
	}

	caller, err := h.store.GetClubMember(clubID, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			respondAPIError(w, ErrNotClubMember)
			return
		}
		slog.Error("removeClubMember: GetClubMember", "err", err)
		respondAPIError(w, ErrServerError)
		return
	}
	// A plain Member may remove only themselves; removing anyone else needs admin.
	if targetID != user.ID && caller.Role != "admin" {
		respondAPIError(w, ErrAdminRequired)
		return
	}

	err = h.store.RemoveClubMember(clubID, targetID)
	if errors.Is(err, store.ErrNotFound) {
		respondAPIError(w, ErrNotClubMember)
		return
	}
	if errors.Is(err, store.ErrLastAdmin) {
		respondAPIError(w, ErrLastAdmin)
		return
	}
	if err != nil {
		slog.Error("removeClubMember", "err", err)
		respondAPIError(w, ErrServerError)
		return
	}
	respond(w, http.StatusOK, map[string]string{"status": "removed"})
}

// updateClubMemberRole promotes or demotes a member (Club Admins only). Demoting
// the sole Admin is blocked so the Club always keeps ≥1 Admin.
func (h *Handler) updateClubMemberRole(w http.ResponseWriter, r *http.Request) {
	clubID := chi.URLParam(r, "id")
	targetID := chi.URLParam(r, "userID")
	if !h.requireClubAdmin(w, r, clubID) {
		return
	}

	var body struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondAPIError(w, ErrInvalidRequestBody)
		return
	}
	if body.Role != "admin" && body.Role != "member" {
		respondAPIError(w, ErrInvalidRequestBody)
		return
	}

	err := h.store.UpdateClubMemberRole(clubID, targetID, body.Role)
	if errors.Is(err, store.ErrNotFound) {
		respondAPIError(w, ErrNotClubMember)
		return
	}
	if errors.Is(err, store.ErrLastAdmin) {
		respondAPIError(w, ErrLastAdmin)
		return
	}
	if err != nil {
		slog.Error("updateClubMemberRole", "err", err)
		respondAPIError(w, ErrServerError)
		return
	}
	respond(w, http.StatusOK, map[string]string{"status": "updated"})
}
