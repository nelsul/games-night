package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// The home page shows who has worked on the project: the repository's contributors on
// GitHub. They are fetched at most every 15 minutes and kept in memory; if GitHub is down
// or rate limited, the last good list is served.

type Contributor struct {
	Login         string `json:"login"`
	AvatarURL     string `json:"avatar_url"`
	URL           string `json:"url"`
	Contributions int    `json:"contributions"`
}

type contributorsCache struct {
	mu      sync.Mutex
	list    []Contributor
	fetched time.Time
	tried   time.Time
}

const contributorsTTL = 15 * time.Minute

func (a *App) handleContributors(w http.ResponseWriter, r *http.Request) {
	c := a.contributors
	c.mu.Lock()
	defer c.mu.Unlock()
	// Retry a failed fetch after a minute, not on every request.
	if time.Since(c.fetched) > contributorsTTL && time.Since(c.tried) > time.Minute {
		c.tried = time.Now()
		ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
		list, err := fetchContributors(ctx, a.cfg.Repo)
		cancel()
		if err != nil {
			a.log.Warn("contributors", "err", err)
		} else {
			c.list, c.fetched = list, time.Now()
		}
	}
	w.Header().Set("Cache-Control", "public, max-age=300")
	writeJSON(w, 200, map[string]any{"repo": "https://github.com/" + a.cfg.Repo, "contributors": nonNilContributors(c.list)})
}

func nonNilContributors(l []Contributor) []Contributor {
	if l == nil {
		return []Contributor{}
	}
	return l
}

// fetchContributors lists the people (not bots) who committed to repo, most commits first.
func fetchContributors(ctx context.Context, repo string) ([]Contributor, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/repos/"+repo+"/contributors?per_page=100", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "games-night")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub answered %s", res.Status)
	}
	var raw []struct {
		Login         string `json:"login"`
		AvatarURL     string `json:"avatar_url"`
		HTMLURL       string `json:"html_url"`
		Type          string `json:"type"`
		Contributions int    `json:"contributions"`
	}
	if err := json.NewDecoder(res.Body).Decode(&raw); err != nil {
		return nil, err
	}
	out := []Contributor{}
	for _, u := range raw {
		if u.Type != "User" {
			continue
		}
		out = append(out, Contributor{Login: u.Login, AvatarURL: u.AvatarURL, URL: u.HTMLURL, Contributions: u.Contributions})
	}
	return out, nil
}
