package api_test

import (
	"net/http"
	"testing"
)

func TestRegister(t *testing.T) {
	srv, _ := newAPITestServer(t)
	res := postReq(t, srv, "/api/auth/register", map[string]any{
		"email":        "alice@example.com",
		"display_name": "Alice",
		"password":     "password123",
		"self_rating":  5,
	}, "")
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", res.StatusCode)
	}
	var body struct {
		Token string `json:"token"`
		User  struct {
			Email       string `json:"email"`
			DisplayName string `json:"display_name"`
			SelfRating  *int   `json:"self_rating"`
		} `json:"user"`
	}
	decodeBody(t, res, &body)
	if body.Token == "" {
		t.Error("expected non-empty token")
	}
	if body.User.Email != "alice@example.com" {
		t.Errorf("expected email alice@example.com, got %q", body.User.Email)
	}
	if body.User.SelfRating == nil || *body.User.SelfRating != 5 {
		t.Errorf("expected self_rating 5, got %v", body.User.SelfRating)
	}
}

func TestRegister_MissingRating(t *testing.T) {
	srv, _ := newAPITestServer(t)
	res := postReq(t, srv, "/api/auth/register", map[string]any{
		"email":        "alice@example.com",
		"display_name": "Alice",
		"password":     "password123",
	}, "")
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing self_rating, got %d", res.StatusCode)
	}
	_ = res.Body.Close()
}

func TestRegister_OutOfRangeRating(t *testing.T) {
	srv, _ := newAPITestServer(t)
	for _, rating := range []int{0, 6, -1} {
		res := postReq(t, srv, "/api/auth/register", map[string]any{
			"email":        "alice@example.com",
			"display_name": "Alice",
			"password":     "password123",
			"self_rating":  rating,
		}, "")
		if res.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected 400 for self_rating %d, got %d", rating, res.StatusCode)
		}
		_ = res.Body.Close()
	}
}

func TestRegister_MissingFields(t *testing.T) {
	srv, _ := newAPITestServer(t)
	res := postReq(t, srv, "/api/auth/register", map[string]any{
		"email": "alice@example.com",
	}, "")
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.StatusCode)
	}
	res.Body.Close()
}

func TestRegister_ShortPassword(t *testing.T) {
	srv, _ := newAPITestServer(t)
	res := postReq(t, srv, "/api/auth/register", map[string]any{
		"email":        "alice@example.com",
		"display_name": "Alice",
		"password":     "short",
	}, "")
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.StatusCode)
	}
	res.Body.Close()
}

func TestRegister_DuplicateEmail(t *testing.T) {
	srv, _ := newAPITestServer(t)
	body := map[string]any{
		"email":        "alice@example.com",
		"display_name": "Alice",
		"password":     "password123",
		"self_rating":  3,
	}
	postReq(t, srv, "/api/auth/register", body, "").Body.Close()
	res := postReq(t, srv, "/api/auth/register", body, "")
	if res.StatusCode != http.StatusConflict {
		t.Fatalf("expected 409, got %d", res.StatusCode)
	}
	res.Body.Close()
}

func TestLogin(t *testing.T) {
	srv, _ := newAPITestServer(t)
	mustRegister(t, srv, "alice@example.com", "Alice", "password123")

	res := postReq(t, srv, "/api/auth/login", map[string]any{
		"email":    "alice@example.com",
		"password": "password123",
	}, "")
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
	var body struct {
		Token string `json:"token"`
	}
	decodeBody(t, res, &body)
	if body.Token == "" {
		t.Error("expected token in response")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	srv, _ := newAPITestServer(t)
	mustRegister(t, srv, "alice@example.com", "Alice", "password123")

	res := postReq(t, srv, "/api/auth/login", map[string]any{
		"email":    "alice@example.com",
		"password": "wrongpassword",
	}, "")
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", res.StatusCode)
	}
	res.Body.Close()
}

func TestLogin_UnknownEmail(t *testing.T) {
	srv, _ := newAPITestServer(t)
	res := postReq(t, srv, "/api/auth/login", map[string]any{
		"email":    "nobody@example.com",
		"password": "password123",
	}, "")
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", res.StatusCode)
	}
	res.Body.Close()
}

func TestMe(t *testing.T) {
	srv, _ := newAPITestServer(t)
	token := mustRegister(t, srv, "alice@example.com", "Alice", "password123")

	res := getReq(t, srv, "/api/auth/me", token)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
	var user struct {
		Email string `json:"email"`
	}
	decodeBody(t, res, &user)
	if user.Email != "alice@example.com" {
		t.Errorf("expected alice@example.com, got %q", user.Email)
	}
}

