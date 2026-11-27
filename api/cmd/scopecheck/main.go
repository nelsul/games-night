// Command scopecheck is the CI "scope" check for pull requests. A pull request is
// either a hub change or a change to exactly one game:
//
//   - hub: anything outside the games' folders. Title starts with "hub: ".
//   - game <id>: only files in that game's folders. Title starts with "<id>: ".
//
// Game folders: api/games/<id>/, api/games/all/<id>.go, web/src/lib/games/<id>/,
// web/static/games/<id>/ and docs/games/<id>/.
//
// It also checks that migrations already on main are never edited, that a game's SQL
// only touches its own schema (the database enforces it too: see api/gamedb.go), that
// hub SQL never touches a game's schema, that games don't import each other, and
// that every game names its creators (creators.go).
// A change that has to span the hub and games (a refactor of the game interface) can
// pass with the "hub+games" label, which only maintainers can set.
//
// Usage (from the repo root, in CI):
//
//	BASE_SHA=… PR_TITLE=… PR_LABELS=a,b go run ./api/cmd/scopecheck
package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

const hub = "hub"

// owner says which part of the project a path belongs to: "hub" or a game id.
func owner(path string) string {
	parts := strings.Split(path, "/")
	switch {
	case len(parts) == 4 && parts[0] == "api" && parts[1] == "games" && parts[2] == "all" && strings.HasSuffix(parts[3], ".go"):
		return strings.TrimSuffix(parts[3], ".go")
	case len(parts) >= 4 && parts[0] == "api" && parts[1] == "games" && parts[2] != "all":
		return parts[2]
	case len(parts) >= 5 && parts[0] == "web" && parts[1] == "src" && parts[2] == "lib" && parts[3] == "games":
		return parts[4]
	case len(parts) >= 4 && parts[0] == "web" && parts[1] == "static" && parts[2] == "games":
		return parts[3]
	case len(parts) >= 3 && parts[0] == "docs" && parts[1] == "games":
		return parts[2]
	}
	return hub
}

type change struct {
	status string // A, M, D, R…
	path   string
	from   string // for renames
}

type report struct{ errors, notes []string }

func (r *report) fail(format string, a ...any) {
	r.errors = append(r.errors, fmt.Sprintf(format, a...))
}
func (r *report) note(format string, a ...any) { r.notes = append(r.notes, fmt.Sprintf(format, a...)) }

var (
	commentRe   = regexp.MustCompile(`(?s)--[^\n]*|/\*.*?\*/`)
	stringRe    = regexp.MustCompile(`'(?:[^']|'')*'`)
	qualifiedRe = regexp.MustCompile(`(?i)\b([a-z_][a-z0-9_]*)\s*\.\s*([a-z_][a-z0-9_]*)\b`)
	ident       = `([a-z_][a-z0-9_]*(?:\s*\.\s*[a-z_][a-z0-9_]*)?)`
	// What a statement creates or changes, by how it starts.
	targetRes = []*regexp.Regexp{
		regexp.MustCompile(`(?is)^create\s+(?:unique\s+)?index\s+(?:concurrently\s+)?(?:if\s+not\s+exists\s+)?(?:[a-z0-9_]+\s+)?on\s+(?:only\s+)?` + ident),
		regexp.MustCompile(`(?is)^create\s+(?:or\s+replace\s+)?(?:unlogged\s+)?(?:table|view|materialized\s+view|sequence|type)\s+(?:if\s+not\s+exists\s+)?` + ident),
		regexp.MustCompile(`(?is)^alter\s+(?:table|sequence|view|materialized\s+view|type|index)\s+(?:if\s+exists\s+)?(?:only\s+)?` + ident),
		regexp.MustCompile(`(?is)^drop\s+(?:table|index|view|materialized\s+view|sequence|type)\s+(?:if\s+exists\s+)?` + ident),
		regexp.MustCompile(`(?is)^insert\s+into\s+` + ident),
		regexp.MustCompile(`(?is)^update\s+(?:only\s+)?` + ident),
		regexp.MustCompile(`(?is)^delete\s+from\s+(?:only\s+)?` + ident),
		regexp.MustCompile(`(?is)^truncate\s+(?:table\s+)?` + ident),
		regexp.MustCompile(`(?is)^comment\s+on\s+(?:table|column|index)\s+` + ident),
	}
	forbiddenRe  = regexp.MustCompile(`(?i)\b(set\s+(?:local\s+)?role|reset\s+role|set\s+session\s+authorization|search_path|create\s+schema|drop\s+schema|alter\s+schema|grant|revoke|alter\s+default\s+privileges|create\s+role|alter\s+role|drop\s+role|create\s+extension|copy|security\s+definer|create\s+(?:or\s+replace\s+)?function|create\s+(?:or\s+replace\s+)?procedure|create\s+trigger|do)\b|\$[a-z_]*\$`)
	referencesRe = regexp.MustCompile(`(?i)\breferences\s+public\s*\.\s*(?:rooms|users)\b`)
)

