package backdoor

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"slices"
	"testing"
	"time"
)

var t0 = time.Date(2026, 11, 26, 21, 0, 0, 0, time.UTC)

func players(n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = fmt.Sprintf("p%d", i)
	}
	return out
}

func setup(t *testing.T, n int, opts Options, seed uint64) *State {
	t.Helper()
	s, err := Setup(players(n), opts, rand.New(rand.NewPCG(seed, 1)), t0)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func must(t *testing.T, s *State, p string, a Action) {
	t.Helper()
	if err := Apply(s, p, a, t0); err != nil {
		t.Fatalf("%s %s: %v", p, a.Type, err)
	}
}

func mustFail(t *testing.T, s *State, p string, a Action) {
	t.Helper()
	before, _ := json.Marshal(s)
	if err := Apply(s, p, a, t0); err == nil {
		t.Fatalf("%s %s: expected a rule error", p, a.Type)
	} else if _, ok := err.(*RuleError); !ok {
		t.Fatalf("%s %s: not a RuleError: %v", p, a.Type, err)
	}
	after, _ := json.Marshal(s)
	if string(before) != string(after) {
		t.Fatalf("%s %s: a rejected action changed the state", p, a.Type)
	}
}

func yes() *bool { b := true; return &b }
func no() *bool  { b := false; return &b }

func byRole(s *State, r Role) []string {
	var out []string
	for _, p := range s.Players {
		if s.Roles[p] == r {
			out = append(out, p)
		}
	}
	return out
}

func devsOnly(s *State) []string {
	var out []string
	for _, p := range s.Players {
		if !s.Roles[p].IsHacker() {
			out = append(out, p)
		}
	}
	return out
}

func hackers(s *State) []string {
	var out []string
	for _, p := range s.Players {
		if s.Roles[p].IsHacker() {
			out = append(out, p)
		}
	}
	return out
}

func allReady(t *testing.T, s *State) {
	for _, p := range s.Players {
		must(t, s, p, Action{Type: "ready"})
	}
}

// runSprint proposes squad, has everyone approve, and pushes: members in dirty push a backdoor.
func runSprint(t *testing.T, s *State, squad []string, dirty map[string]bool) {
	t.Helper()
	leader := s.leaderID()
	must(t, s, leader, Action{Type: "select", Squad: squad})
	must(t, s, leader, Action{Type: "propose"})
	for _, p := range s.Players {
		must(t, s, p, Action{Type: "vote", Approve: yes()})
	}
	for _, p := range squad {
		b := dirty[p]
		must(t, s, p, Action{Type: "push", Backdoor: &b})
	}
}

func TestRolesPerTableSize(t *testing.T) {
	for n := MinPlayers; n <= MaxPlayers; n++ {
		s := setup(t, n, Options{}, uint64(n))
		if got := len(hackers(s)); got != hackersFor[n] {
			t.Errorf("%d players: %d hackers, want %d", n, got, hackersFor[n])
		}
		if len(byRole(s, Analyst)) != 1 || len(byRole(s, SocialEngineer)) != 1 {
			t.Errorf("%d players: the Analyst and the Social Engineer must always be in play", n)
		}
		if len(byRole(s, ZeroDay))+len(byRole(s, Freelancer))+len(byRole(s, ProductOwner)) != 0 {
			t.Errorf("%d players: optional roles dealt without being turned on", n)
		}
	}
	s := setup(t, 10, Options{ZeroDay: true, Freelancer: true, ProductOwner: true}, 7)
	if len(hackers(s)) != 4 || len(byRole(s, ZeroDay)) != 1 || len(byRole(s, Freelancer)) != 1 || len(byRole(s, ProductOwner)) != 1 {
		t.Fatalf("optional roles: %v", s.Roles)
	}
}

func TestSetupChecksOptions(t *testing.T) {
	if _, err := Setup(players(6), Options{ZeroDay: true}, rand.New(rand.NewPCG(1, 1)), t0); err == nil {
		t.Error("Zero-day accepted with 6 players")
	}
	if _, err := Setup(players(6), Options{Freelancer: true}, rand.New(rand.NewPCG(1, 1)), t0); err == nil {
		t.Error("Freelancer accepted with 6 players")
	}
	if _, err := Setup(players(4), Options{}, rand.New(rand.NewPCG(1, 1)), t0); err == nil {
		t.Error("4 players accepted")
	}
	if _, err := Setup(players(11), Options{}, rand.New(rand.NewPCG(1, 1)), t0); err == nil {
		t.Error("11 players accepted")
	}
	if err := (Options{TimerSeconds: 45}).CheckLobby(); err == nil {
		t.Error("odd timer accepted")
	}
	if _, err := Setup(players(5), Options{ProductOwner: true, TimerSeconds: 120}, rand.New(rand.NewPCG(1, 1)), t0); err != nil {
		t.Errorf("Product Owner with 5 players: %v", err)
	}
}

func TestBadgesThenProposal(t *testing.T) {
	s := setup(t, 5, Options{}, 1)
	mustFail(t, s, s.Players[0], Action{Type: "propose"})
	for _, p := range s.Players[:4] {
		must(t, s, p, Action{Type: "ready"})
	}
	if s.Phase != PhaseBadges {
		t.Fatal("started before everyone was ready")
	}
	must(t, s, s.Players[4], Action{Type: "ready"})
	if s.Phase != PhaseProposal || s.Sprint != 0 {
		t.Fatalf("phase %s sprint %d", s.Phase, s.Sprint)
	}
	mustFail(t, s, "stranger", Action{Type: "ready"})
}

func TestOnlyScrumMasterBuildsTheSquad(t *testing.T) {
	s := setup(t, 5, Options{}, 2)
	allReady(t, s)
	other := s.Players[(s.Leader+1)%5]
	mustFail(t, s, other, Action{Type: "select", Squad: s.Players[:2]})
	mustFail(t, s, s.leaderID(), Action{Type: "select", Squad: s.Players[:3]})                        // too many for sprint 1
	mustFail(t, s, s.leaderID(), Action{Type: "select", Squad: []string{s.Players[0], s.Players[0]}}) // repeated
	mustFail(t, s, s.leaderID(), Action{Type: "select", Squad: []string{"stranger"}})
	must(t, s, s.leaderID(), Action{Type: "select", Squad: s.Players[:1]})
	mustFail(t, s, s.leaderID(), Action{Type: "propose"}) // wrong size
	must(t, s, s.leaderID(), Action{Type: "select", Squad: s.Players[:2]})
	mustFail(t, s, other, Action{Type: "propose"})
	must(t, s, s.leaderID(), Action{Type: "propose"})
	if s.Phase != PhaseReview {
		t.Fatal("not in review")
	}
}

func TestReviewNeedsStrictMajorityAndRotates(t *testing.T) {
	s := setup(t, 6, Options{}, 3)
	allReady(t, s)
	first := s.Leader
	must(t, s, s.leaderID(), Action{Type: "select", Squad: s.Players[:2]})
	must(t, s, s.leaderID(), Action{Type: "propose"})
	for i, p := range s.Players { // 3 of 6: a tie is not a majority
		must(t, s, p, Action{Type: "vote", Approve: map[bool]*bool{true: yes(), false: no()}[i < 3]})
	}
	if s.Phase != PhaseProposal || s.FailedMeetings != 1 || s.Leader != (first+1)%6 {
		t.Fatalf("phase %s, failed %d, leader %d", s.Phase, s.FailedMeetings, s.Leader)
	}
	p := s.Proposals[0]
	if p.Approved || len(p.Votes) != 6 {
		t.Fatalf("proposal record: %+v", p)
	}
	if len(s.Selection) != 0 {
		t.Fatal("the next Scrum Master must start from an empty squad")
	}
	mustFail(t, s, s.Players[0], Action{Type: "vote", Approve: yes()})
}

func TestCannotVoteTwice(t *testing.T) {
	s := setup(t, 5, Options{}, 4)
	allReady(t, s)
	must(t, s, s.leaderID(), Action{Type: "select", Squad: s.Players[:2]})
	must(t, s, s.leaderID(), Action{Type: "propose"})
	must(t, s, s.Players[0], Action{Type: "vote", Approve: yes()})
	mustFail(t, s, s.Players[0], Action{Type: "vote", Approve: no()})
	mustFail(t, s, s.Players[1], Action{Type: "vote"})
}

func TestFiveFailedMeetingsHackersWin(t *testing.T) {
	s := setup(t, 5, Options{}, 5)
	allReady(t, s)
	for m := 0; m < MaxFailedMeetings; m++ {
		must(t, s, s.leaderID(), Action{Type: "select", Squad: s.Players[:2]})
		must(t, s, s.leaderID(), Action{Type: "propose"})
		for _, p := range s.Players {
			must(t, s, p, Action{Type: "vote", Approve: no()})
		}
	}
	if s.Phase != PhaseOver || s.Winner != "hackers" || s.Reason != "meetings" {
		t.Fatalf("phase %s winner %s reason %s", s.Phase, s.Winner, s.Reason)
	}
}

func TestApprovalResetsFailedMeetings(t *testing.T) {
	s := setup(t, 5, Options{}, 6)
	allReady(t, s)
	for m := 0; m < 4; m++ {
		must(t, s, s.leaderID(), Action{Type: "select", Squad: s.Players[:2]})
		must(t, s, s.leaderID(), Action{Type: "propose"})
		for _, p := range s.Players {
			must(t, s, p, Action{Type: "vote", Approve: no()})
		}
	}
	runSprint(t, s, devsOnly(s)[:2], nil)
	if s.FailedMeetings != 0 || s.Sprint != 1 || s.Results[0].Rollback {
		t.Fatalf("failed %d sprint %d results %+v", s.FailedMeetings, s.Sprint, s.Results)
	}
}

func TestPushRules(t *testing.T) {
	s := setup(t, 5, Options{}, 7)
	allReady(t, s)
	dev := devsOnly(s)[0]
	hacker := hackers(s)[0]
	outsider := ""
	for _, p := range s.Players {
		if p != dev && p != hacker {
			outsider = p
		}
	}
	must(t, s, s.leaderID(), Action{Type: "select", Squad: []string{dev, hacker}})
	must(t, s, s.leaderID(), Action{Type: "propose"})
	for _, p := range s.Players {
		must(t, s, p, Action{Type: "vote", Approve: yes()})
	}
	mustFail(t, s, dev, Action{Type: "push", Backdoor: yes()})     // devs can't backdoor
	mustFail(t, s, outsider, Action{Type: "push", Backdoor: no()}) // not in the squad
	mustFail(t, s, dev, Action{Type: "push"})                      // no choice
	must(t, s, dev, Action{Type: "push", Backdoor: no()})
	mustFail(t, s, dev, Action{Type: "push", Backdoor: no()}) // twice
	must(t, s, hacker, Action{Type: "push", Backdoor: yes()})
	r := s.Results[0]
	if !r.Rollback || r.Backdoors != 1 || s.Sprint != 1 || s.Phase != PhaseProposal {
		t.Fatalf("result %+v sprint %d phase %s", r, s.Sprint, s.Phase)
	}
}

func TestSprintFourNeedsTwoBackdoorsWithSeven(t *testing.T) {
	s := setup(t, 7, Options{}, 8)
	allReady(t, s)
	h := hackers(s)
	d := devsOnly(s)
	runSprint(t, s, d[:2], nil)                                                    // deploy
	runSprint(t, s, append([]string{h[0]}, d[:2]...), map[string]bool{h[0]: true}) // rollback
	runSprint(t, s, append([]string{h[1]}, d[:2]...), map[string]bool{h[1]: true}) // rollback
	runSprint(t, s, append([]string{h[0]}, d[:3]...), map[string]bool{h[0]: true}) // 1 backdoor in sprint 4: deploy
	if got := s.Results[3]; got.Rollback || got.Backdoors != 1 {
		t.Fatalf("sprint 4 with one backdoor: %+v", got)
	}
	runSprint(t, s, append([]string{h[0], h[1]}, d[:2]...), map[string]bool{h[0]: true}) // sprint 5: rollback
	if s.Phase != PhaseOver || s.Winner != "hackers" || s.Reason != "rollbacks" {
		t.Fatalf("phase %s winner %s reason %s", s.Phase, s.Winner, s.Reason)
	}
}

func TestDeploysLeadToTheSocialEngineersGuess(t *testing.T) {
	for _, hit := range []bool{true, false} {
		s := setup(t, 5, Options{}, 9)
		allReady(t, s)
		d := devsOnly(s)
		runSprint(t, s, d[:2], nil)
		runSprint(t, s, d[:3], nil)
		runSprint(t, s, d[:2], nil)
		if s.Phase != PhaseGuess {
			t.Fatalf("phase %s after 3 deploys", s.Phase)
		}
		se := byRole(s, SocialEngineer)[0]
		analyst := byRole(s, Analyst)[0]
		mustFail(t, s, analyst, Action{Type: "guess", Target: se})
		mustFail(t, s, se, Action{Type: "guess", Target: se})
		target := analyst
		if !hit {
			for _, p := range d {
				if p != analyst {
					target = p
					break
				}
			}
		}
		must(t, s, se, Action{Type: "guess", Target: target})
		want := map[bool]string{true: "hackers", false: "devs"}[hit]
		if s.Phase != PhaseOver || s.Winner != want {
			t.Fatalf("hit=%v: winner %s", hit, s.Winner)
		}
	}
}

func TestViewKeepsSecrets(t *testing.T) {
	s := setup(t, 10, Options{ZeroDay: true, Freelancer: true, ProductOwner: true}, 10)
	zero := byRole(s, ZeroDay)[0]
	free := byRole(s, Freelancer)[0]
	analyst := byRole(s, Analyst)[0]
	se := byRole(s, SocialEngineer)[0]

	ids := func(ks []Known) []string {
		var out []string
		for _, k := range ks {
			out = append(out, k.ID)
		}
		slices.Sort(out)
		return out
	}
	for _, p := range s.Players {
		v := ViewFor(s, p)
		if v.Roles != nil {
			t.Fatal("roles leaked before the end")
		}
		if v.Me.Role != s.Roles[p] {
			t.Fatalf("%s sees the wrong own role", p)
		}
		switch s.Roles[p] {
		case Dev, Freelancer:
			if len(v.Me.Knows) != 0 {
				t.Errorf("%s (%s) knows %v", p, s.Roles[p], v.Me.Knows)
			}
		case Analyst:
			got := ids(v.Me.Knows)
			if slices.Contains(got, zero) || len(got) != 3 || !slices.Contains(got, free) {
				t.Errorf("Analyst sees %v (zero-day %s, freelancer %s)", got, zero, free)
			}
		case ProductOwner:
			got := ids(v.Me.Knows)
			want := []string{analyst, zero}
			slices.Sort(want)
			if !slices.Equal(got, want) {
				t.Errorf("Product Owner sees %v, want %v", got, want)
			}
			for _, k := range v.Me.Knows {
				if k.As != "analyst_candidate" {
					t.Errorf("Product Owner learns too much: %+v", k)
				}
			}
		case Hacker, SocialEngineer, ZeroDay:
			got := ids(v.Me.Knows)
			if slices.Contains(got, free) || len(got) != 2 {
				t.Errorf("%s (%s) sees %v", p, s.Roles[p], got)
			}
			if p != se && !slices.Contains(got, se) {
				t.Errorf("%s doesn't know the Social Engineer", p)
			}
		}
		if v.SocialEngineer != "" {
			t.Error("Social Engineer public before the guess")
		}
	}
}

func TestViewHidesWhoPushedWhat(t *testing.T) {
	s := setup(t, 5, Options{}, 11)
	allReady(t, s)
	h := hackers(s)[0]
	d := devsOnly(s)[0]
	must(t, s, s.leaderID(), Action{Type: "select", Squad: []string{h, d}})
	must(t, s, s.leaderID(), Action{Type: "propose"})
	for _, p := range s.Players {
		must(t, s, p, Action{Type: "vote", Approve: yes()})
	}
	must(t, s, h, Action{Type: "push", Backdoor: yes()})
	for _, p := range s.Players {
		raw, _ := json.Marshal(ViewFor(s, p))
		var generic map[string]any
		_ = json.Unmarshal(raw, &generic)
		if _, ok := generic["pushes"]; ok {
			t.Fatal("individual pushes in the view")
		}
	}
	v := ViewFor(s, d)
	if !v.Me.InSquad || v.Me.Pushed || v.Me.CanBackdoor {
		t.Fatalf("dev in squad: %+v", v.Me)
	}
	for _, pv := range v.Players {
		if pv.ID == h && !pv.Pushed {
			t.Fatal("others should see that the hacker already pushed (not what)")
		}
	}
}

func TestViewAfterTheEndRevealsEverything(t *testing.T) {
	s := setup(t, 5, Options{}, 12)
	allReady(t, s)
	h := hackers(s)
	d := devsOnly(s)
	runSprint(t, s, []string{h[0], d[0]}, map[string]bool{h[0]: true})
	runSprint(t, s, []string{h[0], d[0], d[1]}, map[string]bool{h[0]: true})
	runSprint(t, s, []string{h[1], d[0]}, map[string]bool{h[1]: true})
	v := ViewFor(s, d[0])
	if v.Phase != PhaseOver || len(v.Roles) != 5 || v.SocialEngineer == "" || len(v.Proposals) != 3 {
		t.Fatalf("end view: %+v", v)
	}
}
