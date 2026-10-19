// batchwire_test.go — T-74 slice 1's own battery: the CLI-parity
// wiring's laws over the app's injected seams (the hermetic law —
// every network/process door faked at construction):
//   - the pipeline-switches gate + the no-URLs early return;
//   - the warmup family's doors: the dead Ollama + the FileNotFoundError
//     autostart refusal + the autostart-comes-up ladder, the
//     single-model auto-switch + the model-prompt door + the declined
//     best-stand-in, the llama.cpp abort + the detect-and-switch save,
//     the cloud /models pre-flight's three verdicts;
//   - the import fetcher (the one-link grammar + the _inbox tables +
//     the counters);
//   - the dry-run law (the banner + the entry count + NO vault writes
//   - NO seals);
//   - the finish chain (the seals in the reference's order) + THE
//     PHASE-5 CLEAR (the mirror-push ack legs + the all-clear verdict
//     shapes over crafted trackers);
//   - the websites phase's wiring (ON with a vault runs the phase; ON
//     without a vault speaks the honest warning and keeps the batch).
package app

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/assadigit/second-brain-project/internal/batch"
	"github.com/assadigit/second-brain-project/internal/githubpipeline"
	"github.com/assadigit/second-brain-project/internal/llmclient"
	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/store"
)

// ---------------------------------------------------------------------------
// The wiring's own fakes.
// ---------------------------------------------------------------------------

// wireOllama is the configurable Ollama fake (the warmup family's
// doors): List fails for the first listFailures calls; Chat fails for
// the models in chatErr.
type wireOllama struct {
	mu           sync.Mutex
	models       []string
	listFailures int
	listErr      error
	chatErr      map[string]error
}

func (f *wireOllama) List(context.Context) ([]string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.listFailures > 0 {
		f.listFailures--
		return nil, f.listErr
	}
	return f.models, nil
}

func (f *wireOllama) Chat(ctx context.Context,
	req *llmclient.OllamaChatRequest) (*llmclient.OllamaChatResponse, error) {
	f.mu.Lock()
	err := f.chatErr[req.Model]
	f.mu.Unlock()
	if err != nil {
		return nil, err
	}
	return &llmclient.OllamaChatResponse{
		Model:   req.Model,
		Message: &llmclient.OllamaChatMessage{Role: "assistant", Content: "ok"},
	}, nil
}

// logCapture is the BatchEvents log recorder.
type logCapture struct {
	mu   sync.Mutex
	logs []string
}

func (c *logCapture) log(msg, level string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.logs = append(c.logs, level+"|"+msg)
}

func (c *logCapture) joined() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return strings.Join(c.logs, "\n")
}

func (c *logCapture) has(part string) bool {
	return strings.Contains(c.joined(), part)
}

// joinLogsHas answers whether the bus logs contain part.
func joinLogsHas(r *busRec, part string) bool {
	return strings.Contains(joinLogs(r), part)
}

// newWiredApp builds a hermetic App over the fake engine doors (the
// wiring tests' harness).
func newWiredApp(t *testing.T, cfg *storage.OrderedMap) (*App, *busRec) {
	t.Helper()
	appDir := t.TempDir()
	t.Setenv("GITCURATOR_HOME", appDir)
	cfgPath := filepath.Join(appDir, "config.json")
	if err := storage.WriteConfigFile(cfgPath, cfg); err != nil {
		t.Fatal(err)
	}
	a := New(cfgPath, false)
	rec := &busRec{}
	rec.subscribe(a.Bus)
	a.GHClientFactory = func(string) githubpipeline.RepoFetcher { return &appFakeFetcher{} }
	a.BannerDownload = func(_, _, _ string) string { return "" }
	return a, rec
}

// craftedEngine builds a bare engine for the unit doors (the warmup
// family + the finish legs) — in-package, exactly like the CLI's
// engine tests.
func craftedEngine(a *App, cfg *storage.OrderedMap, lc *logCapture) *batchEngine {
	return &batchEngine{
		a:   a,
		cfg: cfg,
		ev:  BatchEvents{Log: lc.log},
		fin: &BatchResult{},
	}
}

// ---------------------------------------------------------------------------
// The pipeline gates.
// ---------------------------------------------------------------------------

