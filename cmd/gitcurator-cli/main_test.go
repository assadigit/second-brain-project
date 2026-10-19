package main

// main_test.go — the main package's own doors: the --cli strip law
// (main.py's routing, ported to the exe) and the --version stamp door
// (doc 10 §5 — the Go packaging addition; the reference's argparse has
// no --version, a documented divergence).

import (
	"bytes"
	"strings"
	"testing"
)

func TestVersionDoor(t *testing.T) {
	cases := [][]string{
		{"--version"},
		{"-version"},
		{"--version", "--status"},
		{"--cli", "--version"}, // the strip happens first (main.py's law)
	}
	for _, argv := range cases {
		buf := &bytes.Buffer{}
		if rc := run(buf, argv); rc != 0 {
			t.Errorf("%v: rc = %d, want 0", argv, rc)
		}
		want := "gitcurator-cli dev\n"
		if buf.String() != want {
			t.Errorf("%v: output = %q, want %q", argv, buf.String(), want)
		}
	}
}

// TestVersionStampFeedsTheDoor — the stamp mechanism is real: changing
// the version variable changes the printed stamp (the -ldflags
// -X main.version=… build lands here).
func TestVersionStampFeedsTheDoor(t *testing.T) {
	old := version
	version = "0.92.0-go.1"
	defer func() { version = old }()
	buf := &bytes.Buffer{}
	if rc := run(buf, []string{"--version"}); rc != 0 {
		t.Fatalf("rc = %d, want 0", rc)
	}
	if buf.String() != "gitcurator-cli 0.92.0-go.1\n" {
		t.Errorf("stamped output = %q", buf.String())
	}
}

// TestCLIStripLaw — "--cli" never reaches the CLI's parser: it is a
// spelling of the router, stripped verbatim (main.py: "[a for a in
// sys.argv[1:] if a != '--cli']"). Proof: --cli alone routes to the
// CLI's no-flag door (the friendly help card, exit 0) instead of the
// unknown-flag error (exit 2), and --cli --login still finds --login.
func TestCLIStripLaw(t *testing.T) {
	// --cli alone: after the strip the CLI sees an empty argv → the
	// help card (exit 0). The banner goes to the real stdout (the CLI
	// binds its own streams — the test only asserts the exit code).
	if rc := run(&bytes.Buffer{}, []string{"--cli"}); rc != 0 {
		t.Errorf("--cli alone: rc = %d, want 0 (the help card — the flag was stripped)", rc)
	}
	// --cli --version: the strip happens BEFORE the version door, so
	// the version door wins over any CLI parse.
	buf := &bytes.Buffer{}
	run(buf, []string{"--status", "--cli", "--version"})
	if !strings.HasPrefix(buf.String(), "gitcurator-cli ") {
		t.Errorf("--status --cli --version: the strip must keep the version door first, got %q", buf.String())
	}
}
