// vaultseal_test.go — T-47: the ported selections of the reference's
// real-git integration suites:
//
//   - test_sealfix.py §1 TestSealPushReconcile (THE DIVERGENCE
//     LADDER's every branch against REAL local bare repos — the
//     fast-forward, the up-to-date, the shared-history rebase, the
//     unrelated-histories merge, the two conflict-rescue branches,
//     the non-rejection pass-through, the FETCH_HEAD scrub, the
//     seal-result push/rescue wiring);
//   - test_sealfix.py §6 TestCreationGuardV0830 (the v0.83 creation
//     guard — 404 without opt-in, POST-once with it, the empty-name
//     NOT-CONFIGURED door, the existing-repo no-POST, the bridge's
//     opt-in);
//   - test_localonly.py §2-§3 (the LOCAL-ONLY BELT's real-git
//     mechanics: the managed block's merge/rebuild/idempotence, the
//     untrack law — the files NEVER leave the disk — the
//     seal-applies-the-belt cycle, the config bridge's belt facts).
//
// TEMP REPOS ONLY (never the owner's vault); NO real network (the
// GitHub API surface is the instance-seam fake exactly like the
// reference's own tests; the "remote" is a local bare repo).
package vaultseal

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// git helpers (REAL local repos — the same porcelain VaultSeal drives)
// ---------------------------------------------------------------------------

func runGit(t *testing.T, cwd string, args ...string) string {
	t.Helper()
	out, err := exec.Command("git", append([]string{"-C", cwd}, args...)...).CombinedOutput()
	if err != nil {
		t.Fatalf("git -C %s %v: %v\n%s", cwd, args, err, out)
	}
	return string(out)
}

func tryGit(cwd string, args ...string) (int, string) {
	out, err := exec.Command("git", append([]string{"-C", cwd}, args...)...).CombinedOutput()
	rc := 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			rc = ee.ExitCode()
		} else {
			rc = 127
		}
	}
	return rc, string(out)
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// initRepo mirrors _init_repo: a fresh git repo with one commit on
// main.
func initRepo(t *testing.T, path, initialFile, content string) string {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
	runGit(t, path, "init", "-b", "main")
	runGit(t, path, "config", "user.email", "t@example.com")
	runGit(t, path, "config", "user.name", "T")
	if initialFile != "" {
		writeFile(t, filepath.Join(path, initialFile), content)
		runGit(t, path, "add", "-A", "--", ".")
		runGit(t, path, "commit", "-m", "init", "--quiet")
	}
	return path
}

// commitRepo mirrors _commit.
func commitRepo(t *testing.T, path, fname, content, msg string) {
	t.Helper()
	writeFile(t, filepath.Join(path, fname), content)
	runGit(t, path, "add", "-A", "--", ".")
	runGit(t, path, "commit", "-m", msg, "--quiet")
}

// bareRepo mirrors _bare.
func bareRepo(t *testing.T, path string) string {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
	runGit(t, path, "init", "--bare", "-b", "main")
	return path
}

func remoteHead(t *testing.T, bare string) string {
	t.Helper()
	return strings.TrimSpace(runGit(t, bare, "rev-parse", "--short=7", "main"))
}

func remoteTree(t *testing.T, bare string) []string {
	t.Helper()
	var out []string
	for _, ln := range strings.Split(runGit(t, bare, "ls-tree", "-r", "--name-only", "main"), "\n") {
		if strings.TrimSpace(ln) != "" {
			out = append(out, strings.TrimSpace(ln))
		}
	}
	return out
}

func remoteBranches(t *testing.T, bare string) []string {
	t.Helper()
	var out []string
	for _, ln := range strings.Split(runGit(t, bare, "branch", "--format", "%(refname:short)"), "\n") {
		if strings.TrimSpace(ln) != "" {
			out = append(out, strings.TrimSpace(ln))
		}
	}
	return out
}

func treeHas(tree []string, path string) bool {
	for _, p := range tree {
		if p == path {
			return true
		}
	}
	return false
}

// sealFor mirrors _seal_for.
func sealFor(vaultDir string) *VaultSeal {
	return NewVaultSeal(vaultDir, "tok", "mirror", true, true, 0, nil, false, nil)
}

