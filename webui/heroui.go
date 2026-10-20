// heroui.go — T-77: the main window chrome's binding wrapper (the
// SIXTH bound struct): the hero button machine's shell doors (the ONE
// click door with the whole flow — the guards, the fetch, the process,
// the stop), the Test Connectivity battery (the ConnectionTestDialog's
// engine + events), and the More menu's live doors. The engine laws
// live in internal/app (herostate.go — the pure machine, the paint
// laws, the formats); this file binds them to the wails window, the
// established thin-shell law (doc 04 §3: "ui — thin shells over app").
//
// THE HERO FLOW (hero.py's two-stage law): SYNC → the queue fetch over
// Commands.SyncFetchQueue (Option B's Worker poll — the reference's
// Telethon check_bot_queue) → PROCESS (the fetched queue, or the Input
// page's import file when nothing was fetched) → Stop while the batch
// runs (the cooperative between-items stop — the SAME button). The
// state crossings ride the hero:state event (the render + the progress
// row's live format); the 200 ms mirror (the reference's
// _run_mirror_timer) keeps the button honest against batches other
// doors started.
//
// THE BATTERY (test_connection_modal.py): the four rows (Vaults, LLM,
// GitHub, the Worker — ADR-0004 Am.3's re-scope) render live from the
// cc:section/cc:result events; Start Syncing unlocks only when every
// row is ok (the payload carries the gate verdict).
package main

import (
	"context"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/assadigit/second-brain-project/internal/app"
	"github.com/assadigit/second-brain-project/internal/connectioncheck"
	"github.com/assadigit/second-brain-project/internal/websitepipeline"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// HeroUI is the main window chrome's bound surface (the wailsjs
// bindings land in frontend/wailsjs/go/main/HeroUI.js).
type HeroUI struct {
	settings *app.Settings
	commands *app.Commands
	ctx      context.Context

	// machine is hero.py's GUI-only state (the pure engine).
	machine *app.HeroMachine

	// the fetched queue's facts (the process leg reads them; the
	// confirm gate counts them).
	pendingRepos     []string
	pendingWebsites  []string
	intakeDuplicates int
	rawCount         int

	// mu serializes the machine + the queue facts (wails runs each
	// bound call on its own goroutine; the mirror ticks alongside).
	mu sync.Mutex

	// mirrorCancel stops the 200 ms mirror loop at shutdown.
	mirrorCancel context.CancelFunc
}

// NewHeroUI wires the chrome's shell over the Settings + Commands
// surfaces.
func NewHeroUI(settings *app.Settings, commands *app.Commands) *HeroUI {
	return &HeroUI{
		settings: settings,
		commands: commands,
		machine:  app.NewHeroMachine(),
	}
}

// bindCtx captures the wails context + starts the mirror loop (the
// reference's _run_mirror_timer, 200 ms — the defensive law: a batch
// another door started flips this button to Stop whatever the local
// state thinks).
func (u *HeroUI) bindCtx(ctx context.Context) {
	u.ctx = ctx
	mctx, cancel := context.WithCancel(context.Background())
	u.mirrorCancel = cancel
	go u.mirrorLoop(mctx)
}

// mirrorLoop is _run_mirror_timer: every 200 ms, mirror the batch
// flag onto the machine; a changed render crosses as hero:state.
func (u *HeroUI) mirrorLoop(ctx context.Context) {
	t := time.NewTicker(200 * time.Millisecond)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			if u.commands == nil {
				continue
			}
			u.mu.Lock()
			changed := u.machine.Mirror(u.commands.BatchBusy())
			u.mu.Unlock()
			if changed {
				u.emitState("")
			}
		}
	}
}

// stopMirror halts the loop (the shell's shutdown door calls it).
func (u *HeroUI) stopMirror() {
	if u.mirrorCancel != nil {
		u.mirrorCancel()
	}
}

// ---------------------------------------------------------------------------
// The state crossing.
// ---------------------------------------------------------------------------