func cleanSQL(sql string) string {
	return stringRe.ReplaceAllString(commentRe.ReplaceAllString(sql, " "), "''")
}

// checkGameSQL: a game's migration may only name its own schema, plus foreign keys to
// public.rooms and public.users. Functions, triggers and DO blocks are out: the check
// can't read inside them (the database role would still stop them).
func checkGameSQL(id, file, sql string, r *report) {
	s := cleanSQL(sql)
	if m := forbiddenRe.FindString(s); m != "" {
		r.fail("%s: %q is not allowed in a game migration (plain tables, indexes, views and data only; the hub sets up the game's role and schema)", file, strings.ToUpper(strings.TrimSpace(m)))
	}
	scan := referencesRe.ReplaceAllString(s, "references hub_key")
	for _, m := range qualifiedRe.FindAllStringSubmatch(scan, -1) {
		schema := strings.ToLower(m[1])
		if schema == id || (schema != "public" && !knownSchema(schema)) {
			continue // its own schema, or alias.column
		}
		r.fail("%s: refers to %s.%s; a game may only use its own schema %q (and REFERENCES public.rooms / public.users)", file, m[1], m[2], id)
	}
	for _, stmt := range strings.Split(s, ";") {
		stmt = strings.TrimSpace(stmt)
		for _, re := range targetRes {
			if m := re.FindStringSubmatch(stmt); m != nil {
				target := strings.ReplaceAll(strings.ToLower(m[1]), " ", "")
				if !strings.HasPrefix(target, id+".") {
					r.fail("%s: %q must be qualified with the game's schema, as %s.%s", file, target, id, target)
				}
				break
			}
		}
	}
}

var gameSchemas []string

func knownSchema(s string) bool { return slices.Contains(gameSchemas, s) || s == "pg_catalog" }

// checkHubSQL: hub migrations never touch a game's schema.
func checkHubSQL(file, sql string, r *report) {
	s := cleanSQL(sql)
	for _, m := range qualifiedRe.FindAllStringSubmatch(s, -1) {
		if slices.Contains(gameSchemas, strings.ToLower(m[1])) {
			r.fail("%s: hub migrations must not touch the %q game's schema", file, m[1])
		}
	}
}

var goImportRe = regexp.MustCompile(`"github\.com/nelsul/games-night/api/games/([a-z0-9_]+)`)
var webImportRe = regexp.MustCompile(`(?:\$lib/games/|\.\./\.\./games/|\.\./)([a-z0-9_]+)/`)

// checkImports: a game's code only uses the shared games package and its own files.
func checkImports(id, file, src string, r *report) {
	if strings.HasSuffix(file, ".go") {
		for _, m := range goImportRe.FindAllStringSubmatch(src, -1) {
			if m[1] != id && m[1] != "all" {
				r.fail("%s: imports the %q game; games must not depend on each other", file, m[1])
			}
		}
		if strings.Contains(src, `"github.com/nelsul/games-night/api"`) {
			r.fail("%s: imports the hub; games may only use api/games", file)
		}
		return
	}
	if strings.HasPrefix(file, "web/src/lib/games/") {
		for _, m := range webImportRe.FindAllStringSubmatch(src, -1) {
			if m[1] != id && slices.Contains(gameSchemas, m[1]) {
				r.fail("%s: imports from the %q game; games must not depend on each other", file, m[1])
			}
		}
	}
}

