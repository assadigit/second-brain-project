package cli

// cli.go — the command surface + dispatch (cli.py's build_parser /
// cli_main, the Option-B honest surface per ADR-0004 Am.3 / rider R-0).
//
// THE HONEST FLAG SURFACE (T-60's decision, never a silent no-op):
//   - IMPLEMENTED in this slice: --status, --list-dead, --reset-dead,
//     --config, --vault, --yes/-y, --no-color, --dry-run, --strikes
//     (the last four parsed everywhere; the batch engines that consume
//     them land with T-62).
//   - NOT-YET cards (explicit, exit 1): --init (the wizard), the
//     --sync door's live beats (its pure CHECK half runs now; the
//     battery/button/fetch engine lands with T-62), --auto,
//     --retry-failed, --import-file, --test-connection, --detect-llm,
//     --hand-delivery, --chrome-retry.
//   - REFUSAL cards (ADR-0004 Am.3 accepted losses, exit 1): --login,
//     --mark-read, --from-id/--to-id, --offset-start/--count,
//     --single-id — the Telethon-only modes; the Go app ships no
//     MTProto client, the queue rides the Worker HTTP API, and
//     mirror-push is the read acknowledgement.

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/assadigit/second-brain-project/internal/connectioncheck"
	"github.com/assadigit/second-brain-project/internal/llmclient"
)

// CLI carries the shell's seams: the output/input streams, the clock
// (the frozen-clock law), the TTY doors, and the status card's probe
// seams (the instance-seam pattern).
type CLI struct {
	Out       io.Writer // stdout
	ErrW      io.Writer // stderr (the argparse error path)
	In        io.Reader // stdin (the prompt doors)
	Now       func() time.Time
	IsATTYIn  func() bool
	IsATTYOut func() bool

	lineReader *bufio.Reader

	// The status card's seams (the _gui_symbol / net_guard / ollama
	// import doors map to injectable functions here).
	IsDir         func(string) bool
	ProbeInternet func(cfg any, timeoutS float64) connectioncheck.ProbeInternetResult
	AIStatus      func(cfg any, timeoutS float64) connectioncheck.AIStatusResult
	OllamaModels  func(baseURL string, timeoutS float64) ([]string, error)
	ProbeLlama    func(baseURL, apiKey string, timeoutS float64) *llmclient.ProbeResult
	DetectLlama   func(apiKey string) *llmclient.ProbeResult
}

// NewCLI builds the default shell bound to the real streams and the
// ported engine packages.
func NewCLI() *CLI {
	return &CLI{
		Out:           os.Stdout,
		ErrW:          os.Stderr,
		In:            os.Stdin,
		Now:           time.Now,
		IsATTYIn:      isATTYStdin,
		IsATTYOut:     isATTYStdout,
		IsDir:         func(p string) bool { fi, err := os.Stat(p); return err == nil && fi.IsDir() },
		ProbeInternet: defaultProbeInternet,
		AIStatus:      defaultAIStatus,
		OllamaModels:  defaultOllamaModels,
		ProbeLlama:    defaultProbeLlama,
		DetectLlama:   defaultDetectLlama,
	}
}

// Main is the CLI entry (called by cmd/gitcurator-cli).
func Main(argv []string) int {
	c := NewCLI()
	return c.Run(argv)
}

// Args is the parsed surface (argparse's Namespace twin).
type Args struct {
	Init, Login, Auto, Sync   bool
	ListDead, ResetDead       bool
	HandDelivery, ChromeRetry bool
	Status, TestConnection    bool
	RetryFailed, MarkRead     bool
	Yes, DryRun, NoColor      bool
	DetectLLM                 string
	FromID, ToID              *int
	OffsetStart, Count        *int
	SingleID                  *int
	ImportFile, Vault, Config string
	Strikes                   *int

	// the raw optional-int holders (resolved onto the *int fields
	// AFTER Parse — argparse's Namespace carries present-vs-absent)
	fromID, toID       *optInt
	offsetStart, count *optInt
	singleID, strikes  *optInt
}

// optInt is an optional int flag (argparse type=int, default None).
type optInt struct {
	v   int
	set bool
}

func (o *optInt) String() string {
	if o == nil || !o.set {
		return ""
	}
	return strconv.Itoa(o.v)
}

func (o *optInt) Set(s string) error {
	n, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	o.v, o.set = n, true
	return nil
}

func (o *optInt) ptr() *int {
	if o == nil || !o.set {
		return nil
	}
	v := o.v
	return &v
}