// HeroStatePayload is the hero:state event + the evidence door's
// shape: the machine's state word, the button's render, and the
// progress row's live format + count ("" = keep the current).
type HeroStatePayload struct {
	State  string         `json:"state"`
	Render app.HeroRender `json:"render"`
	Format string         `json:"format"`
	Count  string         `json:"count"`
}

// emitState crosses the machine's current render (and, when non-empty,
// the progress row's format + count) as the hero:state event.
func (u *HeroUI) emitState(format string) {
	if u.ctx == nil {
		return
	}
	u.mu.Lock()
	payload := HeroStatePayload{
		State:  u.machine.State(),
		Render: u.machine.Render(),
		Format: format,
	}
	u.mu.Unlock()
	runtime.EventsEmit(u.ctx, "hero:state", payload)
}

// applySentinel records the sentinel's online verdict onto the machine
// (the offline advisory's seat — hero.py's _sentinel_online) and
// refreshes the PROCESS count when a queue check landed fresh counts.
func (u *HeroUI) applySentinel(ev app.Sentinel) {
	online := ev.Net.Online
	u.mu.Lock()
	u.machine.SetSentinelOnline(&online)
	u.mu.Unlock()
}

// onBatchFinished is the batch-finished crossing's hero half: the
// cleared flag restores SYNC (the mirror's belt catches it too — this
// is the deterministic first path).
func (u *HeroUI) onBatchFinished() {
	u.mu.Lock()
	changed := u.machine.Mirror(false)
	u.mu.Unlock()
	if changed {
		u.emitState("")
	}
}

// ---------------------------------------------------------------------------
// THE HERO CLICK — the one door (hero.py's _on_hero_clicked + the
// stage legs).
// ---------------------------------------------------------------------------

// HeroClickResult is the click door's verdict: the state crossing +
// the guard boxes the frontend renders. The blocking legs (the fetch,
// the batch) run INSIDE the door — the desk events flow on their own
// transport while the promise waits (the StartScan law).
type HeroClickResult struct {
	State  string         `json:"state"`
	Render app.HeroRender `json:"render"`
	Format string         `json:"format"`
	Count  string         `json:"count"`
	// Box: a themed message box to render (the vault guard, the
	// nothing-to-sync guard) — the click's whole answer.
	Box *app.HeroBox `json:"box"`
	// Offline: the continue-anyway question — the frontend asks and
	// answers through HeroAnswerOffline.
	Offline bool `json:"offline"`
	// Confirm: the >10 process confirm gate's box — Continue runs
	// HeroRunQueue.
	Confirm *app.HeroBox `json:"confirm"`
}

func (u *HeroUI) snapshotResult(format, count string) *HeroClickResult {
	u.mu.Lock()
	defer u.mu.Unlock()
	return &HeroClickResult{
		State:  u.machine.State(),
		Render: u.machine.Render(),
		Format: format,
		Count:  count,
	}
}

// HeroClick is THE hero button's one click door. importFile is the
// Input page's live import-file field (the reference reads the widget;
// the port's save-on-change law makes the live field the frontend's to
// pass — empty when no file is picked).
func (u *HeroUI) HeroClick(importFile string) *HeroClickResult {
	vault := u.settings.LiveVaultPath()
	vaultValid := vault != "" && isDirPath(vault)
	queueConfigured := u.settings.WorkerQueueEnabled()
	importReady := strings.TrimSpace(importFile) != ""

	u.mu.Lock()
	plan := u.machine.Click(vaultValid, queueConfigured, importReady)
	u.mu.Unlock()

	switch plan.Action {
	case app.HeroActStop:
		return u.stopLeg()
	case app.HeroActProcess:
		return u.processLeg(false)
	case app.HeroActImport:
		return u.importLeg(importFile)
	case app.HeroActFetch:
		return u.fetchLeg(importFile)
	}
	// The guards: the boxes / the offline ask are the whole answer.
	res := u.snapshotResult("", "")
	res.Box = plan.VaultBox
	if plan.VaultBox == nil {
		res.Box = plan.NothingBox
	}
	res.Offline = plan.OfflineAsk
	return res
}

