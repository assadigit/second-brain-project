// Package vaulthygiene ports core/vault_hygiene.py — v0.68.0 THE
// STRAY-NOTES DOOR (the shared pass), widened by v0.70.0 (THE LINKED-
// GHOST DOOR + THE APP-OWNED BODYLESS FINGERPRINT + THE ONE WALK) and
// v0.71.0 (THE ONE-OPEN LAW's shared cache) and guarded by v0.73.0 (THE
// LOCK LAW's hygiene predicate).
//
// Spec: docs/port-blueprint/module-specs/spec-core-vault-hygiene.md;
// task T-46 (Phase 4). The reference's collect_real_note_identities
// (341-389) is DEAD upstream (superseded by the one walk) and is not
// ported — its observable behavior lives inside RunVaultHygiene's walk,
// where the identity harvest, the stub candidacy and the link targets
// are harvested from ONE bounded read per file.

package vaulthygiene

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/assadigit/second-brain-project/internal/constants"
	"github.com/assadigit/second-brain-project/internal/dryrun"
	"github.com/assadigit/second-brain-project/internal/helpers"
	"github.com/assadigit/second-brain-project/internal/notereads"
	"github.com/assadigit/second-brain-project/internal/storage"
)

// HygieneSkipDirs mirrors HYGIENE_SKIP_DIRS: the folders the hygiene
// walk never enters (system / hidden / app-machinery folders — the same
// set the rest of the vault walks honor, plus the Library mirror tree).
// Exact-name membership: an unlisted dot-folder (e.g. ".stfolder") IS
// walked, exactly as the reference's `d not in HYGIENE_SKIP_DIRS` does.
var HygieneSkipDirs = []string{
	".obsidian", ".git", ".trash", "_missing", "_moc", "_inbox",
	"attachments", "Library", "__pycache__",
}

var hygieneSkipSet = func() map[string]bool {
	m := make(map[string]bool, len(HygieneSkipDirs))
	for _, d := range HygieneSkipDirs {
		m[d] = true
	}
	return m
}()

// StubMaxBytes mirrors STUB_MAX_BYTES (256): the empty-stub door's
// documented size ceiling. The classifier itself judges by SHAPE alone
// (blank lines + at most one heading); the constant is the GUI mixin's
// parity surface, kept exported exactly as the reference keeps it.
const StubMaxBytes = 256

// AppOwnedBodylessMaxBytes mirrors APP_OWNED_BODYLESS_MAX_BYTES
// (16,384): the app-owned bodyless door's documented frontmatter
// ceiling. Same law as StubMaxBytes — the classifier judges by proof
// (managed_by) + shape (the empty body); the constant is the parity
// surface.
const AppOwnedBodylessMaxBytes = 16384

// walkReadMaxRunes mirrors _WALK_READ_MAX_BYTES (1,048,576): how much
// of each note the one walk reads — identities live in the head,
// wiki-links live anywhere, whole notes are read (bounded) so the
// linked-ghost fingerprint sees every [[link]]. Python's text-mode
// read(n) counts CHARACTERS — the bound is runes here (T-45's lesson;
// the reference's constant is named BYTES but the text read counts
// chars). Notes bigger than this cap contribute their stem only.
const walkReadMaxRunes = 1048576

// progressEvery mirrors _PROGRESS_EVERY (25): how often the one walk
// speaks its progress (files walked).
const progressEvery = 25

// wikilinkRE mirrors _WIKILINK_RE: one [[wiki-link]] — [[target]],
// [[target|alias]], [[target#anchor]], [[target#anchor|alias]]; embeds
// (![[…]]) match too (the inner brackets).
var wikilinkRE = regexp.MustCompile(
	`\[\[([^\[\]<>|]+?)(?:#[^\[\]<>|]*)?(?:\|[^\[\]]*)?\]\]`)

// lockTagWords mirrors _LOCK_TAG_WORDS: the lock tag door's vocabulary
// (vault_scan's own grammar, spelled here so the module stays
// import-light — the fold matches vault_scan._fold_tag's identity law).
var lockTagWords = map[string]bool{"lock": true, "locked": true}

// foldNonAlnumRE mirrors the identity fold `[^a-z0-9]+`.
var foldNonAlnumRE = regexp.MustCompile(`[^a-z0-9]+`)

