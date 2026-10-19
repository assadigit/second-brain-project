// herostate_test.go — T-77's acceptance battery, part 1: the hero
// state machine (the spec's own list): the born law + the SYNC render's
// exact words; the click routing through every guard (the vault box,
// the offline advisory's ask + the owner's-word-wins continuation, the
// nothing-to-sync split); the fetch routing (the fetched-queue flip
// with the both-pipelines arithmetic + the 🟢 line, the import-file
// ready path, the caught-up law's closing-line choice, the error
// restore); the mirror (the missed running transition, the cleared-flag
// SYNC restore, the fetching keep); the stopping render; the >10
// confirm gate; the sentinel labels' paint laws; the progress row's
// live formats.
package app

import (
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/assadigit/second-brain-project/internal/connectioncheck"
)

// ---------------------------------------------------------------------------
// The born state + the renders.
// ---------------------------------------------------------------------------

func TestHeroBornStateAndSyncRender(t *testing.T) {
	m := NewHeroMachine()
	if m.State() != HeroSync {
		t.Fatalf("born state = %q, want sync", m.State())
	}
	r := m.Render()
	if r.Label != "SYNC" || r.Glyph != "refresh" || !r.Enabled || r.Danger {
		t.Fatalf("SYNC render = %+v", r)
	}
	// The FULL SYNC tooltip's own sentences (hero.py:474-486).
	for _, want := range []string{
		"FULL SYNC — Stage 1: fetch every UNDONE item from",
		"The button then becomes PROCESS — click it to start the batch.",
		"While a batch runs this button becomes Stop — click to cancel.",
		"⚡ Fast Sync",
	} {
		if !strings.Contains(r.Tooltip, want) {
			t.Fatalf("SYNC tooltip missing %q:\n%s", want, r.Tooltip)
		}
	}
}

func TestHeroFetchingRenderIsDisabled(t *testing.T) {
	m := NewHeroMachine()
	m.BeginFetch()
	if m.State() != HeroFetching {
		t.Fatalf("state = %q, want fetching", m.State())
	}
	r := m.Render()
	if r.Label != "FETCHING…" || r.Glyph != "loader" || r.Enabled {
		t.Fatalf("FETCHING render = %+v", r)
	}
	if r.Tooltip != "Fetching undone items from the Telegram bot…" {
		t.Fatalf("FETCHING tooltip = %q", r.Tooltip)
	}
	// A fossil click while fetching does nothing.
	if plan := m.Click(true, true, false); plan.Action != HeroActNone {
		t.Fatalf("fetching click = %+v, want none", plan)
	}
}

func TestHeroProcessRenderCountsAndTooltips(t *testing.T) {
	m := NewHeroMachine()
	m.SetCounts(4, 3)
	out := m.ApplyFetchResult(true, 4, 3, false)
	if out.NewState != HeroProcess {
		t.Fatalf("state = %q, want process", out.NewState)
	}
	r := m.Render()
	if r.Label != "PROCESS (7)" || r.Glyph != "play" || !r.Enabled || r.Danger {
		t.Fatalf("PROCESS render = %+v", r)
	}
	if !strings.Contains(r.Tooltip, "start curating the 7 fetched undone item(s)") ||
		!strings.Contains(r.Tooltip, "more than 10 items") {
		t.Fatalf("PROCESS tooltip = %q", r.Tooltip)
	}
	// The zero-count rendering: nothing fetched — the Input mode words.
	m2 := NewHeroMachine()
	m2.ApplyFetchResult(true, 0, 0, true)
	r2 := m2.Render()
	if r2.Label != "PROCESS" {
		t.Fatalf("zero-count label = %q", r2.Label)
	}
	if !strings.Contains(r2.Tooltip, "runs the selected Input mode") {
		t.Fatalf("zero-count tooltip = %q", r2.Tooltip)
	}
}

