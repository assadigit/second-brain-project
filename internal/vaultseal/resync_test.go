// resync_test.go — T-47: the ported selections of test_resyncfix.py
// (v0.36.0 the mirror resync) — the deletions-reach-the-mirror report,
// the fresh-vault reconciliation (no union-merge resurrection), the
// vault-wins-over-web-edits law, idempotence, the delta counts,
// describe()/to_dict(), the guards, and resync_from_config's
// explicit-action semantics. REAL local bare mirrors, no network.
package vaultseal

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/assadigit/second-brain-project/internal/storage"
)

func initVault(t *testing.T, path string) string {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
	runGit(t, path, "init", "-b", "main")
	runGit(t, path, "config", "user.email", "t@example.com")
	runGit(t, path, "config", "user.name", "T")
	return path
}

func resyncSealer(vault string) *VaultSeal {
	return NewVaultSeal(vault, "tok", "mirror", true, true, 0, nil, false, nil)
}

func isAncestorOf(t *testing.T, repo, maybeAncestor, descendant string) bool {
	t.Helper()
	rc, _ := tryGit(repo, "merge-base", "--is-ancestor", maybeAncestor, descendant)
	return rc == 0
}

// seedMirrorWithVault mirrors _seed_mirror_with_vault: push the
// vault's files to the mirror (the pre-bug state); returns the old
// tip sha.
func seedMirrorWithVault(t *testing.T, vault, bare string, files ...string) string {
	t.Helper()
	for _, name := range files {
		writeFile(t, filepath.Join(vault, name), "# "+name+"\n"+strings.Repeat("x", 100))
	}
	runGit(t, vault, "add", "-A")
	runGit(t, vault, "commit", "-m", "seal: the old state", "--quiet")
	runGit(t, vault, "push", bare, "HEAD:refs/heads/main")
	return strings.TrimSpace(runGit(t, bare, "rev-parse", "main"))
}

func TestResync_DeletionsReachTheMirror(t *testing.T) {
	// The owner's exact report: wrongly-created notes deleted in
	// Obsidian must leave the GitHub mirror too.
	tmp := t.TempDir()
	bare := bareRepo(t, filepath.Join(tmp, "mirror.git"))
	vault := initVault(t, filepath.Join(tmp, "vault"))
	oldTip := seedMirrorWithVault(t, vault, bare,
		"good-note.md", "wrong-youtube-note.md", "wrong-share-google.md")
	for _, name := range []string{"wrong-youtube-note.md", "wrong-share-google.md"} {
		if err := os.Remove(filepath.Join(vault, name)); err != nil {
			t.Fatal(err)
		}
	}

	res := resyncSealer(vault).Resync(bare)

	if !res.OK() {
		t.Fatalf("resync failed: %s", res.Describe())
	}
	if !res.Pushed {
		t.Fatal("expected pushed")
	}
	if got := remoteTree(t, bare); len(got) != 1 || got[0] != "good-note.md" {
		t.Fatalf("remote tree: %v", got)
	}
	if res.FilesDeleted != 2 {
		t.Fatalf("files_deleted: %d", res.FilesDeleted)
	}
	if res.FilesAdded != 0 {
		t.Fatalf("files_added: %d", res.FilesAdded)
	}
	if !strings.HasPrefix(res.CommitMessage, "resync:") {
		t.Fatalf("commit message: %q", res.CommitMessage)
	}
	// history kept: the old tip is still an ancestor of the new main
	if !isAncestorOf(t, bare, oldTip, "main") {
		t.Fatal("the old tip stopped being an ancestor")
	}
	// and the local vault is left clean (the commit landed on HEAD)
	if got := runGit(t, vault, "status", "--porcelain"); strings.TrimSpace(got) != "" {
		t.Fatalf("the vault is not clean: %q", got)
	}
}

