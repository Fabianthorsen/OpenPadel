package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJoinSession(t *testing.T) {
	srv, _ := newAPITestServer(t)
	sessID, _ := mustCreateSession(t, srv, "")

	res := postReq(t, srv, "/api/sessions/"+sessID+"/players", map[string]any{
		"name":   "Alice",
		"rating": 3,
	}, "")
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", res.StatusCode)
	}
	var player struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Active bool   `json:"active"`
	}
	decodeBody(t, res, &player)
	if player.ID == "" {
		t.Error("expected non-empty player ID")
	}
	if player.Name != "Alice" {
		t.Errorf("expected name 'Alice', got %q", player.Name)
	}
	if !player.Active {
		t.Error("expected player to be active")
	}
}

func TestJoinSession_GuestWithRating(t *testing.T) {
	srv, _ := newAPITestServer(t)
	sessID, _ := mustCreateSession(t, srv, "")

	res := postReq(t, srv, "/api/sessions/"+sessID+"/players", map[string]any{
		"name":   "Alice",
		"rating": 5,
	}, "")
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", res.StatusCode)
	}
	var player struct {
		Rating int `json:"rating"`
	}
	decodeBody(t, res, &player)
	if player.Rating != 5 {
		t.Errorf("expected guest rating 5, got %d", player.Rating)
	}
}

func TestJoinSession_InvalidRating(t *testing.T) {
	srv, _ := newAPITestServer(t)
	sessID, _ := mustCreateSession(t, srv, "")

	res := postReq(t, srv, "/api/sessions/"+sessID+"/players", map[string]any{
		"name":   "Alice",
		"rating": 9,
	}, "")
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for out-of-range rating, got %d", res.StatusCode)
	}
	_ = res.Body.Close()
}

func TestJoinSession_GuestMissingRating(t *testing.T) {
	srv, _ := newAPITestServer(t)
	sessID, _ := mustCreateSession(t, srv, "")

	// An anonymous guest self-joining by link must pick a skill level (#210).
	res := postReq(t, srv, "/api/sessions/"+sessID+"/players", map[string]any{
		"name": "Alice",
	}, "")
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for guest join without rating, got %d", res.StatusCode)
	}
	var body struct {
		Error string `json:"error"`
	}
	decodeBody(t, res, &body)
	if body.Error != "rating_required" {
		t.Errorf("expected error 'rating_required', got %q", body.Error)
	}
}

func TestJoinSession_AdminGuestWithoutRating(t *testing.T) {
	srv, _ := newAPITestServer(t)
	sessID, adminToken := mustCreateSession(t, srv, "")

	// An admin adding a guest by name does not have to supply a rating — the
	// guest falls back to the median.
	res := postReq(t, srv, "/api/sessions/"+sessID+"/players", map[string]any{
		"name": "Alice",
	}, adminToken)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 for admin-added guest, got %d", res.StatusCode)
	}
	var player struct {
		Rating int `json:"rating"`
	}
	decodeBody(t, res, &player)
	if player.Rating != 3 {
		t.Errorf("expected median guest rating 3, got %d", player.Rating)
	}
}

func TestJoinSession_EmptyName(t *testing.T) {
	srv, _ := newAPITestServer(t)
	sessID, _ := mustCreateSession(t, srv, "")

	res := postReq(t, srv, "/api/sessions/"+sessID+"/players", map[string]any{
		"name": "",
	}, "")
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.StatusCode)
	}
	res.Body.Close()
}

func TestJoinSession_SessionNotFound(t *testing.T) {
	srv, _ := newAPITestServer(t)

	res := postReq(t, srv, "/api/sessions/XXXX/players", map[string]any{
		"name": "Alice",
	}, "")
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", res.StatusCode)
	}
	res.Body.Close()
}

func TestJoinSession_AlreadyStarted(t *testing.T) {
	srv, _ := newAPITestServer(t)
	sessID, adminToken, _ := setupStartedSession(t, srv)

	res := postReq(t, srv, "/api/sessions/"+sessID+"/players", map[string]any{
		"name": "Eve",
	}, adminToken)
	if res.StatusCode != http.StatusConflict {
		t.Fatalf("expected 409 for already-started session, got %d", res.StatusCode)
	}
	res.Body.Close()
}

