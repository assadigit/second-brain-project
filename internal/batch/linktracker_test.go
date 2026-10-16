// Package batch, file linktracker_test.go — T-25 characterization
// (Phase 2): the LinkTracker half.
//
// Ported/landed suites:
//   - test_reconfix.py: TestReconciliation (the owner's 620-link
//     scenario, mixed batches, the old-shape behavior, moved notes,
//     lost notes, the no-leak law, idempotence) + TestVerifyHealing
//     (verify()'s ground-truth healing) + TestAllClearGroundTruth (the
//     all-clear gate's vault truth + the injected-index reuse).
//   - test_retrycry.py: TestReconciliationGithubOnly (website rows are
//     set aside, never cried; the github healing law untouched).
//   - test_reposettled.py: TestTheReconciliationGate (settled github
//     rows never cry — the settled probe goes through the
//     DEFAULT-anchored CacheDB, redirected via GITCURATOR_HOME) +
//     TestGetAllClear (a batch of settled skips is all-clear).
package batch

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/assadigit/second-brain-project/internal/store"
)

// writeManifest writes a manifest with the given rows (the suites'
// _manifest helper).
func writeManifest(t *testing.T, vault string, links []*ManifestLink) {
	t.Helper()
	if err := os.MkdirAll(vault, 0o777); err != nil {
		t.Fatal(err)
	}
	m := &Manifest{BatchID: "b1", Source: "bot",
		TotalLinks: len(links), Links: links}
	data, _ := json.Marshal(m)
	// Write through the same writer the tracker uses (the suites'
	// plain json.dump shape reads back identically).
	if err := os.WriteFile(filepath.Join(vault, "links_manifest.json"),
		[]byte(pyDumpManifest(m)), 0o644); err != nil {
		t.Fatal(err)
	}
	_ = data
}

