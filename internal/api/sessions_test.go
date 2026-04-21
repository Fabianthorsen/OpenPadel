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
	if session.Status != "active" {
		t.Errorf("expected status 'active' after start, got %q", session.Status)
	}
}

func TestStartSession_NotEnoughPlayers(t *testing.T) {
	srv, _ := newAPITestServer(t)
	sessID, adminToken := mustCreateSession(t, srv, "")

	mustJoinSession(t, srv, sessID, "Alice", adminToken)

	res := postReq(t, srv, "/api/sessions/"+sessID+"/start", nil, adminToken)
	if res.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", res.StatusCode)
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
	if session.Status != "complete" {
		t.Errorf("expected status 'complete' after close, got %q", session.Status)
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

// Timed Americano Tests

func TestCreateSession_TimedAmericano(t *testing.T) {
	srv, _ := newAPITestServer(t)
	res := postReq(t, srv, "/api/sessions", map[string]any{
		"courts":                   1,
		"game_mode":                "timed_americano",
		"total_duration_minutes":   120,
		"buffer_seconds":           120,
	}, "")
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", res.StatusCode)
	}
	var body struct {
		ID                   string `json:"id"`
		GameMode             string `json:"game_mode"`
		TotalDurationMinutes *int   `json:"total_duration_minutes"`
		BufferSeconds        *int   `json:"buffer_seconds"`
		Points               int    `json:"points"`
	}
	decodeBody(t, res, &body)
	if body.GameMode != "timed_americano" {
		t.Errorf("expected game_mode='timed_americano', got %q", body.GameMode)
	}
	if body.TotalDurationMinutes == nil || *body.TotalDurationMinutes != 120 {
		t.Errorf("expected TotalDurationMinutes=120, got %v", body.TotalDurationMinutes)
	}
	if body.BufferSeconds == nil || *body.BufferSeconds != 120 {
		t.Errorf("expected BufferSeconds=120, got %v", body.BufferSeconds)
	}
	if body.Points != 0 {
		t.Errorf("expected Points=0 for timed_americano, got %d", body.Points)
	}
}

func TestCreateSession_TimedAmericano_DefaultBuffer(t *testing.T) {
	srv, _ := newAPITestServer(t)
	res := postReq(t, srv, "/api/sessions", map[string]any{
		"courts":                   1,
		"game_mode":                "timed_americano",
		"total_duration_minutes":   120,
	}, "")
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", res.StatusCode)
	}
	var body struct {
		BufferSeconds *int `json:"buffer_seconds"`
	}
	decodeBody(t, res, &body)
	if body.BufferSeconds == nil || *body.BufferSeconds != 120 {
		t.Errorf("expected BufferSeconds=120 (default), got %v", body.BufferSeconds)
	}
}

func TestCreateSession_TimedAmericano_InvalidDuration(t *testing.T) {
	srv, _ := newAPITestServer(t)
	res := postReq(t, srv, "/api/sessions", map[string]any{
		"courts":                   1,
		"game_mode":                "timed_americano",
		"total_duration_minutes":   5,
		"buffer_seconds":           120,
	}, "")
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid duration, got %d", res.StatusCode)
	}
	res.Body.Close()
}

func TestCreateSession_TimedAmericano_InvalidBuffer(t *testing.T) {
	srv, _ := newAPITestServer(t)
	res := postReq(t, srv, "/api/sessions", map[string]any{
		"courts":                   1,
		"game_mode":                "timed_americano",
		"total_duration_minutes":   120,
		"buffer_seconds":           30,
	}, "")
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid buffer, got %d", res.StatusCode)
	}
	res.Body.Close()
}

func TestCreateSession_TimedAmericano_PointsRejected(t *testing.T) {
	srv, _ := newAPITestServer(t)
	res := postReq(t, srv, "/api/sessions", map[string]any{
		"courts":                   1,
		"game_mode":                "timed_americano",
		"total_duration_minutes":   120,
		"buffer_seconds":           120,
		"points":                   24,
	}, "")
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 when points provided, got %d", res.StatusCode)
	}
	res.Body.Close()
}

