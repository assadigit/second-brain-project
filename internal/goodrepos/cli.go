// cli.go — the standalone CLI (goodrepos.py's _main): manual
// publishes, the status snapshot, and the dry-run door. The exit code
// is 0 iff result.ok (1 otherwise); the flag-package's parse-error
// exit 2 stays the reference's argparse surface (not byte-pinned by
// any suite). The process-exit staging cleanup rides the defer (the
// atexit convergence).
package goodrepos

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/assadigit/second-brain-project/internal/storage"
)

// Main mirrors _main(argv): the standalone CLI's surface — --vault
// (required), --repo-name, --create-repo, --token[=$GITHUB_TOKEN],
// --no-push, --dry-run, --json, --status.
func Main(argv []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("goodrepos", flag.ContinueOnError)
	fs.SetOutput(stderr)
	vault := fs.String("vault", "", "path to the Obsidian vault")
	repoName := fs.String("repo-name", "",
		"directory repo name (REQUIRED — no default is assumed anymore, v0.83.0)")
	createRepo := fs.Bool("create-repo", false,
		"v0.83.0 — the one-time explicit opt-in that allows creating the repo when it "+
			"does not exist (off by default; created PRIVATE)")
	tokenDefault := os.Getenv("GITHUB_TOKEN")
	token := fs.String("token", tokenDefault, "GitHub token (default: env GITHUB_TOKEN)")
	noPush := fs.Bool("no-push", false, "commit the directory in the staging dir, never push")
	dryRun := fs.Bool("dry-run", false,
		"build the directory + README into a temp dir and print a summary WITHOUT "+
			"running any git command; the temp dir is kept alive (path reported as "+
			"staging_dir in --json output) so you can inspect it")
	asJSON := fs.Bool("json", false, "print the result as JSON")
	status := fs.Bool("status", false, "print the vault scan status and exit")
	if err := fs.Parse(argv); err != nil {
		return 2
	}
	if *vault == "" {
		fmt.Fprintln(stderr, "goodrepos: the --vault flag is required")
		fs.Usage()
		return 2
	}

	defer CleanupStaging() // the atexit.register convergence

	log := func(msg, level string) {
		fmt.Fprintf(stderr, "[goodrepos] [%s] %s\n", level, msg)
	}

	publisher := NewGoodRepos(*vault, *token, *repoName, !*noPush, true,
		DefaultTimeout, log, *createRepo)
	if *status {
		fmt.Fprintln(stdout, storage.PyDumps(publisher.Status()))
		return 0
	}

	result := publisher.Publish(nil, *dryRun)
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
