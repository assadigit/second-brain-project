package cli

// hands.go — the --hand-delivery (cli.py cmd_hand_delivery, 1032-1088)
// and --chrome-retry (cmd_chrome_retry, 1095-1245) wirings over the
// PORTED engines (internal/handdelivery, internal/chrometabs, the
// websitepipeline legs).
//
// --hand-delivery is COMPLETE: the walled-rows listing
// (handdelivery.WalledRetryRows over the StateDB adapter — the
// decomtable law keeps the concrete adapter at the call site), the
// enqueue (queue.json + README.txt, the open-Chrome leg), the folder
// card.
//
// --chrome-retry wires the LISTING (the retry rows + the 🖐 hand rows
// + the redo set) and the FULL dry-run rehearsal; the live run
// (preflight → the real Chrome delivery → run_batch_visual processing
// the delivered pages IN THE SAME RUN) waits for T-62's batch engine —
// an honest boundary card, never a silent no-op: splitting the deliver
// from the process would diverge from the reference's one-run
// contract.
//
// REALITY-WINS FIND (recorded in spec-cli.md): the reference's
// cmd_chrome_retry --dry-run leg calls deliver_pages_via_chrome
// WITHOUT flipping the global dry-run switch (only the batch worker
// ever flips it) — so the pinned rev's dry-run would launch the real
// Chrome while the card says "Chrome is never launched". The port
// flips the switch for the dry-run leg (the card's own words + the
// deliver's rehearsal law are the contract).

import (
	"fmt"
	"strings"

	"github.com/assadigit/second-brain-project/internal/chrometabs"
	"github.com/assadigit/second-brain-project/internal/dryrun"
	"github.com/assadigit/second-brain-project/internal/handdelivery"
	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/websitepipeline"
)

// stateRowsAdapter adapts websitepipeline.WebsiteStateDB onto
// handdelivery.RetryRowsState (the reference duck-types the pipeline's
// state DB into the picker; the concrete adapter lives here — the CLI
// hand leg — per the decomtable law).
type stateRowsAdapter struct {
	db *websitepipeline.WebsiteStateDB
}

// AllRetryRows mirrors all_retry_rows(): every waiting retry-queue
// row, oldest first, in the picker's row shape.
func (a *stateRowsAdapter) AllRetryRows() ([]handdelivery.RetrySourceRow, error) {
	rows, err := a.db.AllRetryRows()
	if err != nil {
		return nil, err
	}
	out := make([]handdelivery.RetrySourceRow, 0, len(rows))
	for _, r := range rows {
		le := ""
		if r.LastError != nil {
			le = *r.LastError
		}
		out = append(out, handdelivery.RetrySourceRow{
			URL: r.URL, LastError: le, Attempts: int(r.Attempts),
		})
	}
	return out, nil
}

