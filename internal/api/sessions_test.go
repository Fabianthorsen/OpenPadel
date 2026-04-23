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
