package websitepipeline

import (
	"fmt"
	"strings"

	"github.com/assadigit/second-brain-project/internal/links"
)

// master.go — v0.47.0 the master table refresh (refresh_master_table,
// website_pipeline.py:3161-3329): keep the master note honest after
// every batch — every WAITING failure as fresh 'unreviewed' rows, the
// truth stamp, the hand harvest, and v0.60.2's compact-table prune.

// RefreshReport mirrors refresh_master_table's dict:
// {'waiting', 'retired', 'written'}.
type RefreshReport struct {
	Waiting int
	Retired int
	Written int
}

// RefreshMasterTable mirrors refresh_master_table. The v0.60.2 COMPACT
// TABLE: the auto-verdict rows are no longer WRITTEN (the verdict's
// record lives in the state DB and the run logs; writing them here
// would churn — the prune removes them the same pass). Existing rows
// are never duplicated or clobbered (the writer's law); a vault with
// nothing waiting and nothing retired stays table-less (a cheap
// no-op). Tolerated everywhere — the master table is bookkeeping,
// never a batch killer.
func RefreshMasterTable(state *StateDB, vaultPath string, log LogFn) RefreshReport {
	if log == nil {
		log = noopLog
	}
	report := RefreshReport{}
	if vaultPath == "" {
		return report
	}
	// v0.52.0 — a trimmed table gets its full grammar back first (the
	// stamp this pass writes must land in a cell the note can render).
	func() {
		defer func() { recover() }()
		EnsureTableShape(DecommissionTablePath(vaultPath), log)
	}()
	func() {
		defer func() { recover() }() // bookkeeping never kills a batch
		var waitingURLs []string
		waitingNotes := map[string]string{}
		dismissedCache := map[string]bool{}

		isRetired := func(u string) bool {
			if hit, cached := dismissedCache[u]; cached {
				return hit
			}
			hit := false
			if v, err := state.IsDismissed(u); err == nil {
				hit = v
			}
			dismissedCache[u] = hit
			return hit
		}

		retryRows, err := state.AllRetryRows()
		if err == nil {
			for _, row := range retryRows {
				u := row.URL
				if u == "" || isRetired(u) {
					continue // retired — its row comes below, not here
				}
				waitingURLs = append(waitingURLs, u)
				waitingNotes[u] = ptrStr(row.LastError)
				report.Waiting++
			}
		}
		for _, it := range ScanReviewBacklog(vaultPath, nil, isRetired) {
			u := it.URL
			if u == "" {
				continue
			}
			if _, exists := waitingNotes[u]; exists {
				continue
			}
			waitingURLs = append(waitingURLs, u)
			waitingNotes[u] = "placeholder in _review (state row was lost)"
			report.Waiting++
		}
		// v0.49.0 — the human-eye classes: every OTHER app-owned note in
		// _review (low classification confidence, analysis failures,
		// archived rescues).
		for _, it := range ScanReviewNotes(vaultPath, nil) {
			u := it.URL
			if u == "" {
				continue
			}
			if _, exists := waitingNotes[u]; exists {
				continue
			}
			if strings.EqualFold(strings.TrimSpace(it.FetchStatus), "failed") {
				continue // a failed placeholder is a retry row or a
				// lost-row backlog item — both counted above
			}
			cu := links.NormalizeWebsiteURL(u)
			if cu == "" {
				cu = u
			}
			if _, exists := waitingNotes[cu]; exists {
				continue // canonical twin already waiting
			}
			if isRetired(cu) {
				continue
			}
			reason := ReviewNoteReason(it.Path)
			fs := strings.TrimSpace(it.FetchStatus)
			if fs == "" {
				fs = "review"
			}
			note := "in _review (" + fs + "): "
			if reason == "" {
				reason = "waiting for the owner’s move"
			}
			note += reason
			waitingURLs = append(waitingURLs, cu)
			waitingNotes[cu] = note
			report.Waiting++
		}

		// v0.60.2 — the auto-verdict rows are counted for the report,
		// never written (the compact-table law).
		if rows, err := state.DismissedRows("auto-verdict:"); err == nil {
			for _, row := range rows {
				if row.URL == "" {
					continue
				}
				report.Retired++
			}
		}

		n1 := 0
		if len(waitingURLs) > 0 {
			n1 = WriteDecommissionCandidates(vaultPath, waitingURLs, "waiting", nil, waitingNotes, "unreviewed")
		}
		report.Written = n1
		// v0.51.0 — the truth stamp: rows still marked " - " whose
		// retries have since SUCCEEDED are stamped '📁 stored'.
		func() {
			defer func() { recover() }()
			StampStoredRows(state, vaultPath, log)
		}()
		// v0.56.0 — THE HAND'S HARVEST: 🖐 rows whose links are already
		// stored retire to the green checkbox.
		func() {
			defer func() { recover() }()
			HarvestHandRows(state, vaultPath, log)
		}()
		// v0.60.2 — THE PRUNE: terminal rows leave the table.
		func() {
			defer func() { recover() }()
			PruneDecommissionTable(vaultPath, log)
		}()
		if report.Written > 0 {
			log(fmt.Sprintf("📋 Master table refreshed: %d waiting (fetch failures + links parked in _review), %d auto-retired link(s) — %d new row(s) in %s (🪦 dead / ✅ reviewed retire; ♻️ revived re-fetches; 🖐 hand queues for the fourth door — your real Chrome; terminal rows leave the table — the state DB and .trash keep their records)",
				report.Waiting, report.Retired, report.Written, DecommissionTable), "info")
		}
	}()
	return report
}
