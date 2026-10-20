package app

// fanfaredoors.go — T-79f: the runner's automatic triggers — the
// Phase-6 closure row (SESSION-046's scoping decision). These are the
// DATA HALVES of the reference's three batch-finish/hero-machine
// surfaces (pinned at 98f4fc7):
//
//   - _maybe_deliver_hand_queue (processing_control.py:475-520) — the
//     end-of-run hand pass: every 🖐 hand row whose page has not
//     landed yet rides the app's own Chrome pass, ONCE per session
//     per link (the _hand_delivery_tried belt — a link whose scrape
//     failed keeps waiting for its doors instead of reopening Chrome
//     every batch); opted out via the web_browser_retry knob.
//   - hero.py's caught-up hand scrape (280-305) — the SAME rows at
//     the caught-up moment, with NO tried-set (the reference's own
//     law at that call site: the caught-up sync opens the tabs itself,
//     every time) and a line of its own.
//   - _maybe_redo_hand_notes (processing_control.py:560-613) — THE
//     NOTE IS THE SUCCESS (v0.57.0): the hand rows whose delivery was
//     a FALSE SUCCESS (the page delivered but the note not properly
//     stored) are redone as a re-run batch; once per session per link
//     (the _hand_redo_tried belt — the LLM's low-confidence answer
//     will not change on an immediate second ask); skipped while a
//     Chrome delivery gathers (those links get their own re-run) and
//     while a batch runs (the running batch's end-of-run pass fires
//     it instead). The 🔁 line speaks here (the reference's own
//     shared-method placement — both call sites hear it).
//   - _maybe_offer_chrome_tab_retry (processing_control.py:1197-1264)
//     — the end-of-run offer: THIS run's fetch-failed links (each
//     with its last error — collect_failed_fetch_links over the
//     worker's _website_summary['results']), minus the fetcher's own
//     retired verdicts (the is_dismissed probe), minus the session's
//     offered set. The OFFER itself (the modal + the answer) is the
//     shell's — the offered-set marking rides the answer door, only
//     the PICKED links ever marked (a declined link was never marked
//     and may be offered again next run — the reference's own law).
//
// NOT new engines: every scan/collector here already landed
// (ScanMasterHandRows / ScanMasterRedoRows / CollectFailedFetchLinks)
// — this file is the follow-up WIRING over them, the Settings-method
// shape of T-79a…T-79e. The SHELL halves (the lines, the
// startChromeTabRetry calls, the re-run batch, the chrome:offer
// crossing, the answer doors) live in webui/heroui.go; the offered
// set + the tried seats are session facts (never persisted).
//
// The recorded divergences (docs/poc-artifacts/t-79f-fanfare-triggers.md):
//   - The reference's _closing/isVisible guards (a hidden PyQt window
//     never offers) have no Go-shell counterpart — the Wails window is
//     the always-visible surface; the same treatment the T-79d/e doors
//     gave them.
//   - The _delivery_rerun_pending flag + the stashed-scorecard story
//     ride the batch-scorecard MODAL (the reference's
//     _stashed_batch_summary), which no T-79 row names — the port's
//     fanfare is the bar marks + the chime, and the re-run batch's
//     own finish tells its own story (T-79d's divergence #2 record,
//     unchanged).

import (
	"fmt"
	"strings"

	"github.com/assadigit/second-brain-project/internal/chrometabs"
	"github.com/assadigit/second-brain-project/internal/dryrun"
	"github.com/assadigit/second-brain-project/internal/githubpipeline"
	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/websitepipeline"
)

// handRowsToLinks maps the scan's rows onto the delivery's link shape
// (the reference's setdefault('error', wall or gesture) law — shared
// by the manual door, the end-of-run pass, and the caught-up scrape).
func handRowsToLinks(rows []websitepipeline.HandRow) []ChromeLinkRow {
	out := make([]ChromeLinkRow, 0, len(rows))
	for _, r := range rows {
		errText := r.Wall
		if errText == "" {
			errText = chromeHandGesture
		}
		out = append(out, ChromeLinkRow{URL: r.URL, Error: errText})
	}
	return out
}

// webBrowserRetryOn reads the fifth door's on/off knob the reference's
// own way (_maybe_offer_chrome_tab_retry:1211-1212: `if cfg.get(
// 'web_browser_retry') is False: return`): a MISSING key (or a null)
// is ON — only an explicitly falsy value opts the AUTOMATIC passes out
// (the manual doors never read this knob).
func webBrowserRetryOn(cfg *storage.OrderedMap) bool {
	if cfg == nil {
		return true
	}
	v, found := cfg.Get("web_browser_retry")
	if !found || v == nil {
		return true // the knob's default is ON
	}
	return pyTruthyCfg(v)
}

