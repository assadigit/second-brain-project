// Package repobanish ports core/repo_banish.py — the GitHub Repos
// vault's banishment gate.
//
// Spec: docs/port-blueprint/module-specs/spec-core-repo-banish.md;
// task T-46 (Phase 4). T-45 (internal/vaultscan) landed the surface
// its vault-kind door consumes — the gate's eyes
// (ScanPendingRepoBanishments) and the burial (BanishMarkedRepos), the
// exact functions vault_scan.py's repos door calls — ported first
// because vaultscan's own acceptance (the ported test_reposcan
// selections) exercises them end to end. The module's remaining doors
// (revive_banished_repos, run_repo_banishment_pass, the ♻️ revival and
// the confirmation-gate orchestration) are T-46's to complete.
//
// The reference imports website_pipeline lazily and never re-detects:
// ONE grammar (the websites twin's own doors), an optional skip. The Go
// port follows the same shape.
package repobanish

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/assadigit/second-brain-project/internal/constants"
	"github.com/assadigit/second-brain-project/internal/dryrun"
	"github.com/assadigit/second-brain-project/internal/links"
	"github.com/assadigit/second-brain-project/internal/notestate"
	"github.com/assadigit/second-brain-project/internal/store"
	"github.com/assadigit/second-brain-project/internal/websitepipeline"
	"github.com/assadigit/second-brain-project/internal/websitepipeline/decomtable"
)

// RepoBanishTallyPrefix mirrors REPO_BANISH_TALLY_PREFIX: the log
// prefix the repos run's tally line carries.
const RepoBanishTallyPrefix = "🗑️ Repo run tally:"

// GitHubURLMark mirrors GITHUB_URL_MARK: the substring every repo
// source must carry to be this door's business (defensive: the GitHub
// vault holds repo notes, but a stray source is not a repo and never a
// repos banishment).
const GitHubURLMark = "github.com/"

// clockNow is the package's clock seam: tests freeze it (the house
// pattern — websitepipeline.clockNow's twin; the confirmed-record rows
// carry today's date).
var clockNow = time.Now

type LogFn = func(msg, level string)

func noopLog(string, string) {}

// IsRepoURL mirrors _is_repo_url: is this source URL a GitHub repo
// link (the only kind the repos banishment ever touches)? Cheap
// substring law — the same test the processing loop's own grammar
// uses.
func IsRepoURL(url string) bool {
	return strings.Contains(url, GitHubURLMark)
}

// PendingRepoBanishment mirrors scan_pending_repo_banishments' items:
// {'url', 'canonical', 'title', 'marker', 'door', 'path'}.
type PendingRepoBanishment struct {
	URL       string
	Canonical string
	Title     string
	Marker    string
	Door      string
	Path      string
}

// PendingRepoBanishmentsReport mirrors the gate's dict:
// {'items': [...], 'kept_handwritten': int}.
type PendingRepoBanishmentsReport struct {
	Items           []PendingRepoBanishment
	KeptHandwritten int
}

