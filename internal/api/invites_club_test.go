package api_test

import (
	"net/http"
	"testing"
)

// TestSessionClubInvite_FansOutRoster covers the happy path: a Club Member fans
// the whole roster out into one Session invite each, and accepting one behaves
// exactly like any other Session invite (the invitee becomes a Player).
func TestSessionClubInvite_FansOutRoster(t *testing.T) {
	srv, s := newAPITestServer(t)
	defer func() { _ = s.Close() }()

	aliceToken := mustRegister(t, srv, "alice@test.local", "Alice", "password123")
	bobToken := mustRegister(t, srv, "bob@test.local", "Bob", "password123")
	carolToken := mustRegister(t, srv, "carol@test.local", "Carol", "password123")

	clubID := mustCreateClub(t, srv, aliceToken, "Bouvet Padel")
	joinClub(t, srv, aliceToken, bobToken, clubID)
	joinClub(t, srv, aliceToken, carolToken, clubID)

	sessionID, _ := mustCreateSession(t, srv, aliceToken)

	res := postReq(t, srv, "/api/sessions/"+sessionID+"/invites/club",
		map[string]any{"club_id": clubID}, aliceToken)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("fan-out: expected 201, got %d", res.StatusCode)
	}
	var created []struct {
		ID       string `json:"id"`
		ToUserID string `json:"to_user_id"`
	}
	decodeBody(t, res, &created)
	// Alice (the caller) is not invited to her own fan-out; Bob and Carol are.
	if len(created) != 2 {
		t.Fatalf("expected 2 invites over the roster, got %d", len(created))
	}

	// Bob sees his invite and accepts it — a plain Session invite, nothing special.
	res = getReq(t, srv, "/api/invites", bobToken)
	var pending []struct {
		ID        string `json:"id"`
		SessionID string `json:"session_id"`
	}
	decodeBody(t, res, &pending)
	if len(pending) != 1 || pending[0].SessionID != sessionID {
		t.Fatalf("expected Bob to hold 1 invite for the session, got %+v", pending)
	}

	res = postReq(t, srv, "/api/invites/"+pending[0].ID+"/accept", nil, bobToken)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("accept: expected 200, got %d", res.StatusCode)
	}
	var player struct {
		SessionID string `json:"session_id"`
	}
	decodeBody(t, res, &player)
	if player.SessionID != sessionID {
		t.Errorf("expected accepted player on session %s, got %s", sessionID, player.SessionID)
	}
}

// TestSessionClubInvite_Idempotent confirms a repeat fan-out and an
// already-joined Member cause neither a duplicate nor an error.
func TestSessionClubInvite_Idempotent(t *testing.T) {
	srv, s := newAPITestServer(t)
	defer func() { _ = s.Close() }()

	aliceToken := mustRegister(t, srv, "alice@test.local", "Alice", "password123")
	bobToken := mustRegister(t, srv, "bob@test.local", "Bob", "password123")
	carolToken := mustRegister(t, srv, "carol@test.local", "Carol", "password123")

	clubID := mustCreateClub(t, srv, aliceToken, "Bouvet Padel")
	joinClub(t, srv, aliceToken, bobToken, clubID)
	joinClub(t, srv, aliceToken, carolToken, clubID)

	sessionID, _ := mustCreateSession(t, srv, aliceToken)

	// First fan-out invites Bob and Carol.
	res := postReq(t, srv, "/api/sessions/"+sessionID+"/invites/club",
		map[string]any{"club_id": clubID}, aliceToken)
	var first []struct{}
	decodeBody(t, res, &first)
	if len(first) != 2 {
		t.Fatalf("first fan-out: expected 2, got %d", len(first))
	}

	// Repeat fan-out: everyone already invited, so nothing new and still 201.
	res = postReq(t, srv, "/api/sessions/"+sessionID+"/invites/club",
		map[string]any{"club_id": clubID}, aliceToken)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("repeat fan-out: expected 201, got %d", res.StatusCode)
	}
	var second []struct{}
	decodeBody(t, res, &second)
	if len(second) != 0 {
		t.Errorf("repeat fan-out: expected 0 new invites, got %d", len(second))
	}
}

