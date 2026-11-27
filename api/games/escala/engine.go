// Package escala is a cooperative guessing game. Every round each person gets a secret
// number from 1 to 100 and a theme (a scale, like "size of animals"). Everyone gives a
// clue that says how big their number is on that scale, without saying the number, and
// the table sorts itself from smallest to largest using only the clues.
package escala

import (
	"math/rand/v2"
	"slices"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	MinPlayers    = 3
	MaxPlayers    = 10
	MaxClueRune   = 30
	DefaultRounds = 3
	// MaxSwapProposals is how many times per round the table may ask to change the theme.
	MaxSwapProposals = 3
)

// Phases of a round.
const (
	PhaseClues  = "clues"  // everyone writes a clue
	PhaseOrder  = "order"  // the table sorts the cards
	PhaseReveal = "reveal" // numbers are shown
	PhaseOver   = "over"
)

// RuleError is a broken rule; its message is shown to players (Portuguese).
type RuleError struct{ Msg string }

func (e *RuleError) Error() string { return e.Msg }

func ruleErr(msg string) error { return &RuleError{Msg: msg} }

// Options are the host's lobby settings.
type Options struct {
	Rounds int `json:"rounds"` // 3, 5 or 7
}

func (o Options) rounds() int {
	if o.Rounds == 0 {
		return DefaultRounds
	}
	return o.Rounds
}

// CheckLobby validates options while people are still joining.
func (o Options) CheckLobby() error {
	switch o.Rounds {
	case 0, 3, 5, 7:
		return nil
	}
	return ruleErr("Escolha 3, 5 ou 7 rodadas.")
}

// Result is how a finished round went.
type Result struct {
	Theme Theme `json:"theme"`
	OK    bool  `json:"ok"`
}

// SwapVote is an open vote to change the round's theme. The proposer's vote is a yes.
type SwapVote struct {
	Proposer string          `json:"proposer"`
	Votes    map[string]bool `json:"votes"`
}

// State is a whole match, secrets included. It never leaves the server as is.
type State struct {
	Players []string          `json:"players"`
	Rounds  int               `json:"rounds"`
	Round   int               `json:"round"` // 0-based
	Phase   string            `json:"phase"`
	Theme   Theme             `json:"theme"`
	Numbers map[string]int    `json:"numbers"`
	Clues   map[string]string `json:"clues"`
	Line    []string          `json:"line"` // placed cards, left (smallest) to right
	Used    []int             `json:"used"` // theme indexes already played
	Results []Result          `json:"results"`
	Swap    *SwapVote         `json:"swap,omitempty"` // an open vote to change the theme
	Swaps   int               `json:"swaps"`          // proposals made this round
	SwapSeq int               `json:"swap_seq"`       // counts closed votes, so phones can announce them
	SwapOK  bool              `json:"swap_ok"`        // how the last closed vote ended
}

// Action is what a player sends.
type Action struct {
	Type  string `json:"type"` // clue | place | reveal | next | swap_propose | swap_vote
	Yes   bool   `json:"yes,omitempty"`
	Text  string `json:"text,omitempty"`
	Index int    `json:"index,omitempty"`
}

// Setup deals the first round.
func Setup(players []string, o Options, rng *rand.Rand, _ time.Time) (*State, error) {
	if len(players) < MinPlayers || len(players) > MaxPlayers {
		return nil, ruleErr("Escala é para 3 a 10 pessoas.")
	}
	if err := o.CheckLobby(); err != nil {
		return nil, err
	}
	s := &State{Players: slices.Clone(players), Rounds: o.rounds()}
	s.deal(rng)
	return s, nil
}

// pickTheme draws a theme that has not been played in this match yet.
func (s *State) pickTheme(rng *rand.Rand) {
	var free []int
	for i := range themes {
		if !slices.Contains(s.Used, i) {
			free = append(free, i)
		}
	}
	if len(free) == 0 {
		s.Used = nil
		for i := range themes {
			free = append(free, i)
		}
	}
	t := free[rng.IntN(len(free))]
	s.Used = append(s.Used, t)
	s.Theme = themes[t]
}

// deal starts a round: a new theme and a new secret number for everyone.
func (s *State) deal(rng *rand.Rand) {
	s.pickTheme(rng)
	nums := rng.Perm(100) // unique numbers, so there is always one right order
	s.Numbers = map[string]int{}
	for i, p := range s.Players {
		s.Numbers[p] = nums[i] + 1
	}
	s.Clues = map[string]string{}
	s.Line = nil
	s.Swap, s.Swaps = nil, 0
	s.Phase = PhaseClues
}