func loadManifest(t *testing.T, vault string) *Manifest {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(vault, "links_manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	return &m
}

// row builds a manifest row (the suites' _row helper).
func row(url, status string, notePath *string, errMsg *string,
	kind string) *ManifestLink {
	return &ManifestLink{URL: url, Normalized: url, Type: kind,
		Status: status, NotePath: notePath, Error: errMsg}
}

func sp(s string) *string { return &s }

// ---------------------------------------------------------------------------
// test_reconfix.py: TestReconciliation
// ---------------------------------------------------------------------------

func TestReconciliationOwnersScenarioAllStored(t *testing.T) {
	// The screenshot: pending/failed/processing rows whose notes are
	// all in the vault → ZERO retries, every row healed + persisted.
	vault := filepath.Join(t.TempDir(), "vault")
	if err := os.MkdirAll(filepath.Join(vault, "AI"), 0o777); err != nil {
		t.Fatal(err)
	}
	var urls []string
	for i := 0; i < 620; i++ {
		urls = append(urls, "https://github.com/o/r"+itoa(i))
	}
	for _, u := range urls {
		writeNote(t, vault, "AI/"+filepath.Base(u)+".md", u)
	}
	var rows []*ManifestLink
	rows = append(rows,
		row(urls[0], "pending", nil, nil, "github"),
		row(urls[1], "processing", nil, nil, "github"),
		row(urls[2], "failed", nil, nil, "github"))
	for _, u := range urls[3:] {
		rows = append(rows, row(u, "pending", nil, nil, "github"))
	}
	writeManifest(t, vault, rows)
	tr := NewLinkTracker(vault)
	retry := tr.GetReconciliationURLs()
	if len(retry) != 0 {
		t.Fatalf("retry = %v; want zero", retry)
	}
	if tr.ReconciledVaultHits != 620 {
		t.Fatalf("reconciled_vault_hits = %d; want 620",
			tr.ReconciledVaultHits)
	}
	healed := loadManifest(t, vault)
	byURL := map[string]*ManifestLink{}
	for _, l := range healed.Links {
		byURL[l.URL] = l
	}
	for _, u := range urls {
		l := byURL[u]
		if l == nil {
			t.Fatalf("row %s vanished", u)
		}
		if l.Status != "skipped" {
			t.Fatalf("healed status = %q for %s", l.Status, u)
		}
		if l.Error == nil || *l.Error != "already in vault (reconciled)" {
			t.Fatalf("healed error = %v for %s", l.Error, u)
		}
		if l.NotePath == nil || *l.NotePath == "" {
			t.Fatalf("healed note_path empty for %s", u)
		}
		if l.ProcessedAt == nil || *l.ProcessedAt == "" {
			t.Fatalf("healed processed_at empty for %s", u)
		}
	}
}

func TestReconciliationMixedBatchReturnsOnlyTheAbsent(t *testing.T) {
	vault := filepath.Join(t.TempDir(), "vault")
	if err := os.MkdirAll(filepath.Join(vault, "AI"), 0o777); err != nil {
		t.Fatal(err)
	}
	writeNote(t, vault, "AI/stored.md", "https://github.com/o/stored")
	writeManifest(t, vault, []*ManifestLink{
		row("https://github.com/o/stored", "pending", nil, nil, "github"),
		row("https://github.com/o/missing", "pending", nil, nil, "github"),
	})
	retry := NewLinkTracker(vault).GetReconciliationURLs()
	if !reflect.DeepEqual(retry, []string{"https://github.com/o/missing"}) {
		t.Fatalf("retry = %v", retry)
	}
}

func TestReconciliationOldShapeManifestOnlyVerdict(t *testing.T) {
	// No vault notes → the v0.32 behavior is untouched.
	vault := filepath.Join(t.TempDir(), "vault")
	if err := os.MkdirAll(vault, 0o777); err != nil {
		t.Fatal(err)
	}
	writeManifest(t, vault, []*ManifestLink{
		row("https://github.com/a/one", "failed", nil, nil, "github"),
		row("https://github.com/b/two", "pending", nil, nil, "github"),
		row("https://github.com/c/three", "processing", nil, nil, "github"),
		row("https://github.com/d/four", "skipped", nil, nil, "github"),
	})
	tr := NewLinkTracker(vault)
	retry := tr.GetReconciliationURLs()
	sort.Strings(retry)
	want := []string{"https://github.com/a/one",
		"https://github.com/b/two", "https://github.com/c/three"}
	if !reflect.DeepEqual(retry, want) {
		t.Fatalf("retry = %v; want %v", retry, want)
	}
	if tr.ReconciledVaultHits != 0 {
		t.Fatalf("hits = %d", tr.ReconciledVaultHits)
	}
}

func TestReconciliationMovedNoteRepointedNotRetried(t *testing.T) {
	vault := filepath.Join(t.TempDir(), "vault")
	if err := os.MkdirAll(vault, 0o777); err != nil {
		t.Fatal(err)
	}
	newPath := writeNote(t, vault, "Moved/one.md", "https://github.com/a/one")
	stale := filepath.Join(vault, "AI", "one.md") // no longer there
	writeManifest(t, vault, []*ManifestLink{
		row("https://github.com/a/one", "processed", sp(stale), nil,
			"github"),
	})
	retry := NewLinkTracker(vault).GetReconciliationURLs()
	if len(retry) != 0 {
		t.Fatalf("retry = %v; want zero (a move is not a loss)", retry)
	}
	if got := loadManifest(t, vault).Links[0].NotePath; got == nil ||
		*got != newPath {
		t.Fatalf("note_path = %v; want %q", got, newPath)
	}
}

func TestReconciliationLostNoteStillRetried(t *testing.T) {
	vault := filepath.Join(t.TempDir(), "vault")
	if err := os.MkdirAll(vault, 0o777); err != nil {
		t.Fatal(err)
	}
	writeManifest(t, vault, []*ManifestLink{
		row("https://github.com/a/gone", "processed",
			sp(filepath.Join(vault, "AI", "gone.md")), nil, "github"),
	})
	retry := NewLinkTracker(vault).GetReconciliationURLs()
	if !reflect.DeepEqual(retry, []string{"https://github.com/a/gone"}) {
		t.Fatalf("retry = %v", retry)
	}
}

func TestReconciliationNoManifestIsEmpty(t *testing.T) {
	vault := t.TempDir()
	if got := NewLinkTracker(vault).GetReconciliationURLs(); got != nil {
		t.Fatalf("retry = %v; want empty", got)
	}
}

func TestReconciliationHealingNeverLeaksIntoTheLiveManifest(t *testing.T) {
	vault := filepath.Join(t.TempDir(), "vault")
	if err := os.MkdirAll(vault, 0o777); err != nil {
		t.Fatal(err)
	}
	writeNote(t, vault, "AI/one.md", "https://github.com/a/one")
	writeManifest(t, vault, []*ManifestLink{
		row("https://github.com/a/one", "pending", nil, nil, "github"),
	})
	tr := NewLinkTracker(vault)
	before, _ := json.Marshal(tr.Manifest)
	tr.GetReconciliationURLs()
	after, _ := json.Marshal(tr.Manifest)
	if string(before) != string(after) {
		t.Fatal("the live manifest mutated during reconciliation")
	}
	if len(tr.Manifest.Links) != 0 {
		t.Fatalf("live links = %d; want 0", len(tr.Manifest.Links))
	}
}

func TestReconciliationSecondReadIsStable(t *testing.T) {
	vault := filepath.Join(t.TempDir(), "vault")
	if err := os.MkdirAll(vault, 0o777); err != nil {
		t.Fatal(err)
	}
	writeNote(t, vault, "AI/one.md", "https://github.com/a/one")
	writeManifest(t, vault, []*ManifestLink{
		row("https://github.com/a/one", "pending", nil, nil, "github"),
	})
	if got := NewLinkTracker(vault).GetReconciliationURLs(); got != nil {
		t.Fatalf("first pass = %v", got)
	}
	tr2 := NewLinkTracker(vault)
	if got := tr2.GetReconciliationURLs(); got != nil {
		t.Fatalf("second pass = %v", got)
	}
	if tr2.ReconciledVaultHits != 0 {
		t.Fatalf("second hits = %d", tr2.ReconciledVaultHits)
	}
}

// ---------------------------------------------------------------------------
// test_reconfix.py: TestVerifyHealing
// ---------------------------------------------------------------------------

func trackerWithManifest(t *testing.T, vault string,
	links []*ManifestLink) *LinkTracker {
	t.Helper()
	tr := NewLinkTracker(vault)
	tr.Manifest = &Manifest{BatchID: "b", Source: "bot",
		TotalLinks: len(links), Links: links}
	return tr
}

func TestVerifyPendingInVaultCountsSkipped(t *testing.T) {
	vault := filepath.Join(t.TempDir(), "vault")
	if err := os.MkdirAll(vault, 0o777); err != nil {
		t.Fatal(err)
	}
	writeNote(t, vault, "AI/one.md", "https://github.com/a/one")
	tr := trackerWithManifest(t, vault, []*ManifestLink{
		row("https://github.com/a/one", "pending", nil, nil, "github"),
		row("https://github.com/z/new", "pending", nil, nil, "github"),
	})
	rep := tr.Verify(nil, nil)
	if rep.GithubSkipped != 1 {
		t.Fatalf("github_skipped = %d; want 1", rep.GithubSkipped)
	}
	if rep.GithubPending != 1 {
		t.Fatalf("github_pending = %d; want 1 (z/new is real work)",
			rep.GithubPending)
	}
	if rep.VerificationPassed {
		t.Fatal("verification passed with real work pending")
	}
	if !rep.AccountingOK {
		t.Fatal("accounting broke")
	}
	l0 := tr.Manifest.Links[0]
	if l0.Status != "skipped" {
		t.Fatalf("healed status = %q", l0.Status)
	}
	if l0.NotePath == nil ||
		!stringsHasSuffix(filepath.ToSlash(*l0.NotePath), "one.md") {
		t.Fatalf("healed note_path = %v", l0.NotePath)
	}
}

func TestVerifyProcessedStalePathRepointed(t *testing.T) {
	vault := filepath.Join(t.TempDir(), "vault")
	if err := os.MkdirAll(vault, 0o777); err != nil {
		t.Fatal(err)
	}
	newPath := writeNote(t, vault, "Moved/one.md", "https://github.com/a/one")
	tr := trackerWithManifest(t, vault, []*ManifestLink{
		row("https://github.com/a/one", "processed",
			sp(filepath.Join(vault, "AI", "one.md")), nil, "github"),
	})
	rep := tr.Verify(nil, nil)
	if rep.GithubProcessed != 1 || rep.GithubFailed != 0 {
		t.Fatalf("processed/failed = %d/%d", rep.GithubProcessed,
			rep.GithubFailed)
	}
	if !rep.VerificationPassed {
		t.Fatal("verification failed on a moved note")
	}
	if got := tr.Manifest.Links[0].NotePath; got == nil || *got != newPath {
		t.Fatalf("note_path = %v; want %q", got, newPath)
	}
}

func TestVerifyProcessedLostNoteStillFails(t *testing.T) {
	vault := t.TempDir()
	if err := os.MkdirAll(vault, 0o777); err != nil {
		t.Fatal(err)
	}
	tr := trackerWithManifest(t, vault, []*ManifestLink{
		row("https://github.com/a/gone", "processed",
			sp(filepath.Join(vault, "AI", "gone.md")), nil, "github"),
	})
	rep := tr.Verify(nil, nil)
	if rep.GithubFailed != 1 {
		t.Fatalf("github_failed = %d; want 1", rep.GithubFailed)
	}
	if rep.VerificationPassed {
		t.Fatal("verification passed on a lost note")
	}
	if got := rep.FailedLinks[0].Error; got == nil ||
		*got != "note file missing" {
		t.Fatalf("error = %v; want 'note file missing'", got)
	}
}

func TestVerifyProcessedValidNoteUnchanged(t *testing.T) {
	vault := t.TempDir()
	if err := os.MkdirAll(vault, 0o777); err != nil {
		t.Fatal(err)
	}
	p := writeNote(t, vault, "AI/one.md", "https://github.com/a/one")
	tr := trackerWithManifest(t, vault, []*ManifestLink{
		row("https://github.com/a/one", "processed", sp(p), nil, "github"),
	})
	rep := tr.Verify(nil, nil)
	if rep.GithubProcessed != 1 || !rep.VerificationPassed {
		t.Fatalf("report = %+v", rep)
	}
}

// ---------------------------------------------------------------------------
// test_reconfix.py: TestAllClearGroundTruth
// ---------------------------------------------------------------------------

func TestAllClearFailedInVaultDoesNotBlock(t *testing.T) {
	vault := t.TempDir()
	if err := os.MkdirAll(vault, 0o777); err != nil {
		t.Fatal(err)
	}
	writeNote(t, vault, "AI/one.md", "https://github.com/a/one")
	tr := trackerWithManifest(t, vault, []*ManifestLink{
		row("https://github.com/a/one", "failed", nil, nil, "github"),
	})
	if !tr.GetAllClear() {
		t.Fatal("a stored note blocked the all-clear")
	}
}

func TestAllClearFailedAbsentStillBlocks(t *testing.T) {
	vault := t.TempDir()
	if err := os.MkdirAll(vault, 0o777); err != nil {
		t.Fatal(err)
	}
	tr := trackerWithManifest(t, vault, []*ManifestLink{
		row("https://github.com/a/gone", "failed", nil, nil, "github"),
	})
	if tr.GetAllClear() {
		t.Fatal("an absent note did not block")
	}
}

func TestAllClearProcessingInVaultDoesNotBlock(t *testing.T) {
	vault := t.TempDir()
	if err := os.MkdirAll(vault, 0o777); err != nil {
		t.Fatal(err)
	}
	writeNote(t, vault, "AI/one.md", "https://github.com/a/one")
	tr := trackerWithManifest(t, vault, []*ManifestLink{
		row("https://github.com/a/one", "processing", nil, nil, "github"),
	})
	if !tr.GetAllClear() {
		t.Fatal("a stored note blocked the all-clear (processing)")
	}
}

func TestAllClearPendingGithubInVaultDoesNotBlock(t *testing.T) {
	vault := t.TempDir()
	if err := os.MkdirAll(vault, 0o777); err != nil {
		t.Fatal(err)
	}
	writeNote(t, vault, "AI/one.md", "https://github.com/a/one")
	tr := trackerWithManifest(t, vault, []*ManifestLink{
		row("https://github.com/a/one", "pending", nil, nil, "github"),
	})
	if !tr.GetAllClear() {
		t.Fatal("a stored note blocked the all-clear (pending)")
	}
}

func TestAllClearPendingGithubAbsentStillBlocks(t *testing.T) {
	vault := t.TempDir()
	if err := os.MkdirAll(vault, 0o777); err != nil {
		t.Fatal(err)
	}
	tr := trackerWithManifest(t, vault, []*ManifestLink{
		row("https://github.com/a/gone", "pending", nil, nil, "github"),
	})
	if tr.GetAllClear() {
		t.Fatal("an absent pending note did not block")
	}
}

func TestAllClearPendingNonGithubNeverBlocks(t *testing.T) {
	// The v29.4 rule: non-GitHub pending is a duplicate already
	// recorded elsewhere (websites pipeline semantics).
	vault := t.TempDir()
	if err := os.MkdirAll(vault, 0o777); err != nil {
		t.Fatal(err)
	}
	tr := trackerWithManifest(t, vault, []*ManifestLink{
		row("https://example.com/x", "pending", nil, nil, "non-github"),
	})
	if !tr.GetAllClear() {
		t.Fatal("a non-github pending row blocked")
	}
}

func TestAllClearInjectedIndexIsReusedNotRebuilt(t *testing.T) {
	vault := t.TempDir()
	if err := os.MkdirAll(vault, 0o777); err != nil {
		t.Fatal(err)
	}
	writeNote(t, vault, "AI/one.md", "https://github.com/a/one")
	idx := NewVaultIndex(vault, nil)
	idx.Rebuild(nil)
	tr := NewLinkTracker(vault)
	tr.SetVaultIndex(idx)
	if tr.groundTruth() != idx {
		t.Fatal("the injected index was not reused")
	}
	tr.Manifest = &Manifest{BatchID: "b", Source: "bot", TotalLinks: 1,
		Links: []*ManifestLink{
			row("https://github.com/a/one", "failed", nil, nil, "github")}}
	if !tr.GetAllClear() {
		t.Fatal("the injected index did not answer")
	}
	if tr.groundTruth() != idx {
		t.Fatal("a second ground-truth read rebuilt the index")
	}
}

func TestVaultHasPublicFace(t *testing.T) {
	vault := t.TempDir()
	if err := os.MkdirAll(vault, 0o777); err != nil {
		t.Fatal(err)
	}
	writeNote(t, vault, "AI/one.md", "https://github.com/a/one")
	tr := NewLinkTracker(vault)
	if !tr.VaultHas("https://github.com/a/one") {
		t.Fatal("vault_has(stored) = false")
	}
	if tr.VaultHas("https://github.com/a/gone") {
		t.Fatal("vault_has(absent) = true")
	}
}

// ---------------------------------------------------------------------------
// test_retrycry.py: TestReconciliationGithubOnly
// ---------------------------------------------------------------------------

func TestWebsiteRowsAreSetAsideNeverCried(t *testing.T) {
	vault := filepath.Join(t.TempDir(), "vault")
	if err := os.MkdirAll(vault, 0o777); err != nil {
		t.Fatal(err)
	}
	writeManifest(t, vault, []*ManifestLink{
		row("https://example.com/walled", "failed", nil,
			sp("403 wall"), "non-github"),
		row("https://example.com/stored2", "processed", sp(""),
			nil, "non-github"),
	})
	tr := NewLinkTracker(vault)
	if got := tr.GetReconciliationURLs(); got != nil {
		t.Fatalf("retry = %v; want empty", got)
	}
	if tr.SetAsideWebsites != 2 {
		t.Fatalf("set_aside_websites = %d; want 2",
			tr.SetAsideWebsites)
	}
}

func TestStoredWebsiteWithValidPathNeedsNoAside(t *testing.T) {
	dir := t.TempDir()
	vault := filepath.Join(dir, "vault")
	if err := os.MkdirAll(vault, 0o777); err != nil {
		t.Fatal(err)
	}
	wsNote := writeNote(t, filepath.Join(dir, "ws"), "Design/site.md",
		"https://example.com/stored")
	writeManifest(t, vault, []*ManifestLink{
		row("https://example.com/stored", "processed", sp(wsNote), nil,
			"non-github"),
	})
	tr := NewLinkTracker(vault)
	if got := tr.GetReconciliationURLs(); got != nil {
		t.Fatalf("retry = %v", got)
	}
	if tr.SetAsideWebsites != 0 {
		t.Fatalf("set_aside_websites = %d; want 0",
			tr.SetAsideWebsites)
	}
}

func TestGithubLawUntouchedHealing(t *testing.T) {
	vault := filepath.Join(t.TempDir(), "vault")
	if err := os.MkdirAll(vault, 0o777); err != nil {
		t.Fatal(err)
	}
	writeNote(t, vault, "AI/one.md", "https://github.com/a/one")
	writeManifest(t, vault, []*ManifestLink{
		row("https://github.com/a/one", "pending", nil, nil, "github"),
		row("https://github.com/a/gone", "pending", nil, nil, "github"),
	})
	tr := NewLinkTracker(vault)
	retry := tr.GetReconciliationURLs()
	if !reflect.DeepEqual(retry, []string{"https://github.com/a/gone"}) {
		t.Fatalf("retry = %v", retry)
	}
	if tr.ReconciledVaultHits != 1 {
		t.Fatalf("hits = %d; want 1", tr.ReconciledVaultHits)
	}
}

// ---------------------------------------------------------------------------
// test_reposettled.py: TestTheReconciliationGate + TestGetAllClear
// ---------------------------------------------------------------------------

// newSettledGate redirects the tracker's settled probe (the
// DEFAULT-anchored CacheDB) into a temp APP_DIR via GITCURATOR_HOME,
// runs THE SETTLEMENT over the seeded repos, and hands back the vault.
func newSettledGate(t *testing.T) (vault string) {
	t.Helper()
	dir := t.TempDir()
	vault = filepath.Join(dir, "ghvault")
	if err := os.MkdirAll(vault, 0o777); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GITCURATOR_HOME", dir)
	cache, err := store.NewCacheDB("cache.db")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { cache.Close() })
	// THE SETTLEMENT has run on this machine
	if _, err := cache.SettleExistingRepos([]string{
		"https://github.com/o/settled-failed",
		"https://github.com/o/settled-pending",
	}); err != nil {
		t.Fatal(err)
	}
	return vault
}

