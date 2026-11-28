// Package games is the seam between the hub and each game. The hub (package main)
// only knows this interface; every game lives in its own package and its own
// Postgres schema, and never reads the hub's tables beyond foreign keys.
package games

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"regexp"
	"time"

	"github.com/jackc/pgx/v5"
)

// Info is what the hub shows about a game.
type Info struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	MinPlayers int    `json:"min_players"`
	MaxPlayers int    `json:"max_players"`
	Available  bool   `json:"available"`
}

// Querier is satisfied by both a pool and a transaction.
type Querier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type Engine interface {
	Info() Info
	// Migrations holds the game's *.sql files. They run, in name order, as the game's
	// own database role (Role) inside its own schema, named after the game's id, and
	// are tracked in <id>.schema_migrations. The role can't touch the hub's tables
	// (beyond pointing at rooms(id) and users(id)) or any other game's schema.
	Migrations() fs.FS
	// CheckOptions validates lobby options sent by the host and returns them normalized.
	CheckOptions(raw json.RawMessage) (json.RawMessage, error)
	// Start deals a new match for players (in seat order).
	Start(ctx context.Context, tx pgx.Tx, roomID string, players []string, options json.RawMessage, now time.Time) error
	// Act applies one player's action to the room's current match.
	Act(ctx context.Context, tx pgx.Tx, roomID, userID string, action json.RawMessage, now time.Time) error
	// View is what userID may see of the room's current match; nil when there is none.
	View(ctx context.Context, q Querier, roomID, userID string) (any, error)
	// Over reports whether the room's current match has ended.
	Over(ctx context.Context, q Querier, roomID string) (bool, error)
	// Stop closes the room's current match (ended early, or the table plays again).
	Stop(ctx context.Context, tx pgx.Tx, roomID string) error
}

type hostKey struct{}

// WithHost marks ctx with the room's host. The hub sets it before calling Act, so a game
// can keep some actions (like fixing the score) to the host.
func WithHost(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, hostKey{}, userID)
}

// IsHost reports whether userID hosts the room the action came from.
func IsHost(ctx context.Context, userID string) bool {
	host, _ := ctx.Value(hostKey{}).(string)
	return host != "" && host == userID
}

// BotPlayer is implemented by games whose matches bots can play. BotStep applies at
// most one action by one of the given bots and reports whether a bot acted; the hub
// calls it again (with a pause, so people can follow) until no bot has anything to do.
type BotPlayer interface {
	BotStep(ctx context.Context, tx pgx.Tx, roomID string, bots []string, now time.Time) (bool, error)
}

// idRe is what a game id may look like; it is also its schema name.
var idRe = regexp.MustCompile(`^[a-z][a-z0-9_]{1,30}$`)

var registry []Engine

// Register adds a game to the hub. Each game calls it from its package's init(), and
// is pulled into the build by a one-line file in api/games/all/<id>.go.
func Register(e Engine) {
	id := e.Info().ID
	if !idRe.MatchString(id) || id == "public" || id == "pg_catalog" {
		panic(fmt.Sprintf("games: invalid game id %q", id))
	}
	for _, other := range registry {
		if other.Info().ID == id {
			panic(fmt.Sprintf("games: %q registered twice", id))
		}
	}
	registry = append(registry, e)
}

// All is every registered game, in registration order.
func All() []Engine { return registry }

// Role is the database role a game's migrations and queries run as.
func Role(id string) string { return "game_" + id }

// UserError is an error whose message is meant for players (in Portuguese).
type UserError interface {
	error
	UserMessage() string
}