func TestHeroRunningAndStoppingRenders(t *testing.T) {
	m := NewHeroMachine()
	if !m.Mirror(true) {
		t.Fatalf("the missed running transition must flip")
	}
	r := m.Render()
	if r.Label != "Stop" || r.Glyph != "stop" || !r.Enabled || !r.Danger {
		t.Fatalf("Stop render = %+v", r)
	}
	if !strings.Contains(r.Tooltip, "The stop is clean — already-processed notes are kept") {
		t.Fatalf("Stop tooltip = %q", r.Tooltip)
	}
	// stop_processing: the button waits disabled "Stopping…" — no double-click.
	m.BeginStopping()
	r = m.Render()
	if r.Label != "Stopping…" || r.Enabled || !r.Danger {
		t.Fatalf("Stopping render = %+v", r)
	}
	if plan := m.Click(true, true, false); plan.Action != HeroActNone {
		t.Fatalf("a winding-down click = %+v, want none", plan)
	}
	// The batch's finish restores SYNC.
	if !m.Mirror(false) {
		t.Fatalf("the cleared flag must restore SYNC")
	}
	if m.State() != HeroSync || m.Render().Label != "SYNC" {
		t.Fatalf("post-finish state/render = %q/%s", m.State(), m.Render().Label)
	}
}

// ---------------------------------------------------------------------------
// The click routing + the guards.
// ---------------------------------------------------------------------------

func TestHeroClickSyncRunsTheFetch(t *testing.T) {
	m := NewHeroMachine()
	plan := m.Click(true, true, false)
	if plan.Action != HeroActFetch || plan.VaultBox != nil || plan.OfflineAsk ||
		plan.NothingBox != nil {
		t.Fatalf("plan = %+v, want the fetch", plan)
	}
}

func TestHeroClickInvalidVaultBox(t *testing.T) {
	m := NewHeroMachine()
	plan := m.Click(false, true, false)
	if plan.Action != HeroActNone || plan.VaultBox == nil {
		t.Fatalf("plan = %+v, want the vault box", plan)
	}
	if plan.VaultBox.Title != "Error" ||
		plan.VaultBox.Body != "Please select a valid Obsidian vault path." {
		t.Fatalf("vault box = %+v", plan.VaultBox)
	}
	if m.State() != HeroSync {
		t.Fatalf("a guard box never moves the state; got %q", m.State())
	}
}

func TestHeroClickOfflineAdvisory(t *testing.T) {
	m := NewHeroMachine()
	offline := false
	m.SetSentinelOnline(&offline)
	// The ask fires BEFORE the queue guards.
	plan := m.Click(true, true, false)
	if plan.Action != HeroActNone || !plan.OfflineAsk {
		t.Fatalf("plan = %+v, want the offline ask", plan)
	}
	// "No" stays in SYNC — retry when back online.
	plan2, line := m.AnswerOffline(false, true, false)
	if plan2.Action != HeroActNone || line.Msg != "" {
		t.Fatalf("refused answer = %+v / %+v", plan2, line)
	}
	if m.State() != HeroSync {
		t.Fatalf("a refused ask stays in SYNC; got %q", m.State())
	}
	// "Continue anyway" speaks the owner's-word-wins line and continues.
	plan3, line3 := m.AnswerOffline(true, true, false)
	if plan3.Action != HeroActFetch {
		t.Fatalf("continued answer = %+v, want the fetch", plan3)
	}
	if line3.Msg != "⚠️ Syncing despite the offline verdict (the owner's word wins)." ||
		line3.Level != "warning" {
		t.Fatalf("the warning line = %+v", line3)
	}
	// A not-yet-observed sentinel (nil) never triggers the ask.
	m2 := NewHeroMachine()
	if plan := m2.Click(true, true, false); plan.OfflineAsk {
		t.Fatalf("an unobserved sentinel must stay silent: %+v", plan)
	}
	// An online verdict neither.
	online := true
	m2.SetSentinelOnline(&online)
	if plan := m2.Click(true, true, false); plan.OfflineAsk || plan.Action != HeroActFetch {
		t.Fatalf("an online sentinel must not ask: %+v", plan)
	}
}

