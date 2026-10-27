package websitepipeline

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/assadigit/second-brain-project/internal/constants"
	"github.com/assadigit/second-brain-project/internal/dryrun"
	"github.com/assadigit/second-brain-project/internal/links"
	"github.com/assadigit/second-brain-project/internal/websitepipeline/decomtable"
)

// review.go — v0.42.0 the _review backlog (scan + retry inputs),
// v0.49.0 the every-app-owned-note scan, v0.56.0 the half-fetched
// sweep. String-scan only — core never grows a YAML dependency.

// ReviewFM mirrors _parse_review_frontmatter's dict: the three keys
// the backlog retry needs (source / fetch_status / managed_by). Nil
// when the file has no frontmatter block or cannot be read.
type ReviewFM struct {
	Source      string
	FetchStatus string
	ManagedBy   string
}

// ParseReviewFrontmatter mirrors _parse_review_frontmatter.
func ParseReviewFrontmatter(path string) *ReviewFM {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	lines := pySplitLines(decodeUTF8Replace(string(data)))
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return nil
	}
	out := &ReviewFM{}
	for _, line := range lines[1:] {
		s := strings.TrimSpace(line)
		if s == "---" {
			break
		}
		if !strings.Contains(s, ":") {
			continue
		}
		key, val, _ := strings.Cut(s, ":")
		key = strings.ToLower(strings.TrimSpace(key))
		val = stripQuotes(strings.TrimSpace(val))
		switch key {
		case "source":
			out.Source = val
		case "fetch_status":
			out.FetchStatus = val
		case "managed_by":
			out.ManagedBy = val
		}
	}
	return out
}

// BacklogItem mirrors scan_review_backlog's items:
// {'url': source, 'path': note_path}.
type BacklogItem struct {
	URL  string
	Path string
}