func (s *State) has(user string) bool { return slices.Contains(s.Players, user) }

// Apply runs one player's action.
func Apply(s *State, user string, a Action, rng *rand.Rand) error {
	if !s.has(user) {
		return ruleErr("Você não está nessa partida.")
	}
	switch a.Type {
	case "clue":
		// A clue can be changed until the numbers are revealed.
		if s.Phase != PhaseClues && s.Phase != PhaseOrder {
			return ruleErr("Não é hora de dar dicas.")
		}
		text := strings.TrimSpace(a.Text)
		if text == "" {
			return ruleErr("Escreva uma dica.")
		}
		if utf8.RuneCountInString(text) > MaxClueRune {
			return ruleErr("A dica pode ter até 30 letras.")
		}
		s.Clues[user] = text
		if s.Phase == PhaseClues && len(s.Clues) == len(s.Players) {
			s.Phase = PhaseOrder
		}
	case "place":
		if s.Phase != PhaseOrder {
			return ruleErr("Não é hora de ordenar.")
		}
		// Take the card out first, so moving it never counts its own old spot.
		line := slices.DeleteFunc(slices.Clone(s.Line), func(p string) bool { return p == user })
		if a.Index < 0 || a.Index > len(line) {
			return ruleErr("Posição inválida.")
		}
		s.Line = slices.Insert(line, a.Index, user)
	case "reveal":
		if s.Phase != PhaseOrder {
			return ruleErr("Ainda não dá para revelar.")
		}
		if s.Swap != nil {
			return ruleErr("Tem uma votação para trocar o tema aberta. Votem antes de revelar.")
		}
		if len(s.Line) != len(s.Players) {
			return ruleErr("Todo mundo precisa colocar sua carta antes.")
		}
		s.Results = append(s.Results, Result{Theme: s.Theme, OK: s.sorted()})
		s.Phase = PhaseReveal
	case "swap_propose":
		if s.Phase != PhaseClues && s.Phase != PhaseOrder {
			return ruleErr("Agora não dá para trocar o tema.")
		}
		if s.Swap != nil {
			return ruleErr("Já tem uma votação aberta.")
		}
		if s.Swaps >= MaxSwapProposals {
			return ruleErr("O grupo já pediu troca de tema vezes demais nesta rodada.")
		}
		s.Swaps++
		s.Swap = &SwapVote{Proposer: user, Votes: map[string]bool{user: true}}
		s.settleSwap(rng)
	case "swap_vote":
		if s.Swap == nil {
			return ruleErr("Não tem votação aberta.")
		}
		s.Swap.Votes[user] = a.Yes
		s.settleSwap(rng)
	case "next":
		if s.Phase != PhaseReveal {
			return ruleErr("A rodada ainda não terminou.")
		}
		s.Round++
		if s.Round >= s.Rounds {
			s.Phase = PhaseOver
			return nil
		}
		s.deal(rng)
	default:
		return ruleErr("Ação inválida.")
	}
	return nil
}

// settleSwap closes the open vote once its result cannot change: a majority of yes
// swaps the theme (and the table writes new clues); no yes majority possible, or a tie,
// keeps it. Numbers stay as they are.
func (s *State) settleSwap(rng *rand.Rand) {
	yes, no := 0, 0
	for _, v := range s.Swap.Votes {
		if v {
			yes++
		} else {
			no++
		}
	}
	n := len(s.Players)
	switch {
	case yes*2 > n:
		s.pickTheme(rng)
		s.Clues = map[string]string{}
		s.Line = nil
		s.Phase = PhaseClues
		s.Swap = nil
		s.SwapSeq++
		s.SwapOK = true
	case no*2 >= n:
		s.Swap = nil
		s.SwapSeq++
		s.SwapOK = false
	}
}

// Score is how many rounds the table sorted right. It is a team result: nobody loses.
func (s *State) Score() int {
	n := 0
	for _, r := range s.Results {
		if r.OK {
			n++
		}
	}
	return n
}

// sorted reports whether the line runs from the smallest number to the largest.
func (s *State) sorted() bool {
	for i := 1; i < len(s.Line); i++ {
		if s.Numbers[s.Line[i-1]] > s.Numbers[s.Line[i]] {
			return false
		}
	}
	return true
}

// Misplaced marks every card that is smaller than a card on its left.
func (s *State) Misplaced() map[string]bool {
	out := map[string]bool{}
	highest := 0
	for _, p := range s.Line {
		if n := s.Numbers[p]; n < highest {
			out[p] = true
		} else {
			highest = n
		}
	}
	return out
}