// LogFn is the injected log (msg, level); ProgressFn is the optional
// progress callable (label, value, maximum).
type LogFn = func(msg, level string)
type ProgressFn = func(label string, value, maximum int)

func noopLog(string, string) {}

// ---------------------------------------------------------------------------
// The classifiers (pure — the worker mixin's classmethods delegate here)
// ---------------------------------------------------------------------------

// IsEmptyStub mirrors is_empty_stub: True when a note is content-free —
// no frontmatter, blank lines and at most one heading line (the shape
// Obsidian leaves when a ghost [[wiki-link]] is clicked into
// existence). Anything richer — frontmatter, body text, a list — is
// somebody's writing and is never touched. The reference's None door
// reads True; Go's zero string carries the same verdict.
func IsEmptyStub(content string) bool {
	if strings.HasPrefix(content, "---") {
		return false // frontmatter = a structured file, never ours to judge
	}
	headings := 0
	for _, line := range helpers.SplitLines(content) {
		stripped := strings.TrimSpace(line)
		if stripped == "" {
			continue
		}
		if strings.HasPrefix(stripped, "#") {
			headings++
			if headings > 1 {
				return false
			}
			continue
		}
		return false // real content — somebody's writing
	}
	return true
}

// IsAppOwnedBodyless mirrors is_app_owned_bodyless (v0.70.0): True when
// a note PROVES the app wrote it (managed_by: GitCurator in the
// frontmatter) and its body never landed — after the frontmatter block
// there is nothing but blank lines and at most one heading line. The
// ownership PROOF is what makes it ours to judge (an owner-written
// frontmatter note with an empty body stays sacred — never touched).
func IsAppOwnedBodyless(content string) bool {
	if !strings.HasPrefix(content, "---") {
		return false // no frontmatter → the empty-stub door's call
	}
	lines := helpers.SplitLines(content)
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return false
	}
	// walk the frontmatter block; find the app's ownership mark
	i := 1
	owned := false
	for i < len(lines) {
		stripped := strings.TrimSpace(lines[i])
		if stripped == "---" || stripped == "..." {
			i++
			break
		}
		key, val, _ := strings.Cut(stripped, ":")
		if strings.TrimSpace(strings.ToLower(key)) == "managed_by" &&
			strings.ToLower(strings.Trim(strings.TrimSpace(val), `"'`)) ==
				constants.ManagedByGitcurator {
			owned = true
		}
		i++
	}
	if !owned {
		return false // the owner's own frontmatter — sacred
	}
	// the body: blank lines + at most one heading, nothing else
	headings := 0
	for _, line := range lines[i:] {
		stripped := strings.TrimSpace(line)
		if stripped == "" {
			continue
		}
		if strings.HasPrefix(stripped, "#") {
			headings++
			if headings > 1 {
				return false
			}
			continue
		}
		return false // real body — a finished note, never ours
	}
	return true
}

// IsLockedFrontmatter mirrors is_locked_frontmatter (v0.73.0 — THE LOCK
// LAW's hygiene predicate, PURE): does this note's frontmatter carry a
// lock door — a truthy `locked:` or `category_locked:` key, or a
// `lock`/`locked` tag (flow or block list, folded past case and
// separators)? A locked note is never retired by the hygiene pass,
// whatever its emptiness. The empty-stub door never reaches here (a
// stub carries NO frontmatter at all); this guards the app-owned
// bodyless door.
func IsLockedFrontmatter(content string) bool {
	if !strings.HasPrefix(content, "---") {
		return false
	}
	lines := helpers.SplitLines(content)
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return false
	}
	inTags := false
	for _, line := range lines[1:] {
		s := strings.TrimSpace(line)
		if s == "---" {
			break
		}
		if inTags && strings.HasPrefix(s, "- ") {
			if lockTagWords[foldLockWord(s[2:])] {
				return true
			}
			continue
		}
		inTags = false
		if !strings.Contains(s, ":") {
			continue
		}
		key, val, _ := strings.Cut(s, ":")
		key = strings.TrimSpace(strings.ToLower(key))
		// val.strip().strip('"').strip("'").strip().lower()
		val = strings.TrimSpace(strings.ToLower(
			strings.TrimSpace(strings.Trim(
				strings.Trim(strings.TrimSpace(val), `"`), "'"))))
		if (key == "locked" || key == "category_locked") &&
			(val == "true" || val == "yes" || val == "1") {
			return true
		}
		if key == "tags" {
			inner := val
			if strings.HasPrefix(inner, "[") && strings.HasSuffix(inner, "]") {
				inner = inner[1 : len(inner)-1]
			} else if inner == "" {
				inTags = true // block style — items follow
			}
			for _, item := range strings.Split(inner, ",") {
				if lockTagWords[foldLockWord(item)] {
					return true
				}
			}
		}
	}
	return false
}

