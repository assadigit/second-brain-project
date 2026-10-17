// reporadar_test.go — T-48: the ported selections of the reference's
// own suite (app/tests/test_reporadar.py — the pure evidence pieces,
// the ladder's every branch, the three slots together, the read-only
// law, and the owner's exact end-to-end scenario):
//
//   - TestRemoteParsing (§1: the https/ssh/git@ forms, the owner
//     lowercased, non-GitHub never an owner);
//   - TestRemoteOrigin (§1: the seal's own signature read from a real
//     .git/config — hermetic: git init only, no network);
//   - TestBlobSha (§1: the canonical git fixtures + the real-git
//     `hash-object` cross-check the spec's test plan demands);
//   - TestSampling (§1: the diversified round-robin, .git/.trash
//     skipped, the blank-path law);
//   - TestDetectBackupRepo (§2: the evidence ladder — remote certain,
//     the unverifiable fall-through, the name evidence (incl. the
//     not-yet-created folder), the fingerprint's dominance, the
//     size-0 live lesson, the candidate crowding cure, the
//     archived/fork filter, ambiguity never a guess, sub-threshold
//     weakness, the no-match law, the honest closing line);
//   - TestRadarAllVaults (§3: the no-token sleep, the rejected token,
//     the empty-slot-detected / owner-slot-kept pair, the missing
//     path, the never-raises doors — a panicking API surface is
//     recovered and SPOKEN);
//   - TestApplyDetections (§3: only filled slots land, owner
//     confidence never overwrites, plain dicts flow, nested sets keep
//     neighbours, idempotence, garbage inputs);
//   - TestReportLines / TestRadarPending (the honest report lines;
//     the quiet-startup gate);
//   - TestReadOnlyLaw (§4: the fake API sees only read paths);
//   - TestManualVaultSlot (§5: the constants' third-slot contract);
//   - TestTheOwnersExactScenario (§7: paths in, repos out — zero
//     typing; the fingerprint story when folders were renamed).
//
// TEMP VAULTS ONLY; NO real network (the API surface is the FakeAPI
// seam exactly like the reference's own tests).
package reporadar

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/assadigit/second-brain-project/internal/constants"
)

// ---------------------------------------------------------------------------
// The fake GitHub API (the real client's exact call shape)
// ---------------------------------------------------------------------------

// fakeRadarAPI mirrors FakeAPI: a fixed account, repo list and tree
// contents; calls records every path requested — the read-only law is
// provable from it.
type fakeRadarAPI struct {
	login  string
	repos  []map[string]any
	trees  map[string][]string // {full_name: [blob_sha, ...]}
	exists map[string]map[string]any
	meOK   bool
	calls  []string
}

func newFakeAPI(opts func(*fakeRadarAPI)) *fakeRadarAPI {
	f := &fakeRadarAPI{login: "owner", meOK: true,
		trees: map[string][]string{}, exists: map[string]map[string]any{}}
	if opts != nil {
		opts(f)
	}
	return f
}

func (f *fakeRadarAPI) Me() string {
	f.calls = append(f.calls, "/user")
	if !f.meOK {
		return ""
	}
	return f.login
}

func (f *fakeRadarAPI) ListRepos() []map[string]any {
	f.calls = append(f.calls, "/user/repos")
	out := make([]map[string]any, len(f.repos))
	copy(out, f.repos)
	return out
}

func (f *fakeRadarAPI) RepoExists(name string) map[string]any {
	f.calls = append(f.calls, "/repos/"+name)
	if p, ok := f.exists[name]; ok && p != nil {
		return p
	}
	return nil
}

func (f *fakeRadarAPI) TreeBlobShas(fullName, ref string) map[string]bool {
	f.calls = append(f.calls, "/repos/"+fullName+"/git/trees/"+ref)
	shas, ok := f.trees[fullName]
	if !ok || shas == nil {
		return nil
	}
	out := map[string]bool{}
	for _, s := range shas {
		out[s] = true
	}
	return out
}

// boomMe / boomRepoExists mirror BoomAPI / BoomLater: the raise doors
// the never-raises law must recover and speak.
type boomMe struct{ fakeRadarAPI }

func (b *boomMe) Me() string { panic(errors.New("boom")) }

type boomRepoExists struct{ fakeRadarAPI }

func (b *boomRepoExists) RepoExists(string) map[string]any { panic(errors.New("boom")) }

// repoPayload mirrors _repo.
func repoPayload(name string, size int, defaultBranch string, archived, fork bool) map[string]any {
	return map[string]any{
		"name":           name,
		"full_name":      "owner/" + name,
		"size":           size,
		"default_branch": defaultBranch,
		"archived":       archived,
		"fork":           fork,
	}
}

func rp(name string) map[string]any { return repoPayload(name, 1000, "main", false, false) }

func gitBlob(content string) string { return BlobSha([]byte(content)) }

