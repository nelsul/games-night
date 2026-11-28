// Package backdoor is the rules engine of Backdoor, a social deduction game: the devs
// at Cudia ship sprints while hidden hackers try to roll them back.
//
// The engine is pure: Setup deals a match, Apply validates and applies one player's
// action, and View builds what one player may see. Nothing here touches the network
// or the database, so every rule is tested in engine_test.go.
package backdoor

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"time"
)

type Role string

const (
	Dev            Role = "dev"
	Analyst        Role = "analyst"
	ProductOwner   Role = "product_owner"
	Hacker         Role = "hacker"
	SocialEngineer Role = "social_engineer"
	ZeroDay        Role = "zero_day" // hidden from the Analyst
	Deepfake       Role = "deepfake" // looks like a second Analyst to the Product Owner
	Freelancer     Role = "freelancer"
)

// IsHacker reports whether the role plays for the hackers.
func (r Role) IsHacker() bool {
	switch r {
	case Hacker, SocialEngineer, ZeroDay, Deepfake, Freelancer:
		return true
	}
	return false
}

type Phase string

const (
	PhaseBadges   Phase = "badges"   // everyone looks at their badge
	PhaseProposal Phase = "proposal" // the Scrum Master builds the squad
	PhaseReview   Phase = "review"   // everyone votes on the squad
	PhasePush     Phase = "push"     // the squad pushes in secret
	PhaseReveal   Phase = "reveal"   // the Scrum Master opens the commits one by one
	PhaseGuess    Phase = "guess"    // the Social Engineer hunts the Analyst
	PhaseOver     Phase = "over"
)

const (
	MinPlayers        = 5
	MaxPlayers        = 16
	SpecialHackersMin = 7 // Zero-day and Freelancer need at least this many players
	Sprints           = 5
	ToWin             = 3 // deploys or rollbacks
	MaxFailedMeetings = 5
)

// About 40% of the table plays for the hackers.
var hackersFor = map[int]int{5: 2, 6: 2, 7: 3, 8: 3, 9: 3, 10: 4, 11: 4, 12: 5, 13: 5, 14: 6, 15: 6, 16: 7}

var squadSizes = map[int][Sprints]int{
	5:  {2, 3, 2, 3, 3},
	6:  {2, 3, 4, 3, 4},
	7:  {2, 3, 3, 4, 4},
	8:  {3, 4, 4, 5, 5},
	9:  {3, 4, 4, 5, 5},
	10: {3, 4, 4, 5, 5},
	11: {3, 4, 4, 5, 5},
	12: {3, 4, 5, 5, 6},
	13: {4, 4, 5, 6, 6},
	14: {4, 5, 5, 6, 6},
	15: {4, 5, 6, 6, 7},
	16: {4, 5, 6, 7, 7},
}

// BackdoorsNeeded is how many backdoors roll back a sprint (sprint is 0-based).
// From 7 players the 4th sprint needs 2; from 13, with squads of 6 or 7, the 5th too.
func BackdoorsNeeded(players, sprint int) int {
	if (sprint == 3 && players >= 7) || (sprint == 4 && players >= 13) {
		return 2
	}
	return 1
}

// Options are what the host sets in the lobby.
type Options struct {
	ProductOwner bool `json:"product_owner"`
	ZeroDay      bool `json:"zero_day"`
	Deepfake     bool `json:"deepfake"`
	Freelancer   bool `json:"freelancer"`
	TimerSeconds int  `json:"timer_seconds"` // 0 (off), 60, 120 or 180
	SquadHistory bool `json:"squad_history"` // show who was in each past sprint during the match
}

// RuleError is a broken rule; its message is shown to players (Portuguese).
type RuleError struct{ Msg string }

func (e *RuleError) Error() string { return e.Msg }

func ruleErr(msg string) error { return &RuleError{Msg: msg} }

// CheckLobby validates options while people are still joining.
func (o Options) CheckLobby() error {
	if o.Deepfake && !o.ProductOwner {
		return ruleErr("O Deepfake só entra junto com o Product Owner: é para ele que o Deepfake se passa pelo Analista.")
	}
	switch o.TimerSeconds {
	case 0, 60, 120, 180:
		return nil
	}
	return ruleErr("Escolha um timer de 1, 2 ou 3 minutos, ou desligue.")
}

