package api_test

import (
	"net/http"
	"testing"
)

// TestGetClubLeaderboard_MemberGate covers the endpoint's own responsibilities:
// members get a 200 with both lists present (serialized as [] not null), a
// non-member is refused with 403, and an anonymous caller with 401. Row
// selection and ranking are covered by the store and domain tests.
func TestGetClubLeaderboard_MemberGate(t *testing.T) {
	srv, s := newAPITestServer(t)
	defer func() { _ = s.Close() }()

	alice := mustRegister(t, srv, "alice@test.local", "Alice", "password123")
	bob := mustRegister(t, srv, "bob@test.local", "Bob", "password123")
	clubID := mustCreateClub(t, srv, alice, "Bouvet Padel")

	// Member: 200 with a well-formed, empty board.
	res := getReq(t, srv, "/api/clubs/"+clubID+"/leaderboard", alice)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("member: expected 200, got %d", res.StatusCode)
	}
	var board struct {
		Ranked      []any `json:"ranked"`
		Provisional []any `json:"provisional"`
	}
	decodeBody(t, res, &board)
	if board.Ranked == nil {
		t.Error("ranked should serialize as [] not null")
	}
	if board.Provisional == nil {
		t.Error("provisional should serialize as [] not null")
	}

	// Non-member: 403.
	res = getReq(t, srv, "/api/clubs/"+clubID+"/leaderboard", bob)
	if res.StatusCode != http.StatusForbidden {
		t.Errorf("non-member: expected 403, got %d", res.StatusCode)
	}

	// Anonymous: 401.
	res = getReq(t, srv, "/api/clubs/"+clubID+"/leaderboard", "")
	if res.StatusCode != http.StatusUnauthorized {
		t.Errorf("anonymous: expected 401, got %d", res.StatusCode)
	}
}
