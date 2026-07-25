// Command seed fills a local dev database with predictable test data so you
// don't have to hand-register users every time. It seeds a fixed roster of
// registered Users (all sharing one password) and a couple of Clubs with mixed
// admin/member rosters, exercising the Club and invite features.
//
// It is idempotent and additive: Users and Clubs that already exist are reused,
// never duplicated, and nothing is ever deleted — so it's safe to re-run. It
// talks to the store directly (real bcrypt hashing, real validation), which is
// why this can't be a plain SQL fixture.
//
// Usage:
//
//	make seed                 # seeds ./openpadel.db (or $DB_PATH)
//	go run ./cmd/seed -db path/to.db
//
// Never point this at a production database.
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"strings"

	"github.com/joho/godotenv"

	"github.com/fabianthorsen/openpadel/internal/domain"
	"github.com/fabianthorsen/openpadel/internal/pairing/americano"
	"github.com/fabianthorsen/openpadel/internal/store"
)

// seedPassword is the shared login for every seeded User — trivial to type when
// logging in as any of them during manual testing.
const seedPassword = "password123"

// seedUser is one registered account to ensure exists. Key is a short handle
// used to wire Club rosters below, not persisted anywhere.
type seedUser struct {
	Key         string
	DisplayName string
	SelfRating  int
}

// seedClub is one Club to ensure exists, owned by Admin (a user Key) with the
// listed Members (user Keys) on its roster.
type seedClub struct {
	Name        string
	Description string
	AvatarIcon  string
	AvatarColor string
	Admin       string
	Members     []string
}

// email derives the deterministic login for a seeded user handle, e.g.
// "alice" -> "alice@openpadel.local".
func email(key string) string { return key + "@openpadel.local" }

var seedUsers = []seedUser{
	{"alice", "Alice", 5},
	{"bob", "Bob", 4},
	{"carol", "Carol", 3},
	{"dave", "Dave", 2},
	{"erik", "Erik", 4},
	{"fiona", "Fiona", 3},
	{"grace", "Grace", 5},
	{"henry", "Henry", 1},
}

var seedClubs = []seedClub{
	{
		Name:        "Bouvet Padel",
		Description: "The original test club.",
		AvatarIcon:  "Trophy",
		AvatarColor: "forest",
		Admin:       "alice",
		Members:     []string{"bob", "carol", "dave"},
	},
	{
		Name:        "Oslo Smashers",
		Description: "A second club for cross-club testing.",
		AvatarIcon:  "Swords",
		AvatarColor: "forest",
		Admin:       "erik",
		Members:     []string{"fiona", "grace"},
	},
}

func main() {
	_ = godotenv.Load()
	dbPath := flag.String("db", env("DB_PATH", "openpadel.db"), "database path")
	flag.Parse()

	s, err := store.Open(*dbPath)
	if err != nil {
		log.Fatalf("open db %q: %v", *dbPath, err)
	}
	defer func() { _ = s.Close() }()

	// Ensure every user exists, remembering each handle's user id for wiring
	// Club rosters. Existing accounts are reused, not recreated.
	ids := make(map[string]string, len(seedUsers))
	created := 0
	for _, u := range seedUsers {
		id, isNew, err := ensureUser(s, u)
		if err != nil {
			log.Fatalf("seed user %s: %v", u.Key, err)
		}
		ids[u.Key] = id
		if isNew {
			created++
		}
	}
	fmt.Printf("users: %d present (%d created, %d already existed)\n", len(seedUsers), created, len(seedUsers)-created)

	for _, c := range seedClubs {
		if err := ensureClub(s, c, ids); err != nil {
			log.Fatalf("seed club %q: %v", c.Name, err)
		}
	}

	if err := ensureHistory(s, ids); err != nil {
		log.Fatalf("seed history: %v", err)
	}

	fmt.Print(summary())
}

// ensureUser creates the user if no account with its email exists yet, and
// returns the user id either way. isNew reports whether a row was created.
func ensureUser(s *store.Store, u seedUser) (id string, isNew bool, err error) {
	existing, err := s.GetUserByEmail(email(u.Key))
	if err == nil {
		return existing.ID, false, nil
	}
	if !errors.Is(err, store.ErrNotFound) {
		return "", false, err
	}
	user, err := s.CreateUser(email(u.Key), u.DisplayName, seedPassword, u.SelfRating)
	if err != nil {
		return "", false, err
	}
	return user.ID, true, nil
}

