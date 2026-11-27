package main

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/go-chi/chi/v5"
)

// Feedback, bug reports and game ideas. Anyone signed in can send one, with their
// name or anonymously (then no user is stored at all). Only admins can read them,
// mark them done or delete them; that is everything admins can do for now.

var feedbackKinds = map[string]bool{"feedback": true, "bug": true, "idea": true}

const (
	feedbackMaxLen = 2000
	feedbackBurst  = 5                // sends per person…
	feedbackWindow = 10 * time.Minute // …in this window
)

// feedbackLimiter keeps the last sends of each person in memory, so anonymous
// feedback can be rate limited without being written down with who sent it.
type feedbackLimiter struct {
	mu   sync.Mutex
	sent map[string][]time.Time
}

func (l *feedbackLimiter) allow(userID string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.sent == nil {
		l.sent = map[string][]time.Time{}
	}
	recent := l.sent[userID][:0]
	for _, t := range l.sent[userID] {
		if now.Sub(t) < feedbackWindow {
			recent = append(recent, t)
		}
	}
	if len(recent) >= feedbackBurst {
		l.sent[userID] = recent
		return false
	}
	l.sent[userID] = append(recent, now)
	return true
}

func (a *App) handleSendFeedback(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Kind      string `json:"kind"`
		Body      string `json:"body"`
		Anonymous bool   `json:"anonymous"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	body := strings.TrimSpace(in.Body)
	switch n := utf8.RuneCountInString(body); {
	case !feedbackKinds[in.Kind]:
		writeError(w, 400, "Escolha se é feedback, bug ou ideia de jogo.")
		return
	case n < 3:
		writeError(w, 400, "Escreva um pouco mais.")
		return
	case n > feedbackMaxLen:
		writeError(w, 400, "Use no máximo 2000 caracteres.")
		return
	}
	me := currentUser(r.Context())
	if !a.feedbackLimit.allow(me, time.Now()) {
		writeError(w, 429, "Você enviou muitos feedbacks agora. Tente de novo daqui a pouco.")
		return
	}
	var author *string
	if !in.Anonymous {
		author = &me
	}
	if _, err := a.db.Exec(r.Context(), `INSERT INTO feedback (kind, body, user_id) VALUES ($1, $2, $3)`, in.Kind, body, author); err != nil {
		a.serverError(w, r, err)
		return
	}
	writeJSON(w, 201, map[string]bool{"ok": true})
}

// requireAdmin lets only admins through (after requireUser).
func (a *App) requireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		admin, err := a.isAdmin(r.Context(), currentUser(r.Context()))
		if err != nil {
			a.serverError(w, r, err)
			return
		}
		if !admin {
			writeError(w, 403, "Só admins podem fazer isso.")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (a *App) isAdmin(ctx context.Context, userID string) (bool, error) {
	var admin bool
	err := a.db.QueryRow(ctx, `SELECT is_admin FROM users WHERE id = $1`, userID).Scan(&admin)
	if isNoRows(err) {
		return false, nil
	}
	return admin, err
}

type FeedbackItem struct {
	ID        string      `json:"id"`
	Kind      string      `json:"kind"`
	Body      string      `json:"body"`
	Author    *PublicUser `json:"author"` // null when anonymous
	CreatedAt time.Time   `json:"created_at"`
	DoneAt    *time.Time  `json:"done_at"`
	DoneBy    string      `json:"done_by,omitempty"` // admin's name
}

// handleListFeedback lists open feedback (newest first), or done with ?status=done.
func (a *App) handleListFeedback(w http.ResponseWriter, r *http.Request) {
	done := r.URL.Query().Get("status") == "done"
	ctx := r.Context()
	rows, err := a.db.Query(ctx, `
		SELECT f.id::text, f.kind, f.body, f.created_at, f.done_at,
		       u.id::text, u.name, u.handle, u.avatar_key, coalesce(d.name, '')
		FROM feedback f
		LEFT JOIN users u ON u.id = f.user_id
		LEFT JOIN users d ON d.id = f.done_by
		WHERE (f.done_at IS NOT NULL) = $1
		ORDER BY coalesce(f.done_at, f.created_at) DESC
		LIMIT 300`, done)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer rows.Close()
	out := []FeedbackItem{}
	var open int
	for rows.Next() {
		var it FeedbackItem
		var uid, name, handle, avatar *string
		if err := rows.Scan(&it.ID, &it.Kind, &it.Body, &it.CreatedAt, &it.DoneAt, &uid, &name, &handle, &avatar, &it.DoneBy); err != nil {
			a.serverError(w, r, err)
			return
		}
		if uid != nil {
			u := a.publicUser(ctx, *uid, deref(name), handle, avatar)
			it.Author = &u
		}
		out = append(out, it)
	}
	if err := rows.Err(); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := a.db.QueryRow(ctx, `SELECT count(*) FROM feedback WHERE done_at IS NULL`).Scan(&open); err != nil {
		a.serverError(w, r, err)
		return
	}
	writeJSON(w, 200, map[string]any{"items": out, "open": open})
}

func (a *App) handleFinishFeedback(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if !uuidRe.MatchString(id) {
		writeError(w, 404, "Não encontrado.")
		return
	}
	tag, err := a.db.Exec(r.Context(), `UPDATE feedback SET done_at = now(), done_by = $2 WHERE id = $1 AND done_at IS NULL`, id, currentUser(r.Context()))
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if tag.RowsAffected() == 0 {
		writeError(w, 404, "Esse feedback não existe ou já foi concluído.")
		return
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}

func (a *App) handleDeleteFeedback(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if !uuidRe.MatchString(id) {
		writeError(w, 404, "Não encontrado.")
		return
	}
	tag, err := a.db.Exec(r.Context(), `DELETE FROM feedback WHERE id = $1`, id)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if tag.RowsAffected() == 0 {
		writeError(w, 404, "Esse feedback não existe.")
		return
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}
