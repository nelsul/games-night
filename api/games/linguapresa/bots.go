package linguapresa

import (
	"math/rand/v2"
	"time"
)

// readyPause is how long a bot narrator waits before starting a turn, so the table can
// read the "Vez do time" announcement before the clock runs.
const readyPause = 5 * time.Second

// BotAction is what a bot does now, or nil. A bot narrator starts its turn, skips a few
// cards and ends it: it never scores, because nobody is guessing. A bot's turn is
// really a pause for its team, and people can still tap "Proibida!" on its cards. Bots
// never press "Próximo", so a person always decides when the table moves on.
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
		if s.Skips < MaxSkips && rng.Float64() < 0.8 {
			return &Action{Type: "skip", Seq: s.CardSeq}
		}
		return &Action{Type: "finish"}
	}
	return nil
}

// botPause is how much longer a bot waits before an action, so people can follow it:
// the announcement of the turn is read before the clock starts, and each card stays a
// moment before the bot moves on.
func botPause(s *State, a Action, now time.Time) time.Duration {
	switch a.Type {
	case "start":
		return max(0, readyPause-now.Sub(s.ReadyAt))
	case "skip", "finish":
		return 1500 * time.Millisecond
	}
	return 0
}
