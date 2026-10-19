package app

// herostate.go — T-77: the main window's hero button machine as a PURE
// engine (gui/main_window/hero.py), main-battery-testable per the house
// law. The machine owns EXACTLY what hero.py's _hero_state owns: the
// GUI-only state word, the stopping flag, the both-pipelines pending
// counts (v0.24.1), and the sentinel's observed-online seat. It never
// touches the pipeline — the running flag stays with the batch doors
// (SyncRunBatch), the fetch stays with SyncFetchQueue; the shell drives
// the machine and renders its verdicts.
//
// THE STATE MACHINE (hero.py's docstring law): SYNC when idle →
// PROCESS once undone items have been fetched → STOP while a batch runs
// (the SAME button, danger fill + square glyph) → back to SYNC when it
// finishes or is cancelled. FETCHING is the disabled in-flight fetch
// rendering. No separate STOP button exists (one primary action).
//
// THE CHROME'S OTHER PURE LAWS live here too:
//   - the progress row's live formats (spec-gui-ui's SegmentBar
//     setFormat strings — the exact vocabulary the tests pin);
//   - the sentinel labels' paint laws (window.py's
//     _apply_net/ai/proxy_status — words + dots + semantics + tooltips,
//     never bare dashes);
//   - the >10 confirm gate (_confirm_batch, v31.1);
//   - the caught-up closing line (v0.63.2's honest law).

import (
	"fmt"
	"strings"

	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/workerclient"
)

// The hero states (hero.py's _hero_state vocabulary).
const (
	HeroSync     = "sync"
	HeroFetching = "fetching"
	HeroProcess  = "process"
	HeroRunning  = "running"
)

// The click actions — what the SHELL must do for one click.
const (
	// HeroActNone: nothing (a disabled button, a winding-down stop, or a
	// guard box already spoke).
	HeroActNone = ""
	// HeroActStop: cancel the running batch (the cooperative stop).
	HeroActStop = "stop"
	// HeroActProcess: run the fetched queue (process_bot_queue's twin).
	HeroActProcess = "process"
	// HeroActImport: nothing was fetched — run the Input mode instead
	// (start_processing, v0.23.0's ONE remaining input mode).
	HeroActImport = "import"
	// HeroActFetch: stage 1 — fetch every UNDONE item (the queue fetch).
	HeroActFetch = "fetch"
)

// The progress row's live formats (spec-gui-ui's SegmentBar law — the
// exact strings; the tests pin them).
const (
	FormatFetching    = "Fetching undone items…"
	FormatImportReady = "Import file ready"
	FormatReady       = "Ready"
	FormatProcessing  = "Processing…"
	FormatDone        = "✅ Done"
	FormatFailed      = "❌ Failed"
)

// FormatReadyToProcess is "{n} ready to process" (the fetched queue's
// own format — v0.07: the counter reflects the fetched queue
// immediately).
func FormatReadyToProcess(n int) string {
	return fmt.Sprintf("%d ready to process", n)
}

// FormatProcessingOf is update_progress's bar format (v31.1:
// "Processing X of Y").
func FormatProcessingOf(current, total int) string {
	if total > 0 {
		return fmt.Sprintf("Processing %d of %d", current, total)
	}
	return FormatReady
}

// ProgressCountText is the standalone X / Y counter beside the bar
// (v33 wireframe "10/20"; v0.07: never "– / –" — a determinate bar
// always has numbers).
func ProgressCountText(current, total int) string {
	if total > 0 {
		return fmt.Sprintf("%d / %d", current, total)
	}
	return "0 / 0"
}

// ---------------------------------------------------------------------------
// The hero machine.
// ---------------------------------------------------------------------------

// HeroBox is one guard's themed message box (the reference's
// _show_custom_message_box / _show_custom_question payload).
type HeroBox struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	// Question marks a yes/no box (the offline guard's Continue anyway?).
	Question bool `json:"question,omitempty"`
}

