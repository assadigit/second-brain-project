package websitepipeline

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/assadigit/second-brain-project/internal/constants"
	"github.com/assadigit/second-brain-project/internal/dryrun"
	"github.com/assadigit/second-brain-project/internal/helpers"
	"github.com/assadigit/second-brain-project/internal/links"
	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/taxonomy"
)

// directorynote.go — T-41: core/website_directory.py (432 LOC) — THE LAW
// sweep + the categorized Website Directory (v0.28.0/v0.35.0).
//
// Two jobs, both run-side, both idempotent, both dry-run aware:
//
//  1. SweepBannedNotes (sweep_banned_notes) — enforce the owner's law
//     against the vault ON DISK: every APP-OWNED note whose source: is on
//     a banned domain moves to <vault>/.trash/banned-domains/ (Obsidian's
//     hidden trash; VaultSeal never backs it up, VaultIndex/mirror/
//     note_state never read it). Hand-written notes are never touched —
//     counted and reported instead (the user's own content is sacred).
//     v0.35.0 — a platform whose EVERY domain is banned is never
//     collected anymore, so its _inbox table is quarantined to the same
//     .trash folder instead of lingering as a dead "collection".
//
//  2. BuildWebsiteDirectory (build_website_directory) — regenerate
//     <vault>/000 📚 Website Directory.md: the consolidated, categorized,
//     clickable index. Categories in taxonomy order → subcategories →
//     one line per site: a wiki-link to the note (alias = its title),
//     the TL;DR one-liner, the pricing word when known, and a ↗ external
//     link. The file is app-owned (kind: directory) so overwriting it
//     never violates the never-rewrite-user-notes rule.
//
// The parse is deliberately LIGHT (the reference's own law: "light regex
// — notes are small"): frontmatter key lines, the first `> **TL;DR:**`
// line, the first `# ` heading. Regex classes carry PYTHON's Unicode \s
// set (U+00A0, \x1c… are whitespace to the reference, not to Go's \s).
// The walk mirrors os.walk's top-down order exactly: a directory's own
// files (sorted) are yielded before ANY of its subdirectories' files —
// filepath.WalkDir interleaves them lexically, so this is a hand-rolled
// recursion, not WalkDir.

// DirectoryFilename mirrors DIRECTORY_FILENAME: the directory note
// (app-owned; sorts to the TOP of the vault so it is the first thing the
// owner sees when the vault opens).
const DirectoryFilename = "000 📚 Website Directory.md"

// BannedQuarantineRelpath mirrors QUARANTINE_RELPATH: where swept
// (banned) notes go — inside Obsidian's hidden .trash so they disappear
// from the vault view but stay recoverable by hand. (The BANISHMENT's
// quarantine is a different folder — BanishQuarantineRelpath
// ".trash/banished".)
const BannedQuarantineRelpath = ".trash/banned-domains"

// directorySkipDirs mirrors _SKIP_DIRS: folders the DIRECTORY walk never
// enters (system/record folders + the pipeline's own queues — they are
// not library notes).
var directorySkipDirs = map[string]bool{
	"_inbox": true, "_review": true, "_missing": true, "_moc": true,
	"attachments": true, ".obsidian": true, ".trash": true, ".git": true,
}

// sweepSkipDirs mirrors _SKIP_DIRS_SWEEP: the SWEEP is broader — it
// ENTERS _review/_missing/_moc (the legacy banned-domain pile lives in
// _review) and only skips the dot-folders and the _inbox record tables
// (handled separately — v0.35.0 quarantines the banned platforms' tables
// whole).
var sweepSkipDirs = map[string]bool{
	"_inbox": true, ".obsidian": true, ".trash": true, ".git": true,
}

// pyWS is Python's Unicode \s class for str patterns (the same set as
// str.isspace(): ASCII whitespace + \x1c-\x1f + U+0085, U+00A0, and the
// Unicode space separators). Go's own \s is ASCII-only — using it would
// leave U+00A0 and friends attached to parsed values.
const pyWS = `[\t\n\v\f\r \x{1c}\x{1d}\x{1e}\x{1f}\x{85}\x{a0}\x{1680}\x{2000}-\x{200a}\x{2028}\x{2029}\x{202f}\x{205f}\x{3000}]`

// fmLineRe mirrors _FM_LINE_RE: one frontmatter key line.
var fmLineRe = regexp.MustCompile(`^([A-Za-z0-9_]+):` + pyWS + `*(.*?)` + pyWS + `*$`)