func TestUpdatePlayerRating(t *testing.T) {
	srv, _ := newAPITestServer(t)
	sessID, adminToken := mustCreateSession(t, srv, "")
	playerID := mustJoinSession(t, srv, sessID, "Alice", adminToken)

	res := patchReq(t, srv, "/api/sessions/"+sessID+"/players/"+playerID+"/rating", map[string]any{
		"rating": 5,
	}, adminToken)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
	var player struct {
		Rating int `json:"rating"`
	}
	decodeBody(t, res, &player)
	if player.Rating != 5 {
		t.Errorf("expected rating 5 in response, got %d", player.Rating)
	}

	// The change persists and re-displays on the session.
	res = getReq(t, srv, "/api/sessions/"+sessID, adminToken)
	var sess struct {
		Players []struct {
			ID     string `json:"id"`
			Rating int    `json:"rating"`
		} `json:"players"`
	}
	decodeBody(t, res, &sess)
	var found bool
	for _, p := range sess.Players {
		if p.ID == playerID {
			found = true
			if p.Rating != 5 {
				t.Errorf("expected persisted rating 5, got %d", p.Rating)
			}
		}
	}
	if !found {
		t.Fatal("player not found in session after rating edit")
	}
}

func TestUpdatePlayerRating_AdminTokenViaHeader(t *testing.T) {
	srv, _ := newAPITestServer(t)
	sessID, adminToken := mustCreateSession(t, srv, "")
	playerID := mustJoinSession(t, srv, sessID, "Alice", adminToken)

	// The web client sends the admin token in the X-Admin-Token header (not the
	// Authorization bearer header). The endpoint must accept it, same as join.
	res := doRequest(t, srv, http.MethodPatch, "/api/sessions/"+sessID+"/players/"+playerID+"/rating", map[string]any{
		"rating": 5,
	}, "", map[string]string{"X-Admin-Token": adminToken})
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 with X-Admin-Token header, got %d", res.StatusCode)
	}
	_ = res.Body.Close()
}

func TestUpdatePlayerRating_RequiresAdmin(t *testing.T) {
	srv, _ := newAPITestServer(t)
	sessID, adminToken := mustCreateSession(t, srv, "")
	playerID := mustJoinSession(t, srv, sessID, "Alice", adminToken)

	res := patchReq(t, srv, "/api/sessions/"+sessID+"/players/"+playerID+"/rating", map[string]any{
		"rating": 5,
	}, "")
	if res.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 without admin token, got %d", res.StatusCode)
	}
	_ = res.Body.Close()
}

func TestUpdatePlayerRating_CreatorTokenNotAuthorized(t *testing.T) {
	srv, _ := newAPITestServer(t)
	// The creating user owns the session (CreatorUserID) but that alone must not
	// authorize a rating edit — only the AdminToken gates it (#211).
	userToken := mustRegister(t, srv, "creator@test.local", "Creator", "password123")
	sessID, adminToken := mustCreateSession(t, srv, userToken)
	playerID := mustJoinSession(t, srv, sessID, "Alice", adminToken)

	res := patchReq(t, srv, "/api/sessions/"+sessID+"/players/"+playerID+"/rating", map[string]any{
		"rating": 5,
	}, userToken)
	if res.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 for creator token without admin token, got %d", res.StatusCode)
	}
	_ = res.Body.Close()
}

func TestUpdatePlayerRating_InvalidRating(t *testing.T) {
	srv, _ := newAPITestServer(t)
	sessID, adminToken := mustCreateSession(t, srv, "")
	playerID := mustJoinSession(t, srv, sessID, "Alice", adminToken)

	res := patchReq(t, srv, "/api/sessions/"+sessID+"/players/"+playerID+"/rating", map[string]any{
		"rating": 9,
	}, adminToken)
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for out-of-range rating, got %d", res.StatusCode)
	}
	_ = res.Body.Close()
}

func TestUpdatePlayerRating_PlayerNotFound(t *testing.T) {
	srv, _ := newAPITestServer(t)
	sessID, adminToken := mustCreateSession(t, srv, "")

	res := patchReq(t, srv, "/api/sessions/"+sessID+"/players/NOPE/rating", map[string]any{
		"rating": 4,
	}, adminToken)
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown player, got %d", res.StatusCode)
	}
	_ = res.Body.Close()
}

