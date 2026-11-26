package main

import (
	"net/http"
	"time"
)

type InviteOut struct {
	ID        string     `json:"id"`
	Code      string     `json:"code"`
	Game      Game       `json:"game"`
	From      PublicUser `json:"from"`
	Players   int        `json:"players"`
	CreatedAt time.Time  `json:"created_at"`
}

// handleInbox is what the badge and the home screen need: pending friend
// requests and invites to rooms that are still waiting in the lobby.
func (a *App) handleInbox(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	me := currentUser(ctx)
	var requests int
	if err := a.db.QueryRow(ctx, `SELECT count(*) FROM friend_requests WHERE to_id = $1 AND status = 'pending'`, me).Scan(&requests); err != nil {
		a.serverError(w, r, err)
		return
	}
	rows, err := a.db.Query(ctx, `
		SELECT i.id, i.created_at, r.code, r.game, u.id, u.name, u.handle, u.avatar_key,
			(SELECT count(*) FROM room_players p WHERE p.room_id = r.id)
		FROM room_invites i
		JOIN rooms r ON r.id = i.room_id AND r.status = 'lobby'
		JOIN users u ON u.id = i.from_id
		WHERE i.to_id = $1 AND i.status = 'pending'
		  AND NOT EXISTS (SELECT 1 FROM room_players p WHERE p.room_id = r.id AND p.user_id = $1)
		ORDER BY i.created_at DESC LIMIT 10`, me)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer rows.Close()
	invites := []InviteOut{}
	for rows.Next() {
		var in InviteOut
		var gameID, uid, name string
		var handle, avatar *string
		if err := rows.Scan(&in.ID, &in.CreatedAt, &in.Code, &gameID, &uid, &name, &handle, &avatar, &in.Players); err != nil {
			a.serverError(w, r, err)
			return
		}
		in.Game, _ = findGame(gameID)
		in.From = a.publicUser(ctx, uid, name, handle, avatar)
		invites = append(invites, in)
	}
	writeJSON(w, 200, map[string]any{"requests": requests, "invites": invites})
}

type InvitableFriend struct {
	PublicUser
	State string `json:"state"` // in_room, invited or none
}

// handleRoomFriends lists your friends with whether they are in the room or already invited.
func (a *App) handleRoomFriends(w http.ResponseWriter, r *http.Request) {
	rr, ok := a.memberRoom(w, r)
	if !ok {
		return
	}
	ctx := r.Context()
	me := currentUser(ctx)
	rows, err := a.db.Query(ctx, `
		SELECT u.id, u.name, u.handle, u.avatar_key,
			CASE
				WHEN EXISTS (SELECT 1 FROM room_players p WHERE p.room_id = $2 AND p.user_id = u.id) THEN 'in_room'
				WHEN EXISTS (SELECT 1 FROM room_invites i WHERE i.room_id = $2 AND i.to_id = u.id AND i.status = 'pending') THEN 'invited'
				ELSE 'none'
			END
		FROM friendships f
		JOIN users u ON u.id = CASE WHEN f.user_a = $1 THEN f.user_b ELSE f.user_a END
		WHERE f.user_a = $1 OR f.user_b = $1
		ORDER BY lower(u.name), u.handle`, me, rr.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer rows.Close()
	out := []InvitableFriend{}
	for rows.Next() {
		var f InvitableFriend
		var id, name string
		var handle, avatar *string
		if err := rows.Scan(&id, &name, &handle, &avatar, &f.State); err != nil {
			a.serverError(w, r, err)
			return
		}
		f.PublicUser = a.publicUser(ctx, id, name, handle, avatar)
		out = append(out, f)
	}
	writeJSON(w, 200, out)
}

func (a *App) handleInviteToRoom(w http.ResponseWriter, r *http.Request) {
	rr, ok := a.memberRoom(w, r)
	if !ok {
		return
	}
	var in struct {
		UserID string `json:"user_id"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	ctx := r.Context()
	me := currentUser(ctx)
	if !uuidRe.MatchString(in.UserID) || in.UserID == me {
		writeError(w, 400, "Escolha um amigo para convidar.")
		return
	}
	friends, err := a.areFriends(ctx, me, in.UserID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if !friends {
		writeError(w, 403, "Só dá para convidar amigos.")
		return
	}
	if rr.Status != "lobby" {
		writeError(w, 409, "A partida já começou nessa sala.")
		return
	}
	inRoom, err := a.isMember(ctx, rr.ID, in.UserID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if inRoom {
		writeJSON(w, 200, map[string]string{"state": "in_room"})
		return
	}
	if _, err := a.db.Exec(ctx, `INSERT INTO room_invites (room_id, from_id, to_id) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`, rr.ID, me, in.UserID); err != nil {
		a.serverError(w, r, err)
		return
	}
	a.notifyUser(in.UserID)
	writeJSON(w, 200, map[string]string{"state": "invited"})
}

func (a *App) handleDeclineInvite(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	ctx := r.Context()
	if _, err := a.db.Exec(ctx, `UPDATE room_invites SET status = 'declined' WHERE id = $1 AND to_id = $2 AND status = 'pending'`, id, currentUser(ctx)); err != nil {
		a.serverError(w, r, err)
		return
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}

// memberRoom loads the open room in {code} and checks the current user is in it.
func (a *App) memberRoom(w http.ResponseWriter, r *http.Request) (roomRow, bool) {
	rr, ok := a.roomFromPath(w, r)
	if !ok {
		return rr, false
	}
	member, err := a.isMember(r.Context(), rr.ID, currentUser(r.Context()))
	if err != nil {
		a.serverError(w, r, err)
		return rr, false
	}
	if !member {
		writeError(w, 403, "Entre na sala primeiro.")
		return rr, false
	}
	return rr, true
}
