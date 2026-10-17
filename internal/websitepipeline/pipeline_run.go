package websitepipeline

import (
	"fmt"
	"strings"

	"github.com/assadigit/second-brain-project/internal/handdelivery"
	"github.com/assadigit/second-brain-project/internal/links"
)

// pipeline_run.go — the four batch drivers (website_pipeline.py:4900-
// 5296): run, run_due_retries, retry_review_backlog,
// retry_master_waiting. All end with _refresh_master_table.

// Run mirrors run: process a batch of non-GitHub links (deduped
// upstream by the caller; dedupe is re-checked per link here anyway).
// shouldContinue is polled between links so a GUI Stop button can end
// the phase cleanly. onProgress fires once per link BEFORE it is
// processed. Within-batch duplicates count too: the caller's position
// advances for every link it handed us, mirroring the GitHub loop.
func (p *Pipeline) Run(urls []string, shouldContinue ShouldContinueFn, onProgress OnProgressFn) []Result {
	var results []Result
	seen := map[string]bool{}
	// v0.48.0 — the fourth door's pull: links whose hand-delivered page
	// is WAITING in the folder join the batch uninvited (the owner's
	// gesture — saving the page — says "process me now").
	if p.handDelivery && p.vaultPath != "" {
		func() {
			defer func() { recover() }() // the pull never breaks the batch
			urlSet := map[string]bool{}
			for _, u := range urls {
				urlSet[links.NormalizeWebsiteURL(u)] = true
			}
			var ready []string
			for _, d := range handdelivery.CollectDelivered(p.vaultPath) {
				if !urlSet[links.NormalizeWebsiteURL(d.URL)] {
					ready = append(ready, d.URL)
				}
			}
			if len(ready) > 0 {
				p.log(fmt.Sprintf("🖐 Hand-delivery: %d delivered page(s) waiting in the hand-delivered folder — they join this batch", len(ready)), "info")
				urls = append(append([]string{}, urls...), ready...)
			}
		}()
	}
	for _, u := range urls {
		if shouldContinue != nil && !shouldContinue() {
			p.log("⏹️ Websites pipeline stopped by user — remaining links stay queued for the next run", "warning")
			break
		}
		if onProgress != nil {
			func() {
				defer func() { recover() }()
				onProgress(u)
			}()
		}
		canonical := links.NormalizeWebsiteURL(u)
		if seen[canonical] {
			// Within-batch duplicate: REPORT it (never silently drop —
			// the run report's counts must add up).
			results = append(results, Result{
				URL: u, Canonical: canonical, Outcome: "skipped",
				Error: "duplicate within batch",
			})
			continue
		}
		seen[canonical] = true
		// v0.63.2 — THE SETTLED LEDGER's gate: a link the owner already
		// sent to the bot and addressed is never fetched again. The
		// owner's 🖐 hand outranks the settlement exactly as it outranks
		// the burned-out retry counter.
		skipSettled := false
		func() {
			defer func() { recover() }() // a broken probe never hides a link
			if canonical == "" {
				return
			}
			if settled, err := p.state.IsSettled(canonical); err == nil && settled &&
				!p.handOurranksSettlement(canonical) {
				skipSettled = true
			}
		}()
		if skipSettled {
			result := Result{
				URL: u, Canonical: canonical, Outcome: "skipped",
				Error: "settled — already sent to the bot and addressed (the owner's law); never re-fetched",
			}
			results = append(results, result)
			p.lastResults = append(p.lastResults, result)
			p.counters["skipped"]++
			p.log(fmt.Sprintf("🤝 %s: settled — already addressed and processed; never fetched again (the owner's law; ♻️ revived in the master table un-settles it)", u), "info")
			continue
		}
		results = append(results, p.ProcessLink(u))
	}
	p.refreshMasterTable()
	return results
}

// handOurranksSettlement mirrors _hand_outranks_settlement: does the
// owner's 🖐 hand outrank the settlement for this link? The gesture on
// the master-table row, or a delivered page already waiting in the
// hand-delivered folder. False on any doubt (the settlement holds).
func (p *Pipeline) handOurranksSettlement(canonical string) bool {
	defer func() { recover() }()
	if p.rowReadsHand(canonical) {
		return true
	}
	if p.handDelivery && p.vaultPath != "" {
		for _, d := range handdelivery.CollectDelivered(p.vaultPath) {
			if SameSource(d.URL, canonical) {
				return true
			}
		}
	}
	return false
}

