package api_test

import (
	"net/http"
	"testing"
)

func TestCreateSession(t *testing.T) {
	srv, _ := newAPITestServer(t)
	res := postReq(t, srv, "/api/sessions", map[string]any{
		"courts":    1,
		"points":    24,
		"game_mode": "americano",
	}, "")
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", res.StatusCode)
	}
	var body struct {
		ID         string `json:"id"`
		AdminToken string `json:"admin_token"`
		Status     string `json:"status"`
		GameMode   string `json:"game_mode"`
	}
	decodeBody(t, res, &body)
	if body.ID == "" {
		t.Error("expected non-empty session ID")
	}
	if body.AdminToken == "" {
		t.Error("expected non-empty admin token")
	}
	if body.Status != "lobby" {
		t.Errorf("expected status 'lobby', got %q", body.Status)
	}
}

func TestCreateSession_InvalidGameMode(t *testing.T) {
	srv, _ := newAPITestServer(t)
	res := postReq(t, srv, "/api/sessions", map[string]any{
		"courts":    1,
		"points":    24,
		"game_mode": "invalid",
	}, "")
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.StatusCode)
	}
	res.Body.Close()
}

func TestCreateSession_InvalidPoints(t *testing.T) {
	srv, _ := newAPITestServer(t)
	res := postReq(t, srv, "/api/sessions", map[string]any{
		"courts":    1,
		"points":    30,
		"game_mode": "americano",
	}, "")
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.StatusCode)
	}
	res.Body.Close()
}

func TestGetSession(t *testing.T) {
	srv, _ := newAPITestServer(t)
	_, adminToken := mustCreateSession(t, srv, "")

	var sessBody struct {
		ID string `json:"id"`
	}
	res := postReq(t, srv, "/api/sessions", map[string]any{
		"courts":    1,
		"points":    24,
		"game_mode": "americano",
	}, "")
	decodeBody(t, res, &sessBody)

	res2 := getReq(t, srv, "/api/sessions/"+sessBody.ID, adminToken)
	if res2.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res2.StatusCode)
	}
	var session struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	decodeBody(t, res2, &session)
	if session.ID != sessBody.ID {
		t.Errorf("expected session ID %q, got %q", sessBody.ID, session.ID)
	}
}

func TestGetSession_NotFound(t *testing.T) {
	srv, _ := newAPITestServer(t)
	res := getReq(t, srv, "/api/sessions/XXXX", "")
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", res.StatusCode)
	}
	res.Body.Close()
}

func TestGetSession_HidesAdminTokenFromNonAdmin(t *testing.T) {
	srv, _ := newAPITestServer(t)
	sessID, _ := mustCreateSession(t, srv, "")

	res := getReq(t, srv, "/api/sessions/"+sessID, "")
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
	var session struct {
		AdminToken string `json:"admin_token"`
	}
	decodeBody(t, res, &session)
	if session.AdminToken != "" {
		t.Error("admin_token should be hidden from non-admin requests")
	}
}

func TestStartSession(t *testing.T) {
	srv, _ := newAPITestServer(t)
	sessID, adminToken, _ := setupStartedSession(t, srv)

	res := getReq(t, srv, "/api/sessions/"+sessID, adminToken)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
	var session struct {
		Status string `json:"status"`
	}
	decodeBody(t, res, &session)
	if session.Status != "playing" {
		t.Errorf("expected status 'playing' after start, got %q", session.Status)
	}
}

func TestStartSession_NotEnoughPlayers(t *testing.T) {
	srv, _ := newAPITestServer(t)
	sessID, adminToken := mustCreateSession(t, srv, "")

	mustJoinSession(t, srv, sessID, "Alice", adminToken)

	res := postReq(t, srv, "/api/sessions/"+sessID+"/start", nil, adminToken)
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.StatusCode)
	}
	res.Body.Close()
}

func TestStartSession_RequiresAdmin(t *testing.T) {
	srv, _ := newAPITestServer(t)
	sessID, adminToken := mustCreateSession(t, srv, "")

	mustJoinSession(t, srv, sessID, "Alice", adminToken)
	mustJoinSession(t, srv, sessID, "Bob", "")
	mustJoinSession(t, srv, sessID, "Charlie", "")
	mustJoinSession(t, srv, sessID, "Diana", "")

	res := postReq(t, srv, "/api/sessions/"+sessID+"/start", nil, "")
	if res.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", res.StatusCode)
	}
	res.Body.Close()
}

