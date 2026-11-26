package main

import (
	"strings"
	"testing"
)

func TestNewRoomCode(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 500; i++ {
		code, err := newRoomCode()
		if err != nil {
			t.Fatal(err)
		}
		if got, ok := normalizeCode(code); !ok || got != code {
			t.Fatalf("generated code %q does not normalize to itself", code)
		}
		seen[code] = true
	}
	if len(seen) < 450 {
		t.Fatalf("too many repeated codes: %d unique of 500", len(seen))
	}
}

func TestNormalizeCode(t *testing.T) {
	cases := map[string]string{
		"kx7q":   "KX7Q",
		" KX7Q ": "KX7Q",
		"KX7":    "",
		"KX7QQ":  "",
		"KX0Q":   "", // 0 is not in the alphabet (looks like O)
		"KXIQ":   "", // I is not in the alphabet (looks like 1)
		"KX-Q":   "",
	}
	for in, want := range cases {
		got, ok := normalizeCode(in)
		if want == "" && ok {
			t.Errorf("normalizeCode(%q) = %q, want invalid", in, got)
		}
		if want != "" && (!ok || got != want) {
			t.Errorf("normalizeCode(%q) = %q, %v; want %q", in, got, ok, want)
		}
	}
	if strings.ContainsAny(codeAlphabet, "01IOL") {
		t.Fatal("alphabet has confusable characters")
	}
}

func TestCleanName(t *testing.T) {
	if got, ok := cleanName("  Ana   Paula "); !ok || got != "Ana Paula" {
		t.Errorf("got %q, %v", got, ok)
	}
	if _, ok := cleanName("   "); ok {
		t.Error("blank name accepted")
	}
	if _, ok := cleanName(strings.Repeat("á", 25)); ok {
		t.Error("25-letter name accepted")
	}
	if _, ok := cleanName(strings.Repeat("á", 24)); !ok {
		t.Error("24-letter name rejected")
	}
}

func TestNormalizeEmail(t *testing.T) {
	if got, ok := normalizeEmail(" Ana@Example.Test "); !ok || got != "ana@example.test" {
		t.Errorf("got %q, %v", got, ok)
	}
	for _, bad := range []string{"", "ana", "Ana <ana@example.test>", "a@b@c"} {
		if _, ok := normalizeEmail(bad); ok {
			t.Errorf("accepted %q", bad)
		}
	}
}

func TestCleanHandle(t *testing.T) {
	ok := map[string]string{"@Nelsul": "nelsul", " ana.paula_2 ": "ana.paula_2", "abc": "abc"}
	for in, want := range ok {
		if got, valid := cleanHandle(in); !valid || got != want {
			t.Errorf("cleanHandle(%q) = %q, %v; want %q", in, got, valid, want)
		}
	}
	for _, bad := range []string{"ab", "ana paula", "ana-paula", "joão", strings.Repeat("a", 21), "@"} {
		if _, valid := cleanHandle(bad); valid {
			t.Errorf("accepted %q", bad)
		}
	}
}

func TestOrderedPair(t *testing.T) {
	a, b := orderedPair("b", "a")
	if a != "a" || b != "b" {
		t.Fatalf("got %q, %q", a, b)
	}
}

func TestCleanMessage(t *testing.T) {
	if got, ok := cleanMessage("  oi\r\n\n\n\n gente  "); !ok || got != "oi\n\n gente" {
		t.Errorf("got %q, %v", got, ok)
	}
	if _, ok := cleanMessage(" \n "); ok {
		t.Error("blank message accepted")
	}
	if _, ok := cleanMessage(strings.Repeat("é", 301)); ok {
		t.Error("301 characters accepted")
	}
	if _, ok := cleanMessage(strings.Repeat("é", 300)); !ok {
		t.Error("300 characters rejected")
	}
}
