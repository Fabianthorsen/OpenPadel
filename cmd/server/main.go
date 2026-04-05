package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"github.com/fabianthorsen/nottennis/internal/api"
	"github.com/fabianthorsen/nottennis/internal/email"
	"github.com/fabianthorsen/nottennis/internal/store"
)

func main() {
	// Load .env if present (local dev). Silently ignored in production.
	godotenv.Load()
	dbPath := env("DB_PATH", "nottennis.db")
	port := env("PORT", "8080")
	appURL := env("APP_URL", "http://localhost:5173")
	resendKey := env("RESEND_API_KEY", "")
	resendFrom := env("RESEND_FROM", "NotTennis <noreply@nottennis.app>")

	s, err := store.Open(dbPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer s.Close()

	emailClient := email.NewClient(resendKey, resendFrom)
	r := api.NewRouter(s, emailClient, appURL)

	log.Printf("listening on :%s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("server: %v", err)
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
