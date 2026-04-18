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

// Timed Americano Tests

func TestSubmitScore_TimedAmericano_FreeScoring(t *testing.T) {
	srv, _ := newAPITestServer(t)
	// Create timed americano session
	userToken := mustRegister(t, srv, "admin@test.local", "Admin", "password123")
	sessRes := postReq(t, srv, "/api/sessions", map[string]any{
		"courts":                   1,
		"game_mode":                "timed_americano",
		"total_duration_minutes":   120,
		"buffer_seconds":           120,
	}, userToken)
	if sessRes.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", sessRes.StatusCode)
	}
	var sess struct {
		ID         string `json:"id"`
		AdminToken string `json:"admin_token"`
	}
	decodeBody(t, sessRes, &sess)

	// Join 4 players
	mustJoinSession(t, srv, sess.ID, "Alice", userToken)
	mustJoinSession(t, srv, sess.ID, "Bob", "")
	mustJoinSession(t, srv, sess.ID, "Charlie", "")
	mustJoinSession(t, srv, sess.ID, "Diana", "")

	// Start session
	mustStartSession(t, srv, sess.ID, sess.AdminToken)

	// Get current round
	res := getReq(t, srv, "/api/sessions/"+sess.ID+"/rounds/current", sess.AdminToken)
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

	// Submit free score (not constrained to a sum)
	res2 := putReq(t, srv, "/api/sessions/"+sess.ID+"/matches/"+matchID+"/score", map[string]any{
		"score_a": 23,
		"score_b": 19,
	}, sess.AdminToken)
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
	if match.Score.A != 23 || match.Score.B != 19 {
		t.Errorf("expected score 23-19, got %d-%d", match.Score.A, match.Score.B)
	}
}

func TestSubmitScore_TimedAmericano_ZeroZero(t *testing.T) {
	srv, _ := newAPITestServer(t)
	// Create timed americano session
	userToken := mustRegister(t, srv, "admin@test.local", "Admin", "password123")
	sessRes := postReq(t, srv, "/api/sessions", map[string]any{
		"courts":                   1,
		"game_mode":                "timed_americano",
		"total_duration_minutes":   120,
		"buffer_seconds":           120,
	}, userToken)
	var sess struct {
		ID         string `json:"id"`
		AdminToken string `json:"admin_token"`
	}
	decodeBody(t, sessRes, &sess)

	// Join 4 players and start
	for i, name := range []string{"Alice", "Bob", "Charlie", "Diana"} {
		token := ""
		if i == 0 {
			token = userToken
		}
		mustJoinSession(t, srv, sess.ID, name, token)
	}
	mustStartSession(t, srv, sess.ID, sess.AdminToken)

	// Get current round
	res := getReq(t, srv, "/api/sessions/"+sess.ID+"/rounds/current", sess.AdminToken)
	var round struct {
		Matches []struct {
			ID string `json:"id"`
		} `json:"matches"`
	}
	decodeBody(t, res, &round)
	matchID := round.Matches[0].ID

	// Submit 0-0 score (valid for timed americano)
	res2 := putReq(t, srv, "/api/sessions/"+sess.ID+"/matches/"+matchID+"/score", map[string]any{
		"score_a": 0,
		"score_b": 0,
	}, sess.AdminToken)
	if res2.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res2.StatusCode)
	}
}

func TestSubmitScore_TimedAmericano_NegativeScore(t *testing.T) {
	srv, _ := newAPITestServer(t)
	// Create timed americano session
	userToken := mustRegister(t, srv, "admin@test.local", "Admin", "password123")
	sessRes := postReq(t, srv, "/api/sessions", map[string]any{
		"courts":                   1,
		"game_mode":                "timed_americano",
		"total_duration_minutes":   120,
		"buffer_seconds":           120,
	}, userToken)
	var sess struct {
		ID         string `json:"id"`
		AdminToken string `json:"admin_token"`
	}
	decodeBody(t, sessRes, &sess)

	// Join 4 players and start
	for i, name := range []string{"Alice", "Bob", "Charlie", "Diana"} {
		token := ""
		if i == 0 {
			token = userToken
		}
		mustJoinSession(t, srv, sess.ID, name, token)
	}
	mustStartSession(t, srv, sess.ID, sess.AdminToken)

	// Get current round
	res := getReq(t, srv, "/api/sessions/"+sess.ID+"/rounds/current", sess.AdminToken)
	var round struct {
		Matches []struct {
			ID string `json:"id"`
		} `json:"matches"`
	}
	decodeBody(t, res, &round)
	matchID := round.Matches[0].ID

	// Attempt to submit negative score
	res2 := putReq(t, srv, "/api/sessions/"+sess.ID+"/matches/"+matchID+"/score", map[string]any{
		"score_a": -5,
		"score_b": 10,
	}, sess.AdminToken)
	if res2.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for negative score, got %d", res2.StatusCode)
	}
	res2.Body.Close()
}

