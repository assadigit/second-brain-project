// app_test.go — the DI-root production-seam tests (T-71 slice (d)):
// the wiring proof (every seam non-nil, the headless law), the queue
// classification over a fixture vault, the backup zip's laws, the
// scan job's paths over fixture vaults (the clean library, the gate
// round-trip declined + confirmed), and the batch smoke over the
// engine fakes (the runner composition wired bus-ward).
package app

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/assadigit/second-brain-project/internal/githubpipeline"
	"github.com/assadigit/second-brain-project/internal/llmclient"
	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/store"
)

func TestNewWiresTheProductionSeams(t *testing.T) {
	a := New(filepath.Join(t.TempDir(), "missing-config.json"), false)
	if a.Bus == nil || a.Commands == nil || a.Lifecycle == nil {
		t.Fatal("the composition is missing a root part")
	}
	if a.Bus.Headless {
		t.Fatal("a GUI launch must not be headless")
	}
	d := a.Commands.Deps
	if d.Config == nil || d.QueueFetch == nil || d.QueueClassify == nil ||
		d.Batch == nil || d.DriftSeal == nil || d.Scan == nil ||
		d.ApplyPlan == nil || d.SealVault == nil || d.Resync == nil ||
		d.BackupNow == nil || d.Publish == nil || d.Radar == nil {
		t.Fatal("a production seam is unwired")
	}
	// The headless launch flips the bus law.
	h := New(filepath.Join(t.TempDir(), "missing.json"), true)
	if !h.Bus.Headless {
		t.Fatal("the headless flag never reached the bus")
	}
	// A missing config file loads as the empty config.
	cfg := a.loadConfig()
	if cfg == nil || len(cfg.Keys()) != 0 {
		t.Fatalf("missing config = %v, want empty", cfg)
	}
}

// ---------------------------------------------------------------------------
// The queue classification over a fixture vault.
// ---------------------------------------------------------------------------