// HeroAnswerOffline answers the continue-anyway question (a "no" stays
// in SYNC — retry when back online; a "yes" speaks the owner's-word
// line and continues the stage-1 guards).
func (u *HeroUI) HeroAnswerOffline(goOnline bool, importFile string) *HeroClickResult {
	queueConfigured := u.settings.WorkerQueueEnabled()
	importReady := strings.TrimSpace(importFile) != ""
	u.mu.Lock()
	plan, line := u.machine.AnswerOffline(goOnline, queueConfigured, importReady)
	u.mu.Unlock()
	if line.Msg != "" {
		u.settings.SpeakOnBus(line.Msg, line.Level)
	}
	switch plan.Action {
	case app.HeroActFetch:
		return u.fetchLeg(importFile)
	case app.HeroActImport:
		return u.importLeg(importFile)
	}
	res := u.snapshotResult("", "")
	res.Box = plan.NothingBox
	return res
}

// fetchLeg is stage 1 (check_bot_queue + _after_sync_fetch): the
// fetching render, the Worker queue poll, the classification lines,
// and the routing verdict — process (the fetched queue) / import-ready
// / caught-up (the waiting report + the drift seal + the closing line)
// / the error restore.
func (u *HeroUI) fetchLeg(importFile string) *HeroClickResult {
	u.settings.SpeakOnBus("📬 Checking bot queue (the Worker)…", "info")
	u.mu.Lock()
	u.machine.BeginFetch()
	u.mu.Unlock()
	u.emitState(app.FormatFetching)

	rep := u.commands.SyncFetchQueue(context.Background())
	if rep.State != "ok" {
		u.settings.SpeakOnBus("❌ Bot queue check failed: "+rep.Error, "error")
		u.mu.Lock()
		out := u.machine.ApplyFetchResult(false, 0, 0, false)
		u.mu.Unlock()
		u.emitState(out.ProgressFormat)
		return u.snapshotResult(out.ProgressFormat, "")
	}
	for _, l := range app.QueueLogLines(rep) {
		u.settings.SpeakOnBus(l.Msg, l.Level)
	}
	// The Sync view's freshness hook (the reference's
	// _refresh_sync_after_queue rides every check).
	u.settings.SetSyncQueueState(&app.SyncQueueState{
		Checked:         true,
		PendingRepos:    rep.PendingURLs,
		PendingWebsites: rep.PendingWebsites,
		InVault:         rep.InVault,
		WebsitesInVault: rep.WebsitesInVault,
	})
	u.mu.Lock()
	u.pendingRepos = rep.PendingURLs
	u.pendingWebsites = rep.PendingWebsites
	u.intakeDuplicates = rep.DuplicatesRemoved
	u.rawCount = rep.RawCount
	importReady := strings.TrimSpace(importFile) != ""
	out := u.machine.ApplyFetchResult(true, len(rep.PendingURLs),
		len(rep.PendingWebsites), importReady)
	u.mu.Unlock()
	for _, l := range out.Lines {
		u.settings.SpeakOnBus(l.Msg, l.Level)
	}
	if out.CaughtUp {
		waiting, eyes := u.masterWaitingRows()
		if waiting > 0 {
			if l := app.WaitingRowsLine(waiting, eyes); l.Msg != "" {
				u.settings.SpeakOnBus(l.Msg, l.Level)
			}
		}
		if l := app.CaughtUpLine(waiting); l.Msg != "" {
			u.settings.SpeakOnBus(l.Msg, l.Level)
		}
		// v0.89.0's DRIFT SEAL: a caught-up sync still seals the dirty
		// vaults + publishes the directory (the quiet law — a leg that
		// seals nothing and fails nothing stays silent).
		_ = u.commands.SyncDriftSeal(context.Background())
	}
	u.emitState(out.ProgressFormat)
	return u.snapshotResult(out.ProgressFormat, out.CountText)
}

