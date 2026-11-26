package main

import (
	"context"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/jackc/pgx/v5/pgconn"
)

// PublicUser is what other people see about someone.
type PublicUser struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Handle    string `json:"handle"`
	AvatarURL string `json:"avatar_url"`
}

type Me struct {
	PublicUser
	Email         string `json:"email"`
	Onboarded     bool   `json:"onboarded"`
	PhotosEnabled bool   `json:"photos_enabled"`
}

// publicUser builds the JSON for a person from scanned columns.
func (a *App) publicUser(ctx context.Context, id, name string, handle, avatar *string) PublicUser {
	return PublicUser{ID: id, Name: name, Handle: deref(handle), AvatarURL: a.store.URL(ctx, avatar)}
}

func (a *App) loadMe(ctx context.Context, id string) (Me, error) {
	var m Me
	var name string
	var handle, avatar *string
	err := a.db.QueryRow(ctx, `SELECT id, email, name, handle, avatar_key FROM users WHERE id = $1`, id).
		Scan(&m.ID, &m.Email, &name, &handle, &avatar)
	if err != nil {
		return m, err
	}
	m.PublicUser = a.publicUser(ctx, m.ID, name, handle, avatar)
	m.Onboarded = m.Name != "" && m.Handle != ""
	m.PhotosEnabled = a.store != nil
	return m, nil
}

func (a *App) handleMe(w http.ResponseWriter, r *http.Request) {
	me, err := a.loadMe(r.Context(), currentUser(r.Context()))
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	writeJSON(w, 200, me)
}

// cleanName trims a display name and checks its length (1 to 24 letters).
func cleanName(s string) (string, bool) {
	s = strings.Join(strings.Fields(s), " ")
	n := utf8.RuneCountInString(s)
	return s, n >= 1 && n <= 24
}

var handleRe = regexp.MustCompile(`^[a-z0-9._]{3,20}$`)

// cleanHandle lower-cases a typed @handle, drops the "@" and checks the format.
func cleanHandle(s string) (string, bool) {
	h := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(s), "@"))
	return h, handleRe.MatchString(h)
}