// buildFlagSet mirrors build_parser's surface (the honest Option-B
// set: the Telethon-only flags stay recognized so old scripts get the
// refusal card, never an unknown-flag crash).
func (c *CLI) buildFlagSet(a *Args) *flag.FlagSet {
	fs := flag.NewFlagSet("gitcurator-cli", flag.ContinueOnError)
	fs.SetOutput(c.ErrW)
	fs.BoolVar(&a.Init, "init", false, "")
	fs.BoolVar(&a.Login, "login", false, "")
	fs.BoolVar(&a.Auto, "auto", false, "")
	fs.BoolVar(&a.Sync, "sync", false, "")
	fs.BoolVar(&a.ListDead, "list-dead", false, "")
	fs.BoolVar(&a.ResetDead, "reset-dead", false, "")
	fs.BoolVar(&a.HandDelivery, "hand-delivery", false, "")
	fs.BoolVar(&a.ChromeRetry, "chrome-retry", false, "")
	fs.BoolVar(&a.Status, "status", false, "")
	fs.BoolVar(&a.TestConnection, "test-connection", false, "")
	fs.StringVar(&a.DetectLLM, "detect-llm", "", "")
	fs.BoolVar(&a.RetryFailed, "retry-failed", false, "")
	fs.BoolVar(&a.MarkRead, "mark-read", false, "")
	fromID := &optInt{}
	fs.Var(fromID, "from-id", "")
	a.fromID = fromID
	toID := &optInt{}
	fs.Var(toID, "to-id", "")
	a.toID = toID
	offsetStart := &optInt{}
	fs.Var(offsetStart, "offset-start", "")
	a.offsetStart = offsetStart
	count := &optInt{}
	fs.Var(count, "count", "")
	a.count = count
	fs.StringVar(&a.ImportFile, "import-file", "", "")
	singleID := &optInt{}
	fs.Var(singleID, "single-id", "")
	a.singleID = singleID
	fs.StringVar(&a.Vault, "vault", "", "")
	fs.StringVar(&a.Config, "config", "", "")
	strikes := &optInt{}
	fs.Var(strikes, "strikes", "")
	a.strikes = strikes
	fs.BoolVar(&a.Yes, "yes", false, "")
	fs.BoolVar(&a.Yes, "y", false, "")
	fs.BoolVar(&a.DryRun, "dry-run", false, "")
	fs.BoolVar(&a.NoColor, "no-color", false, "")
	return fs
}

// hasHelpArg answers whether argv explicitly asks for --help (any of
// argparse's spellings) — printed to stdout, exit 0, before parsing.
func hasHelpArg(argv []string) bool {
	for _, a := range argv {
		if a == "-h" || a == "--help" || a == "-help" || a == "--h" {
			return true
		}
	}
	return false
}

// Run mirrors cli_main: parse, the color door, then the dispatch order
// (init → login → status → test-connection → detect-llm → list-dead →
// reset-dead → hand-delivery → chrome-retry → mark-read → auto → sync
// → retry-failed → manual batch; no flag → the friendly help card).
func (c *CLI) Run(argv []string) int {
	hardenStdio()
	if hasHelpArg(argv) {
		fmt.Fprint(c.Out, c.helpText())
		return 0
	}
	var a Args
	fs := c.buildFlagSet(&a)
	if err := fs.Parse(argv); err != nil {
		return 2
	}
	a.FromID = a.fromID.ptr()
	a.ToID = a.toID.ptr()
	a.OffsetStart = a.offsetStart.ptr()
	a.Count = a.count.ptr()
	a.SingleID = a.singleID.ptr()
	a.Strikes = a.strikes.ptr()
	if a.DetectLLM != "" && a.DetectLLM != "ollama" && a.DetectLLM != "llamacpp" {
		fmt.Fprintf(c.ErrW, "gitcurator-cli: error: argument --detect-llm: invalid choice: '%s' (choose from 'ollama', 'llamacpp')\n", a.DetectLLM)
		return 2
	}

	EnableColors(false)
	if a.NoColor {
		DisableColors()
	}

	switch {
	case a.Init:
		return c.notYetCard("--init", "the first-run wizard lands with the rest of T-60 (Phase 5)")
	case a.Login:
		return c.refusalCard("--login",
			"the Telegram (MTProto) login flow was dropped in the Go port — the bot queue rides the Cloudflare Worker HTTP API, and no Telegram session is ever needed")
	case a.Status:
		return c.cmdStatus(&a)
	case a.TestConnection:
		return c.notYetCard("--test-connection", "the connection battery's CLI wiring lands with the rest of T-60 (Phase 5)")
	case a.DetectLLM != "":
		return c.notYetCard("--detect-llm", "the Detect & Set engine lands with the rest of T-60 (Phase 5)")
	case a.ListDead:
		return c.cmdListDead(&a)
	case a.ResetDead:
		return c.cmdResetDead(&a)
	case a.HandDelivery:
		return c.notYetCard("--hand-delivery", "the fourth door lands with the rest of T-60 (Phase 5)")
	case a.ChromeRetry:
		return c.notYetCard("--chrome-retry", "the fifth door lands with the rest of T-60 (Phase 5)")
	case a.MarkRead:
		return c.refusalCard("--mark-read",
			"marking the Telethon bot messages read was dropped in the Go port — mirror-push is the read acknowledgement (spec-worker-contract §8.4)")
	case a.Auto:
		return c.notYetCard("--auto", "the batch engine lands with T-62 (Phase 5)")
	case a.Sync:
		return c.cmdSync(&a)
	case a.RetryFailed:
		return c.notYetCard("--retry-failed", "the batch engine lands with T-62 (Phase 5)")
	case a.ImportFile != "" || a.SingleID != nil || a.FromID != nil ||
		a.ToID != nil || a.OffsetStart != nil || a.Count != nil:
		return c.manualBatchCard(&a)
	}
	return c.helpCard()
}