// specialHackers is how many hacker seats the options take: the Social Engineer,
// always, plus every optional hacker role that is on.
func (o Options) specialHackers() int {
	n := 1
	for _, on := range []bool{o.ZeroDay, o.Deepfake, o.Freelancer} {
		if on {
			n++
		}
	}
	return n
}

// CheckStart validates options for the actual table size. Every role that is on must
// be dealt, so the table needs enough hacker seats for all of them.
func (o Options) CheckStart(players int) error {
	if err := o.CheckLobby(); err != nil {
		return err
	}
	if players < MinPlayers || players > MaxPlayers {
		return ruleErr("O Backdoor é para 5 a 16 pessoas.")
	}
	if (o.ZeroDay || o.Freelancer) && players < SpecialHackersMin {
		return ruleErr("Zero-day e Freelancer só entram com 7 pessoas ou mais. Desligue essas opções ou chame mais gente.")
	}
	if seats := hackersFor[players]; o.specialHackers() > seats {
		return ruleErr(fmt.Sprintf("Com %d pessoas são %d hackers, e a Engenharia Social já ocupa uma vaga: não cabem todos os papéis de hacker ligados. Desligue algum ou chame mais gente.", players, seats))
	}
	return nil
}

type Proposal struct {
	Sprint   int             `json:"sprint"`
	Leader   string          `json:"leader"`
	Squad    []string        `json:"squad"`
	Votes    map[string]bool `json:"votes"` // true = approve
	Approved bool            `json:"approved"`
}

type SprintResult struct {
	Squad     []string `json:"squad"`
	Backdoors int      `json:"backdoors"`
	Rollback  bool     `json:"rollback"`
}

// State is the whole match, stored as JSON. Pushes and roles are secret: only View
// decides who sees what.
type State struct {
	Players        []string        `json:"players"` // seat order
	Roles          map[string]Role `json:"roles"`
	Options        Options         `json:"options"`
	Phase          Phase           `json:"phase"`
	Sprint         int             `json:"sprint"` // 0-based
	Leader         int             `json:"leader"` // index into Players: the Scrum Master
	FailedMeetings int             `json:"failed_meetings"`
	Ready          map[string]bool `json:"ready"`
	Selection      []string        `json:"selection"`
	Votes          map[string]bool `json:"votes"`
	Pushes         map[string]bool `json:"pushes"`  // true = backdoor
	Commits        []bool          `json:"commits"` // the sprint's pushes, shuffled, while they are opened
	Opened         int             `json:"opened"`  // how many of Commits are open
	Proposals      []Proposal      `json:"proposals"`
	Results        []SprintResult  `json:"results"`
	Guess          string          `json:"guess,omitempty"`
	Winner         string          `json:"winner,omitempty"` // devs or hackers
	Reason         string          `json:"reason,omitempty"` // rollbacks, meetings, guess, guess_missed
	PhaseStartedAt time.Time       `json:"phase_started_at"`
}

// Setup deals roles at random and picks a random first Scrum Master.
func Setup(players []string, opts Options, rng *rand.Rand, now time.Time) (*State, error) {
	if err := opts.CheckStart(len(players)); err != nil {
		return nil, err
	}
	n := len(players)
	hackerRoles := []Role{SocialEngineer}
	if opts.ZeroDay {
		hackerRoles = append(hackerRoles, ZeroDay)
	}
	if opts.Deepfake {
		hackerRoles = append(hackerRoles, Deepfake)
	}
	if opts.Freelancer {
		hackerRoles = append(hackerRoles, Freelancer)
	}
	for len(hackerRoles) < hackersFor[n] {
		hackerRoles = append(hackerRoles, Hacker)
	}
	devRoles := []Role{Analyst}
	if opts.ProductOwner {
		devRoles = append(devRoles, ProductOwner)
	}
	for len(devRoles) < n-hackersFor[n] {
		devRoles = append(devRoles, Dev)
	}
	deck := append(hackerRoles, devRoles...)
	if len(deck) != n { // CheckStart makes room for every role that is on
		return nil, fmt.Errorf("backdoor: dealt %d roles for %d players", len(deck), n)
	}
	rng.Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })

	s := &State{
		Players: slices.Clone(players),
		Roles:   map[string]Role{},
		Options: opts,
		Phase:   PhaseBadges,
		Leader:  rng.IntN(n),
		Ready:   map[string]bool{},
		Votes:   map[string]bool{},
		Pushes:  map[string]bool{},
	}
	for i, p := range players {
		s.Roles[p] = deck[i]
	}
	s.enter(PhaseBadges, now)
	return s, nil
}

