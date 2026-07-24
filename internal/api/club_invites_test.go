package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// userID reads the authenticated user's id via /api/auth/me.
func userID(t *testing.T, srv *httptest.Server, token string) string {
	t.Helper()
	res := getReq(t, srv, "/api/auth/me", token)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("userID: expected 200, got %d", res.StatusCode)
	}
	var body struct {
		ID string `json:"id"`
	}
	decodeBody(t, res, &body)
	if body.ID == "" {
		t.Fatalf("userID: empty id")
	}
	return body.ID
}

// TestClubInviteFlow covers the acceptance path: a Member invites a User, the
// invitee sees it in their pending list, accepts, and lands on the roster.
func TestClubInviteFlow(t *testing.T) {
	srv, s := newAPITestServer(t)
	defer func() { _ = s.Close() }()

	aliceToken := mustRegister(t, srv, "alice@test.local", "Alice", "password123")
	bobToken := mustRegister(t, srv, "bob@test.local", "Bob", "password123")
	bobID := userID(t, srv, bobToken)

	clubID := mustCreateClub(t, srv, aliceToken, "Bouvet Padel")

	// Alice invites Bob.
	res := postReq(t, srv, "/api/clubs/"+clubID+"/invites", map[string]any{"to_user_id": bobID}, aliceToken)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("invite: expected 201, got %d", res.StatusCode)
	}

	// Bob sees the invite in his pending list.
	res = getReq(t, srv, "/api/clubs/invites", bobToken)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("list invites: expected 200, got %d", res.StatusCode)
	}
	var invites []struct {
		ID        string `json:"id"`
		ClubID    string `json:"club_id"`
		ClubName  string `json:"club_name"`
		InviterID string `json:"inviter_id"`
	}
	decodeBody(t, res, &invites)
	if len(invites) != 1 {
		t.Fatalf("expected 1 pending invite, got %d", len(invites))
	}
	if invites[0].ClubID != clubID || invites[0].ClubName != "Bouvet Padel" {
		t.Errorf("unexpected invite: %+v", invites[0])
	}

	// Bob accepts.
	res = putReq(t, srv, "/api/clubs/invites/"+invites[0].ID+"/accept", nil, bobToken)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("accept: expected 200, got %d", res.StatusCode)
	}

	// Bob is now on the roster.
	res = getReq(t, srv, "/api/clubs/"+clubID, bobToken)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("detail after accept: expected 200, got %d", res.StatusCode)
	}
	var detail struct {
		RosterCount int    `json:"roster_count"`
		MyRole      string `json:"my_role"`
	}
	decodeBody(t, res, &detail)
	if detail.RosterCount != 2 {
		t.Errorf("expected roster 2 after accept, got %d", detail.RosterCount)
	}
	if detail.MyRole != "member" {
		t.Errorf("expected role member, got %q", detail.MyRole)
	}

	// The pending list is now empty.
	res = getReq(t, srv, "/api/clubs/invites", bobToken)
	decodeBody(t, res, &invites)
	if len(invites) != 0 {
		t.Errorf("expected 0 pending after accept, got %d", len(invites))
	}
}

// TestClubInviteDecline confirms a declined invite does not add the invitee.
func TestClubInviteDecline(t *testing.T) {
	srv, s := newAPITestServer(t)
	defer func() { _ = s.Close() }()

	aliceToken := mustRegister(t, srv, "alice@test.local", "Alice", "password123")
	bobToken := mustRegister(t, srv, "bob@test.local", "Bob", "password123")
	bobID := userID(t, srv, bobToken)

	clubID := mustCreateClub(t, srv, aliceToken, "Bouvet Padel")

	res := postReq(t, srv, "/api/clubs/"+clubID+"/invites", map[string]any{"to_user_id": bobID}, aliceToken)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("invite: expected 201, got %d", res.StatusCode)
	}
	var inv struct {
		ID string `json:"id"`
	}
	decodeBody(t, res, &inv)

	res = putReq(t, srv, "/api/clubs/invites/"+inv.ID+"/decline", nil, bobToken)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("decline: expected 200, got %d", res.StatusCode)
	}

	// Bob is not a member and cannot view the club.
	res = getReq(t, srv, "/api/clubs/"+clubID, bobToken)
	if res.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403 for non-member after decline, got %d", res.StatusCode)
	}

	// Pending list is empty.
	res = getReq(t, srv, "/api/clubs/invites", bobToken)
	var invites []struct{}
	decodeBody(t, res, &invites)
	if len(invites) != 0 {
		t.Errorf("expected 0 pending after decline, got %d", len(invites))
	}
}