func TestWireBothPipelinesOff(t *testing.T) {
	cfg := storage.NewOrderedMap()
	vault := t.TempDir()
	cfg.Set("vault_path", vault)
	pipes := storage.NewOrderedMap()
	pipes.Set("github", false)
	pipes.Set("websites", false)
	cfg.Set("pipelines", pipes)
	a, rec := newWiredApp(t, cfg)

	res := a.Commands.SyncRunBatch(context.Background(), BatchOptions{
		URLs: []string{"https://github.com/octo-owner/curated-repo"},
	})
	if !res.Success || res.Message != "Both pipelines are off — nothing to do." {
		t.Fatalf("res = %+v", res)
	}
	if !joinLogsHas(rec, "⛔ Both pipelines are switched OFF") {
		t.Fatalf("the gate never spoke: %s", joinLogs(rec))
	}
	if rec.countBatches() != 1 {
		t.Fatalf("BatchFinished fired %d times", rec.countBatches())
	}
}

func TestWireNoURLsFound(t *testing.T) {
	cfg := storage.NewOrderedMap()
	cfg.Set("vault_path", t.TempDir())
	a, rec := newWiredApp(t, cfg)
	a.OllamaFactory = func(string) llmclient.OllamaAPI { return &wireOllama{models: []string{"m"}} }

	res := a.Commands.SyncRunBatch(context.Background(), BatchOptions{})
	if !res.Success || res.Message != "No URLs found." {
		t.Fatalf("res = %+v", res)
	}
	if !joinLogsHas(rec, "No URLs found to process.") {
		t.Fatalf("the honest line never spoke: %s", joinLogs(rec))
	}
}

// ---------------------------------------------------------------------------
// The warmup family's doors.
// ---------------------------------------------------------------------------

func TestWireOllamaDeadAborts(t *testing.T) {
	cfg := storage.NewOrderedMap()
	cfg.Set("vault_path", t.TempDir())
	a, rec := newWiredApp(t, cfg)
	a.OllamaFactory = func(string) llmclient.OllamaAPI {
		return &wireOllama{listFailures: 99, listErr: errors.New("connection refused")}
	}
	a.AutostartServe = func() (int, string) { return 0, "FileNotFoundError" }

	res := a.Commands.SyncRunBatch(context.Background(), BatchOptions{
		URLs: []string{"https://github.com/octo-owner/curated-repo"},
	})
	if res.Success || res.Message != "Ollama not available" {
		t.Fatalf("res = %+v", res)
	}
	j := joinLogs(rec)
	for _, want := range []string{
		"❌ Ollama is not running: connection refused",
		"🚀 Ollama server not reachable — trying to start it automatically",
		"❌ 'ollama' was not found on PATH",
	} {
		if !strings.Contains(j, want) {
			t.Fatalf("missing %q in:\n%s", want, j)
		}
	}
}

func TestWireOllamaAutostartComesUp(t *testing.T) {
	cfg := storage.NewOrderedMap()
	cfg.Set("vault_path", t.TempDir())
	a, rec := newWiredApp(t, cfg)
	// The first List fails (the warmup's probe), the autostart's first
	// poll then succeeds.
	a.OllamaFactory = func(string) llmclient.OllamaAPI {
		return &wireOllama{models: []string{"m"}, listFailures: 1,
			listErr: errors.New("connection refused")}
	}
	a.AutostartServe = func() (int, string) { return 4321, "" }

	res := a.Commands.SyncRunBatch(context.Background(), BatchOptions{
		URLs: []string{"https://github.com/octo-owner/missing-repo"},
	})
	j := joinLogs(rec)
	for _, want := range []string{
		"   started 'ollama serve' (PID 4321)",
		"✅ Ollama is up (probe 1/16)",
		"🔥 Warming up Ollama model",
	} {
		if !strings.Contains(j, want) {
			t.Fatalf("missing %q in:\n%s", want, j)
		}
	}
	// A 404 repo URL finishes through the loop without any LLM call.
	if res.Message == "Ollama not available" {
		t.Fatalf("the autostart never recovered: %s", j)
	}
}

