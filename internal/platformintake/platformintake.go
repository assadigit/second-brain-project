package platformintake

// platformintake.go — gui/platform_intake.py's port: the per-platform
// _inbox table writers (write_inbox_links_by_platform), the prune pass
// (prune_inbox_tables), the table-vault law (_inbox_table_vault), and
// the display-name map (PLATFORM_INFO).
//
// v0.35.0 laws preserved: banned domains are NEVER collected (no note,
// no _review, no _inbox row — the manifest's blocked bucket is the
// count); links already stored as notes are not re-collected; every
// table is PRUNED after writing so legacy rows of both kinds leave.
// The row grammar is byte-exact: `| - | YYYY-MM-DD | <scrubbed url> |
// <netloc> | <source> | unreviewed | |`, the header carries the
// platform display name + the "Last updated:" stamp, and new rows
// insert AFTER the last data row (an append, never a rewrite).
//
// The clock stamps (the row date + the header's "Last updated:") ride
// the Clock seam — goldens freeze it (the house law).
//
// T-62 slice 2's need: _create_inbox_notes (the worker's intake leg
// when the Websites pipeline is OFF) and the websites phase's
// post-batch _inbox prune both land on this module.

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/assadigit/second-brain-project/internal/batch"
	"github.com/assadigit/second-brain-project/internal/dryrun"
	"github.com/assadigit/second-brain-project/internal/links"
	"github.com/assadigit/second-brain-project/internal/storage"
)

// LogFn mirrors log_callback(msg, level).
type LogFn func(msg, level string)

// Clock is the datetime.now() seam (the frozen-clock law).
var Clock = time.Now

// platformDisplay mirrors _PLATFORM_DISPLAY (the presentation names).
var platformDisplay = map[string]string{
	"x_twitter":        "🐦 X / Twitter",
	"reddit":           "👽 Reddit",
	"youtube":          "📺 YouTube",
	"linkedin":         "💼 LinkedIn",
	"medium":           "✍️ Medium",
	"huggingface":      "🤗 HuggingFace",
	"arxiv":            "📄 arXiv",
	"package_registry": "📦 Package Registry",
	"other":            "🔗 Other",
}

// PlatformInfo mirrors PLATFORM_INFO: platform key → (display, file).
var PlatformInfo = func() map[string][2]string {
	out := map[string][2]string{}
	for key, display := range platformDisplay {
		file := "other_links.md"
		if f, ok := links.PlatformTableFiles[key]; ok {
			file = f
		}
		out[key] = [2]string{display, file}
	}
	return out
}()

// InboxTableVault mirrors _inbox_table_vault (v0.20.0): every
// non-GitHub link's table row belongs to the WEBSITES vault when one is
// configured; without one the GitHub vault keeps the legacy fallback so
// links are never silently lost.
func InboxTableVault(config map[string]any) string {
	if config == nil {
		return ""
	}
	webVault := strings.TrimSpace(pyString(config["website_vault_path"]))
	if webVault != "" {
		return webVault
	}
	return pyString(config["vault_path"])
}

// storedProbe builds the url → bool "is a note in the vault" probe (the
// Websites VaultIndex over the table vault, keyed with the website
// normalizer). nil when the vault can't be scanned (the caller then
// skips the stored-check rather than false-positive rows).
func storedProbe(vaultPath string) func(string) bool {
	index := batch.NewVaultIndex(vaultPath,
		func(u string) string { return links.NormalizeWebsiteURL(u) })
	if index == nil {
		return nil
	}
	index.Rebuild(nil)
	return func(url string) bool { return index.HasURL(url) }
}

// InboxRowURL mirrors _inbox_row_url: the URL column of one table row
// (column 4, the writer's dedupe parse). ” when not a data row.
func InboxRowURL(line string) string {
	if !strings.HasPrefix(line, "| ") || !strings.Contains(line, "http") {
		return ""
	}
	parts := strings.Split(line, "|")
	if len(parts) < 4 {
		return ""
	}
	u := strings.TrimSpace(parts[3])
	if strings.HasPrefix(u, "http") {
		return u
	}
	return ""
}