// ScanPendingRepoBanishments mirrors scan_pending_repo_banishments
// (v0.65.0 — THE REPOS GATE'S EYES): what WOULD be banished this run —
// both doors, one list, deduped by canonical URL (the repos grammar:
// links.NormalizeURL — trailing slash, query, fragment gone). A URL
// already resting in the dismissed list (an earlier banishment — the
// burial's own DB gate) is never re-asked. Pure file reads + one
// optional DB read; no network; never raises.
func ScanPendingRepoBanishments(vaultPath string, noteStateDB *notestate.DB,
	log LogFn) PendingRepoBanishmentsReport {
	if log == nil {
		log = noopLog
	}
	out := PendingRepoBanishmentsReport{}
	if vaultPath == "" || !isDir(vaultPath) {
		return out
	}
	dismissed := map[string]bool{}
	if noteStateDB != nil {
		if set, err := noteStateDB.DismissedSet(notestate.VaultGithub); err == nil && set != nil {
			dismissed = set
		}
	}
	seen := map[string]bool{}
	for _, it := range websitepipeline.ScanBanishedNotes(vaultPath, log, nil) {
		if !it.AppOwned {
			out.KeptHandwritten++
			continue
		}
		if !IsRepoURL(it.URL) {
			continue // a non-repo source in the GitHub vault — not this
			// door's business (defensive; there are none)
		}
		canonical := links.NormalizeURL(it.URL)
		if canonical == "" || seen[canonical] {
			continue
		}
		if dismissed[canonical] {
			continue // an earlier banishment's own gate — never re-asked
		}
		seen[canonical] = true
		title := canonical
		if it.Path != "" {
			base := filepath.Base(it.Path)
			title = strings.TrimSuffix(base, filepath.Ext(base))
		}
		if title == "" {
			title = canonical
		}
		marker := it.Marker
		if marker == "" {
			marker = "delete mark"
		}
		door := it.Door
		if door == "" {
			door = "note tag"
		}
		out.Items = append(out.Items, PendingRepoBanishment{
			URL: it.URL, Canonical: canonical, Title: title,
			Marker: marker, Door: door, Path: it.Path,
		})
	}
	// the master-table door: the vault's own DECOMMISSIONED.md, the
	// same grammar the websites twin reads (a 🗑 Status on a repo row
	// is a pending ask until it is confirmed-stamped).
	table := websitepipeline.ScanDecommissionTable(vaultPath)
	for _, u := range sortedStringKeys(table) {
		status := table[u]
		if !statusBanished(status) {
			continue
		}
		if strings.Contains(strings.ToLower(status), "confirmed") {
			continue // history, not a pending ask
		}
		if !IsRepoURL(u) {
			continue // a websites row in this vault's table — never the
			// repos burial's call
		}
		canonical := links.NormalizeURL(u)
		if canonical == "" || seen[canonical] {
			continue
		}
		if dismissed[canonical] {
			continue
		}
		seen[canonical] = true
		title := canonical
		if i := strings.Index(canonical, GitHubURLMark); i >= 0 {
			owner := canonical[i+len(GitHubURLMark):]
			owner = strings.TrimRight(owner, "/")
			if owner != "" {
				title = owner
			}
		}
		out.Items = append(out.Items, PendingRepoBanishment{
			URL: u, Canonical: canonical, Title: title,
			Marker: status, Door: "master-table gesture", Path: "",
		})
	}
	return out
}

// statusBanished mirrors _status_banished (the websites twin's own
// website_pipeline._status_is_banished, imported once so the grammar
// never forks).
func statusBanished(status string) bool {
	return decomtable.StatusIsBanished(status)
}

// findRepoNoteForURL mirrors _find_repo_note_for_url (v0.65.0): the
// note FILE for one repo URL (the table-gesture burial needs it: the
// row names the repo, the note holds the file). The processed-repos
// ledger's row first — verified on disk (the app's own note:
// managed_by gitcurator + the same source — a row pointing at a
// stranger or at nothing is not a delivery); then a walk of the
// library for any app-owned note whose source normalizes to the
// canonical (the row was lost, the note was moved by hand). ” when
// nothing is found — the burial proceeds DB-only. Never raises.
func findRepoNoteForURL(vaultPath, url string, cache *store.CacheDB,
	log LogFn) string {
	if log == nil {
		log = noopLog
	}
	canonical := links.NormalizeURL(url)
	if canonical == "" {
		return ""
	}
	if cache != nil {
		if row := cache.ProcessedRowFor(canonical); row != nil && row.NotePath != nil {
			path := *row.NotePath
			if path != "" && isFile(path) {
				if fm := websitepipeline.ParseReviewFrontmatter(path); fm != nil {
					if strings.ToLower(fm.ManagedBy) ==
						constants.ManagedByGitcurator &&
						links.NormalizeURL(fm.Source) == canonical {
						return path
					}
				}
			}
		}
	}
	if vaultPath == "" || !isDir(vaultPath) {
		return ""
	}
	found := ""
	walkRepoVault(vaultPath, func(path string) {
		if found != "" {
			return
		}
		fm := websitepipeline.ParseReviewFrontmatter(path)
		if fm == nil ||
			strings.ToLower(fm.ManagedBy) != constants.ManagedByGitcurator {
			return
		}
		if links.NormalizeURL(fm.Source) == canonical {
			found = path
		}
	})
	return found
}

// walkRepoVault mirrors the reference's pruned os.walk: dirs sorted,
// '_inbox' and dot-folders skipped, files sorted, files of a directory
// handled BEFORE its subdirectories (the house walk law — the same
// shape websitepipeline's own walk keeps).
func walkRepoVault(vaultPath string, visit func(path string)) {
	walkRepoDir(vaultPath, visit)
}