func TestWireWarmupAutoSwitch(t *testing.T) {
	cfg := storage.NewOrderedMap()
	vault := t.TempDir()
	cfg.Set("vault_path", vault)
	a, rec := newWiredApp(t, cfg)
	a.OllamaFactory = func(string) llmclient.OllamaAPI {
		return &wireOllama{models: []string{"other-model"},
			chatErr: map[string]error{"qwythos-9b": errors.New("model not found")}}
	}
	saved := map[string]any{}
	a.WriteLiveConfig = func(c *storage.OrderedMap) bool {
		if v, ok := c.Get("ollama"); ok {
			if om, is := v.(*storage.OrderedMap); is {
				m, _ := om.Get("model")
				saved["model"] = m
			}
		}
		return true
	}

	res := a.Commands.SyncRunBatch(context.Background(), BatchOptions{
		URLs: []string{"https://github.com/octo-owner/curated-repo"},
	})
	if !res.Success {
		t.Fatalf("res = %+v\n%s", res, joinLogs(rec))
	}
	j := joinLogs(rec)
	for _, want := range []string{
		"⚠️ Model warmup failed: model not found",
		"🔄 Configured model 'qwythos-9b' not found — Ollama has only 'other-model'. Auto-switching to it (saved to Settings).",
	} {
		if !strings.Contains(j, want) {
			t.Fatalf("missing %q in:\n%s", want, j)
		}
	}
	if saved["model"] != "other-model" {
		t.Fatalf("the live save wrote model=%v", saved["model"])
	}
}

func TestWireWarmupModelPromptDoor(t *testing.T) {
	a := &App{}
	oll := &wireOllama{models: []string{"aaa", "bbb"},
		chatErr: map[string]error{"qwythos-9b": errors.New("boom")}}
	a.OllamaFactory = func(string) llmclient.OllamaAPI { return oll }
	cfg := storage.NewOrderedMap()
	lc := &logCapture{}
	saved := map[string]any{}
	a.WriteLiveConfig = func(c *storage.OrderedMap) bool {
		if v, ok := c.Get("ollama"); ok {
			if om, is := v.(*storage.OrderedMap); is {
				m, _ := om.Get("model")
				saved["model"] = m
			}
		}
		return true
	}
	e := craftedEngine(a, cfg, lc)
	e.ev.ModelPrompt = func(configured string, available []string) string { return "bbb" }

	if e.warmupFamily() {
		t.Fatalf("the ladder aborted: %s", lc.joined())
	}
	j := lc.joined()
	for _, want := range []string{
		"🔄 Switched to 'bbb' (your choice — applied to the rest of the batch and saved).",
		"✅ Replacement model 'bbb' warmed up",
	} {
		if !strings.Contains(j, want) {
			t.Fatalf("missing %q in:\n%s", want, j)
		}
	}
	if saved["model"] != "bbb" || e.warmModel != "bbb" {
		t.Fatalf("model = %v / %v", saved["model"], e.warmModel)
	}
}

func TestWireWarmupPromptDeclinedAutoPicks(t *testing.T) {
	a := &App{}
	oll := &wireOllama{models: []string{"qwythos-9b-tiny", "zzz"},
		chatErr: map[string]error{"qwythos-9b": errors.New("boom")}}
	a.OllamaFactory = func(string) llmclient.OllamaAPI { return oll }
	cfg := storage.NewOrderedMap()
	lc := &logCapture{}
	a.WriteLiveConfig = func(*storage.OrderedMap) bool { return true }
	e := craftedEngine(a, cfg, lc)
	e.ev.ModelPrompt = func(string, []string) string { return "" } // declined
	if e.warmupFamily() {
		t.Fatalf("the ladder aborted: %s", lc.joined())
	}
	if !lc.has("Auto-selected") || !lc.has("as the closest installed stand-in") {
		t.Fatalf("the auto-pick never spoke: %s", lc.joined())
	}
	// PickBestModel ranks the closest cousin first.
	if e.warmModel != "qwythos-9b-tiny" {
		t.Fatalf("warmModel = %q, want the closest stand-in", e.warmModel)
	}
}

func TestWireLlamaCppAbort(t *testing.T) {
	cfg := storage.NewOrderedMap()
	cfg.Set("vault_path", t.TempDir())
	cfg.Set("llm_provider", "llamacpp")
	cfg.Set("llamacpp_api_url", "http://127.0.0.1:9999/v1")
	a, rec := newWiredApp(t, cfg)
	a.ProbeLlama = func(baseURL, apiKey string, timeoutS float64) *llmclient.ProbeResult {
		return &llmclient.ProbeResult{Found: false}
	}
	a.DetectLlama = func(string) *llmclient.ProbeResult { return nil }

	res := a.Commands.SyncRunBatch(context.Background(), BatchOptions{
		URLs: []string{"https://github.com/octo-owner/curated-repo"},
	})
	if res.Success || res.Message != "llama.cpp server not detected" {
		t.Fatalf("res = %+v", res)
	}
	if !joinLogsHas(rec, "❌ No llama.cpp server detected") {
		t.Fatalf("the abort line never spoke: %s", joinLogs(rec))
	}
}

