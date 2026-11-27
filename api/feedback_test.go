package main

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// feedbackApp is a real API on a throwaway database, with a regular person and an admin.
func feedbackApp(t *testing.T) (*pgxpool.Pool, func(method, path, token, body string) (int, string)) {
	t.Helper()
	db := testDB(t)
	migrateAll(t, db)
	ctx := context.Background()
	for _, u := range []struct{ handle, admin string }{{"ana", "false"}, {"chefe", "true"}} {
		_, err := db.Exec(ctx, `WITH u AS (INSERT INTO users (email, name, handle, is_admin) VALUES ($1 || '@example.test', $1, $1, $2::boolean) RETURNING id)
			INSERT INTO sessions (token_hash, user_id, expires_at) SELECT $3, id, now() + interval '1 day' FROM u`, u.handle, u.admin, hashToken("tok-"+u.handle))
		if err != nil {
			t.Fatal(err)
		}
	}
	a := &App{db: db, hub: NewHub(), bots: newBotRunner(), contributors: &contributorsCache{}, log: slog.New(slog.NewTextHandler(io.Discard, nil))}
	h := a.routes()
	call := func(method, path, token, body string) (int, string) {
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.AddCookie(&http.Cookie{Name: sessionCookie, Value: token})
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec.Code, rec.Body.String()
	}
	return db, call
}

type feedbackList struct {
	Items []FeedbackItem `json:"items"`
	Open  int            `json:"open"`
}

func listFeedback(t *testing.T, call func(string, string, string, string) (int, string), status string) feedbackList {
	t.Helper()
	code, body := call("GET", "/api/admin/feedback?status="+status, "tok-chefe", "")
	if code != 200 {
		t.Fatalf("admin list: %d %s", code, body)
	}
	var l feedbackList
	if err := json.Unmarshal([]byte(body), &l); err != nil {
		t.Fatal(err)
	}
	return l
}

func TestFeedbackNamedAnonymousAndAdminOnly(t *testing.T) {
	_, call := feedbackApp(t)

	if code, _ := call("POST", "/api/feedback", "tok-ana", `{"kind":"bug","body":"O timer some no iPhone."}`); code != 201 {
		t.Fatalf("named send: %d", code)
	}
	if code, _ := call("POST", "/api/feedback", "tok-ana", `{"kind":"idea","body":"Um jogo de mímica.","anonymous":true}`); code != 201 {
		t.Fatalf("anonymous send: %d", code)
	}
	for _, bad := range []string{`{"kind":"spam","body":"oi oi oi"}`, `{"kind":"bug","body":"  a "}`, `{"kind":"bug","body":"` + strings.Repeat("x", 2001) + `"}`} {
		if code, _ := call("POST", "/api/feedback", "tok-ana", bad); code != 400 {
			t.Errorf("%.40s… accepted (%d)", bad, code)
		}
	}
	if code, _ := call("POST", "/api/feedback", "", `{"kind":"bug","body":"sem login"}`); code != 401 {
		t.Errorf("signed out send: %d", code)
	}

	// Regular people can't read, close or delete feedback.
	for _, req := range [][2]string{{"GET", "/api/admin/feedback"}, {"POST", "/api/admin/feedback/00000000-0000-0000-0000-000000000000/done"}, {"DELETE", "/api/admin/feedback/00000000-0000-0000-0000-000000000000"}} {
		if code, _ := call(req[0], req[1], "tok-ana", ""); code != 403 {
			t.Errorf("%s %s as a regular person: %d", req[0], req[1], code)
		}
	}

	l := listFeedback(t, call, "open")
	if len(l.Items) != 2 || l.Open != 2 {
		t.Fatalf("open list: %+v", l)
	}
	var named, anon FeedbackItem
	for _, it := range l.Items {
		if it.Kind == "bug" {
			named = it
		} else {
			anon = it
		}
	}
	if named.Author == nil || named.Author.Handle != "ana" {
		t.Errorf("named feedback author: %+v", named.Author)
	}
	if anon.Author != nil {
		t.Errorf("anonymous feedback shows %+v", anon.Author)
	}

	if code, _ := call("POST", "/api/admin/feedback/"+named.ID+"/done", "tok-chefe", ""); code != 200 {
		t.Fatalf("done: %d", code)
	}
	if code, _ := call("POST", "/api/admin/feedback/"+named.ID+"/done", "tok-chefe", ""); code != 404 {
		t.Errorf("done twice: %d", code)
	}
	if l := listFeedback(t, call, "done"); len(l.Items) != 1 || l.Items[0].DoneBy != "chefe" || l.Open != 1 {
		t.Errorf("done list: %+v", l)
	}
	if code, _ := call("DELETE", "/api/admin/feedback/"+anon.ID, "tok-chefe", ""); code != 200 {
		t.Fatalf("delete: %d", code)
	}
	if l := listFeedback(t, call, "open"); len(l.Items) != 0 || l.Open != 0 {
		t.Errorf("after delete: %+v", l)
	}
}

func TestAnonymousFeedbackStoresNoOne(t *testing.T) {
	db, call := feedbackApp(t)
	call("POST", "/api/feedback", "tok-ana", `{"kind":"feedback","body":"Curti muito!","anonymous":true}`)
	var n int
	if err := db.QueryRow(context.Background(), `SELECT count(*) FROM feedback WHERE user_id IS NULL AND done_by IS NULL`).Scan(&n); err != nil || n != 1 {
		t.Fatalf("anonymous rows without a user: %d (%v)", n, err)
	}
}

func TestFeedbackRateLimit(t *testing.T) {
	_, call := feedbackApp(t)
	for i := range feedbackBurst {
		if code, body := call("POST", "/api/feedback", "tok-ana", `{"kind":"feedback","body":"mensagem de teste","anonymous":true}`); code != 201 {
			t.Fatalf("send %d: %d %s", i, code, body)
		}
	}
	if code, _ := call("POST", "/api/feedback", "tok-ana", `{"kind":"feedback","body":"mais uma"}`); code != 429 {
		t.Errorf("over the limit: %d", code)
	}
	// Others aren't affected.
	if code, _ := call("POST", "/api/feedback", "tok-chefe", `{"kind":"feedback","body":"do admin"}`); code != 201 {
		t.Errorf("another person: %d", code)
	}
}

func TestMeSaysWhoIsAdmin(t *testing.T) {
	_, call := feedbackApp(t)
	for token, want := range map[string]bool{"tok-ana": false, "tok-chefe": true} {
		_, body := call("GET", "/api/me", token, "")
		var me Me
		_ = json.Unmarshal([]byte(body), &me)
		if me.IsAdmin != want {
			t.Errorf("%s is_admin = %v", token, me.IsAdmin)
		}
	}
}
