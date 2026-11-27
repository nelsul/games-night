package escala

// PlayerView is one person at the table as everyone may see them.
type PlayerView struct {
	ID      string `json:"id"`
	Clue    string `json:"clue,omitempty"`
	HasClue bool   `json:"has_clue"`
	Placed  bool   `json:"placed"`
}

// Card is a placed card. Number and Wrong only appear once the round is revealed.
type Card struct {
	Player string `json:"player"`
	Number *int   `json:"number,omitempty"`
	Wrong  bool   `json:"wrong,omitempty"`
}

// SwapView is an open vote to change the theme; votes are public.
type SwapView struct {
	Proposer string          `json:"proposer"`
	Votes    map[string]bool `json:"votes"`
}

// MyView is what only this player knows.
type MyView struct {
	Number int `json:"number"` // their own secret number
}

// MatchView is the match as one player may see it. Other people's numbers are sent
// only after the reveal.
type MatchView struct {
	Phase   string       `json:"phase"`
	Round   int          `json:"round"`
	Rounds  int          `json:"rounds"`
	Score   int          `json:"score"` // rounds sorted right so far
	Theme   Theme        `json:"theme"`
	Players []PlayerView `json:"players"`
	Line    []Card       `json:"line"`
	Results []Result     `json:"results"`
	Swap    *SwapView    `json:"swap,omitempty"`
	// SwapsLeft is how many more theme changes the table may ask for this round.
	SwapsLeft int    `json:"swaps_left"`
	SwapSeq   int    `json:"swap_seq"`
	SwapOK    bool   `json:"swap_ok"`
	Me        MyView `json:"me"`
}

// ViewFor builds what user may see.
func ViewFor(s *State, user string) *MatchView {
	// After the last reveal the phase is "over", but the numbers are still on the table.
	revealed := s.Phase == PhaseReveal || (s.Phase == PhaseOver && len(s.Line) == len(s.Players))
	v := &MatchView{
		Phase: s.Phase, Round: s.Round, Rounds: s.Rounds, Score: s.Score(), Theme: s.Theme,
		Results: s.Results, Line: []Card{},
	}
	placed := map[string]bool{}
	for _, p := range s.Line {
		placed[p] = true
	}
	v.SwapsLeft, v.SwapSeq, v.SwapOK = max(MaxSwapProposals-s.Swaps, 0), s.SwapSeq, s.SwapOK
	if s.Swap != nil {
		v.Swap = &SwapView{Proposer: s.Swap.Proposer, Votes: s.Swap.Votes}
	}
	if v.Results == nil {
		v.Results = []Result{}
	}
	for _, p := range s.Players {
		pv := PlayerView{ID: p, Placed: placed[p]}
		if c, ok := s.Clues[p]; ok {
			pv.HasClue = true
			// Clues are public as soon as they are written, so the table can avoid repeating each other.
			pv.Clue = c
		}
		v.Players = append(v.Players, pv)
	}
	wrong := s.Misplaced()
	for _, p := range s.Line {
		c := Card{Player: p}
		if revealed {
			n := s.Numbers[p]
			c.Number, c.Wrong = &n, wrong[p]
		}
		v.Line = append(v.Line, c)
	}
	if s.has(user) {
		v.Me.Number = s.Numbers[user]
	}
	return v
}