func TestWireLlamaCppDetectAndSwitch(t *testing.T) {
	cfg := storage.NewOrderedMap()
	cfg.Set("vault_path", t.TempDir())
	cfg.Set("llm_provider", "llamacpp")
	cfg.Set("llamacpp_api_url", "http://127.0.0.1:9999/v1")
	a, rec := newWiredApp(t, cfg)
	a.ProbeLlama = func(baseURL, apiKey string, timeoutS float64) *llmclient.ProbeResult {
		return &llmclient.ProbeResult{Found: false}
	}
	ready := true
	pm := "new-model"
	a.DetectLlama = func(string) *llmclient.ProbeResult {
		return &llmclient.ProbeResult{Found: true, BaseURL: "http://127.0.0.1:8080",
			Models: []string{"new-model"}, PropsModel: &pm, Ready: &ready}
	}
	saved := map[string]any{}
	a.WriteLiveConfig = func(c *storage.OrderedMap) bool {
		v, _ := c.Get("llamacpp_model")
		saved["llamacpp_model"] = v
		v2, _ := c.Get("llamacpp_api_url")
		saved["llamacpp_api_url"] = v2
		return true
	}

	res := a.Commands.SyncRunBatch(context.Background(), BatchOptions{
		URLs: []string{"https://github.com/octo-owner/missing-repo"},
	})
	j := joinLogs(rec)
	for _, want := range []string{
		"🔄 llama.cpp not at http://127.0.0.1:9999/v1 — detected at http://127.0.0.1:8080 (switching for this batch and saving to Settings).",
		"🔄 llama.cpp model auto-detected: 'new-model' — saved to Settings.",
		"🦙 Using llama.cpp server http://127.0.0.1:8080 · model 'new-model'",
	} {
		if !strings.Contains(j, want) {
			t.Fatalf("missing %q in:\n%s", want, j)
		}
	}
	if saved["llamacpp_model"] != "new-model" || saved["llamacpp_api_url"] != "http://127.0.0.1:8080/v1" {
		t.Fatalf("the switch never saved: %v", saved)
	}
	if res.Message == "llama.cpp server not detected" {
		t.Fatalf("the batch aborted after the switch: %s", j)
	}
}

func TestWireCloudPreflightVerdicts(t *testing.T) {
	// The missing-model warning + the all-present line + the
	// CloudLLMError shape's ℹ️ line (the reference's two excepts).
	mk := func(list func(url, key string, timeoutS float64) ([]string, error)) (*App, *logCapture, *batchEngine) {
		a := &App{CloudListModels: list}
		cfg := storage.NewOrderedMap()
		cfg.Set("llm_provider", "cloud")
		cfg.Set("cloud_model", "configured-model")
		lc := &logCapture{}
		e := craftedEngine(a, cfg, lc)
		return a, lc, e
	}

	// (a) missing model → the warning.
	_, lc, e := mk(func(string, string, float64) ([]string, error) {
		return []string{"other-model"}, nil
	})
	if e.warmupFamily() {
		t.Fatalf("aborted: %s", lc.joined())
	}
	if !lc.has("⚠️ Endpoint pre-flight: 'configured-model' (model) is NOT in the /models list (1 listed)") {
		t.Fatalf("the missing-model warning never spoke: %s", lc.joined())
	}
	if !lc.has("☁️ Using Cloud API (OpenAI-compatible)") {
		t.Fatalf("the flavor line never spoke: %s", lc.joined())
	}

	// (b) all present → the ✅ line.
	_, lc2, e2 := mk(func(string, string, float64) ([]string, error) {
		return []string{"configured-model", "x"}, nil
	})
	if e2.warmupFamily() {
		t.Fatalf("aborted: %s", lc2.joined())
	}
	if !lc2.has("✅ Endpoint pre-flight: 2 model(s) listed; every configured model is available.") {
		t.Fatalf("the all-present line never spoke: %s", lc2.joined())
	}

	// (c) the CloudLLMError shape → the ℹ️ line (not the ⚠️ one).
	_, lc3, e3 := mk(func(string, string, float64) ([]string, error) {
		return nil, &CloudLLMError{}
	})
	if e3.warmupFamily() {
		t.Fatalf("aborted: %s", lc3.joined())
	}
	if !lc3.has("ℹ️ /models pre-flight unavailable") {
		t.Fatalf("the CloudLLMError line never spoke: %s", lc3.joined())
	}
	if lc3.has("⚠️ Endpoint pre-flight failed") {
		t.Fatalf("the wrong except spoke: %s", lc3.joined())
	}
}

