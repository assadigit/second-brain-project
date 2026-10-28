// laws_test.go — T-47's own safety laws (the task's Go laws, pinned):
// the env scrubbing (GIT_DIR-family stripped, GIT_TERMINAL_PROMPT=0
// set), the token NEVER in .git/config (the token-less remote), the
// redaction of the token in ANY captured git output, and the CLI's
// flag surface + exit codes.
package vaultseal

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGitEnv_ScrubsTheHijackVarsAndNeverPrompts(t *testing.T) {
	tmp := t.TempDir()
	vault := initRepo(t, filepath.Join(tmp, "vault"), "note.md", "content\n")
	// (the repo is built BEFORE the hijack vars land — the test's own
	// git calls are the unscrubbed parent env)
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

	// the hijack is actually defeated: a repo inside the vault answers
	// rev-parse even with GIT_DIR pointing elsewhere in the parent env
	s := NewVaultSeal(vault, "tok", "mirror", true, true, 0, nil, false, nil)
	rc, out := s.gitf("rev-parse", "--show-toplevel")
	if rc != 0 || !strings.Contains(out, vault) {
		t.Fatalf("the hijack var leaked into the run: rc=%d out=%q", rc, out)
	}
}

func TestLaws_TokenNeverTouchesGitConfig(t *testing.T) {
	tmp := t.TempDir()
	vault := initRepo(t, filepath.Join(tmp, "vault"), "note.md", "content\n")
	s := NewVaultSeal(vault, "SUPERSECRETTOKEN", "mirror", true, true, 0, nil, true, nil)
	fakeAPI(s, map[[2]string][2]any{
		{"GET", "/user"}:                  {200, map[string]any{"login": "testuser"}},
		{"GET", "/repos/testuser/mirror"}: {404, map[string]any{"message": "Not Found"}},
		{"POST", "/user/repos"}:           {201, map[string]any{"full_name": "testuser/mirror"}},
	})
	ok, err := s.ensureRepo()
	if !ok {
		t.Fatalf("ensureRepo: %s", err)
	}
	for _, cfg := range []string{
		filepath.Join(vault, ".git", "config"),
	} {
		data, err := os.ReadFile(cfg)
		if err != nil {
			t.Fatalf("%s: %v", cfg, err)
		}
		if strings.Contains(string(data), "SUPERSECRETTOKEN") {
			t.Fatalf("THE TOKEN LAW BROKEN — the token reached %s:\n%s", cfg, data)
		}
		if !strings.Contains(string(data), "https://github.com/testuser/mirror.git") {
			t.Fatalf("the token-less remote is missing:\n%s", data)
		}
	}
}

