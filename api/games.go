package main

import "net/http"

// Game describes one game in the hub's catalog. The rules engine of each game
// plugs in here later; for now the hub only needs its name and table size.
type Game struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	MinPlayers int    `json:"min_players"`
	MaxPlayers int    `json:"max_players"`
	Available  bool   `json:"available"` // false while the game can't be played yet
}

var catalog = []Game{
	{ID: "backdoor", Name: "Backdoor", MinPlayers: 5, MaxPlayers: 10, Available: false},
}

func findGame(id string) (Game, bool) {
	for _, g := range catalog {
		if g.ID == id {
			return g, true
		}
	}
	return Game{}, false
}

func (a *App) handleListGames(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, catalog)
}
