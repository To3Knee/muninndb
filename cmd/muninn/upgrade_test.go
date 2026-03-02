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

func TestIsHomebrewInstall(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{"/opt/homebrew/Cellar/muninn/1.0.0/bin/muninn", true},
		{"/usr/local/opt/muninn/bin/muninn", true},
		{"/opt/homebrew/bin/muninn", true},
		{"/usr/local/Cellar/muninn/1.0.0/bin/muninn", true},
		{"/home/user/.local/bin/muninn", false},
		{"/usr/local/bin/muninn", false},
		{"/tmp/muninn", false},
	}
	for _, tc := range cases {
		got := isHomebrewInstallPath(tc.path)
		if got != tc.want {
			t.Errorf("isHomebrewInstallPath(%q) = %v, want %v", tc.path, got, tc.want)
		}
	}
}

func TestReleaseAssetURL(t *testing.T) {
	url := releaseAssetURL("v1.2.3", "darwin", "arm64")
	want := "https://github.com/scrypster/muninndb/releases/download/v1.2.3/muninn_v1.2.3_darwin_arm64.tar.gz"
	if url != want {
		t.Errorf("got %q, want %q", url, want)
	}

	url = releaseAssetURL("v1.2.3", "linux", "amd64")
	want = "https://github.com/scrypster/muninndb/releases/download/v1.2.3/muninn_v1.2.3_linux_amd64.tar.gz"
	if url != want {
		t.Errorf("got %q, want %q", url, want)
	}

	url = releaseAssetURL("v1.2.3", "windows", "amd64")
	want = "https://github.com/scrypster/muninndb/releases/download/v1.2.3/muninn_v1.2.3_windows_amd64.zip"
	if url != want {
		t.Errorf("got %q, want %q", url, want)
	}
}
