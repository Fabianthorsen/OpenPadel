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

// mustJoinClub has Bob join the club via its current join code and returns his
// user id (read back from the roster).
func mustJoinClub(t *testing.T, srv *httptest.Server, adminToken, joinerToken, clubID string) string {
	t.Helper()
	joinCode := clubJoinCode(t, srv, adminToken, clubID)
	res := postReq(t, srv, "/api/clubs/join", map[string]any{"join_code": joinCode}, joinerToken)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("join club: expected 200, got %d", res.StatusCode)
	}
	_ = res.Body.Close()
	return memberIDByRole(t, srv, adminToken, clubID, "member")
}

// joinClub has a user join the club via its current join code without reading
// back an id — use when the roster has several members and a role lookup would be
// ambiguous.
func joinClub(t *testing.T, srv *httptest.Server, adminToken, joinerToken, clubID string) {
	t.Helper()
	joinCode := clubJoinCode(t, srv, adminToken, clubID)
	res := postReq(t, srv, "/api/clubs/join", map[string]any{"join_code": joinCode}, joinerToken)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("join club: expected 200, got %d", res.StatusCode)
	}
	_ = res.Body.Close()
}

// memberIDByName returns the user_id of the roster member with the given display name.
func memberIDByName(t *testing.T, srv *httptest.Server, token, clubID, name string) string {
	t.Helper()
	res := getReq(t, srv, "/api/clubs/"+clubID, token)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("memberIDByName: detail expected 200, got %d", res.StatusCode)
	}
	var detail struct {
		Members []struct {
			UserID      string `json:"user_id"`
			DisplayName string `json:"display_name"`
		} `json:"members"`
	}
	decodeBody(t, res, &detail)
	for _, m := range detail.Members {
		if m.DisplayName == name {
			return m.UserID
		}
	}
	t.Fatalf("memberIDByName: no member named %q", name)
	return ""
}

// memberIDByRole returns the user_id of the first roster member with the given role.
func memberIDByRole(t *testing.T, srv *httptest.Server, token, clubID, role string) string {
	t.Helper()
	res := getReq(t, srv, "/api/clubs/"+clubID, token)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("memberIDByRole: detail expected 200, got %d", res.StatusCode)
	}
	var detail struct {
		Members []struct {
			UserID string `json:"user_id"`
			Role   string `json:"role"`
		} `json:"members"`
	}
	decodeBody(t, res, &detail)
	for _, m := range detail.Members {
		if m.Role == role {
			return m.UserID
		}
	}
	t.Fatalf("memberIDByRole: no member with role %q", role)
	return ""
}

func TestUpdateClub(t *testing.T) {
	srv, s := newAPITestServer(t)
	defer func() { _ = s.Close() }()

	adminToken := mustRegister(t, srv, "alice@test.local", "Alice", "password123")
	clubID := mustCreateClub(t, srv, adminToken, "Old Name")

	res := patchReq(t, srv, "/api/clubs/"+clubID, map[string]any{
		"name":         "New Name",
		"description":  "Updated description",
		"avatar_icon":  "Trophy",
		"avatar_color": "ocean",
	}, adminToken)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("update: expected 200, got %d", res.StatusCode)
	}
	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		AvatarIcon  string `json:"avatar_icon"`
		AvatarColor string `json:"avatar_color"`
	}
	decodeBody(t, res, &body)
	if body.Name != "New Name" || body.Description != "Updated description" {
		t.Errorf("update: unexpected body %+v", body)
	}
	if body.AvatarIcon != "Trophy" || body.AvatarColor != "ocean" {
		t.Errorf("update: avatar not persisted %+v", body)
	}
}

func TestUpdateClub_NonAdmin(t *testing.T) {
	srv, s := newAPITestServer(t)
	defer func() { _ = s.Close() }()

	adminToken := mustRegister(t, srv, "alice@test.local", "Alice", "password123")
	memberToken := mustRegister(t, srv, "bob@test.local", "Bob", "password123")
	clubID := mustCreateClub(t, srv, adminToken, "Club A")
	mustJoinClub(t, srv, adminToken, memberToken, clubID)

	// A plain member cannot edit the club.
	res := patchReq(t, srv, "/api/clubs/"+clubID, map[string]any{"name": "Hijacked"}, memberToken)
	if res.StatusCode != http.StatusForbidden {
		t.Fatalf("member update: expected 403, got %d", res.StatusCode)
	}
	assertErrorCode(t, res, "admin_required")

	// A non-member is refused too.
	strangerToken := mustRegister(t, srv, "carol@test.local", "Carol", "password123")
	res = patchReq(t, srv, "/api/clubs/"+clubID, map[string]any{"name": "Hijacked"}, strangerToken)
	if res.StatusCode != http.StatusForbidden {
		t.Fatalf("stranger update: expected 403, got %d", res.StatusCode)
	}
	assertErrorCode(t, res, "not_club_member")
}