func writeNote(t *testing.T, vault, rel, url string) string {
	t.Helper()
	path := filepath.Join(vault, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	name := url
	if i := strings.LastIndex(url, "/"); i >= 0 {
		name = url[i+1:]
	}
	text := fmt.Sprintf(`---
source: %s
aliases: []
tags: [cli]
category: "Agents"
date_processed: 2026-10-09
managed_by: "gitcurator"
---
# %s

A curated fixture note with enough body words for every gate.
`, url, name)
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestQueueClassificationOverFixtureVault(t *testing.T) {
	appDir := t.TempDir()
	t.Setenv("GITCURATOR_HOME", appDir)
	vault := t.TempDir()
	writeNote(t, vault, "Agents/one.md", "https://github.com/octo/one")

	cfg := storage.NewOrderedMap()
	cfg.Set("vault_path", vault)

	// Prime the settlement (the one-time law: the first
	// SettleExistingRepos over a FRESH cache would settle the whole
	// queue history — the golden harness's own priming step).
	{
		c, err := store.Open(filepath.Join(appDir, "cache.db"))
		if err != nil {
			t.Fatalf("prime store: %v", err)
		}
		_ = c.Close()
	}
	cache, cerr := store.NewCacheDB(filepath.Join(appDir, "cache.db"))
	if cerr != nil {
		t.Fatal(cerr)
	}
	if _, err := cache.SettleExistingRepos(nil); err != nil {
		t.Fatal(err)
	}
	_ = cache.Close()

	items := []any{
		map[string]any{"url": "https://github.com/octo/one"},   // in vault
		map[string]any{"url": "https://github.com/octo/fresh"}, // pending
		map[string]any{"url": "https://example.com/page"},      // non-GitHub
	}
	var logs []string
	rep := ClassifyQueueLinks(cfg, items, func(msg, level string) {
		logs = append(logs, msg)
	})
	if len(rep.GitHubURLs) != 2 || len(rep.NonGithub) != 1 {
		t.Fatalf("routing = github %v nonGithub %v", rep.GitHubURLs, rep.NonGithub)
	}
	if rep.InVault != 1 {
		t.Fatalf("in-vault = %d, want 1", rep.InVault)
	}
	if len(rep.PendingURLs) != 1 || rep.PendingURLs[0] != "https://github.com/octo/fresh" {
		t.Fatalf("pending = %v", rep.PendingURLs)
	}
	if rep.VaultIndex == nil || *rep.VaultIndex != 1 {
		t.Fatalf("vault index count = %v", rep.VaultIndex)
	}
	if len(rep.PendingWebsites) != 0 || !rep.WebsitesOff {
		t.Fatalf("websites pipeline off: pending=%v off=%v", rep.PendingWebsites, rep.WebsitesOff)
	}
	// The raw count speaks the fetched items.
	if rep.RawCount != 3 {
		t.Fatalf("raw count = %d", rep.RawCount)
	}
}

func TestQueueClassificationNoVaultShowsAll(t *testing.T) {
	cfg := storage.NewOrderedMap() // no vault_path
	items := []any{map[string]any{"url": "https://github.com/octo/x"}}
	rep := ClassifyQueueLinks(cfg, items, nil)
	if len(rep.PendingURLs) != 1 {
		t.Fatalf("no-vault pending = %v, want all", rep.PendingURLs)
	}
	if rep.VaultIndex != nil {
		t.Fatal("a no-vault classify still built an index")
	}
}

// ---------------------------------------------------------------------------
// The backup zip's laws.
// ---------------------------------------------------------------------------

func TestBackupNowZipLaws(t *testing.T) {
	vault := t.TempDir()
	backup := filepath.Join(t.TempDir(), "backups")
	cfg := storage.NewOrderedMap()
	cfg.Set("vault_path", vault)
	cfg.Set("backup_folder", backup)
	cfg.Set("backup_max", 2)

	// The vault: a real note, an excluded config, an excluded lock,
	// an excluded .obsidian workspace file.
	writeNote(t, vault, "Agents/keep.md", "https://github.com/octo/keep")
	_ = os.WriteFile(filepath.Join(vault, "config.json"), []byte("x"), 0o644)
	_ = os.WriteFile(filepath.Join(vault, "note.lock"), []byte("x"), 0o644)
	_ = os.MkdirAll(filepath.Join(vault, ".obsidian"), 0o755)
	_ = os.WriteFile(filepath.Join(vault, ".obsidian", "workspace.json"), []byte("x"), 0o644)
	_ = os.MkdirAll(filepath.Join(vault, ".git"), 0o755)
	_ = os.WriteFile(filepath.Join(vault, ".git", "HEAD"), []byte("x"), 0o644)

	res := RunBackupNow(cfg, nil)
	if !res.Success {
		t.Fatalf("backup = %+v", res)
	}
	if !strings.Contains(res.Message, "vault-") || !strings.Contains(res.Message, "(1 files,") {
		t.Fatalf("verdict = %q, want the file + count shape", res.Message)
	}
	// Inspect the zip: exactly the note, rooted at the vault's parent.
	pile, _ := filepath.Glob(filepath.Join(backup, "vault-*.zip"))
	if len(pile) != 1 {
		t.Fatalf("pile = %v", pile)
	}
	raw, err := os.ReadFile(pile[0])
	if err != nil {
		t.Fatal(err)
	}
	names := zipNames(t, raw)
	if len(names) != 1 || !strings.HasSuffix(names[0], filepath.Join(filepath.Base(vault), "Agents", "keep.md")) {
		t.Fatalf("zip contents = %v, want the one note rooted at the vault folder", names)
	}

	// FIFO rotation: three backups, backup_max=2 → only 2 remain.
	time.Sleep(1100 * time.Millisecond) // the timestamp's 1-second face
	_ = RunBackupNow(cfg, nil)
	time.Sleep(1100 * time.Millisecond)
	_ = RunBackupNow(cfg, nil)
	pile, _ = filepath.Glob(filepath.Join(backup, "vault-*.zip"))
	if len(pile) != 2 {
		t.Fatalf("rotation left %d backups, want 2 (FIFO over backup_max)", len(pile))
	}
}

func zipNames(t *testing.T, raw []byte) []string {
	t.Helper()
	tf := filepath.Join(t.TempDir(), "probe.zip")
	if err := os.WriteFile(tf, raw, 0o644); err != nil {
		t.Fatal(err)
	}
	return zipFileNames(t, tf)
}

// zipFileNames lists the archive's entry names (archive/zip).
func zipFileNames(t *testing.T, path string) []string {
	t.Helper()
	zr, err := zip.OpenReader(path)
	if err != nil {
		t.Fatal(err)
	}
	defer zr.Close()
	var names []string
	for _, f := range zr.File {
		names = append(names, f.Name)
	}
	return names
}

func TestBackupNowGuards(t *testing.T) {
	cfg := storage.NewOrderedMap()
	if r := RunBackupNow(cfg, nil); r.Success || !strings.Contains(r.Message, "vault path") {
		t.Fatalf("no-vault guard = %+v", r)
	}
	cfg.Set("vault_path", t.TempDir())
	if r := RunBackupNow(cfg, nil); r.Success || !strings.Contains(r.Message, "backup folder") {
		t.Fatalf("no-folder guard = %+v", r)
	}
}

// ---------------------------------------------------------------------------
// The scan job's paths.
// ---------------------------------------------------------------------------

func TestScanJobNoVaultsHonestMiss(t *testing.T) {
	appDir := t.TempDir()
	t.Setenv("GITCURATOR_HOME", appDir)
	cfg := storage.NewOrderedMap() // no vaults
	var logs []string
	res := RunVaultScan(cfg, nil, ScanEvents{
		Log:      func(msg, level string) { logs = append(logs, msg+"|"+level) },
		Progress: func(string, int, int) {},
		Confirm: func(any) ScanAnswer {
			t.Fatal("the gate asked with no vaults")
			return ScanAnswer{Verdict: ScanTimeout}
		},
	})
	if !res.Success || res.Verdict != "clean" {
		t.Fatalf("miss story = %+v", res)
	}
	joined := strings.Join(logs, "\n")
	if !strings.Contains(joined, "no Websites vault and no GitHub Repos vault are set") {
		t.Fatalf("the miss never spoke: %v", logs)
	}
}

func TestScanJobCleanLibrary(t *testing.T) {
	appDir := t.TempDir()
	t.Setenv("GITCURATOR_HOME", appDir)
	vault := t.TempDir()
	writeNote(t, vault, "Agents/clean.md", "https://github.com/octo/clean")
	cfg := storage.NewOrderedMap()
	cfg.Set("vault_path", vault)

	var logs []string
	res := RunVaultScan(cfg, []string{"repos"}, ScanEvents{
		Log:      func(msg, level string) { logs = append(logs, msg+"|"+level) },
		Progress: func(string, int, int) {},
		Confirm: func(any) ScanAnswer {
			t.Fatal("the gate asked on a clean library")
			return ScanAnswer{Verdict: ScanTimeout}
		},
	})
	if !res.Success || res.Verdict != "clean" {
		t.Fatalf("clean story = %+v (logs: %v)", res, logs)
	}
	if !res.ReposRan || res.WebsitesRan {
		t.Fatalf("scope ran websites=%v repos=%v, want repos alone", res.WebsitesRan, res.ReposRan)
	}
	joined := strings.Join(logs, "\n")
	if !strings.Contains(joined, "the library is clean") {
		t.Fatalf("the clean line never spoke: %v", logs)
	}
	if !strings.Contains(joined, "the vault holds 1 note(s) in 1 folder(s)") {
		t.Fatalf("the inventory line never spoke: %v", logs)
	}
}

func TestScanJobGateDeclineKeepsEverything(t *testing.T) {
	appDir := t.TempDir()
	t.Setenv("GITCURATOR_HOME", appDir)
	vault := t.TempDir()
	marked := writeMarkedNote(t, vault, "Agents/doomed.md", "https://github.com/octo/doomed")
	cfg := storage.NewOrderedMap()
	cfg.Set("vault_path", vault)

	var asked int
	res := RunVaultScan(cfg, []string{"repos"}, ScanEvents{
		Log:      func(string, string) {},
		Progress: func(string, int, int) {},
		Confirm: func(plan any) ScanAnswer {
			asked++
			return ScanAnswer{Verdict: ScanDeclined}
		},
	})
	if asked != 1 {
		t.Fatalf("the gate ran %d times, want exactly once", asked)
	}
	if res.Verdict != "declined" {
		t.Fatalf("verdict = %q, want declined", res.Verdict)
	}
	if _, err := os.Stat(marked); err != nil {
		t.Fatalf("a declined plan still touched the note: %v", err)
	}
	if _, err := os.Stat(filepath.Join(vault, ".trash")); err == nil {
		t.Fatal("a declined plan still built a trash pile")
	}
}

func TestScanJobGateConfirmAppliesThePlan(t *testing.T) {
	appDir := t.TempDir()
	t.Setenv("GITCURATOR_HOME", appDir)
	vault := t.TempDir()
	marked := writeMarkedNote(t, vault, "Agents/doomed.md", "https://github.com/octo/doomed")
	cfg := storage.NewOrderedMap()
	cfg.Set("vault_path", vault)

	res := RunVaultScan(cfg, []string{"repos"}, ScanEvents{
		Log:      func(string, string) {},
		Progress: func(string, int, int) {},
		Confirm: func(plan any) ScanAnswer {
			return ScanAnswer{Verdict: ScanConfirmed}
		},
	})
	if res.Verdict != "confirmed" {
		t.Fatalf("verdict = %q, want confirmed", res.Verdict)
	}
	if _, err := os.Stat(marked); !os.IsNotExist(err) {
		t.Fatalf("a confirmed plan never removed the marked note: %v", err)
	}
	// The banishment machinery's own door: .trash/banished.
	banished := filepath.Join(vault, ".trash", "banished")
	if _, err := os.Stat(banished); err != nil {
		t.Fatalf("the banished pile never landed: %v", err)
	}
}

// writeMarkedNote writes an app-owned repo note carrying the
// #auto-delete body tag (the v0.60.1 grammar).
func writeMarkedNote(t *testing.T, vault, rel, url string) string {
	t.Helper()
	path := filepath.Join(vault, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	name := url
	if i := strings.LastIndex(url, "/"); i >= 0 {
		name = url[i+1:]
	}
	text := fmt.Sprintf(`---
source: %s
aliases: []
tags: [cli]
category: "Agents"
date_processed: 2026-10-09
managed_by: "gitcurator"
---
# %s

A curated fixture note whose body carries the owner's delete mark #auto-delete and enough words for every gate.

---
*Source: [GitHub](%s)*
`, url, name, url)
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// ---------------------------------------------------------------------------
// The batch smoke over the engine fakes (the DI wiring's proof).
// ---------------------------------------------------------------------------

type appFakeFetcher struct{}

func (f *appFakeFetcher) GetRateLimit() (*githubpipeline.RateLimit, error) {
	return nil, errors.New("no rate limit on the fake")
}

func (f *appFakeFetcher) GetRepo(fullName string) (*githubpipeline.Repo, error) {
	if fullName != "octo-owner/curated-repo" {
		return nil, &githubpipeline.GHError{Status: 404, Message: "404"}
	}
	return &githubpipeline.Repo{
		ID: 4242, FullName: "octo-owner/curated-repo", Name: "curated-repo",
		Fork: false, StargazersCount: 12500, ForksCount: 12,
		Description:       "A curated offline fixture repo.",
		OwnerLogin:        "octo-owner",
		OrganizationLogin: "",
		Language:          "Go",
	}, nil
}

func (f *appFakeFetcher) GetTopics(string) ([]string, error) {
	return []string{"go", "testing"}, nil
}

func (f *appFakeFetcher) CountCommitsSince(string, time.Time) (int, error) {
	return 42, nil
}

func (f *appFakeFetcher) GetReadme(string) (*githubpipeline.ReadmeResult, error) {
	return &githubpipeline.ReadmeResult{
		Content:  "A fixture readme with enough words for the summary gate.",
		RawChars: 56}, nil
}

func (f *appFakeFetcher) GetLanguages(string) ([]string, error) {
	return []string{"Go"}, nil
}

func (f *appFakeFetcher) LatestRelease(string) (*githubpipeline.Release, error) {
	return &githubpipeline.Release{TagName: "v1.0.0",
		PublishedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local)}, nil
}

type appFakeOllama struct{}

func (f *appFakeOllama) List(context.Context) ([]string, error) {
	return []string{"llama3"}, nil
}

func (f *appFakeOllama) Chat(ctx context.Context,
	req *llmclient.OllamaChatRequest) (*llmclient.OllamaChatResponse, error) {
	analysis := `{"confidence": 95, "core_value": "The wiring proof.", ` +
		`"difference": "none — it is a fixture", "features": "fixtures, wiring", ` +
		`"how_it_works": "It answers canned data deterministically.", "summary": ` +
		`"A curated offline fixture repo for the DI smoke.", "tags": ["go"]}`
	return &llmclient.OllamaChatResponse{
		Model:   req.Model,
		Message: &llmclient.OllamaChatMessage{Role: "assistant", Content: analysis},
	}, nil
}

func TestBatchSmokeOverFakes(t *testing.T) {
	appDir := t.TempDir()
	t.Setenv("GITCURATOR_HOME", appDir)
	vault := t.TempDir()

	cfgPath := filepath.Join(appDir, "config.json")
	cfg := storage.NewOrderedMap()
	cfg.Set("vault_path", vault)
	cfg.Set("github_token", "fixture-token")
	cfg.Set("llm_provider", "ollama")
	cfg.Set("ollama_model", "llama3")
	if err := storage.WriteConfigFile(cfgPath, cfg); err != nil {
		t.Fatal(err)
	}

	a := New(cfgPath, false)
	a.GHClientFactory = func(string) githubpipeline.RepoFetcher { return &appFakeFetcher{} }
	a.OllamaFactory = func(string) llmclient.OllamaAPI { return &appFakeOllama{} }
	a.BannerDownload = func(_, _, _ string) string { return "" }

	rec := &busRec{}
	rec.subscribe(a.Bus)
	res := a.Commands.SyncRunBatch(context.Background(), BatchOptions{
		URLs:   []string{"https://github.com/octo-owner/curated-repo"},
		Source: "direct",
	})
	if !res.Success {
		t.Fatalf("batch = %+v (logs: %v)", res, rec.logs)
	}
	if res.Processed != 1 || res.Total != 1 {
		t.Fatalf("counts = %d/%d", res.Processed, res.Total)
	}
	if rec.countBatches() != 1 {
		t.Fatalf("BatchFinished fired %d times", rec.countBatches())
	}
	// The note landed in the vault.
	found := false
	_ = filepath.Walk(vault, func(p string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && strings.HasSuffix(p, ".md") {
			found = true
		}
		return nil
	})
	if !found {
		t.Fatal("the processed note never landed in the vault")
	}
	// The engine's events rode the bus: the manifest line, the
	// analysis line, the finished card.
	joined := joinLogs(rec)
	if !strings.Contains(joined, "Manifest created: 1 GitHub") {
		t.Fatalf("the manifest line never rode: %v", joined)
	}
	if !strings.Contains(joined, "Analyzing 'curated-repo'") {
		t.Fatalf("the analysis line never rode: %v", joined)
	}
}

func joinLogs(r *busRec) string {
	r.mu.Lock()
	defer r.mu.Unlock()
	var b strings.Builder
	for _, l := range r.logs {
		b.WriteString(l.Msg)
		b.WriteString("\n")
	}
	return b.String()
}
