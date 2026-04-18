package api_test

import (
	"net/http"
	"testing"
)

func TestGetRounds(t *testing.T) {
	srv, _ := newAPITestServer(t)
	sessID, adminToken, _ := setupStartedSession(t, srv)

	res := getReq(t, srv, "/api/sessions/"+sessID+"/rounds", adminToken)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
	var body struct {
		Rounds []struct {
			Number int `json:"number"`
		} `json:"rounds"`
	}
	decodeBody(t, res, &body)
	if len(body.Rounds) == 0 {
		t.Error("expected at least 1 round after session start")
	}
}

func TestGetRounds_LobbySession(t *testing.T) {
	srv, _ := newAPITestServer(t)
	sessID, _ := mustCreateSession(t, srv, "")

	res := getReq(t, srv, "/api/sessions/"+sessID+"/rounds", "")
	if res.StatusCode != http.StatusConflict {
		t.Fatalf("expected 409 for lobby session, got %d", res.StatusCode)
	}
	res.Body.Close()
}

func TestGetCurrentRound(t *testing.T) {
	srv, _ := newAPITestServer(t)
	sessID, adminToken, _ := setupStartedSession(t, srv)

	res := getReq(t, srv, "/api/sessions/"+sessID+"/rounds/current", adminToken)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
	var round struct {
		Number  int `json:"number"`
		Matches []struct {
			ID string `json:"id"`
		} `json:"matches"`
	}
	decodeBody(t, res, &round)
	if round.Number != 1 {
		t.Errorf("expected round number 1, got %d", round.Number)
	}
	if len(round.Matches) == 0 {
		t.Error("expected at least 1 match in current round")
	}
}

func TestSubmitScore(t *testing.T) {
	srv, _ := newAPITestServer(t)
	sessID, adminToken, _ := setupStartedSession(t, srv)

	// Get current round to find match ID
	res := getReq(t, srv, "/api/sessions/"+sessID+"/rounds/current", adminToken)
	var round struct {
		Matches []struct {
			ID string `json:"id"`
		} `json:"matches"`
	}
	decodeBody(t, res, &round)
	if len(round.Matches) == 0 {
		t.Fatal("no matches in current round")
	}
	matchID := round.Matches[0].ID

	res2 := putReq(t, srv, "/api/sessions/"+sessID+"/matches/"+matchID+"/score", map[string]any{
		"score_a": 16,
		"score_b": 8,
	}, adminToken)
	if res2.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res2.StatusCode)
	}
	var match struct {
		Score struct {
			A int `json:"a"`
			B int `json:"b"`
		} `json:"score"`
	}
	decodeBody(t, res2, &match)
	if match.Score.A != 16 || match.Score.B != 8 {
		t.Errorf("expected score 16-8, got %d-%d", match.Score.A, match.Score.B)
	}
}

func TestSubmitScore_InvalidSum(t *testing.T) {
	srv, _ := newAPITestServer(t)
	sessID, adminToken, _ := setupStartedSession(t, srv)

	res := getReq(t, srv, "/api/sessions/"+sessID+"/rounds/current", adminToken)
	var round struct {
		Matches []struct {
			ID string `json:"id"`
		} `json:"matches"`
	}
	decodeBody(t, res, &round)
	matchID := round.Matches[0].ID

	// 10+10=20 != 24
	res2 := putReq(t, srv, "/api/sessions/"+sessID+"/matches/"+matchID+"/score", map[string]any{
		"score_a": 10,
		"score_b": 10,
	}, adminToken)
	if res2.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid sum, got %d", res2.StatusCode)
	}
	res2.Body.Close()
}

func TestGetLeaderboard(t *testing.T) {
	srv, _ := newAPITestServer(t)
	sessID, adminToken, _ := setupStartedSession(t, srv)

	res := getReq(t, srv, "/api/sessions/"+sessID+"/leaderboard", adminToken)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
	var body struct {
		Standings []struct {
			Rank int `json:"rank"`
		} `json:"standings"`
	}
	decodeBody(t, res, &body)
	if len(body.Standings) != 4 {
		t.Errorf("expected 4 standings for 4 players, got %d", len(body.Standings))
	}
}

func TestAdvanceRound(t *testing.T) {
	srv, _ := newAPITestServer(t)
	sessID, adminToken, _ := setupStartedSession(t, srv)

	// Score all matches in round 1
	res := getReq(t, srv, "/api/sessions/"+sessID+"/rounds/current", adminToken)
	var round struct {
		Matches []struct {
			ID string `json:"id"`
		} `json:"matches"`
	}
	decodeBody(t, res, &round)
	for _, m := range round.Matches {
		putReq(t, srv, "/api/sessions/"+sessID+"/matches/"+m.ID+"/score", map[string]any{
			"score_a": 16,
			"score_b": 8,
		}, adminToken).Body.Close()
	}

	res2 := postReq(t, srv, "/api/sessions/"+sessID+"/rounds/advance", nil, adminToken)
	if res2.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", res2.StatusCode)
	}
	res2.Body.Close()
}

func TestAdvanceRound_RoundNotComplete(t *testing.T) {
	srv, _ := newAPITestServer(t)
	sessID, adminToken, _ := setupStartedSession(t, srv)

	res := postReq(t, srv, "/api/sessions/"+sessID+"/rounds/advance", nil, adminToken)
	if res.StatusCode != http.StatusConflict {
		t.Fatalf("expected 409 for unscored round, got %d", res.StatusCode)
	}
	res.Body.Close()
}

func TestAdvanceRound_RequiresAdmin(t *testing.T) {
	srv, _ := newAPITestServer(t)
	sessID, adminToken, _ := setupStartedSession(t, srv)

	res := getReq(t, srv, "/api/sessions/"+sessID+"/rounds/current", adminToken)
	var round struct {
		Matches []struct {
			ID string `json:"id"`
		} `json:"matches"`
	}
	decodeBody(t, res, &round)
	for _, m := range round.Matches {
		putReq(t, srv, "/api/sessions/"+sessID+"/matches/"+m.ID+"/score", map[string]any{
			"score_a": 16,
			"score_b": 8,
		}, adminToken).Body.Close()
	}

	res2 := postReq(t, srv, "/api/sessions/"+sessID+"/rounds/advance", nil, "")
	if res2.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", res2.StatusCode)
	}
	res2.Body.Close()
}