// foldLockWord mirrors _fold_lock_word: a tag's lock identity —
// lowercase alphanumerics only (Lock, lock and locked fold to their own
// words).
func foldLockWord(word string) string {
	return foldNonAlnumRE.ReplaceAllString(strings.ToLower(word), "")
}

// HasNoteNameShape mirrors has_note_name_shape: True when a filename
// carries the app's canonical note shape `<repo>_<Category>_<tag>.md` —
// a CATEGORY_KEY appears as an underscore-delimited segment. An owner's
// own filename (my-thoughts.md, ideas.md) never carries a category
// segment.
func HasNoteNameShape(fname string) bool {
	stem := fname
	if strings.HasSuffix(fname, ".md") {
		stem = fname[:len(fname)-3]
	}
	if !strings.Contains(stem, "_") {
		return false
	}
	for _, p := range strings.Split(stem, "_") {
		for _, k := range constants.CategoryKeys {
			if p == k {
				return true
			}
		}
	}
	return false
}

// NormalizeIdentity mirrors normalize_identity: collapse a name to its
// alphanumeric skeleton — the form where subtitle-translator, Subtitle
// Translator and subtitle_translator all meet (case-folded, everything
// that is not a letter or digit stripped). The reference's junk/None
// doors read ” (never raises); Go's zero string carries the same
// verdict.
func NormalizeIdentity(text string) string {
	return foldNonAlnumRE.ReplaceAllString(strings.ToLower(text), "")
}

// typeLineRE mirrors `^type:\s*(.+)$` (multiline, first match).
var typeLineRE = regexp.MustCompile(`(?m)^type:[ \t]*(.*)$`)

// frontmatterTypeOf mirrors _frontmatter_type_of: the `type:`
// frontmatter value of a note already read — ” when it has none (the
// head-scan style of the old path-reader, kept pure so the one walk
// never re-opens a file). The reference scans content[:400]
// (CHARACTERS); the cap is applied here exactly as there.
func frontmatterTypeOf(content string) string {
	if content == "" || !strings.HasPrefix(content, "---") {
		return ""
	}
	head := helpers.CutRunes(content, 400)
	if m := typeLineRE.FindStringSubmatch(head); m != nil {
		return strings.Trim(strings.TrimSpace(m[1]), `"'`)
	}
	return ""
}

