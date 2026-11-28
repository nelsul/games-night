package devaneio

import (
	"math/rand/v2"
	"slices"
	"strings"
	"testing"
	"time"
)

func rng() *rand.Rand { return rand.New(rand.NewPCG(7, 9)) }

var t0 = time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

func ids(n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = string(rune('a' + i))
	}
	return out
}

func newMatch(t *testing.T, n int, o Options) *State {
	t.Helper()
	s, err := Setup(ids(n), o, rng(), t0)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func do(t *testing.T, s *State, user string, a Action) {
	t.Helper()
	if err := Apply(s, user, a, rng()); err != nil {
		t.Fatalf("%s %s: %v", user, a.Type, err)
	}
}

// tell has the teller start the round, and returns who they are.
func tell(t *testing.T, s *State) string {
	t.Helper()
	teller := s.Teller()
	do(t, s, teller, Action{Type: "story", Card: s.Hands[teller][0], Text: "uma dica"})
	return teller
}

// playAll makes everyone else play their first card.
func playAll(t *testing.T, s *State) {
	t.Helper()
	for _, p := range s.Players {
		if _, done := s.Played[p]; !done {
			do(t, s, p, Action{Type: "play", Card: s.Hands[p][0]})
		}
	}
}

// votes makes each voter pick the table position pos(voter).
func votes(t *testing.T, s *State, pos func(voter string) int) {
	t.Helper()
	for _, p := range s.Players {
		if p != s.Teller() {
			do(t, s, p, Action{Type: "vote", Index: pos(p)})
		}
	}
}

// anyOther is a vote for the first card that is not the voter's own.
func anyOther(s *State) func(string) int {
	return func(v string) int {
		for i, o := range s.Table {
			if o != v {
				return i
			}
		}
		return 0
	}
}

func TestSetup(t *testing.T) {
	for _, n := range []int{4, 9, 14} {
		s := newMatch(t, n, Options{})
		seen := map[int]bool{}
		for _, p := range s.Players {
			if len(s.Hands[p]) != HandSize {
				t.Fatalf("%d players: %s has %d cards", n, p, len(s.Hands[p]))
			}
			for _, c := range s.Hands[p] {
				if seen[c] {
					t.Fatalf("card %d dealt twice", c)
				}
				seen[c] = true
			}
		}
		if len(seen)+len(s.Deck) != CardCount {
			t.Fatalf("%d cards in hands and deck, want %d", len(seen)+len(s.Deck), CardCount)
		}
	}
	for _, n := range []int{3, 15} {
		if _, err := Setup(ids(n), Options{}, rng(), t0); err == nil {
			t.Errorf("%d players should be refused", n)
		}
	}
	if (Options{Rounds: 4}).CheckLobby() == nil {
		t.Error("4 rounds should be refused")
	}
}

func TestRoundCount(t *testing.T) {
	if got := newMatch(t, 6, Options{}).Rounds; got != 6 {
		t.Fatalf("one turn each: %d rounds, want 6", got)
	}
	if got := newMatch(t, 6, Options{Rounds: 8}).Rounds; got != 8 {
		t.Fatalf("rounds = %d, want 8", got)
	}
}

func TestStoryRules(t *testing.T) {
	s := newMatch(t, 4, Options{})
	teller := s.Teller()
	other := s.Players[1]
	if err := Apply(s, other, Action{Type: "story", Card: s.Hands[other][0], Text: "x"}, rng()); err == nil {
		t.Fatal("only the teller tells")
	}
	if err := Apply(s, teller, Action{Type: "story", Card: s.Hands[teller][0]}, rng()); err == nil {
		t.Fatal("a clue is required")
	}
	if err := Apply(s, teller, Action{Type: "story", Card: s.Hands[other][0], Text: "x"}, rng()); err == nil {
		t.Fatal("the card must be in the hand of the teller")
	}
	long := strings.Repeat("a", MaxClueRune+1)
	if err := Apply(s, teller, Action{Type: "story", Card: s.Hands[teller][0], Text: long}, rng()); err == nil {
		t.Fatal("a clue that is too long should be refused")
	}
	card := s.Hands[teller][0]
	do(t, s, teller, Action{Type: "story", Card: card, Text: "  uma   dica  "})
	if s.Clue != "uma dica" || s.Phase != PhasePlay || s.TellerCard != card {
		t.Fatalf("clue %q phase %s card %d", s.Clue, s.Phase, s.TellerCard)
	}
	if slices.Contains(s.Hands[teller], card) {
		t.Fatal("the played card must leave the hand")
	}
}

func TestPlayAndVoteRules(t *testing.T) {
	s := newMatch(t, 4, Options{})
	teller := tell(t, s)
	if err := Apply(s, teller, Action{Type: "play", Card: s.Hands[teller][0]}, rng()); err == nil {
		t.Fatal("the teller already played")
	}
	p := s.Players[1]
	if p == teller {
		p = s.Players[2]
	}
	if err := Apply(s, p, Action{Type: "vote"}, rng()); err == nil {
		t.Fatal("no voting before the table opens")
	}
	do(t, s, p, Action{Type: "play", Card: s.Hands[p][0]})
	if err := Apply(s, p, Action{Type: "play", Card: s.Hands[p][0]}, rng()); err == nil {
		t.Fatal("one card each")
	}
	playAll(t, s)
	if s.Phase != PhaseVote || len(s.Table) != 4 {
		t.Fatalf("phase %s, %d cards on the table", s.Phase, len(s.Table))
	}
	if err := Apply(s, teller, Action{Type: "vote", Index: 0}, rng()); err == nil {
		t.Fatal("the teller doesn't vote")
	}
	own := slices.Index(s.Table, p)
	if err := Apply(s, p, Action{Type: "vote", Index: own}, rng()); err == nil {
		t.Fatal("no voting for your own card")
	}
	if err := Apply(s, p, Action{Type: "vote", Index: 9}, rng()); err == nil {
		t.Fatal("no voting for a card that is not there")
	}
}

func TestScoringMixedVotes(t *testing.T) {
	s := newMatch(t, 5, Options{})
	teller := tell(t, s)
	playAll(t, s)
	tp := slices.Index(s.Table, teller)
	var voters []string
	for _, p := range s.Players {
		if p != teller {
			voters = append(voters, p)
		}
	}
	// voters[1] and voters[2] find the teller's card; voters[3] votes for voters[0]'s card,
	// and voters[0] votes for voters[2]'s.
	pos := map[string]int{
		voters[0]: slices.Index(s.Table, voters[2]),
		voters[1]: tp,
		voters[2]: tp,
		voters[3]: slices.Index(s.Table, voters[0]),
	}
	votes(t, s, func(v string) int { return pos[v] })
	if s.Phase != PhaseReveal {
		t.Fatalf("phase %s, want reveal", s.Phase)
	}
	want := map[string]int{
		teller:    3,
		voters[0]: 1,     // fooled voters[3]
		voters[1]: 3,     // found it
		voters[2]: 3 + 1, // found it, and fooled voters[0]
		voters[3]: 0,
	}
	for p, w := range want {
		if s.Gains[p] != w {
			t.Errorf("%s got %d, want %d", p, s.Gains[p], w)
		}
		if s.Scores[p] != w {
			t.Errorf("%s has %d in total, want %d", p, s.Scores[p], w)
		}
	}
}

func TestScoringEverybodyOrNobodyFindsIt(t *testing.T) {
	for _, everybody := range []bool{true, false} {
		s := newMatch(t, 4, Options{})
		teller := tell(t, s)
		playAll(t, s)
		tp := slices.Index(s.Table, teller)
		votes(t, s, func(v string) int {
			if everybody {
				return tp
			}
			for i, o := range s.Table {
				if o != teller && o != v {
					return i
				}
			}
			return -1
		})
		for _, p := range s.Players {
			want := 2
			if p == teller {
				want = 0
			}
			if s.Gains[p] != want {
				t.Errorf("everybody=%v: %s got %d, want %d", everybody, p, s.Gains[p], want)
			}
		}
	}
}

func TestBluffPointsAreCapped(t *testing.T) {
	s := newMatch(t, 8, Options{})
	teller := tell(t, s)
	playAll(t, s)
	tp := slices.Index(s.Table, teller)
	var voters []string
	for _, p := range s.Players {
		if p != teller {
			voters = append(voters, p)
		}
	}
	// Seven voters: voters[0] finds the teller's card, the other six all vote for the
	// card of voters[1]. Only voters[1] can't vote for their own, so they pick another card.
	target := voters[1]
	votes(t, s, func(v string) int {
		switch v {
		case voters[0], voters[1]:
			return tp
		default:
			return slices.Index(s.Table, target)
		}
	})
	// voters[1] found the teller's card (3) and fooled five people, but the bonus caps at 3.
	if got := s.Gains[target]; got != 3+3 {
		t.Fatalf("target got %d, want 6 (3 for finding the card, 3 for the capped bonus)", got)
	}
}

func TestNextRoundRefillsAndEnds(t *testing.T) {
	s := newMatch(t, 4, Options{})
	first := s.Teller()
	for round := 0; round < s.Rounds; round++ {
		tell(t, s)
		playAll(t, s)
		votes(t, s, anyOther(s))
		if err := Apply(s, "a", Action{Type: "story"}, rng()); err == nil {
			t.Fatal("nothing to tell during the reveal")
		}
		do(t, s, "a", Action{Type: "next"})
		if s.Phase == PhaseOver {
			break
		}
		for _, p := range s.Players {
			if len(s.Hands[p]) != HandSize {
				t.Fatalf("round %d: %s has %d cards", round, p, len(s.Hands[p]))
			}
		}
		if s.Teller() == first {
			t.Fatal("the teller must change every round")
		}
	}
	if s.Phase != PhaseOver || s.Round != s.Rounds {
		t.Fatalf("phase %s round %d of %d", s.Phase, s.Round, s.Rounds)
	}
}

func TestDeckNeverLosesACard(t *testing.T) {
	s := newMatch(t, 14, Options{Rounds: 8})
	for s.Phase != PhaseOver {
		tell(t, s)
		playAll(t, s)
		votes(t, s, anyOther(s))
		do(t, s, "a", Action{Type: "next"})
	}
	total := len(s.Deck) + len(s.Discard)
	for _, p := range s.Players {
		total += len(s.Hands[p])
	}
	if total != CardCount {
		t.Fatalf("%d cards in the game, want %d: some got lost or duplicated", total, CardCount)
	}
}

func TestHiddenInformation(t *testing.T) {
	s := newMatch(t, 5, Options{})
	teller := tell(t, s)
	playAll(t, s)
	for _, p := range s.Players {
		v := ViewFor(s, p)
		if len(v.Hand) == 0 {
			t.Fatalf("%s sees no hand", p)
		}
		for _, h := range v.Hand {
			if !slices.Contains(s.Hands[p], h) {
				t.Fatalf("%s sees a card that is not theirs", p)
			}
		}
		mine := 0
		for _, c := range v.Table {
			if c.Owner != "" || c.Teller || len(c.Voters) != 0 {
				t.Fatalf("%s learns who owns a card before the reveal", p)
			}
			if c.Mine {
				mine++
			}
		}
		if mine != 1 {
			t.Fatalf("%s sees %d of their own cards, want 1", p, mine)
		}
		if len(v.Gains) != 0 {
			t.Fatal("no points before the vote ends")
		}
	}
	votes(t, s, anyOther(s))
	v := ViewFor(s, "b")
	if len(v.Table) != 5 {
		t.Fatalf("%d cards shown, want 5", len(v.Table))
	}
	for _, c := range v.Table {
		if c.Owner == "" {
			t.Fatal("the reveal shows every owner")
		}
		if c.Teller != (c.Owner == teller) {
			t.Fatal("the card of the teller is marked wrong")
		}
	}
}

func TestBotsPlayARound(t *testing.T) {
	s := newMatch(t, 4, Options{})
	r := rng()
	for step := 0; step < 100 && s.Phase != PhaseReveal; step++ {
		acted := false
		for _, p := range s.Players {
			if a := BotAction(s, p, r); a != nil {
				if err := Apply(s, p, *a, r); err != nil {
					t.Fatal(err)
				}
				acted = true
				break
			}
		}
		if !acted {
			t.Fatalf("bots stuck in %s", s.Phase)
		}
	}
	if s.Phase != PhaseReveal {
		t.Fatal("the bots never finished the round")
	}
	for _, p := range s.Players {
		if BotAction(s, p, r) != nil {
			t.Fatal("bots wait for a person to move on")
		}
	}
}
