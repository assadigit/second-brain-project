// Command gitcurator-cli is the CLI entry point (console subsystem) of the
// Go port of GitCurator.
//
// Phase 5 (T-60): the CLI shell — the paint box, the honest Option-B
// flag surface with the dispatch order, the full --status card,
// --list-dead/--reset-dead, and THE SYNC DOOR's pure check half. The
// wizard, the test/detect/hand/chrome commands and the batch engines
// land with the rest of T-60 / T-62 (spec-cli).
package main

import (
	"os"

	"github.com/assadigit/second-brain-project/internal/cli"
)

func main() {
	os.Exit(cli.Main(os.Args[1:]))
}
