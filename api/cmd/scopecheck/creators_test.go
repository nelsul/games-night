package main

import (
	"errors"
	"testing"
)

func creatorsReport(src string, lookup userLookup) report {
	var r report
	checkCreators([]string{"quiz"}, func(p string) string {
		if p == "web/src/lib/games/quiz/index.ts" {
			return src
		}
		return ""
	}, lookup, &r)
	return r
}

func TestCreatorsAreRequired(t *testing.T) {
	wantErr(t, creatorsReport("export const info = { id: 'quiz' };", nil), "has no `creators` list")
	wantErr(t, creatorsReport("creators: [],", nil), "is empty")
	wantErr(t, creatorsReport("", nil), "missing")
}

func TestCreatorsMustBeGitHubUsernames(t *testing.T) {
	wantErr(t, creatorsReport("creators: ['@nelsul'],", nil), "without the @")
	wantErr(t, creatorsReport("creators: ['nel sul'],", nil), "not a valid GitHub username")
	wantErr(t, creatorsReport("creators: ['-nelsul'],", nil), "not a valid GitHub username")
	wantErr(t, creatorsReport("creators: ['nelsul', 'NelSul'],", nil), "listed twice")
	wantErr(t, creatorsReport("creators: [author],", nil), "plain list of quoted")
	wantErr(t, creatorsReport("creators: ['nelsul', ...others],", nil), "plain list of quoted")
}

func TestCreatorsMustExist(t *testing.T) {
	lookup := func(n string) (bool, error) {
		switch n {
		case "ghost-friend":
			return false, nil
		case "flaky":
			return false, errors.New("timeout")
		}
		return true, nil
	}
	wantErr(t, creatorsReport("creators: ['nelsul', 'ghost-friend'],", lookup), "is not a GitHub user")
	wantErr(t, creatorsReport("creators: ['flaky'],", lookup), "could not check")
}

func TestValidCreatorsPass(t *testing.T) {
	src := "export const info = {\n\tid: 'quiz',\n\tcreators: [\n\t\t'nelsul',\n\t\t\"a-friend\"\n\t],\n};"
	r := creatorsReport(src, func(string) (bool, error) { return true, nil })
	if len(r.errors) > 0 {
		t.Fatalf("errors: %v", r.errors)
	}
}
