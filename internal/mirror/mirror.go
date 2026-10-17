// mirror.go — mirror.py + mirror_keys.py (v0.91.0): the one-way Library/
// mirror of machine-library notes into the Manual Notes vault.
//
// Guarantees (proven by the ported test_phase5 selections + the golden):
//   - NEVER creates, changes or deletes anything outside the two Library
//     folders (validated components + realpath containment re-verified
//     before every write and delete);
//   - mirror copies carry the mirror_of front-matter marker plus the
//     read-only banner; ONLY marked files under Library/ are touched;
//   - moves propagate (matching by RAW source value);
//   - idempotent (a second run over unchanged sources = zero writes);
//   - dry-run by default (RunMirror(…, apply=false)).
//
// mirror_keys.py exists in the reference only to keep the Python import
// graph acyclic (mirror imports linking for the Phase-6 carry-over, and
// linking needs the mirror_of parsers); the Go port is ONE package, so
// the key + regex + unquote live here (the "byte-for-byte move rule"
// comment in mirror_keys.py is honored in spirit: the names keep the
// reference's shapes).
package mirror

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/assadigit/second-brain-project/internal/dryrun"
	"github.com/assadigit/second-brain-project/internal/helpers"
	"github.com/assadigit/second-brain-project/internal/storage"
)

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

const (
	// LibraryFolder is the manual vault's mirror root ("Library").
	LibraryFolder = "Library"
	// GitHubTree is the GitHub vault's mirror tree name.
	GitHubTree = "GitHub Projects"
	// WebsitesTree is the Websites vault's mirror tree name.
	WebsitesTree = "Websites"
)

// MirrorBanners are the read-only warning callouts (per tree).
var MirrorBanners = map[string]string{
	GitHubTree: "> [!warning] Read-only mirror of a note in the GitHub Projects " +
		"vault — GitCurator keeps this copy in sync and overwrites edits " +
		"made here.",
	WebsitesTree: "> [!warning] Read-only mirror of a note in the Websites vault — " +
		"GitCurator keeps this copy in sync and overwrites edits made here.",
}

// SkippedFolderMarks — the same special folders note_state skips; a note
// inside one is not a library note and is never mirrored (substring match
// on the walked path). v0.20.0: _missing placeholders are NOT library
// notes (unlike VaultIndex, which DOES index them).
var SkippedFolderMarks = []string{
	"_moc", "_inbox", "_missing", "attachments", ".obsidian", ".trash",
}

// MirrorKey marks a file as a machine-written mirror copy (mirror_keys.
// py MIRROR_KEY). Only files carrying it are ever updated or deleted —
// and only under Library/.
const MirrorKey = "mirror_of"

// pyS is Python's str-pattern \s class (Unicode whitespace) — the house
// law (CPython \s is Unicode-aware; Go's is ASCII-only).
const pyS = `[\s\v\x1c-\x1f\x85\p{Z}]`

var (
	// sourceRE mirrors mirror.py _SOURCE_RE (^source:\s*(.+)$, MULTILINE).
	sourceRE = regexp.MustCompile(`(?m)^source:` + pyS + `*(.+)$`)
	// mirrorRE mirrors mirror_keys.py _MIRROR_RE
	// (^mirror_of:\s*(.+)$, MULTILINE).
	mirrorRE = regexp.MustCompile(`(?m)^` + MirrorKey + `:` + pyS + `*(.+)$`)

	// coverLineRE strips `cover: attachments/banners/x.png` front-matter
	// lines (vault-root-relative references that can never resolve in the
	// Manual vault).
	coverLineRE = regexp.MustCompile(
		`(?m)^cover: attachments/banners/\S*[^\r\n]*\r?\n?`)
	// bannerEmbedRE strips `![banner](attachments/banners/x.png)` body
	// embeds (plus the blank line the embed drags along).
	bannerEmbedRE = regexp.MustCompile(
		`(?m)^!\[banner\]\(attachments/banners/[^)\s]*\)[ \t]*\r?\n(\r?\n)?`)
)

// MirrorError is raised when the mirror must refuse to run (safety) —
// never mid-write.
type MirrorError struct{ Msg string }

func (e *MirrorError) Error() string { return e.Msg }

// unquote mirrors mirror_keys.py _unquote: whitespace strip, then the
// quote characters strip from both ends (character-SET, not a pair).
func unquote(value string) string {
	return helpers.StripChars(helpers.Strip(value), "\"'")
}

