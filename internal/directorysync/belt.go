// belt.go — the v0.89.0 LOCAL-ONLY BELT floor (T-47's surface of
// cloud/directory_sync.py): the pattern builders the vaultseal
// mirrors consume (local_only_note_patterns + private_exclusion_
// patterns + load_directory_config + their pure helpers).
//
// The owner's law, verbatim: "remove torrent and illegal or
// copy-righted download sites from github and site and cloudflare,
// and keep those only local." The site's half (the tombstones + the
// private belt) is directory_sync's own; these builders feed the
// VAULTSEAL mirrors — git never stages the belt's notes, and a
// one-time `git rm --cached` untracks what an older seal already
// pushed (the files stay ON DISK, "kept local").
//
// Landed WITH T-47 (the vaultseal config bridges consume them)
// exactly the way T-45 landed the repobanish floor with vaultscan;
// the REST of directory_sync (build_index, publish, the command
// cycle, the config trio) is T-49's own task.
package directorysync

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/assadigit/second-brain-project/assets"
	"github.com/assadigit/second-brain-project/internal/helpers"
	"github.com/assadigit/second-brain-project/internal/storage"
)

// DefaultPrivateNames mirrors DEFAULT_PRIVATE_NAMES: private
// category/subcategory NAMES — a second belt over the
// directory-config.json visibility marks (defense in depth, SPEC
// §12).
var DefaultPrivateNames = []string{"Torrent & Shadow Libraries"}

// DefaultLocalOnlyStems mirrors DEFAULT_LOCAL_ONLY_STEMS: the
// local-only NOTE STEMS, not paths — the seal's .gitignore matches
// them wherever the note lands in the vault (a renamed folder never
// re-leaks one).
var DefaultLocalOnlyStems = []string{
	"CakhiaTV",        // pirate live-football streaming (published 10-05)
	"1Music_cc",       // pirated FLAC/MP3 music downloads
	"cobalt",          // paywall-circumvention media downloader
	"Prowlarr",        // torrent-indexer manager (the *arr stack)
	"Max_Effect",      // pirated Adobe plugins/templates (maxeffect.ir)
	"Flac_Downloader", // Deezer/Qobuz stream ripper
}

// DirectoryConfigFilename mirrors DIRECTORY_CONFIG_FILENAME.
const DirectoryConfigFilename = "directory-config.json"

// prettyFoldRe mirrors the belt's [\\s_]+ fold.
var prettyFoldRe = regexp.MustCompile(`[\s_]+`)

// PrettifyFolderName mirrors _prettify_folder_name: THE MIRROR LAW's
// display floor — a vault folder no registry knows ('Icons_Svg_Assets')
// still publishes, with its underscores read as spaces ('Icons Svg
// Assets'). The same string rides the site records AND the category
// tree; when the owner later declares the folder in
// directory-config.json, the folder-name law translates to the
// declared display name instead and both sides move together.
func PrettifyFolderName(name string) string {
	text := strings.TrimSpace(prettyFoldRe.ReplaceAllString(strings.TrimSpace(name), " "))
	if text != "" {
		return text
	}
	return strings.TrimSpace(name)
}

// gitignoreEscape mirrors _gitignore_escape: escape a literal folder
// name for a .gitignore pattern — the glob chars ([]?*#!\) and a
// trailing space would change the meaning, and a '/' inside a name is
// a PATH separator in a pattern. Conservative backslash escapes, the
// gitignore way.
func gitignoreEscape(pattern string) string {
	var out strings.Builder
	for _, ch := range pattern {
		switch ch {
		case '[', ']', '?', '*', '#', '!', '\\':
			out.WriteRune('\\')
			out.WriteRune(ch)
		case '/':
			out.WriteRune('_')
		default:
			out.WriteRune(ch)
		}
	}
	text := out.String()
	if strings.HasSuffix(text, " ") {
		text = text[:len(text)-1] + "\\ "
	}
	return text
}

// beltFolds mirrors _belt_folds: the folded (lowercased) spellings
// every private name answers to — the name itself and its folder form
// (prettified safe_filename — 'Torrent & Shadow Libraries' ≡ 'Torrent
// Shadow Libraries').
func beltFolds(names []string) map[string]bool {
	folds := map[string]bool{}
	for _, n := range names {
		n = strings.TrimSpace(n)
		if n == "" {
			continue
		}
		folds[strings.ToLower(n)] = true
		folds[strings.ToLower(strings.TrimSpace(PrettifyFolderName(storage.SafeFilename(n))))] = true
	}
	return folds
}