// HeroClickPlan is Click's verdict: the action to run plus the guard
// boxes the shell must render first. A plan with an Action and no boxes
// runs immediately; a box without an action is the whole answer.
type HeroClickPlan struct {
	Action string `json:"action"`
	// VaultBox: the invalid-vault error box (hero.py:130-132).
	VaultBox *HeroBox `json:"vault_box,omitempty"`
	// OfflineAsk: the sentinel has SEEN the internet unreachable — the
	// continue-anyway question comes BEFORE the fetch burns its timeout
	// (v0.85.0's advisory guard; always escapable).
	OfflineAsk bool `json:"offline_ask,omitempty"`
	// NothingBox: no queue configured and no import file picked
	// (hero.py:164-171, re-worded for Option B's Worker queue per
	// ADR-0004 Am.3 — the Telegram credentials sentence goes).
	NothingBox *HeroBox `json:"nothing_box,omitempty"`
}

// HeroLine is one desk line a routing speaks (msg + level).
type HeroLine struct {
	Msg   string `json:"msg"`
	Level string `json:"level"`
}

// HeroFetchOutcome is ApplyFetchResult's verdict: the new state, the
// progress row's live format + count, and the desk lines the routing
// speaks.
type HeroFetchOutcome struct {
	NewState       string     `json:"new_state"`
	ProgressFormat string     `json:"progress_format"`
	CountText      string     `json:"count_text"`
	CountTooltip   string     `json:"count_tooltip"`
	Lines          []HeroLine `json:"lines"`
	// CaughtUp marks the ok-but-nothing-found path: the shell speaks the
	// master-table waiting report + the drift seal + the closing line
	// (v0.89.0's DRIFT SEAL rides here).
	CaughtUp bool `json:"caught_up"`
}

// HeroMachine is hero.py's GUI-only state (pure; the shell owns the
// I/O). Construct with NewHeroMachine.
type HeroMachine struct {
	state string

	// stopping: stop_processing's winding-down render (v0.32 — the
	// button waits DISABLED "Stopping…" until the batch's finish
	// restores SYNC; _batch_running stays true while it winds down).
	stopping bool

	// The both-pipelines pending counts (v0.24.1: pending repos AND
	// pending website links — a GitHub-only count said PROCESS/0 while
	// website links waited in the queue).
	pendingRepos    int
	pendingWebsites int

	// sentinelOnline: the sentinel's observed verdict (nil = not
	// observed yet → the offline guard stays silent, hero.py:140).
	sentinelOnline *bool
}

// NewHeroMachine builds the machine in the born state (SYNC — hero.py's
// getattr default).
func NewHeroMachine() *HeroMachine {
	return &HeroMachine{state: HeroSync}
}

// State answers the current state word.
func (m *HeroMachine) State() string { return m.state }

// Pending answers the both-pipelines count (v0.24.1).
func (m *HeroMachine) Pending() int { return m.pendingRepos + m.pendingWebsites }

// SetSentinelOnline records the sentinel's verdict (nil = not observed;
// a pointer at false is the advisory offline guard's trigger).
func (m *HeroMachine) SetSentinelOnline(online *bool) { m.sentinelOnline = online }

// SetCounts refreshes the pending counts in place (a Settings → Bot
// re-check updates the queue lists under a live PROCESS state —
// hero.py:92-98's fresh-count law).
func (m *HeroMachine) SetCounts(repos, websites int) {
	m.pendingRepos = repos
	m.pendingWebsites = websites
}

// Click is _on_hero_clicked: the routing for ONE click, guards first.
// vaultValid: the configured vault exists; queueConfigured: the Worker
// queue is enabled (Option B's twin of the bot-credentials guard);
// importReady: the Input page's import file field is non-empty.
func (m *HeroMachine) Click(vaultValid, queueConfigured, importReady bool) HeroClickPlan {
	switch m.state {
	case HeroRunning:
		if m.stopping {
			return HeroClickPlan{Action: HeroActNone} // already winding down
		}
		return HeroClickPlan{Action: HeroActStop}
	case HeroProcess:
		if m.Pending() > 0 {
			return HeroClickPlan{Action: HeroActProcess}
		}
		return HeroClickPlan{Action: HeroActImport}
	case HeroSync:
		return m.clickSync(vaultValid, queueConfigured, importReady)
	default: // fetching — the button is disabled; a fossil click does nothing
		return HeroClickPlan{Action: HeroActNone}
	}
}