func TestHeroClickNothingToSyncSplit(t *testing.T) {
	m := NewHeroMachine()
	// No queue configured + an import file picked → the import batch
	// (v0.23.0's ONE remaining input mode from the main button).
	if plan := m.Click(true, false, true); plan.Action != HeroActImport {
		t.Fatalf("import split = %+v", plan)
	}
	// No queue + no file → the nothing-to-sync box (re-worded for the
	// Worker queue per ADR-0004 Am.3).
	plan := m.Click(true, false, false)
	if plan.Action != HeroActNone || plan.NothingBox == nil {
		t.Fatalf("plan = %+v, want the nothing box", plan)
	}
	if plan.NothingBox.Title != "Nothing to sync" ||
		!strings.Contains(plan.NothingBox.Body, "Settings → Bot") ||
		!strings.Contains(plan.NothingBox.Body, "Settings → Input") {
		t.Fatalf("nothing box = %+v", plan.NothingBox)
	}
}

func TestHeroClickProcessAndStopRouting(t *testing.T) {
	m := NewHeroMachine()
	m.ApplyFetchResult(true, 5, 2, false)
	// Pending work → the queue batch.
	if plan := m.Click(true, true, false); plan.Action != HeroActProcess {
		t.Fatalf("process click = %+v", plan)
	}
	// Nothing pending → the Input mode.
	m2 := NewHeroMachine()
	m2.ApplyFetchResult(true, 0, 0, true)
	if plan := m2.Click(true, true, false); plan.Action != HeroActImport {
		t.Fatalf("empty process click = %+v", plan)
	}
	// Running → Stop.
	m3 := NewHeroMachine()
	m3.Mirror(true)
	if plan := m3.Click(true, true, false); plan.Action != HeroActStop {
		t.Fatalf("running click = %+v", plan)
	}
}

// ---------------------------------------------------------------------------
// The fetch routing.
// ---------------------------------------------------------------------------

func TestHeroApplyFetchResultFetchedQueue(t *testing.T) {
	m := NewHeroMachine()
	out := m.ApplyFetchResult(true, 3, 4, false)
	if out.NewState != HeroProcess || out.CaughtUp {
		t.Fatalf("outcome = %+v", out)
	}
	if out.ProgressFormat != "7 ready to process" {
		t.Fatalf("format = %q", out.ProgressFormat)
	}
	if out.CountText != "0 / 7" || out.CountTooltip != "7 fetched item(s) ready to process" {
		t.Fatalf("count = %q / %q", out.CountText, out.CountTooltip)
	}
	if len(out.Lines) != 1 || out.Lines[0].Level != "success" ||
		out.Lines[0].Msg != "🟢 Fetched 3 repo(s) + 4 website link(s) — click PROCESS to start." {
		t.Fatalf("lines = %+v", out.Lines)
	}
	// The one-pipeline shapes (the join law).
	out = m.ApplyFetchResult(true, 0, 2, false)
	if !strings.Contains(out.Lines[0].Msg, "Fetched 2 website link(s)") {
		t.Fatalf("websites-only line = %q", out.Lines[0].Msg)
	}
	out = m.ApplyFetchResult(true, 2, 0, false)
	if !strings.Contains(out.Lines[0].Msg, "Fetched 2 repo(s)") {
		t.Fatalf("repos-only line = %q", out.Lines[0].Msg)
	}
	// The render's count follows the refresh (v0.24.1's both-pipelines
	// law — the label counts repos AND websites).
	if m.Render().Label != "PROCESS (2)" {
		t.Fatalf("label = %q", m.Render().Label)
	}
}

func TestHeroApplyFetchResultImportReady(t *testing.T) {
	m := NewHeroMachine()
	out := m.ApplyFetchResult(true, 0, 0, true)
	if out.NewState != HeroProcess || out.ProgressFormat != "Import file ready" {
		t.Fatalf("outcome = %+v", out)
	}
	if len(out.Lines) != 1 || out.Lines[0].Level != "info" ||
		!strings.Contains(out.Lines[0].Msg, "PROCESS will import the file picked in Settings → Input instead.") {
		t.Fatalf("lines = %+v", out.Lines)
	}
}