// CloudLLMError carries the class-name token the isCloudLLMError shape
// reads (type(exc).__name__ — the SHAPE is the contract; a local test
// type named for the token, never a real credential).
type CloudLLMError struct{}

func (e *CloudLLMError) Error() string { return "the endpoint answered 503" }

// ---------------------------------------------------------------------------
// The import fetcher (the inputs door).
// ---------------------------------------------------------------------------

func TestWireImportMode(t *testing.T) {
	appDir := t.TempDir()
	t.Setenv("GITCURATOR_HOME", appDir)
	vault := t.TempDir()
	cfg := storage.NewOrderedMap()
	cfg.Set("vault_path", vault)
	// Websites OFF: the non-GitHub links land in the _inbox tables.
	cfgPath := filepath.Join(appDir, "config.json")
	if err := storage.WriteConfigFile(cfgPath, cfg); err != nil {
		t.Fatal(err)
	}
	a := New(cfgPath, false)
	a.GHClientFactory = func(string) githubpipeline.RepoFetcher { return &appFakeFetcher{} }
	a.BannerDownload = func(_, _, _ string) string { return "" }
	a.OllamaFactory = func(string) llmclient.OllamaAPI { return &wireOllama{models: []string{"m"}} }
	rec := &busRec{}
	rec.subscribe(a.Bus)

	imp := filepath.Join(appDir, "links.txt")
	text := "# a comment\n" +
		"https://github.com/octo-owner/curated-repo\n" +
		"https://github.com/octo-owner/curated-repo\n" + // duplicate
		"https://example.com/page\n" +
		"not a link at all\n"
	if err := os.WriteFile(imp, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}

	res := a.Commands.SyncRunBatch(context.Background(), BatchOptions{
		Mode:       "import",
		ImportFile: imp,
		Source:     "import",
	})
	if !res.Success {
		t.Fatalf("res = %+v\n%s", res, joinLogs(rec))
	}
	j := joinLogs(rec)
	for _, want := range []string{
		"📄 Loaded 1 GitHub URLs + 1 website URLs from import file.",
		"🔄 1 duplicate address(es) in the file ignored (3 total → 2 unique)",
		"⚠️ 1 line(s) with no recognizable address skipped:",
		"📋 Manifest created: 1 GitHub + 1 non-GitHub links recorded",
	} {
		if !strings.Contains(j, want) {
			t.Fatalf("missing %q in:\n%s", want, j)
		}
	}
	// The non-GitHub link landed in the GitHub vault's _inbox table.
	entries, err := os.ReadDir(filepath.Join(vault, "_inbox"))
	if err != nil || len(entries) == 0 {
		t.Fatalf("the _inbox table never landed: %v (%v)", entries, err)
	}
}

func TestWireImportFileMissing(t *testing.T) {
	cfg := storage.NewOrderedMap()
	cfg.Set("vault_path", t.TempDir())
	a, rec := newWiredApp(t, cfg)
	a.OllamaFactory = func(string) llmclient.OllamaAPI { return &wireOllama{models: []string{"m"}} }

	res := a.Commands.SyncRunBatch(context.Background(), BatchOptions{
		Mode:       "import",
		ImportFile: filepath.Join(t.TempDir(), "missing.txt"),
	})
	if !res.Success || res.Message != "No URLs found." {
		t.Fatalf("res = %+v", res)
	}
	if !joinLogsHas(rec, "Import file not found.") {
		t.Fatalf("the honest line never spoke: %s", joinLogs(rec))
	}
}

// ---------------------------------------------------------------------------
// The dry-run law.
// ---------------------------------------------------------------------------

