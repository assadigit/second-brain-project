package cli

// engine_test.go — T-62 slice 2's unit battery: the engine's lifecycle
// (the mode dispatch's early returns, the crash wrapper's exit law),
// the finish card, the dry-run path (the entry-count line + the
// report + NO vault writes), the warmup family's doors (the dead
// server + the autostart ladder, the single-model auto-switch), the
// heartbeat ticker (T-61's spinner under a real loop), the seal order,
// and the full offline golden path (the fake GH client + the fake
// Ollama — one link, one note, the honest exit code).

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/assadigit/second-brain-project/internal/config"
	"github.com/assadigit/second-brain-project/internal/githubpipeline"
	"github.com/assadigit/second-brain-project/internal/llmclient"
	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/store"
)

// recStatus is the StatusPainter recorder (the golden replay's twin —
// the log sequence without the \r repaints).
type recStatus struct {
	mu    sync.Mutex
	logs  []string
	ticks int
}

func (r *recStatus) Start(label string, total int) {}
func (r *recStatus) Stop(final string)             {}
func (r *recStatus) SetProgress(cur, total int)    {}
func (r *recStatus) SetRepo(repo string)           {}
func (r *recStatus) Tick()                         { r.mu.Lock(); r.ticks++; r.mu.Unlock() }
func (r *recStatus) EraseIfActive()                {}
func (r *recStatus) Log(text, level string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.logs = append(r.logs, level+"|"+text)
}

func (r *recStatus) logLines() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.logs...)
}

func (r *recStatus) tickCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.ticks
}

// fakeOllama answers List + Chat with canned data.
type fakeOllama struct {
	models  []string
	chatErr map[string]error // by model
	chat    map[string]string
	listErr error
	mu      sync.Mutex
	chats   int
}

func (f *fakeOllama) List(ctx context.Context) ([]string, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.models, nil
}

func (f *fakeOllama) Chat(ctx context.Context,
	req *llmclient.OllamaChatRequest) (*llmclient.OllamaChatResponse, error) {
	f.mu.Lock()
	f.chats++
	f.mu.Unlock()
	if err, ok := f.chatErr[req.Model]; ok {
		return nil, err
	}
	if txt, ok := f.chat[req.Model]; ok {
		return &llmclient.OllamaChatResponse{
			Model:   req.Model,
			Message: &llmclient.OllamaChatMessage{Role: "assistant", Content: txt},
		}, nil
	}
	return &llmclient.OllamaChatResponse{
		Model:   req.Model,
		Message: &llmclient.OllamaChatMessage{Role: "assistant", Content: "{}"},
	}, nil
}

// fakeFetcher is the offline GitHub door (the runner_test twin).
type fakeFetcher struct {
	repos map[string]*githubpipeline.Repo
}

func (f *fakeFetcher) GetRateLimit() (*githubpipeline.RateLimit, error) {
	return &githubpipeline.RateLimit{Remaining: 5000,
		Reset: time.Now().Add(time.Hour)}, nil
}

func (f *fakeFetcher) GetRepo(fullName string) (*githubpipeline.Repo, error) {
	if r, ok := f.repos[fullName]; ok {
		return r, nil
	}
	return nil, &githubpipeline.GHError{Status: 404, Message: "404 {'message': 'Not Found'}"}
}

func (f *fakeFetcher) GetTopics(fullName string) ([]string, error) {
	return []string{"go"}, nil
}

func (f *fakeFetcher) CountCommitsSince(fullName string, since time.Time) (int, error) {
	return 42, nil
}

func (f *fakeFetcher) GetReadme(fullName string) (*githubpipeline.ReadmeResult, error) {
	return &githubpipeline.ReadmeResult{Content: "A test readme.",
		RawChars: 15}, nil
}

func (f *fakeFetcher) GetLanguages(fullName string) ([]string, error) {
	return []string{"Go"}, nil
}

func (f *fakeFetcher) LatestRelease(fullName string) (*githubpipeline.Release, error) {
	return &githubpipeline.Release{TagName: "v1.0.0",
		PublishedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}, nil
}

// analysisJSON is the canned LLM answer (the loop's contract).
func analysisJSON() string {
	m := map[string]any{
		"summary":      "A curated test repo.",
		"how_it_works": "It works.",
		"core_value":   "Testing.",
		"features":     "tests",
		"difference":   "none",
		"category":     "developer-tools",
		"confidence":   95,
		"tags":         []any{"go", "testing"},
	}
	b, _ := json.Marshal(m)
	return string(b)
}