func walkRepoDir(dir string, visit func(path string)) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return // never raises — the walk's own law
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names)
	for _, name := range names {
		if !strings.HasSuffix(strings.ToLower(name), ".md") {
			continue
		}
		path := filepath.Join(dir, name)
		if !isFile(path) {
			continue
		}
		visit(path)
	}
	subdirs := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if e.Name() == "_inbox" || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		subdirs = append(subdirs, e.Name())
	}
	sort.Strings(subdirs)
	for _, sub := range subdirs {
		walkRepoDir(filepath.Join(dir, sub), visit)
	}
}

// banishRepoDBReport mirrors _banish_repo_db's dict.
type banishRepoDBReport struct {
	Dismissed     bool
	Settled       bool
	RetryResolved bool
	RowForgotten  bool
}

// banishRepoDB mirrors _banish_repo_db (v0.65.0 — THE BURIAL CORE, DB
// first): the never-fetch gates BEFORE any file moves — the URL is
// dismissed in note_state and settled in the repos ledger, its
// failed_repos row resolved, its processed_repos row forgotten. A
// nil noteStateDB (a dry-run) skips only the dismissal WRITE.
func banishRepoDB(cache *store.CacheDB, noteStateDB *notestate.DB,
	canonical, notePath string, log LogFn) banishRepoDBReport {
	if log == nil {
		log = noopLog
	}
	report := banishRepoDBReport{}
	if canonical == "" {
		return report
	}
	if noteStateDB != nil && !dryrun.IsEnabled() {
		if err := noteStateDB.Dismiss(notestate.VaultGithub, canonical,
			notePath); err == nil {
			report.Dismissed = true
		} else {
			log(fmt.Sprintf("⚠️ Repo banishment note-state write skipped for %s: %v",
				canonical, err), "warning")
		}
	} else if dryrun.IsEnabled() {
		log(fmt.Sprintf("🗑️ %s: the dismissal would be recorded (dry-run "+
			"rehearsal — note_state is never written in a rehearsal)",
			canonical), "info")
	}
	if cache != nil {
		// the banishment is the STRONGER verdict — a leftover 404
		// quarantine row would hand the "Reset 404 Quarantine" ♻️ door
		// the power to silently un-do this burial; the row goes first.
		cache.ResetDeadLinks(canonical)
		if n, err := cache.SettleRepos([]string{canonical}); err == nil {
			report.Settled = n > 0
		} else {
			log(fmt.Sprintf("⚠️ Repo banishment settlement skipped for %s: %v",
				canonical, err), "warning")
		}
		cache.ResolveFailedURLs([]string{canonical})
		report.RetryResolved = true
		report.RowForgotten = cache.ForgetProcessedURL(canonical)
	}
	return report
}

// BanishMarkedReposReport mirrors banish_marked_repos' dict:
// {'marked', 'banished', 'notes_moved', 'kept_handwritten',
// 'rows_written', 'urls'} — urls is the canonical list the run's 🗑️
// tally counts.
type BanishMarkedReposReport struct {
	Marked          int
	Banished        int
	NotesMoved      int
	KeptHandwritten int
	RowsWritten     int
	URLs            []string
}

