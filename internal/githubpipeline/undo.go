// undo.go — T-42: the v22 Feature 6 batch undo — the pre-batch
// snapshot (processing_worker.py:930-945) and the post-batch diff +
// _undo_last_batch.txt writer (1744-1780). The snapshot walks the
// vault for .md files (skipping .obsidian/attachments), the diff
// re-walks after the batch (also skipping .trash — v0.66.0: the
// hygiene pass's retired stubs are NOT "new notes the batch wrote"),
// and the new files list lands in the vault root as the undo list the
// Dashboard's "Undo Last Batch" button consumes.
//
// The walk is os.walk's files-before-subdirs order (the T-41 law),
// with the reference's SUBSTRING root-skip quirk preserved: the skip
// test is `any(skip in root ...)` — a path SEGMENT substring, not a
// component match (a vault rooted at /tmp/attachments-demo/ skips
// whole, exactly like the reference).
package githubpipeline

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/assadigit/second-brain-project/internal/batch"
	"github.com/assadigit/second-brain-project/internal/dryrun"
	"github.com/assadigit/second-brain-project/internal/store"
)

// snapshotSkips / undoSkips mirror the two literal skip lists.
var snapshotSkips = []string{".obsidian", "attachments"}
var undoSkips = []string{".obsidian", "attachments", ".trash"}

// SnapshotBatchFiles mirrors the pre-loop snapshot: the set of .md
// paths that already exist. Best-effort (an unreadable vault yields an
// empty set — undo just won't work that run).
func SnapshotBatchFiles(vaultPath string) map[string]bool {
	files := map[string]bool{}
	if vaultPath == "" {
		return files
	}
	if fi, err := os.Stat(vaultPath); err != nil || !fi.IsDir() {
		return files
	}
	walkVaultMD(vaultPath, snapshotSkips, func(path string) {
		files[path] = true
	})
	return files
}

// walkVaultMD is the pruned os.walk recursion: a directory's files
// before ANY subdirectory's files (the T-41 law), the substring
// root-skip applied BEFORE a directory is entered (a skipped root's
// whole subtree never yields).
func walkVaultMD(dir string, skips []string, visit func(path string)) {
	if rootSkipped(dir, skips) {
		return
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return // os.walk's onerror=None swallows read errors
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".md") {
			visit(filepath.Join(dir, name))
		}
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		walkVaultMD(filepath.Join(dir, e.Name()), skips, visit)
	}
}

// rootSkipped mirrors `any(skip in root for skip in [...])` — the
// substring law.
func rootSkipped(dir string, skips []string) bool {
	for _, s := range skips {
		if strings.Contains(dir, s) {
			return true
		}
	}
	return false
}

// ComputeNewFiles mirrors the diff walk: every .md now in the vault
// that was not in the snapshot. The reference's falsiness quirk is
// preserved — an EMPTY snapshot means NO diff run at all (new_files
// stays empty even if the vault now has files).
func ComputeNewFiles(vaultPath string,
	batchFiles map[string]bool) []string {
	var newFiles []string
	if vaultPath == "" {
		return newFiles
	}
	if fi, err := os.Stat(vaultPath); err != nil || !fi.IsDir() {
		return newFiles
	}
	if len(batchFiles) == 0 {
		return newFiles
	}
	walkVaultMD(vaultPath, undoSkips, func(path string) {
		if !batchFiles[path] {
			newFiles = append(newFiles, path)
		}
	})
	return newFiles
}

// WriteUndoFile mirrors the _undo_last_batch.txt writer: one path per
// line, every line newline-terminated (the ”.join(f + '\n') shape —
// an empty list writes an EMPTY file). Best-effort: any error is
// logged, never raised.
func (r *Runner) WriteUndoFile(vaultPath string, newFiles []string) bool {
	var b strings.Builder
	for _, f := range newFiles {
		b.WriteString(f + "\n")
	}
	undoPath := filepath.Join(vaultPath, "_undo_last_batch.txt")
	if err := dryrun.WriteText(undoPath, b.String()); err != nil {
		r.log("⚠️ Failed to save batch undo list: "+err.Error(), "warning")
		return false
	}
	return true
}