// HarvestLinkTargets mirrors harvest_link_targets (v0.70.0): every
// [[wiki-link]] target in one note's body, normalized to identity form
// (the linked-ghost fingerprint's eyes). Aliases and anchors are
// stripped; path-shaped targets contribute their last segment; embeds
// (![[…]]) count too. ” bodies yield the empty set.
func HarvestLinkTargets(content string) map[string]bool {
	out := map[string]bool{}
	if content == "" {
		return out
	}
	for _, m := range wikilinkRE.FindAllStringSubmatch(content, -1) {
		raw := strings.TrimSpace(m[1])
		if raw == "" {
			continue
		}
		// raw.replace('\\', '/').rstrip('/').rsplit('/', 1)[-1]
		name := strings.ReplaceAll(raw, `\`, "/")
		name = strings.TrimRight(name, "/")
		if i := strings.LastIndex(name, "/"); i >= 0 {
			name = name[i+1:]
		}
		if ident := NormalizeIdentity(name); ident != "" {
			out[ident] = true
		}
	}
	return out
}

// ---------------------------------------------------------------------------
// The identity harvest — what REAL notes are called in this vault
// (the one walk's own half; the dead standalone collector is not ported)
// ---------------------------------------------------------------------------

// inlineAliasesRE mirrors `^aliases:\s*\[(.*)\]` (match at line start —
// no leading whitespace).
var inlineAliasesRE = regexp.MustCompile(`^aliases:[ \t]*\[(.*)\]`)

// bareAliasesRE mirrors `^aliases:\s*$`.
var bareAliasesRE = regexp.MustCompile(`^aliases:[ \t]*$`)

// aliasItemRE mirrors `^\s*-\s*(.+)$`.
var aliasItemRE = regexp.MustCompile(`^[ \t]*-[ \t]+(.*)$`)

// harvestHeadIdentities mirrors _harvest_head_identities: the first H1
// line + the frontmatter `aliases:` items of one note's head,
// normalized (the one walk's identity half — stems come from the
// caller).
func harvestHeadIdentities(content string) map[string]bool {
	out := map[string]bool{}
	lines := helpers.SplitLines(content)
	i := 0
	// ---- the frontmatter block (opening fence = line 0) --------------
	if len(lines) > 0 && strings.TrimSpace(lines[0]) == "---" {
		i = 1
		inAliases := false
		for i < len(lines) {
			stripped := strings.TrimSpace(lines[i])
			if stripped == "---" || stripped == "..." {
				i++
				break
			}
			if m := inlineAliasesRE.FindStringSubmatch(lines[i]); m != nil {
				for _, item := range strings.Split(m[1], ",") {
					if ident := NormalizeIdentity(
						strings.Trim(strings.TrimSpace(item), `"'`)); ident != "" {
						out[ident] = true
					}
				}
				inAliases = false
			} else if bareAliasesRE.MatchString(lines[i]) {
				inAliases = true
			} else if inAliases {
				if m := aliasItemRE.FindStringSubmatch(lines[i]); m != nil {
					if ident := NormalizeIdentity(
						strings.Trim(strings.TrimSpace(m[1]), `"'`)); ident != "" {
						out[ident] = true
					}
				} else {
					inAliases = false
				}
			}
			i++
		}
	}
	// ---- the first H1 after the frontmatter --------------------------
	for _, line := range lines[i:] {
		stripped := strings.TrimSpace(line)
		if strings.HasPrefix(stripped, "# ") {
			if ident := NormalizeIdentity(stripped[2:]); ident != "" {
				out[ident] = true
			}
			break // the FIRST H1 only
		}
	}
	return out
}

// ---------------------------------------------------------------------------
// The pass itself — one vault, any door
// ---------------------------------------------------------------------------

// HygieneCounts mirrors run_vault_hygiene's honest counts dict:
// removed_index / retired_stubs / retired_reports / stub_names (the
// stems, for the story) + v0.70.0's retired_linked and
// retired_app_owned (the new fingerprints' own counts, subsets of
// retired_stubs — counted at the DECISION, like the reference does,
// even when the move itself then fails).
type HygieneCounts struct {
	RemovedIndex    int
	RetiredStubs    int
	RetiredReports  int
	StubNames       []string
	RetiredLinked   int
	RetiredAppOwned int
}

// candidate mirrors the one walk's candidate rows: path, fname, stem
// (normalized), kind ('stub' | 'app_owned').
type candidate struct {
	path  string
	fname string
	stem  string
	kind  string
}