func TestWireDryRunSideEffectFree(t *testing.T) {
	appDir := t.TempDir()
	t.Setenv("GITCURATOR_HOME", appDir)
	vault := t.TempDir()
	cfg := storage.NewOrderedMap()
	cfg.Set("vault_path", vault)
	cfg.Set("github_token", "fixture-token")
	cfgPath := filepath.Join(appDir, "config.json")
	if err := storage.WriteConfigFile(cfgPath, cfg); err != nil {
		t.Fatal(err)
	}
	a := New(cfgPath, false)
	a.GHClientFactory = func(string) githubpipeline.RepoFetcher { return &appFakeFetcher{} }
	a.BannerDownload = func(_, _, _ string) string { return "" }
	a.OllamaFactory = func(string) llmclient.OllamaAPI { return &wireOllama{models: []string{"m"}} }
	rec := &busRec{}
	rec.subscribe(a.Bus)

	res := a.Commands.SyncRunBatch(context.Background(), BatchOptions{
		URLs:   []string{"https://github.com/octo-owner/curated-repo"},
		DryRun: true,
	})
	if !res.Success {
		t.Fatalf("res = %+v\n%s", res, joinLogs(rec))
	}
	j := joinLogs(rec)
	if !strings.Contains(j, "🧪 DRY-RUN: every vault write this batch would perform is") {
		t.Fatalf("the banner never spoke: %s", j)
	}
	if !strings.Contains(j, "🧪 DRY-RUN COMPLETE — ") {
		t.Fatalf("the completion line never spoke: %s", j)
	}
	// NO vault writes: no note landed.
	_ = filepath.Walk(vault, func(p string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && strings.HasSuffix(p, ".md") {
			t.Fatalf("a dry-run batch wrote a note: %s", p)
		}
		return nil
	})
	// NO seals, NO mirror push (side-effect free everywhere).
	if strings.Contains(j, "Sealing vault") || strings.Contains(j, "VaultSeal:") ||
		strings.Contains(j, "Good Repos:") {
		t.Fatalf("a dry-run batch ran the seals: %s", j)
	}
	// The report landed under APP_DIR/reports/dry-runs.
	reports := filepath.Join(appDir, "reports", "dry-runs")
	entries, err := os.ReadDir(reports)
	if err != nil || len(entries) == 0 {
		t.Fatalf("the dry-run report never landed: %v (%v)", entries, err)
	}
}

// ---------------------------------------------------------------------------
// The finish chain: the seals in the reference's order.
// ---------------------------------------------------------------------------

func TestWireFinishSealOrder(t *testing.T) {
	cfg := storage.NewOrderedMap()
	vault := t.TempDir()
	cfg.Set("vault_path", vault)
	cfg.Set("github_token", "fixture-token")
	a, rec := newWiredApp(t, cfg)
	a.OllamaFactory = func(string) llmclient.OllamaAPI { return &wireOllama{models: []string{"m"}} }

	res := a.Commands.SyncRunBatch(context.Background(), BatchOptions{
		URLs: []string{"https://github.com/octo-owner/curated-repo"},
	})
	if !res.Success {
		t.Fatalf("res = %+v\n%s", res, joinLogs(rec))
	}
	j := joinLogs(rec)
	// The reference's exact order: VaultSeal → Good Repos → (Websites →
	// Personal: silent no-ops with nothing configured).
	iSeal := strings.Index(j, "Sealing vault (git commit + push if configured)…")
	iVS := strings.Index(j, "VaultSeal:")
	iGR := strings.Index(j, "Good Repos:")
	if iSeal < 0 || iVS < 0 || iGR < 0 {
		t.Fatalf("a seal line is missing:\n%s", j)
	}
	if !(iSeal < iVS && iVS < iGR) {
		t.Fatalf("the seal order broke: %d %d %d", iSeal, iVS, iGR)
	}
}

// ---------------------------------------------------------------------------
// THE PHASE-5 CLEAR (the ack legs + the verdict shapes).
// ---------------------------------------------------------------------------