// masterWaitingRows is the caught-up check's master-table read
// (_scan_master_waiting: the " - " rows, classified fetch vs eyes;
// best-effort — a broken probe never blocks the caught-up path).
func (u *HeroUI) masterWaitingRows() (waiting, eyes int) {
	defer func() { recover() }() // best-effort, exactly like the reference
	cfg := u.settings.PlainConfig()
	wv := strings.TrimSpace(pyCfgStr(cfg, "website_vault_path"))
	if wv == "" || !isDirPath(wv) {
		return 0, 0
	}
	rows := websitepipeline.ScanMasterWaitingRows(wv, nil,
		func(msg, level string) { u.settings.SpeakOnBus(msg, level) })
	for _, r := range rows {
		switch r.Kind {
		case "fetch":
			waiting++
		case "eyes":
			eyes++
		}
	}
	return waiting, eyes
}

// processLeg is stage 2 (process_bot_queue): the >10 confirm gate
// first (the frontend asks; Continue calls HeroRunQueue), then the
// batch over the fetched queue.
func (u *HeroUI) processLeg(confirmed bool) *HeroClickResult {
	u.mu.Lock()
	n := len(u.pendingRepos) + len(u.pendingWebsites)
	u.mu.Unlock()
	if !confirmed {
		if ask, title, body := app.ConfirmLargeBatch(n, "the bot queue"); ask {
			res := u.snapshotResult("", "")
			res.Confirm = &app.HeroBox{Title: title, Body: body, Question: true}
			return res
		}
	}
	// The batch's whole story rides the bus (the desk lines, the
	// progress, the finish); the click's answer is the state crossing.
	_ = u.runQueueBatch()
	return u.snapshotResult("", "")
}

// HeroRunQueue is the confirmed process door: the batch over the
// fetched queue (both pipelines — v0.24.1's law). The blocking run
// rides inside the door; the desk events flow on the bus while the
// promise waits; the finish restores SYNC (the batch-finished
// crossing's own half).
func (u *HeroUI) HeroRunQueue() *app.BatchResult {
	return u.runQueueBatch()
}

func (u *HeroUI) runQueueBatch() *app.BatchResult {
	u.mu.Lock()
	urls := append([]string(nil), u.pendingRepos...)
	webs := append([]string(nil), u.pendingWebsites...)
	dupes, raw := u.intakeDuplicates, u.rawCount
	u.machine.Mirror(true)
	u.mu.Unlock()
	u.emitState("")

	res := u.commands.SyncRunBatch(context.Background(), app.BatchOptions{
		URLs:             urls,
		NonGithub:        webs,
		Source:           "bot",
		IntakeDuplicates: dupes,
		RawURLCount:      raw,
	})
	u.onBatchFinished()
	return res
}

// importLeg is the legacy input-mode path (start_processing, v0.23.0's
// ONE remaining input mode): the import batch over the picked file.
func (u *HeroUI) importLeg(importFile string) *HeroClickResult {
	path := strings.TrimSpace(importFile)
	if path == "" {
		// The guard said a file was picked; the field went empty since
		// — the honest refusal, not a crashed batch.
		res := u.snapshotResult("", "")
		res.Box = &app.HeroBox{Title: "Nothing to import",
			Body: "The import file field is empty — pick a .txt / .md file in Settings → Input first."}
		return res
	}
	u.mu.Lock()
	u.machine.Mirror(true)
	u.mu.Unlock()
	u.emitState("")
	_ = u.commands.SyncRunBatch(context.Background(), app.BatchOptions{
		Mode:       "import",
		ImportFile: path,
		Source:     "import",
	})
	u.onBatchFinished()
	return u.snapshotResult("", "")
}

// stopLeg is stop_processing: the cooperative stop + the winding-down
// render ("Stopping…", disabled — no double-click).
func (u *HeroUI) stopLeg() *HeroClickResult {
	u.settings.SpeakOnBus("⏹️ Stopping...", "warning")
	_ = u.commands.StopBatch()
	u.mu.Lock()
	u.machine.BeginStopping()
	u.mu.Unlock()
	u.emitState("")
	return u.snapshotResult("", "")
}