// RunVaultHygiene mirrors run_vault_hygiene: the one hygiene pass over
// ONE vault. THE ONE WALK: each .md file is opened and read exactly
// ONCE (identities + stub candidacy + link targets from the same read)
// — the optional readCache (a notereads.Cache) serves the walk's
// bounded reads, so one Scan click opens each note once across every
// pass that touches it (v0.71.0's one-open law); a nil cache takes the
// direct reads, byte-identical. Dry-run aware; never fails loud;
// nothing content-bearing is ever touched.
func RunVaultHygiene(vaultPath string, log LogFn, progress ProgressFn,
	readCache *notereads.Cache) HygieneCounts {
	if log == nil {
		log = noopLog
	}
	out := HygieneCounts{}
	if vaultPath == "" || !isDir(vaultPath) {
		return out
	}
	say := func(msg string) { log(msg, "info") }
	vaultBase := filepath.Base(vaultPath)
	if vaultBase == "" {
		vaultBase = "vault"
	}
	prog := func(value, maximum int) {
		if progress == nil {
			return
		}
		progress(fmt.Sprintf("healing %s", vaultBase), value, maximum)
	}
	retire := func(path, subdir string) bool {
		// Move one file into <vault>/.trash/<subdir>/ — bytes preserved,
		// name uniquified, dry-run aware.
		dstDir := filepath.Join(vaultPath, ".trash", subdir)
		if err := dryrun.MkdirAll(dstDir, true); err != nil {
			return false
		}
		dst := storage.UniquePath(filepath.Join(dstDir, filepath.Base(path)))
		return dryrun.Move(path, dst) == nil
	}

	// ---- 1. the retired master-index scaffold -------------------------
	removedIndex := 0
	indexPath := filepath.Join(vaultPath, "_index.md")
	if isFile(indexPath) {
		head := readHeadDirect(indexPath)
		if strings.HasPrefix(head, "---") &&
			frontmatterTypeOf(head) == "master-index" {
			_ = dryrun.Remove(indexPath)
			removedIndex++
		}
	}
	mocDir := filepath.Join(vaultPath, "_moc")
	if isDir(mocDir) {
		if entries, err := os.ReadDir(mocDir); err == nil {
			for _, e := range entries {
				if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
					continue
				}
				fpath := filepath.Join(mocDir, e.Name())
				if !isFile(fpath) {
					continue
				}
				head := readHeadDirect(fpath)
				if strings.HasPrefix(head, "---") &&
					frontmatterTypeOf(head) == "moc" {
					_ = dryrun.Remove(fpath)
					removedIndex++
				}
			}
			// prune the folder when the app's files were its only
			// residents (an owner file keeps the folder — sacred).
			if dryrun.IsEnabled() {
				dryrun.Record("remove", mocDir, nil,
					"delete the emptied _moc folder")
			} else if dirIsEmpty(mocDir) {
				_ = os.Remove(mocDir)
			}
		}
	}
	out.RemovedIndex = removedIndex

	// ---- 2. THE ONE WALK — identities, stubs, and the links -----------
	// One walk, one bounded read per file, three harvests: the real
	// notes name the twins, the content-free notes are the candidates,
	// and every [[wiki-link]] target names the ghosts.
	totalFiles := countMDFiles(vaultPath)
	walked := 0
	identities := map[string]bool{}
	linked := map[string]int{}
	var candidates []candidate
	walkVaultHygiene(vaultPath, func(fpath string) {
		fname := filepath.Base(fpath)
		walked++
		if walked%progressEvery == 0 {
			prog(walked, totalFiles)
		}
		st, err := os.Stat(fpath)
		if err != nil {
			return // the reference's OSError door — the file says nothing
		}
		stem, _ := storage.PySplitExt(fname)
		stemNorm := NormalizeIdentity(stem)
		if st.Size() > walkReadMaxRunes {
			// huge file: real writing — stem identity only
			if stemNorm != "" {
				identities[stemNorm] = true
			}
			return
		}
		var content string
		if readCache != nil {
			// v0.71.0 — the one-open law: the shared cache serves the
			// bounded read (the same read the scan's other walks ride).
			content = readCache.Text(fpath, walkReadMaxRunes+1)
		} else {
			var derr error
			content, derr = readTextBounded(fpath, walkReadMaxRunes+1)
			if derr != nil {
				return // the open failed — the reference's OSError skip
			}
		}
		// every [[link]] this note carries — the ghost map
		for tgt := range HarvestLinkTargets(content) {
			if tgt != stemNorm { // a self-link vouches for nothing
				linked[tgt]++
			}
		}
		if IsEmptyStub(content) {
			candidates = append(candidates, candidate{
				path: fpath, fname: fname, stem: stemNorm, kind: "stub"})
			return
		}
		if IsLockedFrontmatter(content) {
			// v0.73.0 — THE LOCK LAW: a locked note is never retired,
			// whatever its emptiness — the owner's lock outranks every
			// hygiene fingerprint; unlocking is his hand in Obsidian. Its
			// stem still vouches for its twins below (a locked note is a
			// REAL note) — the pass falls through to the harvest.
		} else if IsAppOwnedBodyless(content) {
			candidates = append(candidates, candidate{
				path: fpath, fname: fname, stem: stemNorm, kind: "app_owned"})
			return
		}
		// a REAL note: its names vouch for its twins
		if stemNorm != "" {
			identities[stemNorm] = true
		}
		for ident := range harvestHeadIdentities(content) {
			identities[ident] = true
		}
	})
	if totalFiles < 1 {
		prog(1, totalFiles)
	} else {
		prog(totalFiles, totalFiles)
	}

	// ---- the retirement decision (fingerprints, all three) -----------
	retiredStubs := 0
	retiredLinked := 0
	retiredAppOwned := 0
	var stubNames []string
	var stubTargets []string
	for _, cand := range candidates {
		stemNorm := cand.stem
		shaped := HasNoteNameShape(cand.fname)
		twin := stemNorm != "" && identities[stemNorm]
		ghost := stemNorm != "" && linked[stemNorm] > 0
		if !(shaped || twin || ghost) {
			// a content-free note that nobody links, that answers to no
			// real note and carries no app shape — the owner's own empty
			// file. Sacred.
			continue
		}
		stubTargets = append(stubTargets, cand.path)
		if ghost {
			retiredLinked++
		}
		if cand.kind == "app_owned" {
			retiredAppOwned++
		}
	}
	for _, fpath := range stubTargets {
		if retire(fpath, "empty-stubs") {
			retiredStubs++
			base := filepath.Base(fpath)
			stem, _ := storage.PySplitExt(base)
			stubNames = append(stubNames, stem)
		}
	}
	out.RetiredStubs = retiredStubs
	out.StubNames = stubNames
	out.RetiredLinked = retiredLinked
	out.RetiredAppOwned = retiredAppOwned

	// ---- 3. the legacy root reports ------------------------------------
	retiredReports := 0
	if entries, err := os.ReadDir(vaultPath); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			fname := e.Name()
			fpath := filepath.Join(vaultPath, fname)
			if strings.HasPrefix(fname, "_processing_report_") &&
				strings.HasSuffix(fname, ".md") && isFile(fpath) {
				if retire(fpath, "retired-reports") {
					retiredReports++
				}
			} else if strings.HasPrefix(fname, "processing_summary_") &&
				strings.HasSuffix(fname, ".txt") && isFile(fpath) {
				if retire(fpath, "retired-reports") {
					retiredReports++
				}
			}
		}
	}
	out.RetiredReports = retiredReports

	// ---- the honest summary ---------------------------------------------
	if removedIndex > 0 {
		say(fmt.Sprintf("🧹 Vault hygiene: the retired master index "+
			"(%d file(s) — _index.md/_moc) left the vault — its links no "+
			"longer weld your graph or block note deletions", removedIndex))
	}
	if retiredStubs > 0 {
		var why []string
		if retiredLinked > 0 {
			why = append(why, fmt.Sprintf("%d linked ghost(s) — empty "+
				"notes other notes pointed at", retiredLinked))
		}
		if retiredAppOwned > 0 {
			why = append(why, fmt.Sprintf("%d app-owned bodyless note(s) — "+
				"failed writes that never got their content", retiredAppOwned))
		}
		names := stubNames
		nameCell := strings.Join(names[:minInt(6, len(names))], ", ")
		if len(names) > 6 {
			nameCell += "…"
		}
		whyCell := ""
		if len(why) > 0 {
			whyCell = fmt.Sprintf(" (%s)", strings.Join(why, "; "))
		}
		say(fmt.Sprintf("🧹 Vault hygiene: %d stray empty note(s) retired "+
			"to .trash/empty-stubs — %s%s — recoverable by hand",
			retiredStubs, nameCell, whyCell))
	}
	if retiredReports > 0 {
		say(fmt.Sprintf("🧹 Vault hygiene: %d legacy report file(s) retired "+
			"from the vault root to .trash/retired-reports (new reports live "+
			"in the app's reports folder, not your vault)", retiredReports))
	}
	if removedIndex == 0 && retiredStubs == 0 && retiredReports == 0 {
		say("🧹 Vault hygiene: clean — no index scaffold, no stray empty " +
			"notes, no legacy reports in the vault.")
	}
	return out
}