// BanishMarkedRepos mirrors banish_marked_repos (v0.65.0 — THE REPOS
// BURIAL): enforce the owner's delete verdicts on the GitHub vault's
// repo notes. Every picked item leaves the library for
// .trash/banished and its URL is banished through every never-fetch
// gate the repos side owns (dismissed + settled + retry resolved + row
// forgotten). The vault gains the record row (🗑️ banished — confirmed
// <date>, ♻️ revivable). Idempotent. Dry-run aware; never raises.
func BanishMarkedRepos(cache *store.CacheDB, noteStateDB *notestate.DB,
	vaultPath string, log LogFn,
	items []PendingRepoBanishment) BanishMarkedReposReport {
	if log == nil {
		log = noopLog
	}
	report := BanishMarkedReposReport{}
	if vaultPath == "" || !isDir(vaultPath) {
		return report
	}
	if items == nil {
		gate := ScanPendingRepoBanishments(vaultPath, noteStateDB, log)
		items = gate.Items
		report.KeptHandwritten = gate.KeptHandwritten
	}
	if len(items) == 0 {
		warnHandwritten(report.KeptHandwritten, &report, log)
		return report
	}
	dateStr := clockNow().Format("2006-01-02")
	seen := map[string]bool{}
	var recordURLsTag []string
	recordNotesTag := map[string]string{}
	var recordRowsTrash []string
	recordNotesTrash := map[string]string{}
	var recordRowsTable []string
	recordNotesTable := map[string]string{}
	for _, it := range items {
		report.Marked++
		canonical := links.NormalizeURL(it.URL)
		if canonical == "" {
			canonical = links.NormalizeURL(it.Canonical)
		}
		if canonical == "" || seen[canonical] {
			continue // two marked notes, one repo — one burial
		}
		seen[canonical] = true
		door := it.Door
		if door == "" {
			door = "note tag"
		}
		viaTrash := door == "trash folder"
		viaTable := door == "master-table gesture"
		notePath := it.Path
		if viaTable && notePath == "" {
			notePath = findRepoNoteForURL(vaultPath, canonical, cache, log)
		}
		banishRepoDB(cache, noteStateDB, canonical, notePath, log)
		report.Banished++
		if notePath != "" && isFile(notePath) {
			if dst := websitepipeline.BanishNoteFile(vaultPath, notePath, log); dst != "" {
				report.NotesMoved++
			}
		} else if !dryrun.IsEnabled() {
			log(fmt.Sprintf("🗑️ %s: banished — the note file was already "+
				"gone (a hand deletion, an earlier banishment); the repo is "+
				"dismissed, settled and row-forgotten — never fetched or "+
				"counted again", canonical), "info")
		}
		switch {
		case viaTrash:
			recordRowsTrash = append(recordRowsTrash, canonical)
			recordNotesTrash[canonical] =
				"🗑️ you moved it to the Trash folder (repo vault)"
		case viaTable:
			recordRowsTable = append(recordRowsTable, canonical)
			marker := it.Marker
			if marker == "" {
				marker = "delete"
			}
			recordNotesTable[canonical] =
				fmt.Sprintf("🗑️ marked in this table: %s", marker)
		default:
			recordURLsTag = append(recordURLsTag, canonical)
			marker := it.Marker
			if marker == "" {
				marker = "delete"
			}
			recordNotesTag[canonical] =
				fmt.Sprintf("🗑️ marked on the repo note: %s", marker)
		}
		if !dryrun.IsEnabled() {
			log(fmt.Sprintf("🗑️ %s: banished — the repo note left the library "+
				"for %s and the repo is dismissed + settled — never fetched, "+
				"never counted, never retried again (♻️ in this table revives "+
				"it)", canonical, websitepipeline.BanishQuarantineRelpath), "info")
		}
	}
	if len(seen) > 0 {
		written := 0
		if len(recordURLsTag) > 0 {
			written += websitepipeline.WriteDecommissionCandidates(vaultPath,
				recordURLsTag, "repo note tag", log, recordNotesTag,
				fmt.Sprintf("🗑️ banished — confirmed %s", dateStr))
		}
		if len(recordRowsTrash) > 0 {
			written += websitepipeline.WriteDecommissionCandidates(vaultPath,
				recordRowsTrash, "repo Trash folder", log, recordNotesTrash,
				fmt.Sprintf("🗑️ banished — confirmed %s", dateStr))
		}
		if len(recordRowsTable) > 0 {
			written += websitepipeline.WriteDecommissionCandidates(vaultPath,
				recordRowsTable, "repo table gesture", log, recordNotesTable,
				fmt.Sprintf("🗑️ banished — confirmed %s", dateStr))
		}
		report.RowsWritten = written
	}
	report.URLs = sortedKeys(seen)
	warnHandwritten(report.KeptHandwritten, &report, log)
	if report.Banished > 0 {
		log(fmt.Sprintf("🗑️ Repos banishment: %d repo note(s) removed from "+
			"the library and banished — dismissed, settled, retry-resolved, "+
			"row-forgotten (never fetched, never counted, never retried "+
			"again); %d hand-written marked note(s) kept (yours); the "+
			"vault's table holds the record rows (♻️ revivable)",
			report.Banished, report.KeptHandwritten), "info")
	}
	return report
}

// warnHandwritten mirrors _warn_handwritten: the sacred law's own
// line — marked notes that are NOT the app's are counted and kept,
// never deleted by this door.
func warnHandwritten(kept int, report *BanishMarkedReposReport, log LogFn) {
	if kept > 0 {
		log(fmt.Sprintf("✍️ kept %d marked note(s) — hand-written (yours); "+
			"the app never deletes what it did not write — delete them by "+
			"hand in Obsidian if you want them gone", kept), "warning")
	}
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func sortedStringKeys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func isDir(p string) bool {
	st, err := os.Stat(p)
	return err == nil && st.IsDir()
}

func isFile(p string) bool {
	st, err := os.Stat(p)
	return err == nil && !st.IsDir()
}