func run(base, title string, labels []string, changes []change, read func(string) string) report {
	var r report
	owners := map[string]bool{}
	for _, c := range changes {
		owners[owner(c.path)] = true
		if c.from != "" {
			owners[owner(c.from)] = true
		}
	}
	var gameIDs []string
	for o := range owners {
		if o != hub {
			gameIDs = append(gameIDs, o)
		}
	}
	slices.Sort(gameIDs)
	mixedOK := slices.Contains(labels, "hub+games")

	kind := hub
	switch {
	case len(changes) == 0:
		r.note("No changed files.")
	case owners[hub] && len(gameIDs) > 0 && !mixedOK:
		r.fail("This pull request changes the hub and games (%s). Split it: one PR for the hub, one per game. (A maintainer can allow an interface change across both with the \"hub+games\" label.)", strings.Join(gameIDs, ", "))
	case !owners[hub] && len(gameIDs) > 1:
		r.fail("This pull request changes more than one game (%s). Open one PR per game.", strings.Join(gameIDs, ", "))
	case !owners[hub] && len(gameIDs) == 1:
		kind = gameIDs[0]
	}
	if len(r.errors) == 0 && len(changes) > 0 {
		if kind == hub {
			r.note("Type: **hub** change%s.", map[bool]string{true: " (with games, allowed by the hub+games label)", false: ""}[len(gameIDs) > 0])
		} else {
			r.note("Type: change to the **%s** game only.", kind)
		}
		if !strings.HasPrefix(strings.ToLower(title), kind+": ") {
			r.fail("The title must start with %q for this kind of pull request (it is %q).", kind+": ", title)
		}
	}

	for _, c := range changes {
		isMigration := strings.Contains(c.path, "/migrations/") && strings.HasSuffix(c.path, ".sql")
		if isMigration && c.status != "A" {
			r.fail("%s: migrations already on main must never be edited, renamed or deleted (production ran them). Add a new numbered file instead.", c.path)
		}
		if c.status == "D" {
			continue
		}
		o := owner(c.path)
		src := read(c.path)
		switch {
		case isMigration && o != hub:
			checkGameSQL(o, c.path, src, &r)
		case isMigration:
			checkHubSQL(c.path, src, &r)
		case o != hub && (strings.HasSuffix(c.path, ".go") || strings.HasSuffix(c.path, ".ts") || strings.HasSuffix(c.path, ".svelte")):
			checkImports(o, c.path, src, &r)
		}
	}
	return r
}

func main() {
	base := os.Getenv("BASE_SHA")
	if base == "" {
		fmt.Println("::error::BASE_SHA is not set")
		os.Exit(2)
	}
	title := strings.TrimSpace(os.Getenv("PR_TITLE"))
	var labels []string
	for _, l := range strings.Split(os.Getenv("PR_LABELS"), ",") {
		if l = strings.TrimSpace(l); l != "" {
			labels = append(labels, l)
		}
	}
	dirs, _ := filepath.Glob("api/games/*/")
	for _, d := range dirs {
		if name := filepath.Base(d); name != "all" {
			gameSchemas = append(gameSchemas, name)
		}
	}

	out, err := exec.Command("git", "diff", "--name-status", "-M", base+"...HEAD").Output()
	if err != nil {
		fmt.Printf("::error::git diff failed: %v\n", err)
		os.Exit(2)
	}
	var changes []change
	sc := bufio.NewScanner(strings.NewReader(string(out)))
	for sc.Scan() {
		f := strings.Split(sc.Text(), "\t")
		if len(f) < 2 {
			continue
		}
		c := change{status: f[0][:1], path: f[len(f)-1]}
		if c.status == "R" && len(f) == 3 {
			c.from = f[1]
		}
		changes = append(changes, c)
	}
	read := func(p string) string {
		b, _ := os.ReadFile(p)
		return string(b)
	}
	r := run(base, title, labels, changes, read)

	var webGames []string
	mods, _ := filepath.Glob("web/src/lib/games/*/")
	for _, d := range mods {
		webGames = append(webGames, filepath.Base(d))
	}
	for _, id := range gameSchemas {
		if !slices.Contains(webGames, id) {
			webGames = append(webGames, id)
		}
	}
	slices.Sort(webGames)
	var lookup userLookup
	if os.Getenv("CREATORS_OFFLINE") == "" {
		lookup = githubLookup(os.Getenv("GITHUB_TOKEN"))
	}
	checkCreators(webGames, read, lookup, &r)

	summary := []string{"## Scope check", ""}
	for _, n := range r.notes {
		summary = append(summary, "- "+n)
		fmt.Println(n)
	}
	for _, e := range r.errors {
		summary = append(summary, "- ❌ "+e)
		fmt.Printf("::error::%s\n", e)
	}
	if len(r.errors) == 0 {
		summary = append(summary, "- ✅ Scope is valid.")
	}
	if p := os.Getenv("GITHUB_STEP_SUMMARY"); p != "" {
		_ = os.WriteFile(p, []byte(strings.Join(summary, "\n")+"\n"), 0o644)
	}
	if len(r.errors) > 0 {
		os.Exit(1)
	}
}
