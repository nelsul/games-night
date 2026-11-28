// Package linguapresa is a team word game. On a team's turn one person, the narrator,
// sees a word and five forbidden words and has to make their team say the word out loud
// without saying any of the forbidden ones. The other teams see the card too and watch:
// whoever hears a forbidden word taps "Proibida!".
package linguapresa

import (
	"math/rand/v2"
	"slices"
	"time"
)

const (
	MinPlayers = 4
	MaxPlayers = 16
	// MaxSkips is how many cards the narrator may skip in one turn.
	MaxSkips       = 3
	DefaultSeconds = 60
	DefaultLaps    = 1
	// ReadTime is how long the narrator has to read the first card before the clock runs.
	ReadTime = 10 * time.Second
	// BuzzCooldown is how long "Proibida!" stays off after a buzz, so a table tapping at
	// once costs one point, not one per tap.
	BuzzCooldown = 5 * time.Second
)

// Phases of a match.
const (
	PhaseReady   = "ready"   // the narrator is about to start the turn
	PhaseTurn    = "turn"    // the clock is running
	PhaseSummary = "summary" // the turn is over, everyone sees how it went
	PhaseOver    = "over"
)

// How a card ended.
const (
	OutcomeHit  = "hit"
	OutcomeBuzz = "buzz"
	OutcomeSkip = "skip"
)

// RuleError is a broken rule; its message is shown to players (Portuguese).
type RuleError struct{ Msg string }

func (e *RuleError) Error() string { return e.Msg }

func ruleErr(msg string) error { return &RuleError{Msg: msg} }

// Options are the host's lobby settings.
type Options struct {
	Mode    string `json:"mode"`    // two | pairs | trios
	Laps    int    `json:"laps"`    // 1 or 2: how many times each person narrates
	Seconds int    `json:"seconds"` // 45, 60 or 90
}

func (o Options) mode() string {
	if o.Mode == "" {
		return "two"
	}
	return o.Mode
}

func (o Options) laps() int {
	if o.Laps == 0 {
		return DefaultLaps
	}
	return o.Laps
}

func (o Options) seconds() int {
	if o.Seconds == 0 {
		return DefaultSeconds
	}
	return o.Seconds
}

// CheckLobby validates options while people are still joining.
func (o Options) CheckLobby() error {
	switch o.Mode {
	case "", "two", "pairs", "trios":
	default:
		return ruleErr("Escolha dois times, duplas ou trios.")
	}
	switch o.Laps {
	case 0, 1, 2:
	default:
		return ruleErr("Escolha 1 ou 2 voltas.")
	}
	switch o.Seconds {
	case 0, 45, 60, 90:
	default:
		return ruleErr("Escolha 45, 60 ou 90 segundos.")
	}
	return nil
}

// Team is a group playing together.
type Team struct {
	Players []string `json:"players"`
	Score   int      `json:"score"`
}

// Played is a card that was dealt during a turn and how it ended.
type Played struct {
	Card    int    `json:"card"`
	Outcome string `json:"outcome"`
}

// State is a whole match, the current card included. It never leaves the server as is.
type State struct {
	Players []string  `json:"players"`
	Teams   []Team    `json:"teams"`
	Seconds int       `json:"seconds"`
	Turns   int       `json:"turns"` // turns in the whole match
	Turn    int       `json:"turn"`  // 0-based
	Phase   string    `json:"phase"`
	Deck    []int     `json:"deck"`
	Pos     int       `json:"pos"`      // next card to draw from Deck
	Card    int       `json:"card"`     // the card on the table, -1 when none
	CardSeq int       `json:"card_seq"` // counts cards drawn, so a late tap can't hit the next card
	Skips   int       `json:"skips"`
	EndsAt  time.Time `json:"ends_at"`
	// ReadUntil ends the reading time at the start of a turn; the clock (EndsAt) starts
	// after it.
	ReadUntil time.Time `json:"read_until"`
	// BuzzOpenAt is when "Proibida!" works again after the last buzz.
	BuzzOpenAt time.Time `json:"buzz_open_at"`
	Played     []Played  `json:"played"`   // the current (or last) turn
	ReadyAt    time.Time `json:"ready_at"` // when the current turn became ready, so bots can wait
}

// Action is what a player sends.
type Action struct {
	Type  string `json:"type"` // start | hit | skip | buzz | time_up | finish | next | score
	Seq   int    `json:"seq,omitempty"`
	Team  int    `json:"team,omitempty"`  // score: which team
	Delta int    `json:"delta,omitempty"` // score: +1 or -1
}

// Setup splits people into teams and gets the first turn ready.
func Setup(players []string, o Options, rng *rand.Rand, now time.Time) (*State, error) {
	if len(players) < MinPlayers || len(players) > MaxPlayers {
		return nil, ruleErr("Língua Presa é para 4 a 16 pessoas.")
	}
	if err := o.CheckLobby(); err != nil {
		return nil, err
	}
	n := len(players)
	var teamCount int
	switch o.mode() {
	case "two":
		teamCount = 2
	case "pairs":
		if n%2 != 0 {
			return nil, ruleErr("Duplas precisam de um número par de pessoas.")
		}
		teamCount = n / 2
	case "trios":
		if n%3 != 0 || n < 6 {
			return nil, ruleErr("Trios precisam de 6, 9, 12 ou 15 pessoas.")
		}
		teamCount = n / 3
	}
	order := slices.Clone(players)
	rng.Shuffle(len(order), func(i, j int) { order[i], order[j] = order[j], order[i] })
	teams := make([]Team, teamCount)
	for i, p := range order {
		teams[i%teamCount].Players = append(teams[i%teamCount].Players, p)
	}
	biggest := 0
	for _, t := range teams {
		biggest = max(biggest, len(t.Players))
	}
	s := &State{
		Players: slices.Clone(players),
		Teams:   teams,
		Seconds: o.seconds(),
		Turns:   o.laps() * biggest * teamCount,
		Phase:   PhaseReady,
		Card:    -1,
		ReadyAt: now,
	}
	s.reshuffle(rng)
	return s, nil
}

