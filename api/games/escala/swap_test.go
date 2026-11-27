package escala

import "testing"

func TestSwapThemeByMajority(t *testing.T) {
	s := newMatch(t, 4, 3)
	clueAll(t, s)
	old := s.Theme
	if err := Apply(s, "a", Action{Type: "swap_propose"}, rng()); err != nil {
		t.Fatal(err)
	}
	if err := Apply(s, "b", Action{Type: "swap_propose"}, rng()); err == nil {
		t.Fatal("only one vote at a time")
	}
	if err := Apply(s, "a", Action{Type: "reveal"}, rng()); err == nil {
		t.Fatal("can't reveal during a vote")
	}
	_ = Apply(s, "b", Action{Type: "swap_vote", Yes: true}, rng())
	if s.Swap == nil {
		t.Fatal("2 of 4 is not a majority yet")
	}
	numbers := map[string]int{}
	for k, v := range s.Numbers {
		numbers[k] = v
	}
	_ = Apply(s, "c", Action{Type: "swap_vote", Yes: true}, rng())
	if s.Swap != nil || s.Theme == old || !s.SwapOK || s.SwapSeq != 1 {
		t.Fatalf("theme should have changed: %+v", s)
	}
	if s.Phase != PhaseClues || len(s.Clues) != 0 || len(s.Line) != 0 {
		t.Fatalf("a new theme restarts clues: %+v", s)
	}
	for k, v := range numbers {
		if s.Numbers[k] != v {
			t.Fatal("numbers must stay when only the theme changes")
		}
	}
}

func TestSwapRejectedKeepsTheme(t *testing.T) {
	s := newMatch(t, 4, 3)
	old := s.Theme
	_ = Apply(s, "a", Action{Type: "swap_propose"}, rng())
	_ = Apply(s, "b", Action{Type: "swap_vote", Yes: false}, rng())
	if s.Swap == nil {
		t.Fatal("1 no of 4 doesn't close it")
	}
	_ = Apply(s, "c", Action{Type: "swap_vote", Yes: false}, rng()) // 1 yes, 2 no: it can't win
	if s.Swap != nil || s.Theme != old || s.SwapOK || s.SwapSeq != 1 {
		t.Fatalf("expected a rejection: %+v", s)
	}
}

func TestSwapLimitAndPhases(t *testing.T) {
	s := newMatch(t, 4, 3)
	for i := 0; i < MaxSwapProposals; i++ {
		if err := Apply(s, "a", Action{Type: "swap_propose"}, rng()); err != nil {
			t.Fatal(err)
		}
		_ = Apply(s, "b", Action{Type: "swap_vote"}, rng())
		_ = Apply(s, "c", Action{Type: "swap_vote"}, rng())
	}
	if err := Apply(s, "a", Action{Type: "swap_propose"}, rng()); err == nil {
		t.Fatal("limit per round")
	}
	if err := Apply(s, "a", Action{Type: "swap_vote", Yes: true}, rng()); err == nil {
		t.Fatal("no vote open")
	}
	clueAll(t, s)
	placeSorted(t, s, false)
	_ = Apply(s, "a", Action{Type: "reveal"}, rng())
	if err := Apply(s, "a", Action{Type: "swap_propose"}, rng()); err == nil {
		t.Fatal("no theme change after the reveal")
	}
	_ = Apply(s, "a", Action{Type: "next"}, rng())
	if s.Swaps != 0 {
		t.Fatal("the limit resets every round")
	}
}

func TestBotsVoteOnSwaps(t *testing.T) {
	s := newMatch(t, 3, 3)
	_ = Apply(s, "a", Action{Type: "swap_propose"}, rng())
	for _, p := range []string{"b", "c"} {
		a := BotAction(s, p, rng())
		if a == nil || a.Type != "swap_vote" {
			t.Fatalf("bots should vote: %+v", a)
		}
	}
}