// engineEnv is one offline engine run's fixtures.
type engineEnv struct {
	c       *CLI
	rec     *recStatus
	cfg     *storage.OrderedMap
	appDir  string
	vault   string
	fetcher *fakeFetcher
	ollama  *fakeOllama
	out     *strings.Builder
}

func newEngineEnv(t *testing.T) *engineEnv {
	t.Helper()
	appDir := t.TempDir()
	t.Setenv("GITCURATOR_HOME", appDir)
	vault := t.TempDir()
	e := &engineEnv{
		appDir:  appDir,
		vault:   vault,
		rec:     &recStatus{},
		out:     &strings.Builder{},
		fetcher: &fakeFetcher{repos: map[string]*githubpipeline.Repo{}},
		ollama:  &fakeOllama{},
	}
	// the config: an alphabetical literal (the saved-order law).
	cfgJSON := `{
  "github_token": "",
  "llm_provider": "ollama",
  "ollama": {"base_url": "http://127.0.0.1:11434", "model": "test-model"},
  "pipelines": {"github": true, "websites": false},
  "vault_path": "` + filepath.ToSlash(vault) + `"
}`
	cfgPath := filepath.Join(appDir, "config.json")
	if err := os.WriteFile(cfgPath, []byte(cfgJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	om, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("config load: %v", err)
	}
	e.cfg = om
	e.fetcher.repos["octo-owner/curated-repo"] = &githubpipeline.Repo{
		ID: 1, FullName: "octo-owner/curated-repo", Name: "curated-repo",
		StargazersCount: 12500, ForksCount: 12,
		Description: "The test repo.", OwnerLogin: "octo-owner",
	}
	e.ollama.models = []string{"test-model"}
	e.ollama.chat = map[string]string{"test-model": analysisJSON()}
	e.c = &CLI{
		Out:       e.out,
		ErrW:      e.out,
		In:        strings.NewReader(""),
		Now:       func() time.Time { return time.Unix(1000, 0) },
		IsATTYIn:  func() bool { return false },
		IsATTYOut: func() bool { return false },
		IsDir: func(p string) bool {
			fi, err := os.Stat(p)
			return err == nil && fi.IsDir()
		},
		NewStatusLineFn: func() StatusPainter { return e.rec },
		GHClientFactory: func(string) githubpipeline.RepoFetcher { return e.fetcher },
		OllamaFactory:   func(string) llmclient.OllamaAPI { return e.ollama },
		// THE HERMETIC BANNER STUB: without it every engine test ran
		// the REAL DownloadBanner — a live 500 KB network fetch whose
		// goroutine could outlive the test (the 15 s join's escape
		// hatch) and record a late atomic-write into a LATER test's
		// dry-run window (SESSION-034's root cause of the golden's
		// intermittent 13th entry — the runner's own doc: "the tests
		// inject the fake banner"; the golden harness always did).
		BannerDownload: func(_, _, _ string) string { return "" },
		WorkerFetch:    func(*storage.OrderedMap) (string, []any) { return "disabled", nil },
	}
	return e
}

// findLog answers whether a recorded log line exists (level|text).
func findLog(lines []string, level, text string) bool {
	for _, l := range lines {
		if l == level+"|"+text {
			return true
		}
	}
	return false
}

// findLogHas answers whether some recorded line contains the needle.
func findLogHas(lines []string, needle string) bool {
	for _, l := range lines {
		if strings.Contains(l, needle) {
			return true
		}
	}
	return false
}

func TestEngineBothPipelinesOff(t *testing.T) {
	e := newEngineEnv(t)
	e.cfg.Set("pipelines", func() *storage.OrderedMap {
		p := storage.NewOrderedMap()
		p.Set("github", false)
		p.Set("websites", false)
		return p
	}())
	code := e.c.RunBatchVisual(e.cfg, RunBatchOptions{Mode: "direct"})
	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (both-off is a success-shaped no-op)", code)
	}
	if !findLog(e.rec.logLines(), "warning",
		"⛔ Both pipelines are switched OFF (Settings → 📁 Vault). "+
			"Nothing was fetched or processed.") {
		t.Fatalf("missing both-off log; got %v", e.rec.logLines())
	}
	out := e.out.String()
	if !strings.Contains(out, "✔ BATCH DONE") {
		t.Fatalf("missing the DONE card:\n%s", out)
	}
	if !strings.Contains(out, "Both pipelines are off — nothing to do.") {
		t.Fatalf("missing the both-off message:\n%s", out)
	}
	if !strings.Contains(out, "in 0m 01s") {
		t.Fatalf("the frozen clock renders the elapsed floor (max(1,int)):\n%s", out)
	}
}

