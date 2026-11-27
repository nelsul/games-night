package capybaras

import (
	"math/rand/v2"
	"slices"
	"time"
)

// Bot strategy: only what a person in the bot's seat could know (its hand, what its
// Binóculo showed, card counts and the table's story), with some randomness. It plays
// a few cards to stay safe and to annoy whoever leads, then draws.

// BotAction is what bot wants to do now, or nil when it is waiting on others.
func BotAction(s *State, bot string, now time.Time, rng *rand.Rand) *Action {
	if s.Phase == PhaseOver || !slices.Contains(s.Players, bot) {
		return nil
	}
	if pd := s.Pending; pd != nil {
		if !now.Before(pd.Deadline) {
			return &Action{Type: "resolve"}
		}
		if s.alive(bot) && s.holds(bot, Nope) >= 0 && botWantsNope(s, bot, rng) {
			return &Action{Type: "nope"}
		}
		return nil
	}
	if !s.alive(bot) {
		return nil
	}
	switch s.Phase {
	case PhaseFavor:
		if s.Favor.From == bot {
			return &Action{Type: "give", Cards: []int{cheapest(s.Hands[bot]).ID}}
		}
		return nil
	case PhaseDefuse:
		if s.current() == bot {
			// Near the top hurts the next player; sometimes deeper, to keep them guessing.
			i := min(len(s.Deck), rng.IntN(3))
			if rng.Float64() < 0.3 {
				i = rng.IntN(len(s.Deck) + 1)
			}
			return &Action{Type: "hide", Index: i}
		}
		return nil
	}
	if s.current() != bot {
		return nil
	}
	return botTurn(s, bot, rng)
}

// botWantsNope: cancel what hurts the bot; save its own card when someone cancels it.
func botWantsNope(s *State, bot string, rng *rand.Rand) bool {
	pd := s.Pending
	mine := pd.Player == bot
	if mine {
		return pd.Canceled() && rng.Float64() < 0.6
	}
	if pd.Canceled() {
		return false // already canceled: let it be
	}
	hurtsMe := pd.Target == bot || (pd.Effect == Current && s.Players[s.nextAlive(s.Turn)] == bot)
	if hurtsMe {
		return rng.Float64() < 0.75
	}
	return rng.Float64() < 0.08
}

func botTurn(s *State, bot string, rng *rand.Rand) *Action {
	hand := s.Hands[bot]
	find := func(kind string) *Card {
		for i := range hand {
			if hand[i].Kind == kind {
				return &hand[i]
			}
		}
		return nil
	}
	hasDive := find(Dive) != nil
	peek := s.Peeks[bot]
	topIsJaguar := len(peek) > 0 && peek[0].Kind == Jaguar
	jaguars := 0
	for _, c := range s.Deck {
		if c.Kind == Jaguar {
			jaguars++
		}
	}
	risk := float64(jaguars) / float64(max(1, len(s.Deck)))
	escape := func() *Action {
		for _, k := range []string{Current, Nap, Swirl} {
			if c := find(k); c != nil {
				return &Action{Type: "play", Cards: []int{c.ID}}
			}
		}
		return nil
	}
	if topIsJaguar {
		if a := escape(); a != nil {
			return a
		}
	}
	if !hasDive && risk > 0.12 && len(peek) == 0 {
		if c := find(Peek); c != nil && rng.Float64() < 0.7 {
			return &Action{Type: "play", Cards: []int{c.ID}}
		}
		if rng.Float64() < risk*2.5 {
			if a := escape(); a != nil {
				return a
			}
		}
	}
	// Pairs and trios steal from whoever holds the most cards.
	if rng.Float64() < 0.55 {
		target := richest(s, bot)
		if target != "" {
			for _, k := range Capys {
				var ids []int
				for _, c := range hand {
					if c.Kind == k {
						ids = append(ids, c.ID)
					}
				}
				if len(ids) >= 3 && !hasDive {
					return &Action{Type: "play", Cards: ids[:3], Target: target, Ask: Dive}
				}
				if len(ids) >= 2 {
					return &Action{Type: "play", Cards: ids[:2], Target: target}
				}
			}
			if c := find(Cuddle); c != nil && rng.Float64() < 0.4 {
				return &Action{Type: "play", Cards: []int{c.ID}, Target: target}
			}
		}
	}
	return &Action{Type: "draw"}
}

func richest(s *State, bot string) string {
	best, n := "", 0
	for _, p := range s.Players {
		if p != bot && s.alive(p) && len(s.Hands[p]) > n {
			best, n = p, len(s.Hands[p])
		}
	}
	return best
}

// cheapest is the card a bot minds giving away the least.
func cheapest(hand []Card) Card {
	rank := map[string]int{Dive: 9, Nope: 7, Current: 6, Nap: 6, Swirl: 5, Peek: 5, Cuddle: 4}
	best := hand[0]
	for _, c := range hand[1:] {
		if rank[c.Kind] < rank[best.Kind] {
			best = c
		}
	}
	return best
}