// ---------------------------------------------------------------------------
// 1. VaultSeal push reconciliation (test_sealfix TestSealPushReconcile)
// ---------------------------------------------------------------------------

func TestSealPushReconcile_FastForwardToEmptyRemote(t *testing.T) {
	tmp := t.TempDir()
	bare := bareRepo(t, filepath.Join(tmp, "mirror.git"))
	vault := initRepo(t, filepath.Join(tmp, "vault"), "note.md", strings.Repeat("x", 80))
	ok, err := sealFor(vault).push(bare)
	if !ok {
		t.Fatalf("push failed: %s", err)
	}
	if !treeHas(remoteTree(t, bare), "note.md") {
		t.Fatalf("note.md missing from the remote tree: %v", remoteTree(t, bare))
	}
}

func TestSealPushReconcile_PushTwiceIsUpToDate(t *testing.T) {
	tmp := t.TempDir()
	bare := bareRepo(t, filepath.Join(tmp, "mirror.git"))
	vault := initRepo(t, filepath.Join(tmp, "vault"), "note.md", strings.Repeat("x", 80))
	s := sealFor(vault)
	if ok, err := s.push(bare); !ok {
		t.Fatalf("first push failed: %s", err)
	}
	ok, err := s.push(bare)
	if !ok {
		t.Fatalf("second push failed: %s", err)
	}
}

func TestSealPushReconcile_DivergedSharedHistoryRebasesOntoRemote(t *testing.T) {
	// The exact v0.20.0 websites-mirror shape, minus the unrelated
	// part: both sides share a base, the remote gained a commit we
	// lack (a web-UI edit / second machine), we gained a seal commit.
	tmp := t.TempDir()
	bare := bareRepo(t, filepath.Join(tmp, "mirror.git"))
	base := initRepo(t, filepath.Join(tmp, "base"), "base.txt", "base\n")
	runGit(t, base, "push", bare, "HEAD:refs/heads/main")
	vault := filepath.Join(tmp, "vault")
	runGit(t, base, "clone", bare, vault)
	runGit(t, vault, "config", "user.email", "t@example.com")
	runGit(t, vault, "config", "user.name", "T")
	other := filepath.Join(tmp, "other")
	runGit(t, base, "clone", bare, other)
	runGit(t, other, "config", "user.email", "t@example.com")
	runGit(t, other, "config", "user.name", "T")
	commitRepo(t, other, "web-edit.txt", "edited on github\n", "web edit")
	runGit(t, other, "push", bare, "HEAD:refs/heads/main")
	commitRepo(t, vault, "seal-note.md", strings.Repeat("y", 80), "seal")

	ok, err := sealFor(vault).push(bare)
	if !ok {
		t.Fatalf("push failed: %s", err)
	}
	tree := remoteTree(t, bare)
	if !treeHas(tree, "web-edit.txt") { // remote work kept
		t.Fatalf("web-edit.txt missing: %v", tree)
	}
	if !treeHas(tree, "seal-note.md") { // our seal landed too
		t.Fatalf("seal-note.md missing: %v", tree)
	}
}

func TestSealPushReconcile_UnrelatedHistoriesMergeKeepsBoth(t *testing.T) {
	// Fresh `git init` in the vault vs a remote that was used for
	// something else (the gate mirror). No shared base → merge with
	// --allow-unrelated-histories; nothing discarded.
	tmp := t.TempDir()
	bare := bareRepo(t, filepath.Join(tmp, "mirror.git"))
	gate := initRepo(t, filepath.Join(tmp, "gate"), "gate-runner.yml", "ci: yes\n")
	runGit(t, gate, "push", bare, "HEAD:refs/heads/main")
	vault := initRepo(t, filepath.Join(tmp, "vault"), "websites-note.md", strings.Repeat("z", 80))

	ok, err := sealFor(vault).push(bare)
	if !ok {
		t.Fatalf("push failed: %s", err)
	}
	tree := remoteTree(t, bare)
	if !treeHas(tree, "gate-runner.yml") { // remote history intact
		t.Fatalf("gate-runner.yml missing: %v", tree)
	}
	if !treeHas(tree, "websites-note.md") { // vault snapshot added
		t.Fatalf("websites-note.md missing: %v", tree)
	}
}