// ---------------------------------------------------------------------------
// Path safety
// ---------------------------------------------------------------------------

// isInside mirrors mirror.py _is_inside: realpath both paths, then
// commonpath equality (child == parent, or child lies under parent).
func isInside(child, parent string) bool {
	childR, err := pyRealpath(child)
	if err != nil {
		return false
	}
	parentR, err := pyRealpath(parent)
	if err != nil {
		return false
	}
	return childR == parentR || strings.HasPrefix(childR, parentR+string(filepath.Separator))
}

// pyRealpath is os.path.realpath's non-strict semantics: absolutize,
// then resolve symlinks component-by-component; components that do not
// exist ride along lexically (realpath never fails on a missing tail).
// A bounded hop count guards symlink loops (ELOOP's quiet cousin).
func pyRealpath(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	parts := strings.Split(strings.TrimPrefix(abs, string(filepath.Separator)), string(filepath.Separator))
	cur := string(filepath.Separator)
	for _, comp := range parts {
		if comp == "" || comp == "." {
			continue
		}
		next := filepath.Join(cur, comp)
		for hops := 0; ; hops++ {
			if hops > 40 {
				break
			}
			link, err := os.Readlink(next)
			if err != nil {
				break
			}
			if filepath.IsAbs(link) {
				next = filepath.Clean(link)
			} else {
				next = filepath.Clean(filepath.Join(cur, link))
			}
		}
		cur = next
	}
	return cur, nil
}

// SafeJoin mirrors mirror.py _safe_join: join root+rel with every
// path-traversal defense on (absolute rel, backslashes, empty/dot/dotdot
// components, drive letters), then re-verify realpath containment.
// Callers never need to check the result again.
func SafeJoin(root, rel string) (string, error) {
	relNorm := strings.ReplaceAll(rel, "\\", "/")
	if strings.HasPrefix(relNorm, "/") {
		return "", &MirrorError{Msg: fmt.Sprintf(
			"unsafe mirror path %s: absolute paths are not allowed",
			helpers.PyRepr(rel))}
	}
	for _, part := range strings.Split(relNorm, "/") {
		if part == "" || part == "." || part == ".." {
			return "", &MirrorError{Msg: fmt.Sprintf(
				"unsafe mirror path %s: empty or parent-reference component %s",
				helpers.PyRepr(rel), helpers.PyRepr(part))}
		}
		if strings.Contains(part, ":") {
			return "", &MirrorError{Msg: fmt.Sprintf(
				"unsafe mirror path %s: drive-letter-like component %s",
				helpers.PyRepr(rel), helpers.PyRepr(part))}
		}
	}
	target := filepath.Join(root, relNorm)
	if !isInside(target, root) {
		return "", &MirrorError{Msg: fmt.Sprintf(
			"unsafe mirror path %s: resolves outside the Library tree",
			helpers.PyRepr(rel))}
	}
	return target, nil
}