func TestAdvanceRound_TimedAmericano(t *testing.T) {
	srv, _ := newAPITestServer(t)
	// Create timed americano session
	userToken := mustRegister(t, srv, "admin@test.local", "Admin", "password123")
	sessRes := postReq(t, srv, "/api/sessions", map[string]any{
		"courts":                   1,
		"game_mode":                "timed_americano",
		"total_duration_minutes":   120,
		"buffer_seconds":           120,
	}, userToken)
	var sess struct {
		ID         string `json:"id"`
		AdminToken string `json:"admin_token"`
	}
	decodeBody(t, sessRes, &sess)

	// Join 4 players and start
	for i, name := range []string{"Alice", "Bob", "Charlie", "Diana"} {
		token := ""
		if i == 0 {
			token = userToken
		}
		mustJoinSession(t, srv, sess.ID, name, token)
	}
	mustStartSession(t, srv, sess.ID, sess.AdminToken)

	// Score all matches in current round
	res := getReq(t, srv, "/api/sessions/"+sess.ID+"/rounds/current", sess.AdminToken)
	var round struct {
		Matches []struct {
			ID string `json:"id"`
		} `json:"matches"`
	}
	decodeBody(t, res, &round)
	for _, m := range round.Matches {
		putReq(t, srv, "/api/sessions/"+sess.ID+"/matches/"+m.ID+"/score", map[string]any{
			"score_a": 15,
			"score_b": 10,
		}, sess.AdminToken).Body.Close()
	}

	// Advance round
	res2 := postReq(t, srv, "/api/sessions/"+sess.ID+"/rounds/advance", nil, sess.AdminToken)
	if res2.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", res2.StatusCode)
	}
	res2.Body.Close()

	// Verify we're on round 2
	res3 := getReq(t, srv, "/api/sessions/"+sess.ID, sess.AdminToken)
	var sess2 struct {
		CurrentRound *int `json:"current_round"`
	}
	decodeBody(t, res3, &sess2)
	if sess2.CurrentRound == nil || *sess2.CurrentRound != 2 {
		t.Errorf("expected CurrentRound=2, got %v", sess2.CurrentRound)
	}
}

func TestTimedAmericanoFullFlow(t *testing.T) {
	srv, _ := newAPITestServer(t)
	// Create and start timed americano session
	userToken := mustRegister(t, srv, "admin@test.local", "Admin", "password123")
	sessRes := postReq(t, srv, "/api/sessions", map[string]any{
		"courts":                   1,
		"game_mode":                "timed_americano",
		"total_duration_minutes":   120,
		"buffer_seconds":           120,
	}, userToken)
	var sess struct {
		ID         string `json:"id"`
		AdminToken string `json:"admin_token"`
	}
	decodeBody(t, sessRes, &sess)

	// Join 4 players
	for i, name := range []string{"Alice", "Bob", "Charlie", "Diana"} {
		token := ""
		if i == 0 {
			token = userToken
		}
		mustJoinSession(t, srv, sess.ID, name, token)
	}
	mustStartSession(t, srv, sess.ID, sess.AdminToken)

	// Play one round
	res := getReq(t, srv, "/api/sessions/"+sess.ID+"/rounds/current", sess.AdminToken)
	var round struct {
		Matches []struct {
			ID string `json:"id"`
		} `json:"matches"`
	}
	decodeBody(t, res, &round)
	if len(round.Matches) == 0 {
		t.Fatal("expected at least 1 match")
	}

	// Score all matches with varied scores
	for _, m := range round.Matches {
		scoreRes := putReq(t, srv, "/api/sessions/"+sess.ID+"/matches/"+m.ID+"/score", map[string]any{
			"score_a": 23,
			"score_b": 19,
		}, sess.AdminToken)
		if scoreRes.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", scoreRes.StatusCode)
		}
		scoreRes.Body.Close()
	}

	// Advance round
	advRes := postReq(t, srv, "/api/sessions/"+sess.ID+"/rounds/advance", nil, sess.AdminToken)
	if advRes.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", advRes.StatusCode)
	}
	advRes.Body.Close()

	// Get leaderboard (session should still be active)
	res = getReq(t, srv, "/api/sessions/"+sess.ID+"/leaderboard", sess.AdminToken)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
	var lb map[string]any
	decodeBody(t, res, &lb)
	standings := lb["standings"].([]any)
	if len(standings) == 0 {
		t.Error("expected at least 1 player in leaderboard")
	}
}
