// heroui_test.go — T-77's chrome binding wrapper wiring tests: the
// hero click door's guards (the vault box, the nothing-to-sync split,
// the offline advisory's ask + the owner's-word continuation), the
// fetch flow over faked queue seams (the fetching render, the fetched
// routing with the 🟢 line, the caught-up path's closing line + the
// drift seal, the error restore), the process confirm gate + the
// batch's running→SYNC restore, the battery door's four rows + the
// all-ok gate + the busy refusal, the sentinel seat, and the crossing
// payloads' field tags.
package main

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

	"github.com/assadigit/second-brain-project/internal/app"
	"github.com/assadigit/second-brain-project/internal/connectioncheck"
	"github.com/assadigit/second-brain-project/internal/constants"
	"github.com/assadigit/second-brain-project/internal/githubpipeline"
	"github.com/assadigit/second-brain-project/internal/handdelivery"
	"github.com/assadigit/second-brain-project/internal/llmclient"
	"github.com/assadigit/second-brain-project/internal/mirror"
	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/store"
	"github.com/assadigit/second-brain-project/internal/websitepipeline"
)

// heroRig is the shared wiring harness: faked queue/battery seams,
// a log-capturing subscriber, and the HeroUI wrapper.
type heroRig struct {
	ui       *HeroUI
	commands *app.Commands
	cfg      *storage.OrderedMap
	spoken   *[]string
	mu       *sync.Mutex

	queueState string
	queueItems []any
	batchCalls int
	batchModes []string
	// batchOpts snapshots every recorded batch's full options (the
	// routing law's pin: the URLs/NonGithub split, never just the mode).
	batchOpts   []app.BatchOptions
	batchResult *app.BatchResult
	batchBlock  chan struct{} // non-nil → the fake batch waits for the close

	// botQueueItems feeds the Settings' BotQueueFetch seam (T-79c's
	// doors — independent of the Commands' queue fakes).
	botQueueItems []any

	// The Chrome doors' fakes (T-79d): the fifth-door delivery counter
	// + the gate (nil = no block), and the hand-enqueue seam.
	chromeDeliverCalls int
	deliverGate        chan struct{}
	handEnqueueFn      func(vault string, urls []string,
		walls map[string]string, log func(msg, level string),
		openChrome bool) (int, int)
}

func newHeroRig(t *testing.T) *heroRig {
	t.Helper()
	appDir := t.TempDir()
	t.Setenv("GITCURATOR_HOME", appDir)
	bus := app.NewBus()
	cfg := storage.NewOrderedMap()
	cfg.Set("vault_path", t.TempDir())
	settings := app.NewSettings(bus, cfg, &app.OpLock{}, app.SettingsDeps{})
	var spoken []string
	var mu sync.Mutex
	bus.Subscribe(&app.Handlers{LogLine: func(l app.LogLine) {
		mu.Lock()
		spoken = append(spoken, l.Msg)
		mu.Unlock()
	}})

	rig := &heroRig{cfg: cfg, spoken: &spoken, mu: &mu, queueState: "ok"}
	rig.commands = app.NewCommands(bus, app.Deps{
		Config: func() *storage.OrderedMap { return cfg },
		QueueFetch: func(cfg *storage.OrderedMap) (string, []any) {
			return rig.queueState, rig.queueItems
		},
		QueueClassify: func(cfg *storage.OrderedMap, items []any,
			log func(msg, level string)) *app.QueueReport {
			rep := &app.QueueReport{State: "ok", Success: true}
			if rig.queueState != "ok" {
				rep.State = rig.queueState
				rep.Success = false
			}
			return rep
		},
		Batch: func(ctx context.Context, cfg *storage.OrderedMap,
			opts app.BatchOptions, ev app.BatchEvents) *app.BatchResult {
			rig.mu.Lock()
			rig.batchCalls++
			rig.batchModes = append(rig.batchModes, opts.Mode+"/"+opts.Source)
			rig.batchOpts = append(rig.batchOpts, app.BatchOptions{
				URLs:             append([]string(nil), opts.URLs...),
				NonGithub:        append([]string(nil), opts.NonGithub...),
				Source:           opts.Source,
				Mode:             opts.Mode,
				ImportFile:       opts.ImportFile,
				IntakeDuplicates: opts.IntakeDuplicates,
				RawURLCount:      opts.RawURLCount,
				DryRun:           opts.DryRun,
			})
			block := rig.batchBlock
			result := rig.batchResult
			rig.mu.Unlock()
			if block != nil {
				select {
				case <-block:
				case <-ctx.Done():
				}
			}
			if result != nil {
				return result
			}
			return &app.BatchResult{Success: true, Message: "the fake batch finished"}
		},
	})
	rig.ui = NewHeroUI(settings, rig.commands)
	return rig
}

func (r *heroRig) logs() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return strings.Join(*r.spoken, "\n")
}

// ---------------------------------------------------------------------------
// The click guards.
// ---------------------------------------------------------------------------

func TestHeroClickVaultGuardBox(t *testing.T) {
	rig := newHeroRig(t)
	rig.cfg.Set("vault_path", "/not/a/real/dir")
	res := rig.ui.HeroClick("")
	if res.Box == nil || res.Box.Title != "Error" ||
		res.Box.Body != "Please select a valid Obsidian vault path." {
		t.Fatalf("the vault guard = %+v", res.Box)
	}
	if res.State != app.HeroSync {
		t.Fatalf("a guard box never moves the state; got %q", res.State)
	}
}

func TestHeroClickNothingToSyncSplit(t *testing.T) {
	rig := newHeroRig(t)
	// No Worker queue configured + no import file → the nothing box
	// (re-worded for Option B's queue).
	res := rig.ui.HeroClick("")
	if res.Box == nil || res.Box.Title != "Nothing to sync" ||
		!strings.Contains(res.Box.Body, "Settings → Bot") {
		t.Fatalf("the nothing box = %+v", res.Box)
	}
	// An import file picked → the import batch (v0.23.0's ONE other
	// input mode).
	res = rig.ui.HeroClick("/tmp/links.txt")
	if res.Box != nil {
		t.Fatalf("the import split must run, not box: %+v", res.Box)
	}
	rig.mu.Lock()
	calls := rig.batchCalls
	rig.mu.Unlock()
	if calls != 1 {
		t.Fatalf("the import batch never ran (%d calls)", calls)
	}
}

func TestHeroClickOfflineAdvisoryAndContinuation(t *testing.T) {
	rig := newHeroRig(t)
	offline := false
	rig.ui.applySentinel(app.Sentinel{
		Net: appSentinelNet(offline), Verdict: "warn-now",
	})
	res := rig.ui.HeroClick("")
	if !res.Offline {
		t.Fatalf("the offline ask must fire before the fetch: %+v", res)
	}
	// "No" stays in SYNC — nothing runs, nothing asks (hero.py:150:
	// "stay in SYNC state; retry when back online").
	res = rig.ui.HeroAnswerOffline(false, "")
	if res.Offline || res.Box != nil || res.State != app.HeroSync {
		t.Fatalf("a refused ask stays put: %+v", res)
	}
	rig.mu.Lock()
	calls := rig.batchCalls
	rig.mu.Unlock()
	if calls != 0 {
		t.Fatalf("a refused ask never runs anything (%d batch calls)", calls)
	}
	// "Continue anyway" speaks the owner's-word line + fetches (the
	// queue must be configured for the continuation to reach stage 1 —
	// the reference's creds guard).
	rig.cfg.Set("cloudflare_enabled", true)
	rig.cfg.Set("cloudflare_worker_url", "https://worker.example")
	rig.cfg.Set("cloudflare_install_id", "install")
	rig.cfg.Set("cloudflare_shared_secret", "secret")
	rig.ui.applySentinel(app.Sentinel{
		Net: appSentinelNet(offline), Verdict: "offline",
	})
	res = rig.ui.HeroAnswerOffline(true, "")
	if !strings.Contains(rig.logs(), "⚠️ Syncing despite the offline verdict (the owner's word wins).") {
		t.Fatalf("the continuation line missing:\n%s", rig.logs())
	}
	if !strings.Contains(rig.logs(), "✅ All caught up — nothing undone in the bot queue.") {
		t.Fatalf("the caught-up closing line missing:\n%s", rig.logs())
	}
}

// appSentinelNet builds the probe shape (the connectioncheck type).
func appSentinelNet(online bool) connectioncheck.ProbeInternetResult {
	return connectioncheck.ProbeInternetResult{Online: online}
}

// ---------------------------------------------------------------------------
// The fetch flow.
// ---------------------------------------------------------------------------

func TestHeroFetchFlowFetchedQueue(t *testing.T) {
	rig := newHeroRig(t)
	// A configured Worker queue (the fetch's gate): the pairing keys.
	rig.cfg.Set("cloudflare_enabled", true)
	rig.cfg.Set("cloudflare_worker_url", "https://worker.example")
	rig.cfg.Set("cloudflare_install_id", "install")
	rig.cfg.Set("cloudflare_shared_secret", "secret")

	// The classify fake needs the pending lists: extend the rig's fake
	// through the report.
	pending := []string{"https://github.com/octo/one"}
	pendingWeb := []string{"https://example.com/page"}
	rig.commands = app.NewCommands(rig.commands.Bus, app.Deps{
		Config: func() *storage.OrderedMap { return rig.cfg },
		QueueFetch: func(cfg *storage.OrderedMap) (string, []any) {
			return "ok", []any{}
		},
		QueueClassify: func(cfg *storage.OrderedMap, items []any,
			log func(msg, level string)) *app.QueueReport {
			return &app.QueueReport{State: "ok", Success: true,
				PendingURLs: pending, PendingWebsites: pendingWeb}
		},
		Batch: rig.commands.Deps.Batch,
	})
	rig.ui = NewHeroUI(rig.ui.settings, rig.commands)

	res := rig.ui.HeroClick("")
	if res.State != app.HeroProcess || res.Format != "2 ready to process" {
		t.Fatalf("the fetch routing = %q / %q", res.State, res.Format)
	}
	if res.Count != "0 / 2" {
		t.Fatalf("the count = %q", res.Count)
	}
	if !strings.Contains(rig.logs(),
		"🟢 Fetched 1 repo(s) + 1 website link(s) — click PROCESS to start.") {
		t.Fatalf("the fetched line missing:\n%s", rig.logs())
	}
	if res.Render.Label != "PROCESS (2)" {
		t.Fatalf("the render label = %q", res.Render.Label)
	}

	// PROCESS with ≤10 items runs the batch directly (no confirm).
	res = rig.ui.HeroClick("")
	if res.Confirm != nil {
		t.Fatalf("a small queue never asks: %+v", res.Confirm)
	}
	rig.mu.Lock()
	calls, modes := rig.batchCalls, rig.batchModes
	rig.mu.Unlock()
	if calls != 1 || modes[0] != "/bot" {
		t.Fatalf("the queue batch = %d calls, %v (the engine's empty mode IS direct)", calls, modes)
	}
	if res.State != app.HeroSync {
		t.Fatalf("the finish restores SYNC; got %q", res.State)
	}
}

func TestHeroFetchErrorRestore(t *testing.T) {
	rig := newHeroRig(t)
	rig.cfg.Set("cloudflare_enabled", true)
	rig.cfg.Set("cloudflare_worker_url", "https://worker.example")
	rig.cfg.Set("cloudflare_install_id", "i")
	rig.cfg.Set("cloudflare_shared_secret", "s")
	rig.queueState = "error"
	res := rig.ui.HeroClick("")
	if res.State != app.HeroSync || res.Format != "Ready" {
		t.Fatalf("the error restore = %q / %q", res.State, res.Format)
	}
	if !strings.Contains(rig.logs(), "❌ Bot queue check failed:") {
		t.Fatalf("the honest failure line missing:\n%s", rig.logs())
	}
}