// ---------------------------------------------------------------------------
// (a1) The end-of-run hand pass (_maybe_deliver_hand_queue's data half).
// ---------------------------------------------------------------------------

// HandQueuePass is the end-of-run hand pass's data half
// (processing_control.py:486-505): the knob → the pipeline → the vault
// guard ladder (every step a QUIET no-op — the pass never costs the
// owner his scorecard), then the master table's 🖐 rows minus the
// session's tried set. The tried set is updated ONLY here — a link
// whose Chrome scrape failed keeps waiting for its doors (the manual
// More door) instead of reopening Chrome on every batch.
func (s *Settings) HandQueuePass() []ChromeLinkRow {
	cfg := s.cfg
	if cfg == nil {
		return nil
	}
	if !webBrowserRetryOn(cfg) {
		return nil // the knob opts the automatic pass out
	}
	if !pyTruthyCfg(getOrCfg(getMapCfg(cfg, "pipelines"), "websites", false)) {
		return nil
	}
	vault := strings.TrimSpace(pyStrCfg(cfg, "website_vault_path"))
	if vault == "" || !fsIsDirPath(vault) {
		return nil
	}
	rows := websitepipeline.ScanMasterHandRows(vault, nil)
	if len(rows) == 0 {
		return nil
	}
	links := handRowsToLinks(rows)
	s.mu.Lock()
	if s.handDeliveryTried == nil {
		s.handDeliveryTried = map[string]bool{}
	}
	kept := make([]ChromeLinkRow, 0, len(links))
	for _, l := range links {
		if s.handDeliveryTried[l.URL] {
			continue // once per session per link — the tried belt
		}
		s.handDeliveryTried[l.URL] = true
		kept = append(kept, l)
	}
	s.mu.Unlock()
	if len(kept) == 0 {
		return nil // the quiet no-op is nil, never an empty slice
	}
	return kept
}

// ---------------------------------------------------------------------------
// (a2) hero.py's caught-up hand scrape (280-281) — the same rows, the
// caught-up moment's own law.
// ---------------------------------------------------------------------------

// CaughtUpHandRows is the caught-up sync's hand scrape data half
// (hero.py:280-281 over _scan_master_hand's own guards: no vault, no
// pipeline, or a running batch means no scan). NO tried-set — the
// reference's caught-up call site filters nothing: the caught-up sync
// opens the tabs itself every time (a hand row is a request for the
// app's own Chrome pass, not a note to self).
func (s *Settings) CaughtUpHandRows(batchBusy bool) []ChromeLinkRow {
	if batchBusy {
		return nil // a running batch owns the table's truth
	}
	cfg := s.cfg
	if cfg == nil {
		return nil
	}
	if !pyTruthyCfg(getOrCfg(getMapCfg(cfg, "pipelines"), "websites", false)) {
		return nil
	}
	vault := strings.TrimSpace(pyStrCfg(cfg, "website_vault_path"))
	if vault == "" || !fsIsDirPath(vault) {
		return nil
	}
	return handRowsToLinks(websitepipeline.ScanMasterHandRows(vault, nil))
}

// ---------------------------------------------------------------------------
// (b) The redo pass (_maybe_redo_hand_notes over _scan_master_redo).
// ---------------------------------------------------------------------------

