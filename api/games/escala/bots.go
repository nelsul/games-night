package escala

import "math/rand/v2"

var botClues = []string{"minúsculo", "pouco", "meio-termo", "bastante", "enorme"}

// BotAction is what a bot does now, or nil if it has nothing to do. Bots write a clue
// matching their number and slot their card in proportionally; they never reveal, so
// a person always decides when the table is done.
func BotAction(s *State, bot string, rng *rand.Rand) *Action {
	n := s.Numbers[bot]
	if s.Swap != nil {
		if _, voted := s.Swap.Votes[bot]; !voted {
			return &Action{Type: "swap_vote", Yes: rng.Float64() < 0.6}
		}
		return nil // wait for the table to settle the vote
	}
	switch s.Phase {
	case PhaseClues:
		if _, done := s.Clues[bot]; !done {
			return &Action{Type: "clue", Text: botClues[(n-1)*len(botClues)/100]}
		}
	case PhaseOrder:
		for _, p := range s.Line {
			if p == bot {
				return nil
			}
		}
		return &Action{Type: "place", Index: len(s.Line) * n / 100}
	}
	return nil
}
