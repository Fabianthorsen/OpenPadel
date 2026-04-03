package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/fabianthorsen/nottennis/internal/store"
)

type Handler struct {
	store *store.Store
}

func NewRouter(s *store.Store) http.Handler {
	h := &Handler{store: s}
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type"},
	}))

	r.Route("/api", func(r chi.Router) {
		r.Post("/sessions", h.createSession)

		r.Route("/sessions/{id}", func(r chi.Router) {
			r.Get("/", h.getSession)
			r.Post("/start", h.startSession)
			r.Post("/players", h.joinSession)
			r.Delete("/players/{playerID}", h.deactivatePlayer)
			r.Get("/rounds", h.getRounds)
			r.Get("/rounds/current", h.getCurrentRound)
			r.Get("/leaderboard", h.getLeaderboard)
			r.Put("/matches/{matchID}/score", h.submitScore)
		})
	})

	return r
}