func TestLaws_TheTokenIsRedactedInAnyCapturedOutput(t *testing.T) {
	tmp := t.TempDir()
	vault := initRepo(t, filepath.Join(tmp, "vault"), "note.md", "content\n")
	s := NewVaultSeal(vault, "SUPERSECRETTOKEN", "mirror", true, true, 0, nil, false, nil)
	// a command that ECHOES the token to the captured output: a remote
	// registered with the one-time token URL, then get-url — the
	// runner's redaction door must mask it (git itself may mask URLs
	// in some messages; this one is deterministic).
	url := "https://x-access-token:SUPERSECRETTOKEN@github.com/o/r.git"
	if rc, out := s.gitf("remote", "add", "tmpmirror", url); rc != 0 {
		t.Fatalf("remote add: %s", out)
	}
	rc, out := s.gitf("remote", "get-url", "tmpmirror")
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

func TestLaws_CLIFlagSurfaceAndExitCodes(t *testing.T) {
	tmp := t.TempDir()
	vault := initRepo(t, filepath.Join(tmp, "vault"), "note.md", "content\n")

	// --status prints the snapshot and exits 0
	var out bytes.Buffer
	var errOut bytes.Buffer
	if rc := Main([]string{"--vault", vault, "--status"}, &out, &errOut); rc != 0 {
		t.Fatalf("--status exit code: %d (stderr: %s)", rc, errOut.String())
	}
	if !strings.Contains(out.String(), `"is_git_repo": true`) {
		t.Fatalf("--status output: %s", out.String())
	}

	// a local seal with --no-push and --json: the commit lands, exit 0
	writeFile(t, filepath.Join(vault, "cli.md"), "cli note\n")
	out.Reset()
	if rc := Main([]string{"--vault", vault, "--no-push", "--json"}, &out, &errOut); rc != 0 {
		t.Fatalf("--no-push --json exit code: %d (stderr: %s)", rc, errOut.String())
	}
	if !strings.Contains(out.String(), `"sealed": true`) {
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

	// a failing seal exits 1: an unconfigured repo name with push on
	out.Reset()
	if rc := Main([]string{"--vault", vault, "--token", "tok"}, &out, &errOut); rc != 1 {
		t.Fatalf("unconfigured seal exit code: %d", rc)
	}
	if !strings.Contains(out.String(), "no backup repo configured") {
		t.Fatalf("describe: %s", out.String())
	}
}

// TestStageAll_AutoRecoversStaleIndexLock pins THE READY-MADE LAW
// (SESSION-081 / go.15): a stale .git/index.lock (the common Windows
// class — a previously-crashed git process leaves it) is auto-cleared
// + the stage retried; the owner clicks SYNC + the drift-seal
// recovers instead of failing with the confusing "Changes not staged
// for commit" message. Also pins: a clean stage succeeds without
// recovery; the warning log rides the recovery; the real git-add
// error surfaces when the stage fails for a non-lock cause.
func TestStageAll_AutoRecoversStaleIndexLock(t *testing.T) {
	tmp := t.TempDir()
	vault := initRepo(t, filepath.Join(tmp, "vault"), "note.md", "initial\n")
	// Dirty the working tree (a modified tracked file to stage).
	writeFile(t, filepath.Join(vault, "note.md"), "changed\n")

	var logs []string
	logFn := func(msg, level string) { logs = append(logs, msg+"|"+level) }
	s := NewVaultSeal(vault, "", "r", false, true, 0, logFn, false, nil)

	// 1. No lock + dirty → stage succeeds, no recovery, no log.
	if err := s.stageAll(); err != nil {
		t.Fatalf("clean stage: %v", err)
	}
	if len(logs) != 0 {
		t.Fatalf("clean stage: unexpected recovery log: %v", logs)
	}
	// The file is staged now (git status --porcelain shows "M  note.md").
	_, porcelain := s.gitf("status", "--porcelain")
	if !strings.HasPrefix(porcelain, "M  note.md") {
		t.Fatalf("clean stage: porcelain = %q; want the file staged (M  note.md)", porcelain)
	}
	// Commit so the tree is clean for the next scenario.
	s.gitf("commit", "-m", "stage1", "--quiet")

	// 2. Dirty again + a stale .git/index.lock → auto-recover.
	writeFile(t, filepath.Join(vault, "note.md"), "changed-again\n")
	lockPath := filepath.Join(vault, ".git", "index.lock")
	writeFile(t, lockPath, "")
	if err := s.stageAll(); err != nil {
		t.Fatalf("stale-lock stage: %v (expected auto-recovery)", err)
	}
	// The lock was deleted.
	if _, statErr := os.Stat(lockPath); !os.IsNotExist(statErr) {
		t.Fatalf("stale-lock stage: the .git/index.lock was not deleted")
	}
	// The recovery warning was logged.
	found := false
	for _, ln := range logs {
		if strings.Contains(ln, "cleared a stale .git/index.lock") && strings.Contains(ln, "warning") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("stale-lock stage: no recovery warning logged; logs=%v", logs)
	}
	// The file is staged (the retry succeeded).
	_, porcelain = s.gitf("status", "--porcelain")
	if !strings.HasPrefix(porcelain, "M  note.md") {
		t.Fatalf("stale-lock stage: porcelain = %q; want the file staged after recovery", porcelain)
	}
}