func TestSettledFailedRowsNeverCry(t *testing.T) {
	vault := newSettledGate(t)
	writeManifest(t, vault, []*ManifestLink{
		row("https://github.com/o/settled-failed", "failed", nil,
			sp("boom"), "github"),
		row("https://github.com/o/missing", "failed", nil, sp("boom"),
			"github"),
	})
	tr := NewLinkTracker(vault)
	urls := tr.GetReconciliationURLs()
	if !reflect.DeepEqual(urls, []string{"https://github.com/o/missing"}) {
		t.Fatalf("urls = %v", urls)
	}
	if tr.SetAsideSettled != 1 {
		t.Fatalf("set_aside_settled = %d; want 1", tr.SetAsideSettled)
	}
}

func TestSettledPendingRowsNeverCry(t *testing.T) {
	vault := newSettledGate(t)
	writeManifest(t, vault, []*ManifestLink{
		row("https://github.com/o/settled-pending", "pending", nil, nil,
			"github"),
	})
	tr := NewLinkTracker(vault)
	if got := tr.GetReconciliationURLs(); got != nil {
		t.Fatalf("urls = %v; want empty", got)
	}
	if tr.SetAsideSettled != 1 {
		t.Fatalf("set_aside_settled = %d; want 1", tr.SetAsideSettled)
	}
}

