// Package cli is the command-line surface of the Go port (spec:
// docs/port-blueprint/module-specs/spec-cli.md; full surface lands with
// task T-60 in Phase 5).
//
// Phase 0 seed: the welcome banner and the --status "no config" path,
// byte-exact with the Python app when colors are disabled (piped output,
// NO_COLOR, or --no-color). Verified by differential harness case
// testdiff/cases/status-no-config. The banner's stale "v0.09.5" version
// string is a preserved Python quirk (cli_terminal.py banner(); the house
// rule is quirks-preserving parity, see 09-testing-and-parity.md §5).
package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// bannerArt mirrors the art string in cli_terminal.py banner() byte for
// byte (raw string literal keeps the source shape 1:1).
const bannerArt = `
   ▄███████▄  ▄████████    ▄████████    ▄████████  ███▄▄▄▄      ▄████████
  ███    ███ ███    ███   ███    ███   ███    ███ ███▀▀▀██▄   ███    ███
  ███    ███ ███    ███   ███    ███   ███    █▀▀ ███    ███  ███    ███
 ▄███▄▄▄▄███▄███▄▄▄▄▄███▄ ███    ███  ▄███▄▄▄     ███    ███ ▄███▄▄▄▄███▄
▀▀▀▀▀▀▀▀▀▀▀ ███▀▀▀▀▀▀▀▀▀  ███    ███ ▀▀▀▀▀▀▀    ▀▀     ███▀ ▀▀▀▀▀   ███▀
▀███████████ ███    ███   ███    ███ ▀███████████ ▄██    ███ ▄██████████▄
  ███    ███ ███    ███   ███    ███          ███ ███    ███ ███    █████
  ███    ███ ███    ███ ████    ████ ▄███████████  ▀█████▀  ███    █████
`

// bannerVersionLine mirrors the second print of banner(): the stale
// "v0.09.5" stamp is a preserved quirk, not a bug to fix during the port.
const bannerVersionLine = "  GitHub Project Curator — CLI v0.09.5  ·  colored · animated · fully automatic"

// Level icons from _LEVEL_STYLE (cli_terminal.py) — used without ANSI
// color, matching Python's color-disabled output.
const (
	iconInfo    = "•"
	iconWarning = "▲"
)

// Banner writes the CLI welcome banner to w.
func Banner(w io.Writer) {
	fmt.Fprint(w, bannerArt)
	fmt.Fprintln(w, bannerVersionLine)
	fmt.Fprintln(w)
}

// configPath mirrors cli_settings.py _config_path: an explicit --config
// argument is made absolute; the default (APP_DIR/config.json) is
// provisional until task T-10 ports spec-core-constants.
func configPath(cliArg string) string {
	if cliArg != "" {
		if abs, err := filepath.Abs(cliArg); err == nil {
			return abs
		}
		return cliArg
	}
	exe, err := os.Executable()
	if err != nil {
		return "config.json"
	}
	return filepath.Join(filepath.Dir(exe), "config.json")
}

// cmdStatus mirrors cli.py cmd_status for the missing-config branch: the
// banner, one timestamped warning line (icon "▲"), exit code 1. The full
// configuration card (existing config) is Phase 5 work (T-60).
func cmdStatus(configArg string) int {
	Banner(os.Stdout)
	path := configPath(configArg)
	if _, err := os.Stat(path); err != nil {
		fmt.Fprintf(os.Stdout, "%s %s No config at %s — run `--cli --init` first.\n",
			time.Now().Format("15:04:05"), iconWarning, path)
		return 1
	}
	fmt.Fprintf(os.Stdout, "%s %s Configuration card — not yet ported (Phase 5, T-60).\n",
		time.Now().Format("15:04:05"), iconInfo)
	return 0
}

// Main is the CLI entry (called by cmd/gitcurator-cli). Phase 0 surface:
// --status and --config only.
func Main(argv []string) int {
	fs := flag.NewFlagSet("gitcurator-cli", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	status := fs.Bool("status", false, "show the configuration status card")
	config := fs.String("config", "", "config file path (default: APP_DIR/config.json)")
	if err := fs.Parse(argv); err != nil {
		return 2
	}
	if *status {
		return cmdStatus(*config)
	}
	fmt.Fprintln(os.Stderr, "gitcurator-cli: no mode flag given — only --status is ported so far (Phase 0 seed; full surface = Phase 5, T-60)")
	return 2
}