// TestClubInvite_AnyMemberCanSend verifies membership (not admin) is the gate.
func TestClubInvite_AnyMemberCanSend(t *testing.T) {
	srv, s := newAPITestServer(t)
	defer func() { _ = s.Close() }()

	aliceToken := mustRegister(t, srv, "alice@test.local", "Alice", "password123")
	bobToken := mustRegister(t, srv, "bob@test.local", "Bob", "password123")
	carolToken := mustRegister(t, srv, "carol@test.local", "Carol", "password123")
	bobID := userID(t, srv, bobToken)
	carolID := userID(t, srv, carolToken)

	clubID := mustCreateClub(t, srv, aliceToken, "Bouvet Padel")

	// Bob joins as a plain member via an invite from admin Alice.
	res := postReq(t, srv, "/api/clubs/"+clubID+"/invites", map[string]any{"to_user_id": bobID}, aliceToken)
	var inv struct {
		ID string `json:"id"`
	}
	decodeBody(t, res, &inv)
	res = putReq(t, srv, "/api/clubs/invites/"+inv.ID+"/accept", nil, bobToken)
	_ = res.Body.Close()

	// Plain member Bob invites Carol — allowed.
	res = postReq(t, srv, "/api/clubs/"+clubID+"/invites", map[string]any{"to_user_id": carolID}, bobToken)
	if res.StatusCode != http.StatusCreated {
		t.Errorf("member invite: expected 201, got %d", res.StatusCode)
	}
}

// TestClubInvite_NonMemberCannotSend rejects an invite from a non-member.
func TestClubInvite_NonMemberCannotSend(t *testing.T) {
	srv, s := newAPITestServer(t)
	defer func() { _ = s.Close() }()

	aliceToken := mustRegister(t, srv, "alice@test.local", "Alice", "password123")
	strangerToken := mustRegister(t, srv, "stranger@test.local", "Stranger", "password123")
	bobToken := mustRegister(t, srv, "bob@test.local", "Bob", "password123")
	bobID := userID(t, srv, bobToken)

	clubID := mustCreateClub(t, srv, aliceToken, "Bouvet Padel")

	res := postReq(t, srv, "/api/clubs/"+clubID+"/invites", map[string]any{"to_user_id": bobID}, strangerToken)
	if res.StatusCode != http.StatusForbidden {
		t.Errorf("non-member invite: expected 403, got %d", res.StatusCode)
	}
	var errBody struct {
		Error string `json:"error"`
	}
	decodeBody(t, res, &errBody)
	if errBody.Error != "not_club_member" {
		t.Errorf("expected not_club_member, got %q", errBody.Error)
	}
}

// TestClubInvite_Duplicate rejects a second pending invite to the same User.
func TestClubInvite_Duplicate(t *testing.T) {
	srv, s := newAPITestServer(t)
	defer func() { _ = s.Close() }()

	aliceToken := mustRegister(t, srv, "alice@test.local", "Alice", "password123")
	bobToken := mustRegister(t, srv, "bob@test.local", "Bob", "password123")
	bobID := userID(t, srv, bobToken)

	clubID := mustCreateClub(t, srv, aliceToken, "Bouvet Padel")

	res := postReq(t, srv, "/api/clubs/"+clubID+"/invites", map[string]any{"to_user_id": bobID}, aliceToken)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("first invite: expected 201, got %d", res.StatusCode)
	}
	res = postReq(t, srv, "/api/clubs/"+clubID+"/invites", map[string]any{"to_user_id": bobID}, aliceToken)
	if res.StatusCode != http.StatusConflict {
		t.Errorf("duplicate invite: expected 409, got %d", res.StatusCode)
	}
	var errBody struct {
		Error string `json:"error"`
	}
	decodeBody(t, res, &errBody)
	if errBody.Error != "already_club_invited" {
		t.Errorf("expected already_club_invited, got %q", errBody.Error)
	}
}

