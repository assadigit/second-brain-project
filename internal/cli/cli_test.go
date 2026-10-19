package cli

// cli_test.go — the honest-surface tests: the flag surface (the
// parser's recognized flags), the dispatch order, the --help snapshot
// (OUR Option-B surface — the golden's help_card_reference section
// records the pre-Option-B reference for contrast), the refusal cards
// (ADR-0004 Am.3 accepted losses), the batch doors' dispatch smoke
// (T-63's flip: --auto/--sync/--retry-failed/--import-file route to
// the REAL drivers — the not-yet cards are retired with the engine),
// and the argparse-law errors.

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/assadigit/second-brain-project/internal/llmclient"
)

func testCLI(out *bytes.Buffer, in string) *CLI {
	return &CLI{
		Out:       out,
		ErrW:      out,
		In:        strings.NewReader(in),
		Now:       frozenClock,
		IsATTYIn:  func() bool { return false },
		IsATTYOut: func() bool { return false },
		IsDir:     func(string) bool { return false },
	}
}

// TestFlagSurfaceRecognized mirrors TheParserLaw's "the flag exists":
// every flag of the honest surface parses (the Telethon-only flags
// stay recognized — refusal, never unknown-flag).
func TestFlagSurfaceRecognized(t *testing.T) {
	flags := [][]string{
		{"--init"}, {"--login"}, {"--auto"}, {"--sync"},
		{"--list-dead"}, {"--reset-dead"}, {"--hand-delivery"},
		{"--chrome-retry"}, {"--status"}, {"--test-connection"},
		{"--detect-llm", "ollama"}, {"--detect-llm", "llamacpp"},
		{"--retry-failed"}, {"--mark-read"},
		{"--from-id", "1", "--to-id", "2"},
		{"--offset-start", "1", "--count", "2"},
		{"--import-file", "urls.txt"}, {"--single-id", "5"},
		{"--vault", "/v"}, {"--config", "/c"}, {"--strikes", "4"},
		{"--yes"}, {"-y"}, {"--dry-run"}, {"--no-color"},
	}
	for _, argv := range flags {
		var a Args
		fs := testCLI(&bytes.Buffer{}, "").buildFlagSet(&a)
		if err := fs.Parse(argv); err != nil {
			t.Errorf("parse %v: %v", argv, err)
		}
	}
}