func TestUpdatePlayerRating_RegisteredPlayerNotEditable(t *testing.T) {
	srv, _ := newAPITestServer(t)
	userToken := mustRegister(t, srv, "player@test.local", "Player", "password123")
	sessID, adminToken := mustCreateSession(t, srv, "")

	// A registered user joins with their account — they own their own rating, so
	// the admin cannot edit it (#211).
	res := postReq(t, srv, "/api/sessions/"+sessID+"/players", map[string]any{"name": "Player"}, userToken)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("join: expected 201, got %d", res.StatusCode)
	}
	var player struct {
		ID string `json:"id"`
	}
	decodeBody(t, res, &player)

	res = patchReq(t, srv, "/api/sessions/"+sessID+"/players/"+player.ID+"/rating", map[string]any{
		"rating": 5,
	}, adminToken)
	if res.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 editing a registered player, got %d", res.StatusCode)
	}
	var body struct {
		Error string `json:"error"`
	}
	decodeBody(t, res, &body)
	if body.Error != "rating_not_editable" {
		t.Errorf("expected error 'rating_not_editable', got %q", body.Error)
	}
}

func TestUpdatePlayerRating_SelfJoinedGuestNotEditable(t *testing.T) {
	srv, _ := newAPITestServer(t)
	sessID, adminToken := mustCreateSession(t, srv, "")

	// A guest who self-joins by link (no admin token) picked their own level, so
	// the admin cannot edit it — only admin-added guests are editable (#211).
	res := postReq(t, srv, "/api/sessions/"+sessID+"/players", map[string]any{"name": "Guest", "rating": 2}, "")
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("join: expected 201, got %d", res.StatusCode)
	}
	var player struct {
		ID string `json:"id"`
	}
	decodeBody(t, res, &player)

	res = patchReq(t, srv, "/api/sessions/"+sessID+"/players/"+player.ID+"/rating", map[string]any{
		"rating": 5,
	}, adminToken)
	if res.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 editing a self-joined guest, got %d", res.StatusCode)
	}
	_ = res.Body.Close()
}

func TestDeactivatePlayer(t *testing.T) {
	srv, _ := newAPITestServer(t)
	sessID, adminToken := mustCreateSession(t, srv, "")

	playerID := mustJoinSession(t, srv, sessID, "Alice", adminToken)

	res := deleteReq(t, srv, "/api/sessions/"+sessID+"/players/"+playerID, adminToken)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
	var body struct {
		Active bool `json:"active"`
	}
	decodeBody(t, res, &body)
	if body.Active {
		t.Error("expected player to be inactive after deactivation")
	}
}

func TestDeactivatePlayer_RequiresAdmin(t *testing.T) {
	srv, _ := newAPITestServer(t)
	sessID, adminToken := mustCreateSession(t, srv, "")
	playerID := mustJoinSession(t, srv, sessID, "Alice", adminToken)

	res := deleteReq(t, srv, "/api/sessions/"+sessID+"/players/"+playerID, "")
	if res.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", res.StatusCode)
	}
	res.Body.Close()
}

func TestDeactivatePlayer_SelfRemoval(t *testing.T) {
	srv, _ := newAPITestServer(t)
	sessID, _ := mustCreateSession(t, srv, "")
	playerID, playerToken := mustJoinSessionWithToken(t, srv, sessID, "Alice")

	// Self-removal proven by the per-player secret (no admin token needed).
	res := doRequest(t, srv, http.MethodDelete, "/api/sessions/"+sessID+"/players/"+playerID, nil, "", map[string]string{
		"X-Player-Token": playerToken,
	})
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for self-removal, got %d", res.StatusCode)
	}
	res.Body.Close()
}