// tldrRe mirrors _TLDR_RE (MULTILINE): the TL;DR callout line.
var tldrRe = regexp.MustCompile(`(?m)^>` + pyWS + `*\*\*TL;DR:\*\*` + pyWS + `*(.+?)` + pyWS + `*$`)

// h1Re mirrors _H1_RE (MULTILINE): the first level-1 heading.
var h1Re = regexp.MustCompile(`(?m)^#` + pyWS + `+(.+?)` + pyWS + `*$`)

// TLDRMax / TitleMax mirror _TLDR_MAX / _TITLE_MAX: the directory line's
// trim caps.
const TLDRMax = 140
const TitleMax = 80

// noteReadCap is _parse_note's read cap: the first 512,000 DECODED
// characters (Python text-mode read counts characters, not bytes —
// readHeadRunes).
const noteReadCap = 512_000

// stripYAMLQuotes mirrors _strip_yaml_quotes: strip one matching pair of
// quotes the note writers may add.
func stripYAMLQuotes(value string) string {
	v := helpers.Strip(value)
	if len(v) >= 2 && v[0] == v[len(v)-1] && (v[0] == '"' || v[0] == '\'') {
		return v[1 : len(v)-1]
	}
	return v
}

// dirNote is _parse_note's returned dict: a light frontmatter + TL;DR +
// H1 parse of one note file. OWNERSHIP is the caller's decision:
// app_owned (managed_by: gitcurator) lets the sweep move app-owned notes
// while counting hand-written ones, and the directory builder index only
// what the app wrote. A nil note = the reference's None (no frontmatter
// — not a GitCurator-shaped note); never an error.
type dirNote struct {
	appOwned    bool
	source      string
	category    string
	subcategory string
	pricing     string
	fetchStatus string
	tldr        string
	title       string
}

// parseDirNote mirrors _parse_note. '---\n<fm lines>\n---\n<body>' — the
// opening '---' has no newline before it, so split('\n---') puts the
// frontmatter in parts[0] and the body (which may itself contain a
// '\n---' rule line) after it; the body is rebuilt with the removed
// separators when the split found more of them.
func parseDirNote(path string) *dirNote {
	text, ok := readHeadRunes(path, noteReadCap)
	if !ok {
		return nil
	}
	if !strings.HasPrefix(text, "---") {
		return nil
	}
	parts := strings.SplitN(text, "\n---", 3)
	if len(parts) < 2 {
		return nil
	}
	fm := map[string]string{}
	for _, line := range pySplitLines(parts[0]) {
		if m := fmLineRe.FindStringSubmatch(line); m != nil {
			fm[m[1]] = stripYAMLQuotes(m[2])
		}
	}
	body := parts[1]
	if len(parts) > 2 {
		body = parts[1] + "\n---" + parts[2]
	}
	tldr := ""
	if m := tldrRe.FindStringSubmatch(body); m != nil {
		tldr = helpers.Strip(m[1])
	}
	if tldr == "—" || tldr == "-" || tldr == "" {
		tldr = ""
	}
	title := ""
	if m := h1Re.FindStringSubmatch(body); m != nil {
		title = helpers.Strip(m[1])
	}
	return &dirNote{
		appOwned:    fm["managed_by"] == constants.ManagedByGitcurator,
		source:      fmGet(fm, "source"),
		category:    fmGet(fm, "category"),
		subcategory: fmGet(fm, "subcategory"),
		pricing:     fmGetOr(fm, "pricing", "unknown"),
		fetchStatus: fmGet(fm, "fetch_status"),
		tldr:        tldr,
		title:       title,
	}
}

// fmGet mirrors fm.get(key, ”) — missing keys read as the empty string
// (the zero value).
func fmGet(fm map[string]string, key string) string {
	if v, ok := fm[key]; ok {
		return v
	}
	return ""
}

// fmGetOr mirrors fm.get(key, default).
func fmGetOr(fm map[string]string, key, def string) string {
	if v, ok := fm[key]; ok {
		return v
	}
	return def
}

