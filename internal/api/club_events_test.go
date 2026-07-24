package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// mustCreateClubEvent creates a Session owned by clubID and returns its id and
// admin token. The caller (token) must be a Member of the Club.
func mustCreateClubEvent(t *testing.T, srv *httptest.Server, token, clubID string) (id, adminToken string) {
	t.Helper()
	res := postReq(t, srv, "/api/sessions", map[string]any{
		"courts":    1,
		"points":    24,
		"game_mode": "americano",
		"club_id":   clubID,
	}, token)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("createClubEvent: expected 201, got %d", res.StatusCode)
	}
	var body struct {
		ID         string `json:"id"`
		AdminToken string `json:"admin_token"`
		ClubID     string `json:"club_id"`
	}
	decodeBody(t, res, &body)
	if body.ClubID != clubID {
		t.Fatalf("createClubEvent: club_id = %q, want %q", body.ClubID, clubID)
	}
	return body.ID, body.AdminToken
}

// TestClubEventCreateAndDiscovery covers the happy path: a Member creates a club
// event with club_id set, and it surfaces in the Club's events feed.
func TestClubEventCreateAndDiscovery(t *testing.T) {
	srv, s := newAPITestServer(t)
	defer func() { _ = s.Close() }()

	adminToken := mustRegister(t, srv, "alice@test.local", "Alice", "password123")
	memberToken := mustRegister(t, srv, "bob@test.local", "Bob", "password123")

	clubID := mustCreateClub(t, srv, adminToken, "Bouvet Padel")
	joinClub(t, srv, adminToken, memberToken, clubID)

	// Any Member (not just admins) can create a club event.
	eventID, _ := mustCreateClubEvent(t, srv, memberToken, clubID)

	res := getReq(t, srv, "/api/clubs/"+clubID+"/events", adminToken)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("getClubEvents: expected 200, got %d", res.StatusCode)
	}
	var events []struct {
		SessionID string `json:"session_id"`
		Status    string `json:"status"`
	}
	decodeBody(t, res, &events)
	if len(events) != 1 {
		t.Fatalf("expected 1 club event, got %d", len(events))
	}
	if events[0].SessionID != eventID {
		t.Errorf("event session_id = %q, want %q", events[0].SessionID, eventID)
	}
	if events[0].Status != "lobby" {
		t.Errorf("event status = %q, want lobby", events[0].Status)
	}

	// A done/absent Session must not appear; a plain (non-club) Session created by
	// the same user stays out of the Club's feed.
	mustCreateSession(t, srv, memberToken)
	res = getReq(t, srv, "/api/clubs/"+clubID+"/events", memberToken)
	decodeBody(t, res, &events)
	if len(events) != 1 {
		t.Fatalf("plain session leaked into club feed: got %d events", len(events))
	}
}

// TestClubEventNonMemberRefused asserts that a User who is not a Member of the
// Club cannot attach a Session to it.
func TestClubEventNonMemberRefused(t *testing.T) {
	srv, s := newAPITestServer(t)
	defer func() { _ = s.Close() }()

	adminToken := mustRegister(t, srv, "alice@test.local", "Alice", "password123")
	outsiderToken := mustRegister(t, srv, "carol@test.local", "Carol", "password123")

	clubID := mustCreateClub(t, srv, adminToken, "Bouvet Padel")

	res := postReq(t, srv, "/api/sessions", map[string]any{
		"courts":    1,
		"points":    24,
		"game_mode": "americano",
		"club_id":   clubID,
	}, outsiderToken)
	if res.StatusCode != http.StatusForbidden {
		t.Fatalf("non-member create: expected 403, got %d", res.StatusCode)
	}
	_ = res.Body.Close()

	// Non-members also can't read the Club's events feed.
	res = getReq(t, srv, "/api/clubs/"+clubID+"/events", outsiderToken)
	if res.StatusCode != http.StatusForbidden {
		t.Fatalf("non-member events: expected 403, got %d", res.StatusCode)
	}
	_ = res.Body.Close()
}

