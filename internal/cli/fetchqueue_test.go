package cli

// fetchqueue_test.go + autocmd_test.go — T-62 slice 2's driver
// battery: the classification half (the routing split, the vault
// filter, the settlement, the websites ladder), the wrong-vault guard
// (interactive + --yes + non-interactive), the >10 confirm, the
// all-caught-up door, and the missing-settings card.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/assadigit/second-brain-project/internal/llmclient"
	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/store"
	"github.com/assadigit/second-brain-project/internal/websitepipeline"
)

func newFetchEnv(t *testing.T) (*CLI, *strings.Builder, string, string) {
	t.Helper()
	appDir := t.TempDir()
	t.Setenv("GITCURATOR_HOME", appDir)
	vault := filepath.Join(appDir, "vault")
	if err := os.MkdirAll(vault, 0o755); err != nil {
		t.Fatal(err)
	}
	out := &strings.Builder{}
	c := &CLI{
		Out:       out,
		ErrW:      out,
		In:        strings.NewReader(""),
		Now:       func() time.Time { return time.Unix(1000, 0) },
		IsATTYIn:  func() bool { return false },
		IsATTYOut: func() bool { return false },
		IsDir: func(p string) bool {
			fi, err := os.Stat(p)
			return err == nil && fi.IsDir()
		},
	}
	return c, out, appDir, vault
}

func newTestConfig(vault string) *storage.OrderedMap {
	cfg := storage.NewOrderedMap()
	cfg.Set("vault_path", vault)
	oll := storage.NewOrderedMap()
	oll.Set("base_url", "http://127.0.0.1:11434")
	oll.Set("model", "test-model")
	cfg.Set("ollama", oll)
	pipes := storage.NewOrderedMap()
	pipes.Set("github", true)
	pipes.Set("websites", false)
	cfg.Set("pipelines", pipes)
	return cfg
}

func TestRouteQueueItems(t *testing.T) {
	items := []any{
		map[string]any{"url": "https://github.com/octo-owner/repo"},
		map[string]any{"url_normalized": "https://example.com/page?x=1"},
		map[string]any{"url_original": "https://octo-owner.github.io/site/"},
		map[string]any{"url": "https://www.youtube.com/watch?v=x"},
		map[string]any{"junk": true},
		"not a dict",
	}
	githubURLs, nonGithub, raw := routeQueueItems(items)
	if raw != 6 {
		t.Fatalf("raw = %d, want 6", raw)
	}
	if len(githubURLs) != 2 { // the repo + the github.io page
		t.Fatalf("github = %v", githubURLs)
	}
	if len(nonGithub) != 2 { // example.com + youtube
		t.Fatalf("non-github = %v", nonGithub)
	}
	// the url_normalized key speaks first.
	if nonGithub[0] != "https://example.com/page?x=1" {
		t.Fatalf("url_normalized precedence broke: %v", nonGithub)
	}
}

func TestClassifyQueueLinksVaultFilter(t *testing.T) {
	c, _, appDir, vault := newFetchEnv(t)
	// one note in the vault (the in-vault dedup), one dead, one settled,
	// one pending.
	writeVaultNote(t, vault, "existing.md", "https://github.com/octo-owner/done-repo")
	cache, err := store.NewCacheDB(filepath.Join(appDir, "cache.db"))
	if err != nil {
		t.Fatal(err)
	}
	cache.ConfirmDead("https://github.com/octo-owner/dead-repo", "404", 3)
	cache.AddProcessed(900, "https://github.com/octo-owner/settled-repo",
		"octo-owner", "settled-repo", "path", "cat")
	// THE SETTLEMENT's first-door law: pre-settle (the meta guard) so
	// this test classifies like every LATER run — the fetched URLs the
	// cache never knew stay pending.
	cache.SettleExistingRepos(nil)
	cache.Close()

	cfg := newTestConfig(vault)
	items := []any{
		map[string]any{"url": "https://github.com/octo-owner/pending-repo"},
		map[string]any{"url": "https://github.com/octo-owner/done-repo"},
		map[string]any{"url": "https://github.com/octo-owner/dead-repo"},
		map[string]any{"url": "https://github.com/octo-owner/settled-repo"},
	}
	out := c.ClassifyQueueLinks(cfg, items)
	if out.State != "" {
		t.Fatalf("state leaked: %q", out.State)
	}
	if len(out.PendingURLs) != 1 || out.PendingURLs[0] != "https://github.com/octo-owner/pending-repo" {
		t.Fatalf("pending = %v", out.PendingURLs)
	}
	if out.InVaultCount != 1 {
		t.Fatalf("in-vault = %d, want 1", out.InVaultCount)
	}
	if out.Decommissioned != 1 {
		t.Fatalf("decommissioned = %d, want 1", out.Decommissioned)
	}
	if out.ReposSettled != 1 {
		t.Fatalf("settled = %d, want 1", out.ReposSettled)
	}
	if out.VaultIndexCount == nil || *out.VaultIndexCount != 1 {
		t.Fatalf("vault index = %v, want 1", out.VaultIndexCount)
	}
	// the websites half is off: no pending websites + the hint flag.
	if len(out.PendingWebsiteURLs) != 0 || !out.WebsitesPipelineOff {
		// (no non-GitHub links here, so the flag stays false)
		if out.WebsitesPipelineOff {
			t.Fatalf("the pipeline-off hint fired with no links")
		}
	}
}