// readTextBounded mirrors the reference's bounded text-mode read
// (open + read(n), errors='replace', universal newlines): the file is
// read whole (the walk's size gate already bounds it at the cap), the
// bytes decoded with the maximal-subpart U+FFFD law, newlines
// translated, then cut at limit runes. An open failure returns the
// error — the caller's OSError door.
func readTextBounded(path string, limit int) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return helpers.CutRunes(
		helpers.PyNewlines(helpers.DecodeUTF8Replace(data)), limit), nil
}

// readHeadDirect mirrors the scaffold's head read (f.read(400) text
// mode): ” on any failure (the reference's except OSError → ”).
func readHeadDirect(path string) string {
	content, err := readTextBounded(path, 400)
	if err != nil {
		return ""
	}
	return content
}

// walkVaultHygiene mirrors the reference's pruned os.walk: HYGIENE_
// SKIP_DIRS pruned by exact name, a directory's files handled BEFORE
// its subdirectories (the house walk law); ReadDir's sorted order is
// the Go house's deterministic stand-in for os.listdir order (the
// retirement decisions are order-independent — the identity and link
// sets complete before any decision is made).
func walkVaultHygiene(dir string, visit func(path string)) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return // never raises — the walk's own law
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if !strings.HasSuffix(e.Name(), ".md") {
			continue // case-sensitive, exactly as the reference's endswith
		}
		path := filepath.Join(dir, e.Name())
		if !isFile(path) {
			continue
		}
		visit(path)
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if hygieneSkipSet[e.Name()] {
			continue
		}
		walkVaultHygiene(filepath.Join(dir, e.Name()), visit)
	}
}