func TestUnsettledRowsKeepTheOldBehavior(t *testing.T) {
	vault := newSettledGate(t)
	writeManifest(t, vault, []*ManifestLink{
		row("https://github.com/o/new-failure", "failed", nil,
			sp("rate limit"), "github"),
	})
	tr := NewLinkTracker(vault)
	if got := tr.GetReconciliationURLs(); !reflect.DeepEqual(got,
		[]string{"https://github.com/o/new-failure"}) {
		t.Fatalf("urls = %v", got)
	}
	if tr.SetAsideSettled != 0 {
		t.Fatalf("set_aside_settled = %d", tr.SetAsideSettled)
	}
}

func TestVaultHealingLawUntouched(t *testing.T) {
	// The v0.38.0 healing law outranks the count either way: a failed
	// row whose note IS in the vault is healed, never retried — settled
	// or not.
	vault := newSettledGate(t)
	writeNote(t, vault, "AI/stored.md", "https://github.com/o/stored")
	writeManifest(t, vault, []*ManifestLink{
		row("https://github.com/o/stored", "failed", nil, sp("old"),
			"github"),
	})
	tr := NewLinkTracker(vault)
	if got := tr.GetReconciliationURLs(); got != nil {
		t.Fatalf("urls = %v", got)
	}
	if tr.ReconciledVaultHits != 1 {
		t.Fatalf("hits = %d; want 1", tr.ReconciledVaultHits)
	}
	if tr.SetAsideSettled != 0 {
		t.Fatalf("set_aside_settled = %d", tr.SetAsideSettled)
	}
}

