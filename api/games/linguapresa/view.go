package linguapresa

import "time"

// TeamView is a team as everyone sees it.
type TeamView struct {
	Players []string `json:"players"`
	Score   int      `json:"score"`
}

// CardView is the card on the table. Only the narrator and the other teams get it.
type CardView struct {
	Word  string   `json:"word"`
	Taboo []string `json:"taboo"`
}

// PlayedView is a card from this (or the last) turn, once it is done. Everyone sees it.
type PlayedView struct {
	Word    string `json:"word"`
	Outcome string `json:"outcome"`
}

// MatchView is what one player may see of a match.
type MatchView struct {
	Phase      string       `json:"phase"`
	Turn       int          `json:"turn"`
	Turns      int          `json:"turns"`
	Teams      []TeamView   `json:"teams"`
	MyTeam     int          `json:"my_team"`
	ActiveTeam int          `json:"active_team"`
	Narrator   string       `json:"narrator"`
	Role       string       `json:"role"` // narrator | mate | watcher
	Seconds    int          `json:"seconds"`
	RemainingM int64        `json:"remaining_ms"` // while the clock runs, reading time included
	ReadingM   int64        `json:"reading_ms"`   // reading time left at the start of a turn
	BuzzWaitM  int64        `json:"buzz_wait_ms"` // until "Proibida!" works again
	Card       *CardView    `json:"card,omitempty"`
	CardSeq    int          `json:"card_seq"`
	SkipsLeft  int          `json:"skips_left"`
	Played     []PlayedView `json:"played"`
	Winners    []int        `json:"winners,omitempty"`
}

// ViewFor builds what user may see. The card on the table goes to the narrator and to
// the teams that are not playing; the narrator's own team never gets it.
func ViewFor(s *State, user string, now time.Time) *MatchView {
	v := &MatchView{
		Phase:      s.Phase,
		Turn:       s.Turn,
		Turns:      s.Turns,
		MyTeam:     s.TeamOf(user),
		ActiveTeam: s.ActiveTeam(),
		Seconds:    s.Seconds,
		CardSeq:    s.CardSeq,
		SkipsLeft:  max(MaxSkips-s.Skips, 0),
		Played:     []PlayedView{},
	}
	if s.Phase == PhaseOver {
		v.ActiveTeam = -1
		v.Winners = s.Winners()
	} else {
		v.Narrator = s.Narrator()
	}
	for _, t := range s.Teams {
		v.Teams = append(v.Teams, TeamView{Players: t.Players, Score: t.Score})
	}
	switch {
	case user == v.Narrator:
		v.Role = "narrator"
	case v.MyTeam == v.ActiveTeam:
		v.Role = "mate"
	default:
		v.Role = "watcher"
	}
	for _, p := range s.Played {
		v.Played = append(v.Played, PlayedView{Word: cards[p.Card].Word, Outcome: p.Outcome})
	}
	if s.Phase == PhaseTurn {
		v.RemainingM = max(s.EndsAt.Sub(now).Milliseconds(), 0)
		v.ReadingM = max(s.ReadUntil.Sub(now).Milliseconds(), 0)
		v.BuzzWaitM = max(s.BuzzOpenAt.Sub(now).Milliseconds(), 0)
		if s.Card >= 0 && v.Role != "mate" {
			c := cards[s.Card]
			v.Card = &CardView{Word: c.Word, Taboo: c.Taboo}
		}
	}
	return v
}