// TestDeleteClub_CascadesMembership covers the API delete path and the
// club_members cascade. The sessions.club_id SET NULL survival behaviour is
// exercised at the store layer in TestDeleteClub_UnlinksSessions, since no API
// path attaches a Session to a Club yet.
func TestDeleteClub_CascadesMembership(t *testing.T) {
	srv, s := newAPITestServer(t)
	defer func() { _ = s.Close() }()

	adminToken := mustRegister(t, srv, "alice@test.local", "Alice", "password123")
	memberToken := mustRegister(t, srv, "bob@test.local", "Bob", "password123")
	clubID := mustCreateClub(t, srv, adminToken, "Doomed Club")
	mustJoinClub(t, srv, adminToken, memberToken, clubID)

	res := deleteReq(t, srv, "/api/clubs/"+clubID, adminToken)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("delete: expected 200, got %d", res.StatusCode)
	}
	_ = res.Body.Close()

	// The club is gone.
	res = getReq(t, srv, "/api/clubs/"+clubID, adminToken)
	if res.StatusCode != http.StatusForbidden && res.StatusCode != http.StatusNotFound {
		t.Errorf("get after delete: expected 403/404, got %d", res.StatusCode)
	}
	_ = res.Body.Close()

	// Membership cascaded away — the former member has no clubs.
	res = getReq(t, srv, "/api/clubs", memberToken)
	var clubs []struct{}
	decodeBody(t, res, &clubs)
	if len(clubs) != 0 {
		t.Errorf("expected 0 clubs after delete, got %d", len(clubs))
	}
}

func TestDeleteClub_NonAdmin(t *testing.T) {
	srv, s := newAPITestServer(t)
	defer func() { _ = s.Close() }()

	adminToken := mustRegister(t, srv, "alice@test.local", "Alice", "password123")
	memberToken := mustRegister(t, srv, "bob@test.local", "Bob", "password123")
	clubID := mustCreateClub(t, srv, adminToken, "Club A")
	mustJoinClub(t, srv, adminToken, memberToken, clubID)

	res := deleteReq(t, srv, "/api/clubs/"+clubID, memberToken)
	if res.StatusCode != http.StatusForbidden {
		t.Fatalf("member delete: expected 403, got %d", res.StatusCode)
	}
	assertErrorCode(t, res, "admin_required")
}

func TestPromoteAndDemoteMember(t *testing.T) {
	srv, s := newAPITestServer(t)
	defer func() { _ = s.Close() }()

	adminToken := mustRegister(t, srv, "alice@test.local", "Alice", "password123")
	memberToken := mustRegister(t, srv, "bob@test.local", "Bob", "password123")
	clubID := mustCreateClub(t, srv, adminToken, "Club A")
	bobID := mustJoinClub(t, srv, adminToken, memberToken, clubID)

	// Promote Bob to admin.
	res := patchReq(t, srv, "/api/clubs/"+clubID+"/members/"+bobID, map[string]any{"role": "admin"}, adminToken)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("promote: expected 200, got %d", res.StatusCode)
	}
	_ = res.Body.Close()
	if got := memberRole(t, srv, adminToken, clubID, bobID); got != "admin" {
		t.Errorf("promote: expected role admin, got %q", got)
	}

	// Bob (now admin) can demote himself back to member since Alice remains admin.
	res = patchReq(t, srv, "/api/clubs/"+clubID+"/members/"+bobID, map[string]any{"role": "member"}, memberToken)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("demote: expected 200, got %d", res.StatusCode)
	}
	_ = res.Body.Close()
	if got := memberRole(t, srv, adminToken, clubID, bobID); got != "member" {
		t.Errorf("demote: expected role member, got %q", got)
	}
}

func TestPromoteDemote_NonAdmin(t *testing.T) {
	srv, s := newAPITestServer(t)
	defer func() { _ = s.Close() }()

	adminToken := mustRegister(t, srv, "alice@test.local", "Alice", "password123")
	memberToken := mustRegister(t, srv, "bob@test.local", "Bob", "password123")
	clubID := mustCreateClub(t, srv, adminToken, "Club A")
	bobID := mustJoinClub(t, srv, adminToken, memberToken, clubID)
	adminID := memberIDByRole(t, srv, adminToken, clubID, "admin")

	// A plain member cannot promote themselves.
	res := patchReq(t, srv, "/api/clubs/"+clubID+"/members/"+bobID, map[string]any{"role": "admin"}, memberToken)
	if res.StatusCode != http.StatusForbidden {
		t.Fatalf("member promote self: expected 403, got %d", res.StatusCode)
	}
	assertErrorCode(t, res, "admin_required")

	// Nor demote the admin.
	res = patchReq(t, srv, "/api/clubs/"+clubID+"/members/"+adminID, map[string]any{"role": "member"}, memberToken)
	if res.StatusCode != http.StatusForbidden {
		t.Fatalf("member demote admin: expected 403, got %d", res.StatusCode)
	}
	assertErrorCode(t, res, "admin_required")
}