// clickSync runs the SYNC click through the reference's guard order:
// the vault, then the offline advisory, then the queue/import split.
func (m *HeroMachine) clickSync(vaultValid, queueConfigured, importReady bool) HeroClickPlan {
	if !vaultValid {
		return HeroClickPlan{VaultBox: &HeroBox{
			Title: "Error",
			Body:  "Please select a valid Obsidian vault path.",
		}}
	}
	if m.sentinelOnline != nil && !*m.sentinelOnline {
		return HeroClickPlan{OfflineAsk: true}
	}
	return m.afterOfflineGuards(queueConfigured, importReady)
}

// AnswerOffline answers the continue-anyway question. A "no" stays in
// SYNC (retry when back online); a "yes" returns the warning line the
// shell speaks (the owner's word wins) plus the continued plan.
func (m *HeroMachine) AnswerOffline(goSync bool,
	queueConfigured, importReady bool) (HeroClickPlan, HeroLine) {
	if !goSync {
		return HeroClickPlan{Action: HeroActNone}, HeroLine{}
	}
	return m.afterOfflineGuards(queueConfigured, importReady), HeroLine{
		Msg:   "⚠️ Syncing despite the offline verdict (the owner's word wins).",
		Level: "warning",
	}
}

// afterOfflineGuards is the queue/import split (hero.py:156-172,
// re-scoped for Option B: the Worker queue replaces the Telegram
// credentials; the import file remains the ONE other input mode).
func (m *HeroMachine) afterOfflineGuards(queueConfigured, importReady bool) HeroClickPlan {
	if queueConfigured {
		return HeroClickPlan{Action: HeroActFetch}
	}
	if importReady {
		return HeroClickPlan{Action: HeroActImport}
	}
	return HeroClickPlan{NothingBox: &HeroBox{
		Title: "Nothing to sync",
		Body: "SYNC fetches undone items from your sync queue (the " +
			"Cloudflare Worker).\n\nSet the Worker's URL (or pair the app) in " +
			"Settings → Bot.\n\nTo import from a file instead, pick it in " +
			"Settings → Input.",
	}}
}

// BeginFetch flips to FETCHING (stage 1 started — the button renders
// disabled until the fetch's verdict lands).
func (m *HeroMachine) BeginFetch() { m.state = HeroFetching }

// FetchBailed restores SYNC after an early bail (the busy-refusal path
// — hero.py:182-188: the reason was already logged, the button returns).
func (m *HeroMachine) FetchBailed() { m.state = HeroSync }