func TestResync_ResyncIsIdempotent(t *testing.T) {
	tmp := t.TempDir()
	bare := bareRepo(t, filepath.Join(tmp, "mirror.git"))
	vault := initVault(t, filepath.Join(tmp, "vault"))
	seedMirrorWithVault(t, vault, bare, "a.md", "b.md")
	if err := os.Remove(filepath.Join(vault, "b.md")); err != nil {
		t.Fatal(err)
	}
	first := resyncSealer(vault).Resync(bare)
	if !first.OK() {
		t.Fatalf("first resync failed: %s", first.Describe())
	}
	second := resyncSealer(vault).Resync(bare)
	if !second.OK() {
		t.Fatalf("second resync failed: %s", second.Describe())
	}
	if second.SkippedReason != "mirror already matches the vault" {
		t.Fatalf("second resync skipped_reason: %q", second.SkippedReason)
	}
	if second.Sealed {
		t.Fatal("the second resync sealed something")
	}
}

func TestResync_FreshVaultVsPopulatedMirrorNeverResurrects(t *testing.T) {
	// The deleted-everything shape: a brand-new local repo (no .git
	// history — the vault folder was deleted and Obsidian recreated
	// it) against a populated mirror. The reconciliation commit makes
	// the mirror tree the vault EXACTLY — no resurrection.
	tmp := t.TempDir()
	bare := bareRepo(t, filepath.Join(tmp, "mirror.git"))
	seed := initVault(t, filepath.Join(tmp, "seed"))
	writeFile(t, filepath.Join(seed, "keep.md"), "# keep\n"+strings.Repeat("x", 100))
	writeFile(t, filepath.Join(seed, "wrong-note.md"), "# wrong\n"+strings.Repeat("x", 100))
	runGit(t, seed, "add", "-A")
	runGit(t, seed, "commit", "-m", "seal: both files", "--quiet")
	runGit(t, seed, "push", bare, "HEAD:refs/heads/main")
	seedTip := strings.TrimSpace(runGit(t, bare, "rev-parse", "main"))

	// the owner's fresh vault: only what should exist
	fresh := initVault(t, filepath.Join(tmp, "fresh-vault"))
	writeFile(t, filepath.Join(fresh, "keep.md"), "# keep\n"+strings.Repeat("x", 100))

	res := resyncSealer(fresh).Resync(bare)

	if !res.OK() {
		t.Fatalf("resync failed: %s", res.Describe())
	}
	if !res.Pushed {
		t.Fatal("expected pushed")
	}
	if got := remoteTree(t, bare); len(got) != 1 || got[0] != "keep.md" {
		t.Fatalf("remote tree: %v", got)
	}
	if res.FilesDeleted != 1 {
		t.Fatalf("files_deleted: %d", res.FilesDeleted)
	}
	// both histories kept: the mirror's old tip is still an ancestor
	if !isAncestorOf(t, bare, seedTip, "main") {
		t.Fatal("the seed tip stopped being an ancestor")
	}
}

func TestResync_MirrorAheadTheVaultWins(t *testing.T) {
	// A web-UI edit on the mirror (or a second machine) does not
	// survive a resync: the vault is the source of truth.
	tmp := t.TempDir()
	bare := bareRepo(t, filepath.Join(tmp, "mirror.git"))
	vault := initVault(t, filepath.Join(tmp, "vault"))
	oldTip := seedMirrorWithVault(t, vault, bare, "note.md")
	other := initVault(t, filepath.Join(tmp, "other"))
	runGit(t, other, "fetch", bare)
	runGit(t, other, "checkout", "-b", "work", "FETCH_HEAD")
	writeFile(t, filepath.Join(other, "web-edit.md"), "# edited on github\n"+strings.Repeat("x", 100))
	runGit(t, other, "add", "-A")
	runGit(t, other, "commit", "-m", "web edit", "--quiet")
	runGit(t, other, "push", bare, "HEAD:refs/heads/main")

	res := resyncSealer(vault).Resync(bare)

	if !res.OK() {
		t.Fatalf("resync failed: %s", res.Describe())
	}
	if got := remoteTree(t, bare); len(got) != 1 || got[0] != "note.md" {
		t.Fatalf("remote tree: %v", got)
	}
	if res.FilesDeleted != 1 {
		t.Fatalf("files_deleted: %d", res.FilesDeleted)
	}
	if !isAncestorOf(t, bare, oldTip, "main") {
		t.Fatal("the old tip stopped being an ancestor")
	}
}

