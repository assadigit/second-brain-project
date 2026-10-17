// laws_test.go — T-48's own safety laws (the task's Go laws, pinned):
// the env scrubbing (GIT_DIR-family stripped, GIT_TERMINAL_PROMPT=0
// set — the goodrepos origin of the law), the token NEVER in
// .git/config, the redaction of the token in ANY captured git output,
// the FETCH_HEAD scrub door, and the CLI's flag surface + exit codes.
package goodrepos

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGitEnv_ScrubsTheHijackVarsAndNeverPrompts(t *testing.T) {
	tmp := t.TempDir()
	vault := filepath.Join(tmp, "vault")
	if err := os.MkdirAll(vault, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GIT_DIR", "/tmp/hijack")
	t.Setenv("GIT_WORK_TREE", "/tmp/elsewhere")
	t.Setenv("GIT_INDEX_FILE", "/tmp/fake-index")
	t.Setenv("GIT_CONFIG", "/tmp/fake-config")
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("GIT_AUTHOR_NAME", "carried-through") // a legitimate var rides

	env := gitEnv()
	seen := map[string]string{}
	for _, kv := range env {
		if i := strings.IndexByte(kv, '='); i > 0 {
			seen[kv[:i]] = kv[i+1:]
		}
	}
	for _, banned := range []string{
		"GIT_DIR", "GIT_WORK_TREE", "GIT_INDEX_FILE", "GIT_CONFIG", "GIT_CONFIG_NOSYSTEM",
	} {
		if _, ok := seen[banned]; ok {
			t.Fatalf("%s survived the scrub", banned)
		}
	}
	if seen["GIT_TERMINAL_PROMPT"] != "0" {
		t.Fatal("GIT_TERMINAL_PROMPT=0 missing")
	}
	if seen["GIT_AUTHOR_NAME"] != "carried-through" {
		t.Fatal("a legitimate var was dropped")
	}
}

func TestLaws_TheTokenIsRedactedInAnyCapturedOutput(t *testing.T) {
	b := newGRBase(t)
	b.defaultVault()
	pub := NewGoodRepos(b.vault, "SUPERSECRETTOKEN", "mirror", false, true, 0, nil, false)
	t.Cleanup(func() { dropStaging(pub) })
	// a staging dir must exist AND be a git repo for the runner's git
	// calls to run at all
	pub.staging = pub.newStaging(true)
	t.Cleanup(func() { _ = os.RemoveAll(pub.staging); unregisterStaging(pub.staging) })
	if rc, out := pub.gitf("init", "-b", "main"); rc != 0 {
		t.Fatalf("init: %s", out)
	}
	// a command that ECHOES the token to the captured output: a remote
	// registered with the one-time token URL, then get-url — the
	// runner's redaction door must mask it.
	url := "https://x-access-token:SUPERSECRETTOKEN@github.com/o/r.git"
	if rc, out := pub.gitf("remote", "add", "tmpmirror", url); rc != 0 {
		t.Fatalf("remote add: %s", out)
	}
	rc, out := pub.gitf("remote", "get-url", "tmpmirror")
	if rc != 0 {
		t.Fatalf("remote get-url: %s", out)
	}
	if strings.Contains(out, "SUPERSECRETTOKEN") {
		t.Fatalf("THE REDACTION LAW BROKEN — the token survived in the output: %q", out)
	}
	if !strings.Contains(out, "***") {
		t.Fatalf("the mask is missing: %q", out)
	}
}

func TestLaws_NoStagingIsA127Door(t *testing.T) {
	b := newGRBase(t)
	pub := NewGoodRepos(b.vault, "", "mirror", false, true, 0, nil, false)
	rc, out := pub.gitf("status", "--porcelain")
	if rc != 127 || !strings.Contains(out, "no staging directory configured") {
		t.Fatalf("rc=%d out=%q", rc, out)
	}
}

func TestLaws_RedactIsANoOpWithoutAToken(t *testing.T) {
	b := newGRBase(t)
	pub := NewGoodRepos(b.vault, "", "mirror", false, true, 0, nil, false)
	if got := pub.redact("xSUPERSECRETTOKENx"); got != "xSUPERSECRETTOKENx" {
		t.Fatalf("redact without a token: %q", got)
	}
	pub2 := NewGoodRepos(b.vault, "TOK", "mirror", false, true, 0, nil, false)
	if got := pub2.redact("nothing here"); got != "nothing here" {
		t.Fatalf("redact without an occurrence: %q", got)
	}
}

func TestLaws_SanitizeFetchHeadNoTokenIsANoOp(t *testing.T) {
	tmp := t.TempDir()
	staging := filepath.Join(tmp, "staging")
	if err := os.MkdirAll(filepath.Join(staging, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	fh := filepath.Join(staging, ".git", "FETCH_HEAD")
	if err := os.WriteFile(fh, []byte("https://x-access-token:SECRETTOKEN@github.com/o/r.git\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	b := newGRBase(t)
	pub := NewGoodRepos(b.vault, "", "r", false, true, 0, nil, false)
	pub.staging = staging
	pub.sanitizeFetchHead()
	data, _ := os.ReadFile(fh)
	if strings.Contains(string(data), "***") { // untouched without a token
		t.Fatalf("no-token sanitize rewrote FETCH_HEAD:\n%s", data)
	}
}

func TestLaws_CLIFlagSurfaceAndExitCodes(t *testing.T) {
	b := newGRBase(t)
	b.defaultVault()

	// --status prints the snapshot and exits 0
	var out bytes.Buffer
	var errOut bytes.Buffer
	if rc := Main([]string{"--vault", b.vault, "--status"}, &out, &errOut); rc != 0 {
		t.Fatalf("--status exit code: %d (stderr: %s)", rc, errOut.String())
	}
	if !strings.Contains(out.String(), `"entries": 3`) {
		t.Fatalf("--status output: %s", out.String())
	}

	// a dry-run builds the tree without git and reports the staging dir
	out.Reset()
	if rc := Main([]string{"--vault", b.vault, "--repo-name", "good-repos",
		"--dry-run", "--json"}, &out, &errOut); rc != 0 {
		t.Fatalf("--dry-run --json exit code: %d (stderr: %s)", rc, errOut.String())
	}
	if !strings.Contains(out.String(), `"dry_run": true`) {
		t.Fatalf("--json output: %s", out.String())
	}
	if !strings.Contains(out.String(), `"staging_dir": "`) {
		t.Fatalf("staging_dir missing: %s", out.String())
	}

	// a local publish with --no-push and --json: the commit lands, exit 0
	out.Reset()
	if rc := Main([]string{"--vault", b.vault, "--repo-name", "good-repos",
		"--no-push", "--json"}, &out, &errOut); rc != 0 {
		t.Fatalf("--no-push --json exit code: %d (stderr: %s)", rc, errOut.String())
	}
	if !strings.Contains(out.String(), `"published": true`) {
		t.Fatalf("--json output: %s", out.String())
	}
	if !strings.Contains(out.String(), `"repo_url": null`) {
		// no token → no repo resolved: the None doors stay null
		t.Fatalf("--json output: %s", out.String())
	}

	// a missing --vault is a flag error (exit 2)
	if rc := Main([]string{}, &out, &errOut); rc != 2 {
		t.Fatalf("missing --vault exit code: %d", rc)
	}

	// an unconfigured repo name is a SKIP (ok=True): the CLI exits 0
	// and describe speaks the v0.83 guard
	out.Reset()
	if rc := Main([]string{"--vault", b.vault}, &out, &errOut); rc != 0 {
		t.Fatalf("unconfigured publish exit code: %d", rc)
	}
	if !strings.Contains(out.String(), "no directory repo configured") {
		t.Fatalf("describe: %s", out.String())
	}
}