// cmdHandDelivery mirrors cmd_hand_delivery: the fourth door on the
// CLI — list every retry-queued link whose last failure was a WALL,
// queue them all for hand-delivery, and open each in the owner's real
// Chrome. Nothing is retired; the links keep waiting.
func (c *CLI) cmdHandDelivery(a *Args) int {
	c.Banner()
	path := configPath(a.Config)
	cfg, ok := c.loadConfig(path)
	if !ok {
		cfg = storage.NewOrderedMap()
	}
	vault := strings.TrimSpace(cfgOrStr(cfg, "website_vault_path"))
	if vault == "" {
		c.CliPrint("website_vault_path is not set — run --cli --init first.", "warning")
		return 1
	}
	dbFile := cacheDBPath()
	if !fileExists(dbFile) {
		c.CliPrint("No cache.db yet — nothing has failed (nothing is walled).", "success")
		return 0
	}
	state, err := websitepipeline.NewWebsiteStateDB(dbFile)
	if err != nil {
		// the reference crashes here (no guard around the state open);
		// the port answers the honest error card instead — a Go note.
		c.CliPrint(fmt.Sprintf("Walled links unavailable (%s)", err), "error")
		return 1
	}
	rows := handdelivery.WalledRetryRows(&stateRowsAdapter{state})
	state.Close()

	fmt.Fprint(c.Out, Paint("Walled links (the fourth door — your real Chrome)", cBold)+"\n")
	fmt.Fprint(c.Out, Rule()+"\n")
	if len(rows) == 0 {
		fmt.Fprint(c.Out, Paint("  ✓ empty — no retry-queued link ends in a wall.", cGreen)+"\n")
		return 0
	}
	for _, r := range rows {
		wall := pySliceRunes(r.Wall, 0, 100)
		line := fmt.Sprintf("attempt %d · the wall: %s", r.Attempts, wall)
		fmt.Fprintf(c.Out, "  %s %s\n", Paint("•", cYellow), r.URL)
		fmt.Fprintf(c.Out, "      %s\n", Paint(line, cDim))
	}
	fmt.Fprint(c.Out, "\n")

	urls := make([]string, 0, len(rows))
	walls := map[string]string{}
	for _, r := range rows {
		urls = append(urls, r.URL)
		walls[r.URL] = r.Wall
	}
	report := handdelivery.EnqueueHandDelivery(vault, urls, walls,
		func(string, string) {}, true, "")
	c.CliPrint(fmt.Sprintf("🖐 %d link(s) queued for hand-delivery, %d opened in your browser.",
		report.Added, report.Opened), "success")
	folder := handdelivery.HandDeliveryDir(vault)
	fmt.Fprint(c.Out, Paint("Next: in Chrome, Ctrl+S → 'Webpage, HTML Only' → the "+
		"suggested filename → this folder:", cDim)+"\n")
	fmt.Fprint(c.Out, Paint(fmt.Sprintf("  %s", folder), cCyan)+"\n")
	fmt.Fprint(c.Out, Paint("The next --auto / --retry-failed run consumes every "+
		"delivered page as a REAL fetch.", cDim)+"\n")
	return 0
}

// chromeLink mirrors the fifth door's link row: {'url', 'error'}.
type chromeLink struct {
	url   string
	errTs string
}

