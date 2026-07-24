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

// clubJoinCode reads a club's current join_code via the owner's club detail.
func clubJoinCode(t *testing.T, srv *httptest.Server, token, clubID string) string {
	t.Helper()
	res := getReq(t, srv, "/api/clubs/"+clubID, token)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("clubJoinCode: expected 200, got %d", res.StatusCode)
	}
	var body struct {
		Club struct {
			JoinCode string `json:"join_code"`
		} `json:"club"`
	}
	decodeBody(t, res, &body)
	if body.Club.JoinCode == "" {
		t.Fatalf("clubJoinCode: empty join_code")
	}
	return body.Club.JoinCode
}

func TestClubJoinFlow(t *testing.T) {
	srv, s := newAPITestServer(t)
	defer func() { _ = s.Close() }()

	adminToken := mustRegister(t, srv, "alice@test.local", "Alice", "password123")
	joinerToken := mustRegister(t, srv, "bob@test.local", "Bob", "password123")

	clubID := mustCreateClub(t, srv, adminToken, "Bouvet Padel")
	joinCode := clubJoinCode(t, srv, adminToken, clubID)

	// Preview requires no auth and exposes name + member count, but never the code.
	res := getReq(t, srv, "/api/clubs/join/"+joinCode, "")
	if res.StatusCode != http.StatusOK {
		t.Fatalf("preview: expected 200, got %d", res.StatusCode)
	}
	var preview struct {
		Name        string `json:"name"`
		MemberCount int    `json:"member_count"`
		JoinCode    string `json:"join_code"`
	}
	decodeBody(t, res, &preview)
	if preview.Name != "Bouvet Padel" {
		t.Errorf("preview: expected name 'Bouvet Padel', got %q", preview.Name)
	}
	if preview.MemberCount != 1 {
		t.Errorf("preview: expected member_count 1, got %d", preview.MemberCount)
	}
	if preview.JoinCode != "" {
		t.Errorf("preview: join_code should not be echoed, got %q", preview.JoinCode)
	}

	// Bob joins via the code.
	res = postReq(t, srv, "/api/clubs/join", map[string]any{"join_code": joinCode}, joinerToken)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("join: expected 200, got %d", res.StatusCode)
	}
	var joinBody struct {
		ID string `json:"id"`
	}
	decodeBody(t, res, &joinBody)
	if joinBody.ID != clubID {
		t.Errorf("join: expected club id %s, got %s", clubID, joinBody.ID)
	}

	// The new member now appears in the roster and can view the club.
	res = getReq(t, srv, "/api/clubs/"+clubID, joinerToken)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("detail after join: expected 200, got %d", res.StatusCode)
	}
	var detail struct {
		Members     []struct{} `json:"members"`
		MyRole      string     `json:"my_role"`
		RosterCount int        `json:"roster_count"`
	}
	decodeBody(t, res, &detail)
	if detail.RosterCount != 2 {
		t.Errorf("expected roster count 2 after join, got %d", detail.RosterCount)
	}
	if detail.MyRole != "member" {
		t.Errorf("expected joiner role 'member', got %q", detail.MyRole)
	}
}

func TestClubJoin_Idempotent(t *testing.T) {
	srv, s := newAPITestServer(t)
	defer func() { _ = s.Close() }()

	adminToken := mustRegister(t, srv, "alice@test.local", "Alice", "password123")
	joinerToken := mustRegister(t, srv, "bob@test.local", "Bob", "password123")

	clubID := mustCreateClub(t, srv, adminToken, "Bouvet Padel")
	joinCode := clubJoinCode(t, srv, adminToken, clubID)

	for i := 0; i < 2; i++ {
		res := postReq(t, srv, "/api/clubs/join", map[string]any{"join_code": joinCode}, joinerToken)
		if res.StatusCode != http.StatusOK {
			t.Fatalf("join #%d: expected 200, got %d", i+1, res.StatusCode)
		}
		_ = res.Body.Close()
	}

	// Joining twice must not create a duplicate roster row.
	res := getReq(t, srv, "/api/clubs/"+clubID, adminToken)
	var detail struct {
		RosterCount int `json:"roster_count"`
	}
	decodeBody(t, res, &detail)
	if detail.RosterCount != 2 {
		t.Errorf("expected roster count 2 after double-join, got %d", detail.RosterCount)
	}
}