// ChromeBoot is the boot paint's facts (the frontend applies them at
// first frame): the persisted dark_mode preference (the reference
// restores it before the window shows).
func (u *HeroUI) ChromeBoot() *ChromeBootPayload {
	dark := false
	if v, ok := u.settings.PlainConfig().Get("dark_mode"); ok {
		if b, is := v.(bool); is {
			dark = b
		}
	}
	return &ChromeBootPayload{DarkMode: dark}
}

// ChromeBootPayload is ChromeBoot's shape.
type ChromeBootPayload struct {
	DarkMode bool `json:"dark_mode"`
}

// HeroState is the evidence door: the machine's current crossing.
func (u *HeroUI) HeroState() *HeroStatePayload {
	u.mu.Lock()
	defer u.mu.Unlock()
	return &HeroStatePayload{
		State:  u.machine.State(),
		Render: u.machine.Render(),
	}
}

// ---------------------------------------------------------------------------
// THE BATTERY — the Test Connectivity door (test_connection_modal.py).
// ---------------------------------------------------------------------------

// CCRowResult is one battery row's final shape (the modal renders the
// events live; this is the settled verdict the gate reads).
type CCRowResult struct {
	Icon    string   `json:"icon"`
	Title   string   `json:"title"`
	Verdict string   `json:"verdict"` // ok | warn | error
	Items   []string `json:"items"`   // the short lines (the full text rode the events)
}

// CCTestResult is the battery door's payload: the four rows + the
// Start Syncing gate verdict + the one-line headline.
type CCTestResult struct {
	Refusal  string        `json:"refusal"`
	Rows     []CCRowResult `json:"rows"`
	AllOK    bool          `json:"all_ok"`
	Headline string        `json:"headline"`
	Level    string        `json:"level"`
}

// the modal's own row icons (the reference's four-row lockup, the
// Worker leg promoted per ADR-0004 Am.3).
var ccRowLockup = []struct{ icon, title string }{
	{"📁", "Vaults"},
	{"🧠", "LLM"},
	{"🐙", "GitHub"},
	{"☁️", "Worker"},
}

// TestConnection is the 🔌 Test Connectivity door (the reference's
// test_all): the guards, the four-section battery over the live
// config, the live events (cc:section / cc:result) driving the modal's
// rows, and the settled verdict — Start Syncing unlocks only when
// every row is ok.
func (u *HeroUI) TestConnection() *CCTestResult {
	if u.commands.BatchBusy() {
		line := "⏳ A batch is running — Test Connection is available when it finishes."
		u.settings.SpeakOnBus(line, "warning")
		return &CCTestResult{Refusal: line}
	}
	u.settings.SpeakOnBus("🔍 Test Connection — checking vaults, LLM, GitHub and the sync queue…", "info")

	emit := func(name string, payload any) {
		if u.ctx != nil {
			runtime.EventsEmit(u.ctx, name, payload)
		}
	}
	// The battery reads the LIVE config (the OrderedMap rides its own
	// cfgGet — the reference's snapshot is the same live values under
	// the port's save-on-change law).
	cfg := u.settings.PlainConfig()

	sections := connectioncheck.RunLocalChecks(cfg,
		func(title string, idx, total int) {
			emit("cc:section", map[string]any{"title": title, "idx": idx, "total": total})
		},
		func(r connectioncheck.Result) {
			// The log streams the full render line (the reference's
			// line-by-line law); the modal's row gets the compact one.
			u.settings.SpeakOnBus("   "+connectioncheck.RenderLine(r),
				connectioncheck.GUILevels[r.Level])
			emit("cc:result", map[string]any{
				"name": r.Name, "level": r.Level,
				"short": connectioncheck.ShortLine(r),
				"full":  connectioncheck.RenderLine(r),
			})
		})
	if len(sections) == 0 {
		// The battery itself crashed — NEVER report all-ok from an
		// empty result; surface the crash instead (the reference's law).
		sections = []connectioncheck.Section{{
			Title: "Check",
			Results: []connectioncheck.Result{{
				Name: "Test battery", Level: "error", Detail: "the check crashed",
			}},
		}}
	}

	rows := make([]CCRowResult, 0, len(sections))
	allOK := true
	for i, sec := range sections {
		verdict := "ok"
		row := CCRowResult{
			Icon:  ccRowLockup[i%len(ccRowLockup)].icon,
			Title: sec.Title,
		}
		if i < len(ccRowLockup) && sec.Title != ccRowLockup[i].title {
			// the degenerate crash section keeps its own title
			row.Title = sec.Title
		}
		for _, r := range sec.Results {
			row.Items = append(row.Items, connectioncheck.ShortLine(r))
			switch r.Level {
			case "error":
				verdict = "error"
			case "warn":
				if verdict != "error" {
					verdict = "warn"
				}
			}
		}
		if verdict != "ok" {
			allOK = false
		}
		row.Verdict = verdict
		rows = append(rows, row)
	}
	sum := connectioncheck.Summarize(sections)
	u.settings.SpeakOnBus("🏁 Test Connection — "+sum.Headline,
		connectioncheck.GUILevels[sum.Level])
	res := &CCTestResult{
		Rows:     rows,
		AllOK:    allOK,
		Headline: sum.Headline,
		Level:    connectioncheck.GUILevels[sum.Level],
	}
	emit("cc:done", res)
	return res
}