// dictReader reads a dict-shaped value out of a decoded JSON tree
// (the house shapes: *storage.OrderedMap or map[string]any); nil when
// the value is not dict-shaped.
type dictReader func(key string) any

func dictOf(v any) (dictReader, bool) {
	switch t := v.(type) {
	case *storage.OrderedMap:
		return func(key string) any {
			val, _ := t.Get(key)
			return val
		}, true
	case map[string]any:
		return func(key string) any {
			return t[key]
		}, true
	}
	return nil, false
}

// strList mirrors the belt's `for x in (value or [])` + str(x) reads:
// a non-list answers nothing, list items render through Python str().
func strList(v any) []string {
	items, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(items))
	for _, x := range items {
		out = append(out, helpers.PyStr(x))
	}
	return out
}

// dictList reads a list of dict-shaped items (the categories /
// subcategories ladders); a non-dict item answers the empty reader
// (Python: `if not isinstance(cat, dict): continue` — the sub-ladder
// skips it, so the empty reader is the same skip).
func dictList(v any) []dictReader {
	items, ok := v.([]any)
	if !ok {
		return nil
	}
	var out []dictReader
	empty := func(string) any { return nil }
	for _, item := range items {
		if get, ok := dictOf(item); ok {
			out = append(out, get)
		} else {
			out = append(out, empty)
		}
	}
	return out
}

// LoadDirectoryConfig mirrors load_directory_config: the vault's
// _system/directory-config.json (owner-editable), falling back to the
// app's shipped default — embedded here (assets/directorysync, the
// doc 04 §1 home; the reference reads app/taxonomy/ from disk — a
// documented convergence) — and then to the floor dict. Never errors;
// the answer is always dict-shaped.
func LoadDirectoryConfig(vault string) any {
	if data, err := os.ReadFile(filepath.Join(SystemDir(vault), DirectoryConfigFilename)); err == nil {
		if v, err := storage.DecodeJSONValue(data); err == nil {
			if _, ok := dictOf(v); ok {
				return v
			}
		}
	}
	if v, err := storage.DecodeJSONValue(assets.WebsiteDirectoryConfig); err == nil {
		if _, ok := dictOf(v); ok {
			return v
		}
	}
	floor := storage.NewOrderedMap()
	floor.Set("categories", []any{})
	floor.Set("tag_aliases", storage.NewOrderedMap())
	shared := make([]any, len(DefaultSharedHosts))
	for i, s := range DefaultSharedHosts {
		shared[i] = s
	}
	floor.Set("shared_host_exceptions", shared)
	return floor
}

// LocalOnlyNotePatterns mirrors local_only_note_patterns: PURE — the
// local-only NOTE globs (the stems + the owner's own extras —
// deduplicated, the shipped stems first). 'CakhiaTV*.md' matches the
// note wherever it lands — including the repos vault's _review twin
// of the same stem.
func LocalOnlyNotePatterns(extra []string) []string {
	var patterns []string
	seen := map[string]bool{}
	for _, stem := range DefaultLocalOnlyStems {
		pat := stem + "*.md"
		if !seen[pat] {
			seen[pat] = true
			patterns = append(patterns, pat)
		}
	}
	for _, pat := range extra {
		pat = strings.TrimSpace(pat)
		if pat != "" && !seen[pat] {
			seen[pat] = true
			patterns = append(patterns, pat)
		}
	}
	return patterns
}

