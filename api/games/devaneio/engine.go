// Package devaneio is a storytelling game played with picture cards. Each round one
// person, the teller, picks a card from their hand and gives a clue for it. Everyone else
// plays the card from their own hand that best fits the clue, the cards are shuffled
// face up, and everyone but the teller votes for the one they think is the teller's.
// The clue has to be good enough that some people find the card, but not all of them.
//
// The pictures are drawn by the web app from the card number; the server only knows
// numbers, so a card is just an int here.
package devaneio

import (
	"math/rand/v2"
	"slices"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	MinPlayers = 4
	MaxPlayers = 14
	// CardCount is how many different pictures exist (the web app draws card 0..CardCount-1).
	CardCount     = 160
	HandSize      = 6
	MaxClueRune   = 60
	DefaultRounds = 5
)

// Phases of a round.
const (
	PhaseStory  = "story"  // the teller picks a card and gives a clue
	PhasePlay   = "play"   // everyone else plays a card that fits the clue
	PhaseVote   = "vote"   // everyone but the teller guesses which card is theirs
	PhaseReveal = "reveal" // the cards show who played them and who voted for them
	PhaseOver   = "over"
)

// RuleError is a broken rule; its message is shown to players (Portuguese).
type RuleError struct{ Msg string }

func (e *RuleError) Error() string { return e.Msg }

func ruleErr(msg string) error { return &RuleError{Msg: msg} }

// Options are the host's lobby settings.
type Options struct {
	// Rounds is 5, 8, or 0 for "one turn each": everyone is the teller once.
	Rounds int `json:"rounds"`
}

// CheckLobby validates options while people are still joining.
func (o Options) CheckLobby() error {
	switch o.Rounds {
	case 0, 5, 8:
		return nil
	}
	return ruleErr("Escolha 5 ou 8 rodadas, ou uma volta.")
}

func (o Options) rounds(players int) int {
	switch o.Rounds {
	case 0:
		return players
	}
	return o.Rounds
}

// State is a whole match, every hand included. It never leaves the server as is.
type State struct {
	Players []string         `json:"players"`
	Rounds  int              `json:"rounds"`
	Round   int              `json:"round"` // 0-based
	Phase   string           `json:"phase"`
	Hands   map[string][]int `json:"hands"`
	Deck    []int            `json:"deck"`
	Discard []int            `json:"discard"`
	Scores  map[string]int   `json:"scores"`

	TellerCard int            `json:"teller_card"`
	Clue       string         `json:"clue"`
	Played     map[string]int `json:"played"` // who played which card this round (the teller too)
	Table      []string       `json:"table"`  // owners of the cards on the table, in table order
	Votes      map[string]int `json:"votes"`  // voter -> position on the table
	Gains      map[string]int `json:"gains"`  // points each person got this round, once it is scored
}

// Action is what a player sends.
type Action struct {
	Type  string `json:"type"` // story | play | vote | next
	Card  int    `json:"card,omitempty"`
	Text  string `json:"text,omitempty"`
	Index int    `json:"index,omitempty"`
}

// Setup shuffles the deck, deals the hands and starts the first round.
func Setup(players []string, o Options, rng *rand.Rand, _ time.Time) (*State, error) {
	if len(players) < MinPlayers || len(players) > MaxPlayers {
		return nil, ruleErr("Devaneio é para 4 a 14 pessoas.")
	}
	if err := o.CheckLobby(); err != nil {
		return nil, err
	}
	s := &State{
		Players: slices.Clone(players),
		Rounds:  o.rounds(len(players)),
		Hands:   map[string][]int{},
		Deck:    rng.Perm(CardCount),
		Scores:  map[string]int{},
	}
	for _, p := range players {
		s.Scores[p] = 0
		s.Hands[p] = nil
		s.refill(p, rng)
	}
	s.startRound()
	return s, nil
}

// Teller is the person who gives the clue this round; the turn goes around the table.
func (s *State) Teller() string { return s.Players[s.Round%len(s.Players)] }

func (s *State) has(p string) bool { return slices.Contains(s.Players, p) }

func (s *State) startRound() {
	s.Phase = PhaseStory
	s.TellerCard = -1
	s.Clue = ""
	s.Played = map[string]int{}
	s.Table = nil
	s.Votes = map[string]int{}
	s.Gains = map[string]int{}
}

// draw takes the top card, shuffling the discard pile back in when the deck runs out.
func (s *State) draw(rng *rand.Rand) int {
	if len(s.Deck) == 0 {
		s.Deck, s.Discard = s.Discard, nil
		rng.Shuffle(len(s.Deck), func(i, j int) { s.Deck[i], s.Deck[j] = s.Deck[j], s.Deck[i] })
	}
	c := s.Deck[0]
	s.Deck = s.Deck[1:]
	return c
}

func (s *State) refill(p string, rng *rand.Rand) {
	for len(s.Hands[p]) < HandSize {
		s.Hands[p] = append(s.Hands[p], s.draw(rng))
	}
}

