package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/nelsul/games-night/api/games"
)

// The hub's side of a match: it checks who may do what in the room (host, member,
// room status) and hands everything about the rules to the room's game engine.

func readRaw(w http.ResponseWriter, r *http.Request) (json.RawMessage, bool) {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 64<<10))
	if err != nil || !json.Valid(body) {
		writeError(w, 400, "Requisição inválida.")
		return nil, false
	}
	return body, true
}

// gameError shows rule errors to players and logs anything else.
func (a *App) gameError(w http.ResponseWriter, r *http.Request, err error) {
	var ue games.UserError
	if errors.As(err, &ue) {
		writeError(w, 400, ue.UserMessage())
		return
	}
	a.serverError(w, r, err)
}

// hostRoom loads the room, checks the caller is its host and that it's in status.
func (a *App) hostRoom(w http.ResponseWriter, r *http.Request, status string) (roomRow, games.Engine, bool) {
	rr, ok := a.memberRoom(w, r)
	if !ok {
		return rr, nil, false
	}
	if rr.HostID != currentUser(r.Context()) {
		writeError(w, 403, "Só o host pode fazer isso.")
		return rr, nil, false
	}
	if rr.Status != status {
		if status == "lobby" {
			writeError(w, 409, "A partida já começou.")
		} else {
			writeError(w, 409, "Não há partida rolando nessa sala.")
		}
		return rr, nil, false
	}
	e, ok := findEngine(rr.GameID)
	if !ok {
		writeError(w, 409, "Esse jogo não está disponível.")
		return rr, nil, false
	}
	return rr, e, true
}

func (a *App) respondRoom(w http.ResponseWriter, r *http.Request, roomID string) {
	go a.broadcastRoom(roomID)
	v, err := a.roomViewFor(r.Context(), roomID, currentUser(r.Context()))
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	writeJSON(w, 200, v)
}

func (a *App) handleRoomOptions(w http.ResponseWriter, r *http.Request) {
	rr, e, ok := a.hostRoom(w, r, "lobby")
	if !ok {
		return
	}
	raw, ok := readRaw(w, r)
	if !ok {
		return
	}
	opts, err := e.CheckOptions(raw)
	if err != nil {
		a.gameError(w, r, err)
		return
	}
	if _, err := a.db.Exec(r.Context(), `UPDATE rooms SET options = $2, updated_at = now() WHERE id = $1`, rr.ID, opts); err != nil {
		a.serverError(w, r, err)
		return
	}
	a.respondRoom(w, r, rr.ID)
}

func (a *App) handleStartMatch(w http.ResponseWriter, r *http.Request) {
	rr, e, ok := a.hostRoom(w, r, "lobby")
	if !ok {
		return
	}
	ctx := r.Context()
	err := pgx.BeginFunc(ctx, a.db, func(tx pgx.Tx) error {
		var status string
		var options json.RawMessage
		if err := tx.QueryRow(ctx, `SELECT status, options FROM rooms WHERE id = $1 FOR UPDATE`, rr.ID).Scan(&status, &options); err != nil {
			return err
		}
		if status != "lobby" {
			return &startError{"A partida já começou."}
		}
		rows, err := tx.Query(ctx, `SELECT user_id::text FROM room_players WHERE room_id = $1 ORDER BY joined_at, user_id`, rr.ID)
		if err != nil {
			return err
		}
		players, err := pgx.CollectRows(rows, pgx.RowTo[string])
		if err != nil {
			return err
		}
		if err := e.Start(ctx, tx, rr.ID, players, options, time.Now()); err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `UPDATE rooms SET status = 'playing', updated_at = now() WHERE id = $1`, rr.ID)
		return err
	})
	if err != nil {
		a.gameError(w, r, err)
		return
	}
	a.wakeBots(rr.ID)
	a.respondRoom(w, r, rr.ID)
}

type startError struct{ msg string }

func (e *startError) Error() string       { return e.msg }
func (e *startError) UserMessage() string { return e.msg }

func (a *App) handleMatchAction(w http.ResponseWriter, r *http.Request) {
	rr, ok := a.memberRoom(w, r)
	if !ok {
		return
	}
	if rr.Status != "playing" {
		writeError(w, 409, "Não há partida rolando nessa sala.")
		return
	}
	e, ok := findEngine(rr.GameID)
	if !ok {
		writeError(w, 409, "Esse jogo não está disponível.")
		return
	}
	raw, ok := readRaw(w, r)
	if !ok {
		return
	}
	ctx := r.Context()
	err := pgx.BeginFunc(ctx, a.db, func(tx pgx.Tx) error {
		if err := e.Act(ctx, tx, rr.ID, currentUser(ctx), raw, time.Now()); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, `UPDATE rooms SET updated_at = now() WHERE id = $1`, rr.ID)
		return err
	})
	if err != nil {
		a.gameError(w, r, err)
		return
	}
	a.wakeBots(rr.ID)
	a.respondRoom(w, r, rr.ID)
}

// handleStopMatch closes the current match and takes the room back to the lobby:
// "play again" after the end, or ending a match early.
func (a *App) handleStopMatch(w http.ResponseWriter, r *http.Request) {
	rr, e, ok := a.hostRoom(w, r, "playing")
	if !ok {
		return
	}
	ctx := r.Context()
	err := pgx.BeginFunc(ctx, a.db, func(tx pgx.Tx) error {
		if err := e.Stop(ctx, tx, rr.ID); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, `UPDATE rooms SET status = 'lobby', updated_at = now() WHERE id = $1`, rr.ID)
		return err
	})
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	a.respondRoom(w, r, rr.ID)
}
