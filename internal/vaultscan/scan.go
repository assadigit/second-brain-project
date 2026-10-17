// Package vaultscan ports core/vault_scan.py (v0.61.0 → v0.83.0) —
// THE VAULT SCAN: the LLM-assisted refiling engine with the
// deterministic "regex brain" (declared-home, sibling-domain,
// taxonomy-tag passes), weighed-evidence scoring, chunked LLM filing,
// validated plan building, and the confirm-then-apply machinery.
//
// Spec: docs/port-blueprint/module-specs/spec-core-vault-scan.md;
// task T-45 (Phase 4). The `kind` door runs the same grammar for the
// Websites vault ('websites') and the GitHub Repos vault ('repos').
//
// The four laws (the reference's own header, kept):
//
//  1. ONE deletion grammar — the scan's deletions ARE
//     websitepipeline.ScanPendingBanishments (websites) or
//     repobanish.ScanPendingRepoBanishments (repos); this module never
//     re-detects, never re-words, never invents a second grammar.
//  2. The plan is a PROPOSAL — nothing moves or deletes until the
//     owner answers (the confirm gate is the caller's law).
//  3. Existing notes are never rewritten — a move is a FILE move; the
//     folder IS the category, so the state row is re-pointed; content
//     is never touched (the census stamp is the ONE sanctioned
//     frontmatter edit after the move).
//  4. The LLM is a guest — its JSON is VALIDATED before the owner ever
//     sees it: only listed candidates, only legal destinations, names
//     sanitized; anything rejected is dropped with a log line, never a
//     crash.
package vaultscan

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/assadigit/second-brain-project/internal/chrometabs"
	"github.com/assadigit/second-brain-project/internal/constants"
	"github.com/assadigit/second-brain-project/internal/helpers"
	"github.com/assadigit/second-brain-project/internal/links"
	"github.com/assadigit/second-brain-project/internal/notereads"
	"github.com/assadigit/second-brain-project/internal/websitepipeline"
)

// ScanPrefix mirrors SCAN_PREFIX: the scan's own log prefix.
const ScanPrefix = "🔍 Vault scan:"

// DefaultTooBroad mirrors DEFAULT_TOO_BROAD: a leaf folder holding
// more notes than this reads as TOO BROAD (config
// 'scan_too_broad_threshold' overrides).
const DefaultTooBroad = 40

// DefaultMaxNotes mirrors DEFAULT_MAX_NOTES: how many candidate notes
// ride the LLM asks at all (config 'scan_max_notes' overrides).
const DefaultMaxNotes = 300

// DefaultLLMBatch mirrors DEFAULT_LLM_BATCH: how many candidates ride
// ONE LLM call (config 'scan_llm_batch' overrides) — the chunked
// filing pass (v0.70.0).
const DefaultLLMBatch = 60

// ProtectedFolders mirrors _PROTECTED_FOLDERS: folders the scan never
// touches and never proposes as destinations.
var ProtectedFolders = []string{"_review", "_moc", "_inbox"}

// ProtectedFoldersRepos mirrors _PROTECTED_FOLDERS_REPOS: the repos
// vault's own protected set (the 404 machinery owns _missing).
var ProtectedFoldersRepos = []string{"_review", "_moc", "_inbox", "_missing"}

// VaultKindWebsites / VaultKindRepos mirror the vault-kind door's two
// spellings.
const (
	VaultKindWebsites = "websites"
	VaultKindRepos    = "repos"
)

// FrontmatterReadRunes mirrors _FRONTMATTER_READ_BYTES (16,000): the
// cheap head read — Python's text-mode read(16000) counts CHARACTERS
// (the constant is named BYTES in the reference but the read is the
// text-mode one; the spec's migration note preserves the read
// as-is). Named Runes here to keep the Go side honest about what the
// reference actually does.
const FrontmatterReadRunes = 16000

// WeightTopK mirrors WEIGHT_TOP_K: how many candidate homes ride one
// digest line.
const WeightTopK = 3

// WeightTieMargin mirrors WEIGHT_TIE_MARGIN: when the top two weighed
// homes sit within this margin the digest says CLOSE CALL.
const WeightTieMargin = 0.08

// StrongEvidenceWeight mirrors STRONG_EVIDENCE_WEIGHT: a top home at
// or above this weight is STRONG evidence.
const StrongEvidenceWeight = 0.55

