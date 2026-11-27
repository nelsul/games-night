// Package capybaras is the rules engine of Crazy Capybaras, a card game of nerve: on
// your turn you play as many cards as you like and then draw one. Somewhere in the
// deck there are jaguars (onças). Draw one without a Dive (mergulho) in your hand and
// you are out. The last capybara standing wins.
//
// The engine is pure: Setup deals a match, Apply validates and applies one action and
// ViewFor builds what one player may see. Every rule is tested in engine_test.go.
package capybaras

import (
	"math/rand/v2"
	"slices"
	"time"
)

const (
	MinPlayers = 2
	MaxPlayers = 8
	HandSize   = 7 // cards dealt besides the Dive
	// ReactWindow is how long the table has to answer a played card with a Nem Vem.
	// Every Nem Vem gives everyone this long again.
	ReactWindow = 3 * time.Second
	keepEvents  = 40
)

// Card kinds.
const (
	Jaguar  = "jaguar"  // Onça: draw it without a Dive and you're out
	Dive    = "dive"    // Mergulho: saves you from a Jaguar
	Nope    = "nope"    // Nem Vem: cancels the card just played (and Nem Vem cancels Nem Vem)
	Current = "current" // Correnteza: end your turn without drawing; the next player takes 2 turns
	Nap     = "nap"     // Soneca: end one turn without drawing
	Peek    = "peek"    // Binóculo: see the top 3 cards
	Swirl   = "swirl"   // Redemoinho: shuffle the deck
	Cuddle  = "cuddle"  // Cafuné: someone gives you a card of their choice
)

// Capybara cards do nothing alone: a pair steals a random card, a trio asks for one.
var Capys = []string{"orange", "shades", "dj", "spa", "surf"}

func isCapy(kind string) bool { return slices.Contains(Capys, kind) }

// Kinds lists every card kind (for asking with a trio).
var Kinds = append([]string{Jaguar, Dive, Nope, Current, Nap, Peek, Swirl, Cuddle}, Capys...)

// Phases.
const (
	PhasePlay   = "play"   // the current player plays cards or draws
	PhaseDefuse = "defuse" // the current player dove away from a Jaguar and hides it back
	PhaseFavor  = "favor"  // someone picks the card they give after a Cafuné
	PhaseOver   = "over"
)

type Card struct {
	ID   int    `json:"id"`
	Kind string `json:"kind"`
}

// Pending is a card (or combo) waiting out the Nem Vem window before it takes effect.
type Pending struct {
	Player   string    `json:"player"`
	Cards    []Card    `json:"cards"`
	Effect   string    `json:"effect"` // nap, current, peek, swirl, cuddle, pair, trio
	Target   string    `json:"target,omitempty"`
	Ask      string    `json:"ask,omitempty"` // the kind asked for with a trio
	Nopes    int       `json:"nopes"`
	Deadline time.Time `json:"deadline"`
}

func (p *Pending) Canceled() bool { return p.Nopes%2 == 1 }

// Event is one line of the table's story. Card is only shown to the players in Private
// when Private is set (a stolen or given card); everyone else sees that it happened.
type Event struct {
	Seq     int      `json:"seq"`
	Type    string   `json:"type"`
	Player  string   `json:"player,omitempty"`
	Target  string   `json:"target,omitempty"`
	Kinds   []string `json:"kinds,omitempty"`
	Card    *Card    `json:"card,omitempty"`
	Private []string `json:"private,omitempty"`
	OK      bool     `json:"ok,omitempty"`
}

type Favor struct {
	From string `json:"from"` // gives a card
	To   string `json:"to"`
}

// State is a whole match, stored as JSON. Hands and the deck's order are secret.
type State struct {
	Players   []string          `json:"players"` // seat order
	Hands     map[string][]Card `json:"hands"`
	Deck      []Card            `json:"deck"` // [0] is the top
	Discard   []Card            `json:"discard"`
	Out       []string          `json:"out"` // eaten, in order
	Turn      int               `json:"turn"`
	TurnsLeft int               `json:"turns_left"`
	Phase     string            `json:"phase"`
	Pending   *Pending          `json:"pending,omitempty"`
	Favor     *Favor            `json:"favor,omitempty"`
	Hiding    *Card             `json:"hiding,omitempty"` // the Jaguar being hidden back
	Peeks     map[string][]Card `json:"peeks"`            // what Binóculo showed, per player
	PeekSeq   int               `json:"peek_seq"`
	Events    []Event           `json:"events"`
	Seq       int               `json:"seq"`
	Winner    string            `json:"winner,omitempty"`
}

