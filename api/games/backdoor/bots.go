package backdoor

import (
	"math/rand/v2"
	"slices"
)

// Bot strategy. A bot only uses what its own badge and the public table show (the
// same knowledge a person in its seat would have), plus a pinch of randomness so it
// isn't predictable. It is meant to be a fair, slightly naive teammate, not a solver.

// BotAction is the next thing bot wants to do, or nil when it's waiting on others.
func BotAction(s *State, bot string, rng *rand.Rand) *Action {
	role, ok := s.Roles[bot]
	if !ok {
		return nil
	}
	switch s.Phase {
	case PhaseBadges:
		if !s.Ready[bot] {
			return &Action{Type: "ready"}
		}
	case PhaseProposal:
		if s.leaderID() != bot {
			return nil
		}
		if len(s.Selection) != s.SquadSize() {
			return &Action{Type: "select", Squad: botSquad(s, bot, rng)}
		}
		return &Action{Type: "propose"}
	case PhaseReview:
		if _, done := s.Votes[bot]; !done {
			v := botVote(s, bot, rng)
			return &Action{Type: "vote", Approve: &v}
		}
	case PhasePush:
		if _, done := s.Pushes[bot]; !done && slices.Contains(s.Selection, bot) {
			b := role.IsHacker() && botBackdoor(s, bot, rng)
			return &Action{Type: "push", Backdoor: &b}
		}
	case PhaseGuess:
		if role == SocialEngineer {
			return &Action{Type: "guess", Target: botGuess(s, bot, rng)}
		}
	}
	return nil
}

// knownHackers is who the bot's badge says plays for the hackers.
func knownHackers(s *State, bot string) map[string]bool {
	out := map[string]bool{}
	for _, k := range knowledge(s, bot) {
		if k.As != "analyst_candidate" {
			out[k.ID] = true
		}
	}
	return out
}

// suspicion grows for people in squads that had backdoors (more for more backdoors
// in a smaller squad) and shrinks for people in squads that deployed clean.
// Votes count too: hackers tend to approve the squads they sabotage.
func suspicion(s *State) map[string]float64 {
	out := map[string]float64{}
	approved := 0
	for _, pr := range s.Proposals {
		if !pr.Approved {
			continue
		}
		if approved >= len(s.Results) {
			break
		}
		r := s.Results[approved]
		approved++
		for _, p := range r.Squad {
			if r.Backdoors > 0 {
				out[p] += 2 * float64(r.Backdoors) / float64(len(r.Squad))
			} else {
				out[p] -= 0.3
			}
		}
		for p, yes := range pr.Votes {
			switch {
			case r.Backdoors > 0 && yes:
				out[p] += 0.25
			case r.Backdoors > 0 && !yes:
				out[p] -= 0.2
			}
		}
	}
	return out
}

// botBackdoor: hackers sometimes play clean in the first sprint to earn trust, and
// when teammates share a squad only as many backdoor as the sprint needs.
func botBackdoor(s *State, bot string, rng *rand.Rand) bool {
	need := BackdoorsNeeded(len(s.Players), s.Sprint)
	known := knownHackers(s, bot)
	mates := []string{bot}
	for _, p := range s.Selection {
		if known[p] && s.Roles[bot] != Freelancer {
			mates = append(mates, p)
		}
	}
	slices.Sort(mates)
	if idx := slices.Index(mates, bot); idx >= need && len(mates) > need {
		return rng.Float64() < 0.15 // someone else on the team has it covered
	}
	if s.Sprint == 0 {
		return rng.Float64() < 0.5
	}
	return rng.Float64() < 0.92
}

// botSquad: itself, then the least suspicious others. The Analyst leaves out the
// hackers it sees; a hacker leaves out its teammates to look clean.
func botSquad(s *State, bot string, rng *rand.Rand) []string {
	avoid := knownHackers(s, bot)
	sus := suspicion(s)
	others := []string{}
	for _, p := range s.Players {
		if p != bot && !avoid[p] {
			others = append(others, p)
		}
	}
	rng.Shuffle(len(others), func(i, j int) { others[i], others[j] = others[j], others[i] })
	slices.SortStableFunc(others, func(a, b string) int {
		switch {
		case sus[a] < sus[b]:
			return -1
		case sus[a] > sus[b]:
			return 1
		}
		return 0
	})
	squad := []string{bot}
	for _, p := range others {
		if len(squad) == s.SquadSize() {
			break
		}
		squad = append(squad, p)
	}
	// Not enough "clean" people (a big squad): fill with whoever is left.
	for _, p := range s.Players {
		if len(squad) == s.SquadSize() {
			break
		}
		if !slices.Contains(squad, p) {
			squad = append(squad, p)
		}
	}
	return squad
}

func botVote(s *State, bot string, rng *rand.Rand) bool {
	role := s.Roles[bot]
	known := knownHackers(s, bot)
	squad := s.Selection
	hasKnownHacker := slices.ContainsFunc(squad, func(p string) bool { return known[p] })
	inSquad := slices.Contains(squad, bot)

	if role.IsHacker() {
		// Approve squads with a hacker in them; otherwise mostly stall, which also
		// pushes the devs toward the fifth failed meeting.
		if inSquad || hasKnownHacker {
			return true
		}
		return rng.Float64() < 0.25
	}
	// Devs: the fifth failed meeting loses the game, so take the risk.
	if s.FailedMeetings == MaxFailedMeetings-1 {
		return true
	}
	// The Analyst blocks squads with hackers it sees, but not always (more firmly when
	// one more rollback loses), and otherwise votes like any dev, so the Social
	// Engineer can't spot it from the votes alone.
	if role == Analyst && hasKnownHacker {
		if s.rollbacks() == ToWin-1 {
			return rng.Float64() < 0.05
		}
		return rng.Float64() < 0.3
	}
	sus := suspicion(s)
	worst := 0.0
	for _, p := range squad {
		if p != bot {
			worst = max(worst, sus[p])
		}
	}
	switch {
	case worst >= 0.6:
		return rng.Float64() < 0.08
	case worst > 0.2:
		return rng.Float64() < 0.3
	case inSquad:
		return rng.Float64() < 0.9
	case s.Sprint == 0:
		return rng.Float64() < 0.6
	case s.FailedMeetings >= 2:
		return rng.Float64() < 0.65 // don't drift toward the fifth failed meeting
	default:
		return rng.Float64() < 0.4
	}
}

// botGuess: the Analyst tends to vote against squads with hackers and for clean ones,
// so pick the non-hacker whose votes matched that best.
func botGuess(s *State, bot string, rng *rand.Rand) string {
	known := knownHackers(s, bot)
	score := map[string]float64{}
	candidates := []string{}
	for _, p := range s.Players {
		if p != bot && !known[p] {
			candidates = append(candidates, p)
			score[p] = rng.Float64() * 0.5 // breaks ties
		}
	}
	for _, pr := range s.Proposals {
		dirty := slices.ContainsFunc(pr.Squad, func(p string) bool { return known[p] || p == bot })
		for p, approve := range pr.Votes {
			if _, ok := score[p]; !ok {
				continue
			}
			if dirty && !approve {
				score[p] += 1
			}
			if !dirty && approve {
				score[p] += 0.3
			}
		}
	}
	best := candidates[0]
	for _, p := range candidates {
		if score[p] > score[best] {
			best = p
		}
	}
	return best
}
