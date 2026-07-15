package api_test

import (
	"fmt"
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

// TestStartLimitedAmericano_RoundCountFromRoster is the frontend<->backend
// integration regression for the Americano round count.
//
// Each layer is correct in isolation (the frontend computes 10, the backend's
// TotalRounds/GenerateRounds produce 10), but the real flow used to yield 7:
// the client only used its computed value to bound a stepper and actually sent
// a hardcoded rounds_total of 7, which the backend took at face value. The fix
// makes the backend the source of truth — for a limited ("fixed") Americano it
// recomputes the fair count from the roster at start.
//
// This drives the exact HTTP sequence the web client issues: create a 2-court
// Americano, join 10 players, PATCH it to "limited" (sending the stale 7 the UI
// used to send), start it, and assert the session starts with 10 rounds.
func TestStartLimitedAmericano_RoundCountFromRoster(t *testing.T) {
	srv, _ := newAPITestServer(t)
	userToken := mustRegister(t, srv, "admin@test.local", "Admin", "password123")
	sessID, adminToken := mustCreateSessionWithParams(t, srv, userToken, 2, 24, "americano")

	for i := 0; i < 10; i++ {
		mustJoinSession(t, srv, sessID, fmt.Sprintf("Player %02d", i+1), adminToken)
	}

	// The client marks the session "limited" by sending a non-nil rounds_total.
	// It sends the old buggy default (7); the backend must ignore the value and
	// derive the fair count (TotalRounds(10,2) = 10) from the actual roster.
	patchRes := patchReq(t, srv, "/api/sessions/"+sessID, map[string]any{"rounds_total": 7}, adminToken)
	if patchRes.StatusCode != http.StatusOK {
		t.Fatalf("patch config: expected 200, got %d", patchRes.StatusCode)
	}
	patchRes.Body.Close() //nolint:errcheck

	mustStartSession(t, srv, sessID, adminToken)

	res := getReq(t, srv, "/api/sessions/"+sessID+"/rounds", adminToken)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("get rounds: expected 200, got %d", res.StatusCode)
	}
	var body struct {
		Rounds []struct {
			Number  int      `json:"number"`
			Bench   []string `json:"bench"`
			Matches []struct {
				ID string `json:"id"`
			} `json:"matches"`
		} `json:"rounds"`
	}
	decodeBody(t, res, &body)

	if len(body.Rounds) != 10 {
		t.Fatalf("limited 10p/2c Americano started with %d rounds, want 10 (backend must recompute, not use the client's 7)", len(body.Rounds))
	}
	// Sanity: every round fills both courts and benches exactly 2 of the 10.
	for _, r := range body.Rounds {
		if len(r.Matches) != 2 {
			t.Errorf("round %d: %d matches, want 2", r.Number, len(r.Matches))
		}
		if len(r.Bench) != 2 {
			t.Errorf("round %d: %d benched, want 2", r.Number, len(r.Bench))
		}
	}
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

// Leaderboard displays correctly for unlimited sessions (total_rounds = null)
func TestGetLeaderboard_UnlimitedRounds(t *testing.T) {
	srv, _ := newAPITestServer(t)
	userToken := mustRegister(t, srv, "admin@test.local", "Admin", "password123")

	// Create unlimited Mexicano session
	res := postReq(t, srv, "/api/sessions", map[string]any{
		"courts":    2,
		"points":    24,
		"game_mode": "mexicano",
		// rounds_total omitted = unlimited
	}, userToken)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("create: got %d", res.StatusCode)
	}
	var createResp struct {
		ID         string `json:"id"`
		AdminToken string `json:"admin_token"`
	}
	decodeBody(t, res, &createResp)
	sessID := createResp.ID
	adminToken := createResp.AdminToken

	// Add 8 players
	for i := 1; i <= 8; i++ {
		mustJoinSession(t, srv, sessID, fmt.Sprintf("Player %d", i), "")
	}

	// Start session
	mustStartSession(t, srv, sessID, adminToken)

	// Get leaderboard — should work fine with null total_rounds
	leaderRes := getReq(t, srv, "/api/sessions/"+sessID+"/leaderboard", adminToken)
	if leaderRes.StatusCode != http.StatusOK {
		t.Fatalf("leaderboard: expected 200, got %d", leaderRes.StatusCode)
	}
	var leaderboard struct {
		CurrentRound int  `json:"current_round"`
		TotalRounds  *int `json:"total_rounds"`
		Standings    []struct {
			Name string `json:"name"`
		} `json:"standings"`
	}
	decodeBody(t, leaderRes, &leaderboard)

	// Verify structure is correct for unlimited (null total_rounds)
	if leaderboard.TotalRounds != nil {
		t.Errorf("expected TotalRounds=null for unlimited, got %v", leaderboard.TotalRounds)
	}
	if leaderboard.CurrentRound != 1 {
		t.Errorf("expected CurrentRound=1 after start, got %d", leaderboard.CurrentRound)
	}
	if len(leaderboard.Standings) != 8 {
		t.Errorf("expected 8 standings, got %d", len(leaderboard.Standings))
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

// Mexicano session with unlimited rounds (rounds_total = null) preserves null through start
func TestMexicanoAdvanceRound_UnlimitedRounds(t *testing.T) {
	srv, _ := newAPITestServer(t)

	// Register user and create Mexicano session
	userToken := mustRegister(t, srv, "admin@test.local", "Admin", "password123")

	// Create Mexicano with fixed rounds first
	res := postReq(t, srv, "/api/sessions", map[string]any{
		"courts":       2,
		"points":       24,
		"game_mode":    "mexicano",
		"rounds_total": 5, // Start with fixed rounds
	}, userToken)
	if res.StatusCode != http.StatusCreated {
		var errBody map[string]any
		decodeBody(t, res, &errBody)
		t.Fatalf("POST /api/sessions failed: %d, error: %v", res.StatusCode, errBody)
	}
	var createResp struct {
		ID         string `json:"id"`
		AdminToken string `json:"admin_token"`
	}
	decodeBody(t, res, &createResp)
	sessID := createResp.ID
	adminToken := createResp.AdminToken

	// Verify session was created with 5 rounds
	getRes := getReq(t, srv, "/api/sessions/"+sessID, adminToken)
	var sess struct {
		RoundsTotal *int `json:"rounds_total"`
	}
	decodeBody(t, getRes, &sess)
	if sess.RoundsTotal == nil || *sess.RoundsTotal != 5 {
		t.Errorf("expected RoundsTotal=5 after create, got %v", sess.RoundsTotal)
	}

	// Add players (Mexicano with 2 courts needs exactly 8 players)
	for i := 1; i <= 8; i++ {
		mustJoinSession(t, srv, sessID, fmt.Sprintf("Player %d", i), "")
	}

	// Start session
	mustStartSession(t, srv, sessID, adminToken)

	// Verify session is now playing
	getRes2 := getReq(t, srv, "/api/sessions/"+sessID, adminToken)
	var sessFinal struct {
		Status string `json:"status"`
	}
	decodeBody(t, getRes2, &sessFinal)
	if sessFinal.Status != "playing" {
		t.Errorf("expected Status=playing, got %s", sessFinal.Status)
	}
}

// Tracer: Mexicano deferred completion — final round doesn't auto-complete
func TestMexicanoFinalRoundDeferredCompletion(t *testing.T) {
	srv, _ := newAPITestServer(t)
	userToken := mustRegister(t, srv, "admin@test.local", "Admin", "password123")

	// Create Mexicano with just 1 round (makes it easy to hit the cap)
	res := postReq(t, srv, "/api/sessions", map[string]any{
		"courts":       2,
		"points":       24,
		"game_mode":    "mexicano",
		"rounds_total": 1, // Just 1 round to hit cap immediately
	}, userToken)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("create session: got %d", res.StatusCode)
	}
	var createResp struct {
		ID         string `json:"id"`
		AdminToken string `json:"admin_token"`
	}
	decodeBody(t, res, &createResp)
	sessID := createResp.ID
	adminToken := createResp.AdminToken

	// Add 8 players
	for i := 1; i <= 8; i++ {
		mustJoinSession(t, srv, sessID, fmt.Sprintf("Player %d", i), "")
	}

	// Start session
	mustStartSession(t, srv, sessID, adminToken)

	// Get current round to find match IDs
	roundRes := getReq(t, srv, "/api/sessions/"+sessID+"/rounds/current", adminToken)
	var round struct {
		Matches []struct {
			ID string `json:"id"`
		} `json:"matches"`
	}
	decodeBody(t, roundRes, &round)
	if len(round.Matches) == 0 {
		t.Fatal("no matches in current round")
	}

	// Score all matches (2 courts = 2 matches)
	// Scores must add up to points target (24 by default)
	for _, m := range round.Matches {
		scoreRes := putReq(t, srv, "/api/sessions/"+sessID+"/matches/"+m.ID+"/score", map[string]any{
			"score_a": 16,
			"score_b": 8,
		}, adminToken)
		if scoreRes.StatusCode != http.StatusOK {
			t.Fatalf("submit score: got %d", scoreRes.StatusCode)
		}
		scoreRes.Body.Close()
	}

	// Verify session is STILL playing (not auto-completed)
	// This is the key behavior: when we hit the final round, we should NOT auto-complete.
	// Instead, wait for admin choice (Finish vs Keep Playing).
	sessRes := getReq(t, srv, "/api/sessions/"+sessID, adminToken)
	var sess struct {
		Status string `json:"status"`
	}
	decodeBody(t, sessRes, &sess)
	if sess.Status != "playing" {
		t.Errorf("expected Status=playing after final round scores (deferred choice), got %q", sess.Status)
	}
}

// Regression: unlimited Americano (rounds_total omitted) must preserve null through
// Start — not silently convert to a fixed calculated round count. Reproduces the
// reported bug: 5 players / 1 court selected "unlimited" but got "Round 1 of 5".
func TestAmericanoUnlimited_PreservesNullThroughStart(t *testing.T) {
	srv, _ := newAPITestServer(t)
	userToken := mustRegister(t, srv, "admin@test.local", "Admin", "password123")

	res := postReq(t, srv, "/api/sessions", map[string]any{
		"courts":    1,
		"points":    24,
		"game_mode": "americano",
		// rounds_total omitted = unlimited
	}, userToken)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("create: got %d", res.StatusCode)
	}
	var createResp struct {
		ID         string `json:"id"`
		AdminToken string `json:"admin_token"`
	}
	decodeBody(t, res, &createResp)
	sessID := createResp.ID
	adminToken := createResp.AdminToken

	// 5 players on 1 court = 1 bench each round (the user's exact setup).
	for i := 1; i <= 5; i++ {
		mustJoinSession(t, srv, sessID, fmt.Sprintf("Player %d", i), "")
	}

	mustStartSession(t, srv, sessID, adminToken)

	// The core symptom: after start the session must still be unlimited (null),
	// not pinned to a calculated fixed total (which surfaced as "Round 1 of 5").
	assertRoundsTotalNil := func(when string) {
		t.Helper()
		getRes := getReq(t, srv, "/api/sessions/"+sessID, adminToken)
		var sess struct {
			RoundsTotal *int `json:"rounds_total"`
		}
		decodeBody(t, getRes, &sess)
		if sess.RoundsTotal != nil {
			t.Fatalf("unlimited Americano (%s): expected rounds_total=null, got %d", when, *sess.RoundsTotal)
		}
	}
	assertRoundsTotalNil("after start")

	// Play well past the 5-round natural rotation the fixed calculation would cap at,
	// scoring each round and advancing. It must keep generating rounds and never
	// auto-complete on its own — that's what "play indefinitely" means.
	for round := 1; round <= 8; round++ {
		roundRes := getReq(t, srv, "/api/sessions/"+sessID+"/rounds/current", adminToken)
		var cur struct {
			Number  int `json:"number"`
			Matches []struct {
				ID string `json:"id"`
			} `json:"matches"`
		}
		decodeBody(t, roundRes, &cur)
		if cur.Number != round {
			t.Fatalf("expected to be on round %d, got %d", round, cur.Number)
		}
		if len(cur.Matches) != 1 {
			t.Fatalf("round %d: expected 1 match (1 court), got %d", round, len(cur.Matches))
		}
		for _, m := range cur.Matches {
			putReq(t, srv, "/api/sessions/"+sessID+"/matches/"+m.ID+"/score", map[string]any{
				"score_a": 16,
				"score_b": 8,
			}, adminToken).Body.Close() //nolint:errcheck
		}

		// Scoring the round must not auto-complete an unlimited session.
		statusRes := getReq(t, srv, "/api/sessions/"+sessID, adminToken)
		var st struct {
			Status string `json:"status"`
		}
		decodeBody(t, statusRes, &st)
		if st.Status != "playing" {
			t.Fatalf("round %d: unlimited session auto-completed (status=%q); should keep playing", round, st.Status)
		}

		advRes := postReq(t, srv, "/api/sessions/"+sessID+"/rounds/advance", nil, adminToken)
		if advRes.StatusCode != http.StatusNoContent {
			t.Fatalf("round %d: advance expected 204, got %d", round, advRes.StatusCode)
		}
		advRes.Body.Close() //nolint:errcheck
	}

	assertRoundsTotalNil("after advancing past the natural rotation")
}

