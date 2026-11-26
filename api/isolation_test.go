package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"math/rand/v2"
	"os"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nelsul/games-night/api/games"
)

// These tests need a real Postgres where the user may create databases and roles
// (TEST_DATABASE_URL, set in CI). Each run uses a fresh, throwaway database.

func testDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	ctx := context.Background()
	admin, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	name := fmt.Sprintf("gn_test_%d", rand.Uint32())
	if _, err := admin.Exec(ctx, `CREATE DATABASE `+name); err != nil {
		t.Fatal(err)
	}
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		t.Fatal(err)
	}
	cfg.ConnConfig.Database = name
	db, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Close()
		_, _ = admin.Exec(context.Background(), `DROP DATABASE `+name+` WITH (FORCE)`)
		admin.Close()
	})
	return db
}

// fakeGame is a game with hand-written migrations, to try what a game may and may not do.
type fakeGame struct {
	id  string
	sql map[string]string
}

func (g fakeGame) Info() games.Info {
	return games.Info{ID: g.id, Name: g.id, MinPlayers: 1, MaxPlayers: 2}
}
func (g fakeGame) Migrations() fs.FS {
	m := fstest.MapFS{}
	for name, sql := range g.sql {
		m[name] = &fstest.MapFile{Data: []byte(sql)}
	}
	return m
}
func (fakeGame) CheckOptions(raw json.RawMessage) (json.RawMessage, error) { return raw, nil }
func (fakeGame) Start(context.Context, pgx.Tx, string, []string, json.RawMessage, time.Time) error {
	return nil
}
func (fakeGame) Act(context.Context, pgx.Tx, string, string, json.RawMessage, time.Time) error {
	return nil
}
func (fakeGame) View(context.Context, games.Querier, string, string) (any, error) { return nil, nil }
func (fakeGame) Over(context.Context, games.Querier, string) (bool, error)        { return false, nil }
func (fakeGame) Stop(context.Context, pgx.Tx, string) error                       { return nil }

func migrateAll(t *testing.T, db *pgxpool.Pool, extra ...games.Engine) map[string]error {
	t.Helper()
	ctx := context.Background()
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn}))
	hub, _ := fs.Sub(migrationsFS, "migrations")
	if err := migrate(ctx, db, log, hub, "public", ""); err != nil {
		t.Fatalf("hub migrations: %v", err)
	}
	out := map[string]error{}
	for _, e := range append(games.All(), extra...) {
		id := e.Info().ID
		if err := setupGameDB(ctx, db, e); err != nil {
			t.Fatalf("set up %s: %v", id, err)
		}
		out[id] = migrate(ctx, db, log, e.Migrations(), id, games.Role(id))
	}
	return out
}

func TestEveryGameMigratesInItsOwnSchema(t *testing.T) {
	db := testDB(t)
	for id, err := range migrateAll(t, db) {
		if err != nil {
			t.Errorf("%s: %v", id, err)
		}
	}
}

func TestGamesCannotTouchTheHubOrEachOther(t *testing.T) {
	db := testDB(t)
	results := migrateAll(t, db,
		fakeGame{id: "goodgame", sql: map[string]string{
			"001.sql": `CREATE TABLE goodgame.things (id int PRIMARY KEY, room_id uuid REFERENCES public.rooms(id) ON DELETE CASCADE)`,
		}},
		fakeGame{id: "sneaky", sql: map[string]string{"001.sql": `CREATE TABLE backdoor.stolen (x int)`}},
		fakeGame{id: "vandal", sql: map[string]string{"001.sql": `UPDATE public.users SET name = 'x'`}},
		fakeGame{id: "squatter", sql: map[string]string{"001.sql": `CREATE TABLE public.mine (x int)`}},
		fakeGame{id: "peeker", sql: map[string]string{"001.sql": `CREATE TABLE peeker.emails AS SELECT email FROM public.users`}},
	)
	if results["goodgame"] != nil {
		t.Errorf("a well-behaved game failed to migrate: %v", results["goodgame"])
	}
	for _, id := range []string{"sneaky", "vandal", "squatter", "peeker"} {
		if err := results[id]; err == nil || !strings.Contains(err.Error(), "permission denied") {
			t.Errorf("%s: want permission denied, got %v", id, err)
		}
	}

	// At runtime, a game's queries can't reach the hub's data either.
	ctx := context.Background()
	backdoor, _ := findEngine("backdoor")
	for _, sql := range []string{
		`SELECT email FROM public.users`,
		`SELECT * FROM public.sessions`,
		`DELETE FROM public.rooms`,
		`SELECT * FROM goodgame.things`,
	} {
		err := pgx.BeginFunc(ctx, db, func(tx pgx.Tx) error {
			return asGame(ctx, tx, backdoor, func() error {
				_, err := tx.Exec(ctx, sql)
				return err
			})
		})
		if err == nil || !strings.Contains(err.Error(), "permission denied") {
			t.Errorf("backdoor ran %q: %v", sql, err)
		}
	}
}

// A whole Backdoor match through the hub's role switching: the game's role has
// exactly the rights it needs (its schema plus foreign keys to rooms and users).
func TestBackdoorPlaysUnderItsRole(t *testing.T) {
	db := testDB(t)
	migrateAll(t, db)
	ctx := context.Background()
	var players []string
	for i := range 5 {
		var id string
		if err := db.QueryRow(ctx, `INSERT INTO users (email, name, handle) VALUES ($1, $2, $3) RETURNING id`,
			fmt.Sprintf("p%d@example.test", i), fmt.Sprintf("P%d", i), fmt.Sprintf("p%d_test", i)).Scan(&id); err != nil {
			t.Fatal(err)
		}
		players = append(players, id)
	}
	var room string
	if err := db.QueryRow(ctx, `INSERT INTO rooms (code, game, host_id) VALUES ('TEST', 'backdoor', $1) RETURNING id`, players[0]).Scan(&room); err != nil {
		t.Fatal(err)
	}
	e, _ := findEngine("backdoor")
	run := func(fn func(tx pgx.Tx) error) {
		t.Helper()
		if err := pgx.BeginFunc(ctx, db, func(tx pgx.Tx) error { return asGame(ctx, tx, e, func() error { return fn(tx) }) }); err != nil {
			t.Fatal(err)
		}
	}
	run(func(tx pgx.Tx) error { return e.Start(ctx, tx, room, players, json.RawMessage(`{}`), time.Now()) })
	for _, p := range players {
		run(func(tx pgx.Tx) error { return e.Act(ctx, tx, room, p, json.RawMessage(`{"type":"ready"}`), time.Now()) })
	}
	app := &App{db: db, hub: NewHub(), log: slog.Default()}
	v, err := app.gameView(ctx, e, room, players[0])
	if err != nil || v == nil {
		t.Fatalf("view: %v %v", v, err)
	}
	// Deleting the room (as the hub) cascades into the game's tables.
	if _, err := db.Exec(ctx, `DELETE FROM rooms WHERE id = $1`, room); err != nil {
		t.Fatalf("delete room: %v", err)
	}
}