// Finish is the post-loop completion chain (1727-1889) in the
// reference's exact order: the undo diff, the 100% progress, the
// LinkTracker verification (with the websites vault's _inbox as the
// extra inbox dir), the summary log, the hygiene pass seam, the final
// report, and the batch summary — then the finished message with the
// aggregate lines. The shell closes its own cache/note-state handles
// around this call (the reference closes them between the websites
// phase and the undo diff; nothing in the chain needs them).
func (r *Runner) Finish() *batch.BatchSummary {
	// v22 Feature 6 — compute the list of NEW .md files.
	newNotes := 0
	vaultPath := cfgStr(r.Config, "vault_path")
	vaultIsDir := false
	if vaultPath != "" {
		if fi, err := os.Stat(vaultPath); err == nil && fi.IsDir() {
			vaultIsDir = true
		}
	}
	if vaultIsDir {
		newFiles := ComputeNewFiles(vaultPath, r.batchFiles)
		newNotes = len(newFiles)
		if ok := r.WriteUndoFile(vaultPath, newFiles); ok && len(newFiles) > 0 {
			r.log("↩️ Batch undo saved: "+itoa(len(newFiles))+
				" new files can be undone via Dashboard → 'Undo Last Batch'",
				"info")
		}
	}

	// Final progress update to 100%.
	r.position = r.Total
	r.emitProgress()

	// v23 — Phase 3: VERIFICATION.
	var report *batch.VerifyReport
	if r.LinkTracker != nil {
		func() {
			defer func() { recover() }()
			var extra []string
			if wv := cfgStr(r.Config, "website_vault_path"); wv != "" {
				extra = append(extra, filepath.Join(wv, "_inbox"))
			}
			report = r.LinkTracker.Verify(r.Log, extra)
			if !report.VerificationPassed {
				r.log("⚠️ "+itoa(len(report.FailedLinks))+
					" links failed verification — will retry on next run",
					"warning")
			}
		}()
	}

	// Generate the summary txt log.
	summaryPath := r.GenerateSummaryLog()

	// v0.66.0 — the vault hygiene pass (the seam; the module is a
	// later task's port — nil skips it exactly like a broken pass
	// would, best-effort and never a batch killer).
	if r.HygienePass != nil {
		func() {
			defer func() { recover() }()
			r.HygienePass()
		}()
	}

	// v25 pre-flight: the comprehensive final report — always
	// generated so the user has a complete audit trail.
	reportPath := ""
	func() {
		defer func() { recover() }()
		reportPath = r.GenerateFinalReport(report)
		if reportPath != "" {
			r.log("📊 Final report saved: "+reportPath, "success")
		}
	}()

	// v0.39.0 — the structured end-of-batch summary.
	var websites *batch.WebsitesCounters
	if r.websites != nil {
		websites = &batch.WebsitesCounters{
			Processed: r.websites.Counters["processed"],
			Review:    r.websites.Counters["review"],
			Skipped:   r.websites.Counters["skipped"],
			Failed:    r.websites.Counters["failed"],
		}
	}
	summary := batch.BuildBatchSummary(!r.running, r.Processed, r.Total,
		newNotes, websites, r.LinkTracker, reportPath, summaryPath)

	msg := "Processed " + itoa(r.Processed) + " out of " +
		itoa(r.Total) + " repos."
	if r.websites != nil {
		msg += " Websites: " + itoa(r.websites.Counters["processed"]) +
			" processed, " + itoa(r.websites.Counters["review"]) +
			" to review, " + itoa(r.websites.Counters["skipped"]) + " skipped."
	}
	if summaryPath != "" {
		msg += " Summary log: " + summaryPath
	}
	if r.Finished != nil {
		func() {
			defer func() { recover() }()
			r.Finished(true, msg)
		}()
	}
	r.log("🏁 Done. Processed "+itoa(r.Processed)+" repos.", "info")
	// The aggregate lines (v0.08 / v0.63.3 / v0.12.0).
	if r.DeadSkipped != 0 {
		r.log("🚫 "+itoa(r.DeadSkipped)+" dead link(s) skipped — 404 "+
			"quarantine (confirmed after "+
			itoa(store.DeadLinkThreshold)+" attempts in earlier runs; "+
			"More ▸ View 404 Quarantine to manage).", "info")
	}
	if r.SettledReposSkipped != 0 {
		r.log("🤝 "+itoa(r.SettledReposSkipped)+" settled repo link(s) "+
			"skipped — already sent to the bot and addressed (the owner's "+
			"law: only repos added from now on are fetched; ♻️ Reset 404 "+
			"Quarantine un-settles any of them).", "info")
	}
	if r.DismissedSkipped != 0 {
		r.log("🗑️ "+itoa(r.DismissedSkipped)+" dismissed link(s) skipped — "+
			"you deleted their notes (run report lists them).", "info")
	}
	if summaryPath != "" {
		r.log("📝 Summary log saved: "+summaryPath, "success")
	}
	return summary
}

// itoa is the strconv twin (kept for log-line brevity).
func itoa(n int) string { return strconv.Itoa(n) }