// RunDueRetries mirrors run_due_retries: retry fetch-failed links
// whose backoff elapsed (SPEC §4.3 — "retried automatically up to 3
// times over several days"). Called by the batch BEFORE the fresh
// links. v0.60.1 — the count's honesty (split_dismissed_links);
// v0.63.2 — THE SETTLED LEDGER (split_settled_links).
func (p *Pipeline) RunDueRetries(shouldContinue ShouldContinueFn, onProgress OnProgressFn) []Result {
	due, err := p.state.DueRetries()
	if err != nil || len(due) == 0 {
		return nil
	}
	split1 := SplitDismissedLinks(p.state, due, p.log)
	due = split1.Kept
	split2 := SplitSettledLinks(p.state, due, p.log)
	due = split2.Kept
	if len(due) == 0 {
		return nil
	}
	p.log(fmt.Sprintf("🔁 Retrying %d fetch-failed website link(s) whose backoff elapsed", len(due)), "info")
	var results []Result
	for _, u := range due {
		if shouldContinue != nil && !shouldContinue() {
			p.log("⏹️ Websites retry pass stopped by user", "warning")
			break
		}
		if onProgress != nil {
			func() {
				defer func() { recover() }()
				onProgress(u)
			}()
		}
		results = append(results, p.ProcessLink(u))
	}
	p.counters["retried"] = len(due)
	p.refreshMasterTable()
	return results
}

// RetryReviewBacklog mirrors retry_review_backlog (v0.42.0): retry the
// scanned _review backlog — the whole retry queue is RE-ARMED first
// (attempts=0; the wall pile is exactly the links whose 3 automatic
// retries burned out under the OLD fetcher), a placeholder whose state
// row was LOST gets the row re-registered from the disk truth, and
// process_link does the rest. Any OTHER app-owned failed placeholder
// for the same source (legacy _v1/_v2 stacking) is cleaned after the
// verdict: one source, one note.
func (p *Pipeline) RetryReviewBacklog(items []BacklogItem, shouldContinue ShouldContinueFn, onProgress OnProgressFn) []Result {
	var results []Result
	if len(items) == 0 {
		return results
	}
	func() {
		defer func() { recover() }() // bookkeeping never kills the batch
		if rearmed, err := p.state.RearmRetries(); err == nil && rearmed > 0 {
			p.log(fmt.Sprintf("🔁 _review backlog retry: re-armed %d queued retry(ies) — their 3 attempts burned out under the old fetcher; every link gets a fresh set", rearmed), "info")
		}
	}()
	// Consolidate the scan: one source may own several placeholder
	// files; process the URL once, clean the leftovers after the
	// verdict (dict order = scan order = sorted filenames).
	byURL := map[string][]string{}
	var byURLOrder []string
	for _, it := range items {
		if it.URL == "" {
			continue
		}
		if _, exists := byURL[it.URL]; !exists {
			byURLOrder = append(byURLOrder, it.URL)
		}
		byURL[it.URL] = append(byURL[it.URL], it.Path)
	}
	// v0.63.2 — THE SETTLED LEDGER: settled placeholders never re-fetch
	// (the 🖐 gesture outranks the settlement).
	var settledKeys []string
	func() {
		defer func() { recover() }()
		for _, u := range byURLOrder {
			cu := links.NormalizeWebsiteURL(u)
			if cu == "" {
				continue
			}
			if settled, err := p.state.IsSettled(cu); err == nil && settled &&
				!p.handOurranksSettlement(cu) {
				settledKeys = append(settledKeys, u)
			}
		}
	}()
	if len(settledKeys) > 0 {
		for _, u := range settledKeys {
			delete(byURL, u)
			for i, k := range byURLOrder {
				if k == u {
					byURLOrder = append(byURLOrder[:i], byURLOrder[i+1:]...)
					break
				}
			}
		}
		p.log(fmt.Sprintf("🤝 %d settled link(s) left the _review backlog retry — already sent to the bot and addressed; never fetched again (♻️ revived in the master table un-settles any of them)", len(settledKeys)), "info")
	}
	p.counters["retried"] += len(byURL)
	for _, u := range byURLOrder {
		paths := byURL[u]
		if shouldContinue != nil && !shouldContinue() {
			p.log("⏹️ _review backlog retry stopped by user — the remaining placeholders keep waiting", "warning")
			break
		}
		if onProgress != nil {
			func() {
				defer func() { recover() }()
				onProgress(u)
			}()
		}
		canonical := links.NormalizeWebsiteURL(u)
		var prior *ProcessedRow
		if pr, err := p.state.ProcessedRow(canonical); err == nil {
			prior = pr
		}
		if prior == nil && len(paths) > 0 {
			// State row lost — the note on disk is the proof. Register
			// the row the older app SHOULD have written so process_link
			// routes this link through the upgrade path, not "already in
			// the vault". Never clobbers an existing row (prior is nil).
			func() {
				defer func() { recover() }()
				if err := p.state.MarkProcessed(canonical, paths[0], "", "", "failed"); err == nil {
					if rr, err := p.state.RetryRow(canonical); err == nil && rr == nil {
						_ = p.state.EnqueueRetry(canonical,
							"re-discovered _review placeholder (state row was lost)")
					}
					p.log(fmt.Sprintf("🗂️ %s: state row was lost — re-registered its _review placeholder before retrying", u), "info")
				}
			}()
		}
		r := p.ProcessLink(u)
		freshPath := r.NotePath
		// Stale duplicate cleanup: only when THIS run produced a note
		// (processed / review) or the link is already stored for real —
		// never on a dismissed or blocked skip.
		if r.Outcome == "processed" || r.Outcome == "review" ||
			(r.Outcome == "skipped" && strings.Contains(r.Error, "already in")) {
			for _, path := range paths {
				if path != "" && path != freshPath {
					p.removeScannedPlaceholder(path)
				}
			}
		}
		results = append(results, r)
	}
	p.refreshMasterTable()
	return results
}

