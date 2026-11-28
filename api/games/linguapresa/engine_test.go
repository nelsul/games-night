package linguapresa

import (
	"math/rand/v2"
	"slices"
	"strings"
	"testing"
	"time"
)

func rng() *rand.Rand { return rand.New(rand.NewPCG(1, 2)) }

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

func do(t *testing.T, s *State, user string, a Action, at time.Time) {
	t.Helper()
	if err := Apply(s, user, a, at, rng()); err != nil {
		t.Fatalf("%s %s: %v", user, a.Type, err)
	}
}

func TestCards(t *testing.T) {
	if len(cards) < 100 {
		t.Fatalf("want at least 100 cards, have %d", len(cards))
	}
	seen := map[string]bool{}
	for _, c := range cards {
		if seen[c.Word] {
			t.Errorf("duplicate card %q", c.Word)
		}
		seen[c.Word] = true
		if len(c.Taboo) != 5 {
			t.Errorf("%q needs exactly 5 forbidden words, has %d", c.Word, len(c.Taboo))
		}
		for _, w := range c.Taboo {
			if strings.EqualFold(w, c.Word) {
				t.Errorf("%q forbids itself", c.Word)
			}
		}
	}
}

func TestTeams(t *testing.T) {
	cases := []struct {
		n     int
		o     Options
		teams int
		err   bool
	}{
		{4, Options{}, 2, false},
		{5, Options{}, 2, false},
		{9, Options{}, 2, false},
		{6, Options{Mode: "pairs"}, 3, false},
		{10, Options{Mode: "pairs"}, 5, false},
		{5, Options{Mode: "pairs"}, 0, true},
		{6, Options{Mode: "trios"}, 2, false},
		{9, Options{Mode: "trios"}, 3, false},
		{8, Options{Mode: "trios"}, 0, true},
		{3, Options{}, 0, true},
		{16, Options{}, 2, false},
		{16, Options{Mode: "pairs"}, 8, false},
		{15, Options{Mode: "trios"}, 5, false},
		{17, Options{}, 0, true},
	}
	for _, c := range cases {
		s, err := Setup(ids(c.n), c.o, rng(), t0)
		if c.err {
			if err == nil {
				t.Errorf("%d players %+v: want an error", c.n, c.o)
			}
			continue
		}
		if err != nil {
			t.Errorf("%d players %+v: %v", c.n, c.o, err)
			continue
		}
		if len(s.Teams) != c.teams {
			t.Errorf("%d players %+v: %d teams, want %d", c.n, c.o, len(s.Teams), c.teams)
		}
		total := 0
		for _, tm := range s.Teams {
			total += len(tm.Players)
		}
		if total != c.n {
			t.Errorf("everyone must be on a team: %d of %d", total, c.n)
		}
	}
}

func TestOptions(t *testing.T) {
	for _, o := range []Options{{Mode: "x"}, {Laps: 3}, {Seconds: 30}} {
		if o.CheckLobby() == nil {
			t.Errorf("%+v should be refused", o)
		}
	}
}

func TestNarratorsRotate(t *testing.T) {
	s := newMatch(t, 6, Options{})
	if s.Turns != 2*3 {
		t.Fatalf("turns = %d, want 6", s.Turns)
	}
	seen := map[string]bool{}
	for s.Turn = 0; s.Turn < s.Turns; s.Turn++ {
		n := s.Narrator()
		if seen[n] {
			t.Fatalf("%s narrates twice in one lap", n)
		}
		seen[n] = true
		if s.TeamOf(n) != s.ActiveTeam() {
			t.Fatalf("narrator %s is not on the active team", n)
		}
	}
	if len(seen) != 6 {
		t.Fatalf("%d narrators, want everyone", len(seen))
	}
}

func startTurn(t *testing.T, s *State) string {
	t.Helper()
	n := s.Narrator()
	do(t, s, n, Action{Type: "start"}, t0)
	return n
}

func TestTurnScoring(t *testing.T) {
	s := newMatch(t, 4, Options{})
	n := startTurn(t, s)
	team := s.ActiveTeam()
	do(t, s, n, Action{Type: "hit", Seq: s.CardSeq}, t0.Add(5*time.Second))
	do(t, s, n, Action{Type: "hit", Seq: s.CardSeq}, t0.Add(9*time.Second))
	do(t, s, n, Action{Type: "skip", Seq: s.CardSeq}, t0.Add(10*time.Second))
	other := s.Teams[1-team].Players[0]
	do(t, s, other, Action{Type: "buzz", Seq: s.CardSeq}, t0.Add(12*time.Second))
	if got := s.Teams[team].Score; got != 1 {
		t.Fatalf("score = %d, want 2 hits - 1 buzz = 1", got)
	}
	if len(s.Played) != 4 {
		t.Fatalf("played = %d, want 4", len(s.Played))
	}
}

