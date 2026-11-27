package capybaras

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"slices"
	"strings"
	"testing"
	"time"
)

var t0 = time.Date(2026, 11, 27, 21, 0, 0, 0, time.UTC)

func players(n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = fmt.Sprintf("p%d", i)
	}
	return out
}

func setup(t *testing.T, n int, seed uint64) *State {
	t.Helper()
	s, err := Setup(players(n), rand.New(rand.NewPCG(seed, seed)), t0)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func rng() *rand.Rand { return rand.New(rand.NewPCG(1, 2)) }

func must(t *testing.T, s *State, p string, a Action, now time.Time) {
	t.Helper()
	if err := Apply(s, p, a, now, rng()); err != nil {
		t.Fatalf("%s %+v: %v", p, a, err)
	}
}

func mustFail(t *testing.T, s *State, p string, a Action, now time.Time) {
	t.Helper()
	before, _ := json.Marshal(s)
	if err := Apply(s, p, a, now, rng()); err == nil {
		t.Fatalf("%s %+v: expected an error", p, a)
	}
	after, _ := json.Marshal(s)
	if string(before) != string(after) {
		t.Fatalf("%s %+v: a refused action changed the state", p, a)
	}
}

// give hands p a fresh card of kind and returns it.
func give(s *State, p, kind string) Card {
	c := Card{ID: 10000 + len(s.Hands[p]) + 100*len(s.Events) + s.Seq*7, Kind: kind}
	for slices.ContainsFunc(s.Hands[p], func(x Card) bool { return x.ID == c.ID }) {
		c.ID++
	}
	s.Hands[p] = append(s.Hands[p], c)
	return c
}

func count(cards []Card, kind string) int {
	n := 0
	for _, c := range cards {
		if c.Kind == kind {
			n++
		}
	}
	return n
}

func TestSetupDealsEveryone(t *testing.T) {
	for n := MinPlayers; n <= MaxPlayers; n++ {
		s := setup(t, n, uint64(n))
		ids := map[int]bool{}
		all := slices.Clone(s.Deck)
		for _, p := range s.Players {
			h := s.Hands[p]
			if len(h) != HandSize+1 || count(h, Dive) < 1 || count(h, Jaguar) != 0 {
				t.Fatalf("n=%d %s hand: %v", n, p, h)
			}
			all = append(all, h...)
		}
		for _, c := range all {
			if ids[c.ID] {
				t.Fatalf("n=%d: card id %d twice", n, c.ID)
			}
			ids[c.ID] = true
		}
		if got := count(s.Deck, Jaguar); got != n-1 {
			t.Fatalf("n=%d: %d jaguars", n, got)
		}
		if len(s.Deck) < 2*n {
			t.Fatalf("n=%d: deck too small (%d)", n, len(s.Deck))
		}
	}
	if _, err := Setup(players(1), rng(), t0); err == nil {
		t.Fatal("a table of one")
	}
	if _, err := Setup(players(9), rng(), t0); err == nil {
		t.Fatal("a table of nine")
	}
}

func TestOnlyTheCurrentPlayerPlaysAndDraws(t *testing.T) {
	s := setup(t, 3, 1)
	cur := s.current()
	other := s.Players[s.nextAlive(s.Turn)]
	mustFail(t, s, other, Action{Type: "draw"}, t0)
	c := give(s, other, Nap)
	mustFail(t, s, other, Action{Type: "play", Cards: []int{c.ID}}, t0)
	mustFail(t, s, "stranger", Action{Type: "draw"}, t0)
	mine := give(s, cur, Nope)
	mustFail(t, s, cur, Action{Type: "play", Cards: []int{mine.ID}}, t0) // Nem Vem alone
	capy := give(s, cur, "orange")
	mustFail(t, s, cur, Action{Type: "play", Cards: []int{capy.ID}}, t0) // a lone capybara
	other2 := give(s, cur, "dj")
	mustFail(t, s, cur, Action{Type: "play", Cards: []int{capy.ID, other2.ID}, Target: other}, t0) // not a pair
	mustFail(t, s, cur, Action{Type: "play", Cards: []int{999999}}, t0)                            // not in hand
}

func TestNemVemWindowAndChains(t *testing.T) {
	s := setup(t, 3, 2)
	cur := s.current()
	a := s.Players[s.nextAlive(s.Turn)]
	b := s.Players[s.nextAlive(s.nextAlive(s.Turn))]
	nap := give(s, cur, Nap)
	must(t, s, cur, Action{Type: "play", Cards: []int{nap.ID}}, t0)
	mustFail(t, s, cur, Action{Type: "draw"}, t0.Add(time.Second))    // wait for the card
	mustFail(t, s, a, Action{Type: "resolve"}, t0.Add(2*time.Second)) // window still open
	give(s, a, Nope)
	give(s, b, Nope)
	must(t, s, a, Action{Type: "nope"}, t0.Add(2*time.Second))
	if !s.Pending.Canceled() || !s.Pending.Deadline.Equal(t0.Add(2*time.Second+ReactWindow)) {
		t.Fatalf("after one Nem Vem: %+v", s.Pending)
	}
	must(t, s, b, Action{Type: "nope"}, t0.Add(4*time.Second)) // Nem Vem on the Nem Vem
	mustFail(t, s, a, Action{Type: "resolve"}, t0.Add(6*time.Second))
	must(t, s, a, Action{Type: "resolve"}, t0.Add(4*time.Second+ReactWindow))
	// Not canceled in the end: the nap passed the turn without drawing.
	if s.current() != a || s.Pending != nil {
		t.Fatalf("turn %s pending %+v", s.current(), s.Pending)
	}
	mustFail(t, s, a, Action{Type: "nope"}, t0.Add(10*time.Second)) // nothing to cancel
}

func TestCanceledCardDoesNothing(t *testing.T) {
	s := setup(t, 2, 3)
	cur := s.current()
	other := s.Players[s.nextAlive(s.Turn)]
	c := give(s, cur, Current)
	must(t, s, cur, Action{Type: "play", Cards: []int{c.ID}}, t0)
	give(s, other, Nope)
	must(t, s, other, Action{Type: "nope"}, t0.Add(time.Second))
	must(t, s, other, Action{Type: "resolve"}, t0.Add(time.Second+ReactWindow))
	if s.current() != cur || s.TurnsLeft != 1 {
		t.Fatalf("a canceled Correnteza moved the turn: %s %d", s.current(), s.TurnsLeft)
	}
}

func TestCorrenteza(t *testing.T) {
	s := setup(t, 3, 4)
	cur := s.current()
	next := s.Players[s.nextAlive(s.Turn)]
	c := give(s, cur, Current)
	must(t, s, cur, Action{Type: "play", Cards: []int{c.ID}}, t0)
	must(t, s, cur, Action{Type: "resolve"}, t0.Add(ReactWindow))
	if s.current() != next || s.TurnsLeft != 2 {
		t.Fatalf("turn %s with %d", s.current(), s.TurnsLeft)
	}
	safeTop(s)
	must(t, s, next, Action{Type: "draw"}, t0.Add(5*time.Second))
	if s.current() != next || s.TurnsLeft != 1 {
		t.Fatal("the second turn is the same player's")
	}
	must(t, s, next, Action{Type: "draw"}, t0.Add(6*time.Second))
	if s.current() == next {
		t.Fatal("two turns, then the next player")
	}
}

// safeTop moves the jaguars away from the top two cards.
func safeTop(s *State) {
	slices.SortStableFunc(s.Deck, func(a, b Card) int {
		if (a.Kind == Jaguar) == (b.Kind == Jaguar) {
			return 0
		}
		if a.Kind == Jaguar {
			return 1
		}
		return -1
	})
}

func jaguarOnTop(s *State) {
	i := slices.IndexFunc(s.Deck, func(c Card) bool { return c.Kind == Jaguar })
	c := s.Deck[i]
	s.Deck = slices.Delete(s.Deck, i, i+1)
	s.Deck = slices.Insert(s.Deck, 0, c)
}

func TestDiveAndHide(t *testing.T) {
	s := setup(t, 3, 5)
	cur := s.current()
	jaguarOnTop(s)
	dives := count(s.Hands[cur], Dive)
	size := len(s.Deck)
	must(t, s, cur, Action{Type: "draw"}, t0)
	if s.Phase != PhaseDefuse || count(s.Hands[cur], Dive) != dives-1 || len(s.Deck) != size-1 {
		t.Fatalf("phase %s", s.Phase)
	}
	other := s.Players[s.nextAlive(s.Turn)]
	mustFail(t, s, other, Action{Type: "hide", Index: 0}, t0)
	mustFail(t, s, cur, Action{Type: "hide", Index: len(s.Deck) + 1}, t0)
	must(t, s, cur, Action{Type: "hide", Index: 2}, t0)
	if s.Deck[2].Kind != Jaguar || s.Phase != PhasePlay || s.current() != other {
		t.Fatalf("jaguar at 2? %v, turn %s", s.Deck[2], s.current())
	}
	// Nobody else learns where it went.
	raw, _ := json.Marshal(ViewFor(s, other, t0))
	if strings.Contains(string(raw), `"index"`) || strings.Contains(string(raw), `"deck":[`) {
		t.Fatal("the deck order leaked")
	}
}

func TestEatenAndLastCapybaraWins(t *testing.T) {
	s := setup(t, 2, 6)
	cur := s.current()
	other := s.Players[s.nextAlive(s.Turn)]
	s.Hands[cur] = slices.DeleteFunc(s.Hands[cur], func(c Card) bool { return c.Kind == Dive })
	jaguarOnTop(s)
	must(t, s, cur, Action{Type: "draw"}, t0)
	if s.Phase != PhaseOver || s.Winner != other || !slices.Equal(s.Out, []string{cur}) || len(s.Hands[cur]) != 0 {
		t.Fatalf("phase %s winner %s out %v", s.Phase, s.Winner, s.Out)
	}
	mustFail(t, s, other, Action{Type: "draw"}, t0)
}

func TestEatenPlayerIsSkipped(t *testing.T) {
	s := setup(t, 3, 7)
	cur := s.current()
	s.Hands[cur] = slices.DeleteFunc(s.Hands[cur], func(c Card) bool { return c.Kind == Dive })
	jaguarOnTop(s)
	must(t, s, cur, Action{Type: "draw"}, t0)
	if s.Phase != PhasePlay || s.current() == cur || s.alive(cur) {
		t.Fatal("eaten player still playing")
	}
	for range 4 {
		next := s.nextAlive(s.Turn)
		if s.Players[next] == cur {
			t.Fatal("turn order includes the eaten player")
		}
		s.Turn = next
	}
}

func TestPairTrioAndCuddle(t *testing.T) {
	s := setup(t, 3, 8)
	cur := s.current()
	victim := s.Players[s.nextAlive(s.Turn)]
	x, y := give(s, cur, "spa"), give(s, cur, "spa")
	before := len(s.Hands[victim])
	must(t, s, cur, Action{Type: "play", Cards: []int{x.ID, y.ID}, Target: victim}, t0)
	must(t, s, cur, Action{Type: "resolve"}, t0.Add(ReactWindow))
	if len(s.Hands[victim]) != before-1 {
		t.Fatal("the pair stole nothing")
	}
	last := s.Events[len(s.Events)-1]
	if last.Type != "steal" || last.Card == nil {
		t.Fatalf("steal event %+v", last)
	}
	// Others don't see which card it was; the two players do.
	third := s.Players[s.nextAlive(s.nextAlive(s.Turn))]
	for _, e := range ViewFor(s, third, t0).Events {
		if e.Type == "steal" && e.Card != nil {
			t.Fatal("a bystander saw the stolen card")
		}
	}
	seen := false
	for _, e := range ViewFor(s, victim, t0).Events {
		seen = seen || (e.Type == "steal" && e.Card != nil)
	}
	if !seen {
		t.Fatal("the victim should see what was stolen")
	}

	// Trio: ask for a Dive the victim holds.
	a, b, c := give(s, cur, "surf"), give(s, cur, "surf"), give(s, cur, "surf")
	give(s, victim, Dive)
	dives := count(s.Hands[cur], Dive)
	mustFail(t, s, cur, Action{Type: "play", Cards: []int{a.ID, b.ID, c.ID}, Target: victim, Ask: "bacon"}, t0)
	must(t, s, cur, Action{Type: "play", Cards: []int{a.ID, b.ID, c.ID}, Target: victim, Ask: Dive}, t0.Add(10*time.Second))
	must(t, s, cur, Action{Type: "resolve"}, t0.Add(10*time.Second+ReactWindow))
	if count(s.Hands[cur], Dive) != dives+1 {
		t.Fatal("the trio didn't get the Dive")
	}

	// Cafuné: the target picks what to give.
	cd := give(s, cur, Cuddle)
	must(t, s, cur, Action{Type: "play", Cards: []int{cd.ID}, Target: victim}, t0.Add(20*time.Second))
	must(t, s, cur, Action{Type: "resolve"}, t0.Add(20*time.Second+ReactWindow))
	if s.Phase != PhaseFavor {
		t.Fatalf("phase %s", s.Phase)
	}
	mustFail(t, s, cur, Action{Type: "draw"}, t0.Add(30*time.Second))
	mustFail(t, s, third, Action{Type: "give", Cards: []int{s.Hands[third][0].ID}}, t0.Add(30*time.Second))
	gift := s.Hands[victim][0]
	must(t, s, victim, Action{Type: "give", Cards: []int{gift.ID}}, t0.Add(30*time.Second))
	if !slices.ContainsFunc(s.Hands[cur], func(c Card) bool { return c.ID == gift.ID }) || s.Phase != PhasePlay {
		t.Fatal("the gift didn't arrive")
	}
}

func TestPeekAndSwirl(t *testing.T) {
	s := setup(t, 3, 9)
	cur := s.current()
	other := s.Players[s.nextAlive(s.Turn)]
	pk := give(s, cur, Peek)
	must(t, s, cur, Action{Type: "play", Cards: []int{pk.ID}}, t0)
	must(t, s, cur, Action{Type: "resolve"}, t0.Add(ReactWindow))
	if v := ViewFor(s, cur, t0); !slices.Equal(v.Me.Peek, s.Deck[:3]) {
		t.Fatal("Binóculo should show the top 3")
	}
	if v := ViewFor(s, other, t0); len(v.Me.Peek) != 0 {
		t.Fatal("someone else saw the Binóculo")
	}
	sw := give(s, cur, Swirl)
	must(t, s, cur, Action{Type: "play", Cards: []int{sw.ID}}, t0.Add(10*time.Second))
	must(t, s, cur, Action{Type: "resolve"}, t0.Add(10*time.Second+ReactWindow))
	if len(ViewFor(s, cur, t0).Me.Peek) != 0 {
		t.Fatal("a shuffle makes the old peek useless")
	}
}

func TestViewKeepsHandsSecret(t *testing.T) {
	s := setup(t, 4, 10)
	for _, p := range s.Players {
		v := ViewFor(s, p, t0)
		raw, _ := json.Marshal(v)
		for _, other := range s.Players {
			if other == p {
				continue
			}
			for _, c := range s.Hands[other] {
				if slices.ContainsFunc(v.Me.Hand, func(x Card) bool { return x.ID == c.ID }) {
					t.Fatalf("%s sees %s's card", p, other)
				}
			}
		}
		if strings.Contains(string(raw), `"hands"`) {
			t.Fatal("hands in the view")
		}
		for _, pv := range v.Players {
			if pv.Cards != len(s.Hands[pv.ID]) {
				t.Fatal("card counts are public")
			}
		}
	}
}