// TestClubAdminHasNoSessionPower is the authorization boundary from the spec: a
// Club Admin who does not hold the club event's AdminToken is refused every
// Session admin action (start/score/close/cancel/kick). A Club role never
// bridges into Session authority.
func TestClubAdminHasNoSessionPower(t *testing.T) {
	srv, s := newAPITestServer(t)
	defer func() { _ = s.Close() }()

	// Alice creates (and thus admins) the Club; Bob is a plain Member.
	clubAdminToken := mustRegister(t, srv, "alice@test.local", "Alice", "password123")
	memberToken := mustRegister(t, srv, "bob@test.local", "Bob", "password123")

	clubID := mustCreateClub(t, srv, clubAdminToken, "Bouvet Padel")
	joinClub(t, srv, clubAdminToken, memberToken, clubID)

	// Bob creates the club event, so Bob holds its AdminToken — not Alice.
	eventID, eventAdminToken := mustCreateClubEvent(t, srv, memberToken, clubID)

	// Alice (Club Admin, no event AdminToken) is refused starting the lobby event.
	res := postReq(t, srv, "/api/sessions/"+eventID+"/start", nil, clubAdminToken)
	if res.StatusCode != http.StatusForbidden {
		t.Fatalf("club admin start: expected 403, got %d", res.StatusCode)
	}
	_ = res.Body.Close()

	// Bob (the real Session admin) fills and starts the event so score/close have
	// a running Session to act on.
	p1 := mustJoinSession(t, srv, eventID, "Alice", eventAdminToken)
	mustJoinSession(t, srv, eventID, "Bob", "")
	mustJoinSession(t, srv, eventID, "Charlie", "")
	mustJoinSession(t, srv, eventID, "Diana", "")
	mustStartSession(t, srv, eventID, eventAdminToken)

	// Fetch a match id via Bob's admin token for the score attempt.
	roundRes := getReq(t, srv, "/api/sessions/"+eventID+"/rounds/current", eventAdminToken)
	var round struct {
		Matches []struct {
			ID string `json:"id"`
		} `json:"matches"`
	}
	decodeBody(t, roundRes, &round)
	if len(round.Matches) == 0 {
		t.Fatal("started event has no matches")
	}
	matchID := round.Matches[0].ID

	// Every Session admin action must refuse Alice, who holds no event AdminToken.
	cases := []struct {
		name string
		res  *http.Response
	}{
		{"score", putReq(t, srv, "/api/sessions/"+eventID+"/matches/"+matchID+"/score", map[string]any{"score_a": 16, "score_b": 8}, clubAdminToken)},
		{"close", postReq(t, srv, "/api/sessions/"+eventID+"/close", nil, clubAdminToken)},
		{"cancel", deleteReq(t, srv, "/api/sessions/"+eventID, clubAdminToken)},
		{"kick", deleteReq(t, srv, "/api/sessions/"+eventID+"/players/"+p1, clubAdminToken)},
	}
	for _, c := range cases {
		if c.res.StatusCode != http.StatusForbidden {
			t.Errorf("club admin %s: expected 403, got %d", c.name, c.res.StatusCode)
		}
		_ = c.res.Body.Close()
	}

	// Sanity: the event is still live and un-cancelled — the boundary held.
	res = getReq(t, srv, "/api/sessions/"+eventID, memberToken)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("event should still exist after refused admin actions, got %d", res.StatusCode)
	}
	_ = res.Body.Close()
}

// TestClubEventGuestJoin asserts a Guest can still join a club event via its
// public Session join link (by name), exactly like any other Session.
func TestClubEventGuestJoin(t *testing.T) {
	srv, s := newAPITestServer(t)
	defer func() { _ = s.Close() }()

	memberToken := mustRegister(t, srv, "bob@test.local", "Bob", "password123")
	clubID := mustCreateClub(t, srv, memberToken, "Bouvet Padel")
	eventID, _ := mustCreateClubEvent(t, srv, memberToken, clubID)

	// No auth token — a Guest joining by name.
	guestID, _ := mustJoinSessionWithToken(t, srv, eventID, "Zoe")
	if guestID == "" {
		t.Fatal("guest join returned empty player id")
	}
}