func writeVaultFile(t *testing.T, root, rel, content string) {
	t.Helper()
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func makeVault(t *testing.T, tmp, name string, files map[string]string) string {
	t.Helper()
	root := filepath.Join(tmp, name)
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	// deterministic creation order: sorted names per directory (Go's
	// ReadDir walks sorted either way)
	keys := make([]string, 0, len(files))
	for k := range files {
		keys = append(keys, k)
	}
	sortStrings(keys)
	for _, rel := range keys {
		writeVaultFile(t, root, rel, files[rel])
	}
	return root
}

// makeRemoteVault mirrors _make_remote_vault: a vault that IS a git
// repo with a GitHub origin — the seal's own signature (evidence 1).
// Hermetic: git init only, no network.
func makeRemoteVault(t *testing.T, tmp, name, owner, repo string) string {
	t.Helper()
	root := makeVault(t, tmp, name, map[string]string{"note.md": "# hi\n"})
	out, err := exec.Command("git", "-C", root, "init", "-b", "main").CombinedOutput()
	if err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	cfg := filepath.Join(root, ".git", "config")
	f, err := os.OpenFile(cfg, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err := f.WriteString("[remote \"origin\"]\n\turl = https://github.com/" +
		owner + "/" + repo + ".git\n"); err != nil {
		t.Fatal(err)
	}
	return root
}

func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}

// ---------------------------------------------------------------------------
// 1. The pure evidence pieces
// ---------------------------------------------------------------------------

func TestRemoteParsing_HTTPSForms(t *testing.T) {
	cases := [][2]string{
		{"https://github.com/owner/repo", "owner/repo"},
		{"https://github.com/owner/repo.git", "owner/repo"},
		{"https://github.com/owner/repo/", "owner/repo"},
	}
	for _, c := range cases {
		if got := ParseRemoteRepo(c[0]); got != c[1] {
			t.Errorf("parse_remote_repo(%q) = %q, want %q", c[0], got, c[1])
		}
	}
}

func TestRemoteParsing_SSHAndGitForms(t *testing.T) {
	cases := [][2]string{
		{"git@github.com:owner/repo.git", "owner/repo"},
		{"ssh://git@github.com/owner/repo.git", "owner/repo"},
	}
	for _, c := range cases {
		if got := ParseRemoteRepo(c[0]); got != c[1] {
			t.Errorf("parse_remote_repo(%q) = %q, want %q", c[0], got, c[1])
		}
	}
}

func TestRemoteParsing_OwnerIsLowercased(t *testing.T) {
	if got := ParseRemoteRepo("https://github.com/Owner/Repo"); got != "owner/Repo" {
		t.Errorf("got %q", got)
	}
}

func TestRemoteParsing_NonGitHubIsNone(t *testing.T) {
	for _, u := range []string{
		"https://gitlab.com/owner/repo", "", "https://github.com/git/x", // 'git' is not an owner
	} {
		if got := ParseRemoteRepo(u); got != "" {
			t.Errorf("parse_remote_repo(%q) = %q, want empty", u, got)
		}
	}
}

func TestRemoteOrigin_ReadsTheSealsSignature(t *testing.T) {
	tmp := t.TempDir()
	vault := makeRemoteVault(t, tmp, "my-vault", "assadigit", "Ali-personal-obsidian-vault")
	// the owner is lowercased, the repo keeps its own case (GitHub
	// repo names are case-preserving; the API lookup is tolerant)
	if got := RemoteOriginRepo(vault); got != "assadigit/Ali-personal-obsidian-vault" {
		t.Errorf("got %q", got)
	}
}

func TestRemoteOrigin_MissingGitConfigIsNone(t *testing.T) {
	tmp := t.TempDir()
	vault := makeVault(t, tmp, "plain", map[string]string{"a.md": "x"})
	if got := RemoteOriginRepo(vault); got != "" {
		t.Errorf("got %q", got)
	}
}

func TestRemoteOrigin_MissingFolderIsNone(t *testing.T) {
	tmp := t.TempDir()
	if got := RemoteOriginRepo(filepath.Join(tmp, "nope")); got != "" {
		t.Errorf("got %q", got)
	}
	if got := RemoteOriginRepo(""); got != "" {
		t.Errorf("got %q", got)
	}
}

func TestBlobSha_KnownGitBlobs(t *testing.T) {
	// the canonical git fixtures — the same SHA GitHub's tree API
	// reports for the same bytes, whatever the path
	if got := BlobSha([]byte("")); got != "e69de29bb2d1d6434b8b29ae775ad8c2e48c5391" {
		t.Errorf("empty blob: %s", got)
	}
	if got := BlobSha([]byte("hello\n")); got != "ce013625030ba8dba906f756967f9e9ca394464a" {
		t.Errorf("hello blob: %s", got)
	}
}

func TestBlobSha_MatchesRealGit(t *testing.T) {
	// the spec's test plan: blob_sha correctness vs REAL git hashes
	tmp := t.TempDir()
	for _, content := range []string{"", "hello\n", "the vault's content\n", "unicode ⭐ content\n"} {
		path := filepath.Join(tmp, "f"+string(content[:minInt(len(content), 3)])+".md")
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		out, err := exec.Command("git", "hash-object", path).CombinedOutput()
		if err != nil {
			t.Fatalf("git hash-object: %v\n%s", err, out)
		}
		want := strings.TrimSpace(string(out))
		if got := BlobSha([]byte(content)); got != want {
			t.Errorf("blob_sha(%q) = %s, git says %s", content, got, want)
		}
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TestBlobSha_ContentAddressedNotPathAddressed(t *testing.T) {
	if BlobSha([]byte("same")) != BlobSha([]byte("same")) {
		t.Error("same content, different shas")
	}
}

func TestSampling_DiversifiedRoundRobin(t *testing.T) {
	// two folders, one huge: the sample must NOT be monopolised by the
	// huge folder (round-robin across top-level groups)
	tmp := t.TempDir()
	files := map[string]string{}
	for i := 0; i < 30; i++ {
		files[fmt.Sprintf("big/f%02d.md", i)] = "x"
	}
	for i := 0; i < 3; i++ {
		files[fmt.Sprintf("sml/a%d.md", i)] = "y"
	}
	root := makeVault(t, tmp, "vault", files)
	sample := SampleVaultFiles(root, 10, 0)
	if len(sample) != 10 {
		t.Fatalf("len(sample) = %d", len(sample))
	}
	tops := map[string]bool{}
	for _, s := range sample {
		tops[strings.Split(s, "/")[0]] = true
	}
	if len(tops) != 2 || !tops["big"] || !tops["sml"] { // both voices present
		t.Errorf("tops: %v", tops)
	}
}

func TestSampling_GitAndTrashSkipped(t *testing.T) {
	tmp := t.TempDir()
	root := makeVault(t, tmp, "vault", map[string]string{
		"n.md": "x", ".git/config": "junk", ".trash/old.md": "x",
	})
	sample := SampleVaultFiles(root, 0, 0)
	if len(sample) != 1 || sample[0] != "n.md" {
		t.Fatalf("sample: %v", sample)
	}
}

func TestSampling_BlankPathIsNeverTheWorkingFolder(t *testing.T) {
	// the caught bug (v0.84.0 dev): Path("") resolves to "." and a
	// blank vault path fingerprinted the APP'S OWN source tree
	tmp := t.TempDir()
	if got := SampleVaultFiles("", 0, 0); got != nil {
		t.Errorf("blank: %v", got)
	}
	if got := SampleVaultFiles("   ", 0, 0); got != nil {
		t.Errorf("spaces: %v", got)
	}
	if got := SampleVaultFiles(filepath.Join(tmp, "gone"), 0, 0); got != nil {
		t.Errorf("gone: %v", got)
	}
}

// ---------------------------------------------------------------------------
// 2. detect_backup_repo — the evidence ladder, one vault at a time
// ---------------------------------------------------------------------------

func TestDetectBackupRepo_NoPathIsAReportNotAnError(t *testing.T) {
	det := DetectBackupRepo("github", "", newFakeAPI(nil), "owner")
	if det.RepoName() != "" {
		t.Errorf("repo_name: %s", det.RepoName())
	}
	if !strings.Contains(det.Detail(), "no vault path") {
		t.Errorf("detail: %s", det.Detail())
	}
}

func TestDetectBackupRepo_NoOwnerIsAReportNotAnError(t *testing.T) {
	det := DetectBackupRepo("github", "/v", newFakeAPI(nil), "")
	if det.RepoName() != "" {
		t.Errorf("repo_name: %s", det.RepoName())
	}
	if !strings.Contains(det.Detail(), "could not resolve") {
		t.Errorf("detail: %s", det.Detail())
	}
}

func TestDetectBackupRepo_Evidence1RemoteCertain(t *testing.T) {
	tmp := t.TempDir()
	vault := makeRemoteVault(t, tmp, "Ali-personal-obsidian-vault",
		"assadigit", "Ali-personal-obsidian-vault")
	api := newFakeAPI(func(f *fakeRadarAPI) {
		f.exists["assadigit/Ali-personal-obsidian-vault"] =
			map[string]any{"name": "Ali-personal-obsidian-vault"}
	})
	det := DetectBackupRepo("personal", vault, api, "assadigit")
	if det.RepoName() != "Ali-personal-obsidian-vault" {
		t.Errorf("repo_name: %s", det.RepoName())
	}
	if det.Confidence() != "certain" || det.Method() != "remote" {
		t.Errorf("confidence/method: %s/%s", det.Confidence(), det.Method())
	}
}

func TestDetectBackupRepo_RemoteUnverifiableFallsThrough(t *testing.T) {
	// the origin points somewhere the account cannot see → the remote
	// is NOT trusted; detection continues (no crash)
	tmp := t.TempDir()
	vault := makeRemoteVault(t, tmp, "v", "owner", "ghost-repo")
	api := newFakeAPI(func(f *fakeRadarAPI) {
		f.exists["owner/ghost-repo"] = map[string]any{"_error": "HTTP 403"}
	})
	det := DetectBackupRepo("github", vault, api, "owner")
	if det.RepoName() != "" { // nothing else to find
		t.Errorf("repo_name: %s", det.RepoName())
	}
}

func TestDetectBackupRepo_Evidence2NameCertain(t *testing.T) {
	// the cloned-vault case: the folder's own name IS the repo's
	tmp := t.TempDir()
	vault := makeVault(t, tmp, "github-projects-obsidian-vault", map[string]string{"n.md": "x"})
	api := newFakeAPI(func(f *fakeRadarAPI) {
		f.exists["owner/github-projects-obsidian-vault"] =
			map[string]any{"name": "github-projects-obsidian-vault"}
	})
	det := DetectBackupRepo("github", vault, api, "owner")
	if det.RepoName() != "github-projects-obsidian-vault" {
		t.Errorf("repo_name: %s", det.RepoName())
	}
	if det.Confidence() != "certain" || det.Method() != "name" {
		t.Errorf("confidence/method: %s/%s", det.Confidence(), det.Method())
	}
}

func TestDetectBackupRepo_Evidence2WorksForANotYetCreatedFolder(t *testing.T) {
	// the Websites vault "will be created" case: the promised path
	// names the repo even before the folder exists
	tmp := t.TempDir()
	promised := filepath.Join(tmp, "my-awesome-websites-directory")
	api := newFakeAPI(func(f *fakeRadarAPI) {
		f.exists["owner/my-awesome-websites-directory"] =
			map[string]any{"name": "my-awesome-websites-directory"}
	})
	det := DetectBackupRepo("websites", promised, api, "owner")
	if det.RepoName() != "my-awesome-websites-directory" {
		t.Errorf("repo_name: %s", det.RepoName())
	}
	if det.Method() != "name" {
		t.Errorf("method: %s", det.Method())
	}
}

func TestDetectBackupRepo_Evidence3FingerprintDominant(t *testing.T) {
	tmp := t.TempDir()
	files := map[string]string{}
	for i := 0; i < 6; i++ {
		files[fmt.Sprintf("n%d.md", i)] = fmt.Sprintf("# note %d\n", i) + strings.Repeat("z", 50)
	}
	vault := makeVault(t, tmp, "moved-vault", files) // not the repo's name
	var shas []string
	keys := make([]string, 0, len(files))
	for k := range files {
		keys = append(keys, k)
	}
	sortStrings(keys)
	for _, k := range keys {
		shas = append(shas, gitBlob(files[k]))
	}
	api := newFakeAPI(func(f *fakeRadarAPI) {
		f.repos = []map[string]any{rp("the-real-backup"), rp("something-else")}
		f.trees["owner/the-real-backup"] = shas
		f.trees["owner/something-else"] = []string{gitBlob("unrelated")}
	})
	det := DetectBackupRepo("github", vault, api, "owner")
	if det.RepoName() != "the-real-backup" {
		t.Errorf("repo_name: %s", det.RepoName())
	}
	if det.Confidence() != "strong" || det.Method() != "fingerprint" {
		t.Errorf("confidence/method: %s/%s", det.Confidence(), det.Method())
	}
	if det.Matched() != 6 || det.Sampled() != 6 {
		t.Errorf("matched/sampled: %d/%d", det.Matched(), det.Sampled())
	}
}

func TestDetectBackupRepo_SizeZeroRepoIsStillFingerprinted(t *testing.T) {
	// THE LIVE-ACCOUNT CATCH (v0.84.0 dev): GitHub's repos API reports
	// size in KB ROUNDED DOWN — a size-0 repo is NOT empty; it must
	// stay a fingerprint candidate.
	tmp := t.TempDir()
	content := "# the personal vault\n" + strings.Repeat("y", 1100)
	vault := makeVault(t, tmp, "renamed-folder", map[string]string{"README.md": content})
	api := newFakeAPI(func(f *fakeRadarAPI) {
		f.repos = []map[string]any{repoPayload("Ali-personal-obsidian-vault", 0, "main", false, false)}
		f.trees["owner/Ali-personal-obsidian-vault"] = []string{gitBlob(content)}
	})
	det := DetectBackupRepo("personal", vault, api, "owner")
	if det.RepoName() != "Ali-personal-obsidian-vault" {
		t.Errorf("repo_name: %s", det.RepoName())
	}
	if det.Method() != "fingerprint" {
		t.Errorf("method: %s", det.Method())
	}
	// a 1/1 full-content match is legitimate evidence (threshold
	// collapses to the sample size below MIN_DOMINANT_MATCHES)
	if det.Matched() != 1 || det.Sampled() != 1 {
		t.Errorf("matched/sampled: %d/%d", det.Matched(), det.Sampled())
	}
}

func TestDetectBackupRepo_TheTrueBackupSurvivesCandidateCrowding(t *testing.T) {
	// THE LIVE-ACCOUNT CATCH, part two: with the fetch cap too low, an
	// account's vault-ish names alone crowded the TRUE backup out of
	// the candidate list. Every non-archived, non-fork repo is a
	// candidate now (24 fit; bigger accounts rank vault-ish first).
	tmp := t.TempDir()
	files := map[string]string{}
	for i := 0; i < 4; i++ {
		files[fmt.Sprintf("n%d.md", i)] = fmt.Sprintf("# note %d\n", i)
	}
	vault := makeVault(t, tmp, "renamed", files)
	var shas []string
	keys := make([]string, 0, len(files))
	for k := range files {
		keys = append(keys, k)
	}
	sortStrings(keys)
	for _, k := range keys {
		shas = append(shas, gitBlob(files[k]))
	}
	api := newFakeAPI(func(f *fakeRadarAPI) {
		for i := 0; i < 20; i++ {
			f.repos = append(f.repos, repoPayload(fmt.Sprintf("filler-%02d-vault", i), 5000, "main", false, false))
		}
		f.repos = append(f.repos, repoPayload("Ali-personal-obsidian-vault", 0, "main", false, false))
		f.trees["owner/Ali-personal-obsidian-vault"] = shas
	})
	det := DetectBackupRepo("personal", vault, api, "owner")
	if det.RepoName() != "Ali-personal-obsidian-vault" {
		t.Errorf("repo_name: %s", det.RepoName())
	}
}

func TestDetectBackupRepo_ArchivedAndForkReposAreNotCandidates(t *testing.T) {
	tmp := t.TempDir()
	files := map[string]string{}
	for i := 0; i < 4; i++ {
		files[fmt.Sprintf("n%d.md", i)] = fmt.Sprintf("# note %d\n", i)
	}
	vault := makeVault(t, tmp, "v", files)
	var shas []string
	keys := make([]string, 0, len(files))
	for k := range files {
		keys = append(keys, k)
	}
	sortStrings(keys)
	for _, k := range keys {
		shas = append(shas, gitBlob(files[k]))
	}
	api := newFakeAPI(func(f *fakeRadarAPI) {
		f.repos = []map[string]any{
			repoPayload("archived-backup", 5000, "main", true, false),
			repoPayload("forked-backup", 5000, "main", false, true),
		}
		f.trees["owner/archived-backup"] = shas
		f.trees["owner/forked-backup"] = shas
	})
	det := DetectBackupRepo("github", vault, api, "owner")
	if det.RepoName() != "" { // never looked at them
		t.Errorf("repo_name: %s", det.RepoName())
	}
}

func TestDetectBackupRepo_FingerprintAmbiguityIsAReportNeverAGuess(t *testing.T) {
	// two repos share the content near-equally → NO winner
	tmp := t.TempDir()
	files := map[string]string{}
	for i := 0; i < 4; i++ {
		files[fmt.Sprintf("n%d.md", i)] = fmt.Sprintf("# note %d\n", i)
	}
	vault := makeVault(t, tmp, "v", files)
	var shas []string
	keys := make([]string, 0, len(files))
	for k := range files {
		keys = append(keys, k)
	}
	sortStrings(keys)
	for _, k := range keys {
		shas = append(shas, gitBlob(files[k]))
	}
	api := newFakeAPI(func(f *fakeRadarAPI) {
		f.repos = []map[string]any{rp("twin-a"), rp("twin-b")}
		f.trees["owner/twin-a"] = shas
		f.trees["owner/twin-b"] = shas
	})
	det := DetectBackupRepo("github", vault, api, "owner")
	if det.RepoName() != "" {
		t.Errorf("repo_name: %s", det.RepoName())
	}
	if !strings.Contains(det.Detail(), "ambiguous") {
		t.Errorf("detail: %s", det.Detail())
	}
}

func TestDetectBackupRepo_FingerprintBelowMinMatchesIsNotAWinner(t *testing.T) {
	// 5 files sampled, only 2 shared with the single candidate →
	// sub-threshold weakness: a report, never a guess
	tmp := t.TempDir()
	files := map[string]string{}
	for i := 0; i < 5; i++ {
		files[fmt.Sprintf("n%d.md", i)] = fmt.Sprintf("# note %d\n", i)
	}
	vault := makeVault(t, tmp, "v", files)
	var shas []string
	keys := make([]string, 0, len(files))
	for k := range files {
		keys = append(keys, k)
	}
	sortStrings(keys)
	for _, k := range keys {
		shas = append(shas, gitBlob(files[k]))
	}
	api := newFakeAPI(func(f *fakeRadarAPI) {
		f.repos = []map[string]any{rp("solo")}
		f.trees["owner/solo"] = shas[:2]
	})
	det := DetectBackupRepo("github", vault, api, "owner")
	if det.RepoName() != "" {
		t.Errorf("repo_name: %s", det.RepoName())
	}
	if !strings.Contains(det.Detail(), "not enough evidence") {
		t.Errorf("detail: %s", det.Detail())
	}
	if det.Matched() != 2 || det.Sampled() != 5 {
		t.Errorf("matched/sampled: %d/%d", det.Matched(), det.Sampled())
	}
}

func TestDetectBackupRepo_NoMatchAndNoCreation(t *testing.T) {
	tmp := t.TempDir()
	files := map[string]string{}
	for i := 0; i < 5; i++ {
		files[fmt.Sprintf("n%d.md", i)] = fmt.Sprintf("# note %d\n", i)
	}
	vault := makeVault(t, tmp, "v", files)
	api := newFakeAPI(func(f *fakeRadarAPI) {
		f.repos = []map[string]any{rp("other")}
		f.trees["owner/other"] = []string{gitBlob("unrelated")}
	})
	det := DetectBackupRepo("github", vault, api, "owner")
	if det.RepoName() != "" {
		t.Errorf("repo_name: %s", det.RepoName())
	}
	if !strings.Contains(det.Detail(), "no repo on the account shares") {
		t.Errorf("detail: %s", det.Detail())
	}
}

func TestDetectBackupRepo_MissingFolderNoNameNoFingerprintSaysTheLaw(t *testing.T) {
	// the folder does not exist AND no repo carries its name → the
	// honest closing line: nothing was created (v0.83.0 law)
	tmp := t.TempDir()
	promised := filepath.Join(tmp, "no-such-vault-yet")
	det := DetectBackupRepo("github", promised, newFakeAPI(nil), "owner")
	if det.RepoName() != "" {
		t.Errorf("repo_name: %s", det.RepoName())
	}
	if !strings.Contains(det.Detail(), "nothing was created") {
		t.Errorf("detail: %s", det.Detail())
	}
	if !strings.Contains(det.Detail(), "next radar pass will find it") {
		t.Errorf("detail: %s", det.Detail())
	}
}

// ---------------------------------------------------------------------------
// 3. radar_all_vaults / apply_detections — the three slots together
// ---------------------------------------------------------------------------

func radarCfg(over map[string]any) map[string]any {
	cfg := map[string]any{"github_token": "tok"}
	for k, v := range over {
		cfg[k] = v
	}
	return cfg
}

func TestRadarAllVaults_NoTokenSleeps(t *testing.T) {
	out := RadarAllVaults(map[string]any{"github_token": "", "vault_path": "/v"}, nil)
	if len(out) != 3 {
		t.Fatalf("kinds: %v", out)
	}
	for _, det := range out {
		if det.RepoName() != "" {
			t.Error("repo_name should be empty")
		}
	}
	if !strings.Contains(out["github"].Detail(), "no GitHub token") {
		t.Errorf("detail: %s", out["github"].Detail())
	}
}

func TestRadarAllVaults_RejectedTokenIsReported(t *testing.T) {
	api := newFakeAPI(func(f *fakeRadarAPI) { f.meOK = false })
	out := RadarAllVaults(radarCfg(map[string]any{"vault_path": "/v"}), api)
	if !strings.Contains(out["github"].Detail(), "rejected") {
		t.Errorf("detail: %s", out["github"].Detail())
	}
	if out["github"].RepoName() != "" {
		t.Errorf("repo_name: %s", out["github"].RepoName())
	}
}

func TestRadarAllVaults_EmptySlotDetectedOwnerSlotKept(t *testing.T) {
	tmp := t.TempDir()
	ghVault := makeVault(t, tmp, "github-projects-obsidian-vault", map[string]string{"a.md": "x"})
	webVault := makeVault(t, tmp, "my-awesome-websites-directory", map[string]string{"b.md": "y"})
	perVault := makeVault(t, tmp, "Ali-personal-obsidian-vault", map[string]string{"c.md": "z"})
	api := newFakeAPI(func(f *fakeRadarAPI) {
		f.exists["owner/github-projects-obsidian-vault"] =
			map[string]any{"name": "github-projects-obsidian-vault"}
		f.exists["owner/my-awesome-websites-directory"] =
			map[string]any{"name": "my-awesome-websites-directory"}
		f.exists["owner/Ali-personal-obsidian-vault"] =
			map[string]any{"name": "Ali-personal-obsidian-vault"}
	})
	cfg := radarCfg(map[string]any{
		"vault_path": ghVault, "website_vault_path": webVault, "manual_vault_path": perVault,
	})
	out := RadarAllVaults(cfg, api)
	if out["github"].RepoName() != "github-projects-obsidian-vault" {
		t.Errorf("github: %s", out["github"].RepoName())
	}
	if out["websites"].RepoName() != "my-awesome-websites-directory" {
		t.Errorf("websites: %s", out["websites"].RepoName())
	}
	if out["personal"].RepoName() != "Ali-personal-obsidian-vault" {
		t.Errorf("personal: %s", out["personal"].RepoName())
	}

	// the owner's typed word wins: a SET slot is reported as owner
	cfg2 := radarCfg(map[string]any{
		"vault_path":         ghVault,
		"vaultseal":          map[string]any{"repo_name": "my-own-choice"},
		"website_vault_path": webVault,
		"manual_vault_path":  perVault,
	})
	out2 := RadarAllVaults(cfg2, api)
	if out2["github"].RepoName() != "my-own-choice" {
		t.Errorf("github: %s", out2["github"].RepoName())
	}
	if out2["github"].Confidence() != "owner" || out2["github"].Method() != "config" {
		t.Errorf("confidence/method: %s/%s", out2["github"].Confidence(), out2["github"].Method())
	}
	// and the other two still detect
	if out2["websites"].RepoName() != "my-awesome-websites-directory" {
		t.Errorf("websites: %s", out2["websites"].RepoName())
	}
}

func TestRadarAllVaults_MissingPathSlotReportsNotConfigured(t *testing.T) {
	out := RadarAllVaults(radarCfg(nil), newFakeAPI(nil))
	if out["github"].RepoName() != "" {
		t.Errorf("repo_name: %s", out["github"].RepoName())
	}
	if !strings.Contains(out["github"].Detail(), "no vault path") {
		t.Errorf("detail: %s", out["github"].Detail())
	}
}

func TestRadarAllVaults_NeverRaisesEvenWhenDetectionExplodes(t *testing.T) {
	api := &boomMe{}
	out := RadarAllVaults(radarCfg(map[string]any{"vault_path": "/v"}), api)
	if len(out) != 3 {
		t.Fatalf("kinds: %v", out)
	}
	for _, det := range out {
		if det.RepoName() != "" {
			t.Error("repo_name should be empty")
		}
	}
	// the explosion is SPOKEN, not swallowed silently
	if !strings.Contains(out["github"].Detail(), "account probe failed") {
		t.Errorf("detail: %s", out["github"].Detail())
	}

	api2 := &boomRepoExists{fakeRadarAPI{login: "owner", meOK: true}}
	out2 := RadarAllVaults(radarCfg(map[string]any{"vault_path": "/v"}), api2)
	if len(out2) != 3 {
		t.Fatalf("kinds: %v", out2)
	}
	if !strings.Contains(out2["github"].Detail(), "detection failed") {
		t.Errorf("detail: %s", out2["github"].Detail())
	}
}

func applyResults() map[string]any {
	return map[string]any{
		"github":   NewDetection("kind", "github", "repo_name", "gh-repo", "confidence", "certain"),
		"websites": NewDetection("kind", "websites", "repo_name", "", "detail", "nothing"),
		"personal": NewDetection("kind", "personal", "repo_name", "per-repo", "confidence", "strong"),
	}
}

func TestApplyDetections_OnlyFilledSlotsLand(t *testing.T) {
	cfg := map[string]any{}
	applied := ApplyDetections(cfg, applyResults())
	if applied != 2 {
		t.Fatalf("applied: %d", applied)
	}
	if cfg["vaultseal"].(map[string]any)["repo_name"] != "gh-repo" {
		t.Errorf("vaultseal.repo_name: %v", cfg["vaultseal"])
	}
	if cfg["manual_repo_name"] != "per-repo" {
		t.Errorf("manual_repo_name: %v", cfg["manual_repo_name"])
	}
	if _, ok := cfg["website_repo_name"]; ok { // stayed EMPTY
		t.Error("website_repo_name should stay absent")
	}
}

func TestApplyDetections_OwnerConfidenceNeverOverwrites(t *testing.T) {
	cfg := map[string]any{"website_repo_name": "typed-by-owner"}
	results := applyResults()
	results["websites"] = NewDetection("kind", "websites", "repo_name", "detected",
		"confidence", "owner")
	applied := ApplyDetections(cfg, results)
	if applied != 2 { // github + personal only
		t.Fatalf("applied: %d", applied)
	}
	if cfg["website_repo_name"] != "typed-by-owner" {
		t.Errorf("website_repo_name: %v", cfg["website_repo_name"])
	}
}

func TestApplyDetections_PlainDictsFlowToo(t *testing.T) {
	// the worker payload shape (Detection IS a dict — but the plain
	// dict crossing the signal boundary must work just the same)
	cfg := map[string]any{}
	plain := map[string]any{
		"github":   map[string]any{"repo_name": "gh-repo", "confidence": "certain"},
		"websites": map[string]any{"repo_name": "", "confidence": ""},
		"personal": map[string]any{"repo_name": "p", "confidence": "strong"},
	}
	if applied := ApplyDetections(cfg, plain); applied != 2 {
		t.Fatalf("applied: %d", applied)
	}
	if cfg["vaultseal"].(map[string]any)["repo_name"] != "gh-repo" {
		t.Errorf("vaultseal.repo_name: %v", cfg["vaultseal"])
	}
	if cfg["manual_repo_name"] != "p" {
		t.Errorf("manual_repo_name: %v", cfg["manual_repo_name"])
	}
}

func TestApplyDetections_NestedSetKeepsNeighbours(t *testing.T) {
	cfg := map[string]any{"vaultseal": map[string]any{"enabled": true, "auto_push": false}}
	ApplyDetections(cfg, applyResults())
	nested := cfg["vaultseal"].(map[string]any)
	if nested["repo_name"] != "gh-repo" {
		t.Errorf("repo_name: %v", nested["repo_name"])
	}
	if nested["enabled"] != true { // kept
		t.Error("enabled was dropped")
	}
	if nested["auto_push"] != false { // kept
		t.Error("auto_push was dropped")
	}
}

func TestApplyDetections_Idempotent(t *testing.T) {
	cfg := map[string]any{}
	if applied := ApplyDetections(cfg, applyResults()); applied != 2 {
		t.Fatalf("applied: %d", applied)
	}
	if applied := ApplyDetections(cfg, applyResults()); applied != 0 {
		t.Fatalf("second applied: %d", applied)
	}
}

func TestApplyDetections_GarbageInputsNeverRaise(t *testing.T) {
	if got := ApplyDetections(map[string]any{}, nil); got != 0 {
		t.Errorf("nil results: %d", got)
	}
	if got := ApplyDetections(map[string]any{}, map[string]any{"github": "junk"}); got != 0 {
		t.Errorf("junk entry: %d", got)
	}
	if got := ApplyDetections(nil, map[string]any{}); got != 0 {
		t.Errorf("nil config: %d", got)
	}
}

func TestReportLines_OneLinePerVaultWithTheHonestVerbs(t *testing.T) {
	results := map[string]any{
		"github":   NewDetection("kind", "github", "repo_name", "gh", "confidence", "owner", "method", "config"),
		"websites": NewDetection("kind", "websites", "repo_name", "web", "confidence", "certain", "method", "name"),
		"personal": NewDetection("kind", "personal", "repo_name", "", "vault_path", "/p", "detail", "nothing conclusive"),
	}
	lines := ReportLines(results)
	if len(lines) != 3 {
		t.Fatalf("lines: %v", lines)
	}
	if !strings.Contains(lines[0], "gh (already set — kept)") {
		t.Errorf("line 0: %s", lines[0])
	}
	if !strings.Contains(lines[1], "web — detected via the vault folder's name") {
		t.Errorf("line 1: %s", lines[1])
	}
	if !strings.Contains(lines[2], "no backup repo found — nothing conclusive") {
		t.Errorf("line 2: %s", lines[2])
	}
}

func TestReportLines_PlainDictsFlowToo(t *testing.T) {
	lines := ReportLines(map[string]any{
		"github": map[string]any{"repo_name": "x", "confidence": "certain", "method": "remote"},
	})
	if len(lines) != 1 {
		t.Fatalf("lines: %v", lines)
	}
	if !strings.Contains(lines[0], "detected via its own git remote") {
		t.Errorf("line: %s", lines[0])
	}
}

func TestRadarPending_NoTokenIsNotPending(t *testing.T) {
	if RadarPending(map[string]any{
		"vault_path": "/v", "vaultseal": map[string]any{"repo_name": ""},
	}) {
		t.Error("pending without a token")
	}
}

func TestRadarPending_TokenPathEmptySlotIsPending(t *testing.T) {
	if !RadarPending(map[string]any{"github_token": "t", "vault_path": "/v"}) {
		t.Error("should be pending")
	}
}

func TestRadarPending_FullyWiredIsNotPending(t *testing.T) {
	cfg := map[string]any{
		"github_token": "t", "vault_path": "/v",
		"vaultseal":          map[string]any{"repo_name": "a"},
		"website_vault_path": "/w", "website_repo_name": "b",
		"manual_vault_path": "/p", "manual_repo_name": "c",
	}
	if RadarPending(cfg) {
		t.Error("fully wired should not be pending")
	}
}

func TestRadarPending_OneEmptySlotAmongThreeIsPending(t *testing.T) {
	cfg := map[string]any{
		"github_token": "t", "vault_path": "/v",
		"vaultseal":          map[string]any{"repo_name": "a"},
		"website_vault_path": "/w", "website_repo_name": "b",
		"manual_vault_path": "/p", // manual_repo_name empty
	}
	if !RadarPending(cfg) {
		t.Error("one empty slot should be pending")
	}
}

func TestRadarPending_PathsButEverySlotFilledAndNoPaths(t *testing.T) {
	if RadarPending(map[string]any{"github_token": "t"}) {
		t.Error("no paths should not be pending")
	}
	if RadarPending(map[string]any{}) {
		t.Error("empty config should not be pending")
	}
}

// ---------------------------------------------------------------------------
// 4. The read-only law (the creation guard stands)
// ---------------------------------------------------------------------------

func TestReadOnlyLaw_FakeAPISeesOnlyReadPaths(t *testing.T) {
	tmp := t.TempDir()
	vault := makeVault(t, tmp, "github-projects-obsidian-vault", map[string]string{"a.md": "x"})
	api := newFakeAPI(func(f *fakeRadarAPI) {
		f.exists["owner/github-projects-obsidian-vault"] =
			map[string]any{"name": "github-projects-obsidian-vault"}
	})
	cfg := map[string]any{"github_token": "t", "vault_path": vault}
	out := RadarAllVaults(cfg, api)
	if out["github"].RepoName() != "github-projects-obsidian-vault" {
		t.Errorf("repo_name: %s", out["github"].RepoName())
	}
	for _, path := range api.calls {
		if !strings.HasPrefix(path, "/user") && !strings.HasPrefix(path, "/repos/") {
			t.Errorf("non-read path requested: %s", path)
		}
	}
}

// ---------------------------------------------------------------------------
// 5. The third vault's slot — the constants contract
// ---------------------------------------------------------------------------

func TestManualVaultSlot_ConstantsDefaultIsEmpty(t *testing.T) {
	example := constants.ConfigExample()
	if example["manual_repo_name"] != "" {
		t.Errorf("manual_repo_name: %v", example["manual_repo_name"])
	}
	if example["website_repo_name"] != "" {
		t.Errorf("website_repo_name: %v", example["website_repo_name"])
	}
	if example["manual_vault_path"] != "" {
		t.Errorf("manual_vault_path: %v", example["manual_vault_path"])
	}
}

// ---------------------------------------------------------------------------
// 7. The end-to-end story: one wizard pass, zero typing
// ---------------------------------------------------------------------------

func TestTheOwnersExactScenario_PathsInReposOut(t *testing.T) {
	tmp := t.TempDir()
	ghVault := makeVault(t, tmp, "github-projects-obsidian-vault", map[string]string{"g.md": "github"})
	webVault := makeVault(t, tmp, "my-awesome-websites-directory", map[string]string{"w.md": "web"})
	perVault := makeVault(t, tmp, "Ali-personal-obsidian-vault", map[string]string{"p.md": "personal"})
	api := newFakeAPI(func(f *fakeRadarAPI) {
		f.exists["owner/github-projects-obsidian-vault"] =
			map[string]any{"name": "github-projects-obsidian-vault"}
		f.exists["owner/my-awesome-websites-directory"] =
			map[string]any{"name": "my-awesome-websites-directory"}
		f.exists["owner/Ali-personal-obsidian-vault"] =
			map[string]any{"name": "Ali-personal-obsidian-vault"}
	})
	// exactly what the wizard / startup radar build:
	cfg := map[string]any{
		"github_token":       "tok",
		"vault_path":         ghVault,
		"website_vault_path": webVault,
		"manual_vault_path":  perVault,
	}
	if !RadarPending(cfg) { // work to do
		t.Fatal("should be pending")
	}
	results := RadarAllVaults(cfg, api)
	resultsAny := map[string]any{}
	for k, v := range results {
		resultsAny[k] = v
	}
	applied := ApplyDetections(cfg, resultsAny)
	if applied != 3 { // all three, no typing
		t.Fatalf("applied: %d", applied)
	}
	if cfg["vaultseal"].(map[string]any)["repo_name"] != "github-projects-obsidian-vault" {
		t.Errorf("vaultseal.repo_name: %v", cfg["vaultseal"])
	}
	if cfg["website_repo_name"] != "my-awesome-websites-directory" {
		t.Errorf("website_repo_name: %v", cfg["website_repo_name"])
	}
	if cfg["manual_repo_name"] != "Ali-personal-obsidian-vault" {
		t.Errorf("manual_repo_name: %v", cfg["manual_repo_name"])
	}
	if RadarPending(cfg) { // and never wakes again
		t.Error("should not be pending after the fill")
	}
}

func TestTheOwnersExactScenario_FingerprintStoryWhenFoldersWereRenamed(t *testing.T) {
	// the vault folders were renamed locally; only the CONTENT
	// identifies the backups (path-independent blob SHAs)
	tmp := t.TempDir()
	files := map[string]string{}
	for i := 0; i < 5; i++ {
		files[fmt.Sprintf("notes/n%d.md", i)] = fmt.Sprintf("# n%d\n", i) + strings.Repeat("v", 40)
	}
	vault := makeVault(t, tmp, "renamed-folder", files)
	var shas []string
	keys := make([]string, 0, len(files))
	for k := range files {
		keys = append(keys, k)
	}
	sortStrings(keys)
	for _, k := range keys {
		shas = append(shas, gitBlob(files[k]))
	}
	api := newFakeAPI(func(f *fakeRadarAPI) {
		f.repos = []map[string]any{rp("github-projects-obsidian-vault"), rp("Ali-personal-obsidian-vault")}
		f.trees["owner/github-projects-obsidian-vault"] = shas[:5]
		f.trees["owner/Ali-personal-obsidian-vault"] = []string{gitBlob("different content")}
	})
	cfg := map[string]any{"github_token": "tok", "vault_path": vault}
	results := RadarAllVaults(cfg, api)
	if results["github"].RepoName() != "github-projects-obsidian-vault" {
		t.Errorf("repo_name: %s", results["github"].RepoName())
	}
	if results["github"].Method() != "fingerprint" {
		t.Errorf("method: %s", results["github"].Method())
	}
}