func TestHeroProcessConfirmGate(t *testing.T) {
	rig := newHeroRig(t)
	rig.cfg.Set("cloudflare_enabled", true)
	rig.cfg.Set("cloudflare_worker_url", "https://w.example")
	rig.cfg.Set("cloudflare_install_id", "i")
	rig.cfg.Set("cloudflare_shared_secret", "s")
	eleven := make([]string, 11)
	for i := range eleven {
		eleven[i] = "https://github.com/octo/repo" + string(rune('a'+i))
	}
	rig.commands = app.NewCommands(rig.commands.Bus, app.Deps{
		Config:     func() *storage.OrderedMap { return rig.cfg },
		QueueFetch: func(cfg *storage.OrderedMap) (string, []any) { return "ok", nil },
		QueueClassify: func(cfg *storage.OrderedMap, items []any, log func(string, string)) *app.QueueReport {
			return &app.QueueReport{State: "ok", Success: true, PendingURLs: eleven}
		},
		Batch: rig.commands.Deps.Batch,
	})
	rig.ui = NewHeroUI(rig.ui.settings, rig.commands)

	res := rig.ui.HeroClick("") // the fetch
	if res.State != app.HeroProcess {
		t.Fatalf("the fetch routing = %q", res.State)
	}
	res = rig.ui.HeroClick("") // PROCESS with 11 → the confirm gate
	if res.Confirm == nil || res.Confirm.Title != "Confirm Large Batch" ||
		res.Confirm.Body != "This will process 11 items (the bot queue).\n\nContinue?" {
		t.Fatalf("the confirm gate = %+v", res.Confirm)
	}
	rig.mu.Lock()
	calls := rig.batchCalls
	rig.mu.Unlock()
	if calls != 0 {
		t.Fatalf("an unanswered confirm never runs the batch (%d calls)", calls)
	}
	// Continue → the confirmed door runs it.
	_ = rig.ui.HeroRunQueue()
	rig.mu.Lock()
	calls = rig.batchCalls
	rig.mu.Unlock()
	if calls != 1 {
		t.Fatalf("the confirmed batch never ran (%d calls)", calls)
	}
}

// ---------------------------------------------------------------------------
// The battery door.
// ---------------------------------------------------------------------------

func TestBatteryDoorFourRowsAndGate(t *testing.T) {
	rig := newHeroRig(t)
	res := rig.ui.TestConnection()
	if res.Refusal != "" {
		t.Fatalf("the battery refused: %s", res.Refusal)
	}
	if len(res.Rows) != 4 {
		t.Fatalf("the battery rows = %d, want the four", len(res.Rows))
	}
	wantTitles := []string{"Vaults", "LLM", "GitHub", "Worker"}
	wantIcons := []string{"📁", "🧠", "🐙", "☁️"}
	for i, row := range res.Rows {
		if row.Title != wantTitles[i] || row.Icon != wantIcons[i] {
			t.Fatalf("row[%d] = %s %s", i, row.Icon, row.Title)
		}
		if row.Verdict == "" {
			t.Fatalf("row[%d] has no verdict", i)
		}
	}
	if !strings.Contains(rig.logs(), "🏁 Test Connection — ") {
		t.Fatalf("the verdict line missing:\n%s", rig.logs())
	}
	// The gate verdict is honest (a default config is not all-ok: no
	// LLM, no GitHub token).
	if res.AllOK {
		t.Fatalf("an unconfigured battery must not say all-ok")
	}
}