func TestWhoMayDoWhat(t *testing.T) {
	s := newMatch(t, 4, Options{})
	n := s.Narrator()
	mate := ""
	for _, p := range s.Teams[s.ActiveTeam()].Players {
		if p != n {
			mate = p
		}
	}
	rival := s.Teams[1-s.ActiveTeam()].Players[0]
	if err := Apply(s, mate, Action{Type: "start"}, t0, rng()); err == nil {
		t.Fatal("only the narrator may start")
	}
	do(t, s, n, Action{Type: "start"}, t0)
	seq := s.CardSeq
	at := t0.Add(time.Second)
	for _, c := range []struct {
		user string
		a    Action
	}{
		{mate, Action{Type: "hit", Seq: seq}},
		{rival, Action{Type: "hit", Seq: seq}},
		{rival, Action{Type: "skip", Seq: seq}},
		{n, Action{Type: "buzz", Seq: seq}},
		{mate, Action{Type: "buzz", Seq: seq}},
		{n, Action{Type: "hit", Seq: seq + 5}},
	} {
		if err := Apply(s, c.user, c.a, at, rng()); err == nil {
			t.Errorf("%s %+v should be refused", c.user, c.a)
		}
	}
}

func TestSkipLimit(t *testing.T) {
	s := newMatch(t, 4, Options{})
	n := startTurn(t, s)
	for i := 0; i < MaxSkips; i++ {
		do(t, s, n, Action{Type: "skip", Seq: s.CardSeq}, t0.Add(time.Second))
	}
	if err := Apply(s, n, Action{Type: "skip", Seq: s.CardSeq}, t0.Add(time.Second), rng()); err == nil {
		t.Fatal("a fourth skip should be refused")
	}
}

func TestTimeUp(t *testing.T) {
	s := newMatch(t, 4, Options{Seconds: 45})
	n := startTurn(t, s)
	if err := Apply(s, "a", Action{Type: "time_up"}, t0.Add(10*time.Second), rng()); err == nil {
		t.Fatal("the clock has not run out")
	}
	late := t0.Add(46 * time.Second)
	if err := Apply(s, n, Action{Type: "hit", Seq: s.CardSeq}, late, rng()); err == nil {
		t.Fatal("no hits after the time is up")
	}
	do(t, s, "a", Action{Type: "time_up"}, late)
	if s.Phase != PhaseSummary || s.Card != -1 {
		t.Fatalf("phase %s card %d after time up", s.Phase, s.Card)
	}
}

func TestMatchEndsAndWinners(t *testing.T) {
	s := newMatch(t, 4, Options{})
	for s.Phase != PhaseOver {
		n := startTurn(t, s)
		if s.ActiveTeam() == 0 {
			do(t, s, n, Action{Type: "hit", Seq: s.CardSeq}, t0.Add(time.Second))
		}
		do(t, s, n, Action{Type: "finish"}, t0.Add(2*time.Second))
		do(t, s, "a", Action{Type: "next"}, t0.Add(3*time.Second))
	}
	if s.Turn != s.Turns {
		t.Fatalf("turn %d of %d", s.Turn, s.Turns)
	}
	if w := s.Winners(); !slices.Equal(w, []int{0}) {
		t.Fatalf("winners = %v, want team 0", w)
	}
}

func TestCardHiddenFromNarratorsTeam(t *testing.T) {
	s := newMatch(t, 6, Options{})
	n := startTurn(t, s)
	for _, p := range s.Players {
		v := ViewFor(s, p, t0.Add(time.Second))
		mate := s.TeamOf(p) == s.ActiveTeam() && p != n
		switch {
		case mate && v.Card != nil:
			t.Fatalf("%s is on the narrator's team and must not see the card", p)
		case !mate && v.Card == nil:
			t.Fatalf("%s should see the card", p)
		}
	}
	// Between turns nobody sees a card, so nothing can leak early.
	do(t, s, n, Action{Type: "finish"}, t0.Add(time.Second))
	for _, p := range s.Players {
		if ViewFor(s, p, t0.Add(2*time.Second)).Card != nil {
			t.Fatal("no card on the table between turns")
		}
	}
}

func TestBotsPlayATurnWithoutScoring(t *testing.T) {
	s := newMatch(t, 4, Options{})
	bot := s.Narrator()
	r := rng()
	for i := 0; i < 60 && s.Phase != PhaseSummary; i++ {
		a := BotAction(s, bot, t0.Add(time.Second), r)
		if a == nil {
			t.Fatalf("bot stuck in %s", s.Phase)
		}
		if a.Type == "hit" {
			t.Fatal("a bot narrator must never mark a hit: nobody is guessing")
		}
		if err := Apply(s, bot, *a, t0.Add(time.Second), r); err != nil {
			t.Fatal(err)
		}
	}
	if s.Phase != PhaseSummary {
		t.Fatal("the bot never ended its turn")
	}
	for i, tm := range s.Teams {
		if tm.Score != 0 {
			t.Fatalf("team %d scored %d from a bot's turn", i, tm.Score)
		}
	}
	if BotAction(s, bot, t0.Add(time.Second), rng()) != nil {
		t.Fatal("bots wait for a person to move on")
	}
}

func TestBotWaitsBeforeStarting(t *testing.T) {
	s := newMatch(t, 4, Options{})
	start := Action{Type: "start"}
	if got := botPause(s, start, t0.Add(time.Second)); got != readyPause-time.Second {
		t.Fatalf("pause after 1s = %v, want %v", got, readyPause-time.Second)
	}
	if got := botPause(s, start, t0.Add(readyPause+time.Second)); got != 0 {
		t.Fatalf("pause after the wait = %v, want 0", got)
	}
	// A new turn restarts the wait from the moment it became ready.
	s.Phase = PhaseSummary
	later := t0.Add(time.Minute)
	do(t, s, "a", Action{Type: "next"}, later)
	if got := botPause(s, start, later); got != readyPause {
		t.Fatalf("pause at the start of turn 2 = %v, want %v", got, readyPause)
	}
}
