// Package games is the seam between the hub and each game. The hub (package main)
// only knows this interface; every game lives in its own package and its own
// Postgres schema, and never reads the hub's tables beyond foreign keys.
package games

import (
	"context"
	"encoding/json"
	"io/fs"
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
	// Schema is the game's own Postgres schema; Migrations holds its *.sql files,
	// applied in name order and tracked in <schema>.schema_migrations.
	Schema() string
	Migrations() fs.FS
	// CheckOptions validates lobby options sent by the host and returns them normalized.
	CheckOptions(raw json.RawMessage) (json.RawMessage, error)
	// Start deals a new match for players (in seat order).
	Start(ctx context.Context, tx pgx.Tx, roomID string, players []string, options json.RawMessage, now time.Time) error
	// Act applies one player's action to the room's current match.
	Act(ctx context.Context, tx pgx.Tx, roomID, userID string, action json.RawMessage, now time.Time) error
	// View is what userID may see of the room's current match; nil when there is none.
	View(ctx context.Context, q Querier, roomID, userID string) (any, error)
	// Stop closes the room's current match (ended early, or the table plays again).
	Stop(ctx context.Context, tx pgx.Tx, roomID string) error
}

// BotPlayer is implemented by games whose matches bots can play. BotStep applies at
// most one action by one of the given bots and reports whether a bot acted; the hub
// calls it again (with a pause, so people can follow) until no bot has anything to do.
type BotPlayer interface {
	BotStep(ctx context.Context, tx pgx.Tx, roomID string, bots []string, now time.Time) (bool, error)
}

// UserError is an error whose message is meant for players (in Portuguese).
type UserError interface {
	error
	UserMessage() string
}