// countMDFiles mirrors _count_md_files: the cheap pre-count for the
// progress story (listdir only, no opens).
func countMDFiles(vaultPath string) int {
	total := 0
	var walk func(dir string)
	walk = func(dir string) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return
		}
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
				total++
			}
		}
		for _, e := range entries {
			if e.IsDir() && !hygieneSkipSet[e.Name()] {
				walk(filepath.Join(dir, e.Name()))
			}
		}
	}
	walk(vaultPath)
	return total
}

func dirIsEmpty(path string) bool {
	entries, err := os.ReadDir(path)
	return err == nil && len(entries) == 0
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ---------------------------------------------------------------------------
// The config-aware door — both machine vaults, any caller
// ---------------------------------------------------------------------------

// HealMachineVaults mirrors heal_machine_vaults: run the hygiene pass
// over BOTH machine vaults the config names (the GitHub projects vault
// vault_path + the Websites vault website_vault_path). Returns a map
// with an entry ONLY for each vault that exists (the reference's
// {'github': counts, 'websites': counts} — Go's key set carries the
// same law). The optional progress and readCache ride through.
func HealMachineVaults(config map[string]any, log LogFn,
	progress ProgressFn, readCache *notereads.Cache) map[string]HygieneCounts {
	healed := map[string]HygieneCounts{}
	if config == nil {
		return healed
	}
	for _, row := range [][2]string{
		{"github", "vault_path"},
		{"websites", "website_vault_path"},
	} {
		label, key := row[0], row[1]
		path := cfgStr(config, key)
		if path == "" || !isDir(path) {
			continue
		}
		vaultName := fmt.Sprintf("%s vault", label)
		vaultProgress := func(lbl string, value, maximum int) {
			// the vault's own name rides the label — the counter's
			// tooltip says exactly which vault is being walked.
			if progress != nil {
				progress(fmt.Sprintf("%s · %s", vaultName, lbl), value, maximum)
			}
		}
		healed[label] = RunVaultHygiene(path, log, vaultProgress, readCache)
	}
	return healed
}

// cfgStr reads a config string (Python str(v or ”).strip() — falsy
// values read ”).
func cfgStr(config map[string]any, key string) string {
	v, ok := config[key]
	if !ok || v == nil {
		return ""
	}
	switch vv := v.(type) {
	case string:
		return strings.TrimSpace(vv)
	case bool:
		if !vv {
			return ""
		}
		return "True"
	case int:
		if vv == 0 {
			return ""
		}
	case int64:
		if vv == 0 {
			return ""
		}
	case float64:
		if vv == 0 {
			return ""
		}
	}
	return strings.TrimSpace(helpers.PyStr(v))
}

// HygieneTotal mirrors hygiene_total: the one number the doors speak —
// how many files the pass retired across the vaults (scaffold + stubs
// + reports). Pure; a nil map reads 0 (the reference's junk-tolerance
// doors are type-narrowed away — Go's counts are typed ints).
func HygieneTotal(healed map[string]HygieneCounts) int {
	total := 0
	for _, counts := range healed {
		total += counts.RemovedIndex
		total += counts.RetiredStubs
		total += counts.RetiredReports
	}
	return total
}

func isDir(p string) bool {
	st, err := os.Stat(p)
	return err == nil && st.IsDir()
}

func isFile(p string) bool {
	st, err := os.Stat(p)
	return err == nil && !st.IsDir()
}