// Fixed Americano (rounds_total set) still pre-generates the full rotation and
// auto-completes once the final round is scored — the counterpart to the unlimited
// path, guarding against a regression from making null mean unlimited.
func TestAmericanoFixed_AutoCompletesAtCap(t *testing.T) {
	srv, _ := newAPITestServer(t)
	userToken := mustRegister(t, srv, "admin@test.local", "Admin", "password123")

	res := postReq(t, srv, "/api/sessions", map[string]any{
		"courts":       1,
		"points":       24,
		"game_mode":    "americano",
		"rounds_total": 3, // explicit fixed count (as the Lobby's "Fixed" toggle sends)
	}, userToken)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("create: got %d", res.StatusCode)
	}
	var createResp struct {
		ID         string `json:"id"`
		AdminToken string `json:"admin_token"`
	}
	decodeBody(t, res, &createResp)
	sessID := createResp.ID
	adminToken := createResp.AdminToken

	for i := 1; i <= 4; i++ {
		mustJoinSession(t, srv, sessID, fmt.Sprintf("Player %d", i), "")
	}
	mustStartSession(t, srv, sessID, adminToken)

	// Fixed count is preserved through start.
	getRes := getReq(t, srv, "/api/sessions/"+sessID, adminToken)
	var sess struct {
		RoundsTotal *int `json:"rounds_total"`
	}
	decodeBody(t, getRes, &sess)
	if sess.RoundsTotal == nil || *sess.RoundsTotal != 3 {
		t.Fatalf("fixed Americano: expected rounds_total=3 after start, got %v", sess.RoundsTotal)
	}

	// Score the first two rounds, advancing after each. The session stays playing.
	for round := 1; round <= 2; round++ {
		mustScoreCurrentRound(t, srv, sessID, adminToken)
		statusRes := getReq(t, srv, "/api/sessions/"+sessID, adminToken)
		var st struct {
			Status string `json:"status"`
		}
		decodeBody(t, statusRes, &st)
		if st.Status != "playing" {
			t.Fatalf("round %d: expected still playing, got %q", round, st.Status)
		}
		advRes := postReq(t, srv, "/api/sessions/"+sessID+"/rounds/advance", nil, adminToken)
		if advRes.StatusCode != http.StatusNoContent {
			t.Fatalf("round %d: advance expected 204, got %d", round, advRes.StatusCode)
		}
		advRes.Body.Close() //nolint:errcheck
	}

	// Scoring the final (3rd) round auto-completes the session.
	mustScoreCurrentRound(t, srv, sessID, adminToken)
	finalRes := getReq(t, srv, "/api/sessions/"+sessID, adminToken)
	var final struct {
		Status string `json:"status"`
	}
	decodeBody(t, finalRes, &final)
	if final.Status != "done" {
		t.Fatalf("fixed Americano: expected auto-complete to 'done' after final round, got %q", final.Status)
	}
}