// RuleError is a broken rule; its message is shown to players (Portuguese).
type RuleError struct{ Msg string }

func (e *RuleError) Error() string { return e.Msg }

func ruleErr(msg string) error { return &RuleError{Msg: msg} }

// deckCounts is how many of each card go in, scaled up for big tables.
func deckCounts(players int) (counts map[string]int, extraDives int) {
	f := 1
	if players > 5 {
		f = 2
	}
	counts = map[string]int{Nope: 5 * f, Current: 4 * f, Nap: 4 * f, Peek: 5 * f, Swirl: 4 * f, Cuddle: 4 * f}
	for _, c := range Capys {
		counts[c] = 4 * f
	}
	switch {
	case players <= 3:
		extraDives = 2
	case players <= 5:
		extraDives = 6 - players
	default:
		extraDives = 2
	}
	return counts, extraDives
}

// Setup deals: everyone gets 7 cards and a Dive, then the Jaguars (one fewer than the
// players) and the spare Dives are shuffled into the deck. A random player starts.
func Setup(players []string, rng *rand.Rand, _ time.Time) (*State, error) {
	if len(players) < MinPlayers || len(players) > MaxPlayers {
		return nil, ruleErr("O Crazy Capybaras é para 2 a 8 pessoas.")
	}
	s := &State{
		Players: slices.Clone(players),
		Hands:   map[string][]Card{},
		Peeks:   map[string][]Card{},
		Phase:   PhasePlay,
	}
	id := 0
	card := func(kind string) Card { id++; return Card{ID: id, Kind: kind} }
	counts, extra := deckCounts(len(players))
	var deck []Card
	for _, kind := range Kinds {
		for range counts[kind] {
			deck = append(deck, card(kind))
		}
	}
	shuffle(deck, rng)
	for _, p := range players {
		hand := slices.Clone(deck[:HandSize])
		deck = deck[HandSize:]
		s.Hands[p] = append(hand, card(Dive))
	}
	for range extra {
		deck = append(deck, card(Dive))
	}
	for range len(players) - 1 {
		deck = append(deck, card(Jaguar))
	}
	shuffle(deck, rng)
	s.Deck = deck
	s.Turn = rng.IntN(len(players))
	s.TurnsLeft = 1
	s.log(Event{Type: "start", Player: s.current()})
	return s, nil
}

func shuffle(c []Card, rng *rand.Rand) {
	rng.Shuffle(len(c), func(i, j int) { c[i], c[j] = c[j], c[i] })
}

func (s *State) current() string { return s.Players[s.Turn] }

func (s *State) alive(p string) bool {
	return slices.Contains(s.Players, p) && !slices.Contains(s.Out, p)
}

func (s *State) aliveCount() int { return len(s.Players) - len(s.Out) }

// nextAlive is the seat after i that is still in the game.
func (s *State) nextAlive(i int) int {
	for k := 1; k <= len(s.Players); k++ {
		j := (i + k) % len(s.Players)
		if s.alive(s.Players[j]) {
			return j
		}
	}
	return i
}

func (s *State) log(e Event) {
	s.Seq++
	e.Seq = s.Seq
	s.Events = append(s.Events, e)
	if len(s.Events) > keepEvents {
		s.Events = s.Events[len(s.Events)-keepEvents:]
	}
}

// take removes the cards with these ids from p's hand and returns them, or nil if p
// doesn't hold them all.
func (s *State) take(p string, ids []int) []Card {
	hand := s.Hands[p]
	var out []Card
	var rest []Card
	for _, c := range hand {
		if slices.Contains(ids, c.ID) {
			out = append(out, c)
		} else {
			rest = append(rest, c)
		}
	}
	if len(out) != len(ids) || hasDup(ids) {
		return nil
	}
	s.Hands[p] = nonNil(rest)
	return out
}