// WriteInboxLinksByPlatform mirrors write_inbox_links_by_platform:
// write non-GitHub links to per-platform files in `<vault>/_inbox/`,
// idempotent by normalized URL, atomic, dry-run aware. Returns the
// total number of NEW rows added across all platforms.
func WriteInboxLinksByPlatform(vaultPath string, nonGithubURLs []string,
	source string, logFn LogFn, blockedDomains []string,
	vaultIndexHas func(string) bool) (totalNew int) {

	if logFn == nil {
		logFn = func(string, string) {}
	}
	defer func() {
		// the reference's whole-body try/except: never raises —
		// the failure is spoken, the count stays 0.
		if r := recover(); r != nil {
			func() {
				defer func() { recover() }()
				logFn(fmt.Sprintf("Failed to update inbox: %v", r), "warning")
			}()
			totalNew = 0
		}
	}()
	totalNew = 0
	if vaultPath == "" || len(nonGithubURLs) == 0 {
		return 0
	}
	inboxFolder := filepath.Join(vaultPath, "_inbox")

	// v0.35.0 — filter 1: banned domains are never collected.
	keptURLs := append([]string(nil), nonGithubURLs...)
	omitted := 0
	if len(blockedDomains) > 0 {
		keptURLs = nil
		for _, url := range nonGithubURLs {
			if links.DomainIsBlocked(url, blockedDomains) {
				omitted++
			} else {
				keptURLs = append(keptURLs, url)
			}
		}
		if omitted > 0 {
			func() {
				defer func() { recover() }()
				logFn(fmt.Sprintf(
					"🚫 %d banned-domain link(s) omitted — never collected "+
						"(no note, no _review, no _inbox row)", omitted), "info")
			}()
		}
	}
	if len(keptURLs) == 0 {
		// Nothing to write — but still prune the legacy tables so old
		// banned/stored rows leave on this pass too.
		PruneInboxTables(vaultPath, blockedDomains, vaultIndexHas, logFn, nil)
		return 0
	}

	// v0.35.0 — filter 2: links already stored as notes are not
	// re-collected (the probe is built lazily, only when a link
	// survived the ban filter).
	alreadyStored := 0
	storedProbeFn := vaultIndexHas
	if storedProbeFn == nil {
		storedProbeFn = storedProbe(vaultPath)
	}
	if storedProbeFn != nil {
		survivors := []string{}
		for _, url := range keptURLs {
			if storedProbeFn(url) {
				alreadyStored++
			} else {
				survivors = append(survivors, url)
			}
		}
		keptURLs = survivors
		if alreadyStored > 0 {
			func() {
				defer func() { recover() }()
				logFn(fmt.Sprintf(
					"🔗 %d link(s) already stored as note(s) in the vault — "+
						"not re-added to _inbox", alreadyStored), "info")
			}()
		}
	}

	// Group URLs by platform (GitHub URLs route to 'other' defensively).
	platformURLs := map[string][]string{}
	var platformOrder []string
	for _, url := range keptURLs {
		platform := links.ClassifyPlatform(url)
		if platform == "github" {
			platform = "other"
		}
		if _, seen := platformURLs[platform]; !seen {
			platformOrder = append(platformOrder, platform)
		}
		platformURLs[platform] = append(platformURLs[platform], url)
	}
	sort.Strings(platformOrder) // Go map ranges are RANDOM — the convergence law

	for _, platform := range platformOrder {
		urls := platformURLs[platform]
		entry, ok := PlatformInfo[platform]
		if !ok {
			entry = [2]string{"🔗 Other", "other_links.md"}
		}
		displayName, filename := entry[0], entry[1]
		tablePath := filepath.Join(inboxFolder, filename)
		// v0.09.5 — Phase 0 (dry-run): recorded, not performed. Only
		// created when there is at least one row to write.
		_ = dryrun.MkdirAll(inboxFolder, true)

		// Read existing URLs for dedupe (and scrub live tokens out of
		// rows written before the fix — atomically, only on change).
		existingURLs := map[string]bool{}
		existingContent := ""
		if data, err := os.ReadFile(tablePath); err == nil {
			existingContent = string(data)
			clean := links.ScrubURLsInText(existingContent)
			if clean != existingContent {
				tmpPath := tablePath + ".tmp"
				if err := os.WriteFile(tmpPath, []byte(clean), 0o644); err == nil {
					_ = os.Rename(tmpPath, tablePath)
					existingContent = clean
					func() {
						defer func() { recover() }()
						logFn(fmt.Sprintf(
							"🔒 %s: scrubbed secret token(s) out of stored URL(s)",
							filename), "info")
					}()
				}
			}
			for _, line := range strings.Split(existingContent, "\n") {
				u := InboxRowURL(line)
				if u != "" {
					existingURLs[links.NormalizeURL(u)] = true
				}
			}
		}

		// Find new URLs.
		var newRows []string
		for _, url := range urls {
			norm := links.NormalizeURL(url)
			if existingURLs[norm] {
				continue
			}
			existingURLs[norm] = true
			domain := netlocOrUnknown(url)
			dateStr := Clock().Format("2006-01-02")
			// v0.21.0 — never store secret query values (the row keeps
			// the URL with the parameter name and a … placeholder).
			newRows = append(newRows, fmt.Sprintf(
				"| - | %s | %s | %s | %s | unreviewed | |",
				dateStr, links.ScrubURLToken(url), domain, source))
		}
		if len(newRows) == 0 {
			continue
		}

		// Build or append the table.
		var content string
		if existingContent == "" {
			var b strings.Builder
			fmt.Fprintf(&b, "# %s — Review Queue\n\n", displayName)
			fmt.Fprintf(&b, "> Auto-updated. DO NOT delete rows — only update the Status column.\n")
			fmt.Fprintf(&b, "> Edit Status to: ✅ reviewed / ❌ ignored\n")
			fmt.Fprintf(&b, "> Last updated: %s\n\n", Clock().Format("2006-01-02 15:04"))
			fmt.Fprintf(&b, "| # | Date | URL | Domain | Source | Status | Notes |\n")
			fmt.Fprintf(&b, "|---|------|-----|--------|--------|--------|-------|\n")
			b.WriteString(strings.Join(newRows, "\n") + "\n")
			content = b.String()
		} else {
			lines := strings.Split(existingContent, "\n")
			lastDataIdx := 0
			for i, line := range lines {
				if strings.HasPrefix(line, "| ") && strings.Contains(line, "http") {
					lastDataIdx = i
				}
			}
			out := append([]string(nil), lines[:lastDataIdx+1]...)
			out = append(out, newRows...)
			out = append(out, lines[lastDataIdx+1:]...)
			// Refresh the "Last updated:" stamp like the reference does.
			stamp := Clock().Format("2006-01-02 15:04")
			for i, ln := range out {
				if strings.Contains(ln, "Last updated:") {
					out[i] = "> Last updated: " + stamp
				}
			}
			content = strings.Join(out, "\n")
		}
		if err := storage.AtomicWriteText(tablePath, content); err == nil {
			totalNew += len(newRows)
			func() {
				defer func() { recover() }()
				logFn(fmt.Sprintf(
					"📥 %s: %d new links → %s",
					displayName, len(newRows), filename), "info")
			}()
		}
	}
	if totalNew > 0 {
		func() {
			defer func() { recover() }()
			logFn(fmt.Sprintf(
				"📥 Total: %d non-GitHub links added to _inbox/ (by platform)",
				totalNew), "info")
		}()
	}
	// v0.35.0 — after every write, prune the tables: legacy banned rows
	// and rows for links that are already stored as notes leave the
	// review queue for good. The probe built for filter 2 is reused —
	// no second vault scan.
	PruneInboxTables(vaultPath, blockedDomains, storedProbeFn, logFn, nil)
	return totalNew
}