func TestHeroApplyFetchResultCaughtUpAndError(t *testing.T) {
	m := NewHeroMachine()
	out := m.ApplyFetchResult(true, 0, 0, false)
	if out.NewState != HeroSync || !out.CaughtUp || out.ProgressFormat != "Ready" {
		t.Fatalf("caught-up outcome = %+v", out)
	}
	if len(out.Lines) != 0 {
		t.Fatalf("the caught-up closing line is the shell's (the waiting report comes first): %+v", out.Lines)
	}
	// v0.63.2's honest closing-line choice.
	if l := CaughtUpLine(0); l.Msg != "✅ All caught up — nothing undone in the bot queue." ||
		l.Level != "success" {
		t.Fatalf("clean closing = %+v", l)
	}
	if l := CaughtUpLine(3); !strings.Contains(l.Msg,
		"the waiting row(s) above are yours (the owner's doors), not the machine's.") {
		t.Fatalf("waiting closing = %+v", l)
	}
	// The waiting-rows report's own grammar.
	if l := WaitingRowsLine(2, 0); !strings.Contains(l.Msg,
		"the master table records 2 ' - ' row(s) whose fetches failed") ||
		strings.Contains(l.Msg, "wait for your eyes") {
		t.Fatalf("waiting line = %q", l.Msg)
	}
	if l := WaitingRowsLine(2, 4); !strings.Contains(l.Msg, "(4 row(s) wait for your eyes)") {
		t.Fatalf("eyes line = %q", l.Msg)
	}
	// The error restore: no lines (the fetch door already spoke), SYNC + Ready.
	out = m.ApplyFetchResult(false, 9, 9, true)
	if out.NewState != HeroSync || out.ProgressFormat != "Ready" || len(out.Lines) != 0 {
		t.Fatalf("error outcome = %+v", out)
	}
}

func TestHeroFetchBailRestoresSync(t *testing.T) {
	m := NewHeroMachine()
	m.BeginFetch()
	m.FetchBailed()
	if m.State() != HeroSync || m.Render().Label != "SYNC" {
		t.Fatalf("post-bail = %q/%s", m.State(), m.Render().Label)
	}
}

// ---------------------------------------------------------------------------
// The mirror.
// ---------------------------------------------------------------------------

func TestHeroMirrorLaws(t *testing.T) {
	m := NewHeroMachine()
	// The missed running transition is caught.
	if !m.Mirror(true) || m.State() != HeroRunning {
		t.Fatalf("the missed transition must flip to running")
	}
	if m.Mirror(true) {
		t.Fatalf("a second running mirror is a no-op")
	}
	// The cleared flag restores SYNC (the batch just ended).
	if !m.Mirror(false) || m.State() != HeroSync {
		t.Fatalf("the cleared flag must restore SYNC")
	}
	// Fetching keeps its rendering while no batch runs (the flag owns
	// the batch, not the fetch).
	m.BeginFetch()
	if m.Mirror(false) {
		t.Fatalf("a cleared flag never touches a fetch")
	}
	if m.State() != HeroFetching {
		t.Fatalf("fetching kept; got %q", m.State())
	}
	// The running flag wins over fetching too (the reference's defensive
	// mirror: `if running:` returns for ANY state — a missed _start_worker
	// transition is caught even mid-fetch).
	if !m.Mirror(true) || m.State() != HeroRunning {
		t.Fatalf("the running flag must win over fetching")
	}
	// A live PROCESS count refreshes in place (a Bot re-check updated
	// the queue lists).
	m2 := NewHeroMachine()
	m2.ApplyFetchResult(true, 2, 0, false)
	m2.SetCounts(2, 5)
	if m2.Render().Label != "PROCESS (7)" {
		t.Fatalf("refreshed label = %q", m2.Render().Label)
	}
	if m2.Mirror(false) {
		t.Fatalf("a process state under a cleared flag stays (the flag belongs to the batch doors)")
	}
}

// ---------------------------------------------------------------------------
// The >10 confirm gate.
// ---------------------------------------------------------------------------