// ApplyFetchResult is _after_sync_fetch: the fetch worker finished and
// the queue counts are in. ok mirrors the fetch's own success verdict
// (State == "ok"); repos/websites are the pending lists' lengths;
// importReady re-reads the Input page's field (the reference re-reads
// the widget at verdict time).
func (m *HeroMachine) ApplyFetchResult(ok bool, repos, websites int,
	importReady bool) HeroFetchOutcome {
	m.pendingRepos = repos
	m.pendingWebsites = websites
	n := repos + websites

	if ok && n > 0 {
		m.state = HeroProcess
		parts := []string{}
		if repos > 0 {
			parts = append(parts, fmt.Sprintf("%d repo(s)", repos))
		}
		if websites > 0 {
			parts = append(parts, fmt.Sprintf("%d website link(s)", websites))
		}
		return HeroFetchOutcome{
			NewState:       HeroProcess,
			ProgressFormat: FormatReadyToProcess(n),
			CountText:      fmt.Sprintf("0 / %d", n),
			CountTooltip:   fmt.Sprintf("%d fetched item(s) ready to process", n),
			Lines: []HeroLine{{
				Msg:   "🟢 Fetched " + strings.Join(parts, " + ") + " — click PROCESS to start.",
				Level: "success",
			}},
		}
	}
	if ok && importReady {
		// Queue all caught up → PROCESS will run the import file.
		m.state = HeroProcess
		return HeroFetchOutcome{
			NewState:       HeroProcess,
			ProgressFormat: FormatImportReady,
			Lines: []HeroLine{{
				Msg:   "✅ Bot queue is all caught up — PROCESS will import the file picked in Settings → Input instead.",
				Level: "info",
			}},
		}
	}
	if ok {
		// The caught-up path: the closing line is the shell's (the
		// waiting-rows report comes first, CaughtUpLine picks the words).
		m.state = HeroSync
		return HeroFetchOutcome{
			NewState:       HeroSync,
			ProgressFormat: FormatReady,
			CaughtUp:       true,
		}
	}
	// The fetch error was already logged by the fetch door.
	m.state = HeroSync
	return HeroFetchOutcome{NewState: HeroSync, ProgressFormat: FormatReady}
}

// CaughtUpLine is the caught-up path's closing line (v0.63.2's honest
// law: with waiting rows reported above, "caught up" is said with the
// rows named — the machine's work is DONE either way, the rest are the
// owner's doors).
func CaughtUpLine(waiting int) HeroLine {
	if waiting > 0 {
		return HeroLine{
			Msg:   "✅ Bot queue caught up — no new websites to fetch; the waiting row(s) above are yours (the owner's doors), not the machine's.",
			Level: "success",
		}
	}
	return HeroLine{
		Msg:   "✅ All caught up — nothing undone in the bot queue.",
		Level: "success",
	}
}

// WaitingRowsLine is the master-table waiting report (hero.py:262-271,
// v0.51.0's check half — the rows are read, never auto-fetched; the
// owner's doors are named). eyes is the _master_waiting_eyes count (0
// hides the parenthetical).
func WaitingRowsLine(waiting, eyes int) HeroLine {
	msg := fmt.Sprintf(
		"📋 Bot queue caught up; the master table records %d ' - ' row(s) whose fetches failed — "+
			"NOT re-fetched on my own (only newly-added websites are fetched). Your doors: set ✅ or 🪦 in the table, "+
			"queue 🖐, ♻️ revive, or More ▸ 🔁 Retry the table's ' - ' rows to fetch them now", waiting)
	if eyes > 0 {
		msg += fmt.Sprintf(" (%d row(s) wait for your eyes)", eyes)
	}
	msg += "."
	return HeroLine{Msg: msg, Level: "info"}
}

// Mirror is _sync_run_button: the pipeline-state mirror onto the
// button. The running flag is the BATCH's own (BatchBusy); the mirror
// only catches missed transitions — running flips the state to running
// (a missed _start_worker rendering), a cleared flag under running
// restores SYNC. Returns whether the state changed (a re-render is the
// shell's call either way — the reference's timer re-renders on diff).
func (m *HeroMachine) Mirror(batchRunning bool) bool {
	if batchRunning {
		if m.state != HeroRunning {
			m.state = HeroRunning
			m.stopping = false
			return true
		}
		return false
	}
	if m.state == HeroRunning {
		m.state = HeroSync
		m.stopping = false
		return true
	}
	return false
}

// BeginStopping is stop_processing's render state: the button waits
// DISABLED ("Stopping…") until the batch's finish restores SYNC — no
// double-click (v0.32). The state stays running while it winds down.
func (m *HeroMachine) BeginStopping() { m.stopping = true }

// Stopping answers whether a stop is winding down.
func (m *HeroMachine) Stopping() bool { return m.stopping }

