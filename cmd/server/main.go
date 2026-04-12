package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"github.com/fabianthorsen/openpadel/internal/api"
	"github.com/fabianthorsen/openpadel/internal/email"
	"github.com/fabianthorsen/openpadel/internal/store"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	// Load .env if present (local dev). Silently ignored in production.
	godotenv.Load()
	dbPath := env("DB_PATH", "openpadel.db")
	port := env("PORT", "8080")
	appURL := env("APP_URL", "http://localhost:5173")
	resendKey := env("RESEND_API_KEY", "")
	resendFrom := env("RESEND_FROM", "OpenPadel <noreply@openpadel.app>")
	vapidPrivate := env("VAPID_PRIVATE_KEY", "")
	vapidPublic := env("VAPID_PUBLIC_KEY", "")

	s, err := store.Open(dbPath)
	if err != nil {
		slog.Error("open db", "err", err)
		os.Exit(1)
	}
	defer s.Close()

	emailClient := email.NewClient(resendKey, resendFrom)
	r := api.NewRouter(s, emailClient, appURL, vapidPrivate, vapidPublic)

	slog.Info("listening", "port", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		slog.Error("server", "err", err)
		os.Exit(1)
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