// manualBatchCard routes the manual batch modes: the import file is
// T-62's engine; the Telegram ID modes are Option-B refusals.
func (c *CLI) manualBatchCard(a *Args) int {
	if a.ImportFile != "" {
		return c.notYetCard("--import-file", "the batch engine lands with T-62 (Phase 5)")
	}
	flag := "--single-id"
	detail := "the single-message Telegram fetch was dropped in the Go port (ADR-0004 Am.3 — Option B); import links with --import-file instead"
	if a.FromID != nil || a.ToID != nil {
		flag = "--from-id/--to-id"
		detail = "the Telegram ID-range fetch was dropped in the Go port (ADR-0004 Am.3 — Option B); import links with --import-file instead"
	} else if a.OffsetStart != nil || a.Count != nil {
		flag = "--offset-start/--count"
		detail = "the Telegram offset fetch was dropped in the Go port (ADR-0004 Am.3 — Option B); import links with --import-file instead"
	}
	return c.refusalCard(flag, detail)
}

// notYetCard is the honest in-development door: explicit, never a
// silent no-op. Exit 1 (the command did not run).
func (c *CLI) notYetCard(flag, what string) int {
	c.Banner()
	c.CliPrint(flag+" is not yet available in the Go port — "+what+".", "warning")
	c.CliPrint("Nothing was run.", "info")
	return 1
}

// refusalCard is the Option-B accepted-loss door (ADR-0004 Am.3 /
// rider R-0): the Telethon-only modes answer an honest refusal, never
// a silent no-op. Exit 1.
func (c *CLI) refusalCard(flag, why string) int {
	c.Banner()
	c.CliPrint(flag+" is unavailable in the Go port: "+why+".", "warning")
	c.CliPrint("Accepted loss per ADR-0004 Am.3 (Option B) — the direct Telegram leg is gone; the queue rides the Worker HTTP API.", "info")
	return 1
}

// helpCard mirrors cli_main's friendly help (the honest Option-B
// surface: the Telethon-only modes are dropped from the card — their
// flags answer refusal cards when used). Exit 0.
func (c *CLI) helpCard() int {
	c.Banner()
	fmt.Fprint(c.Out, Paint("Nothing to do — pick a mode:\n", cBold)+"\n")
	type row struct{ flag, text string }
	rows := []row{
		{"--init", "first-run wizard (saves credentials locally)"},
		{"--auto", "fully automatic run (SYNC → PROCESS → SEAL)"},
		{"--sync", "the fast lane: check → one Enter → sync (add --yes for cron)"},
		{"--status", "vault map + config + cache + note-state summary"},
		{"--test-connection", "vaults · LLM · GitHub · the sync queue — is everything up and ready?"},
		{"--detect-llm X", "Detect & Set a local engine (ollama | llamacpp) — probe, pick, save"},
		{"--list-dead", "list the 404 quarantine (dead links)"},
		{"--reset-dead", "clear the 404 quarantine"},
		{"--hand-delivery", "the fourth door — open walled links in your real Chrome"},
		{"--chrome-retry", "the fifth door — re-fetch failed links through your real Chrome"},
		{"--retry-failed", "reprocess the retry queue"},
		{"--import-file F", "curate URLs from a text file"},
	}
	for _, r := range rows {
		fmt.Fprintf(c.Out, "  %s %s\n", pyLjust(Paint(r.flag, cCyan), 24), r.text)
	}
	fmt.Fprint(c.Out, "\n")
	fmt.Fprint(c.Out, Paint("Windows tip: double-click GitCurator-SYNC.bat for the fast sync lane,", cDim)+"\n")
	fmt.Fprint(c.Out, Paint("or GitCurator-CLI.bat for the automatic run (both live in gitcurator/cli/).", cDim)+"\n")
	return 0
}

