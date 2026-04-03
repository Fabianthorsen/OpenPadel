package api

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/fabianthorsen/nottennis/internal/store"
	"github.com/fabianthorsen/nottennis/internal/ui"
)

type Handler struct {
	store *store.Store
}

func NewRouter(s *store.Store) http.Handler {
	h := &Handler{store: s}
	r := chi.NewRouter()

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("recv: %s %s", r.Method, r.URL.Path)
			next.ServeHTTP(w, r)
		})
	})
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type"},
	}))

	// Session invite pages get Open Graph tags injected so share previews show
	// "<Creator> wants you to join this Padel tournament!"
	r.Get("/s/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		creator := h.store.GetCreatorName(id)
		title := "Join a Padel tournament!"
		desc := "You've been invited to play Padel. Join now!"
		if creator != "" {
			title = creator + " wants you to join!"
			desc = creator + " wants you to join this Padel tournament on NotTennis."
		}
		page := ui.IndexWithOG(title, desc)
		if page == nil {
			page = []byte("frontend not built")
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(page)
	})

	// Serve SvelteKit SPA for all non-API routes
	r.Handle("/*", ui.Handler())

	r.Route("/api", func(r chi.Router) {
		r.Post("/sessions", h.createSession)

		r.Route("/sessions/{id}", func(r chi.Router) {
			r.Get("/", h.getSession)
			r.Delete("/", h.cancelSession)
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
