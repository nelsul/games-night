package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
	"net/http"
	"net/mail"
	"strings"
	"time"
)

const (
	sessionCookie = "gn_session"
	sessionTTL    = 60 * 24 * time.Hour
	codeTTL       = 10 * time.Minute
	codeResend    = 45 * time.Second
	codeAttempts  = 5
)

func hashToken(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func normalizeEmail(s string) (string, bool) {
	s = strings.ToLower(strings.TrimSpace(s))
	addr, err := mail.ParseAddress(s)
	if err != nil || addr.Address != s || len(s) > 200 {
		return "", false
	}
	return s, true
}

func (a *App) handleRequestCode(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Email string `json:"email"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	email, ok := normalizeEmail(in.Email)
	if !ok {
		writeError(w, 400, "Digite um e-mail válido.")
		return
	}
	ctx := r.Context()
	var sentAt time.Time
	err := a.db.QueryRow(ctx, `SELECT sent_at FROM login_codes WHERE email = $1`, email).Scan(&sentAt)
	if err == nil && time.Since(sentAt) < codeResend {
		writeError(w, 429, "Espere alguns segundos para pedir outro código.")
		return
	} else if err != nil && !isNoRows(err) {
		a.serverError(w, r, err)
		return
	}
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	code := fmt.Sprintf("%06d", n.Int64())
	_, err = a.db.Exec(ctx, `
		INSERT INTO login_codes (email, code_hash, attempts, sent_at, expires_at)
		VALUES ($1, $2, 0, now(), now() + $3::interval)
		ON CONFLICT (email) DO UPDATE SET code_hash = EXCLUDED.code_hash, attempts = 0,
			sent_at = now(), expires_at = EXCLUDED.expires_at`,
		email, hashToken(email+":"+code), fmt.Sprintf("%d seconds", int(codeTTL.Seconds())))
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := a.mail.SendCode(email, code); err != nil {
		a.serverError(w, r, fmt.Errorf("send code: %w", err))
		return
	}
	writeJSON(w, 200, map[string]bool{"sent": true})
}

func (a *App) handleVerifyCode(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Email string `json:"email"`
		Code  string `json:"code"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	email, ok := normalizeEmail(in.Email)
	code := strings.TrimSpace(in.Code)
	if !ok || len(code) != 6 {
		writeError(w, 400, "Código inválido.")
		return
	}
	ctx := r.Context()
	tx, err := a.db.Begin(ctx)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback(ctx)

	var hash string
	var attempts int
	var expires time.Time
	err = tx.QueryRow(ctx, `SELECT code_hash, attempts, expires_at FROM login_codes WHERE email = $1 FOR UPDATE`, email).
		Scan(&hash, &attempts, &expires)
	if isNoRows(err) {
		writeError(w, 400, "Peça um novo código.")
		return
	} else if err != nil {
		a.serverError(w, r, err)
		return
	}
	if time.Now().After(expires) || attempts >= codeAttempts {
		writeError(w, 400, "Esse código expirou. Peça um novo.")
		return
	}
	if subtle.ConstantTimeCompare([]byte(hash), []byte(hashToken(email+":"+code))) != 1 {
		_, _ = tx.Exec(ctx, `UPDATE login_codes SET attempts = attempts + 1 WHERE email = $1`, email)
		_ = tx.Commit(ctx)
		writeError(w, 400, "Código incorreto.")
		return
	}
	if _, err := tx.Exec(ctx, `DELETE FROM login_codes WHERE email = $1`, email); err != nil {
		a.serverError(w, r, err)
		return
	}
	var userID string
	err = tx.QueryRow(ctx, `
		INSERT INTO users (email) VALUES ($1)
		ON CONFLICT (email) DO UPDATE SET email = EXCLUDED.email
		RETURNING id`, email).Scan(&userID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		a.serverError(w, r, err)
		return
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	if _, err := tx.Exec(ctx, `INSERT INTO sessions (token_hash, user_id, expires_at) VALUES ($1, $2, $3)`,
		hashToken(token), userID, time.Now().Add(sessionTTL)); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := tx.Commit(ctx); err != nil {
		a.serverError(w, r, err)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name: sessionCookie, Value: token, Path: "/", HttpOnly: true,
		Secure: a.cfg.CookieSecure, SameSite: http.SameSiteLaxMode, MaxAge: int(sessionTTL.Seconds()),
	})
	me, err := a.loadMe(ctx, userID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	writeJSON(w, 200, me)
}

func (a *App) handleLogout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(sessionCookie); err == nil {
		_, _ = a.db.Exec(r.Context(), `DELETE FROM sessions WHERE token_hash = $1`, hashToken(c.Value))
	}
	http.SetCookie(w, &http.Cookie{Name: sessionCookie, Value: "", Path: "/", MaxAge: -1, HttpOnly: true, Secure: a.cfg.CookieSecure, SameSite: http.SameSiteLaxMode})
	writeJSON(w, 200, map[string]bool{"ok": true})
}

func (a *App) requireUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(sessionCookie)
		if err != nil || c.Value == "" {
			writeError(w, 401, "Entre para continuar.")
			return
		}
		var userID string
		err = a.db.QueryRow(r.Context(), `SELECT user_id FROM sessions WHERE token_hash = $1 AND expires_at > now()`, hashToken(c.Value)).Scan(&userID)
		if isNoRows(err) {
			writeError(w, 401, "Entre para continuar.")
			return
		} else if err != nil {
			a.serverError(w, r, err)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userKey, userID)))
	})
}
