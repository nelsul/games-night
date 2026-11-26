package main

import (
	"net/http"

	"github.com/nelsul/games-night/api/games"
	"github.com/nelsul/games-night/api/games/backdoor"
)

// Game is a catalog entry as the site sees it.
type Game = games.Info

// engines is every game the hub can host. Each owns its rules and its own schema.
var engines = []games.Engine{backdoor.New()}

func findEngine(id string) (games.Engine, bool) {
	for _, e := range engines {
		if e.Info().ID == id {
			return e, true
		}
	}
	return nil, false
}

func findGame(id string) (Game, bool) {
	e, ok := findEngine(id)
	if !ok {
		return Game{}, false
	}
	return e.Info(), true
}

func (a *App) handleListGames(w http.ResponseWriter, r *http.Request) {
	out := make([]Game, 0, len(engines))
	for _, e := range engines {
		out = append(out, e.Info())
	}
	writeJSON(w, 200, out)
}
