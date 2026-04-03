package api

import (
	"net/http"
	"strings"
)

type contextKey string

const keyAdminToken contextKey = "admin_token"

// adminOnly checks the Authorization header against the session's admin token.
// It does NOT block the request — it sets a flag that handlers can check.
// This way public endpoints and admin endpoints share the same routes.
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