func TestWebsiteRowsStaySetAside(t *testing.T) {
	// The v0.63.1 law survives this one untouched (regression guard).
	vault := newSettledGate(t)
	writeManifest(t, vault, []*ManifestLink{
		row("https://example.com/walled", "failed", nil, sp("403"),
			"non-github"),
		row("https://github.com/o/settled-failed", "failed", nil,
			sp("boom"), "github"),
	})
	tr := NewLinkTracker(vault)
	if got := tr.GetReconciliationURLs(); got != nil {
		t.Fatalf("urls = %v", got)
	}
	if tr.SetAsideWebsites != 1 {
		t.Fatalf("set_aside_websites = %d; want 1", tr.SetAsideWebsites)
	}
	if tr.SetAsideSettled != 1 {
		t.Fatalf("set_aside_settled = %d; want 1", tr.SetAsideSettled)
	}
}

func TestABrokenCacheNeverResurrectsACry(t *testing.T) {
	// The probe is guarded: with cache.db unreadable the read degrades
	// to the OLD behavior (the manifest's truth), never raises and
	// never hides a settled row it cannot see.
	vault := newSettledGate(t)
	writeManifest(t, vault, []*ManifestLink{
		row("https://github.com/o/settled-failed", "failed", nil,
			sp("boom"), "github"),
	})
	// A FRESH empty cache dir: nothing settled there, the old loud
	// behavior holds (a broken probe never lies quiet).
	t.Setenv("GITCURATOR_HOME", filepath.Join(t.TempDir(),
		"nonexistent-dir"))
	tr := NewLinkTracker(vault)
	urls := tr.GetReconciliationURLs()
	if !reflect.DeepEqual(urls, []string{
		"https://github.com/o/settled-failed"}) {
		t.Fatalf("urls = %v", urls)
	}
}