// ---------------------------------------------------------------------------
// THE MORE MENU's live doors.
// ---------------------------------------------------------------------------

// MoreDoorResult is a More action's honest verdict ("" = the door
// ran; a title renders the themed box).
type MoreDoorResult struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

// MoreOpenDashboard is the 📊 Open Dashboard door: the URL law (the
// config's Worker URL, the honest refusal when unset) + the browser.
func (u *HeroUI) MoreOpenDashboard() *MoreDoorResult {
	title, body := u.settings.OpenDashboardBrowser("")
	return &MoreDoorResult{Title: title, Body: body}
}

// MoreRetryBacklog is the 🔁 Retry _review backlog door (v0.42.0's
// anytime trigger): the review_retry batch mode — the app-owned
// fetch-failed placeholders re-fetched through the full pipeline.
func (u *HeroUI) MoreRetryBacklog() *app.BatchResult {
	return u.moreRetryBatch("review_retry")
}

// MoreRetryMaster is the 🔁 Retry the table's ' - ' rows door
// (v0.51.0's anytime trigger): the master_retry batch mode — the
// master table's waiting rows fetched again before "everything is up
// to date" may be said.
func (u *HeroUI) MoreRetryMaster() *app.BatchResult {
	return u.moreRetryBatch("master_retry")
}

func (u *HeroUI) moreRetryBatch(mode string) *app.BatchResult {
	u.mu.Lock()
	u.machine.Mirror(true)
	u.mu.Unlock()
	u.emitState("")
	res := u.commands.SyncRunBatch(context.Background(), app.BatchOptions{
		Mode:   mode,
		Source: "bot",
	})
	u.onBatchFinished()
	return res
}

// MoreDarkMode is the ⚙️ Settings ▸ 🌙 Dark Mode toggle's save door
// (the theme flip itself is the frontend's — the config persists the
// preference exactly like the reference).
func (u *HeroUI) MoreDarkMode(on bool) *app.SaveResult {
	return u.settings.SaveDarkMode(on)
}

// ---------------------------------------------------------------------------
// THE FIVE PER-TEST DOORS (T-79a) — the More menu's per-test dialogs,
// the reference's connection_tests.py + llamacpp.py flows over the
// Settings engines. Each door streams its own line grammar onto the
// desk while it runs; the two verdict boxes (Cloud + llama.cpp) ride
// the same themed-box shape the dashboard door uses (title "" = no
// box — the log-only doors).
// ---------------------------------------------------------------------------

// MoreTestTelegramGitHub is 🔑 Test Telegram & GitHub (the Option-B
// re-scope: the GitHub token half + the Worker queue leg in the
// Telegram half's seat — ADR-0004 Am.3).
func (u *HeroUI) MoreTestTelegramGitHub() *MoreDoorResult {
	title, body := u.settings.TestTelegramGitHub()
	return &MoreDoorResult{Title: title, Body: body}
}

