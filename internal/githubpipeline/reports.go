// reports.go — T-42: worker/reports.py's writers (94-635) —
// _reports_dir/_reports_dir_for, _generate_summary_log (529-634) and
// _generate_final_report (205-527). v0.66.0's law: the vault is the
// library, not the filing cabinet — the reports live in the app's
// reports/ folder (config-overridable with an absolute path), never
// the vault root.
package githubpipeline

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/assadigit/second-brain-project/internal/batch"
	"github.com/assadigit/second-brain-project/internal/constants"
	"github.com/assadigit/second-brain-project/internal/dryrun"
	"github.com/assadigit/second-brain-project/internal/helpers"
	"github.com/assadigit/second-brain-project/internal/links"
	"github.com/assadigit/second-brain-project/internal/notestate"
)

// ReportsDir mirrors _reports_dir: the app's reports/ folder.
func ReportsDir() string {
	return filepath.Join(constants.AppDir(), "reports")
}

// ReportsDirFor mirrors _reports_dir_for: the CONFIG-aware reports dir
// (an absolute reports_dir overrides; empty/relative fall back).
func ReportsDirFor(config map[string]any) string {
	override := cfgStr(config, "reports_dir")
	if override != "" && filepath.IsAbs(override) {
		return override
	}
	return ReportsDir()
}

// GenerateSummaryLog mirrors _generate_summary_log: the .txt summary
// of processed repos + the summary_keep_last rotation (default 10; 0
// keeps everything). Returns the path, "" on failure (never raises).
func (r *Runner) GenerateSummaryLog() string {
	vaultPath := cfgStr(r.Config, "vault_path")
	if vaultPath == "" {
		return ""
	}
	if fi, err := os.Stat(vaultPath); err != nil || !fi.IsDir() {
		return ""
	}

	timestamp := r.clockNow().Format("20060102_150405")
	filename := "processing_summary_" + timestamp + ".txt"
	reportsDir := ReportsDirFor(r.Config)
	if err := dryrun.MkdirAll(reportsDir, true); err != nil {
		r.log(fmt.Sprintf("Failed to generate summary log: %s", err), "warning")
		return ""
	}
	filePath := filepath.Join(reportsDir, filename)

	processed := r.processedLog
	total := r.Total
	successCount := len(processed)
	skipped := total - successCount

	var lines []string
	lines = append(lines, strings.Repeat("=", 60))
	lines = append(lines, "GITHUB PROJECT CURATOR - PROCESSING SUMMARY")
	lines = append(lines, strings.Repeat("=", 60))
	lines = append(lines, "Date: "+r.clockNow().Format("2006-01-02 15:04:05"))
	lines = append(lines, "Total URLs: "+strconv.Itoa(total))
	lines = append(lines, "Notes created: "+strconv.Itoa(successCount))
	lines = append(lines, "Skipped (duplicates/errors): "+strconv.Itoa(skipped))
	banners := 0
	for _, p := range processed {
		if p.Banner {
			banners++
		}
	}
	lines = append(lines, "Banners downloaded: "+strconv.Itoa(banners))
	lines = append(lines, strings.Repeat("=", 60))
	lines = append(lines, "")

	if len(processed) > 0 {
		lines = append(lines, "PROCESSED REPOS:")
		lines = append(lines, strings.Repeat("-", 60))
		for i, p := range processed {
			lines = append(lines, fmt.Sprintf("%d. %s", i+1, p.Repo))
			lines = append(lines, "   URL: "+p.URL)
			lines = append(lines, "   Category: "+p.Category)
			lines = append(lines, fmt.Sprintf("   Credibility: %s/100",
				pyFloatStr(p.Credibility)))
			lines = append(lines, "   Banner: "+pyBannerYesNo(p.Banner))
			lines = append(lines, "   Note: "+filepath.Base(p.NotePath))
			lines = append(lines, "")
		}
	} else {
		lines = append(lines, "No repos were processed in this run.")
		lines = append(lines, "")
	}

	// Non-GitHub links section.
	nonGithub := r.NonGithubURLs
	if len(nonGithub) > 0 {
		lines = append(lines, strings.Repeat("=", 60))
		if pipelinesWebsitesOn(r.Config) {
			lines = append(lines, "NON-GITHUB LINKS (processed by the Websites pipeline)")
			lines = append(lines, strings.Repeat("=", 60))
			lines = append(lines, "Count: "+strconv.Itoa(len(nonGithub)))
			lines = append(lines, "Outcome: see the run report / the log above")
		} else {
			lines = append(lines, "NON-GITHUB LINKS (not processed — review manually)")
			lines = append(lines, strings.Repeat("=", 60))
			lines = append(lines, "Count: "+strconv.Itoa(len(nonGithub)))
			lines = append(lines, "Stub notes created in: _inbox/ folder")
		}
		lines = append(lines, strings.Repeat("-", 60))
		for i, url := range nonGithub {
			lines = append(lines, fmt.Sprintf("%d. %s", i+1, url))
		}
		lines = append(lines, "")
	}

	lines = append(lines, strings.Repeat("=", 60))
	lines = append(lines, "END OF SUMMARY")
	lines = append(lines, strings.Repeat("=", 60))

	if err := dryrun.WriteText(filePath, strings.Join(lines, "\n")); err != nil {
		r.log(fmt.Sprintf("Failed to generate summary log: %s", err), "warning")
		return ""
	}

	// v0.26.0 — the rotation: keep only the newest summary_keep_last.
	keep := pyIntRaw(r.Config, "summary_keep_last", 10)
	if keep > 0 {
		existing := summaryFiles(reportsDir)
		for _, oldName := range existing[:maxInt(0, len(existing)-keep)] {
			_ = dryrun.Remove(filepath.Join(reportsDir, oldName))
		}
	}
	return filePath
}