func TestConfirmLargeBatchGate(t *testing.T) {
	if ask, _, _ := ConfirmLargeBatch(10, "bot"); ask {
		t.Fatalf("10 items never asks")
	}
	ask, title, body := ConfirmLargeBatch(11, "bot")
	if !ask || title != "Confirm Large Batch" ||
		body != "This will process 11 items (bot).\n\nContinue?" {
		t.Fatalf("gate = %v %q %q", ask, title, body)
	}
}

// ---------------------------------------------------------------------------
// The sentinel labels' paint laws.
// ---------------------------------------------------------------------------

func TestPaintNetCue(t *testing.T) {
	c := PaintNetCue(true, "https://api.github.com answered HTTP 200 in 41 ms")
	if c.Word != "Net: online" || c.DotColor != "#42B36B" || c.Semantic != "success" ||
		c.Tooltip != "https://api.github.com answered HTTP 200 in 41 ms" {
		t.Fatalf("online cue = %+v", c)
	}
	c = PaintNetCue(false, "")
	if c.Word != "Net: offline" || c.DotColor != "#E85D75" || c.Semantic != "error" ||
		c.Tooltip != "offline" {
		t.Fatalf("offline cue = %+v", c)
	}
}

func TestPaintAICue(t *testing.T) {
	cases := []struct {
		state, kind, engine, short string
		wantWord, wantDot, wantSem string
	}{
		{"on", "cloud", "cloud", "on", "AI: on — cloud", "#42B36B", "success"},
		{"on", "local", "Ollama", "on", "AI: on — local (Ollama)", "#42B36B", "success"},
		{"on", "local", "", "on", "AI: on — local (local)", "#42B36B", "success"},
		{"warn", "local", "Ollama", "no models", "AI: warn — no models", "#D9A036", "warning"},
		{"warn", "local", "Ollama", "", "AI: warn", "#D9A036", "warning"},
		{"warn", "local", "Ollama", "warn — no models", "AI: warn — no models", "#D9A036", "warning"},
		{"off", "local", "Ollama", "unreachable", "AI: off — unreachable", "#E85D75", "error"},
		{"off", "local", "Ollama", "", "AI: off", "#E85D75", "error"},
		{"none", "cloud", "cloud", "not configured", "AI: not configured", "#8E8A90", "muted"},
		{"", "", "", "", "AI: not configured", "#8E8A90", "muted"},
	}
	for _, tc := range cases {
		c := PaintAICue(tc.state, tc.kind, tc.engine, "", tc.short)
		if c.Word != tc.wantWord || c.DotColor != tc.wantDot || c.Semantic != tc.wantSem {
			t.Fatalf("cue(%q,%q,%q,%q) = %+v, want %q/%q/%q",
				tc.state, tc.kind, tc.engine, tc.short, c, tc.wantWord, tc.wantDot, tc.wantSem)
		}
	}
	// The tooltip falls back to the word; the detail wins when present.
	if c := PaintAICue("on", "cloud", "cloud", "", ""); c.Tooltip != "on — cloud" {
		t.Fatalf("empty-detail tooltip = %q", c.Tooltip)
	}
	if c := PaintAICue("on", "cloud", "cloud", "the detail", ""); c.Tooltip != "the detail" {
		t.Fatalf("detail tooltip = %q", c.Tooltip)
	}
}

func TestPaintProxyCue(t *testing.T) {
	c := PaintProxyCue("connected", "Proxy OK (127.0.0.1:10808)")
	if c.Word != "Proxy: connected" || c.DotColor != "#42B36B" || c.Semantic != "success" ||
		c.Tooltip != "Proxy OK (127.0.0.1:10808)" {
		t.Fatalf("connected cue = %+v", c)
	}
	c = PaintProxyCue("disabled", "")
	if c.Word != "Proxy: off" || c.DotColor != "#8E8A90" || c.Semantic != "muted" ||
		c.Tooltip != "Proxy disabled" {
		t.Fatalf("disabled cue = %+v", c)
	}
	c = PaintProxyCue("unreachable", "Proxy unreachable (127.0.0.1:10808) — errno 111")
	if c.Word != "Proxy: unreachable" || c.DotColor != "#E85D75" || c.Semantic != "error" {
		t.Fatalf("unreachable cue = %+v", c)
	}
	// The empty-detail fallbacks.
	if c := PaintProxyCue("connected", ""); c.Tooltip != "Proxy OK" {
		t.Fatalf("connected fallback = %q", c.Tooltip)
	}
	if c := PaintProxyCue("unreachable", ""); c.Tooltip != "Proxy unreachable" {
		t.Fatalf("unreachable fallback = %q", c.Tooltip)
	}
}