// CheckVaultPaths mirrors mirror.py check_vault_paths: refuse to run
// before anything is read or written — the manual vault must be set; it
// must not overlap either machine vault in EITHER direction; with
// apply=true the manual vault folder must already exist.
func CheckVaultPaths(manualVault, githubVault, websitesVault string, apply bool) error {
	manual := helpers.Strip(manualVault)
	if manual == "" {
		return &MirrorError{Msg: "manual_vault_path is not set — configure it " +
			"(GUI 📁 Vault page, config.json, or the tool's --manual-vault " +
			"flag) first"}
	}
	for _, tc := range []struct {
		Label   string
		Machine string
	}{
		{GitHubTree, helpers.Strip(githubVault)},
		{WebsitesTree, helpers.Strip(websitesVault)},
	} {
		if tc.Machine == "" {
			continue
		}
		if isInside(tc.Machine, manual) {
			return &MirrorError{Msg: fmt.Sprintf(
				"refusing to run: the manual vault contains the %s vault\n"+
					"  manual:  %s\n  machine:  %s",
				tc.Label, manual, tc.Machine)}
		}
		if isInside(manual, tc.Machine) {
			return &MirrorError{Msg: fmt.Sprintf(
				"refusing to run: the manual vault is inside the %s vault\n"+
					"  manual:  %s\n  machine:  %s",
				tc.Label, manual, tc.Machine)}
		}
	}
	if helpers.Strip(githubVault) == "" && helpers.Strip(websitesVault) == "" {
		return &MirrorError{Msg: "nothing to mirror: neither the GitHub vault " +
			"(vault_path) nor the Websites vault (website_vault_path) is set"}
	}
	if fi, err := os.Stat(manual); err == nil && !fi.IsDir() {
		return &MirrorError{Msg: fmt.Sprintf(
			"the manual vault path is not a folder: %s", manual)}
	}
	library := filepath.Join(manual, LibraryFolder)
	if fi, err := os.Stat(library); err == nil && !fi.IsDir() {
		return &MirrorError{Msg: fmt.Sprintf(
			"'%s' exists in the manual vault but is not a folder: %s",
			LibraryFolder, library)}
	}
	if apply {
		if fi, err := os.Stat(manual); err != nil || !fi.IsDir() {
			return &MirrorError{Msg: fmt.Sprintf(
				"the manual vault folder does not exist — create it first (or "+
					"point --manual-vault at a copy): %s", manual)}
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// Scanning
// ---------------------------------------------------------------------------

// SourceNote is one scanned machine-vault note.
type SourceNote struct {
	Path   string // absolute path in the machine vault
	Rel    string // vault-relative, forward slashes
	Source string // RAW source value (unquoted, exactly as written)
	Text   string // full note text, loaded while planning
}

// scanSourceNotes mirrors mirror.py _scan_source_notes: a read-only walk
// of one machine vault. Returns (unique notes sorted by rel, duplicate
// rels in sorted order, no-source count, read-error count). A source
// value seen twice keeps the first file (sorted) — the mirror writes one
// copy per source.
func scanSourceNotes(vaultPath string) ([]SourceNote, []string, int, int) {
	var notes []SourceNote
	var duplicateRels []string
	noSource, errors := 0, 0
	if vaultPath == "" {
		return notes, duplicateRels, noSource, errors
	}
	if fi, err := os.Stat(vaultPath); err != nil || !fi.IsDir() {
		return notes, duplicateRels, noSource, errors
	}
	// os.walk with the substring skip on the walked root path. ReadDir's
	// sorted order is the house's deterministic stand-in — every output
	// of this scan is sorted afterwards, so the walk order never leaks.
	var walk func(dir string)
	walk = func(dir string) {
		for _, mark := range SkippedFolderMarks {
			if strings.Contains(dir, mark) {
				return
			}
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			return
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
				continue
			}
			fpath := filepath.Join(dir, e.Name())
			head, ok := readHead800(fpath)
			if !ok {
				errors++
				continue
			}
			m := sourceRE.FindStringSubmatchIndex(head)
			if m == nil {
				noSource++
				continue
			}
			rel, relErr := filepath.Rel(vaultPath, fpath)
			if relErr != nil {
				continue
			}
			rel = strings.ReplaceAll(rel, "\\", "/")
			notes = append(notes, SourceNote{
				Path:   fpath,
				Rel:    rel,
				Source: unquote(head[m[2]:m[3]]),
			})
		}
		for _, e := range entries {
			if e.IsDir() {
				walk(filepath.Join(dir, e.Name()))
			}
		}
	}
	walk(vaultPath)
	sort.Slice(notes, func(i, j int) bool { return notes[i].Rel < notes[j].Rel })
	seen := map[string]bool{}
	var unique []SourceNote
	for _, note := range notes {
		if seen[note.Source] {
			duplicateRels = append(duplicateRels, note.Rel)
			continue
		}
		seen[note.Source] = true
		unique = append(unique, note)
	}
	return unique, duplicateRels, noSource, errors
}

// readHead800 reads the first 800 CHARACTERS (code points) of a file,
// utf-8 errors='replace', universal newlines translated (the house
// text-mode read law).
func readHead800(path string) (string, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	return helpers.CutRunes(helpers.PyNewlines(helpers.DecodeUTF8Replace(data)), 800), true
}

// readTextFile reads a whole file the reference's text-mode way (utf-8
// errors='replace' + universal newlines).
func readTextFile(path string) (string, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	return helpers.PyNewlines(helpers.DecodeUTF8Replace(data)), true
}

// mirrorIndex is _scan_mirror_tree's result: source value → sorted rel
// paths, plus the unmarked and error counts.
type mirrorIndex struct {
	BySource   map[string][]string
	Unmarked   int
	ReadErrors int
}

// scanMirrorTree mirrors mirror.py _scan_mirror_tree: index the mirror
// files inside one Library tree. Every .md file is read — the mirror
// tree has NO skip rules, because orphan detection must see every
// marker file the app ever wrote there. Files without the marker are
// counted and NEVER touched.
func scanMirrorTree(treeRoot string) mirrorIndex {
	idx := mirrorIndex{BySource: map[string][]string{}}
	if treeRoot == "" {
		return idx
	}
	if fi, err := os.Stat(treeRoot); err != nil || !fi.IsDir() {
		return idx
	}
	var walk func(dir string)
	walk = func(dir string) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
				continue
			}
			fpath := filepath.Join(dir, e.Name())
			head, ok := readHead800(fpath)
			if !ok {
				idx.ReadErrors++
				continue
			}
			rel, relErr := filepath.Rel(treeRoot, fpath)
			if relErr != nil {
				continue
			}
			rel = strings.ReplaceAll(rel, "\\", "/")
			m := mirrorRE.FindStringSubmatchIndex(head)
			value := ""
			if m != nil {
				value = unquote(head[m[2]:m[3]])
			}
			if value == "" {
				idx.Unmarked++
				continue
			}
			idx.BySource[value] = append(idx.BySource[value], rel)
		}
		for _, e := range entries {
			if e.IsDir() {
				walk(filepath.Join(dir, e.Name()))
			}
		}
	}
	walk(treeRoot)
	for _, rels := range idx.BySource {
		sort.Strings(rels)
	}
	return idx
}