// PrivateExclusionPatterns mirrors private_exclusion_patterns: the
// FULL local-only belt for the Websites vault's seal — the
// private-category FOLDER forms + the local-only NOTE globs + the
// vault's own “local_only“ config list (the owner's extension
// point — _system/directory-config.json). The folder forms are BOTH
// derived from the belt names (the safe_filename spelling — a folder
// renamed later still matches by name) and discovered on disk (the
// vault's actual category/subcategory folders, matched folded against
// the belt). Never raises; the note globs are the floor on every
// error. log may be nil (the silent belt).
func PrivateExclusionPatterns(vault string, log func(string)) []string {
	if log == nil {
		log = func(string) {}
	}
	notePatterns := LocalOnlyNotePatterns(nil)

	get, hasConfig := dictOf(LoadDirectoryConfig(vault))
	if !hasConfig {
		get = func(string) any { return nil }
	}

	// extras = [str(x) for x in (config.get('local_only') or [])
	//            if str(x or '').strip()]
	var extras []string
	for _, x := range strList(get("local_only")) {
		if strings.TrimSpace(x) != "" {
			extras = append(extras, x)
		}
	}
	notePatterns = LocalOnlyNotePatterns(extras)

	// names: the shipped privates + every private-visibility node
	namesSet := map[string]bool{}
	for _, n := range DefaultPrivateNames {
		namesSet[n] = true
	}
	for _, cat := range dictList(get("categories")) {
		if vis, _ := cat("visibility").(string); vis == "private" {
			namesSet[strings.TrimSpace(helpers.PyStr(cat("name")))] = true
		}
		for _, sub := range dictList(cat("subcategories")) {
			if vis, _ := sub("visibility").(string); vis == "private" {
				namesSet[strings.TrimSpace(helpers.PyStr(sub("name")))] = true
			}
		}
	}
	names := make([]string, 0, len(namesSet))
	for n := range namesSet {
		names = append(names, n)
	}
	sort.Strings(names)

	folds := beltFolds(names)
	var folderPatterns []string
	// the belt names' own safe_filename forms (always path-safe)
	for _, n := range names {
		n = strings.TrimSpace(n)
		if n == "" {
			continue
		}
		folderPatterns = append(folderPatterns, storage.SafeFilename(n)+"/")
	}
	// the ACTUAL folders on disk (top two levels), matched folded
	vault = strings.TrimSpace(vault)
	if vault != "" {
		scanTwoLevels(vault, folds, &folderPatterns)
	}
	seen := map[string]bool{}
	out := []string{}
	for _, p := range append(folderPatterns, notePatterns...) {
		if p != "" && !seen[p] {
			seen[p] = true
			out = append(out, p)
		}
	}
	head := out
	if len(head) > 6 {
		head = head[:6]
	}
	ellipsis := ""
	if len(out) > 6 {
		ellipsis = "…"
	}
	log("[directory] local-only belt: " + strconv.Itoa(len(out)) + " pattern(s) for the seal (" +
		strings.Join(head, ", ") + ellipsis + ")")
	return out
}

// scanTwoLevels mirrors the disk half: category dirs (dot-dirs
// skipped) matching the folds fold in escaped; otherwise their
// subcategory dirs matching the folds fold in escaped pairs.
func scanTwoLevels(vault string, folds map[string]bool, out *[]string) {
	fi, err := os.Stat(vault)
	if err != nil || !fi.IsDir() {
		return
	}
	entries, err := os.ReadDir(vault)
	if err != nil {
		return
	}
	names := make([]string, len(entries))
	for i, e := range entries {
		names[i] = e.Name()
	}
	sort.Strings(names)
	for _, entry := range names {
		if strings.HasPrefix(entry, ".") {
			continue
		}
		catDir := filepath.Join(vault, entry)
		if fi, err := os.Stat(catDir); err != nil || !fi.IsDir() {
			continue
		}
		if folds[strings.TrimSpace(strings.ToLower(PrettifyFolderName(entry)))] ||
			folds[strings.TrimSpace(strings.ToLower(entry))] {
			*out = append(*out, gitignoreEscape(entry)+"/")
			continue
		}
		subs, err := os.ReadDir(catDir)
		if err != nil {
			continue // the per-folder OSError door
		}
		subNames := make([]string, len(subs))
		for i, e := range subs {
			subNames[i] = e.Name()
		}
		sort.Strings(subNames)
		for _, sub := range subNames {
			subDir := filepath.Join(catDir, sub)
			if fi, err := os.Stat(subDir); err != nil || !fi.IsDir() {
				continue
			}
			if folds[strings.TrimSpace(strings.ToLower(PrettifyFolderName(sub)))] ||
				folds[strings.TrimSpace(strings.ToLower(sub))] {
				*out = append(*out, gitignoreEscape(entry)+"/"+gitignoreEscape(sub)+"/")
			}
		}
	}
}
