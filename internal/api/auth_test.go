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
	}, "")
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", res.StatusCode)
	}
	var body struct {
		Token string `json:"token"`
		User  struct {
			Email       string `json:"email"`
			DisplayName string `json:"display_name"`
		} `json:"user"`
	}
	decodeBody(t, res, &body)
	if body.Token == "" {
		t.Error("expected non-empty token")
	}
	if body.User.Email != "alice@example.com" {
		t.Errorf("expected email alice@example.com, got %q", body.User.Email)
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