func TestRemoveMember_AdminRemovesOther(t *testing.T) {
	srv, s := newAPITestServer(t)
	defer func() { _ = s.Close() }()

	adminToken := mustRegister(t, srv, "alice@test.local", "Alice", "password123")
	memberToken := mustRegister(t, srv, "bob@test.local", "Bob", "password123")
	clubID := mustCreateClub(t, srv, adminToken, "Club A")
	bobID := mustJoinClub(t, srv, adminToken, memberToken, clubID)

	res := deleteReq(t, srv, "/api/clubs/"+clubID+"/members/"+bobID, adminToken)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("admin remove member: expected 200, got %d", res.StatusCode)
	}
	_ = res.Body.Close()

	// Bob is no longer a member.
	res = getReq(t, srv, "/api/clubs/"+clubID, memberToken)
	if res.StatusCode != http.StatusForbidden {
		t.Errorf("removed member view: expected 403, got %d", res.StatusCode)
	}
	_ = res.Body.Close()
}

func TestRemoveMember_MemberLeaves(t *testing.T) {
	srv, s := newAPITestServer(t)
	defer func() { _ = s.Close() }()

	adminToken := mustRegister(t, srv, "alice@test.local", "Alice", "password123")
	memberToken := mustRegister(t, srv, "bob@test.local", "Bob", "password123")
	clubID := mustCreateClub(t, srv, adminToken, "Club A")
	bobID := mustJoinClub(t, srv, adminToken, memberToken, clubID)

	// Bob removes himself (leave).
	res := deleteReq(t, srv, "/api/clubs/"+clubID+"/members/"+bobID, memberToken)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("self-leave: expected 200, got %d", res.StatusCode)
	}
	_ = res.Body.Close()

	// A plain member cannot remove someone else. Carol and Dave both join; Carol
	// (a member) tries to remove Dave.
	carolToken := mustRegister(t, srv, "carol@test.local", "Carol", "password123")
	joinClub(t, srv, adminToken, carolToken, clubID)
	daveToken := mustRegister(t, srv, "dave@test.local", "Dave", "password123")
	joinClub(t, srv, adminToken, daveToken, clubID)
	daveID := memberIDByName(t, srv, adminToken, clubID, "Dave")

	res = deleteReq(t, srv, "/api/clubs/"+clubID+"/members/"+daveID, carolToken)
	if res.StatusCode != http.StatusForbidden {
		t.Fatalf("member removes other: expected 403, got %d", res.StatusCode)
	}
	assertErrorCode(t, res, "admin_required")
}

func TestSoleAdminBlocked(t *testing.T) {
	srv, s := newAPITestServer(t)
	defer func() { _ = s.Close() }()

	adminToken := mustRegister(t, srv, "alice@test.local", "Alice", "password123")
	memberToken := mustRegister(t, srv, "bob@test.local", "Bob", "password123")
	clubID := mustCreateClub(t, srv, adminToken, "Club A")
	mustJoinClub(t, srv, adminToken, memberToken, clubID)
	adminID := memberIDByRole(t, srv, adminToken, clubID, "admin")

	// The sole admin cannot self-demote.
	res := patchReq(t, srv, "/api/clubs/"+clubID+"/members/"+adminID, map[string]any{"role": "member"}, adminToken)
	if res.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("sole admin demote: expected 422, got %d", res.StatusCode)
	}
	assertErrorCode(t, res, "last_admin")

	// Nor leave.
	res = deleteReq(t, srv, "/api/clubs/"+clubID+"/members/"+adminID, adminToken)
	if res.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("sole admin leave: expected 422, got %d", res.StatusCode)
	}
	assertErrorCode(t, res, "last_admin")

	// After promoting Bob, Alice may leave.
	bobID := memberIDByRole(t, srv, adminToken, clubID, "member")
	res = patchReq(t, srv, "/api/clubs/"+clubID+"/members/"+bobID, map[string]any{"role": "admin"}, adminToken)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("promote bob: expected 200, got %d", res.StatusCode)
	}
	_ = res.Body.Close()
	res = deleteReq(t, srv, "/api/clubs/"+clubID+"/members/"+adminID, adminToken)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("admin leave after promote: expected 200, got %d", res.StatusCode)
	}
	_ = res.Body.Close()
}

// memberRole reads a specific member's role from the roster.
func memberRole(t *testing.T, srv *httptest.Server, token, clubID, userID string) string {
	t.Helper()
	res := getReq(t, srv, "/api/clubs/"+clubID, token)
	var detail struct {
		Members []struct {
			UserID string `json:"user_id"`
			Role   string `json:"role"`
		} `json:"members"`
	}
	decodeBody(t, res, &detail)
	for _, m := range detail.Members {
		if m.UserID == userID {
			return m.Role
		}
	}
	t.Fatalf("memberRole: user %q not on roster", userID)
	return ""
}

// assertErrorCode decodes the response and asserts its error code.
func assertErrorCode(t *testing.T, res *http.Response, want string) {
	t.Helper()
	var body struct {
		Error string `json:"error"`
	}
	decodeBody(t, res, &body)
	if body.Error != want {
		t.Errorf("expected error %q, got %q", want, body.Error)
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