// ---------------------------------------------------------------------------
// The progress row's live formats.
// ---------------------------------------------------------------------------

func TestProgressRowFormats(t *testing.T) {
	if FormatReadyToProcess(12) != "12 ready to process" {
		t.Fatalf("ready format = %q", FormatReadyToProcess(12))
	}
	if FormatProcessingOf(7, 30) != "Processing 7 of 30" {
		t.Fatalf("processing format = %q", FormatProcessingOf(7, 30))
	}
	if FormatProcessingOf(0, 0) != "Ready" {
		t.Fatalf("indeterminate processing format = %q", FormatProcessingOf(0, 0))
	}
	if ProgressCountText(10, 20) != "10 / 20" {
		t.Fatalf("count = %q", ProgressCountText(10, 20))
	}
	if ProgressCountText(0, 0) != "0 / 0" {
		t.Fatalf("never a bare dash: %q", ProgressCountText(0, 0))
	}
	// The terminal formats are the exact vocabulary (spec-gui-ui).
	if FormatFetching != "Fetching undone items…" || FormatImportReady != "Import file ready" ||
		FormatReady != "Ready" || FormatProcessing != "Processing…" ||
		FormatDone != "✅ Done" || FormatFailed != "❌ Failed" {
		t.Fatalf("the format constants drifted")
	}
}

// ---------------------------------------------------------------------------
// The sentinel pass's own laws (app-level): the snapshot emission, the
// offline verdict transitions, and the nil-latency regression (the
// offline pass once panicked into the crash-isolation — T-77's fix).
// ---------------------------------------------------------------------------

