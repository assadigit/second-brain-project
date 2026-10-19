// Command gitcurator-cli is the CLI entry point (console subsystem) of
// the Go port of GitCurator.
//
// T-63 (spec-cli's Shims+launchers section + doc 10 §2-6): the main
// package is the twin of the reference's main.py CLI routing — main.py
// strips --cli from argv and calls cli_main; the dedicated exe IS the
// CLI, so the strip law lives here (the launchers' run lines keep the
// reference's own "--cli --auto --yes" defaults verbatim, and muscle
// memory like "GitCurator-cli.exe --cli --status" keeps working). The
// exit code passes straight through from cli.Main to the process.
//
// The version stamp (doc 10 §2/§5): read the repo's VERSION file at
// build time — the release build command is
//
//	GOFLAGS=-trimpath CGO_ENABLED=0 go build \
//	  -ldflags "-X main.version=$(cat VERSION)" \
//	  -o bin/GitCurator-cli.exe ./cmd/gitcurator-cli
//
// (the zip packaging + the tag are T-65). "dev" is the unstamped
// fallback for go run / go test.
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/assadigit/second-brain-project/internal/cli"
)

// version is stamped at build time via -ldflags "-X main.version=…".
var version = "dev"

func main() {
	os.Exit(run(os.Stdout, os.Args[1:]))
}

// run is main's testable body. Order of the router doors (each mirrors
// a law of the reference):
//
//  1. the --cli strip (main.py: "[a for a in sys.argv[1:] if a != '--cli']"
//     — the flag selects the CLI in the Python world; here it is a no-op
//     spelling, stripped before the CLI's parser ever sees it);
//  2. the --version stamp door (doc 10 §5 — the Go packaging addition;
//     the reference's argparse has no --version: it answers the unknown-
//     flag error instead. The divergence is documented in spec-cli's
//     SESSION-031 shipped record);
//  3. cli.Main — the argparse surface + the dispatch (the exit code
//     passes through unchanged).
func run(out io.Writer, argv []string) int {
	rest := make([]string, 0, len(argv))
	for _, a := range argv {
		if a != "--cli" {
			rest = append(rest, a)
		}
	}
	for _, a := range rest {
		if a == "--version" || a == "-version" {
			fmt.Fprintln(out, "gitcurator-cli", version)
			return 0
		}
	}
	return cli.Main(rest)
}
