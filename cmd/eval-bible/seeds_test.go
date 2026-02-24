package main

import (
	"testing"
)

func TestParseXRef(t *testing.T) {
	tsv := "From Verse\tTo Verse\tVotes\nGen.1.1\tHeb.11.3\t172\nGen.1.1\tJohn.1.1\t151\nMatt.5.3\tLuke.6.20\t88\n"
	xrefs, err := parseXRef([]byte(tsv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	gen11 := xrefs["Genesis 1:1"]
	if len(gen11) != 2 {
		t.Fatalf("want 2 cross-refs for Genesis 1:1, got %d", len(gen11))
	}
	found := map[string]bool{}
	for _, r := range gen11 {
		found[r] = true
	}
	if !found["Hebrews 11:3"] {
		t.Error("missing Hebrews 11:3")
	}
	if !found["John 1:1"] {
		t.Error("missing John 1:1")
	}
}

func TestAbbrevToBook(t *testing.T) {
	cases := [][2]string{
		{"Gen", "Genesis"}, {"Matt", "Matthew"}, {"Rev", "Revelation"},
		{"1Cor", "1 Corinthians"}, {"Ps", "Psalms"},
	}
	for _, c := range cases {
		if got := abbrevToBook(c[0]); got != c[1] {
			t.Errorf("abbrevToBook(%q) = %q, want %q", c[0], got, c[1])
		}
	}
}

func TestSelectSeeds(t *testing.T) {
	xrefs := map[string][]string{
		"John 3:16":  {"r1", "r2", "r3", "r4", "r5", "r6"},
		"John 3:17":  {"r1", "r2"},
		"Romans 5:8": {"r1", "r2", "r3", "r4", "r5", "r6"},
	}
	seeds := selectSeeds(xrefs, 5, 10)
	if len(seeds) != 2 {
		t.Errorf("want 2 seeds (5+ cross-refs), got %d", len(seeds))
	}
}