// ---------------------------------------------------------------------------
// Mirror-note construction
// ---------------------------------------------------------------------------

// stripBannerRefs drops GitHub-note banner references that cannot
// resolve in the mirror (cover line + body embed).
func stripBannerRefs(text string) string {
	text = coverLineRE.ReplaceAllString(text, "")
	text = bannerEmbedRE.ReplaceAllString(text, "")
	return text
}

// lineEndingOf mirrors _line_ending_of: '\r\n' when the FIRST line
// carries one, else '\n'.
func lineEndingOf(text string) string {
	first := text
	if i := strings.IndexByte(text, '\n'); i >= 0 {
		first = text[:i]
	}
	if strings.Contains(first, "\r\n") {
		return "\r\n"
	}
	return "\n"
}

// BuildMirrorNote turns a machine-vault note into its mirror copy:
// inserts `mirror_of: <raw source>` into the front matter (before the
// closing ---), inserts the read-only banner as the first body line,
// and strips banner image references. Content comparison happens in
// newline-translated space (see the reference docstring).
func BuildMirrorNote(content, rawSource, banner string) string {
	text := stripBannerRefs(content)
	nl := lineEndingOf(text)
	mirrorLine := MirrorKey + ": " + rawSource

	lines := strings.Split(text, nl)
	hasFM := len(lines) > 0 && helpers.Strip(lines[0]) == "---"
	closeIdx := -1
	if hasFM {
		for i := 1; i < len(lines); i++ {
			s := helpers.Strip(lines[i])
			if s == "---" || s == "..." {
				closeIdx = i
				break
			}
		}
	}
	var bodyAt int
	if closeIdx >= 0 {
		// insert at closeIdx; body_at past the (now shifted) closing ---
		lines = insertAt(lines, closeIdx, mirrorLine)
		bodyAt = closeIdx + 2
	} else {
		// no (parsable) front matter — wrap the whole note in a fresh block
		fresh := make([]string, 0, len(lines)+4)
		fresh = append(fresh, "---", mirrorLine, "---", "")
		fresh = append(fresh, lines...)
		lines = fresh
		bodyAt = 4
	}
	// banner as the first body line (skip blank lines that follow the
	// front matter; guarantee exactly one blank before and after it)
	for bodyAt < len(lines) && helpers.Strip(lines[bodyAt]) == "" {
		bodyAt++
	}
	if bodyAt > 0 && helpers.Strip(lines[bodyAt-1]) == "---" {
		lines = insertAt(lines, bodyAt, "")
		bodyAt++
	}
	lines = insertAt(lines, bodyAt, banner)
	lines = insertAt(lines, bodyAt+1, "")
	return strings.Join(lines, nl)
}

