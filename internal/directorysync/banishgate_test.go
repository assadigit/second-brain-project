// banishgate_test.go — T-49: the ported selections of the reference's
// TestBanish (test_directorysync.py) — the host-level gate with the
// shared-host exception law. The gate itself landed with T-40; these
// are its reference suite's own tests.
package directorysync

import "testing"

var banishFixtures = []BanishEntry{
	{Host: "banished.example", Reason: "test", BanishedAt: "2026-10-01"},
	{Host: "github.io", URL: "https://someone.github.io/bad/",
		Reason: "url-level", BanishedAt: "2026-10-01"},
}

func TestBanish_HostBanBlocksSubdomains(t *testing.T) {
	if !IsBanished("https://banished.example/", banishFixtures, nil) {
		t.Fatal("host ban")
	}
	if !IsBanished("https://www.banished.example/page", banishFixtures, nil) {
		t.Fatal("www subdomain")
	}
	if !IsBanished("https://sub.banished.example/x", banishFixtures, nil) {
		t.Fatal("subdomain")
	}
}

func TestBanish_UnrelatedHostPasses(t *testing.T) {
	if IsBanished("https://good.example/", banishFixtures, nil) {
		t.Fatal("unrelated host blocked")
	}
}

func TestBanish_SharedHostOnlyBansExactURL(t *testing.T) {
	if !IsBanished("https://someone.github.io/bad/", banishFixtures, nil) {
		t.Fatal("the exact URL")
	}
	if IsBanished("https://someone.github.io/good/", banishFixtures, nil) {
		t.Fatal("a good URL on the shared host")
	}
	if IsBanished("https://other.github.io/bad/", banishFixtures, nil) {
		t.Fatal("another tenant with the banned path")
	}
}

func TestBanish_EmptyListPassesEverything(t *testing.T) {
	if IsBanished("https://anything.example/", nil, nil) {
		t.Fatal("empty list blocked")
	}
}