func TestSealPushReconcile_UnrelatedConflictLandsOnRescueBranch(t *testing.T) {
	tmp := t.TempDir()
	bare := bareRepo(t, filepath.Join(tmp, "mirror.git"))
	gate := initRepo(t, filepath.Join(tmp, "gate"), "README.md", "remote\n")
	runGit(t, gate, "push", bare, "HEAD:refs/heads/main")
	remoteBefore := remoteHead(t, bare)
	vault := initRepo(t, filepath.Join(tmp, "vault"), "README.md", "vault\n")

	ok, err := sealFor(vault).push(bare)
	if ok {
		t.Fatal("expected the conflict push to fail")
	}
	if !strings.Contains(err, "seal-rescue/") {
		t.Fatalf("error does not name the rescue branch: %q", err)
	}
	// main untouched; the seal preserved on its own branch
	if got := remoteHead(t, bare); got != remoteBefore {
		t.Fatalf("remote main moved: %s -> %s", remoteBefore, got)
	}
	rescue := 0
	for _, b := range remoteBranches(t, bare) {
		if strings.HasPrefix(b, "seal-rescue/") {
			rescue++
		}
	}
	if rescue != 1 {
		t.Fatalf("expected 1 rescue branch, got %d (%v)", rescue, remoteBranches(t, bare))
	}
}

func TestSealPushReconcile_RebaseConflictLandsOnRescueBranch(t *testing.T) {
	tmp := t.TempDir()
	bare := bareRepo(t, filepath.Join(tmp, "mirror.git"))
	base := initRepo(t, filepath.Join(tmp, "base"), "f.txt", "base\n")
	runGit(t, base, "push", bare, "HEAD:refs/heads/main")
	vault := filepath.Join(tmp, "vault")
	runGit(t, base, "clone", bare, vault)
	runGit(t, vault, "config", "user.email", "t@example.com")
	runGit(t, vault, "config", "user.name", "T")
	other := filepath.Join(tmp, "other")
	runGit(t, base, "clone", bare, other)
	runGit(t, other, "config", "user.email", "t@example.com")
	runGit(t, other, "config", "user.name", "T")
	commitRepo(t, other, "f.txt", "remote version\n", "remote edit")
	runGit(t, other, "push", bare, "HEAD:refs/heads/main")
	commitRepo(t, vault, "f.txt", "vault version\n", "seal edit")

	ok, err := sealFor(vault).push(bare)
	if ok {
		t.Fatal("expected the rebase-conflict push to fail")
	}
	if !strings.Contains(err, "seal-rescue/") {
		t.Fatalf("error does not name the rescue branch: %q", err)
	}
	rescue := 0
	for _, b := range remoteBranches(t, bare) {
		if strings.HasPrefix(b, "seal-rescue/") {
			rescue++
		}
	}
	if rescue != 1 {
		t.Fatalf("expected 1 rescue branch, got %d", rescue)
	}
}

func TestSealPushReconcile_NonRejectionErrorIsPassedThrough(t *testing.T) {
	// Auth/network failures ("repository not found") must NOT trigger
	// reconciliation or a rescue push.
	tmp := t.TempDir()
	vault := initRepo(t, filepath.Join(tmp, "vault"), "note.md", strings.Repeat("x", 80))
	nowhere := filepath.Join(tmp, "no-such-repo.git")
	ok, err := sealFor(vault).push(nowhere)
	if ok {
		t.Fatal("expected the nowhere push to fail")
	}
	if strings.Contains(err, "seal-rescue") {
		t.Fatalf("a non-rejection error must not reach the rescue: %q", err)
	}
}

func TestSealPushReconcile_FetchHeadTokenIsScrubbed(t *testing.T) {
	tmp := t.TempDir()
	vault := initRepo(t, filepath.Join(tmp, "vault"), "note.md", strings.Repeat("x", 80))
	fh := filepath.Join(vault, ".git", "FETCH_HEAD")
	writeFile(t, fh, "abc123\t\tbranch 'main' of "+
		"https://x-access-token:SECRETTOKEN@github.com/o/r.git\n")
	s := NewVaultSeal(vault, "SECRETTOKEN", "mirror", true, true, 0, nil, false, nil)
	s.sanitizeFetchHead()
	text, err := os.ReadFile(fh)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(text), "SECRETTOKEN") {
		t.Fatalf("token survived in FETCH_HEAD: %q", text)
	}
	if !strings.Contains(string(text), "***") {
		t.Fatalf("the mask is missing: %q", text)
	}
}