func TestClassifyQueueLinksWebsitesLadder(t *testing.T) {
	c, _, appDir, vault := newFetchEnv(t)
	webVault := filepath.Join(appDir, "webvault")
	if err := os.MkdirAll(webVault, 0o755); err != nil {
		t.Fatal(err)
	}
	// one website note in the vault (in-vault, real).
	writeVaultNote(t, webVault, "site.md", "https://www.notesite.dev/")

	cfg := newTestConfig(vault)
	cfg.Set("website_vault_path", webVault)
	pipes := storage.NewOrderedMap()
	pipes.Set("github", true)
	pipes.Set("websites", true)
	cfg.Set("pipelines", pipes)

	// THE SETTLEMENT's first-door law: prime it over the FIRST run's
	// queue history (other urls) so the ladder below classifies like
	// every LATER run — the test's five urls are the NEW arrivals.
	c.ClassifyQueueLinks(cfg, []any{
		map[string]any{"url": "https://www.history-1.dev/"},
		map[string]any{"url": "https://www.history-2.dev/"},
	})

	statePath := filepath.Join(appDir, "cache.db")
	state, err := websitepipeline.NewWebsiteStateDB(statePath)
	if err != nil {
		t.Fatal(err)
	}
	// dismissed / processed / settled / failed-with-retries rows.
	_ = state.Dismiss("https://www.discarded.dev/", "owner deleted")
	_ = state.MarkProcessed("https://www.processed.dev/", "note.md", "cat", "sub", "ok")
	_ = state.EnqueueRetry("https://www.walled.dev/", "403 blocked")
	state.Close()

	items := []any{
		map[string]any{"url": "https://www.newsite.dev/"},
		map[string]any{"url": "https://www.notesite.dev/"},
		map[string]any{"url": "https://www.discarded.dev/"},
		map[string]any{"url": "https://www.processed.dev/"},
		map[string]any{"url": "https://www.walled.dev/"},
	}
	out := c.ClassifyQueueLinks(cfg, items)
	if len(out.PendingWebsiteURLs) != 2 ||
		out.PendingWebsiteURLs[0] != "https://www.newsite.dev/" ||
		out.PendingWebsiteURLs[1] != "https://www.walled.dev/" {
		t.Fatalf("pending websites = %v (in-vault=%d dismissed=%d processed=%d settled=%d)",
			out.PendingWebsiteURLs, out.WebsitesInVault, out.WebsitesDismissed,
			out.WebsitesProcessed, out.WebsitesSettled)
	}
	if out.WebsitesInVault != 1 {
		t.Fatalf("in-vault = %d, want 1", out.WebsitesInVault)
	}
	if out.WebsitesDismissed != 1 {
		t.Fatalf("dismissed = %d, want 1", out.WebsitesDismissed)
	}
	if out.WebsitesProcessed != 1 {
		t.Fatalf("processed = %d, want 1", out.WebsitesProcessed)
	}
	if out.WebsitesVaultIndex == nil || *out.WebsitesVaultIndex != 1 {
		t.Fatalf("websites index = %v", out.WebsitesVaultIndex)
	}
}