// The player id must not authorize removal — it's visible to everyone in the
// lobby. Only the per-player secret (or an admin token) may deactivate (#241).
func TestDeactivatePlayer_SelfRemoval_RejectsSpoofedID(t *testing.T) {
	srv, _ := newAPITestServer(t)
	sessID, _ := mustCreateSession(t, srv, "")
	playerID, playerToken := mustJoinSessionWithToken(t, srv, sessID, "Alice")

	// The old spoofable header (the player id) must no longer grant removal.
	res := doRequest(t, srv, http.MethodDelete, "/api/sessions/"+sessID+"/players/"+playerID, nil, "", map[string]string{
		"X-Player-Id": playerID,
	})
	if res.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 for spoofed player id, got %d", res.StatusCode)
	}
	_ = res.Body.Close()

	// A wrong secret is likewise rejected.
	res = doRequest(t, srv, http.MethodDelete, "/api/sessions/"+sessID+"/players/"+playerID, nil, "", map[string]string{
		"X-Player-Token": playerToken + "x",
	})
	if res.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 for wrong token, got %d", res.StatusCode)
	}
	_ = res.Body.Close()
}

// The self-removal secret must never appear in the shared session listing —
// only in the join response for the joining client (#241).
func TestGetSession_OmitsPlayerToken(t *testing.T) {
	srv, _ := newAPITestServer(t)
	sessID, _ := mustCreateSession(t, srv, "")
	mustJoinSessionWithToken(t, srv, sessID, "Alice")

	res := getReq(t, srv, "/api/sessions/"+sessID, "")
	var sess struct {
		Players []map[string]any `json:"players"`
	}
	decodeBody(t, res, &sess)
	if len(sess.Players) == 0 {
		t.Fatal("expected at least one player")
	}
	for _, p := range sess.Players {
		if _, ok := p["player_token"]; ok {
			t.Errorf("player_token leaked in session listing: %v", p["player_token"])
		}
	}
}

// countActivePlayers reads the session and returns how many of its players are
// still active.
func countActivePlayers(t *testing.T, srv *httptest.Server, sessID string) int {
	t.Helper()
	res := getReq(t, srv, "/api/sessions/"+sessID, "")
	if res.StatusCode != http.StatusOK {
		t.Fatalf("getSession: expected 200, got %d", res.StatusCode)
	}
	var sess struct {
		Players []struct {
			Active bool `json:"active"`
		} `json:"players"`
	}
	decodeBody(t, res, &sess)
	n := 0
	for _, p := range sess.Players {
		if p.Active {
			n++
		}
	}
	return n
}

// A registered user can leave a session using only their auth token — no stored
// player id — which is what makes it work retroactively for legacy sessions.
func TestLeaveSession_RegisteredUser(t *testing.T) {
	srv, _ := newAPITestServer(t)
	token := mustRegister(t, srv, "alice@example.com", "Alice", "password123")
	sessID, _ := mustCreateSession(t, srv, "")

	// Alice joins as a registered user (player carries her user_id).
	mustJoinSession(t, srv, sessID, "Alice", token)
	if got := countActivePlayers(t, srv, sessID); got != 1 {
		t.Fatalf("expected 1 active player after join, got %d", got)
	}

	res := postReq(t, srv, "/api/sessions/"+sessID+"/leave", nil, token)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("leave: expected 200, got %d", res.StatusCode)
	}
	_ = res.Body.Close()

	if got := countActivePlayers(t, srv, sessID); got != 0 {
		t.Fatalf("expected 0 active players after leave, got %d", got)
	}
}

func TestLeaveSession_NotAMember(t *testing.T) {
	srv, _ := newAPITestServer(t)
	token := mustRegister(t, srv, "alice@example.com", "Alice", "password123")
	sessID, adminToken := mustCreateSession(t, srv, "")
	mustJoinSession(t, srv, sessID, "Bob", adminToken) // someone else, a guest

	res := postReq(t, srv, "/api/sessions/"+sessID+"/leave", nil, token)
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 when not a member, got %d", res.StatusCode)
	}
	_ = res.Body.Close()
}

func TestLeaveSession_RequiresAuth(t *testing.T) {
	srv, _ := newAPITestServer(t)
	sessID, adminToken := mustCreateSession(t, srv, "")
	mustJoinSession(t, srv, sessID, "Alice", adminToken)

	res := postReq(t, srv, "/api/sessions/"+sessID+"/leave", nil, "")
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 without auth, got %d", res.StatusCode)
	}
	_ = res.Body.Close()
}

