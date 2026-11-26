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
