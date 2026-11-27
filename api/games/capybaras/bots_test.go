package capybaras

import (
	"math/rand/v2"
	"testing"
	"time"
)

// Tables of bots always finish a match without breaking a rule.
func TestBotsPlayWholeMatches(t *testing.T) {
	for n := MinPlayers; n <= MaxPlayers; n++ {
		for seed := range uint64(15) {
			s := setup(t, n, seed*31+uint64(n))
			r := rand.New(rand.NewPCG(seed, 99))
			now := t0
			for step := 0; s.Phase != PhaseOver; step++ {
				if step > 5000 {
					t.Fatalf("n=%d seed=%d: no end after 5000 steps", n, seed)
				}
				acted := false
				for _, bot := range s.Players {
					if a := BotAction(s, bot, now, r); a != nil {
						if err := Apply(s, bot, *a, now, r); err != nil {
							t.Fatalf("n=%d seed=%d: %s %+v: %v", n, seed, bot, *a, err)
						}
						acted = true
						break
					}
				}
				now = now.Add(time.Second)
				if !acted && s.Pending == nil {
					t.Fatalf("n=%d seed=%d: stuck in %s", n, seed, s.Phase)
				}
			}
			if s.Winner == "" || len(s.Out) != n-1 {
				t.Fatalf("n=%d: winner %q out %v", n, s.Winner, s.Out)
			}
		}
	}
}