func TestCloseSession(t *testing.T) {
	srv, _ := newAPITestServer(t)
	sessID, adminToken, _ := setupStartedSession(t, srv)

	res := postReq(t, srv, "/api/sessions/"+sessID+"/close", nil, adminToken)
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", res.StatusCode)
	}
	res.Body.Close()

	res2 := getReq(t, srv, "/api/sessions/"+sessID, adminToken)
	var session struct {
		Status string `json:"status"`
	}
	decodeBody(t, res2, &session)
	if session.Status != "done" {
		t.Errorf("expected status 'done' after close, got %q", session.Status)
	}
}

func TestCloseSession_RequiresAdmin(t *testing.T) {
	srv, _ := newAPITestServer(t)
	sessID, _, _ := setupStartedSession(t, srv)

	res := postReq(t, srv, "/api/sessions/"+sessID+"/close", nil, "wrong-token")
	if res.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", res.StatusCode)
	}
	res.Body.Close()
}

func TestCancelSession(t *testing.T) {
	srv, _ := newAPITestServer(t)
	sessID, adminToken := mustCreateSession(t, srv, "")

	res := deleteReq(t, srv, "/api/sessions/"+sessID, adminToken)
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", res.StatusCode)
	}
	res.Body.Close()

	res2 := getReq(t, srv, "/api/sessions/"+sessID, "")
	if res2.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 after cancel, got %d", res2.StatusCode)
	}
	res2.Body.Close()
}

func TestCancelSession_RequiresAdmin(t *testing.T) {
	srv, _ := newAPITestServer(t)
	sessID, _ := mustCreateSession(t, srv, "")

	res := deleteReq(t, srv, "/api/sessions/"+sessID, "")
	if res.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", res.StatusCode)
	}
	res.Body.Close()
}

// Integration tests for state machine & validation (Issue #66)

func TestStateFlow_LobbyToPlayingToDone(t *testing.T) {
	srv, _ := newAPITestServer(t)
	sessID, adminToken, _ := setupStartedSession(t, srv)

	// Verify session is in playing state
	res := getReq(t, srv, "/api/sessions/"+sessID, adminToken)
	var sess struct {
		Status string `json:"status"`
	}
	decodeBody(t, res, &sess)
	if sess.Status != "playing" {
		t.Fatalf("expected status 'playing', got %q", sess.Status)
	}

	// Close the session (manual end)
	res2 := postReq(t, srv, "/api/sessions/"+sessID+"/close", nil, adminToken)
	if res2.StatusCode != http.StatusNoContent {
		t.Fatalf("close: expected 204, got %d", res2.StatusCode)
	}
	res2.Body.Close()

	// Verify session is now in done state
	res3 := getReq(t, srv, "/api/sessions/"+sessID, adminToken)
	decodeBody(t, res3, &sess)
	if sess.Status != "done" {
		t.Fatalf("expected status 'done' after close, got %q", sess.Status)
	}
}

func TestStateFlow_DeleteInLobby(t *testing.T) {
	srv, _ := newAPITestServer(t)
	sessID, adminToken := mustCreateSession(t, srv, "")

	// Join some players
	mustJoinSession(t, srv, sessID, "Alice", adminToken)
	mustJoinSession(t, srv, sessID, "Bob", "")

	// Delete the session in lobby state
	res := deleteReq(t, srv, "/api/sessions/"+sessID, adminToken)
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", res.StatusCode)
	}
	res.Body.Close()

	// Verify session no longer exists
	res2 := getReq(t, srv, "/api/sessions/"+sessID, "")
	if res2.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 after delete, got %d", res2.StatusCode)
	}
	res2.Body.Close()
}