func (s *State) enter(p Phase, now time.Time) {
	s.Phase = p
	s.PhaseStartedAt = now
}

// SquadSize is how many people the current sprint needs.
func (s *State) SquadSize() int { return squadSizes[len(s.Players)][s.Sprint] }

func (s *State) leaderID() string { return s.Players[s.Leader] }

func (s *State) deploys() int   { return s.count(false) }
func (s *State) rollbacks() int { return s.count(true) }

func (s *State) count(rollback bool) int {
	n := 0
	for _, r := range s.Results {
		if r.Rollback == rollback {
			n++
		}
	}
	return n
}

func (s *State) socialEngineer() string {
	for _, p := range s.Players {
		if s.Roles[p] == SocialEngineer {
			return p
		}
	}
	return ""
}

func (s *State) nextLeader() { s.Leader = (s.Leader + 1) % len(s.Players) }

func (s *State) finish(winner, reason string, now time.Time) {
	s.Winner, s.Reason = winner, reason
	s.enter(PhaseOver, now)
}

// Action is one thing a player does. Only the fields its Type needs are set.
type Action struct {
	Type     string   `json:"type"` // ready, select, propose, vote, push, open, continue, guess
	Squad    []string `json:"squad,omitempty"`
	Approve  *bool    `json:"approve,omitempty"`
	Backdoor *bool    `json:"backdoor,omitempty"`
	Target   string   `json:"target,omitempty"`
}

// Apply checks an action against the rules and applies it. On error the state is unchanged.
func Apply(s *State, player string, a Action, now time.Time) error {
	if _, ok := s.Roles[player]; !ok {
		return ruleErr("Você não está nessa partida.")
	}
	switch a.Type {
	case "ready":
		return s.ready(player, now)
	case "select":
		return s.selectSquad(player, a.Squad)
	case "propose":
		return s.propose(player, now)
	case "vote":
		return s.vote(player, a.Approve, now)
	case "push":
		return s.push(player, a.Backdoor, now)
	case "open":
		return s.open(player, now)
	case "continue":
		return s.closeSprint(player, now)
	case "guess":
		return s.guess(player, a.Target, now)
	}
	return ruleErr("Ação desconhecida.")
}

func (s *State) ready(player string, now time.Time) error {
	if s.Phase != PhaseBadges {
		return ruleErr("A partida já começou.")
	}
	s.Ready[player] = true
	if len(s.Ready) == len(s.Players) {
		s.enter(PhaseProposal, now)
	}
	return nil
}

func (s *State) selectSquad(player string, squad []string) error {
	if s.Phase != PhaseProposal {
		return ruleErr("Agora não é hora de montar o squad.")
	}
	if player != s.leaderID() {
		return ruleErr("Só o Scrum Master monta o squad.")
	}
	if len(squad) > s.SquadSize() {
		return ruleErr("Esse squad tem gente demais.")
	}
	seen := map[string]bool{}
	for _, p := range squad {
		if _, ok := s.Roles[p]; !ok || seen[p] {
			return ruleErr("Escolha pessoas da partida, sem repetir.")
		}
		seen[p] = true
	}
	s.Selection = slices.Clone(squad)
	return nil
}

func (s *State) propose(player string, now time.Time) error {
	if s.Phase != PhaseProposal {
		return ruleErr("Agora não é hora de abrir o PR.")
	}
	if player != s.leaderID() {
		return ruleErr("Só o Scrum Master abre o PR.")
	}
	if len(s.Selection) != s.SquadSize() {
		return ruleErr("Escolha exatamente quem vai no squad.")
	}
	s.Votes = map[string]bool{}
	s.enter(PhaseReview, now)
	return nil
}