func writeVaultNote(t *testing.T, vault, name, source string) {
	t.Helper()
	body := "---\nsource: " + source + "\ncategory: tools\n---\n\n# " + name + "\n"
	if err := os.WriteFile(filepath.Join(vault, name), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestWrongVaultGuardInteractive(t *testing.T) {
	c, out, _, vault := newFetchEnv(t)
	cfg := newTestConfig(vault)
	pending := []string{"https://github.com/octo-owner/one"}
	idx := 0
	vaultIdx := 0
	// answer "n" (EOF answers "" → the default "n" path)
	c.In = strings.NewReader("\n")
	stop := c.wrongVaultGuard(cfg, vault, pending, 0, &vaultIdx, 7, false, true)
	if !stop {
		t.Fatalf("the guard must stop on the wrong-vault signature")
	}
	if !strings.Contains(out.String(), "This looks like a WRONG VAULT PATH") {
		t.Fatalf("missing the card:\n%s", out.String())
	}
	if !strings.Contains(out.String(), "Cancelled — nothing was processed. Fix vault_path first.") {
		t.Fatalf("missing the cancel line:\n%s", out.String())
	}
	_ = idx
	// answer "y" → the run continues.
	c2, out2, _, vault2 := newFetchEnv(t)
	c2.In = strings.NewReader("y\n")
	stop2 := c2.wrongVaultGuard(newTestConfig(vault2), vault2, pending, 0, &vaultIdx, 7, false, true)
	if stop2 {
		t.Fatalf("the y answer must continue")
	}
	if !strings.Contains(out2.String(), "Process all 1 links into this empty vault anyway?") {
		t.Fatalf("missing the ask:\n%s", out2.String())
	}
}

func TestWrongVaultGuardYesMode(t *testing.T) {
	c, out, _, vault := newFetchEnv(t)
	vaultIdx := 0
	stop := c.wrongVaultGuard(newTestConfig(vault), vault,
		[]string{"https://github.com/octo-owner/one"}, 0, &vaultIdx, 7, true, true)
	if !stop {
		t.Fatalf("--yes must never re-process on this signature")
	}
	if !strings.Contains(out.String(), "(--yes mode never re-processes on this signature.") {
		t.Fatalf("missing the --yes card:\n%s", out.String())
	}
}

func TestWrongVaultGuardNonInteractive(t *testing.T) {
	c, _, _, vault := newFetchEnv(t)
	vaultIdx := 0
	// the sync door's shape: non-interactive → stop without asking.
	stop := c.wrongVaultGuard(newTestConfig(vault), vault,
		[]string{"https://github.com/octo-owner/one"}, 0, &vaultIdx, 7, false, false)
	if !stop {
		t.Fatalf("the non-interactive guard must stop")
	}
}

func TestWrongVaultGuardNotTripped(t *testing.T) {
	c, _, _, vault := newFetchEnv(t)
	vaultIdx := 3
	// an indexed vault (3 notes) never trips the guard.
	if c.wrongVaultGuard(newTestConfig(vault), vault,
		[]string{"u1"}, 0, &vaultIdx, 7, false, true) {
		t.Fatalf("the guard must not trip with an indexed vault")
	}
	// a cache that knows nothing never trips.
	idx := 0
	if c.wrongVaultGuard(newTestConfig(vault), vault,
		[]string{"u1"}, 0, &idx, 0, false, true) {
		t.Fatalf("the guard must not trip with an empty cache")
	}
}

func TestCmdAutoMissingSettings(t *testing.T) {
	c, out, appDir, _ := newFetchEnv(t)
	// a config with NO vault on disk + no telegram keys.
	cfgPath := filepath.Join(appDir, "config.json")
	if err := storage.WriteConfigFile(cfgPath, storage.NewOrderedMap()); err != nil {
		t.Fatal(err)
	}
	code := c.cmdAuto(&Args{Config: cfgPath})
	if code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	if !strings.Contains(out.String(), "Missing settings: vault path, telegram_api_id, telegram_api_hash, telegram_phone, bot_username") {
		t.Fatalf("missing the settings card:\n%s", out.String())
	}
}

func TestCmdAutoAllCaughtUp(t *testing.T) {
	c, out, appDir, vault := newFetchEnv(t)
	cfgPath := filepath.Join(appDir, "config.json")
	cfg := newTestConfig(vault)
	cfg.Set("telegram_api_id", 123)
	cfg.Set("telegram_api_hash", "hash")
	cfg.Set("telegram_phone", "+1")
	cfg.Set("bot_username", "@bot")
	storage.WriteConfigFile(cfgPath, cfg)
	// the Worker answers an empty queue.
	c.WorkerFetch = func(*storage.OrderedMap) (string, []any) {
		return "ok", nil
	}
	// the preflight must pass: token + proxy + LLM doors.
	c.PreflightGithub = func(string) PreflightGHReply {
		return PreflightGHReply{Status: 200, Login: "octo-owner"}
	}
	c.OllamaModels = func(string, float64) ([]string, error) {
		return []string{"test-model"}, nil
	}
	code := c.cmdAuto(&Args{Config: cfgPath})
	if code != 0 {
		t.Fatalf("exit = %d, want 0; out:\n%s", code, out.String())
	}
	if !strings.Contains(out.String(), "All caught up — nothing to process. 🎉") {
		t.Fatalf("missing the caught-up line:\n%s", out.String())
	}
}

func TestCmdAutoFetchFailureCard(t *testing.T) {
	c, out, appDir, vault := newFetchEnv(t)
	cfgPath := filepath.Join(appDir, "config.json")
	cfg := newTestConfig(vault)
	cfg.Set("telegram_api_id", 123)
	cfg.Set("telegram_api_hash", "hash")
	cfg.Set("telegram_phone", "+1")
	cfg.Set("bot_username", "@bot")
	storage.WriteConfigFile(cfgPath, cfg)
	c.WorkerFetch = func(*storage.OrderedMap) (string, []any) {
		return "error", nil
	}
	c.PreflightGithub = func(string) PreflightGHReply {
		return PreflightGHReply{Status: 200, Login: "octo-owner"}
	}
	c.OllamaModels = func(string, float64) ([]string, error) {
		return []string{"test-model"}, nil
	}
	code := c.cmdAuto(&Args{Config: cfgPath})
	if code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	if !strings.Contains(out.String(), "Fetch failed: the Worker queue poll failed") {
		t.Fatalf("missing the failure card:\n%s", out.String())
	}
}

func TestCmdAutoConfirmOver10(t *testing.T) {
	c, out, appDir, vault := newFetchEnv(t)
	cfgPath := filepath.Join(appDir, "config.json")
	cfg := newTestConfig(vault)
	cfg.Set("telegram_api_id", 123)
	cfg.Set("telegram_api_hash", "hash")
	cfg.Set("telegram_phone", "+1")
	cfg.Set("bot_username", "@bot")
	storage.WriteConfigFile(cfgPath, cfg)
	var items []any
	for i := 0; i < 11; i++ {
		items = append(items, map[string]any{
			"url": fmt.Sprintf("https://github.com/octo-owner/repo-%d", i)})
	}
	// pre-settle (a machine that already ran its first --auto)
	if cache, err := store.NewCacheDB(filepath.Join(appDir, "cache.db")); err == nil {
		cache.SettleExistingRepos(nil)
		cache.Close()
	}
	c.WorkerFetch = func(*storage.OrderedMap) (string, []any) {
		return "ok", items
	}
	c.PreflightGithub = func(string) PreflightGHReply {
		return PreflightGHReply{Status: 200, Login: "octo-owner"}
	}
	c.OllamaModels = func(string, float64) ([]string, error) {
		return []string{"test-model"}, nil
	}
	// a TTY stdin + the "n" answer → cancelled, exit 0.
	c.IsATTYIn = func() bool { return true }
	c.In = strings.NewReader("n\n")
	code := c.cmdAuto(&Args{Config: cfgPath})
	if code != 0 {
		t.Fatalf("exit = %d, want 0 (cancelled); out:\n%s", code, out.String())
	}
	if !strings.Contains(out.String(), "11 item(s) ready (11 repo(s) + 0 website(s)) — process them all?") {
		t.Fatalf("missing the confirm ask:\n%s", out.String())
	}
	if !strings.Contains(out.String(), "Cancelled — nothing was processed.") {
		t.Fatalf("missing the cancel line:\n%s", out.String())
	}
}

func TestCmdAutoWrongVaultGuardStops(t *testing.T) {
	c, out, appDir, vault := newFetchEnv(t)
	cfgPath := filepath.Join(appDir, "config.json")
	cfg := newTestConfig(vault)
	cfg.Set("telegram_api_id", 123)
	cfg.Set("telegram_api_hash", "hash")
	cfg.Set("telegram_phone", "+1")
	cfg.Set("bot_username", "@bot")
	storage.WriteConfigFile(cfgPath, cfg)
	// the wrong-vault signature: pending links, empty index, a cache
	// that knows processed repos. THE SETTLEMENT primed first (the
	// machine's first --auto already ran), so the FRESH url stays
	// pending while the cache knows the processed ones.
	cache, _ := store.NewCacheDB(filepath.Join(appDir, "cache.db"))
	cache.AddProcessed(900, "https://github.com/octo-owner/old-repo",
		"octo-owner", "old-repo", "path", "cat")
	cache.SettleExistingRepos(nil)
	cache.Close()
	c.WorkerFetch = func(*storage.OrderedMap) (string, []any) {
		return "ok", []any{map[string]any{"url": "https://github.com/octo-owner/fresh-repo"}}
	}
	c.PreflightGithub = func(string) PreflightGHReply {
		return PreflightGHReply{Status: 200, Login: "octo-owner"}
	}
	c.OllamaModels = func(string, float64) ([]string, error) {
		return []string{"test-model"}, nil
	}
	c.In = strings.NewReader("\n") // EOF → "n"
	code := c.cmdAuto(&Args{Config: cfgPath, Yes: true})
	if code != 1 {
		t.Fatalf("exit = %d, want 1 (--yes stops on the signature)", code)
	}
	if !strings.Contains(out.String(), "This looks like a WRONG VAULT PATH") {
		t.Fatalf("missing the guard card:\n%s", out.String())
	}
	if !strings.Contains(out.String(), "Dedup ground truth — vault index: 0 note(s) · cache: 1 processed repo(s)") {
		t.Fatalf("missing the ground-truth line:\n%s", out.String())
	}
}

func TestCmdAutoQueueCard(t *testing.T) {
	c, out, appDir, vault := newFetchEnv(t)
	cfgPath := filepath.Join(appDir, "config.json")
	cfg := newTestConfig(vault)
	cfg.Set("telegram_api_id", 123)
	cfg.Set("telegram_api_hash", "hash")
	cfg.Set("telegram_phone", "+1")
	cfg.Set("bot_username", "@bot")
	storage.WriteConfigFile(cfgPath, cfg)
	if cache, err := store.NewCacheDB(filepath.Join(appDir, "cache.db")); err == nil {
		cache.SettleExistingRepos(nil)
		cache.Close()
	}
	c.WorkerFetch = func(*storage.OrderedMap) (string, []any) {
		return "ok", []any{
			map[string]any{"url": "https://github.com/octo-owner/repo-1"},
			map[string]any{"url": "https://www.somewhere.dev/page"},
		}
	}
	c.PreflightGithub = func(string) PreflightGHReply {
		return PreflightGHReply{Status: 200, Login: "octo-owner"}
	}
	c.OllamaModels = func(string, float64) ([]string, error) {
		return []string{"test-model"}, nil
	}
	// no TTY → the >10 confirm never asks; the batch proceeds (the
	// engine runs offline over the fake doors).
	c.NewStatusLineFn = func() StatusPainter { return &recStatus{} }
	c.BannerDownload = func(owner, repoName, folder string) string { return "" }
	// a dead Ollama aborts the batch early — enough to see the queue
	// card. (The full engine path is engine_test's golden path.)
	c.OllamaFactory = func(string) llmclient.OllamaAPI {
		return &fakeOllama{listErr: fmt.Errorf("refused")}
	}
	c.AutostartServe = func() (int, string) { return 0, "FileNotFoundError" }
	code := c.cmdAuto(&Args{Config: cfgPath})
	if code != 1 {
		t.Fatalf("exit = %d, want 1 (the dead Ollama aborts)", code)
	}
	if !strings.Contains(out.String(), "Queue: 1 repo(s) pending · 0 already in vault · 0 decommissioned · 0 website(s) pending · 1 non-GitHub") {
		t.Fatalf("missing the queue card:\n%s", out.String())
	}
	if !strings.Contains(out.String(), "1 non-GitHub link(s) are waiting as _inbox rows") {
		t.Fatalf("missing the pipeline-off hint:\n%s", out.String())
	}
}
