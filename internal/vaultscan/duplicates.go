package vaultscan

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/assadigit/second-brain-project/internal/helpers"
	"github.com/assadigit/second-brain-project/internal/links"
	"github.com/assadigit/second-brain-project/internal/websitepipeline"
)

// nonAlnumRE mirrors the fold laws' character class: [^a-z0-9]+.
var nonAlnumRE = regexp.MustCompile(`[^a-z0-9]+`)

// KeepRow mirrors a consolidation proposal's keep cell.
type KeepRow struct {
	File        string
	Path        string
	Source      string
	Title       string
	Category    string
	Subcategory string
}

// MergeRow mirrors a consolidation proposal's merge rows.
type MergeRow struct {
	File   string
	Path   string
	Source string
	Title  string
}

// Consolidation mirrors scan_site_duplicates' proposals:
// {'site', 'keep', 'merge', 'urls'}.
type Consolidation struct {
	Site  string
	Keep  KeepRow
	Merge []MergeRow
	URLs  []string
}

// ScanSiteDuplicates mirrors scan_site_duplicates (v0.64.0 — ONE NOTE
// PER SITE, the scan's own eyes: pure frontmatter reads, no LLM
// anywhere). Every app-owned note with an http source groups by its
// site key (the host; the repos kind groups by the canonical repo
// URL); a group holding a REAL note plus any other note of the same
// site becomes ONE proposal. Hand notes never ride; core notes never
// ride; locked notes never ride EITHER side. Placeholders alone
// produce nothing. Returns the proposals in the walk's order; never
// raises.
func ScanSiteDuplicates(inv *Inventory, log LogFn, kind string) []Consolidation {
	if log == nil {
		log = noopLog
	}
	var out []Consolidation
	func() {
		defer func() { recover() }() // the reference's outer try/except
		groups := map[string][]NoteRow{}
		var order []string
		for _, n := range inv.Notes {
			if !n.AppOwned || n.Core {
				continue
			}
			if n.Locked {
				// THE LOCK LAW: a locked note is invisible to the
				// magnet — never merged, never keeper, never even a
				// group member.
				continue
			}
			src := helpers.Strip(n.Source)
			if !strings.HasPrefix(strings.ToLower(src), "http://") &&
				!strings.HasPrefix(strings.ToLower(src), "https://") {
				continue
			}
			site := wpSiteKey(src)
			if isReposKind(kind) {
				site = repoURLKey(src)
			}
			if site == "" {
				continue
			}
			if _, ok := groups[site]; !ok {
				order = append(order, site)
			}
			groups[site] = append(groups[site], n)
		}
		for _, site := range order {
			notes := groups[site]
			if len(notes) < 2 {
				continue
			}
			var real []NoteRow
			for _, n := range notes {
				fs := n.FetchStatus
				if fs == "" {
					fs = "full"
				}
				if !n.Protected && fs != "failed" {
					real = append(real, n)
				}
			}
			if len(real) == 0 {
				continue // placeholders alone — the review owns them
			}
			// the keeper rank: (len(urlparse(source).path.rstrip('/')),
			// the walk's own deterministic order) — sorted, [0] wins.
			ranked := make([]NoteRow, len(real))
			copy(ranked, real)
			sort.SliceStable(ranked, func(i, j int) bool {
				pi := rankPathLen(ranked[i].Source)
				pj := rankPathLen(ranked[j].Source)
				if pi != pj {
					return pi < pj
				}
				return ranked[i].Rel < ranked[j].Rel
			})
			keep := ranked[0]
			var merge []NoteRow
			for _, n := range notes {
				if n.Path != keep.Path {
					merge = append(merge, n)
				}
			}
			if len(merge) == 0 {
				continue
			}
			keepRow := KeepRow{
				File: keep.File, Path: keep.Path, Source: keep.Source,
				Title:       keep.Title,
				Category:    keep.Category,
				Subcategory: keep.Subcategory,
			}
			if keepRow.Title == "" {
				keepRow.Title = keep.File
			}
			var mergeRows []MergeRow
			var urls []string
			for _, m := range merge {
				title := m.Title
				if title == "" {
					title = m.File
				}
				mergeRows = append(mergeRows, MergeRow{
					File: m.File, Path: m.Path, Source: m.Source,
					Title: title,
				})
				urls = append(urls, m.Source)
			}
			out = append(out, Consolidation{
				Site: site, Keep: keepRow, Merge: mergeRows, URLs: urls})
		}
		if len(out) > 0 {
			what := "site"
			if isReposKind(kind) {
				what = "repo"
			}
			log(fmt.Sprintf("🧲 %d %s(s) hold more than one note — the "+
				"plan consolidates each into its %s's own note (the "+
				"regex half of the scan, no LLM asked)",
				len(out), what, what), "info")
		}
	}()
	return out
}

// rankPathLen mirrors the keeper rank's path length:
// len(urlparse(source).path.rstrip('/')).
func rankPathLen(source string) int {
	return len(strings.TrimRight(pyURLParsePath(source), "/"))
}