// iterNoteFiles mirrors _iter_note_files: every .md under vaultPath
// except the system folders, in os.walk's top-down order — each
// directory's own files (sorted) BEFORE any of its subdirectories'
// files, subdirectories recursed in sorted order. Dot-folders and the
// skip-set folders are pruned; dot-FILES are not (only directories carry
// the dot rule). The directory note itself is excluded by name. Symlinked
// directories are never followed (os.walk's followlinks=False; they ride
// the files list and fail the .md suffix or the parse instead).
func iterNoteFiles(vaultPath string, forSweep bool, visit func(path string)) {
	if vaultPath == "" || !isDirPath(vaultPath) {
		return
	}
	skip := directorySkipDirs
	if forSweep {
		skip = sweepSkipDirs
	}
	walkNoteDir(vaultPath, skip, visit)
}

// walkNoteDir is the pruned os.walk recursion (files before subdirs).
func walkNoteDir(dir string, skip map[string]bool, visit func(path string)) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return // os.walk with onerror=None swallows read errors
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(helpers.PyLower(name), ".md") && name != DirectoryFilename {
			path := filepath.Join(dir, name)
			if !isFilePath(path) {
				// A symlinked directory rides Python's dirs list (DirEntry.
				// is_dir() FOLLOWS links) and followlinks=False never
				// descends it — never yielded, never scanned. A Stat-based
				// file check sees the same thing.
				continue
			}
			visit(path)
		}
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if skip[name] || strings.HasPrefix(name, ".") {
			continue
		}
		walkNoteDir(filepath.Join(dir, name), skip, visit)
	}
}

// countReviewNotes mirrors _count_review_notes: how many notes wait in
// the pipeline's _review folder (listdir + .md count — a subdirectory
// named *.md counts too; the reference's own quirk).
func countReviewNotes(vaultPath string) int {
	entries, err := os.ReadDir(filepath.Join(vaultPath, "_review"))
	if err != nil {
		return 0
	}
	n := 0
	for _, e := range entries {
		if strings.HasSuffix(helpers.PyLower(e.Name()), ".md") {
			n++
		}
	}
	return n
}

// ---------------------------------------------------------------------------
// 1. THE LAW — the banned-domain sweep
// ---------------------------------------------------------------------------

// SweptNote is one moved banned note: (url, src, dst).
type SweptNote struct {
	URL string
	Src string
	Dst string
}

// SweptTable is one quarantined banned-platform table: (platform, src,
// dst).
type SweptTable struct {
	Platform string
	Src      string
	Dst      string
}

// SweepReport mirrors sweep_banned_notes' report dict.
type SweepReport struct {
	Moved           []SweptNote
	KeptHandwritten int
	Scanned         int
	MovedTables     []SweptTable
}