// ensureClub creates the Club (owned by its Admin) if the Admin doesn't already
// own one by that name, then makes sure every listed Member is on the roster.
// Both steps tolerate pre-existing state, so re-running never duplicates.
func ensureClub(s *store.Store, c seedClub, ids map[string]string) error {
	adminID, ok := ids[c.Admin]
	if !ok {
		return fmt.Errorf("unknown admin handle %q", c.Admin)
	}

	clubID, err := findClubByName(s, adminID, c.Name)
	if err != nil {
		return err
	}
	if clubID == "" {
		club, err := s.CreateClub(c.Name, c.Description, c.AvatarIcon, c.AvatarColor, adminID)
		if err != nil {
			return err
		}
		clubID = club.ID
		fmt.Printf("club %q created (admin %s)\n", c.Name, c.Admin)
	} else {
		fmt.Printf("club %q already existed\n", c.Name)
	}

	for _, key := range c.Members {
		memberID, ok := ids[key]
		if !ok {
			return fmt.Errorf("unknown member handle %q", key)
		}
		if err := s.JoinClub(clubID, memberID); err != nil && !errors.Is(err, store.ErrAlreadyMember) {
			return fmt.Errorf("add member %s: %w", key, err)
		}
	}
	return nil
}

// findClubByName returns the id of a Club named name that adminID owns/belongs
// to, or "" if none. Clubs have no name-unique constraint, so this scopes the
// idempotency check to the seeding admin's own clubs.
func findClubByName(s *store.Store, adminID, name string) (string, error) {
	clubs, err := s.GetUserClubs(adminID)
	if err != nil {
		return "", err
	}
	for _, c := range clubs {
		if c.Name == name {
			return c.ID, nil
		}
	}
	return "", nil
}

// historyLineups are the four-player rosters for the seeded past tournaments —
// varied overlapping subsets so several users accrue history with different
// finishing ranks. Each is one Americano on one court (exactly 4 players).
var historyLineups = [][]string{
	{"alice", "bob", "carol", "dave"},
	{"erik", "fiona", "grace", "henry"},
	{"alice", "erik", "carol", "grace"},
	{"bob", "dave", "fiona", "henry"},
	{"alice", "fiona", "bob", "grace"},
}

// ensureHistory plays out a handful of completed tournaments so seeded users have
// non-empty history and career stats to look at. It drives the real store paths
// (pairing → save rounds → start → score every match → complete), so the
// leaderboard and placement stats compute exactly as in production. Idempotent:
// if the first user already has history, it assumes the set is present and skips.
func ensureHistory(s *store.Store, ids map[string]string) error {
	hist, err := s.GetTournamentHistory(ids["alice"])
	if err != nil {
		return err
	}
	if len(hist) > 0 {
		fmt.Println("history: already present, skipping")
		return nil
	}

	for i, lineup := range historyLineups {
		if err := seedTournament(s, ids, lineup); err != nil {
			return fmt.Errorf("tournament %d: %w", i+1, err)
		}
	}
	fmt.Printf("history: seeded %d completed tournaments\n", len(historyLineups))
	return nil
}

// seedTournament runs one Americano to completion with random scores.
func seedTournament(s *store.Store, ids map[string]string, lineup []string) error {
	const points = 24
	sess, err := s.CreateSession(domain.SessionInput{
		Courts:   1,
		Points:   points,
		GameMode: domain.ModeAmericano,
	}, ids[lineup[0]])
	if err != nil {
		return err
	}

	players := make([]domain.Player, 0, len(lineup))
	for _, key := range lineup {
		p, err := s.CreatePlayer(sess.ID, titleCase(key), ids[key], false)
		if err != nil {
			return err
		}
		players = append(players, *p)
	}

	totalRounds := americano.TotalRounds(len(players), 1)
	if err := s.SaveRounds(sess.ID, americano.GenerateRounds(players, 1, totalRounds)); err != nil {
		return err
	}
	if err := s.StartSession(sess.ID, totalRounds, nil); err != nil {
		return err
	}

	rounds, err := s.GetRounds(sess.ID)
	if err != nil {
		return err
	}
	for _, rd := range rounds {
		for _, m := range rd.Matches {
			a := rand.IntN(points + 1) // random split summing to points
			if _, err := s.UpdateScore(m.ID, a, points-a); err != nil {
				return err
			}
		}
	}
	return s.CompleteSession(sess.ID, false)
}

// titleCase upper-cases the first letter of a handle ("alice" -> "Alice"),
// matching the seeded display names.
func titleCase(k string) string {
	if k == "" {
		return k
	}
	return strings.ToUpper(k[:1]) + k[1:]
}

func summary() string {
	out := "\nseed complete. shared password: " + seedPassword + "\nlogins:\n"
	for _, u := range seedUsers {
		out += fmt.Sprintf("  %-24s (%s)\n", email(u.Key), u.DisplayName)
	}
	return out
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