// RedoHandNotesScan is the redo pass's data half
// (processing_control.py:573-604): the quiet guards (pipeline, vault,
// a gathering Chrome delivery, a running batch), then the false-
// success rows — the delivered hand links whose notes are NOT properly
// stored — minus the session's tried set (once per session per link).
// The 🔁 line speaks HERE (the reference's own shared-method
// placement — the chain and the caught-up path both hear it), with
// the first row's reason riding along exactly as the reference
// formats it.
func (s *Settings) RedoHandNotesScan(chromeBusy, batchBusy bool) []websitepipeline.RedoRow {
	cfg := s.cfg
	if cfg == nil {
		return nil
	}
	if !pyTruthyCfg(getOrCfg(getMapCfg(cfg, "pipelines"), "websites", false)) {
		return nil
	}
	vault := strings.TrimSpace(pyStrCfg(cfg, "website_vault_path"))
	if vault == "" || !fsIsDirPath(vault) {
		return nil
	}
	if chromeBusy {
		return nil // the delivery's own re-run owns those links now
	}
	if batchBusy {
		return nil // the running batch's end-of-run pass will fire it
	}
	var rows []websitepipeline.RedoRow
	func() {
		defer func() { recover() }() // a broken scan never breaks the pass
		var state *websitepipeline.WebsiteStateDB
		if dryrun.IsEnabled() {
			state, _ = websitepipeline.NewWebsiteStateDB(
				dryrun.ShadowCachePath(cacheDBPathOf(cfg)))
		} else {
			state, _ = websitepipeline.NewWebsiteStateDB(cacheDBPathOf(cfg))
		}
		if state == nil {
			return // nothing can be proven false — the hermetic law
		}
		defer state.Close()
		rows = websitepipeline.ScanMasterRedoRows(vault, state,
			func(msg, level string) { s.SpeakOnBus(msg, level) })
	}()
	if len(rows) == 0 {
		return nil
	}
	s.mu.Lock()
	if s.handRedoTried == nil {
		s.handRedoTried = map[string]bool{}
	}
	kept := make([]websitepipeline.RedoRow, 0, len(rows))
	for _, r := range rows {
		if s.handRedoTried[r.URL] {
			continue // the LLM's answer will not change on an immediate second ask
		}
		s.handRedoTried[r.URL] = true
		kept = append(kept, r)
	}
	s.mu.Unlock()
	if len(kept) == 0 {
		return nil
	}
	firstReason := strings.TrimSpace(kept[0].Reason)
	line := fmt.Sprintf(
		"🔁 %d hand note(s) NOT properly stored — "+
			"refetching and re-generating them now (the delivered "+
			"page is re-read, the LLM writes the proper, categorized "+
			"note; a ✅ is only earned by the note itself",
		len(kept))
	if firstReason != "" {
		line += "; first reason: " + firstReason
	}
	line += ")"
	s.SpeakOnBus(line, "info")
	return kept
}

// ---------------------------------------------------------------------------
// (c) The end-of-run offer (_maybe_offer_chrome_tab_retry's data half).
// ---------------------------------------------------------------------------

// ChromeOfferFromResults is the end-of-run offer's data half
// (processing_control.py:1210-1253): the quiet guard ladder (the
// knob, the pipeline, the vault), then THIS run's fetch-failed links —
// collect_failed_fetch_links over the run's own results — minus the
// fetcher's retired verdicts (the is_dismissed probe; a state question
// that cannot be asked offers them anyway) minus the session's offered
// set (a link the owner already picked is never re-offered; a DECLINED
// link was never marked and may be offered again next run — the
// reference's own law).
func (s *Settings) ChromeOfferFromResults(
	entries []githubpipeline.WebsiteResultEntry) []ChromeLinkRow {
	cfg := s.cfg
	if cfg == nil {
		return nil
	}
	if !webBrowserRetryOn(cfg) {
		return nil // opted out (the config knob, default ON)
	}
	if !pyTruthyCfg(getOrCfg(getMapCfg(cfg, "pipelines"), "websites", false)) {
		return nil
	}
	vault := strings.TrimSpace(pyStrCfg(cfg, "website_vault_path"))
	if vault == "" || !fsIsDirPath(vault) {
		return nil
	}
	// The run's results as the collector's own shape.
	results := make([]websitepipeline.Result, 0, len(entries))
	for _, e := range entries {
		results = append(results, websitepipeline.Result{
			URL: e.URL, Canonical: e.Canonical, Outcome: e.Outcome,
			NotePath: e.NotePath, Category: e.Category,
			Subcategory: e.Subcategory, FetchStatus: e.FetchStatus,
			Error: e.Error,
		})
	}
	// The dismissed probe: a broken state question never hides a link
	// (the reference's `except Exception: pass # offer them`).
	var links []chrometabs.Link
	func() {
		defer func() { recover() }() // tolerated everywhere — an offer
		// that cannot be made never costs the owner his scorecard
		var state *websitepipeline.WebsiteStateDB
		if dryrun.IsEnabled() {
			state, _ = websitepipeline.NewWebsiteStateDB(
				dryrun.ShadowCachePath(cacheDBPathOf(cfg)))
		} else {
			state, _ = websitepipeline.NewWebsiteStateDB(cacheDBPathOf(cfg))
		}
		if state != nil {
			defer state.Close()
			links = chrometabs.CollectFailedFetchLinks(results, func(canonical string) bool {
				dismissed, err := state.IsDismissed(canonical)
				if err != nil {
					panic(err) // caught right here — the degraded probe offers
				}
				return dismissed
			})
			return
		}
		links = chrometabs.CollectFailedFetchLinks(results, nil)
	}()
	if len(links) == 0 {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]ChromeLinkRow, 0, len(links))
	for _, l := range links {
		if s.chromeTabOffered[l.URL] {
			continue // the session's offered set
		}
		out = append(out, ChromeLinkRow{URL: l.URL, Error: l.Error})
	}
	if len(out) == 0 {
		return nil // the quiet no-op is nil, never an empty slice
	}
	return out
}
