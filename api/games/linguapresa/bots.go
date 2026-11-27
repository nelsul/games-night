package linguapresa

import (
	"math/rand/v2"
	"time"
)

// BotAction is what a bot does now, or nil. A bot narrator starts its turn, plays a few
// cards (mostly hits) and ends it. Bots never press "Próximo", so a person always
// decides when the table moves on, and they never buzz.
func BotAction(s *State, bot string, now time.Time, rng *rand.Rand) *Action {
	if s.Phase == PhaseOver || s.Narrator() != bot {
		return nil
	}
	switch s.Phase {
	case PhaseReady:
		return &Action{Type: "start"}
	case PhaseTurn:
		if !now.Before(s.EndsAt) {
			return &Action{Type: "time_up"}
		}
		if len(s.Played) >= 2 && rng.Float64() < 0.35 {
			return &Action{Type: "finish"}
		}
		if s.Skips < MaxSkips && rng.Float64() < 0.12 {
			return &Action{Type: "skip", Seq: s.CardSeq}
		}
		return &Action{Type: "hit", Seq: s.CardSeq}
	}
	return nil
}