// SweepBannedNotes mirrors sweep_banned_notes: move every app-owned note
// on a banned domain out of the vault's library into
// <vault>/.trash/banned-domains/.
//
// Only APP-OWNED notes move (managed_by: gitcurator). Hand-written notes
// with a banned source: are the owner's — counted, reported, never
// touched. Every app-owned banned note leaves — _review placeholders AND
// real notes alike (the law has no "but it processed fine" exception).
// After the note walk, the banned PLATFORMS' _inbox tables are
// quarantined whole: a table moves only when EVERY domain of its
// platform is on the ban list (partially-banned platforms — and 'other',
// which is not a platform — keep their tables; their banned ROWS are
// pruned by the intake writer's prune pass instead).
//
// Dry-run aware (moves are recorded, not performed). Idempotent — the
// quarantine folder is skipped by the walk.
func SweepBannedNotes(vaultPath string, blockedDomains []string, log LogFn) SweepReport {
	if log == nil {
		log = noopLog
	}
	report := SweepReport{}
	if vaultPath == "" || !isDirPath(vaultPath) {
		return report
	}
	if len(blockedDomains) == 0 {
		return report
	}
	quarantineDir := filepath.Join(vaultPath, filepath.FromSlash(BannedQuarantineRelpath))
	iterNoteFiles(vaultPath, true, func(path string) {
		report.Scanned++
		note := parseDirNote(path)
		if note == nil {
			return
		}
		u := note.source
		if u == "" {
			return
		}
		if !links.DomainIsBlocked(u, blockedDomains) {
			return
		}
		if !note.appOwned {
			// The owner's own writing — sacred. Counted + reported only.
			report.KeptHandwritten++
			return
		}
		dst := storage.UniquePath(filepath.Join(quarantineDir, filepath.Base(path)))
		if err := sweepMoveFile(path, dst, quarantineDir); err != nil {
			rel, relErr := filepath.Rel(vaultPath, path)
			if relErr != nil {
				rel = path
			}
			log(fmt.Sprintf("⚠️ Law sweep could not move %s: %v", rel, err), "warning")
			return
		}
		report.Moved = append(report.Moved, SweptNote{URL: u, Src: path, Dst: dst})
	})
	if len(report.Moved) > 0 {
		log(fmt.Sprintf("🧹 Law sweep: moved %d banned-domain note(s) to %s — the manifest's blocked bucket is the record", len(report.Moved), BannedQuarantineRelpath), "info")
	}
	if report.KeptHandwritten > 0 {
		log(fmt.Sprintf("✍️ Law sweep: %d hand-written note(s) with banned sources kept — they are yours; delete them by hand if you want them gone", report.KeptHandwritten), "info")
	}
	// v0.35.0 — the banned platforms' _inbox tables are collections of
	// links the app must never collect anymore; quarantine them whole.
	// (Row-level pruning for the remaining tables happens in the intake
	// writer's prune pass — gui/platform_intake.prune_inbox_tables.)
	banned := map[string]bool{}
	for _, d := range blockedDomains {
		e := strings.TrimLeft(helpers.PyLower(helpers.Strip(d)), ".")
		if e != "" {
			banned[e] = true
		}
	}
	// links.PlatformDomains carries the reference's insertion order —
	// the order the report's moved_tables rides ('other' is not a
	// platform: no domains, never moved).
	for _, p := range links.PlatformDomains {
		allBanned := true
		for _, d := range p.Domains {
			e := strings.TrimLeft(helpers.PyLower(helpers.Strip(d)), ".")
			if !banned[e] {
				allBanned = false
				break
			}
		}
		if !allBanned {
			continue // partially banned (or not a platform) → keep the table
		}
		tableFile := links.PlatformTableFiles[p.Key]
		src := filepath.Join(vaultPath, "_inbox", tableFile)
		if !isFilePath(src) {
			continue
		}
		dst := storage.UniquePath(filepath.Join(quarantineDir, tableFile))
		if err := sweepMoveFile(src, dst, quarantineDir); err != nil {
			log(fmt.Sprintf("⚠️ Law sweep could not move _inbox/%s: %v", tableFile, err), "warning")
			continue
		}
		report.MovedTables = append(report.MovedTables, SweptTable{Platform: p.Key, Src: src, Dst: dst})
	}
	if len(report.MovedTables) > 0 {
		names := make([]string, 0, len(report.MovedTables))
		for _, t := range report.MovedTables {
			names = append(names, "_inbox/"+filepath.Base(t.Src))
		}
		sort.Strings(names)
		log(fmt.Sprintf("🧹 Law sweep: quarantined %d banned-platform table(s) (%s) — those platforms are never collected (notes or review) anymore", len(report.MovedTables), strings.Join(names, ", ")), "info")
	}
	return report
}

// sweepMoveFile is the sweep's one move: read the file as text
// (errors='replace' — invalid bytes become U+FFFD; universal newlines —
// \r\n and lone \r become \n, exactly the reference's text-mode
// round-trip: a quarantined \r\n note lands normalized), then the
// dry-run-seamed makedirs/write/remove trio. The OSError contract is
// the caller's (one warning line, the note stays, the walk continues).
func sweepMoveFile(src, dst, quarantineDir string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	content := helpers.PyNewlines(decodeUTF8Replace(string(data)))
	if err := dryrun.MkdirAll(quarantineDir, true); err != nil {
		return err
	}
	if err := dryrun.WriteText(dst, content); err != nil {
		return err
	}
	return dryrun.Remove(src)
}

// ---------------------------------------------------------------------------
// 2. The Website Directory — the consolidated categorized index
// ---------------------------------------------------------------------------

// trimRunes mirrors _trim: trim to limit CODEPOINTS at a word boundary
// with an ellipsis.
func trimRunes(text string, limit int) string {
	text = helpers.Strip(text)
	runes := []rune(text)
	if len(runes) <= limit {
		return text
	}
	head := string(runes[:limit])
	cut := head
	if idx := strings.LastIndex(cut, " "); idx >= 0 {
		cut = cut[:idx]
	}
	if cut == "" {
		cut = head
	}
	cut = rstripChars(cut, " ,;:.-")
	return cut + "…"
}

