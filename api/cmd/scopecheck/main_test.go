package main

import (
	"strings"
	"testing"
)

func TestOwner(t *testing.T) {
	cases := map[string]string{
		"api/games/backdoor/engine.go":            "backdoor",
		"api/games/backdoor/migrations/002_x.sql": "backdoor",
		"api/games/all/backdoor.go":               "backdoor",
		"api/games/games.go":                      "hub",
		"api/rooms.go":                            "hub",
		"web/src/lib/games/backdoor/Badge.svelte": "backdoor",
		"web/src/lib/games.ts":                    "hub",
		"web/static/games/quiz/sound.mp3":         "quiz",
		"docs/games/quiz/rules.md":                "quiz",
		"docs/agents/domain.md":                   "hub",
		".github/workflows/ci.yml":                "hub",
		"web/src/routes/sala/[code]/+page.svelte": "hub",
		"web/src/lib/components/Avatar.svelte":    "hub",
	}
	for path, want := range cases {
		if got := owner(path); got != want {
			t.Errorf("owner(%q) = %q, want %q", path, got, want)
		}
	}
}

func check(t *testing.T, title string, labels []string, files map[string]string, statuses ...string) report {
	t.Helper()
	gameSchemas = []string{"backdoor", "quiz"}
	var changes []change
	i := 0
	for p := range files {
		st := "A"
		if i < len(statuses) {
			st = statuses[i]
		}
		changes = append(changes, change{status: st, path: p})
		i++
	}
	return run("base", title, labels, changes, func(p string) string { return files[p] })
}

func wantErr(t *testing.T, r report, substr string) {
	t.Helper()
	for _, e := range r.errors {
		if strings.Contains(e, substr) {
			return
		}
	}
	t.Errorf("want an error containing %q, got %v", substr, r.errors)
}

func wantOK(t *testing.T, r report) {
	t.Helper()
	if len(r.errors) > 0 {
		t.Errorf("want no errors, got %v", r.errors)
	}
}

func TestPRTypes(t *testing.T) {
	wantOK(t, check(t, "hub: better lobby", nil, map[string]string{"api/rooms.go": "", "web/src/routes/+page.svelte": ""}))
	wantOK(t, check(t, "quiz: new game", nil, map[string]string{
		"api/games/quiz/game.go":                 "package quiz\nimport \"github.com/nelsul/games-night/api/games\"",
		"api/games/all/quiz.go":                  "package all\nimport _ \"github.com/nelsul/games-night/api/games/quiz\"",
		"api/games/quiz/migrations/001_init.sql": "CREATE TABLE quiz.rounds (id int PRIMARY KEY, room_id uuid REFERENCES public.rooms(id) ON DELETE CASCADE);",
		"web/src/lib/games/quiz/index.ts":        "import type { GameModule } from '$lib/games';",
	}))
	wantErr(t, check(t, "hub: x", nil, map[string]string{"api/rooms.go": "", "api/games/quiz/game.go": ""}), "changes the hub and games")
	wantOK(t, check(t, "hub: interface", []string{"hub+games"}, map[string]string{"api/games/games.go": "", "api/games/quiz/game.go": ""}))
	wantErr(t, check(t, "quiz: x", nil, map[string]string{"api/games/quiz/a.go": "", "api/games/backdoor/b.go": ""}), "more than one game")
	wantErr(t, check(t, "fix stuff", nil, map[string]string{"api/rooms.go": ""}), `must start with "hub: "`)
	wantErr(t, check(t, "hub: wrong", nil, map[string]string{"api/games/quiz/a.go": ""}), `must start with "quiz: "`)
}

func TestMigrationsAreAppendOnly(t *testing.T) {
	wantErr(t, check(t, "hub: x", nil, map[string]string{"api/migrations/001_init.sql": "SELECT 1"}, "M"), "must never be edited")
	wantErr(t, check(t, "quiz: x", nil, map[string]string{"api/games/quiz/migrations/001.sql": ""}, "D"), "must never be edited")
}

func TestGameSQL(t *testing.T) {
	bad := map[string]string{
		"CREATE TABLE backdoor.stolen (x int);":                  "refers to backdoor.stolen",
		"CREATE TABLE things (x int);":                           "must be qualified",
		"UPDATE public.users SET name = 'x';":                    "refers to public.users",
		"CREATE TABLE quiz.e AS SELECT email FROM public.users;": "refers to public.users",
		"GRANT SELECT ON quiz.t TO PUBLIC;":                      "GRANT",
		"SET search_path TO public;":                             "SEARCH_PATH",
		"DO $$ BEGIN PERFORM 1; END $$;":                         "is not allowed",
		"CREATE INDEX i ON rounds (x);":                          "must be qualified",
		"INSERT INTO rounds VALUES (1);":                         "must be qualified",
		"ALTER TABLE public.rooms ADD COLUMN x int;":             "refers to public.rooms",
	}
	for sql, want := range bad {
		wantErr(t, check(t, "quiz: x", nil, map[string]string{"api/games/quiz/migrations/002.sql": sql}), want)
	}
	wantOK(t, check(t, "quiz: x", nil, map[string]string{"api/games/quiz/migrations/002.sql": `
		-- rounds of a match; public.users is mentioned only in this comment
		CREATE TABLE quiz.rounds (
			id int PRIMARY KEY,
			room_id uuid NOT NULL REFERENCES public.rooms(id) ON DELETE CASCADE,
			user_id uuid REFERENCES public.users (id) ON UPDATE CASCADE,
			note text DEFAULT 'public.users is just text here',
			updated_at timestamptz
		);
		CREATE UNIQUE INDEX rounds_room_idx ON quiz.rounds (room_id);
		ALTER TABLE quiz.rounds ADD COLUMN score int;
		UPDATE quiz.rounds r SET score = 0 WHERE r.score IS NULL;
		INSERT INTO quiz.rounds (id, room_id) SELECT 1, r.id FROM quiz.rounds r WHERE false;
	`}))
}

func TestHubSQLLeavesGamesAlone(t *testing.T) {
	wantErr(t, check(t, "hub: x", nil, map[string]string{"api/migrations/009.sql": "DROP TABLE backdoor.matches;"}), "must not touch")
	wantOK(t, check(t, "hub: x", nil, map[string]string{"api/migrations/009.sql": "ALTER TABLE rooms ADD COLUMN x int;"}))
}

func TestGamesDontImportEachOther(t *testing.T) {
	wantErr(t, check(t, "quiz: x", nil, map[string]string{
		"api/games/quiz/game.go": `import "github.com/nelsul/games-night/api/games/backdoor"`,
	}), `imports the "backdoor" game`)
	wantErr(t, check(t, "quiz: x", nil, map[string]string{
		"api/games/quiz/game.go": `import "github.com/nelsul/games-night/api"`,
	}), "imports the hub")
	wantErr(t, check(t, "quiz: x", nil, map[string]string{
		"web/src/lib/games/quiz/Match.svelte": `import Badge from '$lib/games/backdoor/Badge.svelte';`,
	}), `from the "backdoor" game`)
	wantOK(t, check(t, "quiz: x", nil, map[string]string{
		"web/src/lib/games/quiz/Match.svelte": `import Avatar from '$lib/components/Avatar.svelte'; import X from './art/X.svelte';`,
	}))
}
