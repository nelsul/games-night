package main

import (
	"net/http"

	"github.com/nelsul/games-night/api/games"
	_ "github.com/nelsul/games-night/api/games/all" // registers every game
)

// Game is a catalog entry as the site sees it.
type Game = games.Info

func findEngine(id string) (games.Engine, bool) {
	for _, e := range games.All() {
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
	out := []Game{}
	for _, e := range games.All() {
		out = append(out, e.Info())
	}
	writeJSON(w, 200, out)
}