// PruneInboxTables mirrors prune_inbox_tables (v0.35.0): drop the rows
// the app must not collect from every `<vault>/_inbox/*.md` platform
// table — banned-domain rows, rows already stored as notes (the probe
// or the stored_urls set), everything else kept VERBATIM. The rewritten
// file stays byte-identical apart from the dropped rows and the
// "Last updated:" stamp. Never raises. Returns rows dropped.
func PruneInboxTables(vaultPath string, blockedDomains []string,
	vaultIndexHas func(string) bool, logFn LogFn,
	storedURLs []string) int {
	if logFn == nil {
		logFn = func(string, string) {}
	}
	droppedTotal := 0
	defer func() {
		if recover() != nil {
			// never raises
		}
	}()
	if vaultPath == "" {
		return 0
	}
	inboxFolder := filepath.Join(vaultPath, "_inbox")
	if fi, err := os.Stat(inboxFolder); err != nil || !fi.IsDir() {
		return 0
	}
	storedProbeFn := vaultIndexHas
	storedSet := map[string]bool{}
	for _, u := range storedURLs {
		storedSet[u] = true
	}
	entries, err := os.ReadDir(inboxFolder)
	if err != nil {
		return 0
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names) // sorted(os.listdir) — the convergence law

	for _, fname := range names {
		tablePath := filepath.Join(inboxFolder, fname)
		data, err := os.ReadFile(tablePath)
		if err != nil {
			continue
		}
		lines := strings.Split(string(data), "\n")
		// Lazy shared probe: only build the vault scan when the first
		// candidate row actually needs the "is it a note?" answer.
		dropped := 0
		keptLines := []string{}
		dirty := false
		for _, line := range lines {
			rowURL := InboxRowURL(line)
			if rowURL == "" {
				keptLines = append(keptLines, line)
				continue
			}
			drop := false
			if len(blockedDomains) > 0 &&
				links.DomainIsBlocked(rowURL, blockedDomains) {
				drop = true
			}
			if !drop {
				if storedSet[links.NormalizeWebsiteURL(rowURL)] {
					drop = true
				}
			}
			if !drop && (vaultIndexHas != nil || storedURLs == nil) {
				if storedProbeFn == nil {
					storedProbeFn = storedProbe(vaultPath)
				}
				if storedProbeFn != nil {
					func() {
						defer func() {
							if recover() != nil {
								drop = false
							}
						}()
						drop = storedProbeFn(rowURL)
					}()
				}
			}
			if drop {
				dropped++
				dirty = true
			} else {
				keptLines = append(keptLines, line)
			}
		}
		if !dirty {
			continue
		}
		// Refresh the "Last updated:" stamp like the writer does.
		now := Clock().Format("2006-01-02 15:04")
		for i, ln := range keptLines {
			if strings.Contains(ln, "Last updated:") {
				keptLines[i] = "> Last updated: " + now
			}
		}
		content := strings.Join(keptLines, "\n")
		if err := storage.AtomicWriteText(tablePath, content); err == nil {
			droppedTotal += dropped
			func() {
				defer func() { recover() }()
				logFn(fmt.Sprintf(
					"🧹 %s: pruned %d row(s) — banned domain(s) and link(s) "+
						"already stored as note(s) in the vault", fname, dropped), "info")
			}()
		}
	}
	return droppedTotal
}

// netlocOrUnknown mirrors urlparse(url).netloc or "unknown".
func netlocOrUnknown(u string) string {
	netloc := pyNetloc(u)
	if netloc == "" {
		return "unknown"
	}
	return netloc
}

// pyNetloc mirrors Python urlparse().netloc for the URL shapes this
// table writes (the links package's own twin, local to keep the
// surface stable).
func pyNetloc(u string) string {
	if i := strings.IndexByte(u, '#'); i >= 0 {
		u = u[:i]
	}
	if i := strings.Index(u, "://"); i >= 0 {
		rest := u[i+3:]
		if j := strings.IndexAny(rest, "/?#"); j >= 0 {
			return rest[:j]
		}
		return rest
	}
	return ""
}

// pyString mirrors str(x or ”) for the config seats this module reads.
func pyString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
