package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
)

type ctxKey int

const userKey ctxKey = 1

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError sends {"error": msg}. Messages are shown to users, so they are in Portuguese.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func (a *App) serverError(w http.ResponseWriter, r *http.Request, err error) {
	a.log.Error("request failed", "path", r.URL.Path, "err", err)
	writeError(w, 500, "Algo deu errado. Tente de novo.")
}

func decodeJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		writeError(w, 400, "Requisição inválida.")
		return false
	}
	return true
}

func currentUser(ctx context.Context) string {
	id, _ := ctx.Value(userKey).(string)
	return id
}

func isNoRows(err error) bool { return errors.Is(err, pgx.ErrNoRows) }

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
