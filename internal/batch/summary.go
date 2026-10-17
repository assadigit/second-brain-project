// summary.go — T-42: the v0.39.0 end-of-batch scorecard
// (gui/processing_worker.py::_build_batch_summary, lines 1894-1956) —
// the structured summary MainWindow's success fanfare (chime +
// scorecard modal) renders from. Presence of the struct IS the signal
// that the batch reached the natural end-of-batch stage; the early
// returns leave it nil and the GUI keeps the plain message box.
//
// Deliberately pure attribute reads (processed/total/websites/
// link-tracker manifest/is_running) — no I/O, no clocks — so tests
// drive it through the harness shape without a real batch, exactly
// like the reference's own headless __new__ harness
// (tests/test_batchfinish.py::TestBuildBatchSummary).
package batch

// WebsitesCounters is the picked websites tally the modal renders:
// _ws['counters'] filtered to the four keys the summary carries
// (processed / review / skipped / failed — the other counters
// ('retried', 'upgraded') are dropped on the floor by the reference).
type WebsitesCounters struct {
	Processed int
	Review    int
	Skipped   int
	Failed    int
}

// BatchSummary mirrors _build_batch_summary's dict, key for key:
//
//	stopped            True → the user hit Stop (loop broke early)
//	github_processed   repos curated into the GitHub vault
//	github_total       the bar's final denominator
//	new_notes          NEW .md files this batch wrote (undo tally)
//	websites           WebsitesCounters, or nil when the phase never
//	                   ran (pipeline off / no site links)
//	failed_links       manifest rows marked failed (this batch)
//	report_path        _processing_report_*.md ("" when skipped)
//	summary_path       the summary log path
type BatchSummary struct {
	Stopped         bool
	GithubProcessed int
	GithubTotal     int
	NewNotes        int
	Websites        *WebsitesCounters
	FailedLinks     int
	ReportPath      string
	SummaryPath     string
}

// BuildBatchSummary mirrors _build_batch_summary — every read is the
// reference's own defensive shape: a nil tracker reads as zero failed
// links (the modal just shows 0), a nil websites phase renders as nil,
// and int(...) coercions clamp garbage to 0. The tracker's manifest is
// read live: rows whose status == "failed" count (the same rows the
// retry queue will surface).
func BuildBatchSummary(stopped bool, processed, total, newNotes int,
	websites *WebsitesCounters, tracker *LinkTracker,
	reportPath, summaryPath string) *BatchSummary {
	failed := 0
	if tracker != nil {
		// best-effort — a broken manifest just shows 0 (the Python
		// except swallows the AttributeError the same way).
		func() {
			defer func() { recover() }()
			if tracker.Manifest == nil {
				return
			}
			for _, l := range tracker.Manifest.Links {
				if l != nil && l.Status == "failed" {
					failed++
				}
			}
		}()
	}
	return &BatchSummary{
		Stopped:         stopped,
		GithubProcessed: processed,
		GithubTotal:     total,
		NewNotes:        newNotes,
		Websites:        websites,
		FailedLinks:     failed,
		ReportPath:      reportPath,
		SummaryPath:     summaryPath,
	}
}
