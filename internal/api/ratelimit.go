package api

import (
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/httprate"
)

// RateLimitConfig bounds how many requests a single client IP may make to the
// abuse-sensitive auth endpoints (login/register/forgot) within Window before
// being rejected with 429. It is threaded through NewRouter so tests can run
// with a permissive budget and production can set a strict one.
type RateLimitConfig struct {
	Requests int
	Window   time.Duration
}

// DefaultRateLimit is the production auth rate limit: a per-client-IP budget
// generous enough for fat-finger retries but tight enough to blunt online
// password brute-force and registration/reset-email spam (#239).
var DefaultRateLimit = RateLimitConfig{Requests: 10, Window: time.Minute}

// authRateLimiter builds the middleware guarding the auth endpoints. A single
// instance is shared across login/register/forgot so the budget covers a
// client's total auth traffic, keyed by real client IP (see clientIP).
func authRateLimiter(cfg RateLimitConfig) func(http.Handler) http.Handler {
	return httprate.LimitBy(
		cfg.Requests,
		cfg.Window,
		func(r *http.Request) (string, error) { return clientIP(r), nil },
		httprate.WithLimitHandler(func(w http.ResponseWriter, _ *http.Request) {
			respondAPIError(w, ErrRateLimited)
		}),
	)
}

// clientIP resolves the originating client's IP for rate-limiting. Behind Fly's
// proxy the real client address arrives in Fly-Client-IP (trusted, set by the
// proxy); X-Forwarded-For's first hop is the next-best signal; RemoteAddr is the
// local-dev fallback when no proxy is in front. Keying on this — rather than
// RemoteAddr, which is the proxy in production — stops one client from
// exhausting everyone's budget.
func clientIP(r *http.Request) string {
	if ip := strings.TrimSpace(r.Header.Get("Fly-Client-IP")); ip != "" {
		return ip
	}
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if first, _, ok := strings.Cut(xff, ","); ok {
			return strings.TrimSpace(first)
		}
		return strings.TrimSpace(xff)
	}
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}