// wpSiteKey mirrors _wp_site_key: the site key via the pipeline's own
// law (links.SiteKeyOf).
func wpSiteKey(url string) string {
	return links.SiteKeyOf(url)
}

// LockedCanonicals mirrors _locked_canonicals (v0.73.0 — THE LOCK
// LAW's deletion door): the canonical URLs of every locked note in the
// inventory. Pure; never raises.
func LockedCanonicals(inv *Inventory, kind string) map[string]bool {
	out := map[string]bool{}
	for _, n := range inv.Notes {
		if !n.Locked {
			continue
		}
		src := helpers.Strip(n.Source)
		if !strings.HasPrefix(strings.ToLower(src), "http://") &&
			!strings.HasPrefix(strings.ToLower(src), "https://") {
			continue
		}
		var canon string
		if isReposKind(kind) {
			canon = repoURLKey(src)
		} else {
			canon = links.NormalizeWebsiteURL(src)
		}
		if canon != "" {
			out[canon] = true
		}
	}
	return out
}

// FolderTreeLines mirrors _folder_tree_lines.
func FolderTreeLines(inv *Inventory, tooBroad int) []string {
	var lines []string
	for _, f := range inv.Folders {
		mark := ""
		if f.NoteCount > tooBroad {
			mark = "  [TOO BROAD]"
		}
		lines = append(lines, fmt.Sprintf("%s — %d notes%s",
			f.Rel, f.NoteCount, mark))
	}
	return lines
}

// CandidateReason mirrors _candidate_reason: why one note is a move
// candidate (the digest's honest 'why' cell). Pure; junk-tolerant.
func CandidateReason(n *NoteRow) string {
	folder := n.Folder
	if n.InRoot || folder == "" {
		return "orphaned at the vault root"
	}
	low := strings.ToLower(folder)
	if low == "uncategorized" || strings.HasPrefix(low, "uncategorized/") {
		return "parked in Uncategorized"
	}
	if !strings.Contains(folder, "/") {
		return "a stray at its category root (no subfolder of its own)"
	}
	return fmt.Sprintf("filed in a too-broad folder (%s)", folder)
}

// DigestLines mirrors _digest_lines: the candidate digest the LLM
// reads — the WHY cell, the declared home, up to 8 tags, the repos
// evidence (by owner, ⭐ stars, 🔧 language), and the weighed homes
// cell. Pure; never raises.
func DigestLines(candidates []NoteRow, kind string) []string {
	var lines []string
	repos := isReposKind(kind)
	for _, n := range candidates {
		tags := make([]string, 0, len(n.Tags))
		for _, t := range n.Tags {
			if len(tags) == 8 {
				break
			}
			tags = append(tags, t)
		}
		tagStr := strings.Join(tags, ",")
		var bits []string
		title := n.Title
		if title == "" {
			title = n.File
		}
		bits = append(bits, title)
		if repos {
			if owner := helpers.Strip(n.Org); owner != "" {
				bits = append(bits, "by "+owner)
			}
			if n.Stars != 0 {
				bits = append(bits, fmt.Sprintf("⭐ %s",
					helpers.FormatComma(int64(n.Stars))))
			}
			if lang := helpers.Strip(n.PrimaryLang); lang != "" {
				bits = append(bits, "🔧 "+lang)
			}
		} else if n.Domain != "" {
			bits = append(bits, n.Domain)
		}
		now := n.Folder
		if now == "" {
			now = "(vault root)"
		}
		bits = append(bits, "now: "+now)
		bits = append(bits, "why: "+CandidateReason(&n))
		cat := helpers.Strip(n.Category)
		sub := helpers.Strip(n.Subcategory)
		if cat != "" || sub != "" {
			c, s := cat, sub
			if c == "" {
				c = "?"
			}
			if s == "" {
				s = "?"
			}
			bits = append(bits, "declares: "+c+" / "+s)
		}
		if tagStr != "" {
			bits = append(bits, "tags: "+tagStr)
		}
		if homes := HomesLine(&n); homes != "" {
			bits = append(bits, homes)
		}
		lines = append(lines, " - "+strings.Join(bits, " — "))
	}
	return lines
}

// DiskSpelling mirrors _disk_spelling (v0.82.0 — THE LANDABLE MOVE's
// spelling truth): the folder's OWN spelling on the live tree, cell by
// cell (the exact name first, then its folded identity). Pure reads;
// never raises; ” when a cell resists the walk.
func DiskSpelling(vaultPath, rel string) string {
	var parts []string
	for _, p := range strings.Split(strings.ReplaceAll(rel, "\\", "/"), "/") {
		if p != "" {
			parts = append(parts, p)
		}
	}
	if len(parts) == 0 || vaultPath == "" {
		return ""
	}
	var walked []string
	cur := vaultPath
	for _, seg := range parts {
		entries, err := os.ReadDir(cur)
		if err != nil {
			return ""
		}
		match := ""
		seen := false
		for _, e := range entries {
			if e.Name() == seg {
				seen = true
				break
			}
		}
		if seen {
			match = seg
		} else {
			fold := foldSegment(seg)
			for _, e := range entries {
				if foldSegment(e.Name()) == fold &&
					fsIsDir(filepath.Join(cur, e.Name())) {
					match = e.Name()
					break
				}
			}
		}
		if match == "" || !fsIsDir(filepath.Join(cur, match)) {
			return ""
		}
		walked = append(walked, match)
		cur = filepath.Join(cur, match)
	}
	return strings.Join(walked, "/")
}

