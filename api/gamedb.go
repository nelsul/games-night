package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nelsul/games-night/api/games"
)

// Each game gets its own Postgres role, game_<id>, which owns the game's schema (<id>)
// and nothing else. The hub switches to that role around every call into a game, and
// runs the game's migrations as it, so the database itself keeps games apart: a game
// can't read or change the hub's tables (only point foreign keys at rooms.id and
// users.id) or any other game's schema.

// setupGameDB creates the game's role and schema (idempotent) and hands them over.
func setupGameDB(ctx context.Context, db *pgxpool.Pool, e games.Engine) error {
	id := e.Info().ID
	role := pgx.Identifier{games.Role(id)}.Sanitize()
	schema := pgx.Identifier{id}.Sanitize()
	return pgx.BeginFunc(ctx, db, func(tx pgx.Tx) error {
		var exists bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = $1)`, games.Role(id)).Scan(&exists); err != nil {
			return err
		}
		stmts := []string{}
		if !exists {
			stmts = append(stmts, `CREATE ROLE `+role+` NOLOGIN`)
		}
		stmts = append(stmts,
			`GRANT `+role+` TO CURRENT_USER`,
			`CREATE SCHEMA IF NOT EXISTS `+schema+` AUTHORIZATION `+role,
			`ALTER SCHEMA `+schema+` OWNER TO `+role,
			`REVOKE ALL ON SCHEMA `+schema+` FROM PUBLIC`,
			`REVOKE CREATE ON SCHEMA public FROM PUBLIC`,
			`GRANT USAGE ON SCHEMA public TO `+role,
			`GRANT REFERENCES (id) ON public.rooms, public.users TO `+role,
		)
		for _, sql := range stmts {
			if _, err := tx.Exec(ctx, sql); err != nil {
				return fmt.Errorf("%s: %w", sql, err)
			}
		}
		// Tables made before roles existed (or by hand) move to the game's role too.
		rows, err := tx.Query(ctx, `
			SELECT c.relname FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace
			WHERE n.nspname = $1 AND c.relkind IN ('r', 'p', 'S', 'v', 'm')`, id)
		if err != nil {
			return err
		}
		names, err := pgx.CollectRows(rows, pgx.RowTo[string])
		if err != nil {
			return err
		}
		for _, name := range names {
			if _, err := tx.Exec(ctx, `ALTER TABLE `+pgx.Identifier{id, name}.Sanitize()+` OWNER TO `+role); err != nil {
				return err
			}
		}
		return nil
	})
}

// asGame runs fn inside tx as the game's role, then switches back to the hub.
func asGame(ctx context.Context, tx pgx.Tx, e games.Engine, fn func() error) error {
	if _, err := tx.Exec(ctx, `SET LOCAL ROLE `+pgx.Identifier{games.Role(e.Info().ID)}.Sanitize()); err != nil {
		return err
	}
	if err := fn(); err != nil {
		return err
	}
	_, err := tx.Exec(ctx, `RESET ROLE`)
	return err
}

// gameView reads a player's view of the match in a read-only transaction, as the game.
func (a *App) gameView(ctx context.Context, e games.Engine, roomID, userID string) (any, error) {
	var out any
	err := pgx.BeginTxFunc(ctx, a.db, pgx.TxOptions{AccessMode: pgx.ReadOnly}, func(tx pgx.Tx) error {
		return asGame(ctx, tx, e, func() error {
			var err error
			out, err = e.View(ctx, tx, roomID, userID)
			return err
		})
	})
	return out, err
}
