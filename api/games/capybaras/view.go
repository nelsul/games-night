package capybaras

import (
	"slices"
	"time"
)

type PlayerView struct {
	ID    string `json:"id"`
	Cards int    `json:"cards"`
	Out   bool   `json:"out"`
	Turns int    `json:"turns"` // turns left, for the current player
}

type PendingView struct {
	Player   string    `json:"player"`
	Kinds    []string  `json:"kinds"`
	Effect   string    `json:"effect"`
	Target   string    `json:"target,omitempty"`
	Ask      string    `json:"ask,omitempty"`
	Nopes    int       `json:"nopes"`
	Deadline time.Time `json:"deadline"`
}

type Me struct {
	Hand    []Card `json:"hand"`
	Peek    []Card `json:"peek"` // what my Binóculo showed (top first), until the deck changes
	PeekSeq int    `json:"peek_seq"`
	Out     bool   `json:"out"`
	CanNope bool   `json:"can_nope"`
}

// View is what one player may see: their own hand, how many cards everyone else has,
// the deck's size (never its order), the discard pile and the table's story.
type View struct {
	Phase      string       `json:"phase"`
	Players    []PlayerView `json:"players"`
	Turn       string       `json:"turn"`
	TurnsLeft  int          `json:"turns_left"`
	Deck       int          `json:"deck"`
	Jaguars    int          `json:"jaguars"` // still in the deck (always one fewer than the players left)
	Discard    []Card       `json:"discard"` // newest last, up to 8
	DiscardLen int          `json:"discard_len"`
	Pending    *PendingView `json:"pending,omitempty"`
	Favor      *Favor       `json:"favor,omitempty"`
	Hiding     bool         `json:"hiding"` // the current player is hiding a Jaguar back
	Events     []Event      `json:"events"`
	Winner     string       `json:"winner,omitempty"`
	Out        []string     `json:"out"`
	ServerTime time.Time    `json:"server_time"` // for the Nem Vem countdown, whatever the phone's clock says
	Me         Me           `json:"me"`
	Kinds      []string     `json:"kinds"`
}

func ViewFor(s *State, p string, now time.Time) View {
	v := View{
		Phase:      s.Phase,
		Turn:       s.current(),
		TurnsLeft:  s.TurnsLeft,
		Deck:       len(s.Deck),
		DiscardLen: len(s.Discard),
		Discard:    nonNil(slices.Clone(s.Discard[max(0, len(s.Discard)-8):])),
		Favor:      s.Favor,
		Hiding:     s.Phase == PhaseDefuse,
		Winner:     s.Winner,
		Out:        append([]string{}, s.Out...),
		ServerTime: now,
		Kinds:      Kinds,
	}
	for _, c := range s.Deck {
		if c.Kind == Jaguar {
			v.Jaguars++
		}
	}
	for _, id := range s.Players {
		pv := PlayerView{ID: id, Cards: len(s.Hands[id]), Out: slices.Contains(s.Out, id)}
		if id == v.Turn {
			pv.Turns = s.TurnsLeft
		}
		v.Players = append(v.Players, pv)
	}
	if pd := s.Pending; pd != nil {
		v.Pending = &PendingView{Player: pd.Player, Kinds: kindsOf(pd.Cards), Effect: pd.Effect, Target: pd.Target, Ask: pd.Ask, Nopes: pd.Nopes, Deadline: pd.Deadline}
	}
	v.Events = []Event{}
	for _, e := range s.Events {
		if e.Private != nil && !slices.Contains(e.Private, p) {
			e.Card = nil
		}
		e.Private = nil
		v.Events = append(v.Events, e)
	}
	v.Me = Me{Hand: nonNil(slices.Clone(s.Hands[p])), Peek: nonNil(slices.Clone(s.Peeks[p])), PeekSeq: s.PeekSeq, Out: slices.Contains(s.Out, p)}
	v.Me.CanNope = s.Pending != nil && s.alive(p) && s.holds(p, Nope) >= 0
	return v
}