func hasDup(ids []int) bool {
	seen := map[int]bool{}
	for _, id := range ids {
		if seen[id] {
			return true
		}
		seen[id] = true
	}
	return false
}

func nonNil(c []Card) []Card {
	if c == nil {
		return []Card{}
	}
	return c
}

func (s *State) holds(p, kind string) int {
	for i, c := range s.Hands[p] {
		if c.Kind == kind {
			return i
		}
	}
	return -1
}

// Action is one thing a player does.
type Action struct {
	Type   string `json:"type"` // play, nope, resolve, draw, hide, give
	Cards  []int  `json:"cards,omitempty"`
	Target string `json:"target,omitempty"`
	Ask    string `json:"ask,omitempty"`
	Index  int    `json:"index,omitempty"`
}

// Apply checks an action against the rules and applies it. On error nothing changes.
func Apply(s *State, p string, a Action, now time.Time, rng *rand.Rand) error {
	if !slices.Contains(s.Players, p) {
		return ruleErr("Você não está nessa partida.")
	}
	if s.Phase == PhaseOver {
		return ruleErr("A partida acabou.")
	}
	switch a.Type {
	case "play":
		return s.play(p, a, now)
	case "nope":
		return s.nope(p, a, now)
	case "resolve":
		return s.resolve(now, rng)
	case "draw":
		return s.draw(p)
	case "hide":
		return s.hide(p, a.Index)
	case "give":
		return s.give(p, a.Cards)
	}
	return ruleErr("Ação desconhecida.")
}

func (s *State) myTurn(p string) error {
	switch {
	case !s.alive(p):
		return ruleErr("A onça já te pegou. Agora é só torcer.")
	case p != s.current():
		return ruleErr("Não é a sua vez.")
	case s.Phase != PhasePlay:
		return ruleErr("Espere a jogada atual terminar.")
	case s.Pending != nil:
		return ruleErr("Espere a carta da mesa fazer efeito.")
	}
	return nil
}

func (s *State) play(p string, a Action, now time.Time) error {
	if err := s.myTurn(p); err != nil {
		return err
	}
	if len(a.Cards) == 0 || len(a.Cards) > 3 {
		return ruleErr("Escolha uma carta, um par ou uma trinca.")
	}
	var kinds []string
	for _, id := range a.Cards {
		i := slices.IndexFunc(s.Hands[p], func(c Card) bool { return c.ID == id })
		if i < 0 {
			return ruleErr("Você não tem essa carta.")
		}
		kinds = append(kinds, s.Hands[p][i].Kind)
	}
	if hasDup(a.Cards) {
		return ruleErr("Escolha cartas diferentes.")
	}
	pend := &Pending{Player: p, Deadline: now.Add(ReactWindow)}
	needsTarget := false
	switch {
	case len(kinds) == 1 && slices.Contains([]string{Current, Nap, Peek, Swirl, Cuddle}, kinds[0]):
		pend.Effect = kinds[0]
		needsTarget = kinds[0] == Cuddle
	case len(kinds) == 1 && kinds[0] == Nope:
		return ruleErr("O Nem Vem só serve para cancelar uma carta que alguém acabou de jogar.")
	case len(kinds) == 1 && (kinds[0] == Dive || kinds[0] == Jaguar):
		return ruleErr("O Mergulho só é usado quando a onça aparece.")
	case len(kinds) == 1:
		return ruleErr("Capivaras sozinhas não fazem nada. Junte um par ou uma trinca iguais.")
	case len(kinds) >= 2 && isCapy(kinds[0]) && !slices.ContainsFunc(kinds, func(k string) bool { return k != kinds[0] }):
		pend.Effect = map[int]string{2: "pair", 3: "trio"}[len(kinds)]
		needsTarget = true
	default:
		return ruleErr("Pares e trincas são de capivaras iguais.")
	}
	if needsTarget {
		if a.Target == p || !s.alive(a.Target) {
			return ruleErr("Escolha outra capivara que ainda está no jogo.")
		}
		if len(s.Hands[a.Target]) == 0 {
			return ruleErr("Essa capivara está sem cartas.")
		}
		pend.Target = a.Target
	}
	if pend.Effect == "trio" {
		if !slices.Contains(Kinds, a.Ask) {
			return ruleErr("Diga qual carta você quer.")
		}
		pend.Ask = a.Ask
	}
	pend.Cards = s.take(p, a.Cards)
	s.Discard = append(s.Discard, pend.Cards...)
	s.Pending = pend
	s.log(Event{Type: "play", Player: p, Target: pend.Target, Kinds: kinds, Card: askCard(pend.Ask)})
	return nil
}

