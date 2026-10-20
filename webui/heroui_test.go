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
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/assadigit/second-brain-project/internal/app"
	"github.com/assadigit/second-brain-project/internal/connectioncheck"
	"github.com/assadigit/second-brain-project/internal/llmclient"
	"github.com/assadigit/second-brain-project/internal/storage"
)

// heroRig is the hermetic chrome: a Settings + Commands pair over
// faked queue/battery seams, a log-capturing subscriber, and the
// HeroUI wrapper.
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
	// (T-79b..e), never for the per-test doors.
	for _, soon := range []string{
		"The linking tools land with their own task",
		"The bot-queue tools land with the queue-tools task",
		"The Chrome doors land with the Chrome-doors GUI task",
		"The graveyard procedure lands with the master-table tools task",
		"The export-zip engine rides the packaging phase",
	} {
		if !strings.Contains(src, soon) {
			t.Fatalf("the remaining tail's landing note missing %q", soon)
		}
	}
}