// OverrideMargin mirrors OVERRIDE_MARGIN: how much lighter the chosen
// home may sit under a STRONG top home before the override reads as
// against-the-evidence.
const OverrideMargin = 0.15

// LockKeys mirrors _LOCK_KEYS: the lock's metadata keys (the
// frontmatter doors) — either spelled key reads as a lock when its
// value is truthy.
var LockKeys = []string{"locked", "category_locked"}

// LockTagWords mirrors _LOCK_TAG_WORDS: a frontmatter tag whose folded
// spelling is one of these locks the note.
var LockTagWords = map[string]bool{"lock": true, "locked": true}

// CuratedKeys mirrors _CURATED_KEYS: the census stamp's metadata key.
var CuratedKeys = []string{"curated"}

// CuratedTagWords mirrors _CURATED_TAG_WORDS: the census stamp's tag
// door (folded).
var CuratedTagWords = map[string]bool{"curated": true}

// taxonomyStopwords mirrors _TAXONOMY_STOPWORDS: the taxonomy pass's
// vocabulary stop set — definition words that name no category.
var taxonomyStopwords = map[string]bool{}

// safeFolderNameRE mirrors _SAFE_FOLDER_NAME: a folder-name cell the
// validator accepts (Title-Case-ish words, hyphens/underscores/
// spaces, no path tricks).
var safeFolderNameRE = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9 _\-&+()']{0,39}$`)

func init() {
	for _, w := range []string{
		"and", "the", "for", "with", "that", "this", "are", "not", "but",
		"from", "into", "your", "you", "its", "their", "them", "they",
		"any", "all", "own", "same", "other", "more", "most", "less",
		"via", "etc", "when", "where", "which", "while", "who", "whom",
		"how", "why", "what", "than", "then", "also", "can", "will",
		"was", "were", "has", "have", "had", "does", "did", "being",
		"been", "because", "such", "both", "each", "few", "some", "none",
		"one", "two", "use", "used", "uses", "using", "like", "just",
		"now", "new", "old", "see", "seen", "per", "pre", "pro", "non",
		"out", "off", "own", "too", "very", "keep", "kept", "made",
		"make", "makes", "get", "gets", "got", "let", "lets",
	} {
		taxonomyStopwords[w] = true
	}
}

// fsExists / fsIsDir / fsIsFile are the stat seam: the reference
// calls os.path.exists / isdir / isfile, and the ported
// _windows_eyes test law patches them to answer as a
// case-insensitive Windows disk would (the owner's machine — the
// two-notes loop's second face replayed on POSIX).
var (
	fsExists = func(p string) bool {
		_, err := os.Stat(p)
		return err == nil
	}
	fsIsDir = func(p string) bool {
		st, err := os.Stat(p)
		return err == nil && st.IsDir()
	}
	fsIsFile = func(p string) bool {
		st, err := os.Stat(p)
		return err == nil && !st.IsDir()
	}
)

type LogFn = func(msg, level string)

func noopLog(string, string) {}

// ProgressFn mirrors the progress(label, value, maximum) callable.
type ProgressFn = func(label string, value, maximum int)

// protectedFoldersFor mirrors _protected_folders: the vault kind's own
// no-touch list. Pure; junk-tolerant (an unknown kind reads as
// websites — the pre-v0.83 law).
func protectedFoldersFor(kind string) []string {
	if isReposKind(kind) {
		return ProtectedFoldersRepos
	}
	return ProtectedFolders
}

// isReposKind mirrors _is_repos_kind. Pure; junk-tolerant.
func isReposKind(kind string) bool {
	return strings.ToLower(helpers.Strip(kind)) == VaultKindRepos
}

// kindLabel mirrors _kind_label: the kind's own name for the story the
// log tells. Pure.
func kindLabel(kind string) string {
	if isReposKind(kind) {
		return "GitHub Repos vault"
	}
	return "Websites vault"
}

// repoURLKey mirrors _repo_url_key: the canonical repo URL
// (links.NormalizeURL — trailing slash, query and fragment gone; the
// same grammar every repos door uses). Never raises.
func repoURLKey(url string) string {
	return links.NormalizeURL(url)
}