func TestBatteryDoorBusyRefusal(t *testing.T) {
	rig := newHeroRig(t)
	// A REAL running batch holds the one-batch lock → the ⏳ refusal
	// (the honest busy state, never a faked flag).
	block := make(chan struct{})
	rig.batchBlock = block
	go func() { _ = rig.ui.HeroRunQueue() }()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if rig.commands.BatchBusy() {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if !rig.commands.BatchBusy() {
		t.Fatal("the fake batch never went busy")
	}
	res := rig.ui.TestConnection()
	if res.Refusal == "" || !strings.Contains(res.Refusal, "A batch is running") {
		t.Fatalf("the busy refusal = %q", res.Refusal)
	}
	close(block)
}

// ---------------------------------------------------------------------------
// The sentinel seat + the crossing payloads' field tags.
// ---------------------------------------------------------------------------

func TestHeroSentinelSeatFeedsTheAdvisory(t *testing.T) {
	rig := newHeroRig(t)
	online := true
	rig.ui.applySentinel(app.Sentinel{Net: appSentinelNetStruct(true), Verdict: "first-online"})
	_ = online
	// An online verdict never asks.
	rig.cfg.Set("cloudflare_enabled", true)
	rig.cfg.Set("cloudflare_worker_url", "https://w.example")
	rig.cfg.Set("cloudflare_install_id", "i")
	rig.cfg.Set("cloudflare_shared_secret", "s")
	if res := rig.ui.HeroClick(""); res.Offline {
		t.Fatalf("an online sentinel must not ask")
	}
	offline := false
	rig.ui.applySentinel(app.Sentinel{Net: appSentinelNetStruct(false), Verdict: "fell-offline"})
	_ = offline
	if res := rig.ui.HeroClick(""); !res.Offline {
		t.Fatalf("a seen-offline verdict must arm the advisory")
	}
}

// appSentinelNetStruct builds the net shape with a latency pointer.
func appSentinelNetStruct(online bool) connectioncheck.ProbeInternetResult {
	out := connectioncheck.ProbeInternetResult{Online: online}
	if online {
		ms := 42
		out.LatencyMs = &ms
	}
	return out
}

func TestHeroCrossingPayloadFieldTags(t *testing.T) {
	rig := newHeroRig(t)
	// The state payload's keys are the frontend's reads (THE
	// FIELD-TAG LAW).
	raw, err := json.Marshal(rig.ui.HeroState())
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{`"state":"sync"`, `"render":`, `"label":"SYNC"`, `"glyph":"refresh"`} {
		if !strings.Contains(string(raw), key) {
			t.Fatalf("the state payload missing %s: %s", key, raw)
		}
	}
	// The hero box's keys.
	raw, err = json.Marshal(&app.HeroBox{Title: "t", Body: "b"})
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != `{"title":"t","body":"b"}` {
		t.Fatalf("the hero box payload = %s", raw)
	}
}

// ---------------------------------------------------------------------------
// The mirror + the stop leg (the winding-down render; the batch-finished
// restore).
// ---------------------------------------------------------------------------

func TestHeroStopLegAndMirrorRestore(t *testing.T) {
	rig := newHeroRig(t)
	// A blocking batch: the Stop click lands while it runs.
	block := make(chan struct{})
	rig.batchBlock = block
	rig.cfg.Set("cloudflare_enabled", true)
	rig.cfg.Set("cloudflare_worker_url", "https://w.example")
	rig.cfg.Set("cloudflare_install_id", "i")
	rig.cfg.Set("cloudflare_shared_secret", "s")
	pending := []string{"https://github.com/octo/one"}
	rig.commands = app.NewCommands(rig.commands.Bus, app.Deps{
		Config:     func() *storage.OrderedMap { return rig.cfg },
		QueueFetch: func(cfg *storage.OrderedMap) (string, []any) { return "ok", nil },
		QueueClassify: func(cfg *storage.OrderedMap, items []any,
			log func(string, string)) *app.QueueReport {
			return &app.QueueReport{State: "ok", Success: true, PendingURLs: pending}
		},
		Batch: rig.commands.Deps.Batch,
	})
	rig.ui = NewHeroUI(rig.ui.settings, rig.commands)

	go func() { _ = rig.ui.HeroClick("") }() // fetch → process
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		rig.ui.mu.Lock()
		st := rig.ui.machine.State()
		rig.ui.mu.Unlock()
		if st == app.HeroProcess {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	go func() { _ = rig.ui.HeroRunQueue() }()
	deadline = time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		rig.ui.mu.Lock()
		st := rig.ui.machine.State()
		rig.ui.mu.Unlock()
		if st == app.HeroRunning {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	rig.ui.mu.Lock()
	st := rig.ui.machine.State()
	rig.ui.mu.Unlock()
	if st != app.HeroRunning {
		t.Fatalf("the batch must flip the machine to running; got %q", st)
	}

	// The Stop leg: the winding-down render (Stopping…, disabled).
	res := rig.ui.HeroClick("")
	if res.Render.Label != "Stopping…" || res.Render.Enabled {
		t.Fatalf("the winding-down render = %+v", res.Render)
	}
	if !strings.Contains(rig.logs(), "⏹️ Stopping...") {
		t.Fatalf("the stop line missing:\n%s", rig.logs())
	}
	// The batch finishes → the deterministic restore.
	close(block)
	deadline = time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		rig.ui.mu.Lock()
		st = rig.ui.machine.State()
		rig.ui.mu.Unlock()
		if st == app.HeroSync {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if st != app.HeroSync {
		t.Fatalf("the finish must restore SYNC; got %q", st)
	}
}

// ---------------------------------------------------------------------------
// The frontend's own laws, read where they render (heroflow.js).
// ---------------------------------------------------------------------------

func heroflowSource(t *testing.T) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("frontend", "heroflow.js"))
	if err != nil {
		t.Fatal(err)
	}
	// the desk's own module carries the batch-finished log row (the
	// summary read) — the chrome's laws span both files.
	desk, err := os.ReadFile(filepath.Join("frontend", "app.js"))
	if err != nil {
		t.Fatal(err)
	}
	return string(raw) + "\n" + string(desk)
}

func TestHeroflowReadsTheTaggedNames(t *testing.T) {
	src := heroflowSource(t)
	// THE FIELD-TAG LAW: the frontend reads the TAGGED names, never the
	// camelCase guesses.
	for _, read := range []string{
		"payload.render", "r.label", "r.enabled", "r.danger", "r.tooltip", "r.glyph",
		"payload.format", "payload.count",
		"st.text", "p.pos", "p.total", "jp.label", "jp.value", "jp.max",
		"bf.ok", "bf.summary", "bf.chime", "bf.chime.play", "bf.chime.dataUrl", "bf.chime.volume",
		"ev.net.online", "ev.net.detail", "ai.state", "ai.kind", "ai.engine",
		"ai.short", "ai.detail", "px.state", "px.detail", "ev.verdict",
		"res.rows", "res.refusal", "boot.dark_mode",
		// T-79a: the per-test doors' verdict-box reads (MoreDoorResult's
		// tagged names — the same shape the dashboard door rides).
		"res.title", "res.body",
	} {
		if !strings.Contains(src, read) {
			t.Fatalf("heroflow.js must read %q (the field-tag law)", read)
		}
	}
}

func TestHeroflowHeroMachineLaws(t *testing.T) {
	src := heroflowSource(t)
	// the danger fill + the disabled law ride the render.
	if !strings.Contains(src, "btn-hero-danger") || !strings.Contains(src, "heroBtn.disabled = !r.enabled") {
		t.Fatal("the render must carry the danger fill + the disabled flag")
	}
	// the offline box's own words (the reference's continue-anyway text).
	for _, want := range []string{
		"⚠️ No internet connection",
		"The 'Net' indicator says you are offline",
		"Continue anyway?",
	} {
		if !strings.Contains(src, want) {
			t.Fatalf("the offline ask missing %q", want)
		}
	}
	// the confirm gate renders the Go-provided box (never its own words).
	if !strings.Contains(src, "res.Confirm") || strings.Contains(src, "Confirm Large Batch") {
		t.Fatal("the confirm gate renders the engine's box")
	}
	// the chime never errors (the QSoundEffect twin's law).
	if !strings.Contains(src, "never an error") {
		t.Fatal("the chime's never-an-error law")
	}
}

func TestHeroflowBatteryModalLaws(t *testing.T) {
	src := heroflowSource(t)
	// the four rows' lockup (the Worker leg promoted per ADR-0004 Am.3).
	for _, want := range []string{
		"icon: '📁', title: 'Vaults'", "icon: '🧠', title: 'LLM'",
		"icon: '🐙', title: 'GitHub'", "icon: '☁️', title: 'Worker'",
	} {
		if !strings.Contains(src, want) {
			t.Fatalf("the battery lockup missing %q", want)
		}
	}
	// Start Syncing born disabled; green only when every row is ok.
	if !strings.Contains(src, "startB.disabled = true") ||
		!strings.Contains(src, "r.done && r.verdict === 'ok'") {
		t.Fatal("the Start Syncing gate law")
	}
	// the row lifecycle: waiting → checking (the spinner) → the verdict
	// words (Connected / Connected (warnings) / Not connected).
	for _, want := range []string{"⏳ Waiting…", "Checking…", "'Connected'", "'Connected (warnings)'", "'Not connected'"} {
		if !strings.Contains(src, want) {
			t.Fatalf("the battery row words missing %q", want)
		}
	}
}

func TestHeroflowMoreMenuLabelList(t *testing.T) {
	src := heroflowSource(t)
	// the reference's own action list, verbatim (ui.py:1854-1944 — the
	// parity requirement; the Option-B re-scopes are the exceptions).
	for _, want := range []string{
		"🔌 Test Connection (all systems)",
		"🔑 Test Telegram & GitHub",
		"🌐 Test Proxy Connection",
		"🧠 Test Ollama",
		"🔌 Test Cloud API",
		"🦙 Test llama.cpp",
		"🪝 Recall hooks (dry-run)",
		"🔗 Build link suggestions",
		"✅ Validate Vault",
		"🔍 Verify Vault",
		"📁 Recategorize Notes",
		"✅ Verify All Processed",
		"📋 Export All Links",
		"🔄 Retry Failed",
		"🚫 View 404 Quarantine",
		"🔁 Retry _review backlog",
		"🔁 Retry the table\\'s \\' - \\' rows",
		"🪦 Decommission dead links",
		"🖐 Hand-deliver walled links",
		"🖐 Scrape hand rows via Chrome",
		"🤖 Chrome tab-retry failed links",
		"🪄 Attach to my Chrome",
		"📤 Export Backup ZIP",
		"📊 Open Dashboard",
		"⚙️ Settings",
		"🌙 Dark Mode",
	} {
		if !strings.Contains(src, want) {
			t.Fatalf("the More menu's verbatim list missing %q", want)
		}
	}
	// the honest-disabled law: the not-yet-landed engines are present
	// AND disabled with their landing note.
	if !strings.Contains(src, "more-item-soon") || !strings.Contains(src, "item.disabled = true") {
		t.Fatal("the honest-disabled law")
	}
	// the reference's own menu button words.
	if !strings.Contains(src, "'More ▾'") {
		t.Fatal("the More button's own words")
	}
}

func TestHeroflowSentinelPaintLaws(t *testing.T) {
	src := heroflowSource(t)
	// the words + the dots + the semantics (window.py's paint laws —
	// the labels never read bare dashes).
	for _, want := range []string{
		"'Net: ' + (ev.net && ev.net.online ? 'online' : 'offline')",
		// THE CONCISE-AI-LABEL LAW (SESSION-076): the chrome's word is
		// the bare "AI: on"; the engine kind/name rides the tooltip only.
		"aiWord = 'AI: on'", "'cloud provider'", "'local ('",
		"'Proxy: connected'", "'Proxy: off'", "'Proxy: unreachable'",
		"#42B36B", "#E85D75", "#D9A036", "#8E8A90",
	} {
		if !strings.Contains(src, want) {
			t.Fatalf("the sentinel paint law missing %q", want)
		}
	}
	// the warn-now box (the ONE modal at open).
	if !strings.Contains(src, "verdict === 'warn-now'") ||
		!strings.Contains(src, "GitCurator could not reach the internet") {
		t.Fatal("the offline warn-now box law")
	}
}

// ---------------------------------------------------------------------------
// T-79a — the five per-test doors (the More menu's per-test dialogs).
// ---------------------------------------------------------------------------

// newPerTestRig builds a HeroUI over a Settings whose per-test seams
// are all faked (the hermetic law) + a log-capturing bus.
func newPerTestRig(t *testing.T) (*HeroUI, *storage.OrderedMap, *[]string, *sync.Mutex) {
	t.Helper()
	appDir := t.TempDir()
	t.Setenv("GITCURATOR_HOME", appDir)
	bus := app.NewBus()
	cfg := storage.NewOrderedMap()
	var spoken []string
	var mu sync.Mutex
	bus.Subscribe(&app.Handlers{LogLine: func(l app.LogLine) {
		mu.Lock()
		spoken = append(spoken, l.Msg)
		mu.Unlock()
	}})
	settings := app.NewSettings(bus, cfg, &app.OpLock{}, app.SettingsDeps{
		GitHubGet: func(path, token, apiBase string, timeoutS float64) connectioncheck.GitHubReply {
			payload := storage.NewOrderedMap()
			payload.Set("login", "alice")
			return connectioncheck.GitHubReply{Status: 200, Payload: payload}
		},
		WorkerCheck: func(config any) *connectioncheck.Result {
			return &connectioncheck.Result{Name: "Bot Worker", Level: connectioncheck.LevelOK,
				Detail: "v0.40.0 — matches this app"}
		},
		ProxyCheck: func(config any, timeoutS float64) connectioncheck.ProxyResult {
			return connectioncheck.ProxyResult{State: "connected", Host: "127.0.0.1", Port: 10808,
				Detail: "Proxy OK (127.0.0.1:10808)"}
		},
		DetectOllama: func(base string, timeoutS float64) llmclient.DetectOllamaResult {
			return llmclient.DetectOllamaResult{Found: true, BaseURL: base,
				Models: []string{"llama3"}, Detail: "ok"}
		},
		OllamaChatPing: func(host, model, prompt string, opts *llmclient.OllamaChatOptions, timeoutS float64) (string, string) {
			return "5", ""
		},
		CloudChatPing: func(apiURL, apiKey, model, prompt string, timeoutS float64) (string, string) {
			return "Hello there!", ""
		},
	})
	ui := NewHeroUI(settings, app.NewCommands(bus, app.Deps{
		Config: func() *storage.OrderedMap { return cfg },
	}))
	return ui, cfg, &spoken, &mu
}

func spokePerTest(t *testing.T, spoken *[]string, mu *sync.Mutex, part string) {
	t.Helper()
	mu.Lock()
	defer mu.Unlock()
	for _, l := range *spoken {
		if strings.Contains(l, part) {
			return
		}
	}
	t.Fatalf("no desk line contains %q (have %d lines)", part, len(*spoken))
}

func TestHeroUIPerTestDoors(t *testing.T) {
	ui, cfg, spoken, mu := newPerTestRig(t)
	cfg.Set("github_token", "tok")
	ollama := storage.NewOrderedMap()
	ollama.Set("base_url", "http://127.0.0.1:11434")
	ollama.Set("model", "llama3")
	cfg.Set("ollama", ollama)

	// the three log-only doors: the lines stream, no box.
	for _, door := range []struct {
		name string
		call func() *MoreDoorResult
		line string
	}{
		{"telegram-github", ui.MoreTestTelegramGitHub, "✅ GitHub token valid (user: alice)"},
		{"telegram-github-worker", ui.MoreTestTelegramGitHub, "☁️ Worker queue: v0.40.0 — matches this app"},
		{"proxy", ui.MoreTestProxy, "✅ Proxy OK (127.0.0.1:10808)"},
		{"ollama", ui.MoreTestOllama, "✅ Ollama is fully functional (connection + model + inference)."},
	} {
		res := door.call()
		if res == nil || res.Title != "" {
			t.Fatalf("%s: the door is log-only, got %+v", door.name, res)
		}
		spokePerTest(t, spoken, mu, door.line)
	}

	// the two box doors: the title/body passthrough.
	cfg.Set("cloud_api_url", "https://api.openai.com/v1")
	cfg.Set("cloud_api_key", "k")
	cfg.Set("cloud_model", "gpt-x")
	res := ui.MoreTestCloud()
	if res == nil || res.Title != "Cloud LLM Connected" {
		t.Fatalf("cloud door: %+v", res)
	}
	if !strings.Contains(res.Body, "\"Hello there!\"") {
		t.Fatalf("cloud body: %q", res.Body)
	}
	spokePerTest(t, spoken, mu, "✅ Cloud LLM is fully functional.")
}

func TestHeroUIPerTestLlamaCppDoor(t *testing.T) {
	// the llama.cpp door rides its own rig (the probe seams faked).
	appDir := t.TempDir()
	t.Setenv("GITCURATOR_HOME", appDir)
	bus := app.NewBus()
	cfg := storage.NewOrderedMap()
	var mu sync.Mutex
	var spoken []string
	bus.Subscribe(&app.Handlers{LogLine: func(l app.LogLine) {
		mu.Lock()
		spoken = append(spoken, l.Msg)
		mu.Unlock()
	}})
	model := "qwen3"
	settings := app.NewSettings(bus, cfg, &app.OpLock{}, app.SettingsDeps{
		ProbeLlama: func(baseURL, apiKey string, timeoutS float64) *llmclient.ProbeResult {
			return &llmclient.ProbeResult{Found: true, BaseURL: baseURL,
				Models: []string{model}, Model: &model, Detail: "llama.cpp found"}
		},
		CloudChatPing: func(apiURL, apiKey, model, prompt string, timeoutS float64) (string, string) {
			if apiURL != "http://127.0.0.1:8080/v1" {
				t.Errorf("chat apiURL = %q", apiURL)
			}
			return "hi", ""
		},
	})
	ui := NewHeroUI(settings, app.NewCommands(bus, app.Deps{
		Config: func() *storage.OrderedMap { return cfg },
	}))
	cfg.Set("llamacpp_api_url", "http://127.0.0.1:8080")
	res := ui.MoreTestLlamaCpp()
	if res == nil || res.Title != "llama.cpp Connected" {
		t.Fatalf("llama door: %+v", res)
	}
	spokePerTest(t, &spoken, &mu, "✅ llama.cpp responded: \"hi\"")
}

// newLinkDoorRig builds the chrome over the linking doors' faked
// seams (T-79b): the Ollama chat lane answering a confirm-yes, the
// embeddings router answering identical vectors, a log capture, and
// the HeroUI wrapper.
func newLinkDoorRig(t *testing.T) (*HeroUI, *storage.OrderedMap, *[]string, *sync.Mutex, *app.OpLock) {
	t.Helper()
	appDir := t.TempDir()
	t.Setenv("GITCURATOR_HOME", appDir)
	bus := app.NewBus()
	cfg := storage.NewOrderedMap()
	lock := &app.OpLock{}
	var spoken []string
	var mu sync.Mutex
	bus.Subscribe(&app.Handlers{LogLine: func(l app.LogLine) {
		mu.Lock()
		spoken = append(spoken, l.Msg)
		mu.Unlock()
	}})
	settings := app.NewSettings(bus, cfg, lock, app.SettingsDeps{
		LinkOllamaChat: func(host, model string, messages []llmclient.Message,
			numCtx int, timeoutS float64) (string, string) {
			return `{"best_used_for": "use when you need to file things", ` +
				`"related": true, "reason": "the same body"}`, ""
		},
		LinkEmbedCall: func(cfg any, modelOverride string) (mirror.EmbedFn, string, error) {
			return func(texts []string) ([][]float64, error) {
				out := make([][]float64, len(texts))
				for i := range texts {
					out[i] = []float64{1, 0, 0}
				}
				return out, nil
			}, "Ollama http://127.0.0.1:11434 \u00b7 nomic-embed-text", nil
		},
	})
	ui := NewHeroUI(settings, app.NewCommands(bus, app.Deps{
		Config: func() *storage.OrderedMap { return cfg },
	}))
	return ui, cfg, &spoken, &mu, lock
}

// writeLinkDoorNote drops one managed note into a vault.
func writeLinkDoorNote(t *testing.T, vault, name string) string {
	t.Helper()
	if err := os.MkdirAll(vault, 0o755); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(vault, name+".md")
	content := "---\nsource: https://github.com/acme/" + name + "\n" +
		"category: Agents\n---\n\n# " + name + "\n\nThe same body.\n"
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestHeroUILinkDoorsValidateAndVerify(t *testing.T) {
	ui, cfg, spoken, mu, _ := newLinkDoorRig(t)
	vault := t.TempDir()
	if err := os.MkdirAll(filepath.Join(vault, ".obsidian"), 0o755); err != nil {
		t.Fatal(err)
	}
	cfg.Set("vault_path", vault)

	// the validate door: log-only, no box.
	res := ui.MoreValidateVault()
	if res == nil || res.Title != "" {
		t.Fatalf("validate is log-only: %+v", res)
	}
	spokePerTest(t, spoken, mu, "\U0001f4c1 Validating vault path...")
	spokePerTest(t, spoken, mu,
		"\U00002705 Vault is valid (contains .obsidian folder): "+vault)

	// the verify door: the report box over the (empty) vault — no
	// manifest yet.
	cfg.Set("vault_path", t.TempDir())
	vres := ui.MoreVerifyVault()
	if vres == nil || vres.Title != "Verify Vault" {
		t.Fatalf("verify box: %+v", vres)
	}
	if !strings.Contains(vres.Body, "No previous manifest found") {
		t.Fatalf("verify body: %q", vres.Body)
	}
	spokePerTest(t, spoken, mu,
		"\u2139\ufe0f No manifest to verify \u2014 process a batch first.")
}

func TestHeroUILinkDoorsToolRunners(t *testing.T) {
	ui, cfg, spoken, mu, _ := newLinkDoorRig(t)
	vault := t.TempDir()
	writeLinkDoorNote(t, vault, "alpha")

	// the recall door with a missing vault: the runner's whole grammar.
	cfg.Set("vault_path", filepath.Join(t.TempDir(), "missing"))
	res := ui.MoreRecallHooksDryRun()
	if res == nil || res.Title != "" {
		t.Fatalf("the tool doors are log-only: %+v", res)
	}
	spokePerTest(t, spoken, mu,
		"\u23f3 Recall hooks (dry-run) running\u2026 (the GUI stays usable)")
	spokePerTest(t, spoken, mu,
		"GitHub vault not set or not found (config vault_path or --vault).")
	spokePerTest(t, spoken, mu,
		"\u274c Recall hooks (dry-run) finished with exit code 1 "+
			"(see the lines above \u2014 the command-line tools print the exact next step).")

	// the link-suggestions door with a full fixture: the one write.
	manual := t.TempDir()
	if err := os.MkdirAll(filepath.Join(manual, "Library"), 0o755); err != nil {
		t.Fatal(err)
	}
	websites := t.TempDir()
	writeLinkDoorNote(t, websites, "beta")
	cfg.Set("vault_path", vault)
	cfg.Set("website_vault_path", websites)
	cfg.Set("manual_vault_path", manual)
	lres := ui.MoreBuildLinkSuggestions()
	if lres == nil || lres.Title != "" {
		t.Fatalf("the tool doors are log-only: %+v", lres)
	}
	spokePerTest(t, spoken, mu,
		"  + related: alpha \u2194 beta \u2014 the same body")
	spokePerTest(t, spoken, mu,
		"\u2705 Link suggestions finished.")
	note, err := os.ReadFile(mirror.SuggestionsPath(manual))
	if err != nil {
		t.Fatalf("the Suggestions note: %v", err)
	}
	if !strings.Contains(string(note), "- [ ] ") {
		t.Fatalf("the pending checkbox: %s", note)
	}
}

func TestHeroUILinkDoorsRecategorizeFlow(t *testing.T) {
	ui, cfg, spoken, mu, _ := newLinkDoorRig(t)
	vault := t.TempDir()
	oldPath := writeLinkDoorNote(t, vault, "alpha")
	cfg.Set("vault_path", vault)

	// the open door: the table payload.
	model := ui.MoreRecategorizeOpen()
	if model == nil || model.BoxTitle != "" || len(model.Notes) != 1 {
		t.Fatalf("the table model: %+v", model)
	}
	if model.Notes[0].Category != "Agents" || len(model.Categories) == 0 {
		t.Fatalf("the row: %+v", model.Notes[0])
	}

	// the invalid-vault box.
	cfg.Set("vault_path", "")
	bbox := ui.MoreRecategorizeOpen()
	if bbox.BoxTitle != "Error" || bbox.BoxBody != "Please select a valid vault first." {
		t.Fatalf("the error box: %+v", bbox)
	}

	// the apply door: the move + the finish line.
	cfg.Set("vault_path", vault)
	changes := []app.RecategorizeChange{{
		Path: oldPath, OldCategory: "Agents", Category: "Tools",
	}}
	res := ui.MoreRecategorizeApply(changes)
	if res == nil || res.Moved != 1 {
		t.Fatalf("moved: %+v", res)
	}
	spokePerTest(t, spoken, mu, "\U0001f4c1 Recategorized 1 notes")
	if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
		t.Fatalf("the old file must be gone")
	}
}

func TestHeroUILinkDoorsBusyGuard(t *testing.T) {
	ui, cfg, spoken, mu, lock := newLinkDoorRig(t)
	cfg.Set("vault_path", t.TempDir())
	// the busy lock (the guard law — the shared OpLock).
	if !lock.Acquire("batch") {
		t.Fatal("acquire")
	}
	defer lock.Release("batch")
	ui.MoreValidateVault()
	ui.MoreRecallHooksDryRun()
	ui.MoreBuildLinkSuggestions()
	if res := ui.MoreVerifyVault(); res.Title != "" {
		t.Fatalf("the guarded verify answers no box: %+v", res)
	}
	if m := ui.MoreRecategorizeOpen(); m.BoxTitle != "" || len(m.Notes) != 0 {
		t.Fatalf("the guarded open answers the empty model: %+v", m)
	}
	spokePerTest(t, spoken, mu,
		"\u23f3 A batch is running \u2014 wait for it to finish before running the linking tools.")
	mu.Lock()
	count := 0
	for _, l := range *spoken {
		if strings.Contains(l, "A batch is running") {
			count++
		}
	}
	mu.Unlock()
	if count != 5 {
		t.Fatalf("all five doors speak the guard line once (have %d)", count)
	}
}

func TestHeroflowPerTestDoorsLive(t *testing.T) {
	src := heroflowSource(t)
	// the five doors are LIVE (T-79a) — each with its dispatch key.
	for _, want := range []string{
		"live: 'testTgGh'", "live: 'testProxy'", "live: 'testOllama'",
		"live: 'testCloud'", "live: 'testLlama'",
		"MoreTestTelegramGitHub", "MoreTestProxy", "MoreTestOllama",
		"MoreTestCloud", "MoreTestLlamaCpp",
	} {
		if !strings.Contains(src, want) {
			t.Fatalf("the per-test doors' live wiring missing %q", want)
		}
	}
	// the verdict boxes ride the SAME themed box the dashboard door
	// uses; a title only (the log-only doors return an empty title and
	// no box opens).
	if !strings.Contains(src, "if (res && res.title)") {
		t.Fatal("the verdict box law (title-only)")
	}
	// T-80: the export-zip row is LIVE too (the standing assignment
	// landed) — the More menu has NO remaining disabled row; the old
	// landing note is gone.
	if strings.Contains(src, "The export-zip engine rides the packaging phase") {
		t.Fatal("the export-zip row is live — its landing note is gone")
	}
	for _, want := range []string{
		"live: 'exportZip'", "MoreExportBackupZip", "exportZipRunning",
	} {
		if !strings.Contains(src, want) {
			t.Fatalf("the export-zip row's live wiring missing %q", want)
		}
	}
	if strings.Contains(src, "The graveyard procedure lands with the master-table tools task") {
		t.Fatal("the T-79e door is live — its landing note is gone")
	}
	// T-79b: the five linking doors are LIVE — each with its dispatch
	// key + binding name; the old landing notes are gone.
	for _, want := range []string{
		"live: 'recall'", "live: 'links'", "live: 'validate'",
		"live: 'verify'", "live: 'recategorize'",
		"MoreRecallHooksDryRun", "MoreBuildLinkSuggestions",
		"MoreValidateVault", "MoreVerifyVault",
		"MoreRecategorizeOpen", "MoreRecategorizeApply",
	} {
		if !strings.Contains(src, want) {
			t.Fatalf("the linking doors' live wiring missing %q", want)
		}
	}
	for _, gone := range []string{
		"The linking tools land with their own task",
		"The vault validation door lands with the linking-tools task",
		"The vault verification door lands with the linking-tools task",
		"The recategorize door lands with the linking-tools task",
	} {
		if strings.Contains(src, gone) {
			t.Fatalf("the landed doors' landing note still present: %q", gone)
		}
	}
	// the recategorize modal's laws: the pinned button row (the
	// long-body law) + the reference's own column heads.
	for _, want := range []string{
		"openRecategorizeDialog", "'New Category'", "'Apply Changes'",
		"linkdoor-box", "oldCategory",
	} {
		if !strings.Contains(src, want) {
			t.Fatalf("the recategorize dialog law missing %q", want)
		}
	}
}

// ---------------------------------------------------------------------------
// T-79c — the bot-queue tool doors (the More menu's third family).
// ---------------------------------------------------------------------------

// newBotDoorUIRig builds a heroRig whose Settings carries the faked
// BotQueueFetch seam (botQueueItems feeds it) + the faked batch + the
// cache under a temp APP_DIR.
func newBotDoorUIRig(t *testing.T) *heroRig {
	t.Helper()
	appDir := t.TempDir()
	t.Setenv("GITCURATOR_HOME", appDir)
	bus := app.NewBus()
	cfg := storage.NewOrderedMap()
	var spoken []string
	var mu sync.Mutex
	bus.Subscribe(&app.Handlers{LogLine: func(l app.LogLine) {
		mu.Lock()
		spoken = append(spoken, l.Msg)
		mu.Unlock()
	}})
	rig := &heroRig{cfg: cfg, spoken: &spoken, mu: &mu, queueState: "ok"}
	settings := app.NewSettings(bus, cfg, &app.OpLock{}, app.SettingsDeps{
		BotQueueFetch: func(c *storage.OrderedMap) (string, []any) {
			rig.mu.Lock()
			items := rig.botQueueItems
			rig.mu.Unlock()
			return "ok", items
		},
	})
	rig.commands = app.NewCommands(bus, app.Deps{
		Config: func() *storage.OrderedMap { return cfg },
		Batch: func(ctx context.Context, c *storage.OrderedMap,
			opts app.BatchOptions, ev app.BatchEvents) *app.BatchResult {
			rig.mu.Lock()
			rig.batchCalls++
			rig.batchModes = append(rig.batchModes, opts.Mode+"/"+opts.Source)
			rig.batchOpts = append(rig.batchOpts, app.BatchOptions{
				URLs:             append([]string(nil), opts.URLs...),
				NonGithub:        append([]string(nil), opts.NonGithub...),
				Source:           opts.Source,
				Mode:             opts.Mode,
				ImportFile:       opts.ImportFile,
				IntakeDuplicates: opts.IntakeDuplicates,
				RawURLCount:      opts.RawURLCount,
				DryRun:           opts.DryRun,
			})
			rig.mu.Unlock()
			return &app.BatchResult{Success: true, Message: "the fake batch finished"}
		},
	})
	rig.ui = NewHeroUI(settings, rig.commands)
	return rig
}

// botDoorCache opens the rig's cache (the temp APP_DIR's own cache.db).
func botDoorCache(t *testing.T) *store.CacheDB {
	t.Helper()
	cache, err := store.NewCacheDB(filepath.Join(constants.AppDir(), "cache.db"))
	if err != nil {
		t.Fatal(err)
	}
	return cache
}

func TestHeroUIBotQueueExportAndVerify(t *testing.T) {
	rig := newBotDoorUIRig(t)
	rig.cfg.Set("bot_username", "mybot")
	vault := t.TempDir()
	rig.cfg.Set("vault_path", vault)
	rig.mu.Lock()
	rig.botQueueItems = []any{
		map[string]any{"url_normalized": "https://github.com/acme/one",
			"forward_count": float64(2)},
		map[string]any{"url_normalized": "https://example.com/page"},
	}
	rig.mu.Unlock()

	// the export door: the box + the file + the two verdict lines.
	res := rig.ui.MoreExportAllBotLinks()
	if res == nil || res.Title != "📋 Links Exported" {
		t.Fatalf("the export box: %+v", res)
	}
	matches, err := filepath.Glob(filepath.Join(vault, "_bot_links_export_*.md"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("the export file: %v (%d)", err, len(matches))
	}
	spokePerTest(t, rig.spoken, rig.mu, "📋 Export saved: "+matches[0])
	spokePerTest(t, rig.spoken, rig.mu, "📊 Total messages: 2 | GitHub: 1 | Non-GitHub: 1 | Duplicates: 0")

	// the verify door: the fresh-cache settlement + the report box
	// over the fetched urls (both accounted for — one settled, the
	// ladder's own truth).
	vres := rig.ui.MoreVerifyAllBotLinks()
	if vres == nil || vres.Title != "✅ VERIFICATION REPORT" {
		t.Fatalf("the verify box: %+v", vres)
	}
	spokePerTest(t, rig.spoken, rig.mu, "📚 Vault index: 0 notes indexed")
	spokePerTest(t, rig.spoken, rig.mu, "✅ Verified: 0 in vault + 0 decommissioned + 1 settled = 1/1 accounted for")
	if len(vres.Missing) != 0 {
		t.Fatalf("the fresh-cache verify misses nothing: %v", vres.Missing)
	}

	// the username refusal (log-only, no box).
	rig.cfg.Set("bot_username", "")
	eres := rig.ui.MoreExportAllBotLinks()
	if eres == nil || eres.Title != "" {
		t.Fatalf("the refusal is log-only: %+v", eres)
	}
	spokePerTest(t, rig.spoken, rig.mu, "❌ Please enter the bot username first.")
}

func TestHeroUIBotQueueRetryFailedTwoPhase(t *testing.T) {
	rig := newBotDoorUIRig(t)
	rig.cfg.Set("vault_path", t.TempDir())

	// the empty queue: the box, no batch.
	out := rig.ui.MoreRetryFailed(false)
	if out.Box == nil || out.Box.Title != "Retry Queue Empty" {
		t.Fatalf("the empty box: %+v", out.Box)
	}
	spokePerTest(t, rig.spoken, rig.mu, "✓ No failed repos to retry.")

	// the stale Continue with no held list re-scans (never a blind
	// run) — the empty box again, still no batch.
	out = rig.ui.MoreRetryFailed(true)
	if out.Box == nil || out.Box.Title != "Retry Queue Empty" {
		t.Fatalf("the stale-continue re-scan: %+v", out.Box)
	}
	rig.mu.Lock()
	calls := rig.batchCalls
	rig.mu.Unlock()
	if calls != 0 {
		t.Fatalf("no batch before the kept list (have %d)", calls)
	}

	// a ≤10 kept list: no gate — the batch runs now (the DIRECT
	// source — bot_source=False), the pending lists cleared (the
	// reference's _bot_queue_urls law), the 🔄 line spoken.
	cache := botDoorCache(t)
	cache.AddFailed("https://github.com/acme/live-one", "boom")
	cache.AddFailed("https://github.com/acme/live-two", "boom")
	_ = cache.Close()
	rig.ui.mu.Lock()
	rig.ui.pendingRepos = []string{"https://github.com/acme/stale-bot-row"}
	rig.ui.pendingWebsites = []string{"https://example.com/stale-web-row"}
	rig.ui.mu.Unlock()
	out = rig.ui.MoreRetryFailed(false)
	if out.Box != nil || out.Confirm != nil || len(out.URLs) != 0 {
		t.Fatalf("a run's answer is the empty payload: %+v", out)
	}
	spokePerTest(t, rig.spoken, rig.mu, "🔄 Retrying 2 previously-failed repos...")
	rig.mu.Lock()
	calls = rig.batchCalls
	modes := append([]string(nil), rig.batchModes...)
	rig.ui.mu.Lock()
	pending := append([]string(nil), rig.ui.pendingRepos...)
	webs := append([]string(nil), rig.ui.pendingWebsites...)
	held := append([]string(nil), rig.ui.retryFailedURLs...)
	rig.ui.mu.Unlock()
	rig.mu.Unlock()
	if calls != 1 {
		t.Fatalf("the batch ran once (have %d)", calls)
	}
	if len(modes) != 1 || modes[0] != "/direct" {
		t.Fatalf("the retry batch's source: %v", modes)
	}
	if len(pending) != 0 || len(webs) != 0 {
		t.Fatalf("the pending lists cleared: %v %v", pending, webs)
	}
	if len(held) != 0 {
		t.Fatalf("the held list cleared after the run: %v", held)
	}

	// a >10 kept list: the confirm gate — the unconfirmed pass asks
	// (no batch), the confirmed pass runs over the HELD urls. A fresh
	// cache (the earlier case's rows still sit in this APP_DIR's own
	// cache.db).
	if err := os.Remove(filepath.Join(constants.AppDir(), "cache.db")); err != nil {
		t.Fatal(err)
	}
	cache = botDoorCache(t)
	urls := make([]string, 11)
	for i := range urls {
		urls[i] = fmt.Sprintf("https://github.com/acme/bulk-%02d", i)
		cache.AddFailed(urls[i], "boom")
	}
	_ = cache.Close()
	out = rig.ui.MoreRetryFailed(false)
	if out.Confirm == nil || out.Confirm.Title != "Confirm Large Batch" ||
		!strings.Contains(out.Confirm.Body, "This will process 11 items (the retry queue).") {
		t.Fatalf("the confirm box: %+v", out.Confirm)
	}
	rig.mu.Lock()
	calls = rig.batchCalls
	rig.mu.Unlock()
	if calls != 1 {
		t.Fatalf("the unconfirmed pass never runs the batch (have %d)", calls)
	}
	out = rig.ui.MoreRetryFailed(true)
	if out == nil || out.Box != nil || out.Confirm != nil {
		t.Fatalf("the confirmed pass's answer is the empty payload: %+v", out)
	}
	rig.mu.Lock()
	calls = rig.batchCalls
	rig.mu.Unlock()
	if calls != 2 {
		t.Fatalf("the confirmed pass ran the batch (have %d)", calls)
	}
	spokePerTest(t, rig.spoken, rig.mu, "🔄 Retrying 11 previously-failed repos...")
}

func TestHeroUIBotQueueResolveDoors(t *testing.T) {
	rig := newBotDoorUIRig(t)
	urlA := "https://github.com/acme/alpha"
	urlB := "https://github.com/acme/beta"

	// Mark as Processed: the wrapper + the all-resolved arithmetic.
	res := rig.ui.MoreResolveMarkProcessed(urlA, []string{urlA, urlB})
	if res == nil || res.AllResolved {
		t.Fatalf("mark processed: %+v", res)
	}
	res = rig.ui.MoreResolveMarkProcessed(urlB, []string{urlB})
	if res == nil || !res.AllResolved {
		t.Fatalf("the last row resolved: %+v", res)
	}
	spokePerTest(t, rig.spoken, rig.mu, "✅ All missing links resolved!")

	// Decommission + Mark ALL.
	res = rig.ui.MoreResolveDecommission(urlA, []string{urlA})
	if res == nil || !res.AllResolved {
		t.Fatalf("decommission: %+v", res)
	}
	spokePerTest(t, rig.spoken, rig.mu, "🗑️ Decommissioned (manual): "+urlA)
	res = rig.ui.MoreResolveMarkAll([]string{urlA, urlB})
	if res == nil || !res.AllResolved {
		t.Fatalf("mark all: %+v", res)
	}
	spokePerTest(t, rig.spoken, rig.mu, "✅ Marked 2 URL(s) as processed (manual)")

	// Process Selected: the 🚀 line + the BOT-source batch + the
	// pending seat's update (_process_missing_urls' assignment).
	bres := rig.ui.MoreResolveProcess([]string{urlA})
	if bres == nil || !bres.Success {
		t.Fatalf("process selected: %+v", bres)
	}
	spokePerTest(t, rig.spoken, rig.mu, "🚀 Processing 1 missing URL(s)...")
	rig.mu.Lock()
	modes := append([]string(nil), rig.batchModes...)
	rig.ui.mu.Lock()
	pending := append([]string(nil), rig.ui.pendingRepos...)
	rig.ui.mu.Unlock()
	rig.mu.Unlock()
	last := modes[len(modes)-1]
	if last != "/bot" {
		t.Fatalf("process selected is a bot-source batch: %v", modes)
	}
	if len(pending) != 1 || pending[0] != urlA {
		t.Fatalf("the pending seat carries the processed urls: %v", pending)
	}
}

// TestHeroflowChromeDoorsLive pins T-79d's frontend law: the four
// Chrome doors LIVE (each with its dispatch key), the two picker
// dialogs (the reference's own words + the checkable rows + the
// buttons pinned outside the scrolling body), the failure modal's
// one listener, AND THE BINDING-IMPORT LAW — every door joined the
// import statement (a door missing from the list throws a silent
// ReferenceError; the desk never speaks).
func TestHeroflowChromeDoorsLive(t *testing.T) {
	src := heroflowSource(t)
	for _, want := range []string{
		"live: 'handDeliver'", "live: 'handRows'",
		"live: 'chromeRetry'", "live: 'attach'",
		"MoreHandDeliverWalled", "MoreHandDeliverQueue",
		"MoreHandRowsDeliver(false)", "MoreHandRowsDeliver(true)",
		"MoreChromeTabRetry()", "MoreChromeTabRetryRun",
		"MoreAttachMyChrome(false)", "MoreAttachMyChrome(true)",
		"MoreOpenHandFolder",
	} {
		if !strings.Contains(src, want) {
			t.Fatalf("the Chrome doors' live wiring missing %q", want)
		}
	}
	// the two pickers: the reference's own words + shapes.
	for _, want := range []string{
		"openHandDeliverDialog", "openChromeRetryDialog",
		"Hand-deliver walled links", "Retry in your real Chrome?",
		"couldn't generate content",
		"Tick the ones worth saving",
		"'Webpage, HTML Only'",
		"'🖐 Queue & open in Chrome'", "'Not now'",
		"'🤖 Retry in Chrome now'", "'Open the folder'",
		"Deliver into: ' + model.folder",
		"pickRowEl", "type = 'checkbox'", "checked = true",
	} {
		if !strings.Contains(src, want) {
			t.Fatalf("the picker dialog's law missing %q", want)
		}
	}
	// the failure modal's ONE listener (the one-owner-per-surface law).
	if !strings.Contains(src, "EventsOn('chrome:failed'") {
		t.Fatal("the chrome:failed listener")
	}
	// THE BINDING-IMPORT LAW — every T-79d door inside the import
	// statement (the 044 lesson: a missing door is a silent
	// ReferenceError).
	if !strings.Contains(src, "MoreHandDeliverWalled, MoreHandDeliverQueue, MoreOpenHandFolder,") ||
		!strings.Contains(src, "MoreHandRowsDeliver, MoreChromeTabRetry, MoreChromeTabRetryRun,") ||
		!strings.Contains(src, "MoreAttachMyChrome,") {
		t.Fatal("the Chrome doors' import statement")
	}
}

func TestHeroflowBotQueueDoorsLive(t *testing.T) {
	src := heroflowSource(t)
	// the three doors are LIVE (T-79c) — each with its dispatch key.
	for _, want := range []string{
		"live: 'verifyAll'", "live: 'exportAll'", "live: 'retryFailed'",
		"MoreVerifyAllBotLinks", "MoreExportAllBotLinks", "MoreRetryFailed(false)",
		"MoreRetryFailed(true)",
		"MoreResolveMarkProcessed", "MoreResolveDecommission",
		"MoreResolveMarkAll", "MoreResolveProcess",
	} {
		if !strings.Contains(src, want) {
			t.Fatalf("the bot-queue doors' live wiring missing %q", want)
		}
	}
	// the manual-resolve dialog: the reference's own words + the
	// button hierarchy + the resolve-through-the-box law.
	for _, want := range []string{
		"openBotResolveDialog", "Manual Resolve — Missing Links",
		"'Process Selected'", "'Mark as Processed'", "'Decommission'",
		"'Mark ALL as Processed'", "'Resolve Missing Links'",
		"'Continue'",
	} {
		if !strings.Contains(src, want) {
			t.Fatalf("the resolve dialog's law missing %q", want)
		}
	}
	// T-80: the export-zip row is LIVE too — no disabled row remains
	// in the More menu; the old landing note is gone.
	if strings.Contains(src, "The export-zip engine rides the packaging phase") {
		t.Fatal("the export-zip row is live — its landing note is gone")
	}
	if strings.Contains(src, "The bot-queue tools land with the queue-tools task") {
		t.Fatal("the T-79c doors are live — their landing note is gone")
	}
	if strings.Contains(src, "The Chrome doors land with the Chrome-doors GUI task") {
		t.Fatal("the T-79d doors are live — their landing note is gone")
	}
	if strings.Contains(src, "The graveyard procedure lands with the master-table tools task") {
		t.Fatal("the T-79e door is live — its landing note is gone")
	}
	// the probe's click leg knows the export pick.
	if !strings.Contains(src, "door === 'export' ? 'Export All Links'") {
		t.Fatal("the probe's export pick")
	}
	// the bindings ride the IMPORT list (the gap the first click-leg
	// run caught: a door missing from the import throws a silent
	// ReferenceError — the desk never speaks; the pin keeps every
	// T-79c door inside the import statement).
	if !strings.Contains(src, "MoreExportAllBotLinks, MoreVerifyAllBotLinks, MoreRetryFailed,") ||
		!strings.Contains(src, "MoreResolveMarkProcessed, MoreResolveDecommission,") {
		t.Fatal("the bot-queue doors' import statement")
	}
}

// ---------------------------------------------------------------------------
// THE CHROME DOORS (T-79d) — the four More-menu doors over the
// ported engines (processing_control.py:615/669/1000/1601 + the
// shared _start_chrome_tab_retry at 1351).
// ---------------------------------------------------------------------------

// newChromeDoorUIRig wires the hero rig with the Chrome doors' faked
// seams (the CDP delivery + the hand enqueue — THE HERMETIC-SEAM law)
// + the websites pipeline on + a temp Websites vault.
func newChromeDoorUIRig(t *testing.T) *heroRig {
	t.Helper()
	rig := newHeroRig(t)
	rig.cfg.Set("pipelines", func() *storage.OrderedMap {
		p := storage.NewOrderedMap()
		p.Set("github", true)
		p.Set("websites", true)
		return p
	}())
	rig.cfg.Set("website_vault_path", t.TempDir())
	// the CDP delivery seam (THE HERMETIC-SEAM law — the real
	// DeliverPagesViaChrome never runs in the shell battery).
	rig.ui.settings.Deps.ChromeDeliver = func(vault string, links []app.ChromeLinkRow,
		log func(msg, level string), config *storage.OrderedMap) app.ChromeDeliveryReport {
		if rig.deliverGate != nil {
			<-rig.deliverGate
		}
		rig.mu.Lock()
		rig.chromeDeliverCalls++
		calls := rig.chromeDeliverCalls
		rig.mu.Unlock()
		if log != nil {
			log(fmt.Sprintf("🤖 the fake fifth door delivered page(s) (pass %d)", calls), "success")
		}
		return app.ChromeDeliveryReport{Delivered: len(links),
			URLs: chromeRowURLs(links)}
	}
	return rig
}

// waitBatchBusy polls the Commands' batch lock until a batch holds it.
func waitBatchBusy(t *testing.T, rig *heroRig) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if rig.ui.commands.BatchBusy() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("the batch never took the lock")
}

// waitBatchCalls polls the fake batch's call count.
func waitBatchCalls(t *testing.T, rig *heroRig, n int) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		rig.mu.Lock()
		calls := rig.batchCalls
		rig.mu.Unlock()
		if calls >= n {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("the batch ran %d times, want %d", rig.batchCalls, n)
}

// waitDeliverCalls polls the fake fifth door's call count.
func waitDeliverCalls(t *testing.T, rig *heroRig, n int) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		rig.mu.Lock()
		calls := rig.chromeDeliverCalls
		rig.mu.Unlock()
		if calls >= n {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("the fifth door ran %d times, want %d", rig.chromeDeliverCalls, n)
}

// chromeRowURLs extracts the rows' urls.
func chromeRowURLs(links []app.ChromeLinkRow) []string {
	out := make([]string, 0, len(links))
	for _, l := range links {
		out = append(out, l.URL)
	}
	return out
}

// waitSpoken polls the desk lines until part arrives (the delivery's
// background goroutine speaks asynchronously — the honest async-wait,
// never a sleep-only dance).
func waitSpoken(t *testing.T, rig *heroRig, part string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		rig.mu.Lock()
		found := false
		for _, l := range *rig.spoken {
			if strings.Contains(l, part) {
				found = true
				break
			}
		}
		rig.mu.Unlock()
		if found {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("no desk line contains %q (have: %s)", part, rig.logs())
}

func TestHeroUIChromeHandDeliverWalled(t *testing.T) {
	rig := newChromeDoorUIRig(t)

	// the empty picker: the box in the reference's own words.
	out := rig.ui.MoreHandDeliverWalled()
	if out.Box == nil || out.Box.Title != "No walled links waiting" ||
		!strings.Contains(out.Box.Body, "When one appears, its _review note and master-table row will say so (the fourth door hint).") {
		t.Fatalf("the empty box: %+v", out.Box)
	}
	if !strings.Contains(out.Folder, filepath.Join("_review", "hand-delivered")) {
		t.Fatalf("the folder: %q", out.Folder)
	}

	// the walled rows: seed the rig APP_DIR's own cache.db.
	state, err := websitepipeline.NewWebsiteStateDB(
		filepath.Join(constants.AppDir(), "cache.db"))
	if err != nil {
		t.Fatal(err)
	}
	if err := state.EnqueueRetry("https://example.com/walled", "HTTP 403 — bot defense"); err != nil {
		t.Fatal(err)
	}
	state.Close()
	out = rig.ui.MoreHandDeliverWalled()
	if out.Box != nil || len(out.Links) != 1 ||
		out.Links[0].URL != "https://example.com/walled" {
		t.Fatalf("the walled rows: %+v", out)
	}

	// the queue door: the picked rows enqueue through the seam + the
	// box answers; an empty list speaks the reference's cancel line.
	rig.ui.settings.Deps.HandEnqueue = func(vault string, urls []string,
		walls map[string]string, log func(msg, level string), openChrome bool) (int, int) {
		if len(urls) != 1 || !openChrome {
			t.Fatalf("the enqueue call: %v open=%v", urls, openChrome)
		}
		return 1, 1
	}
	res := rig.ui.MoreHandDeliverQueue(out.Links)
	if res.Title != "Queued for hand-delivery" ||
		!strings.Contains(res.Body, "1 link(s) queued, 1 opened in your Chrome.") {
		t.Fatalf("the queued box: %+v", res)
	}
	res = rig.ui.MoreHandDeliverQueue(nil)
	if res.Title != "" {
		t.Fatalf("the cancel is log-only: %+v", res)
	}
	spokePerTest(t, rig.spoken, rig.mu, "⏭️ Hand-delivery cancelled — nothing was queued.")
}

func TestHeroUIChromeHandRowsTwoPhase(t *testing.T) {
	rig := newChromeDoorUIRig(t)

	// the empty table: the box, no delivery.
	out := rig.ui.MoreHandRowsDeliver(false)
	if out.Box == nil || out.Box.Title != "No hand rows waiting" {
		t.Fatalf("the empty box: %+v", out.Box)
	}
	rig.mu.Lock()
	calls := rig.chromeDeliverCalls
	rig.mu.Unlock()
	if calls != 0 {
		t.Fatalf("no delivery before the rows (have %d)", calls)
	}

	// the stale Continue with no held rows re-scans (never a blind
	// run) — the empty box again.
	out = rig.ui.MoreHandRowsDeliver(true)
	if out.Box == nil || out.Box.Title != "No hand rows waiting" {
		t.Fatalf("the stale-continue re-scan: %+v", out.Box)
	}

	// two hand rows: no gate — the 🖐 line + the delivery now (the
	// fake door's own line follows; the batch re-runs the delivered
	// pages as the direct WEBSITES batch).
	writeHandRowsTable(t, rigConfigVault(t, rig), [][3]string{
		{"https://example.com/hand-one", "🖐", ""},
		{"https://example.com/hand-two", "🖐", "HTTP 403"},
	})
	out = rig.ui.MoreHandRowsDeliver(false)
	if out.Box != nil || out.Confirm != nil {
		t.Fatalf("a run's answer is the empty payload: %+v", out)
	}
	spokePerTest(t, rig.spoken, rig.mu, "🖐 2 hand row(s) — opening your real Chrome (one tab per link; the live pages become real fetches)")
	waitSpoken(t, rig, "🤖 Chrome tab retry: opening 2 link(s) in your real Chrome")
	waitSpoken(t, rig, "🤖 2 page(s) delivered from your Chrome — re-processing them now through the Websites pipeline as real fetches")
	waitDeliverCalls(t, rig, 1)
	waitBatchCalls(t, rig, 1)
	rig.mu.Lock()
	calls = rig.chromeDeliverCalls
	batchCalls := rig.batchCalls
	rig.mu.Unlock()
	if calls != 1 {
		t.Fatalf("one delivery (have %d)", calls)
	}
	if batchCalls != 1 {
		t.Fatalf("the delivered re-run batch (have %d)", batchCalls)
	}
	rig.mu.Lock()
	modes := append([]string(nil), rig.batchModes...)
	rig.mu.Unlock()
	if len(modes) != 1 || modes[0] != "/direct" {
		t.Fatalf("the re-run's source: %v", modes)
	}

	// the busy box: a delivery running answers before the scan.
	rig.ui.mu.Lock()
	rig.ui.chromeDeliveryBusy = true
	rig.ui.mu.Unlock()
	out = rig.ui.MoreHandRowsDeliver(false)
	if out.Box == nil || out.Box.Title != "A Chrome delivery is running" {
		t.Fatalf("the busy box: %+v", out.Box)
	}
	rig.ui.mu.Lock()
	rig.ui.chromeDeliveryBusy = false
	rig.ui.mu.Unlock()

	// the >10 gate: 11 rows ask; Continue runs over the held rows.
	rows := make([][3]string, 11)
	for i := range rows {
		rows[i] = [3]string{fmt.Sprintf("https://example.com/bulk-%02d", i), "🖐", ""}
	}
	writeHandRowsTable(t, rigConfigVault(t, rig), rows)
	out = rig.ui.MoreHandRowsDeliver(false)
	if out.Confirm == nil || !strings.Contains(out.Confirm.Body, "This will process 11 items (the table's 🖐 hand rows).") {
		t.Fatalf("the >10 gate: %+v", out.Confirm)
	}
	rig.mu.Lock()
	calls = rig.chromeDeliverCalls
	rig.mu.Unlock()
	if calls != 1 {
		t.Fatalf("no delivery before the confirm (have %d)", calls)
	}
	out = rig.ui.MoreHandRowsDeliver(true)
	if out.Box != nil || out.Confirm != nil {
		t.Fatalf("the confirmed pass runs: %+v", out)
	}
	waitSpoken(t, rig, "🖐 11 hand row(s) — opening your real Chrome")
	waitDeliverCalls(t, rig, 2)
	rig.mu.Lock()
	calls = rig.chromeDeliverCalls
	held := append([]app.ChromeLinkRow(nil), rig.ui.handRowsRows...)
	rig.mu.Unlock()
	if calls != 2 {
		t.Fatalf("the held rows delivered (have %d)", calls)
	}
	if len(held) != 0 {
		t.Fatalf("the held list cleared after the run: %v", held)
	}
}

// rigConfigVault answers the rig's Websites vault path.
func rigConfigVault(t *testing.T, rig *heroRig) string {
	t.Helper()
	v, _ := rig.cfg.Get("website_vault_path")
	s, _ := v.(string)
	if s == "" {
		t.Fatal("no Websites vault in the rig config")
	}
	return s
}

// writeHandRowsTable writes the master table (the notes column is the
// wall the hand rows carry).
func writeHandRowsTable(t *testing.T, vault string, rows [][3]string) {
	t.Helper()
	path := websitepipeline.DecommissionTablePath(vault)
	var lines []string
	for _, row := range rows {
		domain := "unknown"
		if i := strings.Index(row[0], "://"); i >= 0 {
			rest := row[0][i+3:]
			if j := strings.IndexAny(rest, "/?#"); j >= 0 {
				rest = rest[:j]
			}
			if rest != "" {
				domain = rest
			}
		}
		lines = append(lines, fmt.Sprintf("| - | 2026-10-09 | %s | %s | test | %s | %s |",
			row[0], domain, row[1], row[2]))
	}
	content := fmt.Sprintf(websitepipeline.GraveyardHeader, "2026-10-09 12:00") +
		strings.Join(lines, "\n") + "\n"
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestHeroUIChromeTabRetryFlow(t *testing.T) {
	rig := newChromeDoorUIRig(t)

	// the empty queue: the box, nothing offered.
	out := rig.ui.MoreChromeTabRetry()
	if out.Box == nil || out.Box.Title != "No failed links waiting" {
		t.Fatalf("the empty box: %+v", out.Box)
	}

	// the failed links: the offer's rows.
	state, err := websitepipeline.NewWebsiteStateDB(
		filepath.Join(constants.AppDir(), "cache.db"))
	if err != nil {
		t.Fatal(err)
	}
	if err := state.EnqueueRetry("https://example.com/failed", "connection reset"); err != nil {
		t.Fatal(err)
	}
	state.Close()
	out = rig.ui.MoreChromeTabRetry()
	if out.Box != nil || len(out.Links) != 1 ||
		out.Links[0].Error != "connection reset" {
		t.Fatalf("the offer rows: %+v", out)
	}

	// "Not now": the cancel line, no delivery, nothing marked.
	rig.ui.MoreChromeTabRetryRun(nil)
	spokePerTest(t, rig.spoken, rig.mu, "⏭️ Chrome tab retry cancelled — the failed links keep waiting (their automatic retries continue).")
	rig.mu.Lock()
	calls := rig.chromeDeliverCalls
	rig.mu.Unlock()
	if calls != 0 {
		t.Fatalf("no delivery on Not-now (have %d)", calls)
	}
	out = rig.ui.MoreChromeTabRetry()
	if len(out.Links) != 1 {
		t.Fatalf("the declined offer stays offerable: %+v", out)
	}

	// the retry-now leg: the picked rows are marked offered + the
	// delivery runs; a re-scan hides the offered link.
	rig.ui.MoreChromeTabRetryRun(out.Links)
	waitSpoken(t, rig, "🤖 Chrome tab retry: opening 1 link(s) in your real Chrome")
	out = rig.ui.MoreChromeTabRetry()
	if out.Box == nil || out.Box.Title != "No failed links waiting" {
		t.Fatalf("the offered filter: %+v", out.Box)
	}

	// the batch-busy refusal: the reference's _batch_running box.
	rig.mu.Lock()
	rig.batchBlock = make(chan struct{})
	rig.mu.Unlock()
	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = rig.ui.HeroRunQueue()
	}()
	waitBatchBusy(t, rig)
	out = rig.ui.MoreChromeTabRetry()
	if out.Box == nil || out.Box.Title != "Batch already running" ||
		out.Box.Body != "A batch is already running — finish or stop it first." {
		t.Fatalf("the busy box: %+v", out.Box)
	}
	close(rig.batchBlock)
	<-done
}

func TestHeroUIChromeDeliveryBacklogMerge(t *testing.T) {
	rig := newChromeDoorUIRig(t)
	// a delivery that waits for our signal (the busy window) — the
	// rig's own seam already rides deliverGate.
	rig.deliverGate = make(chan struct{})

	// the first delivery starts (busy).
	rig.ui.MoreChromeTabRetryRun([]app.ChromeLinkRow{{URL: "https://example.com/one"}})
	waitSpoken(t, rig, "🤖 Chrome tab retry: opening 1 link(s) in your real Chrome")

	// the second request merges into the backlog (the reference's
	// v0.52 law — ONE delivery at a time).
	rig.ui.MoreChromeTabRetryRun([]app.ChromeLinkRow{
		{URL: "https://example.com/two"}, {URL: "https://example.com/three"}})
	spokePerTest(t, rig.spoken, rig.mu, "🖐 A Chrome delivery is already running — 2 link(s) queued behind it (they open when the current tabs finish)")

	// the busy hand-rows door answers the busy box while the
	// delivery runs.
	out := rig.ui.MoreHandRowsDeliver(false)
	if out.Box == nil || out.Box.Title != "A Chrome delivery is running" {
		t.Fatalf("the busy box: %+v", out.Box)
	}

	// release: the backlog drains with its own pass (two links — its
	// own opening + delivered lines).
	close(rig.deliverGate)
	waitDeliverCalls(t, rig, 2)
	waitSpoken(t, rig, "🤖 Chrome tab retry: opening 2 link(s) in your real Chrome")
	waitSpoken(t, rig, "🤖 2 page(s) delivered from your Chrome")
}

func TestHeroUIChromeAttachDoor(t *testing.T) {
	rig := newChromeDoorUIRig(t)
	// no Chrome on the machine: the reference's own remedy box (THE
	// COLLABORATOR LAW — Chrome is optional, the door refuses
	// honestly).
	out := rig.ui.MoreAttachMyChrome(false)
	if out.Box == nil || out.Box.Title != "No Chrome found" ||
		out.Box.Body != "GitCurator could not find Google Chrome on this machine — the fifth door needs the real browser." {
		t.Fatalf("the no-chrome box: %+v", out.Box)
	}
}

// ---------------------------------------------------------------------------
// THE DECOMMISSION PICKER (T-79e) — the graveyard procedure's doors.
// ---------------------------------------------------------------------------

// newDecomUIRig is the decommission picker's shell harness (the
// chrome rig shape: the Websites pipeline on + a real temp vault; the
// door rides local files + the rig APP_DIR's own cache.db — no
// network anywhere near).
func newDecomUIRig(t *testing.T) *heroRig {
	t.Helper()
	rig := newHeroRig(t)
	rig.cfg.Set("pipelines", func() *storage.OrderedMap {
		p := storage.NewOrderedMap()
		p.Set("github", true)
		p.Set("websites", true)
		return p
	}())
	rig.cfg.Set("website_vault_path", t.TempDir())
	return rig
}

// writeDecomPlaceholder seeds one fetch-failed placeholder into the
// rig vault's _review (the scan's own frontmatter contract).
func writeDecomPlaceholder(t *testing.T, rig *heroRig, name, src string) string {
	t.Helper()
	vault := pyCfgStr(rig.cfg, "website_vault_path")
	dir := filepath.Join(vault, "_review")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, name)
	note := fmt.Sprintf("---\nsource: %q\nfetch_status: \"failed\"\nmanaged_by: \"gitcurator\"\n---\n\nbody\n", src)
	if err := os.WriteFile(path, []byte(note), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// TestHeroUIDecommissionScanDoor pins the scan door over the real
// engine: the guards' boxes (the unconfigured fixture speaks the
// reference's own remedy line) + the picker's data over the rig's own
// cache.db + the candidates pre-fill.
func TestHeroUIDecommissionScanDoor(t *testing.T) {
	rig := newDecomUIRig(t)

	// the guard leg: the pipeline off answers the reference's own box.
	rig.cfg.Set("pipelines", func() *storage.OrderedMap {
		p := storage.NewOrderedMap()
		p.Set("github", true)
		p.Set("websites", false)
		return p
	}())
	out := rig.ui.MoreDecommissionScan()
	if out.Box == nil || out.Box.Title != "Websites pipeline is off" ||
		out.Box.Body != "Turn the Websites pipeline on first (Settings → 📁 Vault)." {
		t.Fatalf("the guard box: %+v", out.Box)
	}

	// the picker's data: one waiting link + the pre-filled table.
	rig.cfg.Set("pipelines", func() *storage.OrderedMap {
		p := storage.NewOrderedMap()
		p.Set("github", true)
		p.Set("websites", true)
		return p
	}())
	writeDecomPlaceholder(t, rig, "a-placeholder.md", "https://one.example/lost")
	out = rig.ui.MoreDecommissionScan()
	if out.Box != nil || len(out.Links) != 1 || out.Links[0] != "https://one.example/lost" {
		t.Fatalf("the picker's data: %+v", out)
	}
	if !strings.HasSuffix(out.Table, filepath.Join("_review", "DECOMMISSIONED.md")) {
		t.Fatalf("the table path: %q", out.Table)
	}
	// the candidates pre-fill landed through the door (the one-ledger
	// law's first write).
	rows := websitepipeline.ScanDecommissionTable(pyCfgStr(rig.cfg, "website_vault_path"))
	if st, ok := rows["https://one.example/lost"]; !ok || !strings.Contains(strings.ToLower(st), "unreviewed") {
		t.Fatalf("the candidates row: %q", st)
	}
	spokePerTest(t, rig.spoken, rig.mu, "📋 1 row(s) written to")
}

// TestHeroUIDecommissionBuryDoor pins the answer door over the real
// writers + the real consume: the marking + the immediate enforcement
// + the exact box; the empty list is the dialog's Cancel (the
// reference's cancel line on the desk, no box).
func TestHeroUIDecommissionBuryDoor(t *testing.T) {
	rig := newDecomUIRig(t)
	p1 := writeDecomPlaceholder(t, rig, "a-placeholder.md", "https://one.example/lost")
	if out := rig.ui.MoreDecommissionScan(); out.Box != nil {
		t.Fatalf("the scan: %+v", out.Box)
	}

	res := rig.ui.MoreDecommissionBury([]string{"https://one.example/lost"})
	if res.Title != "Decommissioned" {
		t.Fatalf("the box title: %+v", res)
	}
	if !strings.Contains(res.Body, "1 link(s) marked dead — 1 enforced now, 1 placeholder(s) swept.") ||
		!strings.Contains(res.Body, "(♻️ revived on a row brings a link back to life.)") {
		t.Fatalf("the box body: %q", res.Body)
	}
	spokePerTest(t, rig.spoken, rig.mu, "🪦 Graveyard: 1 link(s) decommissioned, 1 placeholder(s) swept — the _review backlog is clear")
	if _, err := os.Stat(p1); !os.IsNotExist(err) {
		t.Fatal("the placeholder was swept through the door")
	}
	// the dismissal landed in the rig APP_DIR's own cache.db.
	state, err := websitepipeline.NewWebsiteStateDB(
		filepath.Join(constants.AppDir(), "cache.db"))
	if err != nil {
		t.Fatal(err)
	}
	if hit, _ := state.IsDismissed("https://one.example/lost"); !hit {
		t.Fatal("the burial's dismissal lives in the state DB")
	}
	state.Close()

	// the cancel leg: an empty list speaks the reference's line.
	res = rig.ui.MoreDecommissionBury(nil)
	if res.Title != "" || res.Body != "" {
		t.Fatalf("the cancel answers no box: %+v", res)
	}
	spokePerTest(t, rig.spoken, rig.mu, "⏭️ Decommission cancelled — nothing was buried.")
}

// TestHeroUIDecommissionOpenTable pins the reveal door over the
// OpenFileLocation seam.
func TestHeroUIDecommissionOpenTable(t *testing.T) {
	rig := newDecomUIRig(t)
	opened := ""
	rig.ui.settings.Deps.OpenFileLocation = func(p string) { opened = p }
	if !rig.ui.MoreOpenDecommissionTable() {
		t.Fatal("the reveal answers true")
	}
	if !strings.HasSuffix(opened, filepath.Join("_review", "DECOMMISSIONED.md")) {
		t.Fatalf("the reveal target: %q", opened)
	}
}

// TestHeroflowDecommissionDoorsLive pins T-79e's frontend law: the
// menu row LIVE, the dispatch, the picker dialog's own words + the
// UNCHECKED-default law, and THE BINDING-IMPORT LAW (the 044 lesson —
// a door missing from the import statement is a silent
// ReferenceError).
func TestHeroflowDecommissionDoorsLive(t *testing.T) {
	src := heroflowSource(t)
	// the row is LIVE with its dispatch key.
	for _, want := range []string{
		"live: 'decommission'",
		"MoreDecommissionScan()", "MoreDecommissionBury(picked)",
		"MoreDecommissionBury([])", "MoreOpenDecommissionTable()",
		"openDecommissionDialog",
	} {
		if !strings.Contains(src, want) {
			t.Fatalf("the decommission door's live wiring missing %q", want)
		}
	}
	// the picker: the reference's own words + shapes.
	for _, want := range []string{
		"Decommission dead links",
		"Tick the links that are GENUINELY dead — a real 404, a lost page, an abandoned domain.",
		"A buried link is never fetched, never retried, never noticed again; its _review placeholder is swept.",
		"(Unticked links keep waiting — ♻️ revived in the table brings a buried link back.)",
		"The ledger: ' + model.table",
		"'Open the table'", "'Cancel'", "'🪦 Decommission selected'",
		"type = 'checkbox'", "checked = false",
		"Reveal the graveyard table in your file manager — hand-edit the Status column in Obsidian (🪦 dead / ♻️ revived).",
	} {
		if !strings.Contains(src, want) {
			t.Fatalf("the picker dialog's law missing %q", want)
		}
	}
	// THE BINDING-IMPORT LAW — the three doors inside the import
	// statement (the 044 lesson: a missing door is a silent
	// ReferenceError and the desk never speaks).
	if !strings.Contains(src, "MoreDecommissionScan, MoreDecommissionBury, MoreOpenDecommissionTable,") {
		t.Fatal("the decommission doors' import statement")
	}
	// the probe's 'grave' click pick (the evidence leg).
	if !strings.Contains(src, "door === 'grave' ? 'Decommission dead links'") {
		t.Fatal("the probe's grave pick")
	}
}

// ---------------------------------------------------------------------------
// THE FANFARE CHAIN (T-79f) — the runner's automatic triggers.
// ---------------------------------------------------------------------------

// newFanfareUIRig is the chrome-door rig + the caught-up queue facts:
// a configured Worker queue whose classify answers EMPTY pendings (the
// caught-up path's own shape).
func newFanfareUIRig(t *testing.T) *heroRig {
	t.Helper()
	rig := newChromeDoorUIRig(t)
	rig.cfg.Set("cloudflare_enabled", true)
	rig.cfg.Set("cloudflare_worker_url", "https://worker.example")
	rig.cfg.Set("cloudflare_install_id", "install")
	rig.cfg.Set("cloudflare_shared_secret", "secret")
	bus := rig.commands.Bus
	rig.commands = app.NewCommands(bus, app.Deps{
		Config: func() *storage.OrderedMap { return rig.cfg },
		QueueFetch: func(cfg *storage.OrderedMap) (string, []any) {
			return "ok", []any{}
		},
		QueueClassify: func(cfg *storage.OrderedMap, items []any,
			log func(msg, level string)) *app.QueueReport {
			return &app.QueueReport{State: "ok", Success: true}
		},
		Batch: rig.commands.Deps.Batch,
	})
	rig.ui = NewHeroUI(rig.ui.settings, rig.commands)
	return rig
}

// seedFalseSuccess writes the redo pass's subject: a 🖐 row whose page
// was delivered but whose note never landed (the queue row + the page
// file at the suggested name — CollectDelivered reads the QUEUE).
func seedFalseSuccess(t *testing.T, vault, url, wall string) {
	t.Helper()
	handdelivery.EnqueueHandDelivery(vault, []string{url},
		map[string]string{url: wall}, nil, false, "")
	page := filepath.Join(handdelivery.HandDeliveryDir(vault),
		handdelivery.SuggestedFilename(url))
	if err := os.MkdirAll(filepath.Dir(page), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(page, []byte(
		"<html><head><title>Half</title></head><body>delivered</body></html>"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestHeroUIFanfareChainGate(t *testing.T) {
	rig := newFanfareUIRig(t)
	writeHandRowsTable(t, rigConfigVault(t, rig), [][3]string{
		{"https://example.com/hand", "\U0001f590", "HTTP 403"},
	})
	// the gate: NOT natural (the no-work early returns / the crash
	// path) + the user's stop both keep the chain silent.
	rig.ui.onBatchFanfare(app.BatchFinished{OK: true})
	rig.ui.onBatchFanfare(app.BatchFinished{OK: true, Natural: true, Stopped: true})
	rig.mu.Lock()
	calls := rig.chromeDeliverCalls
	batches := rig.batchCalls
	rig.mu.Unlock()
	if calls != 0 || batches != 0 {
		t.Fatalf("the gate holds: %d deliveries, %d batches", calls, batches)
	}
	if strings.Contains(rig.logs(), "hand row(s)") {
		t.Fatalf("the gate holds — no line:\n%s", rig.logs())
	}

	// the natural, not-stopped crossing fires the hand pass.
	rig.ui.onBatchFanfare(app.BatchFinished{OK: true, Natural: true})
	spokePerTest(t, rig.spoken, rig.mu,
		"\U0001f590 1 hand row(s) — your real Chrome opens for them now (one tab per link; GitCurator takes each live page and processes it as a real fetch — no manual saving)")
	waitSpoken(t, rig, "\U0001f916 Chrome tab retry: opening 1 link(s) in your real Chrome")

	// the tried belt: the SAME table answers nothing on the next
	// natural crossing (once per session per link).
	rig.ui.onBatchFanfare(app.BatchFinished{OK: true, Natural: true})
	// the first crossing's delivery is ASYNCHRONOUS (startChromeTabRetry's
	// goroutine) — wait for it to land before the exact count (the
	// belt's own law caps the count at 1 here, so the wait is race-free;
	// the bare read flaked under the multi-package battery's CPU
	// contention — SESSION-059's flake repair, the rig's own waiter).
	waitDeliverCalls(t, rig, 1)
	rig.mu.Lock()
	calls = rig.chromeDeliverCalls
	rig.mu.Unlock()
	if calls != 1 {
		t.Fatalf("the tried belt holds one delivery per link per session (have %d)", calls)
	}
	if got := strings.Count(rig.logs(), "hand row(s) — your real Chrome opens"); got != 1 {
		t.Fatalf("the tried belt is quiet (the line spoke %d times):\n%s", got, rig.logs())
	}
}

func TestHeroUIFanfareChainRedoLeg(t *testing.T) {
	rig := newFanfareUIRig(t)
	seedFalseSuccess(t, rigConfigVault(t, rig), "https://example.com/false-success", "HTTP 403")
	writeHandRowsTable(t, rigConfigVault(t, rig), [][3]string{
		{"https://example.com/false-success", "\U0001f590", "HTTP 403"},
	})
	// NO hand pass fires (the row's page was delivered — the hand scan
	// consumes delivered rows), so the redo leg runs: the \U0001f501 line
	// + the asynchronous re-run batch over the Websites pipeline.
	rig.ui.onBatchFanfare(app.BatchFinished{OK: true, Natural: true})
	spokePerTest(t, rig.spoken, rig.mu, "\U0001f501 1 hand note(s) NOT properly stored — refetching and re-generating them now")
	waitBatchCalls(t, rig, 1)
	rig.mu.Lock()
	modes := append([]string(nil), rig.batchModes...)
	rig.mu.Unlock()
	if len(modes) != 1 || modes[0] != "/direct" {
		t.Fatalf("the redo re-run rides the direct non-GitHub batch: %v", modes)
	}
	// the re-run's own crossing is natural too — its chain finds
	// nothing left (the redo tried belt), no second batch.
	rig.mu.Lock()
	batches := rig.batchCalls
	rig.mu.Unlock()
	if batches > 2 {
		t.Fatalf("the recursion terminates (have %d batches)", batches)
	}
}

func TestHeroUIFanfareCaughtUpHandScrape(t *testing.T) {
	rig := newFanfareUIRig(t)
	writeHandRowsTable(t, rigConfigVault(t, rig), [][3]string{
		{"https://example.com/hand", "\U0001f590", "HTTP 403"},
	})
	res := rig.ui.HeroClick("")
	// hero.py:329-330 — work started: NO caught-up claim; the button
	// holds the disabled FETCHING render; the bar holds the fetch line.
	if res.State != app.HeroFetching || res.Format != app.FormatFetching {
		t.Fatalf("the work-started hold = %q / %q", res.State, res.Format)
	}
	if res.Render.Enabled {
		t.Fatalf("the FETCHING render is disabled: %+v", res.Render)
	}
	if !strings.Contains(rig.logs(),
		"\U0001f590 Bot queue caught up, but 1 \U0001f590 hand row(s) are waiting for your real Chrome — scraping them now (one tab per link, the live pages become real fetches)") {
		t.Fatalf("the caught-up hand line missing:\n%s", rig.logs())
	}
	if strings.Contains(rig.logs(), "All caught up") {
		t.Fatalf("work started — no caught-up claim:\n%s", rig.logs())
	}
	waitSpoken(t, rig, "\U0001f916 Chrome tab retry: opening 1 link(s) in your real Chrome")
	// the delivery's own re-run batch (the delivered pages become real
	// fetches).
	waitBatchCalls(t, rig, 1)
}

func TestHeroUIFanfareCaughtUpRedoPass(t *testing.T) {
	rig := newFanfareUIRig(t)
	vault := rigConfigVault(t, rig)
	seedFalseSuccess(t, vault, "https://example.com/false-success", "HTTP 403")
	writeHandRowsTable(t, vault, [][3]string{
		{"https://example.com/false-success", "\U0001f590", "HTTP 403"},
	})
	res := rig.ui.HeroClick("")
	// the redo batch tells its own story — no caught-up claim, and the
	// machine returns the quiet format ("" keeps the current bar).
	if strings.Contains(rig.logs(), "All caught up") {
		t.Fatalf("work started — no caught-up claim:\n%s", rig.logs())
	}
	if !strings.Contains(rig.logs(), "\U0001f501 1 hand note(s) NOT properly stored") {
		t.Fatalf("the redo line missing:\n%s", rig.logs())
	}
	if res.Format != "" {
		t.Fatalf("the redo case keeps the bar: %q", res.Format)
	}
	waitBatchCalls(t, rig, 1)
	rig.mu.Lock()
	modes := append([]string(nil), rig.batchModes...)
	rig.mu.Unlock()
	if len(modes) != 1 || modes[0] != "/direct" {
		t.Fatalf("the caught-up redo rides the direct batch: %v", modes)
	}
}

func TestHeroUIFanfareCaughtUpPlainClaimIntact(t *testing.T) {
	rig := newFanfareUIRig(t)
	// no hand rows, no false successes: the caught-up claim + the
	// drift seal stay exactly as T-77 landed them (the regression pin).
	res := rig.ui.HeroClick("")
	if res.State != app.HeroSync || res.Format != app.FormatReady {
		t.Fatalf("the plain caught-up = %q / %q", res.State, res.Format)
	}
	if !strings.Contains(rig.logs(), "\u2705 All caught up — nothing undone in the bot queue.") {
		t.Fatalf("the closing line missing:\n%s", rig.logs())
	}
	rig.mu.Lock()
	batches := rig.batchCalls
	rig.mu.Unlock()
	if batches != 0 {
		t.Fatalf("the plain claim starts nothing (have %d)", batches)
	}
}

func TestHeroUIChromeOfferAnswer(t *testing.T) {
	rig := newChromeDoorUIRig(t)

	// the DECLINE: the reference's own declined line (1258-1261) + no
	// delivery + nothing marked (a declined link stays offerable).
	rig.ui.ChromeOfferAnswer(nil)
	spokePerTest(t, rig.spoken, rig.mu, "\u23ed\ufe0f Chrome tab retry declined — the failed links keep waiting (their automatic retries continue; the master table's Status cell decides their fate).")
	rig.mu.Lock()
	calls := rig.chromeDeliverCalls
	rig.mu.Unlock()
	if calls != 0 {
		t.Fatalf("a decline never delivers (have %d)", calls)
	}
	links := rig.ui.settings.ChromeOfferFromResults([]githubpipeline.WebsiteResultEntry{
		{URL: "https://example.com/failed", Outcome: "failed", Error: "reset"},
	})
	if len(links) != 1 || links[0].URL != "https://example.com/failed" {
		t.Fatalf("the declined link stays offerable: %+v", links)
	}

	// the PICK: the offered set marks + the delivery starts + the
	// re-offer is quiet for the picked link.
	rig.ui.ChromeOfferAnswer(links)
	waitSpoken(t, rig, "\U0001f916 Chrome tab retry: opening 1 link(s) in your real Chrome")
	waitDeliverCalls(t, rig, 1)
	if again := rig.ui.settings.ChromeOfferFromResults([]githubpipeline.WebsiteResultEntry{
		{URL: "https://example.com/failed", Outcome: "failed", Error: "reset"},
	}); again != nil {
		t.Fatalf("the picked link is never re-offered this session: %+v", again)
	}
}

// TestHeroflowFanfareWiring pins the frontend's fanfare laws: the ONE
// chrome:offer listener, the SAME dialog for both call sites, the TWO
// answer doors (the manual CANCELLED vs the automatic DECLINED), and
// THE BINDING-IMPORT LAW (ChromeOfferAnswer joined the import).
func TestHeroflowFanfareWiring(t *testing.T) {
	src := heroflowSource(t)
	for _, want := range []string{
		"EventsOn('chrome:offer'",
		"openChromeRetryDialog(payload.links, (picked) => ChromeOfferAnswer(picked))",
		"openChromeRetryDialog(res.links,",
		"(picked) => MoreChromeTabRetryRun(picked)",
		"function openChromeRetryDialog(links, answer)",
		"await answer([]).catch",
		"await answer(picked).catch",
	} {
		if !strings.Contains(src, want) {
			t.Fatalf("the fanfare wiring missing %q", want)
		}
	}
	// THE BINDING-IMPORT LAW (the 044 lesson).
	if !strings.Contains(src, "ChromeBoot, ChromeOfferAnswer } from './wailsjs/go/main/HeroUI.js';") {
		t.Fatal("ChromeOfferAnswer must join the HeroUI import statement")
	}
}

// TestHeroflowExportZipRowLive — T-80's standing assignment landed:
// the 📤 Export Backup ZIP row is a live door — the import (THE
// BINDING-IMPORT LAW), the dispatch case with the in-flight belt (the
// reference's disabled "Exporting..." row), and the verdict box law.
func TestHeroflowExportZipRowLive(t *testing.T) {
	src := heroflowSource(t)
	for _, want := range []string{
		"{ label: '📤 Export Backup ZIP', live: 'exportZip'",
		"MoreExportBackupZip,\n", // the binding import's member
		"if (act.live === 'exportZip') {",
		"exportZipRunning = true;",
		"const res = await MoreExportBackupZip().catch(() => null);",
		"exportZipRunning = false;",
		"? 'Exporting...' : act.label",
	} {
		if !strings.Contains(src, want) {
			t.Fatalf("the export row's live wiring missing %q", want)
		}
	}
	// The honest note is gone (the row was born-disabled with it).
	for _, gone := range []string{
		"The export-zip engine rides the packaging phase",
		"more-item-soon', act.label);\n    item.type = 'button';\n    item.setAttribute('role', 'menuitem');\n    item.title = act.tip || '';\n    if (!act.live) item.disabled = true;\n    item.addEventListener('click', () => { closeMoreMenu(); runMoreAction(act); });\n    menu.appendChild(item);\n  }\n  document.getElementById('app')",
	} {
		if strings.Contains(src, gone) {
			t.Fatalf("stale shape still present: %q", gone[:40])
		}
	}
}

// TestHeroflowExportZipDoor — the shell door's own laws: the guard box
// (the No Vault pair), the save-dialog seam's shape (the title, the
// default filename's stamp law, the ZIP filter), the engine call, the
// desk lines (📤 / ❌), and the result boxes' exact titles.
func TestHeroflowExportZipDoor(t *testing.T) {
	src := readFileOrFatal(t, "heroui.go")
	for _, want := range []string{
		"func (u *HeroUI) MoreExportBackupZip() *MoreDoorResult",
		`app.ExportZipGuards(u.settings.PlainConfig())`,
		`runtime.SaveFileDialog(u.ctx, runtime.SaveDialogOptions{`,
		`Title:           "Export Vault ZIP"`,
		`DefaultFilename: app.ExportDefaultName(time.Now())`,
		`app.RunVaultExport(vault, zipPath, time.Now())`,
		`"📤 Exported: "+message`,
		`"❌ Export failed: "+pyErrShell(err)`,
		`Title: "Export Complete", Body: "Saved to:\n" + message`,
		`Title: "Export Failed", Body: pyErrShell(err)`,
	} {
		if !strings.Contains(src, want) {
			t.Fatalf("the export door's law missing %q", want)
		}
	}
	// The cancel law: an empty pick is a silent no-op.
	i := strings.Index(src, "if zipPath == \"\" {")
	if i < 0 || !strings.Contains(src[i:i+80], "return &MoreDoorResult{}") {
		t.Fatal("the save dialog's cancel law missing")
	}
}

// readFileOrFatal is the shell-source reader (the tests' shared twin).
func readFileOrFatal(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(name)
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return string(data)
}