func TestStartSession_TimedAmericano(t *testing.T) {
	srv, _ := newAPITestServer(t)
	userToken := mustRegister(t, srv, "admin@test.local", "Admin", "password123")
	sessID, adminToken := mustCreateTimedAmericanoSession(t, srv, userToken)

	// Join 4 players
	mustJoinSession(t, srv, sessID, "Alice", userToken)
	mustJoinSession(t, srv, sessID, "Bob", "")
	mustJoinSession(t, srv, sessID, "Charlie", "")
	mustJoinSession(t, srv, sessID, "Diana", "")

	// Start session
	res := postReq(t, srv, "/api/sessions/"+sessID+"/start", nil, adminToken)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
	var sess struct {
		Status               string `json:"status"`
		RoundsTotal          *int   `json:"rounds_total"`
		RoundDurationSeconds *int   `json:"round_duration_seconds"`
		RoundStartedAt       *string `json:"round_started_at"`
		CurrentRound         *int   `json:"current_round"`
	}
	decodeBody(t, res, &sess)
	if sess.Status != "active" {
		t.Errorf("expected status 'active', got %q", sess.Status)
	}
	if sess.RoundsTotal == nil || *sess.RoundsTotal <= 0 {
		t.Errorf("expected RoundsTotal > 0, got %v", sess.RoundsTotal)
	}
	if sess.RoundDurationSeconds == nil || *sess.RoundDurationSeconds <= 0 {
		t.Errorf("expected RoundDurationSeconds > 0, got %v", sess.RoundDurationSeconds)
	}
	if sess.RoundStartedAt == nil {
		t.Errorf("expected RoundStartedAt to be set")
	}
	if sess.CurrentRound == nil || *sess.CurrentRound != 1 {
		t.Errorf("expected CurrentRound=1, got %v", sess.CurrentRound)
	}
}

func TestStartSession_TimedAmericano_NotEnoughPlayers(t *testing.T) {
	srv, _ := newAPITestServer(t)
	userToken := mustRegister(t, srv, "admin@test.local", "Admin", "password123")
	sessID, adminToken := mustCreateTimedAmericanoSession(t, srv, userToken)

	// Join only 3 players (need 4 for 1 court)
	mustJoinSession(t, srv, sessID, "Alice", userToken)
	mustJoinSession(t, srv, sessID, "Bob", "")
	mustJoinSession(t, srv, sessID, "Charlie", "")

	// Start session
	res := postReq(t, srv, "/api/sessions/"+sessID+"/start", nil, adminToken)
	if res.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", res.StatusCode)
	}
	res.Body.Close()
}

func TestCreateSession_TimedAmericano_DefaultInterval(t *testing.T) {
	srv, _ := newAPITestServer(t)
	res := postReq(t, srv, "/api/sessions", map[string]any{
		"courts":                   1,
		"game_mode":                "timed_americano",
		"total_duration_minutes":   120,
		"buffer_seconds":           120,
	}, "")
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", res.StatusCode)
	}
	var body struct {
		IntervalBetweenRoundsMin *int `json:"interval_between_rounds_minutes"`
	}
	decodeBody(t, res, &body)
	if body.IntervalBetweenRoundsMin == nil || *body.IntervalBetweenRoundsMin != 3 {
		t.Errorf("expected IntervalBetweenRoundsMin=3 (default), got %v", body.IntervalBetweenRoundsMin)
	}
}

func TestCreateSession_TimedAmericano_CustomInterval(t *testing.T) {
	srv, _ := newAPITestServer(t)
	res := postReq(t, srv, "/api/sessions", map[string]any{
		"courts":                      1,
		"game_mode":                   "timed_americano",
		"total_duration_minutes":      120,
		"buffer_seconds":              120,
		"interval_between_rounds_minutes": 5,
	}, "")
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", res.StatusCode)
	}
	var body struct {
		IntervalBetweenRoundsMin *int `json:"interval_between_rounds_minutes"`
	}
	decodeBody(t, res, &body)
	if body.IntervalBetweenRoundsMin == nil || *body.IntervalBetweenRoundsMin != 5 {
		t.Errorf("expected IntervalBetweenRoundsMin=5, got %v", body.IntervalBetweenRoundsMin)
	}
}

func TestCreateSession_TimedAmericano_IntervalTooLow(t *testing.T) {
	srv, _ := newAPITestServer(t)
	res := postReq(t, srv, "/api/sessions", map[string]any{
		"courts":                      1,
		"game_mode":                   "timed_americano",
		"total_duration_minutes":      120,
		"buffer_seconds":              120,
		"interval_between_rounds_minutes": 0,
	}, "")
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for interval < 1, got %d", res.StatusCode)
	}
	res.Body.Close()
}

func TestCreateSession_TimedAmericano_IntervalTooHigh(t *testing.T) {
	srv, _ := newAPITestServer(t)
	res := postReq(t, srv, "/api/sessions", map[string]any{
		"courts":                      1,
		"game_mode":                   "timed_americano",
		"total_duration_minutes":      120,
		"buffer_seconds":              120,
		"interval_between_rounds_minutes": 6,
	}, "")
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for interval > 5, got %d", res.StatusCode)
	}
	res.Body.Close()
}