func TestWireMirrorPushLegs(t *testing.T) {
	vault := t.TempDir()
	cachePath := filepath.Join(t.TempDir(), "cache.db")

	// (a) attempted + ok → the ☁️ line.
	{
		a := &App{}
		cfg := storage.NewOrderedMap()
		cfg.Set("vault_path", vault)
		lc := &logCapture{}
		a.WorkerMirrorPush = func(cfg *storage.OrderedMap, entries,
			decommissionEvents []any, vaultIndexHash string) (bool, bool, any) {
			return true, true, nil
		}
		e := craftedEngine(a, cfg, lc)
		e.vaultIndex = batch.NewVaultIndex(vault, nil)
		e.vaultIndex.Rebuild(nil)
		e.cache = mustCache(t, cachePath)
		out := e.pushVaultMirrorAfterBatch()
		if !out.Attempted || !out.OK || out.Entries != 0 {
			t.Fatalf("out = %+v", out)
		}
		if !lc.has("☁️ Worker synced: 0 vault entry(ies) and 0 decommission event(s) pushed") {
			t.Fatalf("the ack line never spoke: %s", lc.joined())
		}
	}
	// (b) attempted + failed → the ⚠️ line.
	{
		a := &App{}
		cfg := storage.NewOrderedMap()
		cfg.Set("vault_path", vault)
		lc := &logCapture{}
		a.WorkerMirrorPush = func(cfg *storage.OrderedMap, entries,
			decommissionEvents []any, vaultIndexHash string) (bool, bool, any) {
			return true, false, "the worker answered 503"
		}
		e := craftedEngine(a, cfg, lc)
		e.vaultIndex = batch.NewVaultIndex(vault, nil)
		e.vaultIndex.Rebuild(nil)
		out := e.pushVaultMirrorAfterBatch()
		if out.Attempted && out.OK {
			t.Fatalf("out = %+v", out)
		}
		if !lc.has(`⚠️ Vault-mirror push failed: "the worker answered 503"`) {
			t.Fatalf("the failure line never spoke: %s", lc.joined())
		}
	}
	// (c) the disabled Worker is silent.
	{
		a := &App{}
		cfg := storage.NewOrderedMap()
		lc := &logCapture{}
		e := craftedEngine(a, cfg, lc)
		e.vaultIndex = batch.NewVaultIndex(vault, nil)
		e.vaultIndex.Rebuild(nil)
		out := e.pushVaultMirrorAfterBatch()
		if out.Attempted {
			t.Fatalf("out = %+v", out)
		}
		if lc.has("Worker synced") || lc.has("Vault-mirror push failed") {
			t.Fatalf("the disabled worker spoke: %s", lc.joined())
		}
	}
	// (d) a nil index pushes nothing.
	{
		a := &App{}
		cfg := storage.NewOrderedMap()
		lc := &logCapture{}
		e := craftedEngine(a, cfg, lc)
		if out := e.pushVaultMirrorAfterBatch(); out.Attempted {
			t.Fatalf("a nil index pushed: %+v", out)
		}
	}
}

func TestWirePhaseFiveClearVerdicts(t *testing.T) {
	vault := t.TempDir()

	// (a) all verified + bot source → the ✓ verdict.
	{
		a := &App{}
		cfg := storage.NewOrderedMap()
		lc := &logCapture{}
		e := craftedEngine(a, cfg, lc)
		e.tracker = batch.NewLinkTracker(vault)
		e.tracker.SetSource("bot")
		e.tracker.Intake([]string{"https://github.com/o/r"}, nil)
		e.tracker.MarkProcessed("https://github.com/o/r", "note.md")
		e.phaseFiveClear(true, true)
		if !lc.has("✓ Bot queue synced — all links verified (Phase 5 CLEAR; the vault-mirror push is the ack).") {
			t.Fatalf("the all-clear never spoke: %s", lc.joined())
		}
	}
	// (b) unverified links → the ⏸️ verdict with the count.
	{
		a := &App{}
		cfg := storage.NewOrderedMap()
		lc := &logCapture{}
		e := craftedEngine(a, cfg, lc)
		e.tracker = batch.NewLinkTracker(vault)
		e.tracker.SetSource("bot")
		e.tracker.Intake([]string{"https://github.com/o/a", "https://github.com/o/b"}, nil)
		e.phaseFiveClear(true, true)
		if !lc.has("⏸️ 2 link(s) not verified — they stay PENDING in the Worker queue") {
			t.Fatalf("the pending verdict never spoke: %s", lc.joined())
		}
	}
	// (c) no tracker → the honest no-manifest line.
	{
		a := &App{}
		cfg := storage.NewOrderedMap()
		lc := &logCapture{}
		e := craftedEngine(a, cfg, lc)
		e.phaseFiveClear(true, true)
		if !lc.has("⏸️ No link manifest for this batch") {
			t.Fatalf("the no-manifest line never spoke: %s", lc.joined())
		}
	}
	// (d) a non-bot batch never speaks the verdict.
	{
		a := &App{}
		cfg := storage.NewOrderedMap()
		lc := &logCapture{}
		e := craftedEngine(a, cfg, lc)
		e.tracker = batch.NewLinkTracker(vault)
		e.tracker.SetSource("direct")
		e.tracker.Intake([]string{"https://github.com/o/r"}, nil)
		e.phaseFiveClear(false, true)
		if lc.has("Phase 5 CLEAR") {
			t.Fatalf("a failed/non-bot batch spoke the verdict: %s", lc.joined())
		}
	}
}

// ---------------------------------------------------------------------------
// THE PHASE-5 CLEAR end-to-end through the live composition (the bot
// batch's own door).
// ---------------------------------------------------------------------------