// reposCategoryReverse mirrors _repos_category_reverse: the
// CATEGORY_FOLDERS reverse map (folder → category key). Pure; never
// raises.
func reposCategoryReverse() map[string]string {
	out := map[string]string{}
	for _, e := range constants.CategoryFolders {
		out[e.Folder] = e.Key
	}
	return out
}

// reposCategoryFolders mirrors _repos_category_folders: the
// CATEGORY_FOLDERS forward map (category key → folder). Pure; never
// raises.
func reposCategoryFolders() map[string]string {
	out := map[string]string{}
	for _, e := range constants.CategoryFolders {
		out[e.Key] = e.Folder
	}
	return out
}

// truthy mirrors _truthy: a frontmatter value's truth — true/yes/1
// (case- and quote-tolerant — the string-scan house law, never a YAML
// dependency). The strip chain is the reference's own: strip, strip
// quotes, strip apostrophes, strip again, lower.
func truthy(val string) bool {
	v := strings.ToLower(helpers.Strip(
		helpers.StripChars(helpers.StripChars(
			helpers.Strip(val), `"`), `'`)))
	return v == "true" || v == "yes" || v == "1"
}

// noteLocked mirrors _note_locked (v0.73.0 — THE LOCK LAW's
// predicate): three frontmatter doors, one verdict — the locked: key,
// the category_locked: key, and a lock/locked tag. Pure;
// junk-tolerant; never raises.
func noteLocked(fm *notereads.Inventory) bool {
	if fm == nil {
		return false
	}
	for _, key := range LockKeys {
		v := ""
		switch key {
		case "locked":
			v = pyBoolWord(fm.Locked)
		case "category_locked":
			v = pyBoolWord(fm.CategoryLocked)
		}
		if truthy(v) {
			return true
		}
	}
	for _, t := range fm.Tags {
		if LockTagWords[foldTag(t)] {
			return true
		}
	}
	return false
}

