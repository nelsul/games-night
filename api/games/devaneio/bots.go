package devaneio

import "math/rand/v2"

var botClues = []string{
	"saudade", "uma viagem sem volta", "domingo de manhã", "segredo guardado", "o começo de tudo",
	"silêncio", "cheiro de chuva", "um desejo", "caminho de casa", "sonho esquisito",
}

// BotAction is what a bot does now, or nil. Bots can't see the pictures, so they play
// at random: a clue from a fixed list, any card, any vote. They never move the table on
// to the next round, so a person always decides when everyone has seen the reveal.
func BotAction(s *State, bot string, rng *rand.Rand) *Action {
	hand := s.Hands[bot]
	switch s.Phase {
	case PhaseStory:
		if bot == s.Teller() && len(hand) > 0 {
			return &Action{Type: "story", Card: hand[rng.IntN(len(hand))], Text: botClues[rng.IntN(len(botClues))]}
		}
	case PhasePlay:
		if _, done := s.Played[bot]; !done && bot != s.Teller() && len(hand) > 0 {
			return &Action{Type: "play", Card: hand[rng.IntN(len(hand))]}
		}
	case PhaseVote:
		if _, done := s.Votes[bot]; done || bot == s.Teller() {
			return nil
		}
		var options []int
		for i, owner := range s.Table {
			if owner != bot {
				options = append(options, i)
			}
		}
		return &Action{Type: "vote", Index: options[rng.IntN(len(options))]}
	}
	return nil
}