func (a *App) handleUpdateMe(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Name   *string `json:"name"`
		Handle *string `json:"handle"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	ctx := r.Context()
	uid := currentUser(ctx)
	if in.Name != nil {
		name, ok := cleanName(*in.Name)
		if !ok {
			writeError(w, 400, "Digite um nome de até 24 letras.")
			return
		}
		if _, err := a.db.Exec(ctx, `UPDATE users SET name = $1 WHERE id = $2`, name, uid); err != nil {
			a.serverError(w, r, err)
			return
		}
	}
	if in.Handle != nil {
		h, ok := cleanHandle(*in.Handle)
		if !ok {
			writeError(w, 400, "O @ precisa ter de 3 a 20 caracteres: letras minúsculas, números, ponto ou _.")
			return
		}
		if _, err := a.db.Exec(ctx, `UPDATE users SET handle = $1 WHERE id = $2`, h, uid); err != nil {
			// The unique index decides when two people pick the same free @ at once.
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				writeError(w, 409, "Esse @ já está em uso.")
				return
			}
			a.serverError(w, r, err)
			return
		}
	}
	if in.Name != nil || in.Handle != nil {
		a.notifyUserRooms(ctx, uid)
	}
	a.handleMe(w, r)
}

// handleHandleAvailable lets the forms check a @ while it is typed.
func (a *App) handleHandleAvailable(w http.ResponseWriter, r *http.Request) {
	h, ok := cleanHandle(r.URL.Query().Get("handle"))
	if !ok {
		writeJSON(w, 200, map[string]any{"handle": h, "valid": false, "available": false})
		return
	}
	var takenBy *string
	err := a.db.QueryRow(r.Context(), `SELECT id FROM users WHERE handle = $1`, h).Scan(&takenBy)
	if err != nil && !isNoRows(err) {
		a.serverError(w, r, err)
		return
	}
	available := takenBy == nil || *takenBy == currentUser(r.Context())
	writeJSON(w, 200, map[string]any{"handle": h, "valid": true, "available": available})
}

type UserWithRelation struct {
	PublicUser
	Relation string `json:"relation"` // self, friend, request_sent, request_received or none
}

// handleSearchUsers finds people whose @ starts with what was typed.
func (a *App) handleSearchUsers(w http.ResponseWriter, r *http.Request) {
	q := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(r.URL.Query().Get("q")), "@"))
	if len(q) < 2 {
		writeJSON(w, 200, []UserWithRelation{})
		return
	}
	q = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(q)
	ctx := r.Context()
	me := currentUser(ctx)
	rows, err := a.db.Query(ctx, `
		SELECT id, name, handle, avatar_key FROM users
		WHERE handle LIKE $1 || '%' AND id <> $2 AND name <> ''
		ORDER BY length(handle), handle LIMIT 10`, q, me)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer rows.Close()
	out := []UserWithRelation{}
	for rows.Next() {
		var id, name string
		var handle, avatar *string
		if err := rows.Scan(&id, &name, &handle, &avatar); err != nil {
			a.serverError(w, r, err)
			return
		}
		out = append(out, UserWithRelation{PublicUser: a.publicUser(ctx, id, name, handle, avatar)})
	}
	rows.Close()
	for i := range out {
		if out[i].Relation, err = a.relation(ctx, me, out[i].ID); err != nil {
			a.serverError(w, r, err)
			return
		}
	}
	writeJSON(w, 200, out)
}

func (a *App) handleUploadAvatar(w http.ResponseWriter, r *http.Request) {
	if a.store == nil {
		writeError(w, 503, "As fotos ainda não estão disponíveis.")
		return
	}
	data, ctype, ext, ok := readImage(w, r, avatarSide)
	if !ok {
		return
	}
	ctx := r.Context()
	uid := currentUser(ctx)
	key, err := a.store.Put(ctx, "avatars/"+uid, ext, ctype, data)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	var old *string
	if err := a.db.QueryRow(ctx, `UPDATE users u SET avatar_key = $1 FROM users o WHERE u.id = o.id AND u.id = $2 RETURNING o.avatar_key`, key, uid).Scan(&old); err != nil {
		a.store.Delete(ctx, key)
		a.serverError(w, r, err)
		return
	}
	a.store.Delete(ctx, deref(old))
	a.notifyUserRooms(ctx, uid)
	a.handleMe(w, r)
}

func (a *App) handleDeleteAvatar(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uid := currentUser(ctx)
	var old *string
	if err := a.db.QueryRow(ctx, `UPDATE users u SET avatar_key = NULL FROM users o WHERE u.id = o.id AND u.id = $1 RETURNING o.avatar_key`, uid).Scan(&old); err != nil {
		a.serverError(w, r, err)
		return
	}
	a.store.Delete(ctx, deref(old))
	a.notifyUserRooms(ctx, uid)
	a.handleMe(w, r)
}

// handleGetUser is someone's card: who they are and how they relate to you (for the
// player card in a room). Only what everyone may see; bots have no friendship.
func (a *App) handleGetUser(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	ctx := r.Context()
	var name string
	var handle, avatar *string
	var isBot bool
	err := a.db.QueryRow(ctx, `SELECT name, handle, avatar_key, is_bot FROM users WHERE id = $1`, id).Scan(&name, &handle, &avatar, &isBot)
	if isNoRows(err) {
		writeError(w, 404, "Não encontramos essa pessoa.")
		return
	} else if err != nil {
		a.serverError(w, r, err)
		return
	}
	out := struct {
		UserWithRelation
		IsBot bool `json:"is_bot"`
	}{UserWithRelation: UserWithRelation{PublicUser: a.publicUser(ctx, id, name, handle, avatar), Relation: "none"}, IsBot: isBot}
	if !isBot {
		if out.Relation, err = a.relation(ctx, currentUser(ctx), id); err != nil {
			a.serverError(w, r, err)
			return
		}
	}
	writeJSON(w, 200, out)
}