// pyBoolWord renders a parsed boolean back to its word (the predicate
// folds the word through truthy again — the unified parse already
// folded, so this rides the same verdict).
func pyBoolWord(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// noteCurated mirrors _note_curated (v0.78.0 — THE CENSUS STAMP's
// predicate): the curated: key (the stamp the scan itself writes) and
// a curated tag (the tags door). Pure; junk-tolerant; never raises.
func noteCurated(fm *notereads.Inventory) bool {
	if fm == nil {
		return false
	}
	if truthy(pyBoolWord(fm.Curated)) {
		return true
	}
	for _, t := range fm.Tags {
		if CuratedTagWords[foldTag(t)] {
			return true
		}
	}
	return false
}

// parseNoteFrontmatter mirrors _parse_note_frontmatter: the scan's own
// frontmatter read (the same string-scan style as the banishment's
// parser — the house law: core never grows a YAML dependency). The
// bounded head read is the text-mode 16,000-character one (translated
// newlines, U+FFFD maximal-subpart decode — the T-41 text-mode laws).
// The grammar rides vault_scan's OWN parser, not note_reads' unified
// one: the reference keeps both, with their micro-differences (flow
// tags REPLACE where the unified parse appends; the strip chain ends
// at the quote-strip without a final strip) — preserved as-is.
func parseNoteFrontmatter(path string) notereads.Inventory {
	out := notereads.Inventory{Tags: []string{}}
	data, err := os.ReadFile(path)
	if err != nil {
		return out
	}
	content := helpers.PyNewlines(helpers.DecodeUTF8Replace(data))
	content = helpers.CutRunes(content, FrontmatterReadRunes)
	lines := helpers.SplitLines(content)
	if len(lines) == 0 || helpers.Strip(lines[0]) != "---" {
		return out
	}
	inTags := false
	for _, line := range lines[1:] {
		s := helpers.Strip(line)
		if s == "---" {
			break
		}
		if inTags && strings.HasPrefix(s, "- ") {
			out.Tags = append(out.Tags,
				helpers.StripChars(helpers.StripChars(
					helpers.Strip(s[2:]), `"`), `'`))
			continue
		}
		inTags = false
		if !strings.Contains(s, ":") {
			continue
		}
		key, val, _ := strings.Cut(s, ":")
		key = strings.ToLower(helpers.Strip(key))
		val = helpers.StripChars(helpers.StripChars(
			helpers.Strip(val), `"`), `'`)
		switch key {
		case "tags":
			inner := val
			if strings.HasPrefix(inner, "[") && strings.HasSuffix(inner, "]") {
				inner = inner[1 : len(inner)-1]
			} else if inner == "" {
				inTags = true
			}
			tags := []string{}
			for _, t := range strings.Split(inner, ",") {
				t = helpers.Strip(t)
				if t == "" {
					continue
				}
				tags = append(tags, helpers.StripChars(
					helpers.StripChars(t, `"`), `'`))
			}
			out.Tags = tags // flow style REPLACES (vault_scan's own law)
		case "source":
			out.Source = val
		case "category":
			out.Category = val
		case "subcategory":
			out.Subcategory = val
		case "managed_by":
			out.ManagedBy = val
		case "kind":
			out.Kind = val
		case "fetch_status":
			out.FetchStatus = val
		case "primary_language":
			out.PrimaryLanguage = val
		case "org":
			out.Org = val
		case "stars":
			out.Stars = val
		case "locked", "category_locked":
			if truthy(val) {
				if key == "locked" {
					out.Locked = true
				} else {
					out.CategoryLocked = true
				}
			}
		case "curated":
			out.Curated = truthy(val)
		}
	}
	return out
}

// pyURLParseNetloc mirrors CPython 3.12's urlparse(url).netloc for the
// inventory's domain cell: the \t\r\n removal (the unsafe-bytes law),
// the case-insensitive scheme, the netloc between "//" and the first
// of "/?#" (case PRESERVED — netloc is not lowercased), and the
// bracket law (unbalanced brackets raise in CPython — the reference's
// except reads "no parse", so an invalid netloc answers "").
func pyURLParseNetloc(rawURL string) string {
	u := strings.Map(func(r rune) rune {
		if r == '\t' || r == '\r' || r == '\n' {
			return -1
		}
		return r
	}, rawURL)
	low := strings.ToLower(u)
	var rest string
	switch {
	case strings.HasPrefix(low, "https://"):
		rest = u[len("https://"):]
	case strings.HasPrefix(low, "http://"):
		rest = u[len("http://"):]
	default:
		return ""
	}
	netloc := rest
	if i := strings.IndexAny(netloc, "/?#"); i >= 0 {
		netloc = netloc[:i]
	}
	if !chrometabs.ValidPyNetloc(netloc) {
		return "" // the bracket law — the reference's except path
	}
	return netloc
}

// pyURLParsePath mirrors CPython 3.12's urlparse(url).path for the
// consolidation keeper rank: the path after netloc, query and fragment
// stripped, then _splitparams (the last path segment's ';params' cut —
// urlparse's legacy for http/https). Unparseable shapes (the bracket
// law's ValueError included) answer "z", the reference's own except
// sentinel.
func pyURLParsePath(rawURL string) string {
	u := strings.Map(func(r rune) rune {
		if r == '\t' || r == '\r' || r == '\n' {
			return -1
		}
		return r
	}, rawURL)
	low := strings.ToLower(u)
	var rest string
	switch {
	case strings.HasPrefix(low, "https://"):
		rest = u[len("https://"):]
	case strings.HasPrefix(low, "http://"):
		rest = u[len("http://"):]
	default:
		return "z"
	}
	path := ""
	if i := strings.IndexAny(rest, "/?#"); i >= 0 {
		path = rest[i:]
	}
	if !chrometabs.ValidPyNetloc(pyNetlocOf(rest)) {
		return "z" // the bracket law — the reference's except sentinel
	}
	if i := strings.Index(path, "#"); i >= 0 {
		path = path[:i] // the fragment split
	}
	if i := strings.Index(path, "?"); i >= 0 {
		path = path[:i] // the query split
	}
	// _splitparams: the ';' lives in the last segment or nowhere.
	if i := strings.Index(path, ";"); i >= 0 {
		if j := strings.LastIndex(path, "/"); j < 0 || i > j {
			path = path[:i]
		}
	}
	return path
}

// pyNetlocOf is the raw netloc slice of a post-scheme rest (up to the
// first of "/?#").
func pyNetlocOf(rest string) string {
	if i := strings.IndexAny(rest, "/?#"); i >= 0 {
		return rest[:i]
	}
	return rest
}

// isCoreNote mirrors _is_core_note (v0.63.0 — THE CORE-NOTE LAW's
// predicate): the Website Directory (the canonical catalog by its
// exact name, then any file whose name ends in "website directory"
// — the owner's spelling survives an emoji or numbering change) and
// every kind: directory note are the library's own catalog — never
// move candidates.
func isCoreNote(name string, fm *notereads.Inventory) bool {
	if name == websitepipeline.DirectoryFilename {
		return true // the canonical catalog, by its exact name
	}
	if strings.HasSuffix(strings.ToLower(name), "website directory.md") {
		return true // the same catalog under the owner's spelling
	}
	if fm != nil {
		return strings.ToLower(helpers.Strip(fm.Kind)) == "directory"
	}
	return false
}

// NoteRow mirrors the inventory's per-note rows.
type NoteRow struct {
	File        string
	Path        string
	Rel         string
	Title       string
	Source      string
	Domain      string
	Category    string
	Subcategory string
	Tags        []string
	FetchStatus string
	AppOwned    bool
	Folder      string
	InRoot      bool
	Core        bool
	Locked      bool
	Curated     bool
	Stars       int
	PrimaryLang string
	Org         string
	Protected   bool
	Homes       []WeighedHome `json:"-"`
	HomesClose  bool          `json:"-"`
}

// FolderCount mirrors the inventory's folders rows.
type FolderCount struct {
	Rel       string
	NoteCount int
}

// Inventory mirrors scan_vault_inventory's dict.
type Inventory struct {
	Folders            []FolderCount
	Notes              []NoteRow
	RootNotes          int
	UncategorizedNotes int
	HandNotes          int
	TrashNotes         int
	CoreNotes          int
	LockedNotes        int
	CuratedNotes       int
	TotalNotes         int
}

// ScanVaultInventory mirrors scan_vault_inventory (v0.61.0 — the
// folder walk + per-note digest, the LLM's eyes): dot-folders and
// _inbox skipped; _review/_moc notes SEEN but never candidates; the
// Trash waiting room counted but never listed; core notes marked;
// locked and curated notes counted (the census law); the repos kind
// carries the repo notes' own metadata. Pure file reads; never
// raises.
func ScanVaultInventory(vaultPath string, log LogFn, progress ProgressFn,
	readCache *notereads.Cache, kind string) Inventory {
	if log == nil {
		log = noopLog
	}
	out := Inventory{}
	if vaultPath == "" || !fsIsDir(vaultPath) {
		return out
	}
	// the cheap pre-count: how many .md the walk will open (listdir
	// only, no file reads) so the progress row has a REAL maximum.
	totalEstimate := 0
	if progress != nil {
		totalEstimate = countWalkMD(vaultPath)
	}
	prog := func(seen int) {
		if progress == nil {
			return
		}
		func() {
			defer func() { recover() }() // the reference's try/except
			base := filepath.Base(vaultPath)
			if base == "" {
				base = "vault"
			}
			progress("reading "+base, seen, max(totalEstimate, 1))
		}()
	}
	folderCounts := map[string]int{}
	folderOrder := []string{}
	seen := 0
	protected := protectedFoldersFor(kind)
	protectedPrefixes := make([]string, len(protected))
	for i, f := range protected {
		protectedPrefixes[i] = f + "/"
	}
	isProtectedRel := func(rel string) bool {
		for _, f := range protected {
			if rel == f {
				return true
			}
		}
		for _, p := range protectedPrefixes {
			if strings.HasPrefix(rel, p) {
				return true
			}
		}
		return false
	}
	walkInventory(vaultPath, "", func(path, relRoot string) {
		name := filepath.Base(path)
		seen++
		if seen%25 == 0 {
			prog(seen)
		}
		var fm notereads.Inventory
		if readCache != nil {
			fm = readCache.Frontmatter(path)
		} else {
			fm = parseNoteFrontmatter(path)
		}
		src := helpers.Strip(fm.Source)
		domain := ""
		if strings.HasPrefix(strings.ToLower(src), "http://") ||
			strings.HasPrefix(strings.ToLower(src), "https://") {
			domain = pyURLParseNetloc(src)
		}
		appOwned := strings.ToLower(helpers.Strip(fm.ManagedBy)) ==
			constants.ManagedByGitcurator
		core := isCoreNote(name, &fm)
		locked := noteLocked(&fm)
		curated := noteCurated(&fm)
		if websitepipeline.IsTrashFolder(relRoot) {
			out.TrashNotes++
			return
		}
		inRoot := relRoot == ""
		folderRel := ""
		if !inRoot {
			folderRel = relRoot
		}
		if folderRel != "" && !isProtectedRel(folderRel) {
			if _, ok := folderCounts[folderRel]; !ok {
				folderOrder = append(folderOrder, folderRel)
			}
			folderCounts[folderRel]++
		}
		if inRoot {
			out.RootNotes++
		}
		lowFolder := strings.ToLower(folderRel)
		if lowFolder == "uncategorized" ||
			strings.HasPrefix(lowFolder, "uncategorized/") {
			out.UncategorizedNotes++
		}
		if !appOwned {
			out.HandNotes++
		}
		if core {
			out.CoreNotes++
		}
		if locked {
			out.LockedNotes++
		}
		if curated {
			out.CuratedNotes++
		}
		out.TotalNotes++
		starsVal := 0
		if s := helpers.Strip(fm.Stars); s != "" {
			if n, ok := pyIntLoose(s); ok {
				starsVal = n
			}
		}
		tags := fm.Tags
		if tags == nil {
			tags = []string{}
		}
		rel := name
		if !inRoot {
			rel = folderRel + "/" + name
		}
		out.Notes = append(out.Notes, NoteRow{
			File: name, Path: path, Rel: rel,
			Title:       strings.TrimSuffix(name, filepath.Ext(name)),
			Source:      src,
			Domain:      domain,
			Category:    helpers.Strip(fm.Category),
			Subcategory: helpers.Strip(fm.Subcategory),
			Tags:        tags,
			FetchStatus: helpers.Strip(fm.FetchStatus),
			AppOwned:    appOwned, Folder: folderRel,
			InRoot: inRoot, Core: core, Locked: locked,
			Curated:     curated,
			Stars:       starsVal,
			PrimaryLang: helpers.Strip(fm.PrimaryLanguage),
			Org:         helpers.Strip(fm.Org),
			Protected:   isProtectedRel(folderRel),
		})
	})
	// sorted(folder_counts.items()) — the reference sorts the folder
	// rows by rel.
	rels := make([]string, 0, len(folderCounts))
	for rel := range folderCounts {
		rels = append(rels, rel)
	}
	sort.Strings(rels)
	for _, rel := range rels {
		out.Folders = append(out.Folders, FolderCount{
			Rel: rel, NoteCount: folderCounts[rel]})
	}
	if seen > 0 {
		prog(seen) // the walk's closing count
	}
	if out.TrashNotes > 0 {
		log("🗑️ "+strconv.Itoa(out.TrashNotes)+" note(s) rest in the Trash folder — "+
			"they await the deletion review (the same gate as your "+
			"auto-delete marks)", "info")
	}
	if out.CoreNotes > 0 {
		log("🏛️ "+strconv.Itoa(out.CoreNotes)+" core catalog note(s) ("+
			kindLabel(kind)+": the directory and its kin) stay at their "+
			"post — core notes are never re-filed", "info")
	}
	if out.LockedNotes > 0 {
		log("🔒 "+strconv.Itoa(out.LockedNotes)+" locked note(s) are out of the "+
			"scan's radar — never moved, never merged, never proposed "+
			"(locked: true / category_locked: true / the lock or "+
			"locked tag in the frontmatter; remove the lock by hand "+
			"in Obsidian to let a note ride again)", "info")
	}
	if out.CuratedNotes > 0 {
		log("🏷️ "+strconv.Itoa(out.CuratedNotes)+" curated note(s) are skipped — "+
			"the scan remembers its own work (they were moved, settled "+
			"or already reviewed by a past scan; curated: true in the "+
			"frontmatter — remove the line by hand in Obsidian and the "+
			"note rides the next scan again)", "info")
	}
	return out
}

// walkInventory mirrors the reference's pruned os.walk: dirs sorted,
// '_inbox' and dot-folders skipped, files sorted, files of a directory
// handled BEFORE its subdirectories (the house walk law).
func walkInventory(vaultPath string, relRoot string,
	visit func(path, relRoot string)) {
	entries, err := os.ReadDir(vaultPath)
	if err != nil {
		return // never raises — the walk's own law
	}
	files := make([]string, 0, len(entries))
	subdirs := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			if e.Name() == "_inbox" || strings.HasPrefix(e.Name(), ".") {
				continue
			}
			subdirs = append(subdirs, e.Name())
			continue
		}
		files = append(files, e.Name())
	}
	sort.Strings(files)
	sort.Strings(subdirs)
	for _, name := range files {
		if !strings.HasSuffix(strings.ToLower(name), ".md") {
			continue
		}
		path := filepath.Join(vaultPath, name)
		if !fsIsFile(path) {
			continue
		}
		visit(path, relRoot)
	}
	for _, sub := range subdirs {
		subRel := sub
		if relRoot != "" {
			subRel = relRoot + "/" + sub
		}
		walkInventory(filepath.Join(vaultPath, sub), subRel, visit)
	}
}