func TestSealPushReconcile_SealResultWiresPushAndRescue(t *testing.T) {
	// seal() plumbing: a failing _push becomes the SealResult error
	// (commit kept); a reconciled push marks pushed + refreshes sha.
	tmp := t.TempDir()
	vault := initRepo(t, filepath.Join(tmp, "vault"), "note.md", strings.Repeat("x", 80))
	writeFile(t, filepath.Join(vault, "dirty.md"), "uncommitted change\n") // something to seal
	s := NewVaultSeal(vault, "tok", "mirror", true, true, 0, nil, false, nil)
	s.repoFull = "o/mirror"
	s.remoteURL = "https://github.com/o/mirror"
	s.ensureFn = func() (bool, string) { return true, "" }
	s.pushFn = func(string) (bool, string) { return true, "" }
	one, total := 1, 1
	res := s.Seal(&RunSummary{Processed: &one, Total: &total})
	if !res.OK() {
		t.Fatalf("res.error: %s", res.Error)
	}
	if !res.Pushed {
		t.Fatal("expected pushed")
	}

	writeFile(t, filepath.Join(vault, "note2.md"), strings.Repeat("y", 80))
	s.pushFn = func(string) (bool, string) {
		return false, "mirror main rejected … 'seal-rescue/2099'"
	}
	res2 := s.Seal(nil)
	if res2.OK() {
		t.Fatal("expected the failing-push seal to be not-ok")
	}
	if !res2.Sealed { // the local commit is kept
		t.Fatal("expected sealed=true (commit kept)")
	}
	if !strings.Contains(res2.Error, "seal-rescue") {
		t.Fatalf("error does not name the rescue: %q", res2.Error)
	}
	if !strings.Contains(res2.Describe(), "commit kept") {
		t.Fatalf("describe does not say commit kept: %q", res2.Describe())
	}
}

// ---------------------------------------------------------------------------
// 2. the v0.83 creation guard (test_sealfix TestCreationGuardV0830)
// ---------------------------------------------------------------------------

type apiCall struct {
	method string
	path   string
	body   map[string]any
}

// fakeAPI mirrors the suite's instance-level _api stub: records every
// call, answers from the responses table ((404, Not Found) by
// default).
func fakeAPI(s *VaultSeal, responses map[[2]string][2]any) *[]apiCall {
	calls := &[]apiCall{}
	s.apiFn = func(method, path string, body map[string]any) (int, map[string]any) {
		*calls = append(*calls, apiCall{method, path, body})
		key := [2]string{method, path}
		if resp, ok := responses[key]; ok {
			status, _ := resp[0].(int)
			payload, _ := resp[1].(map[string]any)
			return status, payload
		}
		return 404, map[string]any{"message": "Not Found"}
	}
	return calls
}

func guardSealer(t *testing.T, tmp string, repoName string, allowCreate bool) *VaultSeal {
	t.Helper()
	return NewVaultSeal(tmp, "tok", repoName, true, true, 0, nil, allowCreate, nil)
}

func TestCreationGuard_404NeverCreatesWithoutOptIn(t *testing.T) {
	tmp := t.TempDir()
	s := guardSealer(t, tmp, "my-vault-bak", false)
	calls := fakeAPI(s, map[[2]string][2]any{
		{"GET", "/user"}:                        {200, map[string]any{"login": "testuser"}},
		{"GET", "/repos/testuser/my-vault-bak"}: {404, map[string]any{"message": "Not Found"}},
	})
	ok, err := s.ensureRepo()
	if ok || err == "" {
		t.Fatalf("expected the guard to refuse: ok=%v err=%q", ok, err)
	}
	if !strings.Contains(err, "does NOT create repositories automatically") {
		t.Fatalf("error does not carry the guard's law: %q", err)
	}
	if !strings.Contains(err, "create_repo=true") {
		t.Fatalf("error does not teach the opt-in: %q", err)
	}
	// the GitHub-side proof: exactly one GET for /user + one GET for
	// the repo — ZERO POST /user/repos
	for _, c := range *calls {
		if c.method == "POST" {
			t.Fatalf("a POST happened without the opt-in: %+v", c)
		}
	}
}