func TestLeaveSession_AfterStartBlocked(t *testing.T) {
	srv, _ := newAPITestServer(t)
	token := mustRegister(t, srv, "alice@example.com", "Alice", "password123")
	sessID, adminToken := mustCreateSession(t, srv, "")

	// Alice (registered) plus three guests fill the single court, then start.
	mustJoinSession(t, srv, sessID, "Alice", token)
	mustJoinSession(t, srv, sessID, "Bob", adminToken)
	mustJoinSession(t, srv, sessID, "Charlie", adminToken)
	mustJoinSession(t, srv, sessID, "Diana", adminToken)
	mustStartSession(t, srv, sessID, adminToken)

	res := postReq(t, srv, "/api/sessions/"+sessID+"/leave", nil, token)
	if res.StatusCode != http.StatusConflict {
		t.Fatalf("expected 409 after start, got %d", res.StatusCode)
	}
	_ = res.Body.Close()
}

func TestDeactivatePlayer_SessionAlreadyStarted(t *testing.T) {
	srv, _ := newAPITestServer(t)
	sessID, adminToken, playerIDs := setupStartedSession(t, srv)

	res := deleteReq(t, srv, "/api/sessions/"+sessID+"/players/"+playerIDs[1], adminToken)
	if res.StatusCode != http.StatusConflict {
		t.Fatalf("expected 409 for started session, got %d", res.StatusCode)
	}
	res.Body.Close()
}

// The creator administers the session (they hold the AdminToken), so they can't
// leave their own roster and orphan it — they must cancel instead. Joining as the
// creator crowns their CreatorPlayer, which the guard keys on.
func TestLeaveSession_CreatorBlocked(t *testing.T) {
	srv, _ := newAPITestServer(t)
	token := mustRegister(t, srv, "alice@example.com", "Alice", "password123")
	sessID, _ := mustCreateSession(t, srv, token) // Alice is the creator
	mustJoinSession(t, srv, sessID, "Alice", token)

	res := postReq(t, srv, "/api/sessions/"+sessID+"/leave", nil, token)
	if res.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("creator leave: expected 422, got %d", res.StatusCode)
	}
	var errBody struct {
		Error string `json:"error"`
	}
	decodeBody(t, res, &errBody)
	if errBody.Error != "creator_cannot_leave" {
		t.Errorf("expected creator_cannot_leave, got %q", errBody.Error)
	}
	if got := countActivePlayers(t, srv, sessID); got != 1 {
		t.Errorf("expected creator still on roster, got %d active", got)
	}
}

// The creator's Player can't be removed even by an admin token — same roster-
// integrity guard, from the admin-initiated removal path.
func TestDeactivatePlayer_CreatorBlocked(t *testing.T) {
	srv, _ := newAPITestServer(t)
	token := mustRegister(t, srv, "alice@example.com", "Alice", "password123")
	sessID, adminToken := mustCreateSession(t, srv, token)
	playerID := mustJoinSession(t, srv, sessID, "Alice", token)

	res := deleteReq(t, srv, "/api/sessions/"+sessID+"/players/"+playerID, adminToken)
	if res.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("remove creator: expected 422, got %d", res.StatusCode)
	}
	var errBody struct {
		Error string `json:"error"`
	}
	decodeBody(t, res, &errBody)
	if errBody.Error != "creator_cannot_leave" {
		t.Errorf("expected creator_cannot_leave, got %q", errBody.Error)
	}
}

// A non-creator player can still leave a creator-owned session normally — the
// guard is scoped to the creator's own Player, not the whole roster.
func TestLeaveSession_NonCreatorStillLeaves(t *testing.T) {
	srv, _ := newAPITestServer(t)
	aliceToken := mustRegister(t, srv, "alice@example.com", "Alice", "password123")
	bobToken := mustRegister(t, srv, "bob@example.com", "Bob", "password123")
	sessID, _ := mustCreateSession(t, srv, aliceToken)
	mustJoinSession(t, srv, sessID, "Alice", aliceToken) // creator
	mustJoinSession(t, srv, sessID, "Bob", bobToken)     // regular member

	res := postReq(t, srv, "/api/sessions/"+sessID+"/leave", nil, bobToken)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("non-creator leave: expected 200, got %d", res.StatusCode)
	}
	_ = res.Body.Close()
	if got := countActivePlayers(t, srv, sessID); got != 1 {
		t.Errorf("expected only creator left on roster, got %d active", got)
	}
}