// countWalkMD mirrors the pre-count walk (listdir only).
func countWalkMD(vaultPath string) int {
	total := 0
	type dirEntry struct{ path string }
	stack := []dirEntry{{vaultPath}}
	for len(stack) > 0 {
		d := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		entries, err := os.ReadDir(d.path)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() {
				if e.Name() == "_inbox" || strings.HasPrefix(e.Name(), ".") {
					continue
				}
				stack = append(stack, dirEntry{filepath.Join(d.path, e.Name())})
				continue
			}
			if strings.HasSuffix(strings.ToLower(e.Name()), ".md") {
				total++
			}
		}
	}
	return total
}

// pyIntLoose mirrors Python's int() on a string for the stars cell:
// optional sign, digits with SINGLE underscores between digit groups
// (PEP 515 — int('42_000') is 42000), surrounding whitespace already
// stripped by the caller. False when the shape is not an int (the
// reference's except → 0).
func pyIntLoose(s string) (int, bool) {
	if s == "" {
		return 0, false
	}
	body := s
	if body[0] == '+' || body[0] == '-' {
		body = body[1:]
	}
	if body == "" {
		return 0, false
	}
	var b strings.Builder
	prevDigit := false
	for i := 0; i < len(body); i++ {
		c := body[i]
		if c >= '0' && c <= '9' {
			b.WriteByte(c)
			prevDigit = true
			continue
		}
		if c == '_' && prevDigit && i+1 < len(body) &&
			body[i+1] >= '0' && body[i+1] <= '9' {
			prevDigit = false // a single underscore between digits
			continue
		}
		return 0, false
	}
	sign := ""
	if s[0] == '+' || s[0] == '-' {
		sign = string(s[0])
	}
	n, err := strconv.Atoi(sign + b.String())
	if err != nil {
		return 0, false
	}
	return n, true
}