// insertAt is Python list.insert for string slices (fresh backing).
func insertAt(lines []string, i int, s string) []string {
	out := make([]string, 0, len(lines)+1)
	out = append(out, lines[:i]...)
	out = append(out, s)
	return append(out, lines[i:]...)
}

// ---------------------------------------------------------------------------
// Plan
// ---------------------------------------------------------------------------

// CreateItem is one planned create: (rel, text, source).
type CreateItem struct {
	Rel    string
	Text   string
	Source string
}

// UpdateItem is one planned update: (rel, text, source).
type UpdateItem struct {
	Rel    string
	Text   string
	Source string
}

// MoveItem is one planned move: (old, new, text, source).
type MoveItem struct {
	Old    string
	New    string
	Text   string
	Source string
}

// ConflictItem is one conflict: (rel, reason).
type ConflictItem struct {
	Rel    string
	Reason string
}

// TreePlan is the action list for one Library tree (mirror.py
// TreePlan — the dataclass field order kept).
type TreePlan struct {
	Label           string
	VaultPath       string
	VaultPresent    bool
	Sources         int
	Keeps           int
	Creates         []CreateItem
	Updates         []UpdateItem
	Moves           []MoveItem
	Deletes         []string
	Conflicts       []ConflictItem
	Duplicates      []string
	SkippedNoSource int
	UnmarkedFiles   int
	ScanErrors      int
}

// Changes is TreePlan.changes.
func (t *TreePlan) Changes() int {
	return len(t.Creates) + len(t.Updates) + len(t.Moves) + len(t.Deletes)
}

// MirrorPlan is the whole run's plan (mirror.py MirrorPlan).
type MirrorPlan struct {
	ManualVault   string
	GithubVault   string
	WebsitesVault string
	Apply         bool
	Trees         map[string]*TreePlan
	Applied       bool
}

// NewMirrorPlan constructs the plan shell (trees filled by RunMirror).
func NewMirrorPlan(manual, github, websites string, apply bool) *MirrorPlan {
	return &MirrorPlan{
		ManualVault:   manual,
		GithubVault:   github,
		WebsitesVault: websites,
		Apply:         apply,
		Trees:         map[string]*TreePlan{},
	}
}

// TotalChanges is MirrorPlan.total_changes.
func (p *MirrorPlan) TotalChanges() int {
	total := 0
	for _, t := range p.Trees {
		total += t.Changes()
	}
	return total
}

// RenderReport mirrors MirrorPlan.render_report byte-for-byte.
func (p *MirrorPlan) RenderReport() string {
	mode := "DRY RUN (nothing written)"
	if p.Applied {
		mode = "APPLIED"
	} else if p.Apply {
		mode = "APPLY (planned)"
	}
	out := []string{
		"GitCurator — Manual Notes Library mirror — " + mode,
		"Manual vault:   " + p.ManualVault,
		"GitHub vault:   " + pyOr(p.GithubVault, "(not set)"),
		"Websites vault: " + pyOr(p.WebsitesVault, "(not set)"),
		"",
	}
	for _, label := range []string{GitHubTree, WebsitesTree} {
		tree, ok := p.Trees[label]
		if !ok {
			continue
		}
		srcState := ""
		switch {
		case tree.VaultPresent:
			srcState = "scanned"
		case tree.VaultPath == "":
			srcState = "not set — skipped"
		default:
			srcState = "folder not found — skipped"
		}
		out = append(out,
			fmt.Sprintf("== %s  →  Library/%s/  (%s) ==", label, label, srcState))
		out = append(out, fmt.Sprintf(
			"  library notes: %d · kept: %d · created: %d · updated: %d "+
				"· moved: %d · deleted: %d",
			tree.Sources, tree.Keeps, len(tree.Creates), len(tree.Updates),
			len(tree.Moves), len(tree.Deletes)))
		var skipped []string
		if tree.SkippedNoSource > 0 {
			skipped = append(skipped, fmt.Sprintf(
				"%d without a source line", tree.SkippedNoSource))
		}
		if len(tree.Duplicates) > 0 {
			skipped = append(skipped, fmt.Sprintf(
				"%d duplicate-source file(s)", len(tree.Duplicates)))
		}
		if tree.UnmarkedFiles > 0 {
			skipped = append(skipped, fmt.Sprintf(
				"%d unmarked file(s) under Library/ (left untouched)",
				tree.UnmarkedFiles))
		}
		if tree.ScanErrors > 0 {
			skipped = append(skipped, fmt.Sprintf(
				"%d unreadable file(s)", tree.ScanErrors))
		}
		if len(skipped) > 0 {
			out = append(out, "  skipped: "+strings.Join(skipped, "; "))
		}
		for _, c := range tree.Creates {
			out = append(out, "  CREATE   "+c.Rel)
		}
		for _, u := range tree.Updates {
			out = append(out, "  UPDATE   "+u.Rel)
		}
		for _, m := range tree.Moves {
			out = append(out, "  MOVE     "+m.Old+"  ->  "+m.New)
		}
		for _, rel := range tree.Deletes {
			out = append(out, "  DELETE   "+rel)
		}
		for _, c := range tree.Conflicts {
			out = append(out, "  CONFLICT "+c.Rel+" — "+c.Reason)
		}
		for _, rel := range tree.Duplicates {
			out = append(out, "  DUP      "+rel)
		}
		out = append(out, "")
	}
	tail := fmt.Sprintf("Totals: %d change(s) needed", p.TotalChanges())
	if p.Applied {
		tail += " — applied."
	} else if !p.Apply {
		tail += " — NOT applied (dry run)."
	} else {
		tail += " — apply was requested."
	}
	out = append(out, tail)
	out = append(out,
		"Safety: only files carrying the mirror_of marker under Library/ "+
			"are ever changed; nothing outside Library/ is touched.")
	return strings.Join(out, "\n") + "\n"
}

