package escala

import (
	"encoding/json"
	"math/rand/v2"
	"strconv"
	"strings"
	"testing"
	"time"
)

func rng() *rand.Rand { return rand.New(rand.NewPCG(1, 2)) }

func newMatch(t *testing.T, n, rounds int) *State {
	t.Helper()
	ps := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}[:n]
	s, err := Setup(ps, Options{Rounds: rounds}, rng(), time.Now())
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func clueAll(t *testing.T, s *State) {
	t.Helper()
	for _, p := range s.Players {
		if err := Apply(s, p, Action{Type: "clue", Text: "x"}, rng()); err != nil {
			t.Fatal(err)
		}
	}
}

// placeSorted puts every card in the right order (or in the reverse order).
func placeSorted(t *testing.T, s *State, reverse bool) {
	t.Helper()
	for _, p := range s.Players {
		idx := 0
		for _, q := range s.Line {
			if (s.Numbers[q] < s.Numbers[p]) != reverse {
				idx++
			}
		}
		if err := Apply(s, p, Action{Type: "place", Index: idx}, rng()); err != nil {
			t.Fatal(err)
		}
	}
}

func TestSetupDealsUniqueNumbers(t *testing.T) {
	s := newMatch(t, 10, 0)
	seen := map[int]bool{}
	for _, n := range s.Numbers {
		if n < 1 || n > 100 || seen[n] {
			t.Fatalf("bad numbers: %v", s.Numbers)
		}
		seen[n] = true
	}
	if s.Rounds != DefaultRounds {
		t.Fatalf("defaults: %+v", s)
	}
}

func TestLimits(t *testing.T) {
	if _, err := Setup([]string{"a", "b"}, Options{}, rng(), time.Now()); err == nil {
		t.Fatal("2 players should be refused")
	}
	if err := (Options{Rounds: 4}).CheckLobby(); err == nil {
		t.Fatal("4 rounds should be refused")
	}
}

func TestCluesThenOrder(t *testing.T) {
	s := newMatch(t, 3, 3)
	if err := Apply(s, "a", Action{Type: "place"}, rng()); err == nil {
		t.Fatal("placing during clues should fail")
	}
	if err := Apply(s, "a", Action{Type: "clue", Text: "  "}, rng()); err == nil {
		t.Fatal("blank clue should fail")
	}
	if err := Apply(s, "zed", Action{Type: "clue", Text: "x"}, rng()); err == nil {
		t.Fatal("outsiders can't act")
	}
	clueAll(t, s)
	if s.Phase != PhaseOrder {
		t.Fatalf("phase %s", s.Phase)
	}
}

func TestRevealNeedsEveryCard(t *testing.T) {
	s := newMatch(t, 3, 3)
	clueAll(t, s)
	_ = Apply(s, "a", Action{Type: "place", Index: 0}, rng())
	if err := Apply(s, "a", Action{Type: "reveal"}, rng()); err == nil {
		t.Fatal("reveal with missing cards should fail")
	}
}

func TestMovingACard(t *testing.T) {
	s := newMatch(t, 3, 3)
	clueAll(t, s)
	_ = Apply(s, "a", Action{Type: "place", Index: 0}, rng())
	_ = Apply(s, "b", Action{Type: "place", Index: 1}, rng())
	_ = Apply(s, "a", Action{Type: "place", Index: 1}, rng()) // a moves after b
	if len(s.Line) != 2 || s.Line[0] != "b" || s.Line[1] != "a" {
		t.Fatalf("line %v", s.Line)
	}
	if err := Apply(s, "a", Action{Type: "place", Index: 9}, rng()); err == nil {
		t.Fatal("out of range should fail")
	}
	if len(s.Line) != 2 {
		t.Fatalf("a failed move must not lose the card: %v", s.Line)
	}
}

func TestPerfectMatchScoresEveryRound(t *testing.T) {
	s := newMatch(t, 4, 3)
	for r := 0; r < 3; r++ {
		clueAll(t, s)
		placeSorted(t, s, false)
		if err := Apply(s, "a", Action{Type: "reveal"}, rng()); err != nil {
			t.Fatal(err)
		}
		if err := Apply(s, "a", Action{Type: "next"}, rng()); err != nil {
			t.Fatal(err)
		}
	}
	if s.Phase != PhaseOver || s.Score() != 3 {
		t.Fatalf("expected 3 of 3: %+v", s)
	}
}

