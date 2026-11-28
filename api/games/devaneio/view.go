package devaneio

import "slices"

// PlayerView is a player as everyone sees them.
type PlayerView struct {
	ID    string `json:"id"`
	Score int    `json:"score"`
	// Done is whether they have done their part of the phase (played a card, voted).
	Done bool `json:"done"`
}

// TableCard is a card on the table. Who played it and who voted for it only show once
// the round is scored.
type TableCard struct {
	Card   int      `json:"card"`
	Mine   bool     `json:"mine"`
	Owner  string   `json:"owner,omitempty"`
	Teller bool     `json:"teller,omitempty"`
	Voters []string `json:"voters,omitempty"`
}

// MatchView is what one player may see of a match.
type MatchView struct {
	Phase   string       `json:"phase"`
	Round   int          `json:"round"`
	Rounds  int          `json:"rounds"`
	Teller  string       `json:"teller"`
	Clue    string       `json:"clue,omitempty"`
	Players []PlayerView `json:"players"`
	// Hand is only the viewer's own cards.
	Hand  []int       `json:"hand"`
	Table []TableCard `json:"table"`
	// MyVote is the position the viewer voted for, or -1.
	MyVote  int            `json:"my_vote"`
	Gains   map[string]int `json:"gains"`
	Winners []string       `json:"winners,omitempty"`
}

// ViewFor builds what user may see. Other people's hands never leave the server, and
// during the vote nobody can tell which card is whose (the teller's included).
func ViewFor(s *State, user string) *MatchView {
	v := &MatchView{
		Phase:   s.Phase,
		Round:   s.Round,
		Rounds:  s.Rounds,
		Teller:  s.Teller(),
		Hand:    slices.Clone(s.Hands[user]),
		Table:   []TableCard{},
		MyVote:  -1,
		Gains:   map[string]int{},
		Players: []PlayerView{},
	}
	if s.Phase == PhaseOver {
		v.Teller = ""
		v.Winners = s.Winners()
	}
	if v.Hand == nil {
		v.Hand = []int{}
	}
	if s.Phase != PhaseStory && s.Phase != PhaseOver {
		v.Clue = s.Clue
	}
	for _, p := range s.Players {
		pv := PlayerView{ID: p, Score: s.Scores[p]}
		switch s.Phase {
		case PhaseStory:
			pv.Done = false
		case PhasePlay:
			_, pv.Done = s.Played[p]
		case PhaseVote:
			if p == s.Teller() {
				pv.Done = true
			} else {
				_, pv.Done = s.Votes[p]
			}
		default:
			pv.Done = true
		}
		v.Players = append(v.Players, pv)
	}
	if pos, ok := s.Votes[user]; ok {
		v.MyVote = pos
	}
	if s.Phase == PhaseVote || s.Phase == PhaseReveal {
		for i, owner := range s.Table {
			tc := TableCard{Card: s.Played[owner], Mine: owner == user}
			if s.Phase == PhaseReveal {
				tc.Owner = owner
				tc.Teller = owner == s.Teller()
				for _, p := range s.Players {
					if pos, ok := s.Votes[p]; ok && pos == i {
						tc.Voters = append(tc.Voters, p)
					}
				}
			}
			v.Table = append(v.Table, tc)
		}
	}
	if s.Phase == PhaseReveal {
		for p, g := range s.Gains {
			v.Gains[p] = g
		}
	}
	return v
}