func TestABatchOfSettledSkipsIsAllClear(t *testing.T) {
	vault := filepath.Join(t.TempDir(), "ghvault")
	if err := os.MkdirAll(vault, 0o777); err != nil {
		t.Fatal(err)
	}
	writeManifest(t, vault, []*ManifestLink{
		row("https://github.com/o/settled-one", "skipped", nil, nil,
			"github"),
		row("https://github.com/o/settled-two", "skipped", nil, nil,
			"github"),
	})
	tr := NewLinkTracker(vault)
	tr.Manifest = tr.LoadPreviousManifest()
	if !tr.GetAllClear() {
		t.Fatal("a batch of settled skips is not all-clear")
	}
}

func TestAGenuinelyFailedRowStillBlocks(t *testing.T) {
	vault := filepath.Join(t.TempDir(), "ghvault")
	if err := os.MkdirAll(vault, 0o777); err != nil {
		t.Fatal(err)
	}
	writeManifest(t, vault, []*ManifestLink{
		row("https://github.com/o/new-failure", "failed", nil,
			sp("boom"), "github"),
	})
	tr := NewLinkTracker(vault)
	tr.Manifest = tr.LoadPreviousManifest()
	if tr.GetAllClear() {
		t.Fatal("a genuinely failed row did not block")
	}
}

// ---------------------------------------------------------------------------
// Small helpers.
// ---------------------------------------------------------------------------

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}

func stringsHasSuffix(s, suf string) bool {
	return len(s) >= len(suf) && s[len(s)-len(suf):] == suf
}