// TestClubInvite_AlreadyMember rejects inviting an existing Member.
func TestClubInvite_AlreadyMember(t *testing.T) {
	srv, s := newAPITestServer(t)
	defer func() { _ = s.Close() }()

	aliceToken := mustRegister(t, srv, "alice@test.local", "Alice", "password123")
	aliceID := userID(t, srv, aliceToken)

	clubID := mustCreateClub(t, srv, aliceToken, "Bouvet Padel")

	res := postReq(t, srv, "/api/clubs/"+clubID+"/invites", map[string]any{"to_user_id": aliceID}, aliceToken)
	if res.StatusCode != http.StatusConflict {
		t.Errorf("invite existing member: expected 409, got %d", res.StatusCode)
	}
	var errBody struct {
		Error string `json:"error"`
	}
	decodeBody(t, res, &errBody)
	if errBody.Error != "already_club_member" {
		t.Errorf("expected already_club_member, got %q", errBody.Error)
	}
}

// TestClubInvite_UnknownUser 404s when the invitee doesn't exist.
func TestClubInvite_UnknownUser(t *testing.T) {
	srv, s := newAPITestServer(t)
	defer func() { _ = s.Close() }()

	aliceToken := mustRegister(t, srv, "alice@test.local", "Alice", "password123")
	clubID := mustCreateClub(t, srv, aliceToken, "Bouvet Padel")

	res := postReq(t, srv, "/api/clubs/"+clubID+"/invites", map[string]any{"to_user_id": "nope"}, aliceToken)
	if res.StatusCode != http.StatusNotFound {
		t.Errorf("unknown invitee: expected 404, got %d", res.StatusCode)
	}
}

// TestClubInvite_OnlyInviteeCanAct rejects a third party accepting/declining.
func TestClubInvite_OnlyInviteeCanAct(t *testing.T) {
	srv, s := newAPITestServer(t)
	defer func() { _ = s.Close() }()

	aliceToken := mustRegister(t, srv, "alice@test.local", "Alice", "password123")
	bobToken := mustRegister(t, srv, "bob@test.local", "Bob", "password123")
	carolToken := mustRegister(t, srv, "carol@test.local", "Carol", "password123")
	bobID := userID(t, srv, bobToken)

	clubID := mustCreateClub(t, srv, aliceToken, "Bouvet Padel")

	res := postReq(t, srv, "/api/clubs/"+clubID+"/invites", map[string]any{"to_user_id": bobID}, aliceToken)
	var inv struct {
		ID string `json:"id"`
	}
	decodeBody(t, res, &inv)

	// Carol (not the invitee) cannot accept Bob's invite.
	res = putReq(t, srv, "/api/clubs/invites/"+inv.ID+"/accept", nil, carolToken)
	if res.StatusCode != http.StatusNotFound {
		t.Errorf("wrong-user accept: expected 404, got %d", res.StatusCode)
	}
}

// TestClubInvite_RequiresAuth rejects unauthenticated listing.
func TestClubInvite_RequiresAuth(t *testing.T) {
	srv, s := newAPITestServer(t)
	defer func() { _ = s.Close() }()

	res := getReq(t, srv, "/api/clubs/invites", "")
	if res.StatusCode != http.StatusUnauthorized {
		t.Errorf("list without auth: expected 401, got %d", res.StatusCode)
	}
}
