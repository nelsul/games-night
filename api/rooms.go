package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Room codes avoid letters and digits that are easy to confuse (0/O, 1/I/L).
const (
	codeAlphabet = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"
	codeLength   = 4
)

func newRoomCode() (string, error) {
	b := make([]byte, codeLength)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(codeAlphabet))))
		if err != nil {
			return "", err
		}
		b[i] = codeAlphabet[n.Int64()]
	}
	return string(b), nil
}

// normalizeCode upper-cases a typed code and checks it could be a room code.
func normalizeCode(s string) (string, bool) {
	s = strings.ToUpper(strings.TrimSpace(s))
	if len(s) != codeLength {
		return "", false
	}
	for _, c := range s {
		if !strings.ContainsRune(codeAlphabet, c) {
			return "", false
		}
	}
	return s, true
}

type RoomPlayer struct {
	PublicUser
	IsHost bool `json:"is_host"`
	Online bool `json:"online"`
}

type RoomView struct {
	Code      string          `json:"code"`
	Game      Game            `json:"game"`
	Status    string          `json:"status"`
	HostID    string          `json:"host_id"`
	CreatedAt time.Time       `json:"created_at"`
	Options   json.RawMessage `json:"options"`
	Players   []RoomPlayer    `json:"players"`
	Match     any             `json:"match"` // the game's view for the person asking, or null
}

type roomRow struct {
	ID     string
	Code   string
	GameID string
	Status string
	HostID string
}

// openRoom finds a room that is still in the lobby or being played.
func (a *App) openRoom(ctx context.Context, q pgx.Tx, code string, forUpdate bool) (roomRow, error) {
	sql := `SELECT id, code, game, status, host_id FROM rooms WHERE code = $1 AND status IN ('lobby', 'playing')`
	if forUpdate {
		sql += ` FOR UPDATE`
	}
	var rr roomRow
	var err error
	if q != nil {
		err = q.QueryRow(ctx, sql, code).Scan(&rr.ID, &rr.Code, &rr.GameID, &rr.Status, &rr.HostID)
	} else {
		err = a.db.QueryRow(ctx, sql, code).Scan(&rr.ID, &rr.Code, &rr.GameID, &rr.Status, &rr.HostID)
	}
	return rr, err
}

func (a *App) isMember(ctx context.Context, roomID, userID string) (bool, error) {
	var ok bool
	err := a.db.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM room_players WHERE room_id = $1 AND user_id = $2)`, roomID, userID).Scan(&ok)
	return ok, err
}

func (a *App) roomView(ctx context.Context, roomID string) (RoomView, error) {
	var v RoomView
	var gameID string
	err := a.db.QueryRow(ctx, `SELECT code, game, status, host_id, created_at, options FROM rooms WHERE id = $1`, roomID).
		Scan(&v.Code, &gameID, &v.Status, &v.HostID, &v.CreatedAt, &v.Options)
	if err != nil {
		return v, err
	}
	v.Game, _ = findGame(gameID)
	rows, err := a.db.Query(ctx, `
		SELECT u.id, u.name, u.handle, u.avatar_key FROM room_players p JOIN users u ON u.id = p.user_id
		WHERE p.room_id = $1 ORDER BY p.joined_at, u.id`, roomID)
	if err != nil {
		return v, err
	}
	defer rows.Close()
	online := a.hub.Online(roomID)
	v.Players = []RoomPlayer{}
	for rows.Next() {
		var p RoomPlayer
		var id, name string
		var handle, avatar *string
		if err := rows.Scan(&id, &name, &handle, &avatar); err != nil {
			return v, err
		}
		p.PublicUser = a.publicUser(ctx, id, name, handle, avatar)
		p.IsHost = p.ID == v.HostID
		p.Online = online[p.ID]
		v.Players = append(v.Players, p)
	}
	return v, rows.Err()
}

// matchView is the game's view of the room's current match for one person, or nil.
func (a *App) matchView(ctx context.Context, roomID, gameID, userID string) any {
	e, ok := findEngine(gameID)
	if !ok {
		return nil
	}
	m, err := e.View(ctx, a.db, roomID, userID)
	if err != nil {
		a.log.Error("match view", "room", roomID, "err", err)
		return nil
	}
	return m
}

// roomViewFor is the room plus the match as userID sees it.
func (a *App) roomViewFor(ctx context.Context, roomID, userID string) (RoomView, error) {
	v, err := a.roomView(ctx, roomID)
	if err != nil {
		return v, err
	}
	v.Match = a.matchView(ctx, roomID, v.Game.ID, userID)
	return v, nil
}

// broadcastRoom sends everyone connected to the room its current state, each with
// their own view of the match.
func (a *App) broadcastRoom(roomID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	v, err := a.roomView(ctx, roomID)
	if err != nil {
		a.log.Error("room view", "room", roomID, "err", err)
		return
	}
	a.hub.SendEach(roomID, func(userID string) any {
		mine := v
		mine.Match = a.matchView(ctx, roomID, v.Game.ID, userID)
		return map[string]any{"type": "room", "room": mine}
	})
}

// notifyUserRooms refreshes every open room the user is in (after a name or photo change).
func (a *App) notifyUserRooms(ctx context.Context, userID string) {
	rows, err := a.db.Query(ctx, `
		SELECT r.id FROM rooms r JOIN room_players p ON p.room_id = r.id
		WHERE p.user_id = $1 AND r.status IN ('lobby', 'playing')`, userID)
	if err != nil {
		return
	}
	var ids []string
	for rows.Next() {
		var id string
		if rows.Scan(&id) == nil {
			ids = append(ids, id)
		}
	}
	rows.Close()
	for _, id := range ids {
		go a.broadcastRoom(id)
	}
}

func (a *App) handleCreateRoom(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Game string `json:"game"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	if _, ok := findGame(in.Game); !ok {
		writeError(w, 400, "Esse jogo não existe.")
		return
	}
	ctx := r.Context()
	uid := currentUser(ctx)
	var roomID string
	for attempt := 0; ; attempt++ {
		code, err := newRoomCode()
		if err != nil {
			a.serverError(w, r, err)
			return
		}
		roomID, err = a.createRoom(ctx, code, in.Game, uid)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" && attempt < 10 {
			continue // that code is taken by another open room: draw again
		}
		if err != nil {
			a.serverError(w, r, err)
			return
		}
		break
	}
	v, err := a.roomView(ctx, roomID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	writeJSON(w, 201, v)
}