// rstripChars mirrors str.rstrip(chars): strip trailing chars in the set.
func rstripChars(s, chars string) string {
	set := map[rune]bool{}
	for _, r := range chars {
		set[r] = true
	}
	rs := []rune(s)
	for len(rs) > 0 && set[rs[len(rs)-1]] {
		rs = rs[:len(rs)-1]
	}
	return string(rs)
}

// aliasSafe mirrors _alias_safe: strip the characters that break a
// wiki-link ALIAS.
func aliasSafe(title string) string {
	for _, ch := range "[]|^#" {
		title = strings.ReplaceAll(title, string(ch), "")
	}
	return helpers.Strip(title)
}

// strIn reports membership in a string slice.
func strIn(list []string, s string) bool {
	for _, e := range list {
		if e == s {
			return true
		}
	}
	return false
}

// categoryOrder mirrors _category_order: taxonomy order first, then any
// leftover/unknown names (e.g. from an older taxonomy) alphabetically.
// (The reference's None-key tail is dead code — tree keys are always
// strings — and is not ported.)
func categoryOrder(namesInUse map[string]bool, tax *taxonomy.Taxonomy) []string {
	var ordered []string
	if tax != nil {
		for _, c := range tax.CategoryNames() {
			if namesInUse[c] {
				ordered = append(ordered, c)
			}
		}
	}
	var leftovers []string
	for n := range namesInUse {
		if !strIn(ordered, n) && helpers.Strip(n) != "" {
			leftovers = append(leftovers, n)
		}
	}
	sort.Strings(leftovers)
	return append(ordered, leftovers...)
}

// directoryEntry is one site line's facts (the tree's leaf dict).
type directoryEntry struct {
	title   string
	relpath string
	url     string
	tldr    string
	pricing string
}

// DirectoryReport mirrors build_website_directory's return dict.
type DirectoryReport struct {
	Path       string
	Sites      int
	Categories int
	Review     int
}