func TestClubJoin_UnknownCode(t *testing.T) {
	srv, s := newAPITestServer(t)
	defer func() { _ = s.Close() }()

	token := mustRegister(t, srv, "alice@test.local", "Alice", "password123")

	// Preview of an unknown code 404s (no auth).
	res := getReq(t, srv, "/api/clubs/join/nope-not-a-real-code", "")
	if res.StatusCode != http.StatusNotFound {
		t.Errorf("preview unknown: expected 404, got %d", res.StatusCode)
	}

	// Joining with an unknown code 404s.
	res = postReq(t, srv, "/api/clubs/join", map[string]any{"join_code": "nope-not-a-real-code"}, token)
	if res.StatusCode != http.StatusNotFound {
		t.Errorf("join unknown: expected 404, got %d", res.StatusCode)
	}
}

func TestClubJoin_RequiresAuth(t *testing.T) {
	srv, s := newAPITestServer(t)
	defer func() { _ = s.Close() }()

	adminToken := mustRegister(t, srv, "alice@test.local", "Alice", "password123")
	clubID := mustCreateClub(t, srv, adminToken, "Bouvet Padel")
	joinCode := clubJoinCode(t, srv, adminToken, clubID)

	res := postReq(t, srv, "/api/clubs/join", map[string]any{"join_code": joinCode}, "")
	if res.StatusCode != http.StatusUnauthorized {
		t.Errorf("join without auth: expected 401, got %d", res.StatusCode)
	}
}

func TestClubJoinCodeRotate(t *testing.T) {
	srv, s := newAPITestServer(t)
	defer func() { _ = s.Close() }()

	adminToken := mustRegister(t, srv, "alice@test.local", "Alice", "password123")
	joinerToken := mustRegister(t, srv, "bob@test.local", "Bob", "password123")

	clubID := mustCreateClub(t, srv, adminToken, "Bouvet Padel")
	oldCode := clubJoinCode(t, srv, adminToken, clubID)

	// Admin rotates the join code.
	res := postReq(t, srv, "/api/clubs/"+clubID+"/join-code/rotate", nil, adminToken)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("rotate: expected 200, got %d", res.StatusCode)
	}
	var body struct {
		JoinCode string `json:"join_code"`
	}
	decodeBody(t, res, &body)
	if body.JoinCode == "" || body.JoinCode == oldCode {
		t.Fatalf("rotate: expected a new non-empty join_code, got %q (old %q)", body.JoinCode, oldCode)
	}

	// The old code no longer previews or joins.
	res = getReq(t, srv, "/api/clubs/join/"+oldCode, "")
	if res.StatusCode != http.StatusNotFound {
		t.Errorf("preview old code: expected 404, got %d", res.StatusCode)
	}
	res = postReq(t, srv, "/api/clubs/join", map[string]any{"join_code": oldCode}, joinerToken)
	if res.StatusCode != http.StatusNotFound {
		t.Errorf("join old code: expected 404, got %d", res.StatusCode)
	}

	// The new code still works.
	res = postReq(t, srv, "/api/clubs/join", map[string]any{"join_code": body.JoinCode}, joinerToken)
	if res.StatusCode != http.StatusOK {
		t.Errorf("join new code: expected 200, got %d", res.StatusCode)
	}
}

func TestClubJoinCodeRotate_NonAdmin(t *testing.T) {
	srv, s := newAPITestServer(t)
	defer func() { _ = s.Close() }()

	adminToken := mustRegister(t, srv, "alice@test.local", "Alice", "password123")
	joinerToken := mustRegister(t, srv, "bob@test.local", "Bob", "password123")

	clubID := mustCreateClub(t, srv, adminToken, "Bouvet Padel")
	joinCode := clubJoinCode(t, srv, adminToken, clubID)

	// A plain member cannot rotate the code.
	postJoin := postReq(t, srv, "/api/clubs/join", map[string]any{"join_code": joinCode}, joinerToken)
	_ = postJoin.Body.Close()

	res := postReq(t, srv, "/api/clubs/"+clubID+"/join-code/rotate", nil, joinerToken)
	if res.StatusCode != http.StatusForbidden {
		t.Errorf("member rotate: expected 403, got %d", res.StatusCode)
	}
	var errBody struct {
		Error string `json:"error"`
	}
	decodeBody(t, res, &errBody)
	if errBody.Error != "admin_required" {
		t.Errorf("expected admin_required, got %q", errBody.Error)
	}

	// A non-member cannot rotate either.
	strangerToken := mustRegister(t, srv, "carol@test.local", "Carol", "password123")
	res = postReq(t, srv, "/api/clubs/"+clubID+"/join-code/rotate", nil, strangerToken)
	if res.StatusCode != http.StatusForbidden {
		t.Errorf("non-member rotate: expected 403, got %d", res.StatusCode)
	}
	decodeBody(t, res, &errBody)
	if errBody.Error != "not_club_member" {
		t.Errorf("expected not_club_member, got %q", errBody.Error)
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