func TestSentinelPassEmitsSnapshotAndOfflineLaw(t *testing.T) {
	appDir := t.TempDir()
	t.Setenv("GITCURATOR_HOME", appDir)
	a := &App{ConfigPath: filepath.Join(appDir, "config.json"), Bus: NewBus()}
	// The hermetic probe (the seam law): offline → online → offline, on
	// the test's own clock.
	online := false
	a.SentinelProbe = func(cfg any, timeoutS float64) connectioncheck.Snapshot {
		snap := connectioncheck.EmptySnapshot()
		snap.Net.Online = online
		if online {
			ms := 37
			snap.Net.LatencyMs = &ms
			snap.Net.Endpoint = "https://api.github.com"
		}
		return snap
	}

	var mu sync.Mutex
	var snaps []Sentinel
	var lines []LogLine
	a.Bus.Subscribe(&Handlers{
		Sentinel: func(ev Sentinel) {
			mu.Lock()
			snaps = append(snaps, ev)
			mu.Unlock()
		},
		LogLine: func(l LogLine) {
			mu.Lock()
			lines = append(lines, l)
			mu.Unlock()
		},
	})

	// FIRST PASS, OFFLINE: the warn-now law (the one launch warning) +
	// the snapshot crossing + the 🛰️ line with the nil-safe latency
	// (the regression: this pass used to panic on *net.LatencyMs).
	a.sentinelPass()
	mu.Lock()
	n := len(snaps)
	mu.Unlock()
	if n != 1 {
		t.Fatalf("the snapshot never crossed (%d)", n)
	}
	mu.Lock()
	ev := snaps[0]
	all := append([]LogLine(nil), lines...)
	mu.Unlock()
	if ev.Verdict != "warn-now" {
		t.Fatalf("the first offline verdict = %q, want warn-now", ev.Verdict)
	}
	if ev.Net.Online {
		t.Fatalf("the probe should be offline in this sandbox-blank config")
	}
	if ev.Net.LatencyMs != nil {
		t.Fatalf("an offline probe carries no latency")
	}
	var sawWarn, sawLine bool
	for _, l := range all {
		if strings.Contains(l.Msg, "No internet connection detected") {
			sawWarn = true
		}
		if strings.Contains(l.Msg, "🛰️ Sentinel: net=offline latency=n/a") {
			sawLine = true
		}
	}
	if !sawWarn || !sawLine {
		t.Fatalf("the offline pass's lines: warn=%v sentinel=%v (full: %v)", sawWarn, sawLine, all)
	}

	// SECOND PASS, STILL OFFLINE: no second warn (warn-ONCE), the
	// verdict settles to "offline".
	a.sentinelPass()
	mu.Lock()
	ev = snaps[len(snaps)-1]
	warns := 0
	for _, l := range lines {
		if strings.Contains(l.Msg, "No internet connection detected") {
			warns++
		}
	}
	mu.Unlock()
	if ev.Verdict != "offline" || warns != 1 {
		t.Fatalf("the warn-once law: verdict=%q warns=%d", ev.Verdict, warns)
	}

	// THIRD PASS, BACK ONLINE: the success line + the back-online
	// verdict (the labels paint green from the same snapshot).
	// (the seat is false from the second pass — the transition reads it)
	online = true
	a.sentinelPass()
	mu.Lock()
	ev = snaps[len(snaps)-1]
	sawBack := false
	for _, l := range lines {
		if strings.Contains(l.Msg, "✅ Net: back online — syncing is available again.") {
			sawBack = true
		}
	}
	mu.Unlock()
	if ev.Verdict != "back-online" || !ev.Net.Online || !sawBack {
		t.Fatalf("the back-online pass: verdict=%q online=%v line=%v", ev.Verdict, ev.Net.Online, sawBack)
	}

	// FOURTH PASS, FELL OFFLINE: the warning line (never a second modal).
	online = false
	a.sentinelPass()
	mu.Lock()
	ev = snaps[len(snaps)-1]
	sawFell := false
	warns = 0
	for _, l := range lines {
		if strings.Contains(l.Msg, "⚠️ Net: fell offline — syncing will fail until the connection returns") {
			sawFell = true
		}
		if strings.Contains(l.Msg, "No internet connection detected") {
			warns++
		}
	}
	mu.Unlock()
	if ev.Verdict != "fell-offline" || !sawFell || warns != 1 {
		t.Fatalf("the fell-offline pass: verdict=%q line=%v warns=%d", ev.Verdict, sawFell, warns)
	}
}

// TestBatchStopDoorWiring pins the cooperative-stop door's arm/clear
// law: the App's seat arms while an engine runs, Commands.StopBatch
// reaches the armed fn, and the clear only disarms ITS OWN handle.
func TestBatchStopDoorWiring(t *testing.T) {
	a := &App{}
	c := NewCommands(NewBus(), Deps{})
	a.Commands = c
	c.stopDoor = a.requestBatchStop

	// No batch: the honest no-op.
	if c.StopBatch() {
		t.Fatal("no batch, no stop")
	}
	var flipped bool
	h := &batchStopHandle{fn: func() { flipped = true }}
	a.setBatchStop(h)
	if !c.StopBatch() {
		t.Fatal("the armed door must answer true")
	}
	if !flipped {
		t.Fatal("the stop never reached the engine's fn")
	}
	// A late clear from a finished engine never disarms a NEWER one.
	h2 := &batchStopHandle{fn: func() {}}
	a.setBatchStop(h2)
	a.clearBatchStop(h) // the OLD handle's clear
	flipped = false
	if !c.StopBatch() {
		t.Fatal("the newer door must stay armed")
	}
	// The engine's own flag law: requestStop marks stopRequested even
	// before the runner exists (the lost-stop window).
	e := &batchEngine{}
	e.requestStop()
	if !e.stopRequested {
		t.Fatal("requestStop must set the engine's flag")
	}
}
