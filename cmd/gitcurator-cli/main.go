// Command gitcurator-cli is the CLI entry point (console subsystem) of the
// Go port of GitCurator.
//
// Phase 0 implements only the surface the differential harness drives
// (status with a missing config file). The full flag surface, terminal
// paint and sync door land in Phase 5 (task T-60, spec-cli).
package main

import (
	"os"

	"github.com/assadigit/second-brain-project/internal/cli"
)

func main() {
	os.Exit(cli.Main(os.Args[1:]))
}