// summaryFiles lists processing_summary_*.txt sorted (the rotation's
// victim order: oldest first).
func summaryFiles(reportsDir string) []string {
	entries, err := os.ReadDir(reportsDir)
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() {
			continue
		}
		if strings.HasPrefix(name, "processing_summary_") &&
			strings.HasSuffix(name, ".txt") {
			out = append(out, name)
		}
	}
	sort.Strings(out)
	return out
}

// GenerateFinalReport mirrors _generate_final_report (205-527): the
// comprehensive Markdown audit trail — summary table, note-state
// section, verification report, per-category, all-processed, failed,
// non-GitHub, websites. Always runs. Returns the path, "" on failure.
func (r *Runner) GenerateFinalReport(
	linkTrackerReport *batch.VerifyReport) string {
	vaultPath := cfgStr(r.Config, "vault_path")
	if vaultPath == "" {
		return ""
	}
	if fi, err := os.Stat(vaultPath); err != nil || !fi.IsDir() {
		return ""
	}

	timestamp := r.clockNow().Format("20060102_150405")
	reportsDir := ReportsDirFor(r.Config)
	if err := dryrun.MkdirAll(reportsDir, true); err != nil {
		r.log(fmt.Sprintf("⚠️ Failed to write final report: %s", err), "warning")
		return ""
	}
	reportPath := filepath.Join(reportsDir, "_processing_report_"+timestamp+".md")

	processed := r.processedLog
	total := r.Total
	successCount := r.Processed
	failedCount := total - successCount
	if failedCount < 0 {
		failedCount = 0
	}

	// Count categories (from _processed_log).
	categories := map[string]int{}
	var categoryOrder []string
	for _, p := range processed {
		if _, ok := categories[p.Category]; !ok {
			categoryOrder = append(categoryOrder, p.Category)
		}
		categories[p.Category]++
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("# 📊 Processing Report — %s",
		r.clockNow().Format("2006-01-02 15:04")))
	lines = append(lines, "")
	source := "import/telegram"
	if r.BotSource {
		source = "bot"
	}
	lines = append(lines, fmt.Sprintf("> Source: `%s` | Batch size: %d | Worker v25",
		source, total))
	lines = append(lines, "")

	// Intake duplicates (Feature 7).
	intakeDupes := r.IntakeDuplicates
	if intakeDupes > 0 {
		uniqueTotal := total + len(r.NonGithubURLs)
		rawTotal := r.RawURLCount
		if rawTotal == 0 {
			rawTotal = uniqueTotal + intakeDupes
		}
		lines = append(lines, fmt.Sprintf(
			"> 🔄 **%d duplicate URL(s) removed during intake** (%d unique from %d total)",
			intakeDupes, uniqueTotal, rawTotal))
		lines = append(lines, "")
	}

	lines = append(lines, "## 📈 Summary")
	lines = append(lines, "")
	lines = append(lines, "| Metric | Count |")
	lines = append(lines, "|--------|-------|")
	lines = append(lines, fmt.Sprintf("| 📬 Total links in batch | %d |", total))
	lines = append(lines, fmt.Sprintf("| ✅ Successfully processed | %d |", successCount))
	lines = append(lines, fmt.Sprintf("| ❌ Failed (will retry) | %d |", failedCount))
	skippedCount := total - successCount - failedCount
	if skippedCount < 0 {
		skippedCount = 0
	}
	if linkTrackerReport != nil {
		skippedCount = linkTrackerReport.GithubSkipped
	}
	lines = append(lines, fmt.Sprintf("| ⏭️ Skipped (dedup) | %d |", skippedCount))
	lines = append(lines, fmt.Sprintf("| 📁 Categories used | %d |", len(categories)))
	lines = append(lines, "")

	// The note-state section (v0.12.0 — Phase 3).
	nsLines := []string{}
	for _, vkey := range []string{"github", "websites"} {
		ns := r.NoteStateRun[vkey]
		if ns == nil {
			continue
		}
		ch := ns.Changes
		ap := ns.Applied
		counts := []int{
			ap.MovedApplied, ap.MovedUnmapped, ap.Dismissed,
			len(ch.Edited), len(ch.Duplicates), len(ch.Unmanaged),
			len(ch.Unmapped), len(ch.Unknown),
		}
		anyCount := false
		for _, c := range counts {
			if c != 0 {
				anyCount = true
				break
			}
		}
		if !anyCount && ns.Baseline == nil {
			continue
		}
		vname := "GitHub"
		if vkey == "websites" {
			vname = "Websites"
		}
		nsLines = append(nsLines, "### "+vname+" vault")
		nsLines = append(nsLines, "")
		if ns.Baseline != nil {
			nsLines = append(nsLines, fmt.Sprintf(
				"- 📋 Baseline recorded: %d notes "+
					"(first run after v0.12.0 — nothing flagged)", *ns.Baseline))
		}
		if ap.MovedApplied != 0 {
			nsLines = append(nsLines, fmt.Sprintf(
				"- 📌 Moves accepted as corrections: %d", ap.MovedApplied))
			for _, ml := range notestate.MoveSummaryLines(ap.Corrections) {
				nsLines = append(nsLines, "  - "+ml)
			}
		}
		if ap.MovedUnmapped != 0 {
			nsLines = append(nsLines, fmt.Sprintf(
				"- 📍 Moved into unmapped folders (kept as-is): %d",
				ap.MovedUnmapped))
		}
		if ap.Dismissed != 0 {
			nsLines = append(nsLines, fmt.Sprintf(
				"- 🗑️ Deleted notes dismissed (never re-added): %d",
				ap.Dismissed))
		}
		if len(ch.Edited) != 0 {
			nsLines = append(nsLines, fmt.Sprintf(
				"- ✏️ Edited by hand (skipped, listed): %d", len(ch.Edited)))
			for i := 0; i < len(ch.Edited) && i < 10; i++ {
				nsLines = append(nsLines, "  - "+ch.Edited[i].Path)
			}
		}
		if len(ch.Duplicates) != 0 {
			nsLines = append(nsLines, fmt.Sprintf(
				"- ⚠️ Duplicates (flagged, untouched): %d", len(ch.Duplicates)))
		}
		if len(ch.Unmanaged) != 0 {
			nsLines = append(nsLines, fmt.Sprintf(
				"- 📄 Unmanaged files (no source, ignored): %d",
				len(ch.Unmanaged)))
		}
		if len(ch.Unmapped) != 0 {
			nsLines = append(nsLines, fmt.Sprintf(
				"- ❓ Notes in unmapped folders (kept, reported): %d",
				len(ch.Unmapped)))
			for i := 0; i < len(ch.Unmapped) && i < 10; i++ {
				nsLines = append(nsLines, "  - "+ch.Unmapped[i].Folder+" — "+
					ch.Unmapped[i].Path)
			}
		}
		if len(ch.Unknown) != 0 {
			nsLines = append(nsLines, fmt.Sprintf(
				"- ❔ Notes with an unrecorded source (listed): %d",
				len(ch.Unknown)))
		}
		nsLines = append(nsLines, "")
	}
	if r.DismissedSkipped != 0 {
		nsLines = append(nsLines, fmt.Sprintf(
			"- 🚫 Dismissed URLs skipped in this batch: %d", r.DismissedSkipped))
		nsLines = append(nsLines, "")
	}
	if len(nsLines) > 0 {
		lines = append(lines, "## 🔄 Note State (moves are corrections)")
		lines = append(lines, "")
		lines = append(lines, nsLines...)
	}

	// The LinkTracker verification report.
	if linkTrackerReport != nil {
		rep := linkTrackerReport
		lines = append(lines, "## 🔍 Verification Report")
		lines = append(lines, "")
		lines = append(lines, "| Check | Result |")
		lines = append(lines, "|-------|--------|")
		lines = append(lines, fmt.Sprintf("| 🔍 Total links verified | %d |", rep.Total))
		lines = append(lines, fmt.Sprintf("| ✅ GitHub processed | %d |", rep.GithubProcessed))
		lines = append(lines, fmt.Sprintf("| ⏭️ GitHub skipped (dedup) | %d |", rep.GithubSkipped))
		lines = append(lines, fmt.Sprintf("| ❌ GitHub failed | %d |", rep.GithubFailed))
		if rep.GithubPending != 0 {
			lines = append(lines, fmt.Sprintf(
				"| ⏳ GitHub pending (unfinished) | %d |", rep.GithubPending))
		}
		if rep.NonGithubRecorded != 0 {
			lines = append(lines, fmt.Sprintf(
				"| ✅ Non-GitHub recorded (inbox) | %d |", rep.NonGithubRecorded))
		}
		if rep.WebsitesProcessed != 0 || rep.WebsitesReview != 0 ||
			rep.WebsitesSkipped != 0 {
			lines = append(lines, fmt.Sprintf(
				"| 🌐 Websites notes | %d |", rep.WebsitesProcessed))
			lines = append(lines, fmt.Sprintf(
				"| 🗂️ Websites in _review (retry scheduled) | %d |", rep.WebsitesReview))
			lines = append(lines, fmt.Sprintf(
				"| ⏭️ Websites skipped (dedup) | %d |", rep.WebsitesSkipped))
		}
		if rep.BlockedRecorded != 0 {
			lines = append(lines, fmt.Sprintf(
				"| 🚫 Blocked/self domains (omitted by design — never "+
					"collected) | %d |", rep.BlockedRecorded))
		}
		if rep.NonGithubPending != 0 {
			lines = append(lines, fmt.Sprintf(
				"| ⏳ Non-GitHub pending (websites pipeline off / no "+
					"vault) | %d |", rep.NonGithubPending))
		}
		if rep.NonGithubFailed != 0 {
			lines = append(lines, fmt.Sprintf(
				"| ❌ Non-GitHub failed | %d |", rep.NonGithubFailed))
		}
		accounting := fmt.Sprintf("%d/%d accounted for", rep.Accounted, rep.Total)
		if !rep.AccountingOK {
			accounting += fmt.Sprintf(" — %d unaccounted!", rep.Unaccounted)
		} else {
			accounting += " ✓"
		}
		lines = append(lines, fmt.Sprintf("| 🧮 Accounting | %s |", accounting))
		verdict := "❌ SOME LINKS NEED RETRY"
		if rep.VerificationPassed && rep.AccountingOK {
			verdict = "✅ ALL LINKS VERIFIED — NO DATA LOSS!"
		}
		lines = append(lines, fmt.Sprintf("| 🎯 Overall verdict | %s |", verdict))
		lines = append(lines, "")
	}

	// By category.
	if len(categories) > 0 {
		type catCount struct {
			name  string
			count int
		}
		var cats []catCount
		for name, count := range categories {
			cats = append(cats, catCount{name, count})
		}
		// sorted(key=lambda x: -x[1]) — the reference's count-desc order
		// (ties keep the dict's insertion order: first-seen category).
		stable := make([]catCount, len(cats))
		copy(stable, cats)
		sort.SliceStable(stable, func(i, j int) bool {
			return stable[i].count > stable[j].count
		})
		lines = append(lines, "## 📁 Repos by Category")
		lines = append(lines, "")
		lines = append(lines, "| Category | Count |")
		lines = append(lines, "|----------|-------|")
		for _, c := range stable {
			lines = append(lines, fmt.Sprintf("| %s | %d |", c.name, c.count))
		}
		lines = append(lines, "")
	}

	// All processed repos.
	if len(processed) > 0 {
		lines = append(lines, "## 📋 All Processed Repos")
		lines = append(lines, "")
		lines = append(lines, "| # | Repo | Category | Credibility | Banner |")
		lines = append(lines, "|---|------|----------|-------------|--------|")
		for i, p := range processed {
			banner := "—"
			if p.Banner {
				banner = "🖼️"
			}
			lines = append(lines, fmt.Sprintf("| %d | %s | %s | %s/100 | %s |",
				i+1, p.Repo, p.Category, pyFloatStr(p.Credibility), banner))
		}
		lines = append(lines, "")
	}

	// Failed links (from LinkTracker).
	if linkTrackerReport != nil {
		var failed []*batch.ManifestLink
		for _, l := range linkTrackerReport.FailedLinks {
			if l != nil {
				failed = append(failed, l)
			}
		}
		if len(failed) > 0 {
			lines = append(lines, "## ❌ Failed Links (Will Retry)")
			lines = append(lines, "")
			for _, fl := range failed {
				errMsg := "unknown error"
				if fl.Error != nil {
					errMsg = *fl.Error
				}
				url := fl.URL
				if url == "" {
					url = "?"
				}
				lines = append(lines, fmt.Sprintf("- `%s` — %s", url, errMsg))
			}
			lines = append(lines, "")
			lines = append(lines, "> Failed links are kept in the manifest and "+
				"surfaced for retry on the next app launch "+
				"(Dashboard → 🔍 Verify Vault).")
			lines = append(lines, "")
		}
	}

	// Non-GitHub links (brief summary).
	if len(r.NonGithubURLs) > 0 {
		if pipelinesWebsitesOn(r.Config) {
			lines = append(lines, "## 🌐 Non-GitHub Links (Websites pipeline)")
			lines = append(lines, "")
			lines = append(lines, fmt.Sprintf("%d link(s) were processed "+
				"by the Websites pipeline (see its section below).",
				len(r.NonGithubURLs)))
			lines = append(lines, "")
		} else {
			// v0.35.0 — banned links are never collected: the platform
			// table counts only the recordable ones, and the omitted
			// count is its own line.
			law := links.BlockedDomainsFromConfig(r.Config)
			platformCounts := map[string]int{}
			var platformOrder []string
			omitted := 0
			for _, u := range r.NonGithubURLs {
				if links.DomainIsBlocked(u, law) {
					omitted++
					continue
				}
				p := links.ClassifyPlatform(u)
				if p == "github" {
					p = "other"
				}
				if _, ok := platformCounts[p]; !ok {
					platformOrder = append(platformOrder, p)
				}
				platformCounts[p]++
			}
			lines = append(lines, "## 📥 Non-GitHub Links (recorded in _inbox/)")
			lines = append(lines, "")
			lines = append(lines, "| Platform | Count |")
			lines = append(lines, "|----------|-------|")
			type pc struct {
				name string
				n    int
			}
			var pcs []pc
			for _, name := range platformOrder {
				pcs = append(pcs, pc{name, platformCounts[name]})
			}
			sort.SliceStable(pcs, func(i, j int) bool { return pcs[i].n > pcs[j].n })
			for _, c := range pcs {
				display := platformDisplayName(c.name)
				lines = append(lines, fmt.Sprintf("| %s | %d |", display, c.n))
			}
			lines = append(lines, "")
			if omitted > 0 {
				lines = append(lines, fmt.Sprintf(
					"🚫 %d banned-domain link(s) omitted — never collected "+
						"(no note, no _review, no _inbox row).", omitted))
				lines = append(lines, "")
			}
		}
	}

	// The websites pipeline's own section (the phase's summary: the
	// vault, the six outcome rows — each only when nonzero — and the
	// per-link results table).
	if r.websites != nil {
		c := r.websites.Counters
		lines = append(lines, "## 🌐 Websites Pipeline")
		lines = append(lines, "")
		lines = append(lines, "Vault: `"+r.websites.Vault+"`")
		lines = append(lines, "")
		lines = append(lines, "| Outcome | Count |")
		lines = append(lines, "|---------|-------|")
		type outcomeRow struct {
			key, label string
		}
		for _, row := range []outcomeRow{
			{"processed", "processed"},
			{"review", "needs review (_review)"},
			{"upgraded", "upgraded from _review"},
			{"retried", "fetch retries attempted"},
			{"skipped", "skipped (already known)"},
			{"failed", "failed"},
		} {
			if c[row.key] != 0 {
				lines = append(lines, fmt.Sprintf("| %s | %d |",
					row.label, c[row.key]))
			}
		}
		lines = append(lines, "")
		if len(r.websites.Results) > 0 {
			lines = append(lines, "| Link | Outcome | Filed under |")
			lines = append(lines, "|------|---------|-------------|")
			for _, res := range r.websites.Results {
				filed := res.Category
				if res.Subcategory != "" {
					filed += " / " + res.Subcategory
				}
				if filed == "" {
					filed = res.Error
				}
				url := res.URL
				if url == "" {
					url = "?"
				}
				outcome := res.Outcome
				if outcome == "" {
					outcome = "?"
				}
				lines = append(lines, fmt.Sprintf(
					"| [%s](%s) | %s | %s |", url, res.URL, outcome, filed))
			}
			lines = append(lines, "")
		}
	}

	lines = append(lines, "---")
	lines = append(lines, fmt.Sprintf("*Report generated: %s*",
		r.clockNow().Format("2006-01-02 15:04:05")))

	if err := dryrun.WriteText(reportPath, strings.Join(lines, "\n")); err != nil {
		r.log(fmt.Sprintf("⚠️ Failed to write final report: %s", err), "warning")
		return ""
	}
	return reportPath
}