func TestAmericanoConstraints_StartSucceeds_1Court4Players(t *testing.T) {
	srv, _ := newAPITestServer(t)
	sessID, adminToken := mustCreateSessionWithParams(t, srv, "", 1, 24, "americano")

	// Join 4 players for 1 court (minimum for Americano)
	mustJoinSession(t, srv, sessID, "Alice", adminToken)
	mustJoinSession(t, srv, sessID, "Bob", "")
	mustJoinSession(t, srv, sessID, "Charlie", "")
	mustJoinSession(t, srv, sessID, "Diana", "")

	// Should succeed
	res := postReq(t, srv, "/api/sessions/"+sessID+"/start", nil, adminToken)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
	res.Body.Close()

	// Verify status changed to playing
	res2 := getReq(t, srv, "/api/sessions/"+sessID, adminToken)
	var sess struct {
		Status string `json:"status"`
	}
	decodeBody(t, res2, &sess)
	if sess.Status != "playing" {
		t.Fatalf("expected 'playing', got %q", sess.Status)
	}
}

func TestAmericanoConstraints_StartFails_1Court3Players(t *testing.T) {
	srv, _ := newAPITestServer(t)
	sessID, adminToken := mustCreateSessionWithParams(t, srv, "", 1, 24, "americano")

	// Join only 3 players for 1 court (below minimum)
	mustJoinSession(t, srv, sessID, "Alice", adminToken)
	mustJoinSession(t, srv, sessID, "Bob", "")
	mustJoinSession(t, srv, sessID, "Charlie", "")

	// Should fail with 400
	res := postReq(t, srv, "/api/sessions/"+sessID+"/start", nil, adminToken)
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.StatusCode)
	}
	res.Body.Close()
}

func TestAmericanoConstraints_StartSucceeds_2Courts8Players(t *testing.T) {
	srv, _ := newAPITestServer(t)
	sessID, adminToken := mustCreateSessionWithParams(t, srv, "", 2, 24, "americano")

	// Join 8 players for 2 courts
	for i, name := range []string{"Alice", "Bob", "Charlie", "Diana", "Eve", "Frank", "Grace", "Henry"} {
		token := ""
		if i == 0 {
			token = adminToken
		}
		mustJoinSession(t, srv, sessID, name, token)
	}

	// Should succeed
	res := postReq(t, srv, "/api/sessions/"+sessID+"/start", nil, adminToken)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
	res.Body.Close()

	// Verify status changed to playing
	res2 := getReq(t, srv, "/api/sessions/"+sessID, adminToken)
	var sess struct {
		Status string `json:"status"`
	}
	decodeBody(t, res2, &sess)
	if sess.Status != "playing" {
		t.Fatalf("expected 'playing', got %q", sess.Status)
	}
}

func TestMexicanoConstraints_StartSucceeds_2Courts8Players(t *testing.T) {
	srv, _ := newAPITestServer(t)
	sessID, adminToken := mustCreateSessionWithParams(t, srv, "", 2, 24, "mexicano")

	// Join exactly 8 players for 2 courts in Mexicano
	for i, name := range []string{"Alice", "Bob", "Charlie", "Diana", "Eve", "Frank", "Grace", "Henry"} {
		token := ""
		if i == 0 {
			token = adminToken
		}
		mustJoinSession(t, srv, sessID, name, token)
	}

	// Should succeed
	res := postReq(t, srv, "/api/sessions/"+sessID+"/start", nil, adminToken)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
	res.Body.Close()

	// Verify status changed to playing
	res2 := getReq(t, srv, "/api/sessions/"+sessID, adminToken)
	var sess struct {
		Status string `json:"status"`
	}
	decodeBody(t, res2, &sess)
	if sess.Status != "playing" {
		t.Fatalf("expected 'playing', got %q", sess.Status)
	}
}

func TestMexicanoConstraints_StartFails_2Courts7Players(t *testing.T) {
	srv, _ := newAPITestServer(t)
	sessID, adminToken := mustCreateSessionWithParams(t, srv, "", 2, 24, "mexicano")

	// Join 7 players (below required 8)
	for i, name := range []string{"Alice", "Bob", "Charlie", "Diana", "Eve", "Frank", "Grace"} {
		token := ""
		if i == 0 {
			token = adminToken
		}
		mustJoinSession(t, srv, sessID, name, token)
	}

	// Should fail with 400
	res := postReq(t, srv, "/api/sessions/"+sessID+"/start", nil, adminToken)
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.StatusCode)
	}
	res.Body.Close()
}