// pyOr mirrors `x or fallback` for the report's vault lines.
func pyOr(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}

// planTree mirrors mirror.py _plan_tree: compute the action list for one
// Library tree (pure — writes nothing). Matching is by RAW source
// value, so a note the owner moved between category folders keeps its
// identity (the stale copy is removed, a fresh one appears — a MOVE).
func planTree(label, vaultPath, treeRoot string) *TreePlan {
	plan := &TreePlan{Label: label, VaultPath: vaultPath}
	var notes []SourceNote
	if vaultPath != "" {
		if fi, err := os.Stat(vaultPath); err == nil && fi.IsDir() {
			plan.VaultPresent = true
			scanned, dupes, noSource, errors := scanSourceNotes(vaultPath)
			notes = scanned
			plan.Duplicates = dupes
			plan.SkippedNoSource = noSource
			plan.ScanErrors = errors
		}
	}
	plan.Sources = len(notes)

	bySource := map[string]*SourceNote{}
	for i := range notes {
		bySource[notes[i].Source] = &notes[i]
	}
	mirrors := scanMirrorTree(treeRoot)
	plan.UnmarkedFiles = mirrors.Unmarked
	plan.ScanErrors += mirrors.ReadErrors

	// Every marker file NOT sitting at its source's current location will
	// be removed — as a MOVE's old path, or as a plain DELETE (orphan /
	// duplicate marker). doomed is the full removal set, used to know
	// which expected paths are freed by deletions.
	doomed := map[string]bool{}
	for sourceValue, rels := range mirrors.BySource {
		expected := ""
		if n, ok := bySource[sourceValue]; ok {
			expected = n.Rel
		}
		for _, rel := range rels {
			if expected == "" || rel != expected {
				doomed[rel] = true
			}
		}
	}
	moveOlds := map[string]bool{}

	// --- one expected copy per source ---
	for i := range notes {
		note := &notes[i]
		expectedAbs, err := SafeJoin(treeRoot, note.Rel)
		if err != nil {
			plan.Conflicts = append(plan.Conflicts, ConflictItem{
				Rel: note.Rel, Reason: "unsafe path: " + err.Error()})
			continue
		}
		existingRels := []string{}
		for _, r := range mirrors.BySource[note.Source] {
			if r == note.Rel {
				existingRels = append(existingRels, r)
			}
		}
		text, ok := readTextFile(note.Path)
		if !ok {
			plan.ScanErrors++
			continue
		}
		note.Text = text
		mirrorText := BuildMirrorNote(note.Text, note.Source, MirrorBanners[label])

		if len(existingRels) > 0 {
			// a marked copy is already at the right spot — compare content
			onDisk, ok := readTextFile(expectedAbs)
			if !ok {
				plan.ScanErrors++
				continue
			}
			// v0.16.0 — Phase 6: the rebuilt copy must carry over the
			// owner-approved Related (auto) block (a rebuild would
			// otherwise silently wipe it); with the block carried over,
			// an unchanged note still compares equal (a KEEP).
			mirrorText = PreserveRelatedBlock(onDisk, mirrorText)
			if onDisk == mirrorText {
				plan.Keeps++
			} else {
				plan.Updates = append(plan.Updates, UpdateItem{
					Rel: note.Rel, Text: mirrorText, Source: note.Source})
			}
			continue
		}

		// no marked copy of ours at the expected spot
		if _, err := os.Stat(expectedAbs); err == nil && !doomed[note.Rel] {
			plan.Conflicts = append(plan.Conflicts, ConflictItem{
				Rel: note.Rel,
				Reason: "a file without the mirror_of marker is in the way " +
					"(left untouched)"})
			continue
		}
		var movedFrom []string
		for _, r := range mirrors.BySource[note.Source] {
			if r != note.Rel {
				movedFrom = append(movedFrom, r)
			}
		}
		if len(movedFrom) > 0 {
			moveOlds[movedFrom[0]] = true
			// v0.16.0 — Phase 6: a MOVED mirror copy keeps its approved
			// Related (auto) block too (read from the old spot).
			if oldAbs, err := SafeJoin(treeRoot, movedFrom[0]); err == nil {
				if oldText, ok := readTextFile(oldAbs); ok {
					mirrorText = PreserveRelatedBlock(oldText, mirrorText)
				}
			}
			plan.Moves = append(plan.Moves, MoveItem{
				Old: movedFrom[0], New: note.Rel, Text: mirrorText,
				Source: note.Source})
		} else {
			plan.Creates = append(plan.Creates, CreateItem{
				Rel: note.Rel, Text: mirrorText, Source: note.Source})
		}
	}

	remaining := make([]string, 0, len(doomed))
	for rel := range doomed {
		if !moveOlds[rel] {
			remaining = append(remaining, rel)
		}
	}
	sort.Strings(remaining)
	plan.Deletes = remaining
	return plan
}

