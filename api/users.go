package main

import (
	"context"
	"net/http"
	"strings"
	"unicode/utf8"
)

// PublicUser is what other people see about someone.
type PublicUser struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

type Me struct {
	PublicUser
	Email         string `json:"email"`
	Onboarded     bool   `json:"onboarded"`
	PhotosEnabled bool   `json:"photos_enabled"`
}

func (a *App) loadMe(ctx context.Context, id string) (Me, error) {
	var m Me
	var avatar *string
	err := a.db.QueryRow(ctx, `SELECT id, email, name, avatar_key FROM users WHERE id = $1`, id).
		Scan(&m.ID, &m.Email, &m.Name, &avatar)
	if err != nil {
		return m, err
	}
	m.AvatarURL = a.store.URL(ctx, avatar)
	m.Onboarded = m.Name != ""
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

func (a *App) handleUpdateMe(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Name *string `json:"name"`
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
		a.notifyUserRooms(ctx, uid)
	}
	a.handleMe(w, r)
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