func TestResync_AdditionsAndDeletionsTogether(t *testing.T) {
	tmp := t.TempDir()
	bare := bareRepo(t, filepath.Join(tmp, "mirror.git"))
	vault := initVault(t, filepath.Join(tmp, "vault"))
	seedMirrorWithVault(t, vault, bare, "old-a.md", "old-b.md")
	if err := os.Remove(filepath.Join(vault, "old-a.md")); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(vault, "new-c.md"), "# new\n"+strings.Repeat("x", 100))

	res := resyncSealer(vault).Resync(bare)

	if !res.OK() {
		t.Fatalf("resync failed: %s", res.Describe())
	}
	got := remoteTree(t, bare)
	sort.Strings(got)
	want := []string{"new-c.md", "old-b.md"}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("remote tree: %v", got)
	}
	if res.FilesDeleted != 1 || res.FilesAdded != 1 {
		t.Fatalf("delta counts: deleted=%d added=%d", res.FilesDeleted, res.FilesAdded)
	}
}

func TestResync_EmptyMirrorFirstResync(t *testing.T) {
	tmp := t.TempDir()
	bare := bareRepo(t, filepath.Join(tmp, "mirror.git"))
	vault := initVault(t, filepath.Join(tmp, "vault"))
	writeFile(t, filepath.Join(vault, "first.md"), "# first\n"+strings.Repeat("x", 100))
	res := resyncSealer(vault).Resync(bare)
	if !res.OK() {
		t.Fatalf("resync failed: %s", res.Describe())
	}
	if !res.Pushed {
		t.Fatal("expected pushed")
	}
	if got := remoteTree(t, bare); len(got) != 1 || got[0] != "first.md" {
		t.Fatalf("remote tree: %v", got)
	}
}

func TestResync_DescribeAndToDictCarryTheDelta(t *testing.T) {
	tmp := t.TempDir()
	bare := bareRepo(t, filepath.Join(tmp, "mirror.git"))
	vault := initVault(t, filepath.Join(tmp, "vault"))
	seedMirrorWithVault(t, vault, bare, "gone.md", "stays.md")
	if err := os.Remove(filepath.Join(vault, "gone.md")); err != nil {
		t.Fatal(err)
	}
	res := resyncSealer(vault).Resync(bare)
	line := res.Describe()
	if !strings.Contains(line, "resync: mirror follows the vault") {
		t.Fatalf("describe: %q", line)
	}
	if !strings.Contains(line, "1 file(s) deleted") {
		t.Fatalf("describe: %q", line)
	}
	d := res.ToMap()
	if v, _ := d.Get("files_deleted"); v != 1 {
		t.Fatalf("to_dict files_deleted: %v", v)
	}
	if v, _ := d.Get("files_added"); v != 0 {
		t.Fatalf("to_dict files_added: %v", v)
	}
	if v, _ := d.Get("pushed"); v != true {
		t.Fatalf("to_dict pushed: %v", v)
	}
}

func TestResync_NoTokenIsAClearError(t *testing.T) {
	tmp := t.TempDir()
	bare := bareRepo(t, filepath.Join(tmp, "mirror.git"))
	vault := initVault(t, filepath.Join(tmp, "vault"))
	writeFile(t, filepath.Join(vault, "a.md"), strings.Repeat("x", 50))
	res := NewVaultSeal(vault, "", "m", true, true, 0, nil, false, nil).Resync(bare)
	if res.OK() {
		t.Fatal("expected failure without a token")
	}
	if res.Pushed {
		t.Fatal("nothing was pushed")
	}
	if !strings.Contains(res.Error, "token") {
		t.Fatalf("error: %q", res.Error)
	}
}

func TestResync_MissingVaultIsAClearError(t *testing.T) {
	tmp := t.TempDir()
	bare := bareRepo(t, filepath.Join(tmp, "mirror.git"))
	res := NewVaultSeal(filepath.Join(tmp, "nope"), "tok", "m",
		true, true, 0, nil, false, nil).Resync(bare)
	if res.OK() {
		t.Fatal("expected failure for a missing vault")
	}
	if !strings.Contains(res.Error, "not found") {
		t.Fatalf("error: %q", res.Error)
	}
}