// HeroRender is _set_hero_state's GUI verdict: the label, the glyph,
// the enabled flag, the fill (danger only for Stop), and the tooltip —
// each state's exact words.
type HeroRender struct {
	Label   string `json:"label"`
	Glyph   string `json:"glyph"` // refresh | loader | play | stop
	Enabled bool   `json:"enabled"`
	Danger  bool   `json:"danger"`
	Tooltip string `json:"tooltip"`
}

// The reference's own tooltip sentences (hero.py:439-486), verbatim.
const (
	heroStopTooltip = "Cancel the running batch (SYNC returns when it stops).\n" +
		"The stop is clean — already-processed notes are kept;\n" +
		"unfinished links are surfaced for retry when it ends."
	heroFetchingTooltip = "Fetching undone items from the Telegram bot…"
	heroSyncTooltip     = "FULL SYNC — Stage 1: fetch every UNDONE item from\n" +
		"the Telegram bot (repos already in the vault,\n" +
		"settled (addressed) and decommissioned ones are\n" +
		"skipped; website links already in the Websites\n" +
		"vault or settled are skipped too), then the batch\n" +
		"curates them with the LLM.\n" +
		"The button then becomes PROCESS — click it to start the batch.\n" +
		"While a batch runs this button becomes Stop — click to cancel.\n" +
		"(The no-AI mirror of the triad is the ⚡ Fast Sync\n" +
		"button — vault → GitHub → the website, no fetch,\n" +
		"no LLM.)"
)

// Render is _set_hero_state: the state's exact label + glyph + enabled
// + fill + tooltip. The shell renders this on every crossing.
func (m *HeroMachine) Render() HeroRender {
	switch m.state {
	case HeroRunning:
		if m.stopping {
			return HeroRender{Label: "Stopping…", Glyph: "stop", Enabled: false,
				Danger: true, Tooltip: heroStopTooltip}
		}
		return HeroRender{Label: "Stop", Glyph: "stop", Enabled: true,
			Danger: true, Tooltip: heroStopTooltip}
	case HeroFetching:
		return HeroRender{Label: "FETCHING…", Glyph: "loader", Enabled: false,
			Tooltip: heroFetchingTooltip}
	case HeroProcess:
		n := m.Pending()
		label := "PROCESS"
		tip := "Nothing was fetched — clicking runs the selected Input mode\n" +
			"(Settings → Input) instead."
		if n > 0 {
			label = fmt.Sprintf("PROCESS (%d)", n)
			tip = fmt.Sprintf("Stage 2 — start curating the %d fetched undone item(s) "+
				"into the Obsidian vault(s).\nA confirmation appears for large "+
				"batches (more than 10 items).", n)
		}
		return HeroRender{Label: label, Glyph: "play", Enabled: true, Tooltip: tip}
	default: // sync — also restores after running/fetching
		return HeroRender{Label: "SYNC", Glyph: "refresh", Enabled: true,
			Tooltip: heroSyncTooltip}
	}
}

// ConfirmLargeBatch is _confirm_batch (v31.1's safety gate): a batch on
// MORE THAN 10 items states the exact count and asks. ask false = no
// gate.
func ConfirmLargeBatch(count int, source string) (ask bool, title, body string) {
	if count <= 10 {
		return false, "", ""
	}
	return true, "Confirm Large Batch",
		fmt.Sprintf("This will process %d items (%s).\n\nContinue?", count, source)
}

// ---------------------------------------------------------------------------
// The sentinel labels' paint laws (window.py:662-748 — the words + dots
// + semantics, never bare dashes; WCAG 1.4.1: color never carries the
// state alone).
// ---------------------------------------------------------------------------

// CueRender is one sentinel label's paint verdict.
type CueRender struct {
	// Word is the visible label ("Net: online", "AI: on — local
	// (Ollama)", "Proxy: connected"…).
	Word string `json:"word"`
	// DotColor is the 12px dot's fill (#42B36B / #D9A036 / #E85D75 /
	// #8E8A90).
	DotColor string `json:"dot_color"`
	// Semantic is the text's tone class (success/warning/error/muted).
	Semantic string `json:"semantic"`
	// Tooltip is the hover story (the detail, or the word when empty).
	Tooltip string `json:"tooltip"`
}