// TestDispatchOrder mirrors cli_main's if-chain: the earlier flag wins.
func TestDispatchOrder(t *testing.T) {
	home := t.TempDir()
	t.Setenv("GITCURATOR_HOME", home)
	t.Setenv("GITCURATOR_USER_CONFIG", filepath.Join(home, "user-config.json"))
	cfgPath := filepath.Join(home, "dispatch-config.json")
	// a minimal config that needs no network through the battery:
	// the cloud provider with no URL (the error row), no GitHub token
	// (the warn row), the temp vault (the ok row).
	if err := os.WriteFile(cfgPath, []byte(
		`{"vault_path": `+strconv.Quote(home)+`, "llm_provider": "cloud"}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		argv []string
		want string
	}{
		{[]string{"--status", "--list-dead", "--reset-dead"}, "No config at"},
		{[]string{"--list-dead", "--reset-dead"}, "404 quarantine (threshold "},
		{[]string{"--init", "--auto", "--sync", "--config", cfgPath}, "First-run setup"},
		{[]string{"--login", "--status"}, "Telegram (MTProto)"},
		{[]string{"--mark-read", "--auto"}, "mirror-push is the read acknowledgement"},
		{[]string{"--auto", "--sync"}, "No config at"},
		{[]string{"--sync", "--retry-failed"}, "SYNC refused"},
		{[]string{"--test-connection", "--sync", "--config", cfgPath}, "Test Connection — is everything up and ready?"},
		{[]string{"--detect-llm", "ollama", "--sync", "--config", cfgPath}, "Detect & Set Ollama — probing"},
		{[]string{"--hand-delivery", "--chrome-retry", "--config", cfgPath}, "website_vault_path is not set"},
		{[]string{"--chrome-retry", "--auto", "--config", cfgPath}, "website_vault_path is not set"},
	}
	for _, tc := range cases {
		buf := &bytes.Buffer{}
		c := testCLI(buf, "")
		c.DetectOllama = func(string, float64) llmclient.DetectOllamaResult {
			return llmclient.DetectOllamaResult{Found: false,
				BaseURL: "http://127.0.0.1:11434",
				Detail:  "URLError: refused"}
		}
		c.Run(tc.argv)
		if !strings.Contains(buf.String(), tc.want) {
			t.Errorf("dispatch %v: output missing %q", tc.argv, tc.want)
		}
	}
}

// TestHelpSnapshot pins OUR honest --help surface (the Option-B set —
// the Telethon-only flags are absent from the teaching surface).
func TestHelpSnapshot(t *testing.T) {
	buf := &bytes.Buffer{}
	c := testCLI(buf, "")
	txt := c.helpText()
	for _, want := range []string{
		"usage: gitcurator-cli [-h] [--init] [--auto] [--sync]",
		"GitCurator visual CLI — colored, animated, fully automatic curation.",
		"--detect-llm {ollama,llamacpp}",
		"the sync queue (the Worker's /health)",
		"Tip: GitCurator-SYNC.bat is the fast sync lane;",
	} {
		if !strings.Contains(txt, want) {
			t.Errorf("help text lost %q", want)
		}
	}
	for _, banned := range []string{"--mark-read", "--from-id", "--to-id",
		"--offset-start", "--count N", "--single-id", "--login"} {
		if strings.Contains(txt, banned) {
			t.Errorf("help text still teaches the dropped Telethon flag %q", banned)
		}
	}
	// --help prints to stdout, exit 0 (the argparse law).
	buf2 := &bytes.Buffer{}
	c2 := testCLI(buf2, "")
	rc := c2.Run([]string{"--help"})
	if rc != 0 {
		t.Errorf("--help rc = %d, want 0", rc)
	}
	if buf2.String() != txt {
		t.Errorf("--help output differs from helpText()")
	}
}

// TestHelpCardSnapshot pins the no-flag friendly card (the honest
// Option-B surface — the Telethon-only lines dropped, the core doors
// taught).
func TestHelpCardSnapshot(t *testing.T) {
	buf := &bytes.Buffer{}
	c := testCLI(buf, "")
	rc := c.Run(nil)
	if rc != 0 {
		t.Errorf("no-flag rc = %d, want 0", rc)
	}
	out := buf.String()
	statusLine := "  " + pyLjust("--status", 24) + " vault map + config + cache + note-state summary"
	deadLine := "  " + pyLjust("--list-dead", 24) + " list the 404 quarantine (dead links)"
	resetLine := "  " + pyLjust("--reset-dead", 24) + " clear the 404 quarantine"
	for _, want := range []string{
		"Nothing to do — pick a mode:",
		statusLine,
		deadLine,
		resetLine,
		"Windows tip: double-click GitCurator-SYNC.bat for the fast sync lane,",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("help card lost %q", want)
		}
	}
	for _, banned := range []string{"--mark-read", "--single-id",
		"--from-id A --to-id B", "--offset-start A --count N", "--login"} {
		if strings.Contains(out, banned) {
			t.Errorf("help card still teaches the dropped Telethon line %q", banned)
		}
	}
}

// TestRefusalCards — the accepted losses answer the honest card with
// the ADR citation, exit 1.
func TestRefusalCards(t *testing.T) {
	cases := []struct {
		argv []string
		want string
	}{
		{[]string{"--login"}, "--login is unavailable in the Go port: the Telegram (MTProto) login flow was dropped"},
		{[]string{"--mark-read"}, "mirror-push is the read acknowledgement (spec-worker-contract §8.4)"},
		{[]string{"--single-id", "7"}, "--single-id is unavailable in the Go port: the single-message Telegram fetch was dropped"},
		{[]string{"--from-id", "1", "--to-id", "2"}, "--from-id/--to-id is unavailable in the Go port: the Telegram ID-range fetch was dropped"},
		{[]string{"--offset-start", "1", "--count", "2"}, "--offset-start/--count is unavailable in the Go port: the Telegram offset fetch was dropped"},
	}
	for _, tc := range cases {
		buf := &bytes.Buffer{}
		c := testCLI(buf, "")
		rc := c.Run(tc.argv)
		if rc != 1 {
			t.Errorf("%v: rc = %d, want 1", tc.argv, rc)
		}
		if !strings.Contains(buf.String(), tc.want) {
			t.Errorf("%v: refusal card missing %q", tc.argv, tc.want)
		}
		if !strings.Contains(buf.String(), "Accepted loss per ADR-0004 Am.3 (Option B)") {
			t.Errorf("%v: refusal card missing the ADR citation", tc.argv)
		}
	}
}

// TestBatchDispatchSmoke — T-63's flip: the four batch doors route
// to the REAL drivers (cmd_auto / cmd_sync_live / cmd_retry_failed /
// cmd_manual_batch), each answering its OWN first door honestly —
// the no-config / missing-settings cards are the drivers' own words,
// never a not-yet card (the engine landed with T-62; the launchers'
// defaults run exactly these doors). No network is touched: the
// crafted configs fail before the preflight seams would fire.
func TestBatchDispatchSmoke(t *testing.T) {
	home := t.TempDir()
	t.Setenv("GITCURATOR_HOME", home)
	t.Setenv("GITCURATOR_USER_CONFIG", filepath.Join(home, "user-config.json"))

	// a config with all five cmd_auto settings but a vault the
	// IsDir seam cannot see — the wrong-vault class of honest stop,
	// straight from the driver's own check.
	cfgPath := filepath.Join(home, "smoke-config.json")
	cfg := `{"vault_path": ` + strconv.Quote(home) +
		`, "telegram_api_id": "12345", "telegram_api_hash": "abcdef",` +
		` "telegram_phone": "+15550000", "bot_username": "curatorbot"}` + "\n"
	if err := os.WriteFile(cfgPath, []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		argv []string
		want string
	}{
		{[]string{"--auto"}, "No config at"},
		{[]string{"--auto", "--config", cfgPath}, "Missing settings: vault path"},
		{[]string{"--sync"}, "SYNC refused"},
		{[]string{"--retry-failed"}, "No config at"},
		{[]string{"--import-file", "urls.txt"}, "No config at"},
		{[]string{"--import-file", "urls.txt", "--config", cfgPath}, "Import file not found: urls.txt"},
	}
	for _, tc := range cases {
		buf := &bytes.Buffer{}
		c := testCLI(buf, "")
		rc := c.Run(tc.argv)
		if rc != 1 {
			t.Errorf("%v: rc = %d, want 1 (the honest stop)", tc.argv, rc)
		}
		if !strings.Contains(buf.String(), tc.want) {
			t.Errorf("%v: output missing %q", tc.argv, tc.want)
		}
		if strings.Contains(buf.String(), "not yet available in the Go port") {
			t.Errorf("%v: a not-yet card leaked through — the dispatch flip is incomplete", tc.argv)
		}
	}
}

// TestArgparseErrors — invalid choice / unknown flag exit 2 (the
// argparse law).
func TestArgparseErrors(t *testing.T) {
	buf := &bytes.Buffer{}
	c := testCLI(buf, "")
	if rc := c.Run([]string{"--detect-llm", "gpt4all"}); rc != 2 {
		t.Errorf("invalid choice rc = %d, want 2", rc)
	}
	if !strings.Contains(buf.String(), "invalid choice: 'gpt4all'") {
		t.Errorf("invalid choice message missing: %q", buf.String())
	}
	buf2 := &bytes.Buffer{}
	c2 := testCLI(buf2, "")
	if rc := c2.Run([]string{"--no-such-flag"}); rc != 2 {
		t.Errorf("unknown flag rc = %d, want 2", rc)
	}
}

// TestAskKeepsDefault mirrors _ask: Enter keeps the default; a typed
// value wins; EOF answers the default.
func TestAskKeepsDefault(t *testing.T) {
	buf := &bytes.Buffer{}
	c := testCLI(buf, "\n")
	if got := c.Ask("Obsidian vault folder", "/keep/me"); got != "/keep/me" {
		t.Errorf("Enter: got %q, want the default", got)
	}
	buf2 := &bytes.Buffer{}
	c2 := testCLI(buf2, "/typed\n")
	if got := c2.Ask("Obsidian vault folder", "/keep/me"); got != "/typed" {
		t.Errorf("typed: got %q, want /typed", got)
	}
	buf3 := &bytes.Buffer{}
	c3 := testCLI(buf3, "")
	if got := c3.Ask("Obsidian vault folder", "/keep/me"); got != "/keep/me" {
		t.Errorf("EOF: got %q, want the default", got)
	}
}

// TestConfigPathMirrors — _config_path's absolute/absolute law.
func filepathIsAbs(p string) bool { return filepath.IsAbs(p) }

func TestConfigPathMirrors(t *testing.T) {
	if got := configPath(""); !strings.HasSuffix(got, "config.json") {
		t.Errorf("default config path = %q, want the APP_DIR config.json", got)
	}
	if got := configPath("missing-config.json"); !filepathIsAbs(got) {
		t.Errorf("explicit path not made absolute: %q", got)
	}
}