// cmdChromeRetry mirrors cmd_chrome_retry's listing + the dry-run
// rehearsal; the live run (preflight → delivery → the batch
// processing) is T-62's honest boundary card.
func (c *CLI) cmdChromeRetry(a *Args) int {
	c.Banner()
	path := configPath(a.Config)
	cfg, ok := c.loadConfig(path)
	if !ok {
		cfg = storage.NewOrderedMap()
	}
	vault := strings.TrimSpace(cfgOrStr(cfg, "website_vault_path"))
	if vault == "" {
		c.CliPrint("website_vault_path is not set — run --cli --init first.", "warning")
		return 1
	}
	if !pyTruthy(getOr(getMap(cfg, "pipelines"), "websites", false)) {
		c.CliPrint("The websites pipeline is off — turn it on first "+
			"(Settings → 📁 Vault, or config \"pipelines\").", "warning")
		return 1
	}

	dbFile := cacheDBPath()
	var links []chromeLink
	if !fileExists(dbFile) {
		// no cache.db only means no MACHINE fetch has failed; the 🖐
		// hand rows still ask for their Chrome scrape
		c.CliPrint("No cache.db yet — no machine fetch has failed; the "+
			"hand rows still get their Chrome pass.", "info")
	} else {
		state, err := websitepipeline.NewWebsiteStateDB(dbFile)
		if err != nil {
			c.CliPrint(fmt.Sprintf("Retry queue unavailable (%s)", err), "error")
			return 1
		}
		retryRows, rerr := state.AllRetryRows()
		if rerr != nil {
			state.Close()
			c.CliPrint(fmt.Sprintf("Retry queue unavailable (%s)", rerr), "error")
			return 1
		}
		for _, row := range retryRows {
			u := strings.TrimSpace(row.URL)
			lu := strings.ToLower(u)
			if !strings.HasPrefix(lu, "http://") && !strings.HasPrefix(lu, "https://") {
				continue
			}
			if chrometabs.IsLoopbackURL(u) {
				continue
			}
			dismissed, _ := state.IsDismissed(u)
			if dismissed {
				continue // the ladder's own verdict answered it
			}
			errTxt := ""
			if row.LastError != nil {
				errTxt = *row.LastError
			}
			links = append(links, chromeLink{u, errTxt})
		}
		state.Close()
	}

	// the 🖐 hand rows join the fifth door's list (the owner's gesture
	// asks for THIS pass)
	for _, hr := range websitepipeline.ScanMasterHandRows(vault, nil) {
		dup := false
		for _, l := range links {
			if l.url == hr.URL {
				dup = true
				break
			}
		}
		if !dup {
			wall := hr.Wall
			if wall == "" {
				wall = "the 🖐 hand gesture in the master table"
			}
			links = append(links, chromeLink{hr.URL, wall})
		}
	}

	// THE NOTE IS THE SUCCESS — the redo set: 🖐 hand links whose page
	// was delivered but whose note is NOT properly stored. Computed
	// BEFORE the empty check (a redo set alone is work too). A scan
	// failure degrades to the honest skip warning (the reference's
	// try/except).
	var redoURLs []string
	if state, err := websitepipeline.NewWebsiteStateDB(dbFile); err != nil {
		c.CliPrint(fmt.Sprintf("⚠️ Redo scan skipped: %s", err), "warning")
	} else {
		for _, r := range websitepipeline.ScanMasterRedoRows(vault, state, nil) {
			if r.URL != "" {
				redoURLs = append(redoURLs, r.URL)
			}
		}
		state.Close()
	}

	fmt.Fprint(c.Out, Paint("Failed links waiting for the fifth door "+
		"(your real Chrome, one tab each)", cBold)+"\n")
	fmt.Fprint(c.Out, Rule()+"\n")
	if len(links) == 0 && len(redoURLs) == 0 {
		fmt.Fprint(c.Out, Paint("  ✓ empty — no fetch-failed link is waiting, no "+
			"hand note needs a redo.", cGreen)+"\n")
		return 0
	}
	if len(redoURLs) > 0 {
		fmt.Fprint(c.Out, Paint("  🔁 redo — hand note(s) not properly stored:", cYellow)+"\n")
		for _, u := range redoURLs {
			fmt.Fprintf(c.Out, "      %s\n", Paint(u, cDim))
		}
	}
	for _, l := range links {
		errTxt := pySliceRunes(l.errTs, 0, 100)
		fmt.Fprintf(c.Out, "  %s %s\n", Paint("•", cYellow), l.url)
		if errTxt != "" {
			fmt.Fprintf(c.Out, "      %s\n", Paint(errTxt, cDim))
		}
	}
	fmt.Fprint(c.Out, "\n")

	if a.DryRun {
		// The rehearsal: the dry-run switch flipped for the delivery
		// and restored right after (the reality-wins fix — the
		// reference forgot to flip it; the card's own words + the
		// deliver's rehearsal law are the contract). The recorded
		// entries are the rehearsal's log — spoken in the card, then
		// cleared (the batch's dry-run report writer is T-62's).
		dryrun.Enable()
		report := chrometabs.DeliverPagesViaChrome(vault, toChromeLinks(links),
			func(string, string) {}, cfgPlainMap(cfg), "", nil, nil)
		dryrun.Disable()
		dryrun.Clear()
		c.CliPrint(fmt.Sprintf("🤖 Dry-run — Chrome is never launched; the fifth "+
			"door rehearsed %d link(s), delivered %d.", len(links),
			report.Delivered), "success")
		return 0
	}

	// The live leg (preflight → the real Chrome delivery → the batch
	// processing of every delivered page) lands with T-62's engine —
	// the honest boundary card, never a silent no-op.
	c.CliPrint("The fifth door's live run — preflight, the Chrome delivery, and "+
		"the batch processing of every delivered page in the same run — "+
		"lands with T-62 (Phase 5).", "warning")
	c.CliPrint("Nothing was launched, nothing was written; the links keep "+
		"waiting in the retry queue.", "info")
	return 1
}

// toChromeLinks converts the listing rows into chrometabs.Link.
func toChromeLinks(links []chromeLink) []chrometabs.Link {
	out := make([]chrometabs.Link, 0, len(links))
	for _, l := range links {
		out = append(out, chrometabs.Link{URL: l.url, Error: l.errTs})
	}
	return out
}

// cfgPlainMap flattens the OrderedMap config into the plain map the
// chrometabs delivery reads (nested dicts recursively).
func cfgPlainMap(cfg *storage.OrderedMap) map[string]any {
	out := map[string]any{}
	if cfg == nil {
		return out
	}
	for _, k := range cfg.Keys() {
		v, _ := cfg.Get(k)
		if om, ok := v.(*storage.OrderedMap); ok {
			out[k] = cfgPlainMap(om)
		} else {
			out[k] = v
		}
	}
	return out
}