// RetryMasterWaiting mirrors retry_master_waiting (v0.51.0 — the
// caught-up check's own retry pass): the " - " rows of the master
// table (no verdict emoji, no decision — the link is valid, its fetch
// failed) are fetched AGAIN before any "everything is up to date" may
// be said. A burned-out retry row gets its counter reborn (ONE row,
// not the whole queue); the "eyes" rows are never fetched — only
// counted and named. Ends with the honest verdict — a re-scan of the
// table: what still waits is said out loud.
func (p *Pipeline) RetryMasterWaiting(shouldContinue ShouldContinueFn, onProgress OnProgressFn) []Result {
	var results []Result
	if p.vaultPath == "" {
		return results
	}
	items := ScanMasterWaitingRows(p.vaultPath, p.state, p.log)
	var fetchItems, eyes []WaitingRow
	for _, it := range items {
		if it.Kind == "fetch" {
			fetchItems = append(fetchItems, it)
		} else if it.Kind == "eyes" {
			eyes = append(eyes, it)
		}
	}
	if len(fetchItems) == 0 {
		if len(eyes) > 0 {
			p.log(fmt.Sprintf("📋 Master table: no ' - ' fetch row is waiting — %d row(s) wait for YOUR eyes (set ✅ reviewed or 🪦 dead in the table; ♻️ revived re-fetches)", len(eyes)), "info")
		} else {
			p.log("✅ Master table: no ' - ' row is waiting — the table agrees with 'everything is up to date'", "success")
		}
		return results
	}
	p.log(fmt.Sprintf("🔁 Master table: %d ' - ' row(s) waiting (valid link, failed fetch, no verdict) — fetching them again before 'everything is up to date' is said", len(fetchItems)), "info")
	p.counters["retried"] += len(fetchItems)
	for _, it := range fetchItems {
		if shouldContinue != nil && !shouldContinue() {
			p.log("⏹️ Master-table waiting pass stopped by user — the remaining rows keep waiting", "warning")
			break
		}
		if onProgress != nil {
			func() {
				defer func() { recover() }()
				onProgress(it.URL)
			}()
		}
		u := it.URL
		canonical := links.NormalizeWebsiteURL(u)
		// A burned-out retry row would be skipped as "no more retries" —
		// the owner's " - " says otherwise: reborn.
		func() {
			defer func() { recover() }()
			if canonical == "" {
				return
			}
			if rr, err := p.state.RetryRow(canonical); err == nil && rr != nil {
				_, _ = p.state.ResetRetryAttempts(canonical)
			}
		}()
		results = append(results, p.ProcessLink(u))
	}
	p.refreshMasterTable()
	// the honest verdict — a re-scan says what still waits
	var still, eyesAfter []WaitingRow
	func() {
		defer func() { recover() }()
		after := ScanMasterWaitingRows(p.vaultPath, p.state, nil)
		for _, a := range after {
			if a.Kind == "fetch" {
				still = append(still, a)
			} else if a.Kind == "eyes" {
				eyesAfter = append(eyesAfter, a)
			}
		}
	}()
	if len(still) > 0 {
		p.log(fmt.Sprintf("⚠️ Master table: %d ' - ' row(s) still wait — the machine doors had their say; the rest is yours: set the Status (🪦 dead / ✅ reviewed), queue 🖐 hand, or retry in your real Chrome (the end-of-run offer / More ▸ 🤖 Chrome tab-retry)", len(still)), "warning")
	} else {
		p.log("✅ Master table: every ' - ' fetch row got its retry — the table is honest now", "success")
	}
	if len(eyesAfter) > 0 {
		p.log(fmt.Sprintf("📋 %d row(s) wait for YOUR eyes (set ✅ or 🪦 in the table; ♻️ revived re-fetches)", len(eyesAfter)), "info")
	}
	return results
}

// refreshMasterTable mirrors _refresh_master_table: one call at every
// driver's end. Tolerated everywhere; the dry-run law applies through
// the writer.
func (p *Pipeline) refreshMasterTable() {
	if p.vaultPath == "" {
		return
	}
	func() {
		defer func() { recover() }() // bookkeeping never kills a batch
		RefreshMasterTable(p.state, p.vaultPath, p.log)
	}()
}
