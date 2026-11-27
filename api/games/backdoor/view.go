package backdoor

import (
	"slices"
	"time"
)

// Known is someone a player knows something about from their badge.
type Known struct {
	ID string `json:"id"`
	// As is what this player sees them as: hacker, social_engineer or zero_day (for
	// hackers and the Analyst), or analyst_candidate (for the Product Owner).
	As string `json:"as"`
}

type Me struct {
	Role        Role    `json:"role"`
	Team        string  `json:"team"` // devs or hackers
	Knows       []Known `json:"knows"`
	Ready       bool    `json:"ready"`
	Vote        *bool   `json:"vote"`   // my vote in the current review
	Pushed      bool    `json:"pushed"` // I already pushed this sprint (never what I pushed)
	InSquad     bool    `json:"in_squad"`
	CanBackdoor bool    `json:"can_backdoor"`
	IsLeader    bool    `json:"is_leader"`
}

type PlayerView struct {
	ID       string `json:"id"`
	IsLeader bool   `json:"is_leader"`
	Ready    bool   `json:"ready"`
	Voted    bool   `json:"voted"`
	Pushed   bool   `json:"pushed"`
}

// View is everything one player may see. Roles of others and individual pushes stay
// on the server until the match is over.
type View struct {
	Phase          Phase           `json:"phase"`
	Sprint         int             `json:"sprint"`
	SquadSizes     [Sprints]int    `json:"squad_sizes"`
	DoubleSprint   int             `json:"double_sprint"` // 0-based sprint needing 2 backdoors, or -1
	SquadSize      int             `json:"squad_size"`
	Leader         string          `json:"leader"`
	FailedMeetings int             `json:"failed_meetings"`
	MaxFailed      int             `json:"max_failed"`
	Players        []PlayerView    `json:"players"`
	Selection      []string        `json:"selection"`
	Results        []SprintResult  `json:"results"`
	Proposals      []Proposal      `json:"proposals"`
	Options        Options         `json:"options"`
	PhaseStartedAt time.Time       `json:"phase_started_at"`
	Commits        []bool          `json:"commits"`                   // the commits opened so far (reveal phase)
	CommitsTotal   int             `json:"commits_total"`             // how many there are to open
	BackdoorsNeed  int             `json:"backdoors_needed"`          // backdoors that roll this sprint back
	SocialEngineer string          `json:"social_engineer,omitempty"` // public once the guess starts
	Me             Me              `json:"me"`
	Winner         string          `json:"winner,omitempty"`
	Reason         string          `json:"reason,omitempty"`
	Guess          string          `json:"guess,omitempty"`
	Roles          map[string]Role `json:"roles,omitempty"` // only when the match is over
}

// ViewFor builds the view of one player.
func ViewFor(s *State, player string) View {
	n := len(s.Players)
	v := View{
		Phase:          s.Phase,
		Sprint:         s.Sprint,
		SquadSizes:     squadSizes[n],
		DoubleSprint:   -1,
		SquadSize:      s.SquadSize(),
		Leader:         s.leaderID(),
		FailedMeetings: s.FailedMeetings,
		MaxFailed:      MaxFailedMeetings,
		Selection:      nonNil(s.Selection),
		Results:        nonNilResults(s.Results),
		Proposals:      nonNilProposals(s.Proposals),
		Options:        s.Options,
		PhaseStartedAt: s.PhaseStartedAt,
		Commits:        []bool{},
		BackdoorsNeed:  BackdoorsNeeded(n, s.Sprint),
	}
	if s.Phase == PhaseReveal {
		v.Commits = slices.Clone(s.Commits[:s.Opened])
		v.CommitsTotal = len(s.Commits)
	}
	if n >= 7 {
		v.DoubleSprint = 3
	}
	for _, p := range s.Players {
		_, voted := s.Votes[p]
		_, pushed := s.Pushes[p]
		v.Players = append(v.Players, PlayerView{
			ID: p, IsLeader: p == s.leaderID(), Ready: s.Ready[p], Voted: voted, Pushed: pushed,
		})
	}
	if s.Phase == PhaseGuess || s.Phase == PhaseOver {
		v.SocialEngineer = s.socialEngineer()
	}
	if s.Phase == PhaseOver {
		v.Winner, v.Reason, v.Guess = s.Winner, s.Reason, s.Guess
		v.Roles = s.Roles
	}

	role, ok := s.Roles[player]
	if !ok {
		return v // not in the match: public information only
	}
	me := Me{Role: role, Team: "devs", Ready: s.Ready[player], IsLeader: player == s.leaderID()}
	if role.IsHacker() {
		me.Team = "hackers"
	}
	if vote, ok := s.Votes[player]; ok {
		me.Vote = &vote
	}
	_, me.Pushed = s.Pushes[player]
	me.InSquad = s.Phase == PhasePush && slices.Contains(s.Selection, player)
	me.CanBackdoor = role.IsHacker()
	me.Knows = knowledge(s, player)
	v.Me = me
	return v
}

// knowledge is what a badge reveals about the others.
func knowledge(s *State, player string) []Known {
	role := s.Roles[player]
	out := []Known{}
	for _, p := range s.Players {
		if p == player {
			continue
		}
		other := s.Roles[p]
		switch {
		case role == Analyst && other.IsHacker() && other != ZeroDay:
			out = append(out, Known{ID: p, As: "hacker"})
		case role == ProductOwner && (other == Analyst || other == ZeroDay):
			out = append(out, Known{ID: p, As: "analyst_candidate"})
		case role.IsHacker() && role != Freelancer && other.IsHacker() && other != Freelancer:
			as := "hacker"
			if other == SocialEngineer || other == ZeroDay {
				as = string(other)
			}
			out = append(out, Known{ID: p, As: as})
		}
	}
	return out
}

func nonNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

func nonNilResults(s []SprintResult) []SprintResult {
	if s == nil {
		return []SprintResult{}
	}
	return s
}

func nonNilProposals(s []Proposal) []Proposal {
	if s == nil {
		return []Proposal{}
	}
	return s
}
