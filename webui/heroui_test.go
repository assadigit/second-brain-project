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
	"github.com/assadigit/second-brain-project/internal/llmclient"
	"github.com/assadigit/second-brain-project/internal/mirror"
	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/store"
)

// heroRig is the shared wiring harness: faked queue/battery seams,
// a log-capturing subscriber, and the HeroUI wrapper.
type heroRig struct {
	ui       *HeroUI
	commands *app.Commands
	cfg      *storage.OrderedMap
	spoken   *[]string
	mu       *sync.Mutex

	queueState  string
	queueItems  []any
	batchCalls  int
	batchModes  []string
	batchResult *app.BatchResult
	batchBlock  chan struct{} // non-nil → the fake batch waits for the close

	// botQueueItems feeds the Settings' BotQueueFetch seam (T-79c's
	// doors — independent of the Commands' queue fakes).
	botQueueItems []any
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
			rig.mu.Unlock()
			if rig.batchBlock != nil {
				select {
				case <-rig.batchBlock:
				case <-ctx.Done():
				}
			}
			if rig.batchResult != nil {
				return rig.batchResult
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
		"'on — cloud'", "'on — local ('", "'Proxy: connected'", "'Proxy: off'", "'Proxy: unreachable'",
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
	// the honest-disabled law still holds — for the REMAINING tail
	// (T-79d/e), never for the per-test, linking, or bot-queue doors.
	for _, soon := range []string{
		"The Chrome doors land with the Chrome-doors GUI task",
		"The graveyard procedure lands with the master-table tools task",
		"The export-zip engine rides the packaging phase",
	} {
		if !strings.Contains(src, soon) {
			t.Fatalf("the remaining tail's landing note missing %q", soon)
		}
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
	// the honest-disabled law still holds — for the REMAINING tail
	// (T-79d/T-79e), never for the bot-queue doors.
	for _, soon := range []string{
		"The Chrome doors land with the Chrome-doors GUI task",
		"The graveyard procedure lands with the master-table tools task",
		"The export-zip engine rides the packaging phase",
	} {
		if !strings.Contains(src, soon) {
			t.Fatalf("the remaining tail's landing note missing %q", soon)
		}
	}
	if strings.Contains(src, "The bot-queue tools land with the queue-tools task") {
		t.Fatal("the T-79c doors are live — their landing note is gone")
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
