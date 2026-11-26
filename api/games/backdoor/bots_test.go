package backdoor

import (
	"math/rand/v2"
	"testing"
)

// Whole tables of bots must always finish a match without breaking a rule.
func TestBotsPlayWholeMatches(t *testing.T) {
	wins := map[string]int{}
	reasons := map[string]int{}
	for n := MinPlayers; n <= MaxPlayers; n++ {
		for seed := uint64(0); seed < 60; seed++ {
			opts := Options{ProductOwner: seed%2 == 0}
			if n >= SpecialHackersMin {
				opts.ZeroDay, opts.Freelancer = seed%3 == 0, seed%5 == 0
			}
			s := setup(t, n, opts, seed+uint64(n)*1000)
			rng := rand.New(rand.NewPCG(seed, uint64(n)))
			for steps := 0; s.Phase != PhaseOver; steps++ {
				if steps > 2000 {
					t.Fatalf("%d players, seed %d: stuck in %s", n, seed, s.Phase)
				}
				acted := false
				for _, bot := range s.Players {
					if a := BotAction(s, bot, rng); a != nil {
						if err := Apply(s, bot, *a, t0); err != nil {
							t.Fatalf("%d players, seed %d: bot %s (%s) %s: %v", n, seed, bot, s.Roles[bot], a.Type, err)
						}
						acted = true
						break
					}
				}
				if !acted {
					t.Fatalf("%d players, seed %d: nobody can act in %s", n, seed, s.Phase)
				}
			}
			wins[s.Winner]++
			reasons[s.Reason]++
		}
	}
	t.Logf("wins %v, reasons %v", wins, reasons)
	if wins["devs"] == 0 || wins["hackers"] == 0 {
		t.Fatalf("one side never wins: %v", wins)
	}
}

func TestBotAnalystNeverPicksAKnownHacker(t *testing.T) {
	for seed := uint64(0); seed < 200; seed++ {
		s := setup(t, 8, Options{}, seed)
		analyst := byRole(s, Analyst)[0]
		allReady(t, s)
		s.Leader = slices_index(s.Players, analyst)
		squad := botSquad(s, analyst, rand.New(rand.NewPCG(seed, 2)))
		for _, p := range squad {
			if s.Roles[p].IsHacker() && s.Roles[p] != ZeroDay {
				t.Fatalf("seed %d: the Analyst picked hacker %s", seed, p)
			}
		}
	}
}

func TestBotDevsNeverBackdoor(t *testing.T) {
	s := setup(t, 6, Options{}, 3)
	allReady(t, s)
	d := devsOnly(s)
	must(t, s, s.leaderID(), Action{Type: "select", Squad: d[:2]})
	must(t, s, s.leaderID(), Action{Type: "propose"})
	for _, p := range s.Players {
		must(t, s, p, Action{Type: "vote", Approve: yes()})
	}
	rng := rand.New(rand.NewPCG(1, 1))
	for i := 0; i < 50; i++ {
		a := BotAction(s, d[0], rng)
		if a == nil || a.Type != "push" || *a.Backdoor {
			t.Fatalf("dev bot push: %+v", a)
		}
	}
}

func slices_index(list []string, v string) int {
	for i, x := range list {
		if x == v {
			return i
		}
	}
	return -1
}