// removeFromHand takes card out of p's hand, or reports that it was not there.
func (s *State) removeFromHand(p string, card int) bool {
	i := slices.Index(s.Hands[p], card)
	if i < 0 {
		return false
	}
	s.Hands[p] = slices.Delete(s.Hands[p], i, i+1)
	return true
}

// Apply runs one player's action.
func Apply(s *State, user string, a Action, rng *rand.Rand) error {
	if !s.has(user) {
		return ruleErr("Você não está nesta partida.")
	}
	teller := s.Teller()
	switch a.Type {
	case "story":
		if s.Phase != PhaseStory {
			return ruleErr("A história desta rodada já foi contada.")
		}
		if user != teller {
			return ruleErr("Só quem conta a história escolhe a carta e a dica.")
		}
		clue := strings.Join(strings.Fields(a.Text), " ")
		if clue == "" {
			return ruleErr("Escreva uma dica.")
		}
		if utf8.RuneCountInString(clue) > MaxClueRune {
			return ruleErr("A dica pode ter até 60 letras.")
		}
		if !s.removeFromHand(user, a.Card) {
			return ruleErr("Essa carta não está na sua mão.")
		}
		s.TellerCard = a.Card
		s.Clue = clue
		s.Played[user] = a.Card
		s.Phase = PhasePlay
		return nil
	case "play":
		if s.Phase != PhasePlay {
			return ruleErr("Ainda não é hora de jogar uma carta.")
		}
		if user == teller {
			return ruleErr("Você já escolheu a sua carta.")
		}
		if _, done := s.Played[user]; done {
			return ruleErr("Você já jogou uma carta.")
		}
		if !s.removeFromHand(user, a.Card) {
			return ruleErr("Essa carta não está na sua mão.")
		}
		s.Played[user] = a.Card
		if len(s.Played) == len(s.Players) {
			s.openTable(rng)
		}
		return nil
	case "vote":
		if s.Phase != PhaseVote {
			return ruleErr("Ainda não é hora de votar.")
		}
		if user == teller {
			return ruleErr("Quem conta a história não vota.")
		}
		if _, done := s.Votes[user]; done {
			return ruleErr("Você já votou.")
		}
		if a.Index < 0 || a.Index >= len(s.Table) {
			return ruleErr("Escolha uma das cartas da mesa.")
		}
		if s.Table[a.Index] == user {
			return ruleErr("Você não pode votar na sua própria carta.")
		}
		s.Votes[user] = a.Index
		if len(s.Votes) == len(s.Players)-1 {
			s.score()
		}
		return nil
	case "next":
		if s.Phase != PhaseReveal {
			return ruleErr("A rodada ainda não terminou.")
		}
		for _, c := range s.Played {
			s.Discard = append(s.Discard, c)
		}
		s.Round++
		if s.Round >= s.Rounds {
			s.Phase = PhaseOver
			return nil
		}
		for _, p := range s.Players {
			s.refill(p, rng)
		}
		s.startRound()
		return nil
	}
	return ruleErr("Ação desconhecida.")
}

// openTable shuffles the played cards so nobody can tell whose is whose.
func (s *State) openTable(rng *rand.Rand) {
	s.Table = slices.Clone(s.Players)
	rng.Shuffle(len(s.Table), func(i, j int) { s.Table[i], s.Table[j] = s.Table[j], s.Table[i] })
	s.Phase = PhaseVote
}

// score hands out the round's points:
//   - everybody found the teller's card, or nobody did: the clue was too easy or too
//     hard, so the teller gets nothing and everyone else gets 2;
//   - otherwise the teller and each person who found the card get 3, and each card that
//     fooled people gives its owner 1 per vote, up to 3.
func (s *State) score() {
	teller := s.Teller()
	tellerPos := slices.Index(s.Table, teller)
	found := 0
	for _, pos := range s.Votes {
		if pos == tellerPos {
			found++
		}
	}
	voters := len(s.Players) - 1
	if found == 0 || found == voters {
		for _, p := range s.Players {
			if p != teller {
				s.Gains[p] = 2
			}
		}
	} else {
		s.Gains[teller] = 3
		fooled := map[string]int{}
		for v, pos := range s.Votes {
			if pos == tellerPos {
				s.Gains[v] += 3
			} else {
				fooled[s.Table[pos]]++
			}
		}
		for owner, n := range fooled {
			s.Gains[owner] += min(n, 3)
		}
	}
	for p, g := range s.Gains {
		s.Scores[p] += g
	}
	s.Phase = PhaseReveal
}

// Winners are the people with the best score.
func (s *State) Winners() []string {
	best := -1
	for _, p := range s.Players {
		best = max(best, s.Scores[p])
	}
	var w []string
	for _, p := range s.Players {
		if s.Scores[p] == best {
			w = append(w, p)
		}
	}
	return w
}