func (a *App) createRoom(ctx context.Context, code, game, hostID string) (string, error) {
	tx, err := a.db.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)
	var id string
	if err := tx.QueryRow(ctx, `INSERT INTO rooms (code, game, host_id) VALUES ($1, $2, $3) RETURNING id`, code, game, hostID).Scan(&id); err != nil {
		return "", err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO room_players (room_id, user_id) VALUES ($1, $2)`, id, hostID); err != nil {
		return "", err
	}
	return id, tx.Commit(ctx)
}

type RoomSummary struct {
	Code      string    `json:"code"`
	Game      Game      `json:"game"`
	Status    string    `json:"status"`
	IsHost    bool      `json:"is_host"`
	Players   int       `json:"players"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (a *App) handleListRooms(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uid := currentUser(ctx)
	rows, err := a.db.Query(ctx, `
		SELECT r.code, r.game, r.status, r.host_id = $1, r.updated_at,
			(SELECT count(*) FROM room_players x WHERE x.room_id = r.id)
		FROM rooms r JOIN room_players p ON p.room_id = r.id AND p.user_id = $1
		WHERE r.status IN ('lobby', 'playing')
		ORDER BY r.updated_at DESC LIMIT 20`, uid)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer rows.Close()
	out := []RoomSummary{}
	for rows.Next() {
		var s RoomSummary
		var gameID string
		if err := rows.Scan(&s.Code, &gameID, &s.Status, &s.IsHost, &s.UpdatedAt, &s.Players); err != nil {
			a.serverError(w, r, err)
			return
		}
		s.Game, _ = findGame(gameID)
		out = append(out, s)
	}
	writeJSON(w, 200, out)
}

// roomFromPath reads {code} and loads the open room, writing a 404 when there is none.
func (a *App) roomFromPath(w http.ResponseWriter, r *http.Request) (roomRow, bool) {
	code, ok := normalizeCode(chi.URLParam(r, "code"))
	if !ok {
		writeError(w, 404, "Não encontramos uma sala com esse código.")
		return roomRow{}, false
	}
	rr, err := a.openRoom(r.Context(), nil, code, false)
	if isNoRows(err) {
		writeError(w, 404, "Não encontramos uma sala com esse código.")
		return rr, false
	} else if err != nil {
		a.serverError(w, r, err)
		return rr, false
	}
	return rr, true
}

func (a *App) handleGetRoom(w http.ResponseWriter, r *http.Request) {
	rr, ok := a.roomFromPath(w, r)
	if !ok {
		return
	}
	v, err := a.roomViewFor(r.Context(), rr.ID, currentUser(r.Context()))
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	writeJSON(w, 200, v)
}

// handleJoinRoom adds the user to the room. Joining a room you are already in is fine.
func (a *App) handleJoinRoom(w http.ResponseWriter, r *http.Request) {
	code, ok := normalizeCode(chi.URLParam(r, "code"))
	if !ok {
		writeError(w, 404, "Não encontramos uma sala com esse código.")
		return
	}
	ctx := r.Context()
	uid := currentUser(ctx)
	tx, err := a.db.Begin(ctx)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback(ctx)
	rr, err := a.openRoom(ctx, tx, code, true)
	if isNoRows(err) {
		writeError(w, 404, "Não encontramos uma sala com esse código.")
		return
	} else if err != nil {
		a.serverError(w, r, err)
		return
	}
	var member bool
	var count int
	if err := tx.QueryRow(ctx, `SELECT COALESCE(bool_or(user_id = $2), false), count(*) FROM room_players WHERE room_id = $1`, rr.ID, uid).Scan(&member, &count); err != nil {
		a.serverError(w, r, err)
		return
	}
	if !member {
		game, _ := findGame(rr.GameID)
		if rr.Status != "lobby" {
			writeError(w, 409, "A partida já começou nessa sala.")
			return
		}
		if count >= game.MaxPlayers {
			writeError(w, 409, "A sala está cheia.")
			return
		}
		if _, err := tx.Exec(ctx, `INSERT INTO room_players (room_id, user_id) VALUES ($1, $2)`, rr.ID, uid); err != nil {
			a.serverError(w, r, err)
			return
		}
		if _, err := tx.Exec(ctx, `UPDATE rooms SET updated_at = now() WHERE id = $1`, rr.ID); err != nil {
			a.serverError(w, r, err)
			return
		}
		// Joining answers any invite to this room.
		if _, err := tx.Exec(ctx, `UPDATE room_invites SET status = 'accepted' WHERE room_id = $1 AND to_id = $2 AND status = 'pending'`, rr.ID, uid); err != nil {
			a.serverError(w, r, err)
			return
		}
	}
	if err := tx.Commit(ctx); err != nil {
		a.serverError(w, r, err)
		return
	}
	if !member {
		go a.broadcastRoom(rr.ID)
	}
	v, err := a.roomViewFor(ctx, rr.ID, uid)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	writeJSON(w, 200, v)
}

// handleLeaveRoom removes the user. If the host leaves, the next person to have
// joined becomes host; an empty room is closed so its code can be reused.
func (a *App) handleLeaveRoom(w http.ResponseWriter, r *http.Request) {
	rr, ok := a.roomFromPath(w, r)
	if !ok {
		return
	}
	ctx := r.Context()
	uid := currentUser(ctx)
	if rr.Status == "playing" {
		writeError(w, 409, "A partida está rolando. Peça para o host encerrar antes de sair.")
		return
	}
	tx, err := a.db.Begin(ctx)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `SELECT 1 FROM rooms WHERE id = $1 FOR UPDATE`, rr.ID); err != nil {
		a.serverError(w, r, err)
		return
	}
	if _, err := tx.Exec(ctx, `DELETE FROM room_players WHERE room_id = $1 AND user_id = $2`, rr.ID, uid); err != nil {
		a.serverError(w, r, err)
		return
	}
	var next *string
	if err := tx.QueryRow(ctx, `SELECT user_id::text FROM room_players WHERE room_id = $1 ORDER BY joined_at, user_id LIMIT 1`, rr.ID).Scan(&next); err != nil && !isNoRows(err) {
		a.serverError(w, r, err)
		return
	}
	switch {
	case next == nil:
		_, err = tx.Exec(ctx, `UPDATE rooms SET status = 'closed', updated_at = now() WHERE id = $1`, rr.ID)
	case rr.HostID == uid:
		_, err = tx.Exec(ctx, `UPDATE rooms SET host_id = $2, updated_at = now() WHERE id = $1`, rr.ID, *next)
	default:
		_, err = tx.Exec(ctx, `UPDATE rooms SET updated_at = now() WHERE id = $1`, rr.ID)
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := tx.Commit(ctx); err != nil {
		a.serverError(w, r, err)
		return
	}
	a.hub.Kick(rr.ID, uid)
	if next != nil {
		go a.broadcastRoom(rr.ID)
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}