// MoreTestProxy is 🌐 Test Proxy Connection (the re-scope: the
// sentinel's ProxyReachable TCP verdict).
func (u *HeroUI) MoreTestProxy() *MoreDoorResult {
	title, body := u.settings.TestProxy()
	return &MoreDoorResult{Title: title, Body: body}
}

// MoreTestOllama is 🧠 Test Ollama — the models + the real inference
// ping ("What is 2 + 3?").
func (u *HeroUI) MoreTestOllama() *MoreDoorResult {
	title, body := u.settings.TestOllama()
	return &MoreDoorResult{Title: title, Body: body}
}

// MoreTestCloud is 🔌 Test Cloud API — the /models pre-flight + the
// "Say hello" ping + the verdict box.
func (u *HeroUI) MoreTestCloud() *MoreDoorResult {
	title, body := u.settings.TestCloudLLM()
	return &MoreDoorResult{Title: title, Body: body}
}

// MoreTestLlamaCpp is 🦙 Test llama.cpp — the /props probe + the
// "Say hello" ping + the verdict box.
func (u *HeroUI) MoreTestLlamaCpp() *MoreDoorResult {
	title, body := u.settings.TestLlamaCpp()
	return &MoreDoorResult{Title: title, Body: body}
}

// ---------------------------------------------------------------------------
// THE LINKING-TOOLS DOORS (T-79b) -- the More menu's second door
// family: the reference's SAFE defaults over the ported engines
// (linking_tools.py:107/116 + dashboard.py:313/500 +
// connection_tests.py:219). The two tool doors stream their captured
// tail at completion (the TestWorker law); Verify Vault answers the
// report box; Recategorize is the two-door dialog (the open walk +
// the apply pass).
// ---------------------------------------------------------------------------

// MoreValidateVault is 📁 Validate Vault -- the fast inline
// flow (log-only, no box).
func (u *HeroUI) MoreValidateVault() *MoreDoorResult {
	u.settings.ValidateVault()
	return &MoreDoorResult{}
}

// MoreVerifyVault is 🔍 Verify Vault -- the manifest
// verification report (the box carries the report the reference
// renders into its dashboard text area).
func (u *HeroUI) MoreVerifyVault() *MoreDoorResult {
	title, body := u.settings.VerifyVault()
	return &MoreDoorResult{Title: title, Body: body}
}

// MoreRecallHooksDryRun is 🪝 Recall hooks (dry-run) -- the
// Phase-6 step-1 tool in its default dry-run mode through the
// TestWorker law (log-only).
func (u *HeroUI) MoreRecallHooksDryRun() *MoreDoorResult {
	u.settings.RecallHooksDryRun()
	return &MoreDoorResult{}
}

// MoreBuildLinkSuggestions is 🔗 Build link suggestions --
// the suggest-mode tool through the TestWorker law (log-only).
func (u *HeroUI) MoreBuildLinkSuggestions() *MoreDoorResult {
	u.settings.BuildLinkSuggestions()
	return &MoreDoorResult{}
}

// MoreRecategorizeOpen is 📁 Recategorize Notes' open door:
// the vault walk + the table payload (a box title answers the
// reference's custom message boxes instead of the table).
func (u *HeroUI) MoreRecategorizeOpen() *app.RecategorizeModel {
	return u.settings.RecategorizeOpen()
}

// MoreRecategorizeApply is the recategorize dialog's Apply door: the
// moves + the cache updates + the finish line.
func (u *HeroUI) MoreRecategorizeApply(
	changes []app.RecategorizeChange) *app.RecategorizeResult {
	return u.settings.RecategorizeApply(changes)
}

func isDirPath(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.IsDir()
}

// pyCfgStr reads one string key off a *storage.OrderedMap (the shell's
// local twin of the app package's unexported helper).
func pyCfgStr(m any, key string) string {
	type getter interface {
		Get(string) (any, bool)
	}
	if g, ok := m.(getter); ok {
		if v, found := g.Get(key); found {
			if s, is := v.(string); is {
				return s
			}
		}
	}
	return ""
}
