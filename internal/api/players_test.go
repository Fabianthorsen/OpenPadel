package api_test

import (
	"net/http"
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
	sessID, adminToken := mustCreateSession(t, srv, "")
	playerID := mustJoinSession(t, srv, sessID, "Alice", adminToken)

	// Self-removal via X-Player-Id header (no admin token needed)
	res := doRequest(t, srv, http.MethodDelete, "/api/sessions/"+sessID+"/players/"+playerID, nil, "", map[string]string{
		"X-Player-Id": playerID,
	})
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for self-removal, got %d", res.StatusCode)
	}
	res.Body.Close()
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
