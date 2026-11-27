package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// Every game names who made it: `creators` in web/src/lib/games/<id>/index.ts, a list
// of one or more GitHub usernames written as plain strings. The game page shows them.

var (
	creatorsRe  = regexp.MustCompile(`(?s)\bcreators\s*:\s*\[(.*?)\]`)
	quotedRe    = regexp.MustCompile(`'([^'\\]*)'|"([^"\\]*)"`)
	ghUserRe    = regexp.MustCompile(`^[A-Za-z0-9](?:-?[A-Za-z0-9])*$`)
	commaOnlyRe = regexp.MustCompile(`^[\s,]*$`)
)

// parseCreators reads the creators list out of a game module's source.
func parseCreators(src string) ([]string, error) {
	m := creatorsRe.FindStringSubmatch(src)
	if m == nil {
		return nil, fmt.Errorf("has no `creators` list")
	}
	var names []string
	for _, q := range quotedRe.FindAllStringSubmatch(m[1], -1) {
		names = append(names, q[1]+q[2])
	}
	if !commaOnlyRe.MatchString(quotedRe.ReplaceAllString(m[1], "")) {
		return nil, fmt.Errorf("`creators` must be a plain list of quoted GitHub usernames, like ['nelsul', 'friend']")
	}
	return names, nil
}

// userLookup reports whether a GitHub user account exists.
type userLookup func(name string) (bool, error)

// checkCreators checks every game's creators: at least one, valid GitHub usernames
// (no @), no repeats, and (when lookup is set) accounts that exist.
func checkCreators(ids []string, read func(string) string, lookup userLookup, r *report) {
	for _, id := range ids {
		file := "web/src/lib/games/" + id + "/index.ts"
		src := read(file)
		if src == "" {
			r.fail("%s: missing; every game needs a module that names its creators", file)
			continue
		}
		names, err := parseCreators(src)
		if err != nil {
			r.fail("%s: %v", file, err)
			continue
		}
		if len(names) == 0 {
			r.fail("%s: `creators` is empty; list the GitHub username of everyone who made the game", file)
			continue
		}
		seen := map[string]bool{}
		for _, n := range names {
			switch {
			case strings.HasPrefix(n, "@"):
				r.fail("%s: creator %q: write the GitHub username without the @", file, n)
				continue
			case len(n) > 39 || !ghUserRe.MatchString(n):
				r.fail("%s: creator %q is not a valid GitHub username", file, n)
				continue
			case seen[strings.ToLower(n)]:
				r.fail("%s: creator %q is listed twice", file, n)
				continue
			}
			seen[strings.ToLower(n)] = true
			if lookup == nil {
				continue
			}
			ok, err := lookup(n)
			switch {
			case err != nil:
				r.fail("%s: could not check creator %q on GitHub: %v", file, n, err)
			case !ok:
				r.fail("%s: creator %q is not a GitHub user", file, n)
			}
		}
		if len(r.errors) == 0 {
			r.note("%s is made by %s.", id, strings.Join(names, ", "))
		}
	}
}

// githubLookup asks the GitHub API whether a username is a person's account.
func githubLookup(token string) userLookup {
	client := &http.Client{Timeout: 10 * time.Second}
	return func(name string) (bool, error) {
		req, err := http.NewRequest("GET", "https://api.github.com/users/"+name, nil)
		if err != nil {
			return false, err
		}
		req.Header.Set("Accept", "application/vnd.github+json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		res, err := client.Do(req)
		if err != nil {
			return false, err
		}
		defer res.Body.Close()
		switch res.StatusCode {
		case http.StatusOK:
			var u struct {
				Type string `json:"type"`
			}
			if err := json.NewDecoder(res.Body).Decode(&u); err != nil {
				return false, err
			}
			return u.Type == "User", nil
		case http.StatusNotFound:
			return false, nil
		}
		return false, fmt.Errorf("GitHub answered %s", res.Status)
	}
}