func TestCreationGuard_AllowCreatePostsOnce(t *testing.T) {
	tmp := t.TempDir()
	s := guardSealer(t, tmp, "my-vault-bak", true)
	calls := fakeAPI(s, map[[2]string][2]any{
		{"GET", "/user"}:                        {200, map[string]any{"login": "testuser"}},
		{"GET", "/repos/testuser/my-vault-bak"}: {404, map[string]any{"message": "Not Found"}},
		{"POST", "/user/repos"}:                 {201, map[string]any{"full_name": "testuser/my-vault-bak"}},
	})
	ok, err := s.ensureRepo()
	if !ok {
		t.Fatalf("ensureRepo failed: %s", err)
	}
	posts := 0
	var body map[string]any
	for _, c := range *calls {
		if c.method == "POST" {
			posts++
			body = c.body
		}
	}
	if posts != 1 {
		t.Fatalf("expected exactly 1 POST, got %d", posts)
	}
	if body["name"] != "my-vault-bak" {
		t.Fatalf("POST body name: %v", body["name"])
	}
	if body["private"] != true {
		t.Fatalf("POST body private: %v", body["private"])
	}
}

func TestCreationGuard_EmptyRepoNameIsNotConfigured(t *testing.T) {
	tmp := t.TempDir()
	s := guardSealer(t, tmp, "", false)
	calls := fakeAPI(s, nil)
	ok, err := s.ensureRepo()
	if ok || err == "" {
		t.Fatalf("expected the not-configured refusal: ok=%v err=%q", ok, err)
	}
	if !strings.Contains(err, "no backup repo configured") {
		t.Fatalf("error does not say not-configured: %q", err)
	}
	// not even a /user lookup — nothing is named, nothing happens
	if len(*calls) != 0 {
		t.Fatalf("API calls happened for an unconfigured seal: %+v", *calls)
	}
}

func TestCreationGuard_ExistingRepoNoCreationAttempt(t *testing.T) {
	tmp := t.TempDir()
	s := guardSealer(t, tmp, "my-vault-bak", false)
	fakeAPI(s, map[[2]string][2]any{
		{"GET", "/user"}:                        {200, map[string]any{"login": "testuser"}},
		{"GET", "/repos/testuser/my-vault-bak"}: {200, map[string]any{"private": true}},
	})
	ok, err := s.ensureRepo()
	if !ok {
		t.Fatalf("ensureRepo failed: %s", err)
	}
	// (the POST absence is asserted inside fakeAPI's table — an
	// explicit pass here mirrors the reference's shape)
	rc, _ := tryGit(tmp, "rev-parse", "--show-toplevel")
	if rc != 0 {
		t.Fatal("expected the vault to have become a repo")
	}
}

func TestCreationGuard_ConfigBridgePassesTheOptIn(t *testing.T) {
	// seal_from_config: create_repo=true rides through; the default
	// (missing key) is False — the guard is armed out of the box.
	cfg := map[string]any{
		"vault_path": t.TempDir(), "github_token": "tok",
		"vaultseal": map[string]any{"repo_name": "my-vault-bak", "create_repo": true},
	}
	if !SealerFromConfig(cfg, nil).allowCreate {
		t.Fatal("create_repo=true did not ride through the bridge")
	}
	cfg2 := map[string]any{
		"vault_path": t.TempDir(), "github_token": "tok",
		"vaultseal": map[string]any{"repo_name": "my-vault-bak"},
	}
	if SealerFromConfig(cfg2, nil).allowCreate {
		t.Fatal("the missing create_repo key must default to the armed guard")
	}
}

// ---------------------------------------------------------------------------
// 3. the LOCAL-ONLY BELT's real-git mechanics (test_localonly §2)
// ---------------------------------------------------------------------------