// helpText is the --help rendering: our honest surface in the
// reference's argparse shape (the Telethon-only flags dropped, the
// Worker queue named where the reference said Telegram).
func (c *CLI) helpText() string {
	var b []byte
	b = append(b, `usage: gitcurator-cli [-h] [--init] [--auto] [--sync] [--list-dead] [--reset-dead]
                      [--hand-delivery] [--chrome-retry] [--status]
                      [--test-connection] [--detect-llm {ollama,llamacpp}]
                      [--retry-failed] [--import-file F] [--vault PATH]
                      [--config PATH] [--strikes N] [--yes] [--dry-run]
                      [--no-color]

GitCurator visual CLI — colored, animated, fully automatic curation.

options:
  -h, --help            show this help message and exit
  --init                first-run wizard: ask for credentials and save config.json locally
  --auto                fully automatic run: bot SYNC → process → seal → publish
  --sync                v0.81.0 — THE SYNC DOOR, the fast lane for repeated use: check
                        config + connection + vaults → one Enter → sync the bot queue
                        and curate it into the vaults; --yes skips the button (cron);
                        failures tell you to fix essentials in the GUI
  --list-dead           list the 404 quarantine (confirmed-dead + attempts)
  --reset-dead          clear the 404 quarantine (every link gets fresh attempts)
  --hand-delivery       v0.48.0 — the fourth door: list every walled link
                        (403/bot-defense/TLS), queue it for hand-delivery and open it
                        in your real Chrome; the saved page becomes a real fetch on the
                        next run
  --chrome-retry        v0.50.0 — the fifth door: re-fetch every failed link
                        AUTOMATICALLY through your own real Chrome (a fresh session,
                        one tab per link, the live DOM taken as the content) and
                        process the delivered pages as real fetches in the same run
  --status              show config summary + cache/retry/404-quarantine stats
  --test-connection     check that everything is up and ready: vaults (found+writable) ·
                        the sync queue (the Worker's /health) · LLM (API/ollama/
                        llama.cpp) · GitHub (token + backup repos)
  --detect-llm {ollama,llamacpp}
                        v0.18.0 — Detect & Set a local LLM engine: probe it, pick the
                        model (menu when several), set provider + model + URL and
                        save. The fast lane the GUI's two Detect & Set buttons use
  --retry-failed        reprocess every URL in the retry queue
  --import-file F       batch-import from a .txt/.md file, one address per line (GitHub
                        repos AND websites; Markdown links, bullets and trailing notes
                        are fine)
  --vault PATH          override the vault path for this run
  --config PATH         config file path (default: app/config.json)
  --strikes N           override notfound_strike_threshold for this run (min 2)
  --yes                 skip the >10-repos confirmation in --auto
  --dry-run             v0.09.5 — safety net: log every vault write instead of
                        performing it (also skips sealing, publishing, and state
                        updates). Combine with any batch mode: --auto, --import-file,
                        --retry-failed
  --no-color            disable colored output

Tip: GitCurator-SYNC.bat is the fast sync lane; GitCurator-CLI.bat runs --auto with one
double-click (both in gitcurator/cli/).
`...)
	return string(b)
}

// cmdSync mirrors cli_sync.py cmd_sync's BEAT 1a (the pure half) —
// the live battery (1b), the button (2) and the fetch+curate engine
// (3) land with T-62; the config's own word answers honestly now.
func (c *CLI) cmdSync(a *Args) int {
	c.Banner()
	path := configPath(a.Config)
	cfg, ok := c.loadConfig(path)
	if !ok {
		c.SyncFailureCard([]string{fmt.Sprintf(
			"no config at %s — the GUI's Settings writes it on its first save", path)}, "")
		return 1
	}
	problems := SyncMissingEssentials(cfg, c.IsDir)
	if len(problems) > 0 {
		c.SyncFailureCard(problems, "")
		return 1
	}
	c.CliPrint("Everything present — but the sync engine itself (the live check battery, "+
		"the Enter button, the fetch + curate run) lands with T-62 (Phase 5); nothing was synced.", "warning")
	return 1
}