func TestMexicanoConstraints_StartFails_2Courts9Players(t *testing.T) {
	srv, _ := newAPITestServer(t)
	sessID, adminToken := mustCreateSessionWithParams(t, srv, "", 2, 24, "mexicano")

	// Join 9 players (above maximum 8)
	for i, name := range []string{"Alice", "Bob", "Charlie", "Diana", "Eve", "Frank", "Grace", "Henry", "Ivy"} {
		token := ""
		if i == 0 {
			token = adminToken
		}
		mustJoinSession(t, srv, sessID, name, token)
	}

	// Should fail with 400
	res := postReq(t, srv, "/api/sessions/"+sessID+"/start", nil, adminToken)
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.StatusCode)
	}
	res.Body.Close()
}

func TestInvalidTransition_CanCloseInLobby(t *testing.T) {
	srv, _ := newAPITestServer(t)
	sessID, adminToken := mustCreateSession(t, srv, "")

	// Verify can close in lobby state
	res := postReq(t, srv, "/api/sessions/"+sessID+"/close", nil, adminToken)
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", res.StatusCode)
	}
	res.Body.Close()

	// Verify status is done
	res2 := getReq(t, srv, "/api/sessions/"+sessID, adminToken)
	var sess struct {
		Status string `json:"status"`
	}
	decodeBody(t, res2, &sess)
	if sess.Status != "done" {
		t.Fatalf("expected 'done', got %q", sess.Status)
	}
}

func TestInvalidTransition_CannotCloseInDone(t *testing.T) {
	srv, _ := newAPITestServer(t)
	sessID, adminToken, _ := setupStartedSession(t, srv)

	// Close the session
	res := postReq(t, srv, "/api/sessions/"+sessID+"/close", nil, adminToken)
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("close: expected 204, got %d", res.StatusCode)
	}
	res.Body.Close()

	// Try to close again - should fail with 409 (conflict)
	res2 := postReq(t, srv, "/api/sessions/"+sessID+"/close", nil, adminToken)
	if res2.StatusCode != http.StatusConflict {
		t.Fatalf("expected 409, got %d", res2.StatusCode)
	}
	res2.Body.Close()
}

func TestInvalidTransition_CannotStartInPlaying(t *testing.T) {
	srv, _ := newAPITestServer(t)
	sessID, adminToken, _ := setupStartedSession(t, srv)

	// Try to start again - should fail with 409 (conflict)
	res := postReq(t, srv, "/api/sessions/"+sessID+"/start", nil, adminToken)
	if res.StatusCode != http.StatusConflict {
		t.Fatalf("expected 409, got %d", res.StatusCode)
	}
	res.Body.Close()
}

func TestPlayerJoin_CanJoinInLobby(t *testing.T) {
	srv, _ := newAPITestServer(t)
	sessID, _ := mustCreateSession(t, srv, "")

	// Player should be able to join in lobby state
	res := postReq(t, srv, "/api/sessions/"+sessID+"/players", map[string]any{"name": "Alice"}, "")
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", res.StatusCode)
	}
	var player struct {
		ID string `json:"id"`
	}
	decodeBody(t, res, &player)
	if player.ID == "" {
		t.Error("expected non-empty player ID")
	}
}

func TestPlayerJoin_CannotJoinInPlaying(t *testing.T) {
	srv, _ := newAPITestServer(t)
	sessID, _, _ := setupStartedSession(t, srv)

	// Player should NOT be able to join once session is playing
	res := postReq(t, srv, "/api/sessions/"+sessID+"/players", map[string]any{"name": "Eve"}, "")
	if res.StatusCode != http.StatusConflict {
		t.Fatalf("expected 409, got %d", res.StatusCode)
	}
	res.Body.Close()
}

func TestPlayerJoin_CannotJoinInDone(t *testing.T) {
	srv, _ := newAPITestServer(t)
	sessID, adminToken, _ := setupStartedSession(t, srv)

	// Close the session
	res := postReq(t, srv, "/api/sessions/"+sessID+"/close", nil, adminToken)
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("close: expected 204, got %d", res.StatusCode)
	}
	res.Body.Close()

	// Player should NOT be able to join in done state
	res2 := postReq(t, srv, "/api/sessions/"+sessID+"/players", map[string]any{"name": "Eve"}, "")
	if res2.StatusCode != http.StatusConflict {
		t.Fatalf("expected 409, got %d", res2.StatusCode)
	}
	res2.Body.Close()
}