// BuildWebsiteDirectory mirrors build_website_directory: (re)generate
// <vault>/000 📚 Website Directory.md.
//
// Groups every app-owned website note by its frontmatter category →
// subcategory (taxonomy order, leftovers last) and writes one line per
// site: a wiki-link to the note, the TL;DR one-liner, the pricing word
// when known, and a ↗ link to the live site. Notes waiting in _review
// are not listed — they are counted in the footer callout so nothing
// looks silently missing. The directory file itself is app-owned (kind:
// directory) and has no source: line, so the VaultIndex dedupe layer
// never sees it. Returns nil when the vault is missing/empty (nothing to
// index yet) or the write failed (logged, never raised).
func BuildWebsiteDirectory(vaultPath string, tax *taxonomy.Taxonomy, config map[string]any, log LogFn) *DirectoryReport {
	if log == nil {
		log = noopLog
	}
	if vaultPath == "" || !isDirPath(vaultPath) {
		return nil
	}
	if tax == nil && config != nil {
		if loaded, err := taxonomy.LoadTaxonomyFromConfig(config); err == nil {
			tax = loaded
		}
	}

	// Collect: category -> subcategory -> entries.
	tree := map[string]map[string][]directoryEntry{}
	sites := 0
	blocked := links.BlockedDomainsFromConfig(config)
	iterNoteFiles(vaultPath, false, func(path string) {
		note := parseDirNote(path)
		if note == nil || !note.appOwned {
			return // only the app's own notes are indexed
		}
		u := note.source
		if u == "" {
			return
		}
		if links.DomainIsBlocked(u, blocked) {
			return // the law: never listed, whatever it is
		}
		sites++
		cat := helpers.Strip(note.category)
		if cat == "" {
			cat = "Uncategorized"
		}
		sub := helpers.Strip(note.subcategory)
		title := note.title
		if title == "" {
			stem, _ := storage.PySplitExt(filepath.Base(path))
			title = stem
		}
		rel, relErr := filepath.Rel(vaultPath, path)
		if relErr != nil {
			rel = path
		}
		subs, ok := tree[cat]
		if !ok {
			subs = map[string][]directoryEntry{}
			tree[cat] = subs
		}
		subs[sub] = append(subs[sub], directoryEntry{
			title:   title,
			relpath: rel,
			url:     u,
			tldr:    note.tldr,
			pricing: note.pricing,
		})
	})

	if len(tree) == 0 {
		return nil
	}

	lines := []string{}
	lines = append(lines, "---")
	lines = append(lines, fmt.Sprintf("managed_by: \"%s\"", constants.ManagedByGitcurator))
	lines = append(lines, "kind: \"directory\"")
	lines = append(lines, fmt.Sprintf("generated: \"%s\"", localDate()))
	lines = append(lines, "---")
	lines = append(lines, "")
	lines = append(lines, "# 📚 Website Directory")
	lines = append(lines, "")
	nCats := 0
	for _, subs := range tree {
		if len(subs) > 0 {
			nCats++
		}
	}
	lines = append(lines, fmt.Sprintf("> [!info] Auto-generated index of the curated library — %d website(s) across %d categor(ies). Click a name to open its note (details inside); ↗ opens the site in your browser.", sites, nCats))
	lines = append(lines, "")

	namesInUse := map[string]bool{}
	for c, subs := range tree {
		if len(subs) > 0 {
			namesInUse[c] = true
		}
	}
	for _, cat := range categoryOrder(namesInUse, tax) {
		subs := tree[cat]
		catTotal := 0
		for _, v := range subs {
			catTotal += len(v)
		}
		lines = append(lines, fmt.Sprintf("## %s (%d)", cat, catTotal))
		lines = append(lines, "")
		// Subcategories in taxonomy order, then the empty-key group last.
		subNames := map[string]bool{}
		for s := range subs {
			subNames[s] = true
		}
		var orderedSubs []string
		if tax != nil && cat != "Uncategorized" {
			for _, s := range tax.SubcategoriesOf(cat) {
				if subNames[s] {
					orderedSubs = append(orderedSubs, s)
				}
			}
		}
		var leftovers []string
		for s := range subNames {
			if !strIn(orderedSubs, s) && s != "" {
				leftovers = append(leftovers, s)
			}
		}
		sort.Strings(leftovers)
		orderedSubs = append(orderedSubs, leftovers...)
		if subNames[""] {
			orderedSubs = append(orderedSubs, "")
		}
		for _, sub := range orderedSubs {
			entries := subs[sub]
			sort.SliceStable(entries, func(i, j int) bool {
				ki, kj := helpers.PyLower(entries[i].title), helpers.PyLower(entries[j].title)
				if ki != kj {
					return ki < kj
				}
				return helpers.PyLower(entries[i].relpath) < helpers.PyLower(entries[j].relpath)
			})
			if sub != "" {
				lines = append(lines, fmt.Sprintf("### %s (%d)", sub, len(entries)))
				lines = append(lines, "")
			}
			for _, e := range entries {
				link := strings.ReplaceAll(e.relpath, "\\", "/")
				if strings.HasSuffix(helpers.PyLower(link), ".md") {
					link = link[:len(link)-3]
				}
				alias := aliasSafe(trimRunes(e.title, TitleMax))
				if alias == "" {
					alias = "Untitled"
				}
				parts := []string{fmt.Sprintf("[[%s|%s]]", link, alias)}
				if e.tldr != "" {
					parts = append(parts, trimRunes(e.tldr, TLDRMax))
				}
				if e.pricing == "free" || e.pricing == "freemium" || e.pricing == "paid" {
					parts = append(parts, e.pricing)
				}
				parts = append(parts, fmt.Sprintf("[↗](%s)", e.url))
				kept := make([]string, 0, len(parts))
				for _, p := range parts {
					if p != "" {
						kept = append(kept, p)
					}
				}
				lines = append(lines, "- "+strings.Join(kept, " — "))
			}
			lines = append(lines, "")
		}
	}

	review := countReviewNotes(vaultPath)
	if review > 0 {
		lines = append(lines, fmt.Sprintf("> [!todo] 📥 %d link(s) still waiting in `_review/` (fetch failed / needs a look) — they join this directory once processed.", review))
		lines = append(lines, "")
	}
	lines = append(lines, "*Regenerated automatically after every curation run — edits here are overwritten.*")

	outPath := filepath.Join(vaultPath, DirectoryFilename)
	if err := dryrun.MkdirAll(vaultPath, true); err != nil {
		log(fmt.Sprintf("⚠️ Website Directory could not be written: %v", err), "warning")
		return nil
	}
	if err := storage.AtomicWriteText(outPath, strings.Join(lines, "\n")+"\n"); err != nil {
		log(fmt.Sprintf("⚠️ Website Directory could not be written: %v", err), "warning")
		return nil
	}
	return &DirectoryReport{Path: outPath, Sites: sites, Categories: nCats, Review: review}
}