func TestResync_SealResultDefaultsUnchanged(t *testing.T) {
	// The two delta fields default to 0 — every existing seal result
	// shape is untouched.
	r := NewSealResult()
	r.Sealed = true
	r.FilesChanged = 3
	if r.FilesAdded != 0 || r.FilesDeleted != 0 {
		t.Fatalf("delta defaults: %d/%d", r.FilesAdded, r.FilesDeleted)
	}
	d := r.ToMap()
	if v, _ := d.Get("files_added"); v != 0 {
		t.Fatalf("to_dict files_added: %v", v)
	}
	if v, _ := d.Get("files_deleted"); v != 0 {
		t.Fatalf("to_dict files_deleted: %v", v)
	}
}

// ---------------------------------------------------------------------------
// resync_from_config — both vaults, explicit-action semantics
// ---------------------------------------------------------------------------

type ctorCapture struct {
	vaultPath string
	token     string
	repoName  string
	autoPush  bool
	enabled   bool
	localOnly []string
}

func captureResyncFromConfig(t *testing.T, config any) ([]ctorCapture, *storage.OrderedMap) {
	t.Helper()
	made := &[]ctorCapture{}
	orig := vaultSealCtor
	vaultSealCtor = func(vaultPath, token, repoName string, autoPush, enabled bool,
		timeout int, log LogFn, allowCreate bool, localOnly []string) *VaultSeal {
		*made = append(*made, ctorCapture{vaultPath, token, repoName, autoPush, enabled, localOnly})
		return NewVaultSeal(vaultPath, token, repoName, autoPush, enabled, timeout, log, allowCreate, localOnly)
	}
	t.Cleanup(func() { vaultSealCtor = orig })
	out := ResyncFromConfig(config, nil)
	return *made, out
}

func TestResyncFromConfig_BothVaultsResynced(t *testing.T) {
	made, out := captureResyncFromConfig(t, map[string]any{
		"vault_path": "/v", "github_token": "tok",
		"website_vault_path": "/w", "website_repo_name": "webrepo",
		"pipelines": map[string]any{"websites": true},
		"vaultseal": map[string]any{"repo_name": "ghrepo"},
	})
	if len(made) != 2 {
		t.Fatalf("expected 2 sealers built, got %d (%+v)", len(made), made)
	}
	gh, web := made[0], made[1]
	if gh.vaultPath != "/v" || gh.repoName != "ghrepo" {
		t.Fatalf("github sealer: %+v", gh)
	}
	if web.vaultPath != "/w" || web.repoName != "webrepo" {
		t.Fatalf("websites sealer: %+v", web)
	}
	for _, m := range made {
		if m.token != "tok" {
			t.Fatalf("token: %+v", m)
		}
		// explicit owner action: the post-run switches never gate it
		if !m.autoPush || !m.enabled {
			t.Fatalf("the explicit-action switches are off: %+v", m)
		}
	}
	if _, ok := out.Get("github"); !ok {
		t.Fatal("the github key is missing")
	}
	if _, ok := out.Get("websites"); !ok {
		t.Fatal("the websites key is missing")
	}
}

func TestResyncFromConfig_WebsitesSkippedWhenPipelineOff(t *testing.T) {
	made, out := captureResyncFromConfig(t, map[string]any{
		"vault_path": "/v", "github_token": "tok",
		"website_vault_path": "/w",
		"pipelines":          map[string]any{"websites": false},
	})
	if len(made) != 1 { // only the GitHub vault
		t.Fatalf("expected 1 sealer built, got %d", len(made))
	}
	v, _ := out.Get("websites")
	res := v.(*SealResult)
	if res.SkippedReason != "websites pipeline off or no websites vault" {
		t.Fatalf("websites skipped_reason: %q", res.SkippedReason)
	}
}

func TestResyncFromConfig_NoVaultsConfigured(t *testing.T) {
	made, out := captureResyncFromConfig(t, map[string]any{"github_token": "tok"})
	if len(made) != 1 { // the GitHub sealer is still built
		t.Fatalf("expected 1 sealer built, got %d", len(made))
	}
	v, _ := out.Get("github")
	res := v.(*SealResult)
	if res.SkippedReason != "no GitHub vault configured" {
		t.Fatalf("github skipped_reason: %q", res.SkippedReason)
	}
}
