package main

import (
	"context"
	"math/rand/v2"
	"net/http"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/nelsul/games-night/api/games"
)

// Bots are house players that fill a room. The host adds and removes them in the
// lobby; during a match the room's game decides their moves (games.BotPlayer), and
// the runner below plays them one at a time with a short, human-like pause.

func (a *App) handleAddBots(w http.ResponseWriter, r *http.Request) {
	rr, _, ok := a.hostRoom(w, r, "lobby")
	if !ok {
		return
	}
	var in struct {
		Count int `json:"count"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	game, _ := findGame(rr.GameID)
	ctx := r.Context()
	err := pgx.BeginFunc(ctx, a.db, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `SELECT 1 FROM rooms WHERE id = $1 FOR UPDATE`, rr.ID); err != nil {
			return err
		}
		var count int
		if err := tx.QueryRow(ctx, `SELECT count(*) FROM room_players WHERE room_id = $1`, rr.ID).Scan(&count); err != nil {
			return err
		}
		n := min(max(in.Count, 0), game.MaxPlayers-count)
		if n <= 0 {
			return nil
		}
		_, err := tx.Exec(ctx, `
			INSERT INTO room_players (room_id, user_id)
			SELECT $1, u.id FROM users u
			WHERE u.is_bot AND NOT EXISTS (SELECT 1 FROM room_players p WHERE p.room_id = $1 AND p.user_id = u.id)
			ORDER BY u.created_at, u.name LIMIT $2`, rr.ID, n)
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `UPDATE rooms SET updated_at = now() WHERE id = $1`, rr.ID)
		return err
	})
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	a.respondRoom(w, r, rr.ID)
}

// handleRemoveBot takes one bot ({id}) or, with "all", every bot out of the lobby.
func (a *App) handleRemoveBots(w http.ResponseWriter, r *http.Request) {
	rr, _, ok := a.hostRoom(w, r, "lobby")
	if !ok {
		return
	}
	ctx := r.Context()
	sql := `DELETE FROM room_players p USING users u WHERE p.room_id = $1 AND u.id = p.user_id AND u.is_bot`
	args := []any{rr.ID}
	if id := r.URL.Query().Get("id"); id != "" {
		if !uuidRe.MatchString(id) {
			writeError(w, 404, "Não encontrado.")
			return
		}
		sql += ` AND u.id = $2`
		args = append(args, id)
	}
	if _, err := a.db.Exec(ctx, sql, args...); err != nil {
		a.serverError(w, r, err)
		return
	}
	a.respondRoom(w, r, rr.ID)
}

// botRunner makes sure at most one loop plays the bots of a room at a time.
type botRunner struct {
	mu      sync.Mutex
	running map[string]bool // room id → a loop is playing
	again   map[string]bool // room id → something changed while it was running
}

func newBotRunner() *botRunner {
	return &botRunner{running: map[string]bool{}, again: map[string]bool{}}
}

// wakeBots lets the bots of a room look at the match again (after a start, an action
// or a reconnect). It returns at once; the bots play in the background.
func (a *App) wakeBots(roomID string) {
	a.bots.mu.Lock()
	defer a.bots.mu.Unlock()
	if a.bots.running[roomID] {
		a.bots.again[roomID] = true
		return
	}
	a.bots.running[roomID] = true
	go a.playBots(roomID)
}

func (a *App) playBots(roomID string) {
	for {
		time.Sleep(time.Duration(900+rand.IntN(1100)) * time.Millisecond)
		acted, err := a.botStep(roomID)
		if err != nil {
			a.log.Error("bot step", "room", roomID, "err", err)
		}
		if acted {
			a.broadcastRoom(roomID)
			continue
		}
		a.bots.mu.Lock()
		if a.bots.again[roomID] {
			a.bots.again[roomID] = false
			a.bots.mu.Unlock()
			continue
		}
		delete(a.bots.running, roomID)
		a.bots.mu.Unlock()
		return
	}
}

func (a *App) botStep(roomID string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var gameID, status string
	if err := a.db.QueryRow(ctx, `SELECT game, status FROM rooms WHERE id = $1`, roomID).Scan(&gameID, &status); err != nil {
		return false, err
	}
	e, ok := findEngine(gameID)
	bp, canBots := e.(games.BotPlayer)
	if !ok || !canBots || status != "playing" {
		return false, nil
	}
	rows, err := a.db.Query(ctx, `SELECT p.user_id::text FROM room_players p JOIN users u ON u.id = p.user_id WHERE p.room_id = $1 AND u.is_bot`, roomID)
	if err != nil {
		return false, err
	}
	bots, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil || len(bots) == 0 {
		return false, err
	}
	var acted bool
	err = pgx.BeginFunc(ctx, a.db, func(tx pgx.Tx) error {
		var err error
		acted, err = bp.BotStep(ctx, tx, roomID, bots, time.Now())
		return err
	})
	return acted, err
}
