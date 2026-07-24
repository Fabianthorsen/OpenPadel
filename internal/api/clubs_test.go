package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func mustCreateClub(t *testing.T, srv *httptest.Server, token string, name string) string {
	t.Helper()
	res := postReq(t, srv, "/api/clubs", map[string]any{
		"name":         name,
		"description":  "Test club",
		"avatar_icon":  "Star",
		"avatar_color": "forest",
	}, token)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("createClub: expected 201, got %d", res.StatusCode)
	}
	var body struct {
		ID string `json:"id"`
	}
	decodeBody(t, res, &body)
	return body.ID
}

func TestCreateClub(t *testing.T) {
	srv, s := newAPITestServer(t)
	defer func() { _ = s.Close() }()

	// Register a user
	token := mustRegister(t, srv, "alice@test.local", "Alice", "password123")

	// Create a club
	res := postReq(t, srv, "/api/clubs", map[string]any{
		"name":         "Bouvet Padel",
		"description":  "Workplace club",
		"avatar_icon":  "Star",
		"avatar_color": "forest",
	}, token)

	if res.StatusCode != http.StatusCreated {
		t.Errorf("expected 201, got %d", res.StatusCode)
	}

	var body struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		JoinCode  string `json:"join_code"`
		CreatedBy string `json:"created_by"`
	}
	decodeBody(t, res, &body)

	if body.Name != "Bouvet Padel" {
		t.Errorf("expected name 'Bouvet Padel', got %q", body.Name)
	}
	if body.JoinCode == "" {
		t.Errorf("expected non-empty join_code")
	}
	if body.CreatedBy != "" {
		// CreatedBy in Club response should be the user ID
		t.Logf("created_by: %s", body.CreatedBy)
	}
}

func TestGetMyClubs(t *testing.T) {
	srv, s := newAPITestServer(t)
	defer func() { _ = s.Close() }()

	// Register two users
	token1 := mustRegister(t, srv, "alice@test.local", "Alice", "password123")
	token2 := mustRegister(t, srv, "bob@test.local", "Bob", "password123")

	// Create a club as Alice
	clubID := mustCreateClub(t, srv, token1, "Club A")

	// Get my clubs as Alice
	res := getReq(t, srv, "/api/clubs", token1)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}

	var clubs []struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		MyRole      string `json:"my_role"`
		RosterCount int    `json:"roster_count"`
	}
	decodeBody(t, res, &clubs)

	if len(clubs) != 1 {
		t.Errorf("expected 1 club, got %d", len(clubs))
	}
	if clubs[0].ID != clubID {
		t.Errorf("expected club ID %s, got %s", clubID, clubs[0].ID)
	}
	if clubs[0].MyRole != "admin" {
		t.Errorf("expected my_role 'admin', got %q", clubs[0].MyRole)
	}
	if clubs[0].RosterCount != 1 {
		t.Errorf("expected roster count 1, got %d", clubs[0].RosterCount)
	}

	// Get my clubs as Bob (should be empty)
	res = getReq(t, srv, "/api/clubs", token2)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}

	var bobClubs []struct{}
	decodeBody(t, res, &bobClubs)

	if len(bobClubs) != 0 {
		t.Errorf("expected 0 clubs for Bob, got %d", len(bobClubs))
	}
}

func TestGetClubDetail(t *testing.T) {
	srv, s := newAPITestServer(t)
	defer func() { _ = s.Close() }()

	// Register a user and create a club
	token := mustRegister(t, srv, "alice@test.local", "Alice", "password123")
	clubID := mustCreateClub(t, srv, token, "Test Club")

	// Get club detail
	res := getReq(t, srv, "/api/clubs/"+clubID, token)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}

	var body struct {
		Club struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"club"`
		Members     []any  `json:"members"`
		IsAdmin     bool   `json:"is_admin"`
		MyRole      string `json:"my_role"`
		RosterCount int    `json:"roster_count"`
	}
	decodeBody(t, res, &body)

	if body.Club.ID != clubID {
		t.Errorf("expected club ID %s, got %s", clubID, body.Club.ID)
	}
	if !body.IsAdmin {
		t.Errorf("expected is_admin=true")
	}
	if body.MyRole != "admin" {
		t.Errorf("expected my_role 'admin', got %q", body.MyRole)
	}
	if body.RosterCount != 1 {
		t.Errorf("expected 1 member, got %d", body.RosterCount)
	}
}

func TestGetClubDetail_NotMember(t *testing.T) {
	srv, s := newAPITestServer(t)
	defer func() { _ = s.Close() }()

	// Register two users
	token1 := mustRegister(t, srv, "alice@test.local", "Alice", "password123")
	token2 := mustRegister(t, srv, "bob@test.local", "Bob", "password123")

	// Create a club as Alice
	clubID := mustCreateClub(t, srv, token1, "Club A")

	// Bob tries to view the club (non-member) — should be refused
	res := getReq(t, srv, "/api/clubs/"+clubID, token2)
	if res.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403 for non-member, got %d", res.StatusCode)
	}
	var body struct {
		Error string `json:"error"`
	}
	decodeBody(t, res, &body)
	if body.Error != "not_club_member" {
		t.Errorf("expected not_club_member error, got %q", body.Error)
	}
}
