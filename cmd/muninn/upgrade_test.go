package main

import "testing"

func TestNewerVersionAvailable(t *testing.T) {
	cases := []struct {
		current, latest string
		want            bool
	}{
		{"v1.0.0", "v1.0.1", true},
		{"v1.0.0", "v1.0.0", false},
		{"v1.0.1", "v1.0.0", false},
		{"v1.2.0", "v2.0.0", true},
		{"dev", "v1.0.0", false},
		{"", "v1.0.0", false},
		{"v1.0.0", "", false},
	}
	for _, tc := range cases {
		got := newerVersionAvailable(tc.current, tc.latest)
		if got != tc.want {
			t.Errorf("newerVersionAvailable(%q, %q) = %v, want %v", tc.current, tc.latest, got, tc.want)
		}
	}
}