func TestEngineNoURLs(t *testing.T) {
	e := newEngineEnv(t)
	code := e.c.RunBatchVisual(e.cfg, RunBatchOptions{Mode: "direct"})
	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (no-URLs is success-shaped)", code)
	}
	if !findLog(e.rec.logLines(), "warning", "No URLs found to process.") {
		t.Fatalf("missing no-URLs log; got %v", e.rec.logLines())
	}
	if !strings.Contains(e.out.String(), "No URLs found.") {
		t.Fatalf("missing the no-URLs message:\n%s", e.out.String())
	}
}

func TestEngineOllamaUnavailable(t *testing.T) {
	e := newEngineEnv(t)
	dead := &fakeOllama{listErr: fmt.Errorf(
		"connect ECONNREFUSED 127.0.0.1:11434")}
	e.c.OllamaFactory = func(string) llmclient.OllamaAPI { return dead }
	e.c.AutostartServe = func() (int, string) { return 0, "FileNotFoundError" }
	code := e.c.RunBatchVisual(e.cfg, RunBatchOptions{
		Mode: "direct",
		URLs: []string{"https://github.com/octo-owner/curated-repo"},
	})
	if code != 1 {
		t.Fatalf("exit code = %d, want 1 (Ollama not available)", code)
	}
	logs := e.rec.logLines()
	if !findLogHas(logs, "❌ Ollama is not running:") {
		t.Fatalf("missing the dead-server line; got %v", logs)
	}
	if !findLogHas(logs, "❌ 'ollama' was not found on PATH") {
		t.Fatalf("missing the PATH line; got %v", logs)
	}
	if !findLogHas(logs, "🚀 Ollama server not reachable") {
		t.Fatalf("missing the autostart line; got %v", logs)
	}
	out := e.out.String()
	if !strings.Contains(out, "✖ BATCH FAILED") {
		t.Fatalf("missing the FAILED card:\n%s", out)
	}
	if !strings.Contains(out, "Ollama not available") {
		t.Fatalf("missing the message:\n%s", out)
	}
}

func TestEngineGoldenPath(t *testing.T) {
	e := newEngineEnv(t)
	code := e.c.RunBatchVisual(e.cfg, RunBatchOptions{
		Mode:      "direct",
		URLs:      []string{"https://github.com/octo-owner/curated-repo"},
		BotSource: true,
	})
	if code != 0 {
		t.Fatalf("exit code = %d, want 0; out:\n%s\nlogs: %v", code,
			e.out.String(), e.rec.logLines())
	}
	logs := e.rec.logLines()
	for _, want := range []string{
		"🔥 Warming up Ollama model 'test-model'...",
		"✅ Model warmed up",
		"📋 Manifest created: 1 GitHub + 0 non-GitHub links recorded",
	} {
		if !findLogHas(logs, want) {
			t.Fatalf("missing %q; got %v", want, logs)
		}
	}
	out := e.out.String()
	if !strings.Contains(out, "✔ BATCH DONE") {
		t.Fatalf("missing the DONE card:\n%s", out)
	}
	if !strings.Contains(out, "Processed 1 out of 1 repos.") {
		t.Fatalf("missing the finished message:\n%s", out)
	}
	// the note landed in the vault.
	var notes []string
	_ = filepath.Walk(e.vault, func(p string, fi os.FileInfo, err error) error {
		if err == nil && strings.HasSuffix(p, ".md") &&
			!strings.HasSuffix(p, "_undo_last_batch.txt") {
			notes = append(notes, p)
		}
		return nil
	})
	if len(notes) == 0 {
		t.Fatalf("no note written under %s", e.vault)
	}
	// the seal order: VaultSeal → Good Repos → (websites + manual
	// silent). The seal lines ride the status log (the recorder).
	logs = e.rec.logLines()
	idxOf := func(needle string) int {
		for i, l := range logs {
			if strings.Contains(l, needle) {
				return i
			}
		}
		return -1
	}
	vsIdx := idxOf("VaultSeal:")
	grIdx := idxOf("Good Repos:")
	if vsIdx < 0 || grIdx < 0 || vsIdx > grIdx {
		t.Fatalf("the seals' order broke (VaultSeal@%d, GoodRepos@%d):\n%v\nOUT:\n%s",
			vsIdx, grIdx, logs, out)
	}
	if idxOf("Websites vault seal:") >= 0 {
		t.Fatalf("the websites seal must be silent while the pipeline is off:\n%v", logs)
	}
	// the Phase-5 CLEAR: bot_source=True + all clear. The verdict
	// rides CliPrint (the stopped status line's log == CliPrint —
	// the landed PhaseFiveClear prints after status.stop()).
	if !strings.Contains(e.out.String(),
		"✓ Bot queue synced — all links verified (Phase 5 CLEAR") {
		t.Fatalf("missing the Phase-5 CLEAR verdict; out:\n%s", e.out.String())
	}
	// T-61's leg: the heartbeat ticker drove the spinner under the loop.
	if e.rec.tickCount() == 0 {
		t.Fatalf("the heartbeat never ticked")
	}
}