// ScanReviewBacklog mirrors scan_review_backlog (v0.42.0): the
// APP-OWNED fetch-failed placeholders of <vault>/_review — the links
// the pre-v0.41 honest-bot User-Agent got walled on. The graveyard
// filter (v0.44.0): a URL already marked dead in DECOMMISSIONED.md is
// BURIED, not backlogged. isDismissed (v0.47.0, optional): a link
// RETIRED by the fetcher's own auto-verdict leaves the backlog too.
// Sorted by filename for a deterministic order; one source may appear
// more than once (legacy _v1/_v2 stacking) — the retry driver
// consolidates. Pure reads; no state DB, no network.
func ScanReviewBacklog(vaultPath string, log LogFn, isDismissed func(string) bool) []BacklogItem {
	if log == nil {
		log = noopLog
	}
	reviewDir := filepath.Join(vaultPath, ReviewFolder)
	if vaultPath == "" {
		return nil
	}
	entries, err := os.ReadDir(reviewDir)
	if err != nil {
		// THE MISSING-FOLDER LAW (the owner's SESSION-075 directive —
		// "Could not list G:\...\_review: The system cannot find the
		// file specified" warned on every sync): a vault WITHOUT a
		// _review folder is the HEALTHY normal (nothing waits for human
		// eyes) — the ENOENT case returns silently; every other error
		// (permissions, a broken junction…) keeps its warning line.
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		if log != nil && vaultPath != "" {
			log(fmt.Sprintf("⚠️ Could not list %s: %v", reviewDir, err), "warning")
		}
		return nil
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	sort.Strings(names)
	dead := map[string]bool{}
	tableRows := ScanDecommissionTable(vaultPath)
	for u, st := range tableRows {
		if statusIsDeadLocal(st) {
			dead[links.NormalizeWebsiteURL(u)] = true
		}
	}
	var items []BacklogItem
	for _, name := range names {
		if !strings.HasSuffix(strings.ToLower(name), ".md") || name == DecommissionTable {
			continue
		}
		path := filepath.Join(reviewDir, name)
		if st, err := os.Stat(path); err != nil || st.IsDir() {
			continue
		}
		fm := ParseReviewFrontmatter(path)
		if fm == nil {
			continue
		}
		if !strings.EqualFold(fm.ManagedBy, constants.ManagedByGitcurator) {
			continue // a human's note — never our call
		}
		if !strings.EqualFold(fm.FetchStatus, "failed") {
			continue // in _review for OTHER reasons — human eyes
		}
		u := strings.TrimSpace(fm.Source)
		if !startsWithHTTP(u) {
			continue
		}
		canonical := links.NormalizeWebsiteURL(u)
		if dead[canonical] {
			continue // v0.44.0 — buried in the graveyard, not backlogged
		}
		if isDismissed != nil && safeProbe(isDismissed, canonical) {
			continue // v0.47.0 — retired (auto-verdict/owner), not
			// waiting — the master table is its ledger. safeProbe
			// carries the reference's try/except law (a broken
			// probe never hides a waiting link).
		}
		items = append(items, BacklogItem{URL: u, Path: path})
	}
	return items
}

// safeProbe runs an injected probe with the reference's never-raise
// law (a broken probe never hides a waiting link).
func safeProbe(probe func(string) bool, u string) (hit bool) {
	defer func() { recover() }()
	return probe(u)
}

// IsDismissedFn is the scan's injected probe seam (a WebsiteStateDB
// partial in the reference).
type IsDismissedFn = func(string) bool

// ReviewNoteItem mirrors scan_review_notes' items:
// {'url', 'path', 'fetch_status'}.
type ReviewNoteItem struct {
	URL         string
	Path        string
	FetchStatus string
}

// ScanReviewNotes mirrors scan_review_notes (v0.49.0): EVERY
// app-owned note waiting in _review (ANY fetch_status — the classes
// that wait for HUMAN eyes: low classification confidence, analysis
// failures, archived rescues). A hand-written note (no frontmatter /
// no managed_by key) is a human's — never our call. Sorted by
// filename; pure file reads; no state DB, no network.
func ScanReviewNotes(vaultPath string, log LogFn) []ReviewNoteItem {
	if log == nil {
		log = noopLog
	}
	if vaultPath == "" {
		return nil
	}
	reviewDir := filepath.Join(vaultPath, ReviewFolder)
	entries, err := os.ReadDir(reviewDir)
	if err != nil {
		// THE MISSING-FOLDER LAW (the twin of the backlog scan's own):
		// no _review folder = the healthy normal, ENOENT returns
		// silently; every other error keeps its warning line.
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		log(fmt.Sprintf("⚠️ Could not list %s: %v", reviewDir, err), "warning")
		return nil
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	sort.Strings(names)
	var items []ReviewNoteItem
	for _, name := range names {
		if !strings.HasSuffix(strings.ToLower(name), ".md") || name == DecommissionTable {
			continue
		}
		path := filepath.Join(reviewDir, name)
		if st, err := os.Stat(path); err != nil || st.IsDir() {
			continue
		}
		fm := ParseReviewFrontmatter(path)
		if fm == nil || !strings.EqualFold(fm.ManagedBy, constants.ManagedByGitcurator) {
			continue // a human's note — never our call
		}
		u := strings.TrimSpace(fm.Source)
		if !startsWithHTTP(u) {
			continue
		}
		items = append(items, ReviewNoteItem{
			URL: u, Path: path, FetchStatus: strings.TrimSpace(fm.FetchStatus),
		})
	}
	return items
}

// ReviewNoteReason mirrors _review_note_reason (v0.49.0): the reason
// line out of a _review note's warning callout — build_review_note
// writes exactly one callout followed by one "> <reason>" line; this
// reads that line back so the master table's Notes column can tell
// the owner WHY a link waits. Tolerant pure read: anything unexpected
// returns ”.
func ReviewNoteReason(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	lines := pySplitLines(decodeUTF8Replace(string(data)))
	for i, line := range lines {
		if !strings.Contains(line, "[!warning]") {
			continue
		}
		for j := i + 1; j < len(lines) && j < i+4; j++ {
			s := strings.TrimSpace(lines[j])
			if strings.HasPrefix(s, ">") {
				reason := strings.TrimSpace(strings.TrimLeft(s, ">"))
				reason = strings.TrimSpace(reason)
				if reason != "" {
					return reason
				}
			} else if s != "" {
				break
			}
		}
	}
	return ""
}

// SweepReviewLeftovers mirrors sweep_review_leftovers (v0.56.0 — THE
// HAND'S HARVEST, the sweep): remove the link's HALF-FETCHED items
// from _review once its proper, categorized note has landed. The
// ownership test is the folder's own frontmatter law
// (managed_by: gitcurator); a hand-written note is a human's, KEPT
// with a warning. The master table and the hand-delivered folder are
// never touched. Dry-run aware; never raises. Returns how many notes
// were swept.
func SweepReviewLeftovers(vaultPath, canonical string, keepPath string, log LogFn) int {
	if log == nil {
		log = noopLog
	}
	if vaultPath == "" || canonical == "" {
		return 0
	}
	reviewDir := filepath.Join(vaultPath, ReviewFolder)
	entries, err := os.ReadDir(reviewDir)
	if err != nil {
		return 0 // no folder / unreadable — nothing to sweep
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	sort.Strings(names)
	swept := 0
	for _, name := range names {
		if !strings.HasSuffix(strings.ToLower(name), ".md") || name == DecommissionTable {
			continue
		}
		path := filepath.Join(reviewDir, name)
		if st, err := os.Stat(path); err != nil || st.IsDir() {
			continue // the hand-delivered folder is a directory
		}
		if keepPath != "" && samePath(path, keepPath) {
			continue
		}
		fm := ParseReviewFrontmatter(path)
		if fm == nil || !strings.EqualFold(fm.ManagedBy, constants.ManagedByGitcurator) {
			// a hand-written note that looks like it belongs to the
			// same source is KEPT with a warning (the sacred law).
			source := ""
			if data, err := os.ReadFile(path); err == nil {
				head := decodeUTF8Replace(string(data))
				lines := pySplitLines(head)
				if len(lines) > 12 {
					lines = lines[:12]
				}
				for _, line := range lines {
					if strings.HasPrefix(strings.ToLower(strings.TrimSpace(line)), "source:") {
						source = strings.TrimSpace(line[len("source:"):])
						break
					}
				}
			}
			if source != "" && SameSource(source, canonical) {
				log(fmt.Sprintf("⚠️ kept %s — it looks hand-written for the same source; the proper note and it now coexist (the duplicate detector's call, never ours)", name), "warning")
			}
			continue
		}
		source := strings.TrimSpace(fm.Source)
		if source == "" || !SameSource(source, canonical) {
			continue // another link's note — untouched
		}
		if err := dryrun.Remove(path); err == nil {
			swept++
			log(fmt.Sprintf("♻️ swept the half-fetched _review item %s — the proper, categorized note replaced it", name), "info")
		} else {
			log(fmt.Sprintf("⚠️ could not sweep %s: %v", path, err), "warning")
		}
	}
	return swept
}

// SameSource mirrors _same_source (v0.56.0): does a note's
// frontmatter source point at the same link as canonical? Both sides
// normalized. Pure; anything unreadable is simply not the same link.
func SameSource(source, canonical string) bool {
	return source != "" && canonical != "" &&
		links.NormalizeWebsiteURL(source) == canonical
}

// samePath mirrors the reference's path-identity comparisons: the
// sweep's keep-path check rides os.path.normpath equality; the twins
// sweep's same-vault guard rides os.path.realpath — EvalSymlinks with
// a cleaned fallback covers both on every platform.
func samePath(a, b string) bool {
	real := func(p string) string {
		if resolved, err := filepath.EvalSymlinks(p); err == nil {
			return filepath.Clean(resolved)
		}
		return filepath.Clean(p)
	}
	return real(a) == real(b)
}

// startsWithHTTP mirrors url.lower().startswith(('http://',
// 'https://')).
func startsWithHTTP(u string) bool {
	l := strings.ToLower(u)
	return strings.HasPrefix(l, "http://") || strings.HasPrefix(l, "https://")
}

// statusIsDeadLocal is the scans' decomtable view.
func statusIsDeadLocal(status string) bool {
	return decomtable.StatusIsDead(status)
}