func orWord(s, word string) string {
	if strings.TrimSpace(s) != "" {
		return s
	}
	return word
}

// PaintNetCue is _apply_net_status (window.py:662-681): online → green
// "Net: online"; offline → red "Net: offline".
func PaintNetCue(online bool, detail string) CueRender {
	word, dot, sem := "offline", "#E85D75", "error"
	if online {
		word, dot, sem = "online", "#42B36B", "success"
	}
	return CueRender{
		Word:     "Net: " + word,
		DotColor: dot,
		Semantic: sem,
		Tooltip:  orWord(detail, word),
	}
}

// PaintAICue is _apply_ai_status (window.py:683-721): the owner's ask,
// verbatim — "i want to be able to SEE AI : on (whether cloud or
// local)". on → "AI: on — cloud" / "AI: on — local (engine)"; warn →
// amber + the short reason; off → red; none → muted "not configured".
// Every state spells a REAL word (the v0.34 law).
func PaintAICue(state, kind, engine, detail, short string) CueRender {
	var word, dot, sem string
	switch state {
	case "on":
		if kind == "cloud" {
			word = "on — cloud"
		} else {
			e := engine
			if e == "" {
				e = "local"
			}
			word = "on — local (" + e + ")"
		}
		dot, sem = "#42B36B", "success"
	case "warn":
		word = short
		if word == "" {
			word = "warn"
		}
		if !strings.HasPrefix(word, "warn") {
			word = "warn — " + word
		}
		dot, sem = "#D9A036", "warning"
	case "off":
		word = short
		if word == "" {
			word = "off"
		}
		if !strings.HasPrefix(word, "off") {
			word = "off — " + word
		}
		dot, sem = "#E85D75", "error"
	default: // 'none' / unknown → not configured
		word, dot, sem = "not configured", "#8E8A90", "muted"
	}
	return CueRender{
		Word:     "AI: " + word,
		DotColor: dot,
		Semantic: sem,
		Tooltip:  orWord(detail, word),
	}
}

// PaintProxyCue is _apply_proxy_status + _set_proxy_status_text
// (window.py:723-748 + input_proxy.py:119-152): the label keeps the
// v31.1 vocabulary — connected → green "Proxy: connected"; disabled →
// muted "Proxy: off"; unreachable → red "Proxy: unreachable".
func PaintProxyCue(state, detail string) CueRender {
	switch state {
	case "connected":
		return CueRender{
			Word:     "Proxy: connected",
			DotColor: "#42B36B",
			Semantic: "success",
			Tooltip:  orWord(detail, "Proxy OK"),
		}
	case "disabled":
		return CueRender{
			Word:     "Proxy: off",
			DotColor: "#8E8A90",
			Semantic: "muted",
			Tooltip:  "Proxy disabled",
		}
	default: // unreachable / unknown
		return CueRender{
			Word:     "Proxy: unreachable",
			DotColor: "#E85D75",
			Semantic: "error",
			Tooltip:  orWord(detail, "Proxy unreachable"),
		}
	}
}

// ---------------------------------------------------------------------------
// The chrome's Settings doors (T-77's thin additions to the surface).
// ---------------------------------------------------------------------------

// WorkerQueueEnabled answers the hero's queue-configured guard (Option
// B's twin of the reference's bot-credentials check): the Worker queue
// is wired — cloudflare_enabled + the URL + the pairing.
func (s *Settings) WorkerQueueEnabled() bool {
	return workerclient.NewSync(cfgPlainMap(s.PlainConfig())).IsEnabled()
}

// SaveDarkMode commits the More ▸ Settings ▸ Dark Mode toggle (one key
// = one merge; the theme flip itself is the frontend's).
func (s *Settings) SaveDarkMode(enabled bool) *SaveResult {
	ui := storage.NewOrderedMap()
	ui.Set("dark_mode", enabled)
	return s.saveMerge(ui)
}