func TestEngineDryRunPath(t *testing.T) {
	e := newEngineEnv(t)
	code := e.c.RunBatchVisual(e.cfg, RunBatchOptions{
		Mode:   "direct",
		URLs:   []string{"https://github.com/octo-owner/curated-repo"},
		DryRun: true,
	})
	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (a dry-run succeeds); out:\n%s", code,
			e.out.String())
	}
	logs := e.rec.logLines()
	if !findLogHas(logs, "🧪 DRY-RUN: every vault write this batch would perform") {
		t.Fatalf("missing the dry-run banner; got %v", logs)
	}
	if !findLogHas(logs, "🧪 DRY-RUN COMPLETE — 0 vault operation(s)") &&
		!findLogHas(logs, "🧪 DRY-RUN COMPLETE — ") {
		t.Fatalf("missing the dry-run completion line; got %v", logs)
	}
	// the report landed under APP_DIR/reports/dry-runs/.
	entries, _ := os.ReadDir(filepath.Join(e.appDir, "reports", "dry-runs"))
	if len(entries) != 1 {
		t.Fatalf("expected exactly one dry-run report, got %d", len(entries))
	}
	// NO note in the vault (writes withheld).
	_ = filepath.Walk(e.vault, func(p string, fi os.FileInfo, err error) error {
		if err == nil && strings.HasSuffix(p, ".md") {
			t.Errorf("a note leaked into the vault in dry-run: %s", p)
		}
		return nil
	})
	// the seals are skipped.
	out := e.out.String()
	if strings.Contains(out, "VaultSeal:") {
		t.Fatalf("a dry-run must not seal:\n%s", out)
	}
	if strings.Contains(out, "✓ Bot queue synced") {
		t.Fatalf("a dry-run must not speak the Phase-5 CLEAR verdict:\n%s", out)
	}
	// the rehearsal records NOTHING: the real cache (the §4.4 classify
	// may open + schema-create it — the reference's own quirk, its
	// classify_changes constructs the full NoteStateDB) carries no
	// processed marks, no failures, no quarantine.
	if cache, err := store.NewCacheDB(filepath.Join(e.appDir, "cache.db")); err == nil {
		if got := len(cache.GetAllProcessedURLs()); got != 0 {
			t.Fatalf("a dry-run leaked %d processed row(s)", got)
		}
		if got := cache.GetFailedCount(); got != 0 {
			t.Fatalf("a dry-run leaked %d failed row(s)", got)
		}
		if got := len(cache.GetQuarantineStats()); got != 0 {
			t.Fatalf("a dry-run leaked %d quarantine row(s)", got)
		}
		cache.Close()
	}
}