func TestMe_Unauthenticated(t *testing.T) {
	srv, _ := newAPITestServer(t)
	res := getReq(t, srv, "/api/auth/me", "")
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", res.StatusCode)
	}
	res.Body.Close()
}

func TestMe_InvalidToken(t *testing.T) {
	srv, _ := newAPITestServer(t)
	res := getReq(t, srv, "/api/auth/me", "invalid-token-xyz")
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", res.StatusCode)
	}
	res.Body.Close()
}

func TestLogout(t *testing.T) {
	srv, _ := newAPITestServer(t)
	token := mustRegister(t, srv, "alice@example.com", "Alice", "password123")

	res := postReq(t, srv, "/api/auth/logout", nil, token)
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", res.StatusCode)
	}
	res.Body.Close()

	res2 := getReq(t, srv, "/api/auth/me", token)
	if res2.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 after logout, got %d", res2.StatusCode)
	}
	res2.Body.Close()
}

func TestUpdateProfile(t *testing.T) {
	srv, _ := newAPITestServer(t)
	token := mustRegister(t, srv, "alice@example.com", "Alice", "password123")

	res := putReq(t, srv, "/api/auth/profile", map[string]any{
		"display_name": "Alice Updated",
		"avatar_icon":  "Star",
		"avatar_color": "blue",
	}, token)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
	var user struct {
		DisplayName string `json:"display_name"`
	}
	decodeBody(t, res, &user)
	if user.DisplayName != "Alice Updated" {
		t.Errorf("expected 'Alice Updated', got %q", user.DisplayName)
	}
}

func TestUpdateProfile_MissingDisplayName(t *testing.T) {
	srv, _ := newAPITestServer(t)
	token := mustRegister(t, srv, "alice@example.com", "Alice", "password123")

	res := putReq(t, srv, "/api/auth/profile", map[string]any{
		"display_name": "",
	}, token)
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.StatusCode)
	}
	res.Body.Close()
}

func TestDeleteAccount(t *testing.T) {
	srv, _ := newAPITestServer(t)
	token := mustRegister(t, srv, "alice@example.com", "Alice", "password123")

	res := deleteReq(t, srv, "/api/auth/account", token)
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", res.StatusCode)
	}
	res.Body.Close()

	res2 := getReq(t, srv, "/api/auth/me", token)
	if res2.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 after account deletion, got %d", res2.StatusCode)
	}
	res2.Body.Close()
}

func TestGetUserSessions(t *testing.T) {
	srv, _ := newAPITestServer(t)
	alice := mustRegister(t, srv, "alice@example.com", "Alice", "password123")

	// Create two sessions as Alice
	sess1ID, sess1Token := mustCreateSession(t, srv, alice)
	sess2ID, sess2Token := mustCreateSession(t, srv, alice)

	// Get Alice's sessions
	res := getReq(t, srv, "/api/auth/sessions", alice)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}

	var body struct {
		Sessions []struct {
			ID         string `json:"id"`
			AdminToken string `json:"admin_token"`
		} `json:"sessions"`
	}
	decodeBody(t, res, &body)

	if len(body.Sessions) != 2 {
		t.Errorf("expected 2 sessions, got %d", len(body.Sessions))
	}

	// Verify admin tokens are included and match
	sessionTokens := map[string]string{
		sess1ID: sess1Token,
		sess2ID: sess2Token,
	}
	found := make(map[string]bool)
	for _, sess := range body.Sessions {
		if sess.AdminToken == "" {
			t.Errorf("expected non-empty AdminToken for session %s", sess.ID)
		}
		if expectedToken, ok := sessionTokens[sess.ID]; ok {
			if sess.AdminToken != expectedToken {
				t.Errorf("expected token %q for session %s, got %q", expectedToken, sess.ID, sess.AdminToken)
			}
			found[sess.ID] = true
		}
	}

	for sessionID := range sessionTokens {
		if !found[sessionID] {
			t.Errorf("expected session %s in response", sessionID)
		}
	}
}

func TestGetUserSessions_Unauthenticated(t *testing.T) {
	srv, _ := newAPITestServer(t)
	res := getReq(t, srv, "/api/auth/sessions", "")
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", res.StatusCode)
	}
	if err := res.Body.Close(); err != nil {
		t.Errorf("failed to close response body: %v", err)
	}
}