func TestWirePhaseFiveClearEndToEnd(t *testing.T) {
	cfg := storage.NewOrderedMap()
	vault := t.TempDir()
	cfg.Set("vault_path", vault)
	cfg.Set("github_token", "fixture-token")
	a, rec := newWiredApp(t, cfg)
	a.OllamaFactory = func(string) llmclient.OllamaAPI { return &wireOllama{models: []string{"m"}} }
	pushCalls := 0
	a.WorkerMirrorPush = func(cfg *storage.OrderedMap, entries,
		decommissionEvents []any, vaultIndexHash string) (bool, bool, any) {
		pushCalls++
		if vaultIndexHash == "" {
			t.Fatal("the index hash was empty")
		}
		return true, true, nil
	}

	res := a.Commands.SyncRunBatch(context.Background(), BatchOptions{
		URLs:   []string{"https://github.com/octo-owner/curated-repo"},
		Source: "bot",
	})
	if !res.Success {
		t.Fatalf("res = %+v\n%s", res, joinLogs(rec))
	}
	if pushCalls != 1 {
		t.Fatalf("the mirror push ran %d times", pushCalls)
	}
	j := joinLogs(rec)
	if !strings.Contains(j, "☁️ Worker synced:") {
		t.Fatalf("the ack never spoke: %s", j)
	}
	if !strings.Contains(j, "✓ Bot queue synced — all links verified (Phase 5 CLEAR") {
		t.Fatalf("the all-clear never spoke: %s", j)
	}
}

// ---------------------------------------------------------------------------
// The websites phase's wiring.
// ---------------------------------------------------------------------------

func TestWireWebsitesPhaseRuns(t *testing.T) {
	appDir := t.TempDir()
	t.Setenv("GITCURATOR_HOME", appDir)
	vault := t.TempDir()
	wvault := t.TempDir()
	cfg := storage.NewOrderedMap()
	cfg.Set("vault_path", vault)
	cfg.Set("github_token", "fixture-token")
	cfg.Set("website_vault_path", wvault)
	cfg.Set("taxonomy_path", "../taxonomy/testdata/website-library-categories.md")
	pipes := storage.NewOrderedMap()
	pipes.Set("github", true)
	pipes.Set("websites", true)
	cfg.Set("pipelines", pipes)
	cfgPath := filepath.Join(appDir, "config.json")
	if err := storage.WriteConfigFile(cfgPath, cfg); err != nil {
		t.Fatal(err)
	}
	a := New(cfgPath, false)
	a.GHClientFactory = func(string) githubpipeline.RepoFetcher { return &appFakeFetcher{} }
	a.BannerDownload = func(_, _, _ string) string { return "" }
	a.OllamaFactory = func(string) llmclient.OllamaAPI { return &wireOllama{models: []string{"m"}} }
	rec := &busRec{}
	rec.subscribe(a.Bus)

	res := a.Commands.SyncRunBatch(context.Background(), BatchOptions{
		URLs: []string{"https://github.com/octo-owner/curated-repo"},
	})
	if !res.Success {
		t.Fatalf("res = %+v\n%s", res, joinLogs(rec))
	}
	j := joinLogs(rec)
	if !strings.Contains(j, "🌐 Websites pipeline: 0 link(s)") {
		t.Fatalf("the phase never announced itself: %s", j)
	}
	if !strings.Contains(j, "🌐 Websites done: 0 processed") {
		t.Fatalf("the phase never finished: %s", j)
	}
}

func TestWireWebsitesNoVaultKeepsBatch(t *testing.T) {
	cfg := storage.NewOrderedMap()
	vault := t.TempDir()
	cfg.Set("vault_path", vault)
	cfg.Set("github_token", "fixture-token")
	pipes := storage.NewOrderedMap()
	pipes.Set("github", true)
	pipes.Set("websites", true)
	cfg.Set("pipelines", pipes)
	a, rec := newWiredApp(t, cfg)
	a.OllamaFactory = func(string) llmclient.OllamaAPI { return &wireOllama{models: []string{"m"}} }

	res := a.Commands.SyncRunBatch(context.Background(), BatchOptions{
		URLs:      []string{"https://github.com/octo-owner/curated-repo"},
		NonGithub: []string{"https://example.com/page"},
	})
	if !res.Success {
		t.Fatalf("res = %+v\n%s", res, joinLogs(rec))
	}
	if !joinLogsHas(rec, "⚠️ Websites pipeline is ON but no Websites vault is set") {
		t.Fatalf("the honest warning never spoke: %s", joinLogs(rec))
	}
}

// mustCache opens a throwaway cache for the mirror-push legs.
func mustCache(t *testing.T, path string) *store.CacheDB {
	c, err := store.NewCacheDB(path)
	if err != nil {
		t.Fatal(err)
	}
	return c
}
