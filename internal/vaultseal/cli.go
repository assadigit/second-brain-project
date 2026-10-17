// cli.go — the standalone CLI (vaultseal.py's _main): manual seals,
// the status snapshot, and the resync door. The exit code is 0 iff
// result.ok (1 otherwise); argparse's parse-error exit 2 stays the
// flag-package's domain (the reference's surface, not byte-pinned by
// any suite).
package vaultseal

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/assadigit/second-brain-project/internal/storage"
)

// Main mirrors _main(argv): the standalone CLI's surface — --vault
// (required), --repo-name, --create-repo, --token[=$GITHUB_TOKEN],
// --no-push, --resync, --json, --status.
func Main(argv []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("vaultseal", flag.ContinueOnError)
	fs.SetOutput(stderr)
	vault := fs.String("vault", "", "path to the Obsidian vault")
	repoName := fs.String("repo-name", "",
		"backup repo name (REQUIRED for pushing — no default is derived anymore, v0.83.0)")
	createRepo := fs.Bool("create-repo", false,
		"v0.83.0 — the one-time explicit opt-in that allows creating the private repo "+
			"when it does not exist (off by default: repos are never created silently)")
	tokenDefault := os.Getenv("GITHUB_TOKEN")
	token := fs.String("token", tokenDefault, "GitHub token (default: env GITHUB_TOKEN)")
	noPush := fs.Bool("no-push", false, "commit locally only, never push")
	resync := fs.Bool("resync", false,
		"v0.36.0 — make the mirror follow the vault EXACTLY (deletions included; "+
			"history kept, never force-pushed)")
	asJSON := fs.Bool("json", false, "print the result as JSON")
	status := fs.Bool("status", false, "print the vault git status and exit")
	if err := fs.Parse(argv); err != nil {
		return 2
	}
	if *vault == "" {
		fmt.Fprintln(stderr, "vaultseal: the --vault flag is required")
		fs.Usage()
		return 2
	}

	log := func(msg, level string) {
		fmt.Fprintf(stderr, "[vaultseal] [%s] %s\n", level, msg)
	}

	sealer := NewVaultSeal(*vault, *token, *repoName, !*noPush, true,
		DefaultTimeout, log, *createRepo, nil)
	if *status {
		fmt.Fprintln(stdout, storage.PyDumps(sealer.Status()))
		return 0
	}

	var result *SealResult
	if *resync {
		result = sealer.Resync("")
	} else {
		result = sealer.Seal(nil)
	}
	if *asJSON {
		fmt.Fprintln(stdout, storage.PyDumps(result.ToMap()))
	} else {
		fmt.Fprintln(stdout, result.Describe())
	}
	if result.OK() {
		return 0
	}
	return 1
}