// Wrong orders never end the match early: the table always plays every round.
func TestMistakesDoNotEndTheMatch(t *testing.T) {
	s := newMatch(t, 4, 5)
	for r := 0; r < 5; r++ {
		if s.Phase == PhaseOver {
			t.Fatalf("match ended early at round %d", r)
		}
		clueAll(t, s)
		placeSorted(t, s, true)
		_ = Apply(s, "a", Action{Type: "reveal"}, rng())
		_ = Apply(s, "a", Action{Type: "next"}, rng())
	}
	if s.Phase != PhaseOver || s.Score() != 0 {
		t.Fatalf("expected the end with 0 of 5: %+v", s)
	}
	if v := ViewFor(s, "a"); v.Score != 0 || len(v.Results) != 5 {
		t.Fatalf("view: %+v", v)
	}
}

func TestChangingAClue(t *testing.T) {
	s := newMatch(t, 3, 3)
	_ = Apply(s, "a", Action{Type: "clue", Text: "primeira"}, rng())
	_ = Apply(s, "a", Action{Type: "clue", Text: "segunda"}, rng())
	if s.Clues["a"] != "segunda" || s.Phase != PhaseClues {
		t.Fatalf("clue should be replaced while clues are open: %+v", s.Clues)
	}
	_ = Apply(s, "b", Action{Type: "clue", Text: "x"}, rng())
	_ = Apply(s, "c", Action{Type: "clue", Text: "x"}, rng())
	// Still changeable while the table is sorting, and that doesn't restart anything.
	if err := Apply(s, "b", Action{Type: "clue", Text: "melhor"}, rng()); err != nil {
		t.Fatal(err)
	}
	if s.Clues["b"] != "melhor" || s.Phase != PhaseOrder {
		t.Fatalf("clue change during order: %+v", s)
	}
	clue := s.Clues["b"]
	placeSorted(t, s, false)
	_ = Apply(s, "a", Action{Type: "reveal"}, rng())
	if err := Apply(s, "b", Action{Type: "clue", Text: "tarde"}, rng()); err == nil || s.Clues["b"] != clue {
		t.Fatal("clues are locked after the reveal")
	}
}

func TestViewHidesOthersNumbers(t *testing.T) {
	s := newMatch(t, 3, 3)
	clueAll(t, s)
	_ = Apply(s, "a", Action{Type: "place", Index: 0}, rng())
	v := ViewFor(s, "b")
	if v.Me.Number != s.Numbers["b"] {
		t.Fatal("player should see their own number")
	}
	if v.Line[0].Number != nil {
		t.Fatal("others' numbers must stay hidden before the reveal")
	}
	raw, _ := json.Marshal(v)
	if strings.Contains(string(raw), `"number":`+strconv.Itoa(s.Numbers["a"])+`}`) {
		t.Fatal("a's number leaked into b's view")
	}
}

func TestNumbersShownAfterReveal(t *testing.T) {
	s := newMatch(t, 3, 3)
	clueAll(t, s)
	placeSorted(t, s, true)
	_ = Apply(s, "a", Action{Type: "reveal"}, rng())
	for _, c := range ViewFor(s, "b").Line {
		if c.Number == nil {
			t.Fatal("numbers should be shown after the reveal")
		}
	}
}

func TestCluesVisibleAsWritten(t *testing.T) {
	s := newMatch(t, 3, 3)
	_ = Apply(s, "a", Action{Type: "clue", Text: "segredo"}, rng())
	for _, p := range ViewFor(s, "b").Players {
		if p.ID == "a" && p.Clue != "segredo" {
			t.Fatal("others should see a clue as soon as it is written")
		}
	}
}

func TestBotsFinishTheirPart(t *testing.T) {
	s := newMatch(t, 3, 3)
	for i := 0; i < 5; i++ {
		for _, p := range s.Players {
			if a := BotAction(s, p, rng()); a != nil {
				if err := Apply(s, p, *a, rng()); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
	if s.Phase != PhaseOrder || len(s.Line) != 3 {
		t.Fatalf("bots should clue and place: %+v", s)
	}
}

func TestThemes(t *testing.T) {
	if len(themes) < 100 {
		t.Fatalf("want at least 100 themes, have %d", len(themes))
	}
	seen := map[string]bool{}
	for _, th := range themes {
		if th.Title == "" || th.Low == "" || th.High == "" || th.Low == th.High {
			t.Fatalf("incomplete theme: %+v", th)
		}
		if seen[th.Title] {
			t.Fatalf("duplicate theme %q", th.Title)
		}
		seen[th.Title] = true
	}
}

// A match of 7 rounds never repeats a theme.
func TestNoRepeatedThemeInAMatch(t *testing.T) {
	s := newMatch(t, 3, 7)
	seen := map[string]bool{}
	for r := 0; r < 7; r++ {
		if seen[s.Theme.Title] {
			t.Fatalf("repeated theme %q", s.Theme.Title)
		}
		seen[s.Theme.Title] = true
		clueAll(t, s)
		placeSorted(t, s, false)
		_ = Apply(s, "a", Action{Type: "reveal"}, rng())
		_ = Apply(s, "a", Action{Type: "next"}, rng())
	}
}