// pyFloatStr renders a float the Python f-string way (str(90.0) →
// "90.0"; the summary log's credibility rows).
func pyFloatStr(f float64) string { return helpers.PyFloatRepr(f) }

// platformDisplayName mirrors _PLATFORM_DISPLAY (the report's platform
// table rows).
func platformDisplayName(key string) string {
	switch key {
	case "x_twitter":
		return "🐦 X / Twitter"
	case "reddit":
		return "👽 Reddit"
	case "youtube":
		return "📺 YouTube"
	case "linkedin":
		return "💼 LinkedIn"
	case "medium":
		return "✍️ Medium"
	case "huggingface":
		return "🤗 HuggingFace"
	case "arxiv":
		return "📄 arXiv"
	case "package_registry":
		return "📦 Package Registry"
	default:
		return "🔗 Other"
	}
}

// pyBannerYesNo renders the Yes/No banner column.
func pyBannerYesNo(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// pipelinesWebsitesOn mirrors config['pipelines']['websites'].
func pipelinesWebsitesOn(config map[string]any) bool {
	pipes, _ := config["pipelines"].(map[string]any)
	if pipes == nil {
		return false
	}
	v, ok := pipes["websites"]
	if !ok || v == nil {
		return false
	}
	if b, ok := v.(bool); ok {
		return b
	}
	return false
}
