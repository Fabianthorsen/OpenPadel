package main

import (
	"log"
	"net/http"
	"os"

	"github.com/fabianthorsen/nottennis/internal/api"
	"github.com/fabianthorsen/nottennis/internal/store"
)

func main() {
	dbPath := env("DB_PATH", "nottennis.db")
	port := env("PORT", "8080")

	s, err := store.Open(dbPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer s.Close()

	r := api.NewRouter(s)

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