// askCard carries the asked kind in an event without any id.
func askCard(kind string) *Card {
	if kind == "" {
		return nil
	}
	return &Card{Kind: kind}
}

func (s *State) nope(p string, a Action, now time.Time) error {
	switch {
	case s.Pending == nil:
		return ruleErr("Não tem nada para cancelar.")
	case !now.Before(s.Pending.Deadline):
		return ruleErr("Tarde demais: a carta já fez efeito.")
	case !s.alive(p):
		return ruleErr("A onça já te pegou. Agora é só torcer.")
	}
	ids := a.Cards
	if len(ids) == 0 {
		if i := s.holds(p, Nope); i >= 0 {
			ids = []int{s.Hands[p][i].ID}
		}
	}
	if len(ids) != 1 {
		return ruleErr("Você não tem um Nem Vem.")
	}
	i := slices.IndexFunc(s.Hands[p], func(c Card) bool { return c.ID == ids[0] })
	if i < 0 || s.Hands[p][i].Kind != Nope {
		return ruleErr("Você não tem um Nem Vem.")
	}
	c := s.take(p, ids)
	s.Discard = append(s.Discard, c...)
	s.Pending.Nopes++
	s.Pending.Deadline = now.Add(ReactWindow)
	s.log(Event{Type: "nope", Player: p, Target: s.Pending.Player, OK: s.Pending.Canceled()})
	return nil
}

// resolve applies the pending card once nobody answered it in time. Any player may ask
// for it (every phone does when its countdown ends).
func (s *State) resolve(now time.Time, rng *rand.Rand) error {
	pend := s.Pending
	switch {
	case pend == nil:
		return ruleErr("Não tem carta esperando.")
	case now.Before(pend.Deadline):
		return ruleErr("Ainda dá tempo de um Nem Vem.")
	}
	s.Pending = nil
	p := pend.Player
	if pend.Canceled() {
		s.log(Event{Type: "canceled", Player: p, Kinds: kindsOf(pend.Cards)})
		return nil
	}
	switch pend.Effect {
	case Nap:
		s.log(Event{Type: "nap", Player: p})
		s.endTurn()
	case Current:
		next := s.nextAlive(s.Turn)
		s.Turn, s.TurnsLeft = next, 2
		s.log(Event{Type: "current", Player: p, Target: s.current()})
	case Peek:
		s.Peeks[p] = slices.Clone(s.Deck[:min(3, len(s.Deck))])
		s.PeekSeq++
		s.log(Event{Type: "peek", Player: p})
	case Swirl:
		shuffle(s.Deck, rng)
		clear(s.Peeks)
		s.log(Event{Type: "swirl", Player: p})
	case Cuddle:
		if len(s.Hands[pend.Target]) == 0 {
			s.log(Event{Type: "favor", Player: pend.Target, Target: p})
			return nil
		}
		s.Phase = PhaseFavor
		s.Favor = &Favor{From: pend.Target, To: p}
		s.log(Event{Type: "cuddle", Player: p, Target: pend.Target})
	case "pair":
		hand := s.Hands[pend.Target]
		if len(hand) == 0 {
			s.log(Event{Type: "steal", Player: p, Target: pend.Target})
			return nil
		}
		c := hand[rng.IntN(len(hand))]
		s.move(pend.Target, p, c)
		s.log(Event{Type: "steal", Player: p, Target: pend.Target, Card: &c, Private: []string{p, pend.Target}, OK: true})
	case "trio":
		i := s.holds(pend.Target, pend.Ask)
		if i < 0 {
			s.log(Event{Type: "ask", Player: p, Target: pend.Target, Card: askCard(pend.Ask)})
			return nil
		}
		c := s.Hands[pend.Target][i]
		s.move(pend.Target, p, c)
		s.log(Event{Type: "ask", Player: p, Target: pend.Target, Card: &c, OK: true})
	}
	return nil
}