// TestSessionClubInvite_NonMemberRejected confirms Club membership is the gate.
func TestSessionClubInvite_NonMemberRejected(t *testing.T) {
	srv, s := newAPITestServer(t)
	defer func() { _ = s.Close() }()

	aliceToken := mustRegister(t, srv, "alice@test.local", "Alice", "password123")
	strangerToken := mustRegister(t, srv, "stranger@test.local", "Stranger", "password123")

	clubID := mustCreateClub(t, srv, aliceToken, "Bouvet Padel")
	sessionID, _ := mustCreateSession(t, srv, strangerToken)

	res := postReq(t, srv, "/api/sessions/"+sessionID+"/invites/club",
		map[string]any{"club_id": clubID}, strangerToken)
	if res.StatusCode != http.StatusForbidden {
		t.Fatalf("non-member fan-out: expected 403, got %d", res.StatusCode)
	}
	var errBody struct {
		Error string `json:"error"`
	}
	decodeBody(t, res, &errBody)
	if errBody.Error != "not_club_member" {
		t.Errorf("expected not_club_member, got %q", errBody.Error)
	}
}

// TestSessionInvite_ClubEventMembersOnly enforces that a club event may only
// invite Members of its Club: a member invite succeeds, a non-member is refused.
func TestSessionInvite_ClubEventMembersOnly(t *testing.T) {
	srv, s := newAPITestServer(t)
	defer func() { _ = s.Close() }()

	aliceToken := mustRegister(t, srv, "alice@test.local", "Alice", "password123")
	bobToken := mustRegister(t, srv, "bob@test.local", "Bob", "password123")
	strangerToken := mustRegister(t, srv, "stranger@test.local", "Stranger", "password123")
	bobID := userID(t, srv, bobToken)
	strangerID := userID(t, srv, strangerToken)

	clubID := mustCreateClub(t, srv, aliceToken, "Bouvet Padel")
	joinClub(t, srv, aliceToken, bobToken, clubID)
	sessionID, _ := mustCreateClubEvent(t, srv, aliceToken, clubID)

	// Member Bob can be invited.
	res := postReq(t, srv, "/api/sessions/"+sessionID+"/invites", map[string]any{"to_user_id": bobID}, aliceToken)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("invite member: expected 201, got %d", res.StatusCode)
	}

	// Non-member Stranger is refused.
	res = postReq(t, srv, "/api/sessions/"+sessionID+"/invites", map[string]any{"to_user_id": strangerID}, aliceToken)
	if res.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("invite non-member: expected 422, got %d", res.StatusCode)
	}
	var errBody struct {
		Error string `json:"error"`
	}
	decodeBody(t, res, &errBody)
	if errBody.Error != "invitee_not_club_member" {
		t.Errorf("expected invitee_not_club_member, got %q", errBody.Error)
	}
}

// TestSessionInvite_NonClubUnrestricted confirms an ordinary (non-club) Session
// still invites anyone.
func TestSessionInvite_NonClubUnrestricted(t *testing.T) {
	srv, s := newAPITestServer(t)
	defer func() { _ = s.Close() }()

	aliceToken := mustRegister(t, srv, "alice@test.local", "Alice", "password123")
	bobToken := mustRegister(t, srv, "bob@test.local", "Bob", "password123")
	bobID := userID(t, srv, bobToken)

	sessionID, _ := mustCreateSession(t, srv, aliceToken)
	res := postReq(t, srv, "/api/sessions/"+sessionID+"/invites", map[string]any{"to_user_id": bobID}, aliceToken)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("invite on non-club session: expected 201, got %d", res.StatusCode)
	}
}

// TestSessionClubInvite_RequiresAuth rejects an unauthenticated fan-out.
func TestSessionClubInvite_RequiresAuth(t *testing.T) {
	srv, s := newAPITestServer(t)
	defer func() { _ = s.Close() }()

	aliceToken := mustRegister(t, srv, "alice@test.local", "Alice", "password123")
	clubID := mustCreateClub(t, srv, aliceToken, "Bouvet Padel")
	sessionID, _ := mustCreateSession(t, srv, aliceToken)

	res := postReq(t, srv, "/api/sessions/"+sessionID+"/invites/club",
		map[string]any{"club_id": clubID}, "")
	if res.StatusCode != http.StatusUnauthorized {
		t.Errorf("unauthenticated fan-out: expected 401, got %d", res.StatusCode)
	}
}