// ResolveDiskFolder mirrors _resolve_disk_folder (v0.71.0 — the
// deterministic passes' shared disk verification): does the proposed
// Cat/Sub destination exist on the LIVE tree (spelling folded)? The
// on-disk spelling is canonicalized cell by cell (v0.82.0 — THE DISK
// SPELLING WINS). foldedKnown memoizes the probes.
func ResolveDiskFolder(vaultPath, dest string,
	foldedKnown map[string]*string) *string {
	key := foldPath(dest)
	if hit, ok := foldedKnown[key]; ok {
		return hit
	}
	var hit *string
	segParts := strings.Split(dest, "/")
	probe := filepath.Join(vaultPath, filepath.Join(segParts...))
	if fsIsDir(probe) {
		s := dest
		hit = &s
	} else {
		// folded disk lookup (Windows-shaped frontmatter meets
		// POSIX-shaped folders)
		parent := filepath.Join(vaultPath, filepath.Join(segParts[:len(segParts)-1]...))
		leafFold := foldSegment(segParts[len(segParts)-1])
		entries, err := os.ReadDir(parent)
		if err == nil {
			for _, f := range entries {
				if foldSegment(f.Name()) == leafFold &&
					fsIsDir(filepath.Join(parent, f.Name())) {
					s := strings.Join(append(
						append([]string{}, segParts[:len(segParts)-1]...),
						f.Name()), "/")
					hit = &s
					break
				}
			}
		}
	}
	if hit != nil {
		// THE DISK SPELLING WINS, always: an exact probe hit is
		// canonicalized cell by cell against the live tree.
		if canon := DiskSpelling(vaultPath, *hit); canon != "" {
			hit = &canon
		}
	}
	foldedKnown[key] = hit
	return hit
}

// foldSegment mirrors _fold_segment: one folder cell's
// case/separator-insensitive identity.
func foldSegment(seg string) string {
	return nonAlnumRE.ReplaceAllString(helpers.PyLower(seg), "")
}

// foldPath mirrors _fold: a folder path folded past case, spaces,
// underscores and hyphens, segment by segment — the fresh-eyes law's
// identity ('Design/Color_Tools' and 'design colortools' fold the
// same). Segment-wise (not one flat squash) so 'A/B' never collides
// with 'A_B'.
func foldPath(rel string) string {
	var parts []string
	for _, p := range strings.Split(
		strings.ReplaceAll(rel, "\\", "/"), "/") {
		p = helpers.Strip(p)
		if p != "" {
			parts = append(parts, p)
		}
	}
	folded := make([]string, len(parts))
	for i, p := range parts {
		folded[i] = foldSegment(p)
	}
	return strings.Join(folded, "/")
}

// foldTag mirrors _fold_tag: a tag's matching identity — lowercase,
// alphanumerics only (a hyphen or underscore never splits a word).
func foldTag(tag string) string {
	return nonAlnumRE.ReplaceAllString(helpers.PyLower(tag), "")
}

// trashFolderNamesLower mirrors _wp.TRASH_FOLDER_NAMES lowered for the
// sanitize law.
var trashFolderNamesLower = func() map[string]bool {
	out := map[string]bool{}
	for _, n := range websitepipeline.TrashFolderNames {
		out[strings.ToLower(n)] = true
	}
	return out
}()

// SanitizeFolder mirrors _sanitize_folder: one validated relative
// folder path (forward slashes, no traversal, no dot-folders, every
// cell a safe name, never INTO the parking lot, never Trash). ”
// when the path is illegal — the caller drops the move.
func SanitizeFolder(rel string) string {
	rel = strings.Trim(
		strings.ReplaceAll(helpers.Strip(rel), "\\", "/"), "/")
	if rel == "" || rel == "." {
		return ""
	}
	parts := strings.Split(rel, "/")
	for _, p := range parts {
		p = helpers.Strip(p)
		if p == "" || strings.HasPrefix(p, ".") ||
			inList(p, ProtectedFolders) || strings.Contains(p, "..") ||
			!safeFolderNameRE.MatchString(p) ||
			trashFolderNamesLower[strings.ToLower(p)] {
			return ""
		}
	}
	if strings.ToLower(parts[0]) == "uncategorized" {
		return "" // never propose INTO the parking lot
	}
	return strings.Join(parts, "/")
}

func inList(s string, list []string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