// ---------------------------------------------------------------------------
// Entry point
// ---------------------------------------------------------------------------

// cleanupEmptyDirs removes folders the deletions left empty — inside
// the tree only. os.Rmdir fails on non-empty folders, so a folder
// holding any other file can never be removed. The tree root itself is
// handled by the caller.
func cleanupEmptyDirs(treeRoot string) {
	if treeRoot == "" {
		return
	}
	if fi, err := os.Stat(treeRoot); err != nil || !fi.IsDir() {
		return
	}
	var walk func(dir string) []string
	walk = func(dir string) []string {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return nil
		}
		var subdirs []string
		for _, e := range entries {
			if e.IsDir() {
				subdirs = append(subdirs, walk(filepath.Join(dir, e.Name()))...)
			}
		}
		return append(subdirs, dir)
	}
	for _, dir := range walk(treeRoot) {
		if dir == treeRoot {
			continue
		}
		if isInside(dir, treeRoot) {
			_ = os.Remove(dir) // os.rmdir semantics: fails on non-empty
		}
	}
}

// readMarkerSource returns the mirror_of value of path ("" + false when
// unreadable/unmarked).
func readMarkerSource(path string) (string, bool) {
	head, ok := readHead800(path)
	if !ok {
		return "", false
	}
	m := mirrorRE.FindStringSubmatchIndex(head)
	if m == nil {
		return "", false
	}
	return unquote(head[m[2]:m[3]]), true
}