// MoveCandidates mirrors _move_candidates: which notes the LLM may
// propose moving — app-owned, not protected, not core, not locked,
// not curated, and either orphaned (vault root), uncategorized, a
// stray at a category root, or filed in a TOO-BROAD folder.
func MoveCandidates(inv *Inventory, tooBroad int) []NoteRow {
	broad := map[string]bool{}
	for _, f := range inv.Folders {
		if f.NoteCount > tooBroad {
			broad[f.Rel] = true
		}
	}
	var out []NoteRow
	for _, n := range inv.Notes {
		if !n.AppOwned || n.Protected || n.Core {
			continue
		}
		if n.Locked {
			continue // THE LOCK LAW: out of the radar
		}
		if n.Curated {
			continue // THE CENSUS STAMP: already verdicted
		}
		folder := n.Folder
		if n.InRoot || folder == "" {
			out = append(out, n) // orphaned — the vault root
			continue
		}
		low := strings.ToLower(folder)
		if low == "uncategorized" || strings.HasPrefix(low, "uncategorized/") {
			out = append(out, n)
			continue
		}
		if !strings.Contains(folder, "/") {
			out = append(out, n) // a stray at the category root
			continue
		}
		if broad[folder] {
			out = append(out, n)
			continue
		}
		for b := range broad {
			if strings.HasPrefix(folder, b+"/") {
				out = append(out, n)
				break
			}
		}
	}
	return out
}
