package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/fabianthorsen/openpadel/internal/email"
	"github.com/fabianthorsen/openpadel/internal/events"
	"github.com/fabianthorsen/openpadel/internal/livescores"
	"github.com/fabianthorsen/openpadel/internal/store"
	"github.com/fabianthorsen/openpadel/internal/ui"
)

type Handler struct {
	store        *store.Store
	live         *livescores.Store
	hub          *events.Hub
	email        *email.Client
	appURL       string
	vapidPrivate string
	vapidPublic  string
}

func NewRouter(s *store.Store, emailClient *email.Client, appURL, vapidPrivate, vapidPublic string) http.Handler {
	h := &Handler{
		store:        s,
		live:         livescores.New(),
		hub:          events.NewHub(),
		email:        emailClient,
		appURL:       appURL,
		vapidPrivate: vapidPrivate,
		vapidPublic:  vapidPublic,
	}
	r := chi.NewRouter()

	r.Use(requestLogger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
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
			desc = creator + " wants you to join this Padel tournament on OpenPadel."
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
		// Auth
		r.Post("/auth/register", h.register)
		r.Post("/auth/login", h.login)
		r.Post("/auth/logout", h.logout)
		r.With(h.requireAuth).Get("/auth/me", h.me)
		r.With(h.requireAuth).Get("/auth/profile", h.profile)
		r.With(h.requireAuth).Put("/auth/profile", h.updateProfile)
		r.With(h.requireAuth).Get("/auth/history", h.history)
		r.With(h.requireAuth).Delete("/auth/account", h.deleteAccount)
		r.Post("/auth/forgot", h.forgotPassword)
		r.Post("/auth/reset", h.resetPassword)

		// Contacts
		r.With(h.requireAuth).Get("/contacts", h.getContacts)
		r.With(h.requireAuth).Post("/contacts", h.addContact)
		r.With(h.requireAuth).Delete("/contacts/{contactID}", h.removeContact)
		r.With(h.requireAuth).Get("/users/search", h.searchUsers)

		// Invites
		r.With(h.requireAuth).Get("/invites", h.getMyInvites)
		r.With(h.requireAuth).Post("/invites/{inviteID}/accept", h.acceptInvite)
		r.With(h.requireAuth).Post("/invites/{inviteID}/decline", h.declineInvite)

		r.Get("/push/vapid-public-key", h.vapidPublicKey)
		r.With(h.requireAuth).Post("/push/subscribe", h.subscribePush)
		r.With(h.requireAuth).Delete("/push/subscribe", h.unsubscribePush)

		r.Post("/sessions", h.createSession)

		r.Route("/sessions/{id}", func(r chi.Router) {
			r.Get("/", h.getSession)
			r.Get("/events", h.hub.ServeSSE())
			r.Delete("/", h.cancelSession)
			r.Post("/close", h.closeSession)
r.Post("/start", h.startSession)
			r.With(h.optionalAuth).Post("/players", h.joinSession)
			r.Get("/invites", h.getSessionInvites)
			r.With(h.requireAuth).Post("/invites", h.sendInvite)
			r.Delete("/players/{playerID}", h.deactivatePlayer)
			r.Get("/rounds", h.getRounds)
			r.Get("/rounds/current", h.getCurrentRound)
			r.Post("/rounds/advance", h.advanceRound)
			r.Get("/leaderboard", h.getLeaderboard)
			r.Put("/matches/{matchID}/score", h.submitScore)
			r.Patch("/matches/{matchID}/score", h.updateLiveScore)
			// Tennis
			r.Put("/tennis/teams", h.setTennisTeams)
			r.Get("/tennis/match", h.getTennisMatch)
			r.Post("/tennis/point/{team}", h.addTennisPoint)
			r.Post("/tennis/server/{team}", h.setTennisServer)
		})
	})

	return r
}