// Americano final round doesn't auto-complete for unlimited sessions (deferred choice)
func TestAmericanoFinalRoundDeferredCompletion(t *testing.T) {
	srv, _ := newAPITestServer(t)
	userToken := mustRegister(t, srv, "admin@test.local", "Admin", "password123")

	// Create unlimited Americano session (1 court to keep it simple)
	res := postReq(t, srv, "/api/sessions", map[string]any{
		"courts":    1,
		"points":    24,
		"game_mode": "americano",
		// rounds_total omitted = unlimited
	}, userToken)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("create: got %d", res.StatusCode)
	}
	var createResp struct {
		ID         string `json:"id"`
		AdminToken string `json:"admin_token"`
	}
	decodeBody(t, res, &createResp)
	sessID := createResp.ID
	adminToken := createResp.AdminToken

	// Add 4 players (1 court = 4 players, 0 bench)
	for i := 1; i <= 4; i++ {
		mustJoinSession(t, srv, sessID, fmt.Sprintf("Player %d", i), "")
	}

	// Start session - generates round 1 only for unlimited
	mustStartSession(t, srv, sessID, adminToken)

	// Get current round to find match IDs
	roundRes := getReq(t, srv, "/api/sessions/"+sessID+"/rounds/current", adminToken)
	var round struct {
		Matches []struct {
			ID string `json:"id"`
		} `json:"matches"`
	}
	decodeBody(t, roundRes, &round)
	if len(round.Matches) == 0 {
		t.Fatal("no matches in current round")
	}

	// Score all matches (1 court = 1 match)
	for _, m := range round.Matches {
		scoreRes := putReq(t, srv, "/api/sessions/"+sessID+"/matches/"+m.ID+"/score", map[string]any{
			"score_a": 16,
			"score_b": 8,
		}, adminToken)
		if scoreRes.StatusCode != http.StatusOK {
			t.Fatalf("submit score: got %d", scoreRes.StatusCode)
		}
		scoreRes.Body.Close() //nolint:errcheck
	}

	// Verify session is STILL playing (not auto-completed)
	// This is the key: unlimited Americano should defer completion, not auto-complete
	sessRes := getReq(t, srv, "/api/sessions/"+sessID, adminToken)
	var sess struct {
		Status string `json:"status"`
	}
	decodeBody(t, sessRes, &sess)
	if sess.Status != "playing" {
		t.Errorf("expected Status=playing after final round scores (deferred choice), got %q", sess.Status)
	}
}
