package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/fabianthorsen/openpadel/internal/api"
	"github.com/fabianthorsen/openpadel/internal/email"
	"github.com/fabianthorsen/openpadel/internal/store"
)

func newAPITestStore(t *testing.T) *store.Store {
	t.Helper()
	f, err := os.CreateTemp("", "openpadel-api-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	t.Cleanup(func() { os.Remove(f.Name()) })

	s, err := store.Open(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func newAPITestServer(t *testing.T) (*httptest.Server, *store.Store) {
	t.Helper()
	s := newAPITestStore(t)
	emailClient := email.NewClient("", "noreply@test.local")
	handler := api.NewRouter(s, emailClient, "http://localhost", "", "")
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv, s
}

func doRequest(t *testing.T, srv *httptest.Server, method, path string, body any, token string, extraHeaders map[string]string) *http.Response {
	t.Helper()
	var reqBody *bytes.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		reqBody = bytes.NewReader(b)
	} else {
		reqBody = bytes.NewReader(nil)
	}
	req, err := http.NewRequest(method, srv.URL+path, reqBody)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	for k, v := range extraHeaders {
		req.Header.Set(k, v)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	return res
}

func getReq(t *testing.T, srv *httptest.Server, path, token string) *http.Response {
	t.Helper()
	return doRequest(t, srv, http.MethodGet, path, nil, token, nil)
}

func postReq(t *testing.T, srv *httptest.Server, path string, body any, token string) *http.Response {
	t.Helper()
	return doRequest(t, srv, http.MethodPost, path, body, token, nil)
}

func deleteReq(t *testing.T, srv *httptest.Server, path, token string) *http.Response {
	t.Helper()
	return doRequest(t, srv, http.MethodDelete, path, nil, token, nil)
}

func putReq(t *testing.T, srv *httptest.Server, path string, body any, token string) *http.Response {
	t.Helper()
	return doRequest(t, srv, http.MethodPut, path, body, token, nil)
}

func decodeBody(t *testing.T, res *http.Response, v any) {
	t.Helper()
	defer res.Body.Close()
	if err := json.NewDecoder(res.Body).Decode(v); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
}

// mustRegister registers a new user and returns the auth token.
func mustRegister(t *testing.T, srv *httptest.Server, emailAddr, name, password string) string {
	t.Helper()
	res := postReq(t, srv, "/api/auth/register", map[string]any{
		"email":        emailAddr,
		"display_name": name,
		"password":     password,
	}, "")
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("register: expected 201, got %d", res.StatusCode)
	}
	var body struct {
		Token string `json:"token"`
	}
	decodeBody(t, res, &body)
	return body.Token
}

// mustCreateSession creates an americano session and returns its ID and admin token.
func mustCreateSession(t *testing.T, srv *httptest.Server, token string) (id, adminToken string) {
	t.Helper()
	res := postReq(t, srv, "/api/sessions", map[string]any{
		"courts":    1,
		"points":    24,
		"game_mode": "americano",
	}, token)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("createSession: expected 201, got %d", res.StatusCode)
	}
	var body struct {
		ID         string `json:"id"`
		AdminToken string `json:"admin_token"`
	}
	decodeBody(t, res, &body)
	return body.ID, body.AdminToken
}

// mustJoinSession joins a session and returns the player ID.
func mustJoinSession(t *testing.T, srv *httptest.Server, sessionID, name, token string) string {
	t.Helper()
	res := postReq(t, srv, "/api/sessions/"+sessionID+"/players", map[string]any{"name": name}, token)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("joinSession(%q): expected 201, got %d", name, res.StatusCode)
	}
	var body struct {
		ID string `json:"id"`
	}
	decodeBody(t, res, &body)
	return body.ID
}

// mustStartSession starts a session (requires admin token and 4+ players already joined).
func mustStartSession(t *testing.T, srv *httptest.Server, sessionID, adminToken string) {
	t.Helper()
	res := postReq(t, srv, "/api/sessions/"+sessionID+"/start", nil, adminToken)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("startSession: expected 200, got %d", res.StatusCode)
	}
	res.Body.Close()
}

// setupStartedSession creates a session, joins 4 players, and starts it.
// Returns the session ID, admin token, and 4 player IDs.
func setupStartedSession(t *testing.T, srv *httptest.Server) (sessID, adminToken string, playerIDs [4]string) {
	t.Helper()
	userToken := mustRegister(t, srv, "admin@test.local", "Admin", "password123")
	sessID, adminToken = mustCreateSession(t, srv, userToken)

	playerIDs[0] = mustJoinSession(t, srv, sessID, "Alice", adminToken)
	playerIDs[1] = mustJoinSession(t, srv, sessID, "Bob", "")
	playerIDs[2] = mustJoinSession(t, srv, sessID, "Charlie", "")
	playerIDs[3] = mustJoinSession(t, srv, sessID, "Diana", "")

	mustStartSession(t, srv, sessID, adminToken)
	return
}
