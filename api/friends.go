package main

import (
	"context"
	"net/http"
	"regexp"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

var uuidRe = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// pathID returns a validated UUID URL parameter, or writes a 404.
func pathID(w http.ResponseWriter, r *http.Request, name string) (string, bool) {
	id := chi.URLParam(r, name)
	if !uuidRe.MatchString(id) {
		writeError(w, 404, "Não encontrado.")
		return "", false
	}
	return id, true
}

// orderedPair returns the two ids sorted, as stored in friendships.
func orderedPair(a, b string) (string, string) {
	if a < b {
		return a, b
	}
	return b, a
}

func (a *App) areFriends(ctx context.Context, x, y string) (bool, error) {
	u1, u2 := orderedPair(x, y)
	var ok bool
	err := a.db.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM friendships WHERE user_a = $1 AND user_b = $2)`, u1, u2).Scan(&ok)
	return ok, err
}

// relation describes how the current user relates to another one.
func (a *App) relation(ctx context.Context, me, other string) (string, error) {
	if me == other {
		return "self", nil
	}
	friends, err := a.areFriends(ctx, me, other)
	if err != nil || friends {
		return "friend", err
	}
	var rel string
	err = a.db.QueryRow(ctx, `
		SELECT CASE WHEN from_id = $1 THEN 'request_sent' ELSE 'request_received' END
		FROM friend_requests WHERE status = 'pending'
		  AND ((from_id = $1 AND to_id = $2) OR (from_id = $2 AND to_id = $1))
		LIMIT 1`, me, other).Scan(&rel)
	if isNoRows(err) {
		return "none", nil
	}
	return rel, err
}

func (a *App) handleListFriends(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	me := currentUser(ctx)
	rows, err := a.db.Query(ctx, `
		SELECT u.id, u.name, u.handle, u.avatar_key
		FROM friendships f
		JOIN users u ON u.id = CASE WHEN f.user_a = $1 THEN f.user_b ELSE f.user_a END
		WHERE f.user_a = $1 OR f.user_b = $1
		ORDER BY lower(u.name), u.handle`, me)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer rows.Close()
	out := []PublicUser{}
	for rows.Next() {
		var id, name string
		var handle, avatar *string
		if err := rows.Scan(&id, &name, &handle, &avatar); err != nil {
			a.serverError(w, r, err)
			return
		}
		out = append(out, a.publicUser(ctx, id, name, handle, avatar))
	}
	writeJSON(w, 200, out)
}

func (a *App) handleRemoveFriend(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	ctx := r.Context()
	u1, u2 := orderedPair(currentUser(ctx), id)
	if _, err := a.db.Exec(ctx, `DELETE FROM friendships WHERE user_a = $1 AND user_b = $2`, u1, u2); err != nil {
		a.serverError(w, r, err)
		return
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}

type RequestOut struct {
	ID        string     `json:"id"`
	User      PublicUser `json:"user"`
	CreatedAt time.Time  `json:"created_at"`
}

func (a *App) handleListRequests(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	me := currentUser(ctx)
	load := func(q string) ([]RequestOut, error) {
		rows, err := a.db.Query(ctx, q, me)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		out := []RequestOut{}
		for rows.Next() {
			var rq RequestOut
			var uid, name string
			var handle, avatar *string
			if err := rows.Scan(&rq.ID, &rq.CreatedAt, &uid, &name, &handle, &avatar); err != nil {
				return nil, err
			}
			rq.User = a.publicUser(ctx, uid, name, handle, avatar)
			out = append(out, rq)
		}
		return out, rows.Err()
	}
	incoming, err := load(`SELECT fr.id, fr.created_at, u.id, u.name, u.handle, u.avatar_key
		FROM friend_requests fr JOIN users u ON u.id = fr.from_id
		WHERE fr.to_id = $1 AND fr.status = 'pending' ORDER BY fr.created_at DESC`)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	outgoing, err := load(`SELECT fr.id, fr.created_at, u.id, u.name, u.handle, u.avatar_key
		FROM friend_requests fr JOIN users u ON u.id = fr.to_id
		WHERE fr.from_id = $1 AND fr.status = 'pending' ORDER BY fr.created_at DESC`)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	writeJSON(w, 200, map[string]any{"incoming": incoming, "outgoing": outgoing})
}

// handleSendRequest asks someone, by @, to be friends. If they already asked
// you, this accepts their request instead.
func (a *App) handleSendRequest(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Handle string `json:"handle"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	ctx := r.Context()
	me := currentUser(ctx)
	h, _ := cleanHandle(in.Handle)
	var to string
	err := a.db.QueryRow(ctx, `SELECT id FROM users WHERE handle = $1`, h).Scan(&to)
	if isNoRows(err) {
		writeError(w, 404, "Ninguém usa esse @.")
		return
	} else if err != nil {
		a.serverError(w, r, err)
		return
	}
	rel, err := a.relation(ctx, me, to)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	switch rel {
	case "self":
		writeError(w, 400, "Esse é você!")
		return
	case "friend", "request_sent":
		writeJSON(w, 200, map[string]string{"relation": rel})
		return
	case "request_received":
		var id string
		if err := a.db.QueryRow(ctx, `SELECT id FROM friend_requests WHERE from_id = $1 AND to_id = $2 AND status = 'pending'`, to, me).Scan(&id); err != nil {
			a.serverError(w, r, err)
			return
		}
		if err := a.acceptRequest(ctx, id, me, to); err != nil {
			a.serverError(w, r, err)
			return
		}
		writeJSON(w, 200, map[string]string{"relation": "friend"})
		return
	}
	if _, err := a.db.Exec(ctx, `INSERT INTO friend_requests (from_id, to_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, me, to); err != nil {
		a.serverError(w, r, err)
		return
	}
	a.notifyUser(to)
	writeJSON(w, 200, map[string]string{"relation": "request_sent"})
}

func (a *App) acceptRequest(ctx context.Context, id, me, from string) error {
	err := pgx.BeginFunc(ctx, a.db, func(tx pgx.Tx) error {
		// Closes this request and a crossed one in the other direction, if any.
		if _, err := tx.Exec(ctx, `UPDATE friend_requests SET status = 'accepted', responded_at = now()
			WHERE status = 'pending' AND (id = $1 OR (from_id = $2 AND to_id = $3))`, id, me, from); err != nil {
			return err
		}
		u1, u2 := orderedPair(me, from)
		_, err := tx.Exec(ctx, `INSERT INTO friendships (user_a, user_b) VALUES ($1, $2) ON CONFLICT DO NOTHING`, u1, u2)
		return err
	})
	if err == nil {
		a.notifyUser(from)
	}
	return err
}

func (a *App) handleRequestAction(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	action := chi.URLParam(r, "action")
	ctx := r.Context()
	me := currentUser(ctx)
	var from, to string
	err := a.db.QueryRow(ctx, `SELECT from_id, to_id FROM friend_requests WHERE id = $1 AND status = 'pending'`, id).Scan(&from, &to)
	if isNoRows(err) {
		writeError(w, 404, "Esse pedido não existe mais.")
		return
	} else if err != nil {
		a.serverError(w, r, err)
		return
	}
	switch {
	case action == "cancel" && from == me:
		_, err = a.db.Exec(ctx, `UPDATE friend_requests SET status = 'cancelled', responded_at = now() WHERE id = $1`, id)
		a.notifyUser(to)
	case action == "accept" && to == me:
		err = a.acceptRequest(ctx, id, me, from)
	case action == "decline" && to == me:
		_, err = a.db.Exec(ctx, `UPDATE friend_requests SET status = 'declined', responded_at = now() WHERE id = $1`, id)
	default:
		writeError(w, 404, "Não encontrado.")
		return
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}
