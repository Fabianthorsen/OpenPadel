package api

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/fabianthorsen/openpadel/internal/domain"
	"github.com/fabianthorsen/openpadel/internal/store"
)

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)
		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", ww.Status(),
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}

type contextKey string

const keyAdminToken contextKey = "admin_token"
const contextKeyUser contextKey = "user"

// extractAdminToken reads a Bearer token from the Authorization header.
// Used for session admin checks (token == session.admin_token).
func extractAdminToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if after, ok := strings.CutPrefix(auth, "Bearer "); ok {
		return strings.TrimSpace(after)
	}
	return ""
}

func isAdmin(token, sessionAdminToken string) bool {
	return token != "" && token == sessionAdminToken
}

// optionalAuth resolves a bearer token to a user if present, but never
// rejects the request. Use userFromContext to read the result downstream.
func (h *Handler) optionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractAdminToken(r)
		if token != "" {
			if user, err := h.store.GetUserByToken(token); err == nil {
				r = r.WithContext(context.WithValue(r.Context(), contextKeyUser, user))
			}
		}
		next.ServeHTTP(w, r)
	})
}

// requireAuth rejects unauthenticated requests with 401.
func (h *Handler) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractAdminToken(r)
		if token == "" {
			respondError(w, http.StatusUnauthorized, "authentication required")
			return
		}
		user, err := h.store.GetUserByToken(token)
		if err != nil {
			if err == store.ErrNotFound {
				respondError(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}
			respondError(w, http.StatusInternalServerError, "could not verify token")
			return
		}
		r = r.WithContext(context.WithValue(r.Context(), contextKeyUser, user))
		next.ServeHTTP(w, r)
	})
}

func userFromContext(r *http.Request) *domain.User {
	u, _ := r.Context().Value(contextKeyUser).(*domain.User)
	return u
}