func (s *State) vote(player string, approve *bool, now time.Time) error {
	if s.Phase != PhaseReview {
		return ruleErr("Agora não é hora de votar.")
	}
	if approve == nil {
		return ruleErr("Escolha aprovar ou pedir mudanças.")
	}
	if _, done := s.Votes[player]; done {
		return ruleErr("Você já votou.")
	}
	s.Votes[player] = *approve
	if len(s.Votes) < len(s.Players) {
		return nil
	}
	yes := 0
	for _, v := range s.Votes {
		if v {
			yes++
		}
	}
	approved := yes*2 > len(s.Players)
	s.Proposals = append(s.Proposals, Proposal{
		Sprint: s.Sprint, Leader: s.leaderID(), Squad: slices.Clone(s.Selection),
		Votes: s.Votes, Approved: approved,
	})
	s.Votes = map[string]bool{}
	if approved {
		s.FailedMeetings = 0
		s.Pushes = map[string]bool{}
		s.enter(PhasePush, now)
		return nil
	}
	s.FailedMeetings++
	if s.FailedMeetings >= MaxFailedMeetings {
		s.finish("hackers", "meetings", now)
		return nil
	}
	s.nextLeader()
	s.Selection = nil
	s.enter(PhaseProposal, now)
	return nil
}

func (s *State) push(player string, backdoor *bool, now time.Time) error {
	if s.Phase != PhasePush {
		return ruleErr("Agora não é hora do push.")
	}
	if !slices.Contains(s.Selection, player) {
		return ruleErr("Só quem está no squad faz push.")
	}
	if backdoor == nil {
		return ruleErr("Escolha o seu push.")
	}
	if *backdoor && !s.Roles[player].IsHacker() {
		return ruleErr("Devs só fazem push limpo.")
	}
	if _, done := s.Pushes[player]; done {
		return ruleErr("Você já fez o push.")
	}
	s.Pushes[player] = *backdoor
	if len(s.Pushes) < len(s.Selection) {
		return nil
	}
	// The commits are shuffled so the order they're opened in says nothing about who
	// pushed them. Apply has no rng of its own; the time of the last push seeds one.
	s.Commits = s.Commits[:0]
	for _, p := range s.Selection {
		s.Commits = append(s.Commits, s.Pushes[p])
	}
	rng := rand.New(rand.NewPCG(uint64(now.UnixNano()), uint64(s.Sprint)))
	rng.Shuffle(len(s.Commits), func(i, j int) { s.Commits[i], s.Commits[j] = s.Commits[j], s.Commits[i] })
	s.Opened = 0
	s.Pushes = map[string]bool{}
	s.enter(PhaseReveal, now)
	return nil
}

// open shows the next commit to the whole table.
func (s *State) open(player string, now time.Time) error {
	if s.Phase != PhaseReveal {
		return ruleErr("Agora não é hora de abrir os commits.")
	}
	if player != s.leaderID() {
		return ruleErr("Só o Scrum Master abre os commits.")
	}
	if s.Opened >= len(s.Commits) {
		return ruleErr("Todos os commits já foram abertos.")
	}
	s.Opened++
	s.PhaseStartedAt = now
	return nil
}

// closeSprint records the sprint once every commit is open and moves the match on.
func (s *State) closeSprint(player string, now time.Time) error {
	if s.Phase != PhaseReveal {
		return ruleErr("Agora não é hora de fechar a sprint.")
	}
	if player != s.leaderID() {
		return ruleErr("Só o Scrum Master fecha a sprint.")
	}
	if s.Opened < len(s.Commits) {
		return ruleErr("Abra todos os commits antes.")
	}
	backdoors := 0
	for _, b := range s.Commits {
		if b {
			backdoors++
		}
	}
	s.Results = append(s.Results, SprintResult{
		Squad: slices.Clone(s.Selection), Backdoors: backdoors,
		Rollback: backdoors >= BackdoorsNeeded(len(s.Players), s.Sprint),
	})
	s.Commits, s.Opened = nil, 0
	s.Selection = nil
	switch {
	case s.rollbacks() >= ToWin:
		s.finish("hackers", "rollbacks", now)
	case s.deploys() >= ToWin:
		s.enter(PhaseGuess, now)
	default:
		s.Sprint++
		s.nextLeader()
		s.enter(PhaseProposal, now)
	}
	return nil
}

func (s *State) guess(player, target string, now time.Time) error {
	if s.Phase != PhaseGuess {
		return ruleErr("Agora não é hora do palpite.")
	}
	if s.Roles[player] != SocialEngineer {
		return ruleErr("Só a Engenharia Social dá o palpite.")
	}
	if _, ok := s.Roles[target]; !ok || target == player {
		return ruleErr("Escolha outra pessoa da partida.")
	}
	s.Guess = target
	if s.Roles[target] == Analyst {
		s.finish("hackers", "guess", now)
	} else {
		s.finish("devs", "guess_missed", now)
	}
	return nil
}
