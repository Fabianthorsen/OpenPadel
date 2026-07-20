package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fabianthorsen/openpadel/internal/api"
	"github.com/fabianthorsen/openpadel/internal/email"
)

// newAPITestServerRL builds a test server whose auth endpoints enforce the given
// rate limit, for exercising the limiter directly (the shared helper is
// deliberately permissive).
func newAPITestServerRL(t *testing.T, cfg api.RateLimitConfig) *httptest.Server {
	t.Helper()
	s := newAPITestStore(t)
	emailClient := email.NewClient("", "noreply@test.local")
	handler := api.NewRouter(s, emailClient, "http://localhost", "", "", cfg)
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv
}

// The abuse-sensitive auth endpoints must start returning 429 once one client
// exceeds the configured budget within the window (#239).
func TestAuthRateLimit_BlocksBurst(t *testing.T) {
	const limit = 3
	endpoints := []struct {
		path string
		body map[string]any
	}{
		{"/api/auth/login", map[string]any{"email": "x@y.z", "password": "nope"}},
		{"/api/auth/register", map[string]any{"email": "", "display_name": "", "password": ""}},
		{"/api/auth/forgot", map[string]any{"email": "x@y.z"}},
		{"/api/auth/reset", map[string]any{"token": "", "password": ""}},
	}
	for _, ep := range endpoints {
		t.Run(ep.path, func(t *testing.T) {
			// Fresh server (fresh counter) per endpoint.
			srv := newAPITestServerRL(t, api.RateLimitConfig{Requests: limit, Window: time.Minute})

			// The first `limit` requests pass the limiter (their status is
			// whatever the handler returns — never 429).
			for i := 0; i < limit; i++ {
				res := postReq(t, srv, ep.path, ep.body, "")
				if res.StatusCode == http.StatusTooManyRequests {
					t.Fatalf("request %d was rate-limited before the limit was reached", i+1)
				}
				_ = res.Body.Close()
			}

			// The next request over the limit is rejected with 429.
			res := postReq(t, srv, ep.path, ep.body, "")
			_ = res.Body.Close()
			if res.StatusCode != http.StatusTooManyRequests {
				t.Fatalf("expected 429 after %d requests, got %d", limit, res.StatusCode)
			}
		})
	}
}

// The limiter must bucket by the real client IP (Fly-Client-IP behind the Fly
// proxy), not by the shared proxy address — so one abusive client can't exhaust
// everyone else's budget (#239).
func TestAuthRateLimit_PerClientIP(t *testing.T) {
	const limit = 2
	srv := newAPITestServerRL(t, api.RateLimitConfig{Requests: limit, Window: time.Minute})
	body := map[string]any{"email": "x@y.z", "password": "nope"}

	exhaust := func(ip string) int {
		var last int
		for i := 0; i < limit+1; i++ {
			res := doRequest(t, srv, http.MethodPost, "/api/auth/login", body, "",
				map[string]string{"Fly-Client-IP": ip})
			last = res.StatusCode
			_ = res.Body.Close()
		}
		return last
	}

	// First client burns through its budget and gets blocked.
	if got := exhaust("1.1.1.1"); got != http.StatusTooManyRequests {
		t.Fatalf("client 1.1.1.1: expected 429 over limit, got %d", got)
	}

	// A different client IP has its own budget — its first request must not be
	// collateral-limited by the first client.
	res := doRequest(t, srv, http.MethodPost, "/api/auth/login", body, "",
		map[string]string{"Fly-Client-IP": "2.2.2.2"})
	_ = res.Body.Close()
	if res.StatusCode == http.StatusTooManyRequests {
		t.Fatal("client 2.2.2.2 was rate-limited on its first request; limiter is not keying per client IP")
	}
}