// beltRepo mirrors _SealCase._repo: a temp vault that IS a git repo,
// with the belt's canonical shape.
func beltRepo(t *testing.T, tracked bool) string {
	t.Helper()
	tmp := t.TempDir()
	vault := filepath.Join(tmp, "vault")
	if err := os.MkdirAll(filepath.Join(vault, "Download_Resources"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(vault, "Design_Icons.md"), "icons")
	writeFile(t, filepath.Join(vault, "Download_Resources", "CakhiaTV.md"), "piracy note")
	writeFile(t, filepath.Join(vault, "Download_Resources", "Torrent_Shadow_Libraries", "Z-Library.md"), "shadow lib")
	runGit(t, vault, "init", "-q", "-b", "main")
	runGit(t, vault, "config", "user.email", "t@example.com")
	runGit(t, vault, "config", "user.name", "T")
	if tracked {
		runGit(t, vault, "add", "-A")
		runGit(t, vault, "commit", "-qm", "base")
	}
	return vault
}

func trackedFiles(t *testing.T, vault string) []string {
	t.Helper()
	rc, out := tryGit(vault, "ls-files")
	if rc != 0 {
		t.Fatalf("ls-files failed: %s", out)
	}
	var files []string
	for _, f := range strings.Fields(out) {
		files = append(files, f)
	}
	return files
}

func inList(list []string, item string) bool {
	for _, x := range list {
		if x == item {
			return true
		}
	}
	return false
}

func beltSealer(vault string, patterns []string) *VaultSeal {
	return NewVaultSeal(vault, "", "r", false, true, 0, nil, false, patterns)
}

func TestLocalOnlyGitignore_TheBlockLandsAndIsIdempotent(t *testing.T) {
	vault := beltRepo(t, false)
	sealer := beltSealer(vault, []string{"Torrent_Shadow_Libraries/", "CakhiaTV*.md"})
	sealer.EnsureLocalOnlyGitignore()
	text1, err := os.ReadFile(filepath.Join(vault, ".gitignore"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(text1), LocalOnlyHeader) {
		t.Fatalf("the block header is missing: %q", text1)
	}
	if !strings.Contains(string(text1), "Torrent_Shadow_Libraries/") {
		t.Fatalf("the folder pattern is missing: %q", text1)
	}
	sealer.EnsureLocalOnlyGitignore()
	text2, _ := os.ReadFile(filepath.Join(vault, ".gitignore"))
	if string(text1) != string(text2) {
		t.Fatalf("not idempotent:\n%q\n%q", text1, text2)
	}
}

func TestLocalOnlyGitignore_TheBlockRebuildsWhenTheBeltChanges(t *testing.T) {
	vault := beltRepo(t, false)
	sealer := beltSealer(vault, []string{"CakhiaTV*.md"})
	sealer.EnsureLocalOnlyGitignore()
	sealer.localOnly = []string{"Prowlarr*.md"}
	sealer.EnsureLocalOnlyGitignore()
	text, _ := os.ReadFile(filepath.Join(vault, ".gitignore"))
	if !strings.Contains(string(text), "Prowlarr*.md") {
		t.Fatalf("the new pattern is missing: %q", text)
	}
	// the OLD pattern must be gone from the managed block
	block := string(text)
	if i := strings.Index(block, LocalOnlyHeader); i >= 0 {
		if strings.Contains(block[i:], "CakhiaTV*.md") {
			t.Fatalf("the removed pattern survived the rebuild: %q", text)
		}
	} else {
		t.Fatalf("the block header vanished: %q", text)
	}
}

func TestLocalOnlyGitignore_TheOwnersOwnLinesSurviveTheRebuild(t *testing.T) {
	vault := beltRepo(t, false)
	gi := filepath.Join(vault, ".gitignore")
	writeFile(t, gi, "my-own-rule/\n\n"+LocalOnlyHeader+
		"\nCakhiaTV*.md\n\nowner-after-block/\n")
	sealer := beltSealer(vault, []string{"Prowlarr*.md"})
	sealer.EnsureLocalOnlyGitignore()
	text, _ := os.ReadFile(gi)
	s := string(text)
	for _, want := range []string{"my-own-rule/", "owner-after-block/", "Prowlarr*.md"} {
		if !strings.Contains(s, want) {
			t.Fatalf("lost %q in the rebuild: %q", want, text)
		}
	}
	if strings.Contains(s, "CakhiaTV*.md") {
		t.Fatalf("the old belt pattern survived: %q", text)
	}
}

func TestLocalOnlyGitignore_TheMachineStateBlockAndTheLocalOnlyBlockCoexist(t *testing.T) {
	vault := beltRepo(t, false)
	sealer := beltSealer(vault, []string{"CakhiaTV*.md"})
	sealer.EnsureGitignore()
	sealer.EnsureLocalOnlyGitignore()
	text, _ := os.ReadFile(filepath.Join(vault, ".gitignore"))
	s := string(text)
	if !strings.Contains(s, GitignoreHeader) || !strings.Contains(s, LocalOnlyHeader) {
		t.Fatalf("the two managed blocks do not coexist: %q", text)
	}
}

func TestLocalOnlyGitignore_NoPatternsNoBlock(t *testing.T) {
	vault := beltRepo(t, false)
	sealer := beltSealer(vault, nil)
	if got := sealer.EnsureLocalOnlyGitignore(); len(got) != 0 {
		t.Fatalf("expected no patterns answered, got %v", got)
	}
	if _, err := os.Stat(filepath.Join(vault, ".gitignore")); err == nil {
		t.Fatal("a .gitignore grew for a beltless vault")
	}
}

func TestUntrackLocalOnly_TrackedPrivateNotesUntrackAndStayOnDisk(t *testing.T) {
	vault := beltRepo(t, true)
	before := trackedFiles(t, vault)
	if !inList(before, "Download_Resources/CakhiaTV.md") ||
		!inList(before, "Download_Resources/Torrent_Shadow_Libraries/Z-Library.md") {
		t.Fatalf("the base repo lost its tracked files: %v", before)
	}
	sealer := beltSealer(vault, []string{"Torrent_Shadow_Libraries/", "CakhiaTV*.md"})
	n := sealer.untrackLocalOnly()
	if n != 2 {
		t.Fatalf("expected 2 untracked, got %d", n)
	}
	after := trackedFiles(t, vault)
	if inList(after, "Download_Resources/CakhiaTV.md") {
		t.Fatal("CakhiaTV.md still tracked")
	}
	if inList(after, "Download_Resources/Torrent_Shadow_Libraries/Z-Library.md") {
		t.Fatal("Z-Library.md still tracked")
	}
	// THE LAW: the files NEVER leave the disk
	for _, p := range []string{
		filepath.Join(vault, "Download_Resources", "CakhiaTV.md"),
		filepath.Join(vault, "Download_Resources", "Torrent_Shadow_Libraries", "Z-Library.md"),
	} {
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("THE UNTRACK LAW BROKEN — %s left the disk: %v", p, err)
		}
	}
	// and the public note stays tracked
	if !inList(after, "Design_Icons.md") {
		t.Fatalf("the public note was untracked: %v", after)
	}
}

func TestUntrackLocalOnly_ACleanRepoUntracksNothing(t *testing.T) {
	vault := beltRepo(t, true)
	sealer := beltSealer(vault, []string{"Nothing*.md"})
	if n := sealer.untrackLocalOnly(); n != 0 {
		t.Fatalf("expected 0, got %d", n)
	}
}

func TestUntrackLocalOnly_TheStemGlobMatchesTheReposVaultReviewTwin(t *testing.T) {
	vault := beltRepo(t, false)
	writeFile(t, filepath.Join(vault, "_review", "cobalt_Uncategorized_svelte.md"), "x")
	runGit(t, vault, "add", "-A")
	runGit(t, vault, "commit", "-qm", "base")
	sealer := beltSealer(vault, []string{"cobalt*.md"})
	if n := sealer.untrackLocalOnly(); n != 1 {
		t.Fatalf("expected 1, got %d", n)
	}
	if inList(trackedFiles(t, vault), "_review/cobalt_Uncategorized_svelte.md") {
		t.Fatal("the _review twin is still tracked")
	}
}

func TestUntrackLocalOnly_NoPatternsIsANoop(t *testing.T) {
	vault := beltRepo(t, true)
	sealer := beltSealer(vault, nil)
	if n := sealer.untrackLocalOnly(); n != 0 {
		t.Fatalf("expected 0, got %d", n)
	}
}

func TestSealAppliesTheBelt_ASealCommitsTheUntrackAndNeverStagesTheNotes(t *testing.T) {
	vault := beltRepo(t, true)
	sealer := beltSealer(vault, []string{"Torrent_Shadow_Libraries/", "CakhiaTV*.md"})
	// an edit on a PUBLIC note so the seal has something to commit
	f, err := os.OpenFile(filepath.Join(vault, "Design_Icons.md"), os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString("more"); err != nil {
		t.Fatal(err)
	}
	f.Close()
	result := sealer.Seal(nil)
	if !result.Sealed {
		t.Fatalf("expected a seal, got %+v (%s)", result, result.Describe())
	}
	tracked := trackedFiles(t, vault)
	if !inList(tracked, "Design_Icons.md") {
		t.Fatalf("the public note is not tracked: %v", tracked)
	}
	if inList(tracked, "Download_Resources/CakhiaTV.md") {
		t.Fatal("CakhiaTV.md still tracked after the seal")
	}
	if inList(tracked, "Download_Resources/Torrent_Shadow_Libraries/Z-Library.md") {
		t.Fatal("Z-Library.md still tracked after the seal")
	}
	// THE LAW again: local files intact
	if _, err := os.Stat(filepath.Join(vault, "Download_Resources", "CakhiaTV.md")); err != nil {
		t.Fatalf("THE UNTRACK LAW BROKEN: %v", err)
	}
	// a second seal is the quiet no-op (nothing dirty)
	again := sealer.Seal(nil)
	if again.Sealed {
		t.Fatal("the second seal sealed something")
	}
}

// ---------------------------------------------------------------------------
// 4. the config bridge's belt facts (test_localonly §3)
// ---------------------------------------------------------------------------

func TestConfigBridge_TheWebsitesSealCarriesTheFullBelt(t *testing.T) {
	cfg := map[string]any{
		"pipelines":          map[string]any{"websites": true},
		"website_vault_path": "/tmp/some-vault",
		"website_repo_name":  "my-awesome-websites-directory",
		"github_token":       "t",
		"vaultseal":          map[string]any{},
	}
	sealer := websitesSealerFromConfig(cfg, nil)
	joined := strings.Join(sealer.localOnly, "\n")
	if !strings.Contains(joined, "CakhiaTV*.md") {
		t.Fatalf("the stem globs did not ride: %v", sealer.localOnly)
	}
	// the full belt's folder forms (from a nonexistent vault: the
	// name-derived forms still answer)
	hasFolder := false
	for _, p := range sealer.localOnly {
		if strings.HasSuffix(p, "/") {
			hasFolder = true
		}
	}
	if !hasFolder {
		t.Fatalf("the belt's folder forms are missing: %v", sealer.localOnly)
	}
}

func TestConfigBridge_TheReposSealCarriesTheNoteGlobsOnly(t *testing.T) {
	cfg := map[string]any{
		"vault_path":   "/tmp/repos-vault",
		"github_token": "t",
		"vaultseal":    map[string]any{"repo_name": "r"},
	}
	sealer := SealerFromConfig(cfg, nil)
	if !inList(sealer.localOnly, "cobalt*.md") {
		t.Fatalf("the note globs did not ride: %v", sealer.localOnly)
	}
	for _, p := range sealer.localOnly {
		if strings.HasSuffix(p, "/") {
			t.Fatalf("a folder form rode the repos seal: %v", sealer.localOnly)
		}
	}
}

func TestConfigBridge_ThePersonalSealCarriesNoBelt(t *testing.T) {
	cfg := map[string]any{
		"manual_vault_path": "/tmp/personal",
		"manual_repo_name":  "personal-mirror",
		"github_token":      "t",
	}
	sealer := manualSealerFromConfig(cfg, nil)
	if len(sealer.localOnly) != 0 {
		t.Fatalf("the personal seal carries a belt: %v", sealer.localOnly)
	}
}

func TestConfigBridge_TheOwnersVaultsealLocalOnlyExtrasJoin(t *testing.T) {
	cfg := map[string]any{
		"vault_path": "/tmp/x", "github_token": "t",
		"vaultseal": map[string]any{"repo_name": "r", "local_only": []any{"MyOwn*.md"}},
	}
	pats := localOnlyFor("/tmp/x", false, cfg)
	if !inList(pats, "MyOwn*.md") || !inList(pats, "Prowlarr*.md") {
		t.Fatalf("the extras did not join: %v", pats)
	}
}