// reshuffle builds a new draw order that never starts with the card on the table.
func (s *State) reshuffle(rng *rand.Rand) {
	s.Deck = rng.Perm(len(cards))
	if len(s.Deck) > 1 && s.Deck[0] == s.Card {
		s.Deck[0], s.Deck[1] = s.Deck[1], s.Deck[0]
	}
	s.Pos = 0
}

func (s *State) draw(rng *rand.Rand) {
	if s.Pos >= len(s.Deck) {
		s.reshuffle(rng)
	}
	s.Card = s.Deck[s.Pos]
	s.Pos++
	s.CardSeq++
}

// ActiveTeam is the index of the team playing this turn.
func (s *State) ActiveTeam() int { return s.Turn % len(s.Teams) }

// Narrator is who gives the clues this turn; each team rotates through its people.
func (s *State) Narrator() string {
	t := s.Teams[s.ActiveTeam()]
	return t.Players[(s.Turn/len(s.Teams))%len(t.Players)]
}

// TeamOf is the index of the team a player is on, or -1.
func (s *State) TeamOf(p string) int {
	for i, t := range s.Teams {
		if slices.Contains(t.Players, p) {
			return i
		}
	}
	return -1
}

// Winners are the teams with the best score once the match is over.
func (s *State) Winners() []int {
	best := s.Teams[0].Score
	for _, t := range s.Teams {
		best = max(best, t.Score)
	}
	var w []int
	for i, t := range s.Teams {
		if t.Score == best {
			w = append(w, i)
		}
	}
	return w
}

func (s *State) record(outcome string) {
	s.Played = append(s.Played, Played{Card: s.Card, Outcome: outcome})
}

func (s *State) endTurn() {
	s.Phase = PhaseSummary
	s.Card = -1
}

// Apply runs one player's action.
func Apply(s *State, user string, a Action, now time.Time, rng *rand.Rand) error {
	if !slices.Contains(s.Players, user) {
		return ruleErr("Você não está nesta partida.")
	}
	switch a.Type {
	case "start":
		if s.Phase != PhaseReady {
			return ruleErr("O turno já começou.")
		}
		if user != s.Narrator() {
			return ruleErr("Só quem narra pode começar o turno.")
		}
		s.Phase = PhaseTurn
		s.ReadUntil = now.Add(ReadTime)
		s.EndsAt = s.ReadUntil.Add(time.Duration(s.Seconds) * time.Second)
		s.BuzzOpenAt = time.Time{}
		s.Played = nil
		s.Skips = 0
		s.draw(rng)
		return nil
	case "hit", "skip", "buzz":
		if s.Phase != PhaseTurn {
			return ruleErr("O turno não está rodando.")
		}
		if !now.Before(s.EndsAt) {
			return ruleErr("O tempo acabou.")
		}
		if a.Seq != s.CardSeq {
			return ruleErr("Essa carta já passou.")
		}
		narrator := user == s.Narrator()
		// While the narrator reads, nobody can score or buzz; a card can still be skipped.
		if a.Type != "skip" && now.Before(s.ReadUntil) {
			return ruleErr("Ainda é a hora de ler a carta.")
		}
		switch a.Type {
		case "hit":
			if !narrator {
				return ruleErr("Só quem narra marca o acerto.")
			}
			s.Teams[s.ActiveTeam()].Score++
			s.record(OutcomeHit)
		case "skip":
			if !narrator {
				return ruleErr("Só quem narra pode pular.")
			}
			if s.Skips >= MaxSkips {
				return ruleErr("Você já pulou o máximo de cartas.")
			}
			s.Skips++
			s.record(OutcomeSkip)
		case "buzz":
			if s.TeamOf(user) == s.ActiveTeam() {
				return ruleErr("O seu time não pode apertar a campainha.")
			}
			if now.Before(s.BuzzOpenAt) {
				return ruleErr("A campainha acabou de tocar. Espere uns segundos.")
			}
			s.Teams[s.ActiveTeam()].Score--
			s.BuzzOpenAt = now.Add(BuzzCooldown)
			s.record(OutcomeBuzz)
		}
		s.draw(rng)
		return nil
	case "time_up":
		if s.Phase != PhaseTurn {
			return ruleErr("O turno não está rodando.")
		}
		if now.Before(s.EndsAt) {
			return ruleErr("Ainda tem tempo.")
		}
		s.endTurn()
		return nil
	case "finish":
		if s.Phase != PhaseTurn {
			return ruleErr("O turno não está rodando.")
		}
		if user != s.Narrator() {
			return ruleErr("Só quem narra encerra o turno.")
		}
		s.endTurn()
		return nil
	case "next":
		if s.Phase != PhaseSummary {
			return ruleErr("O turno ainda não acabou.")
		}
		s.Turn++
		if s.Turn >= s.Turns {
			s.Phase = PhaseOver
		} else {
			s.Phase = PhaseReady
			s.ReadyAt = now
		}
		return nil
	}
	return ruleErr("Ação desconhecida.")
}

// Adjust fixes a team's score by one point, for the host to undo a mistake.
func Adjust(s *State, team, delta int) error {
	if s.Phase == PhaseOver {
		return ruleErr("A partida já acabou.")
	}
	if team < 0 || team >= len(s.Teams) {
		return ruleErr("Esse time não existe.")
	}
	if delta != 1 && delta != -1 {
		return ruleErr("Ajuste um ponto por vez.")
	}
	s.Teams[team].Score += delta
	return nil
}