func TestEngineWarmupAutoSwitch(t *testing.T) {
	e := newEngineEnv(t)
	// the configured model is missing; Ollama has exactly one other.
	e.ollama.models = []string{"other-model"}
	e.ollama.chat = map[string]string{"other-model": analysisJSON()}
	e.ollama.chatErr = map[string]error{"test-model": fmt.Errorf(
		"ollama error: model 'test-model' not found, try pulling it first")}
	e.cfg.Set("__config_path__", filepath.Join(e.appDir, "config.json"))
	code := e.c.RunBatchVisual(e.cfg, RunBatchOptions{
		Mode: "direct",
		URLs: []string{"https://github.com/octo-owner/curated-repo"},
	})
	if code != 0 {
		t.Fatalf("exit code = %d, want 0; logs: %v\nout:\n%s", code,
			e.rec.logLines(), e.out.String())
	}
	logs := e.rec.logLines()
	if !findLogHas(logs, "🔄 Configured model 'test-model' not found — Ollama has only "+
		"'other-model'. Auto-switching to it (saved to Settings).") {
		t.Fatalf("missing the auto-switch line; got %v", logs)
	}
	if !findLogHas(logs, "LLM model switched to 'other-model' (ollama) — saving config.") {
		t.Fatalf("missing the model-changed line; got %v", logs)
	}
	// the switch persisted: the saved config carries the new model.
	saved, err := os.ReadFile(filepath.Join(e.appDir, "config.json"))
	if err != nil {
		t.Fatalf("the config save never happened: %v", err)
	}
	if !strings.Contains(string(saved), "other-model") {
		t.Fatalf("the saved config lacks the switched model:\n%s", string(saved))
	}
}

func TestEngineImportMode(t *testing.T) {
	e := newEngineEnv(t)
	imp := filepath.Join(e.appDir, "links.txt")
	if err := os.WriteFile(imp, []byte(
		"https://github.com/octo-owner/curated-repo\n\n# a comment\nnot a link\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	code := e.c.RunBatchVisual(e.cfg, RunBatchOptions{
		Mode:       "import",
		ImportFile: imp,
	})
	if code != 0 {
		t.Fatalf("exit code = %d, want 0; out:\n%s\nlogs: %v", code,
			e.out.String(), e.rec.logLines())
	}
	logs := e.rec.logLines()
	if !findLogHas(logs, "📄 Loaded 1 GitHub URLs + 0 website URLs from import file.") {
		t.Fatalf("missing the loaded line; got %v", logs)
	}
	if !findLogHas(logs, "⚠️ 1 line(s) with no recognizable address skipped:") {
		t.Fatalf("missing the unparsed line; got %v", logs)
	}
	if findLogHas(logs, "The import file contained no addresses") {
		// the "no addresses" line only fires when BOTH lists are empty —
		// here one address parsed, so the line must be absent.
		t.Fatalf("the empty-file line fired wrongly; got %v", logs)
	}
}

func TestEngineCrashWrapper(t *testing.T) {
	e := newEngineEnv(t)
	// a cache that cannot open: NewCacheDB panics through the wrapper.
	// point the app dir at a file so cache.db cannot be created:
	blocked := filepath.Join(e.appDir, "blocked")
	if err := os.WriteFile(blocked, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GITCURATOR_HOME", blocked)
	oll := e.ollama
	e.c.OllamaFactory = func(string) llmclient.OllamaAPI { return oll }
	code := e.c.RunBatchVisual(e.cfg, RunBatchOptions{
		Mode: "direct",
		URLs: []string{"https://github.com/octo-owner/curated-repo"},
	})
	if code != 1 {
		t.Fatalf("exit code = %d, want 1 (a crashed batch)", code)
	}
	if !findLogHas(e.rec.logLines(), "💥 Batch crashed:") {
		t.Fatalf("missing the crash card; got %v", e.rec.logLines())
	}
	if !strings.Contains(e.out.String(), "✖ BATCH FAILED") {
		t.Fatalf("missing the FAILED card:\n%s", e.out.String())
	}
}

func TestEngineFinishElapsedFloor(t *testing.T) {
	// the elapsed law: max(1, int(now - started)) — the frozen clock
	// answers 0 seconds → the card still says 1s.
	e := newEngineEnv(t)
	e.cfg.Set("pipelines", func() *storage.OrderedMap {
		p := storage.NewOrderedMap()
		p.Set("github", false)
		p.Set("websites", false)
		return p
	}())
	start := time.Unix(2000, 0)
	now := start
	e.c.Now = func() time.Time { return now }
	_ = e.c.RunBatchVisual(e.cfg, RunBatchOptions{Mode: "direct"})
	if !strings.Contains(e.out.String(), "in 0m 01s") {
		t.Fatalf("the elapsed floor broke:\n%s", e.out.String())
	}
}