// RunMirror plans (and with apply=true performs) the Library mirror
// sync. Dry-run by default: the returned plan describes every action
// and the manual vault is not touched. With apply=true every safety
// check re-runs first, deletions happen before writes, and each write
// target is re-verified to be inside its Library tree and either free
// or a mirror-marked copy of the SAME source.
func RunMirror(manualVault string, githubVault, websitesVault string,
	apply bool, log func(string)) (*MirrorPlan, error) {

	say := func(msg string) {
		if log != nil {
			log(msg)
		}
	}

	githubVault = helpers.Strip(githubVault)
	websitesVault = helpers.Strip(websitesVault)
	if err := CheckVaultPaths(manualVault, githubVault, websitesVault, apply); err != nil {
		return nil, err
	}

	manual := helpers.Strip(manualVault)
	libraryRoot := filepath.Join(manual, LibraryFolder)
	plan := NewMirrorPlan(manual, githubVault, websitesVault, apply)

	treeSpecs := []struct {
		Label string
		Vault string
	}{
		{GitHubTree, githubVault},
		{WebsitesTree, websitesVault},
	}
	for _, ts := range treeSpecs {
		if ts.Vault == "" {
			continue
		}
		treeRoot := filepath.Join(libraryRoot, ts.Label)
		plan.Trees[ts.Label] = planTree(ts.Label, ts.Vault, treeRoot)
	}
	if len(plan.Trees) == 0 {
		return nil, &MirrorError{Msg: "nothing to mirror: neither machine " +
			"vault path is configured"}
	}

	if !apply {
		for _, ts := range treeSpecs {
			if tree, ok := plan.Trees[ts.Label]; ok {
				say(fmt.Sprintf("[%s] %d library notes — %d change(s) planned (dry run)",
					tree.Label, tree.Sources, tree.Changes()))
			}
		}
		return plan, nil
	}

	// ---- apply: deletions first (incl. move old paths), then writes ----
	for _, ts := range treeSpecs {
		tree, ok := plan.Trees[ts.Label]
		if !ok {
			continue
		}
		treeRoot := filepath.Join(libraryRoot, ts.Label)
		var dels []string
		for _, d := range tree.Deletes {
			dels = append(dels, d)
		}
		for _, m := range tree.Moves {
			dels = append(dels, m.Old)
		}
		for _, rel := range dels {
			target, err := SafeJoin(treeRoot, rel)
			if err != nil {
				return nil, err
			}
			if !isInside(target, treeRoot) {
				return nil, &MirrorError{Msg: fmt.Sprintf(
					"delete escaped the Library tree: %s", target)}
			}
			if _, found := readMarkerSource(target); !found {
				// not ours any more (changed since the plan) — never delete
				say(fmt.Sprintf("[%s] SKIP delete %s — no mirror marker",
					ts.Label, rel))
				continue
			}
			if err := dryrun.Remove(target); err != nil {
				return nil, err
			}
		}
		type writeItem struct{ Rel, Text, Source string }
		var writes []writeItem
		for _, c := range tree.Creates {
			writes = append(writes, writeItem{c.Rel, c.Text, c.Source})
		}
		for _, u := range tree.Updates {
			writes = append(writes, writeItem{u.Rel, u.Text, u.Source})
		}
		for _, m := range tree.Moves {
			writes = append(writes, writeItem{m.New, m.Text, m.Source})
		}
		for _, w := range writes {
			target, err := SafeJoin(treeRoot, w.Rel)
			if err != nil {
				return nil, err
			}
			if !isInside(target, treeRoot) {
				return nil, &MirrorError{Msg: fmt.Sprintf(
					"write escaped the Library tree: %s", target)}
			}
			if _, statErr := os.Stat(target); statErr == nil {
				marker, _ := readMarkerSource(target)
				if marker != w.Source {
					// a file that is not our mirror copy of this source is
					// in the way — never touch it
					say(fmt.Sprintf(
						"[%s] SKIP write %s — target is not our mirror copy "+
							"of this source", ts.Label, w.Rel))
					continue
				}
			}
			if err := dryrun.MkdirAll(filepath.Dir(target), true); err != nil {
				return nil, err
			}
			if err := storage.AtomicWriteText(target, w.Text); err != nil {
				return nil, err
			}
		}
		cleanupEmptyDirs(treeRoot)
		// an entirely emptied tree folder is tidied away (rmdir only
		// succeeds when nothing else is left in it)
		if fi, err := os.Stat(treeRoot); err == nil && fi.IsDir() {
			_ = os.Remove(treeRoot)
		}
	}
	// tidy an entirely emptied Library root too (never removes anything
	// that still holds files — rmdir fails on non-empty folders)
	if fi, err := os.Stat(libraryRoot); err == nil && fi.IsDir() {
		_ = os.Remove(libraryRoot)
	}

	plan.Applied = true
	for _, ts := range treeSpecs {
		if tree, ok := plan.Trees[ts.Label]; ok {
			say(fmt.Sprintf("[%s] mirrored %d notes — %d change(s) applied",
				tree.Label, tree.Sources, tree.Changes()))
		}
	}
	return plan, nil
}
