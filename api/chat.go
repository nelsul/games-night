package main

import (
	"net/http"
	"strings"
	"time"
	"unicode/utf8"
)

// The room chat: everyone in the room, in the lobby and during matches. Messages go
// out over the room's WebSocket as {"type":"chat","message":…}; the history endpoint
// fills the gap after a reconnect.

const (
	chatHistory  = 100
	chatMaxRunes = 300
	chatCooldown = 700 * time.Millisecond
)

type ChatMessage struct {
	ID        int64      `json:"id"`
	User      PublicUser `json:"user"`
	Body      string     `json:"body"`
	CreatedAt time.Time  `json:"created_at"`
}

// cleanMessage trims a message and squeezes runs of blank lines.
func cleanMessage(s string) (string, bool) {
	s = strings.TrimSpace(strings.ReplaceAll(s, "\r\n", "\n"))
	for strings.Contains(s, "\n\n\n") {
		s = strings.ReplaceAll(s, "\n\n\n", "\n\n")
	}
	n := utf8.RuneCountInString(s)
	return s, n >= 1 && n <= chatMaxRunes
}

func (a *App) handleListMessages(w http.ResponseWriter, r *http.Request) {
	rr, ok := a.memberRoom(w, r)
	if !ok {
		return
	}
	ctx := r.Context()
	rows, err := a.db.Query(ctx, `
		SELECT * FROM (
			SELECT m.id, m.body, m.created_at, u.id, u.name, u.handle, u.avatar_key
			FROM room_messages m JOIN users u ON u.id = m.user_id
			WHERE m.room_id = $1 ORDER BY m.id DESC LIMIT $2
		) t ORDER BY 1`, rr.ID, chatHistory)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer rows.Close()
	out := []ChatMessage{}
	for rows.Next() {
		var m ChatMessage
		var uid, name string
		var handle, avatar *string
		if err := rows.Scan(&m.ID, &m.Body, &m.CreatedAt, &uid, &name, &handle, &avatar); err != nil {
			a.serverError(w, r, err)
			return
		}
		m.User = a.publicUser(ctx, uid, name, handle, avatar)
		out = append(out, m)
	}
	writeJSON(w, 200, out)
}

func (a *App) handleSendMessage(w http.ResponseWriter, r *http.Request) {
	rr, ok := a.memberRoom(w, r)
	if !ok {
		return
	}
	var in struct {
		Body string `json:"body"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	body, ok := cleanMessage(in.Body)
	if !ok {
		writeError(w, 400, "Escreva uma mensagem de até 300 caracteres.")
		return
	}
	ctx := r.Context()
	uid := currentUser(ctx)
	var last *time.Time
	if err := a.db.QueryRow(ctx, `SELECT max(created_at) FROM room_messages WHERE room_id = $1 AND user_id = $2`, rr.ID, uid).Scan(&last); err != nil {
		a.serverError(w, r, err)
		return
	}
	if last != nil && time.Since(*last) < chatCooldown {
		writeError(w, 429, "Calma! Espere um instante entre as mensagens.")
		return
	}
	var m ChatMessage
	if err := a.db.QueryRow(ctx, `INSERT INTO room_messages (room_id, user_id, body) VALUES ($1, $2, $3) RETURNING id, created_at`,
		rr.ID, uid, body).Scan(&m.ID, &m.CreatedAt); err != nil {
		a.serverError(w, r, err)
		return
	}
	me, err := a.loadMe(ctx, uid)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	m.User, m.Body = me.PublicUser, body
	a.hub.Send(rr.ID, map[string]any{"type": "chat", "message": m})
	writeJSON(w, 201, m)
}