func kindsOf(cards []Card) []string {
	out := make([]string, len(cards))
	for i, c := range cards {
		out[i] = c.Kind
	}
	return out
}

func (s *State) move(from, to string, c Card) {
	s.Hands[from] = nonNil(slices.DeleteFunc(slices.Clone(s.Hands[from]), func(x Card) bool { return x.ID == c.ID }))
	s.Hands[to] = append(s.Hands[to], c)
}

// endTurn uses up one of the current player's turns and passes on when none are left.
func (s *State) endTurn() {
	s.TurnsLeft--
	if s.TurnsLeft <= 0 {
		s.Turn = s.nextAlive(s.Turn)
		s.TurnsLeft = 1
	}
}

func (s *State) draw(p string) error {
	if err := s.myTurn(p); err != nil {
		return err
	}
	if len(s.Deck) == 0 {
		return ruleErr("O monte acabou.")
	}
	c := s.Deck[0]
	s.Deck = s.Deck[1:]
	clear(s.Peeks)
	if c.Kind != Jaguar {
		s.Hands[p] = append(s.Hands[p], c)
		s.log(Event{Type: "draw", Player: p, Card: &c, Private: []string{p}})
		s.endTurn()
		return nil
	}
	if i := s.holds(p, Dive); i >= 0 {
		d := s.take(p, []int{s.Hands[p][i].ID})
		s.Discard = append(s.Discard, d...)
		s.Hiding = &c
		s.Phase = PhaseDefuse
		s.log(Event{Type: "dive", Player: p})
		return nil
	}
	// Eaten: the hand and the Jaguar leave the game.
	s.Discard = append(s.Discard, s.Hands[p]...)
	s.Discard = append(s.Discard, c)
	s.Hands[p] = []Card{}
	s.Out = append(s.Out, p)
	s.log(Event{Type: "eaten", Player: p})
	if s.aliveCount() == 1 {
		s.Winner = s.Players[s.nextAlive(s.Turn)]
		s.Phase = PhaseOver
		s.log(Event{Type: "win", Player: s.Winner})
		return nil
	}
	s.Turn = s.nextAlive(s.Turn)
	s.TurnsLeft = 1
	return nil
}

// hide puts the Jaguar back in the deck at index (0 = on top), in secret.
func (s *State) hide(p string, index int) error {
	switch {
	case s.Phase != PhaseDefuse || p != s.current():
		return ruleErr("Agora não é hora de esconder a onça.")
	case index < 0 || index > len(s.Deck):
		return ruleErr("Escolha um lugar no monte.")
	}
	s.Deck = slices.Insert(s.Deck, index, *s.Hiding)
	s.Hiding = nil
	s.Phase = PhasePlay
	clear(s.Peeks)
	s.log(Event{Type: "hide", Player: p})
	s.endTurn()
	return nil
}

func (s *State) give(p string, ids []int) error {
	if s.Phase != PhaseFavor || s.Favor == nil || p != s.Favor.From {
		return ruleErr("Ninguém te pediu um cafuné agora.")
	}
	if len(ids) != 1 {
		return ruleErr("Escolha uma carta para dar.")
	}
	i := slices.IndexFunc(s.Hands[p], func(c Card) bool { return c.ID == ids[0] })
	if i < 0 {
		return ruleErr("Você não tem essa carta.")
	}
	c := s.Hands[p][i]
	to := s.Favor.To
	s.move(p, to, c)
	s.Favor = nil
	s.Phase = PhasePlay
	s.log(Event{Type: "give", Player: p, Target: to, Card: &c, Private: []string{p, to}})
	return nil
}
