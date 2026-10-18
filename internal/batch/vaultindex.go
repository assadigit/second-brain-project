// Package batch, file vaultindex.go — the port of
// gui/vault_index.py (the in-memory URL→note index; the reconciliation
// ground truth). Task T-25 (Phase 2); spec: spec-gui-cache-db.md
// §vault_index.py.
//
// gui/link_helpers.py's normalize_url is links.NormalizeURL (T-11);
// links.SiteKeyOf is the site map's key.
package batch

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/assadigit/second-brain-project/internal/helpers"
	"github.com/assadigit/second-brain-project/internal/links"
	"github.com/assadigit/second-brain-project/internal/storage"
)

// pyS is Python's str-pattern \s class (Unicode whitespace) — the
// ported frontmatter patterns carry Python \s semantics.
const pyS = `[\s\v\x1c-\x1f\x85\p{Z}]`

var (
	sourceRE    = regexp.MustCompile(`(?m)^source:` + pyS + `*(.+)$`)
	siteLinksRE = regexp.MustCompile(`(?m)^site_links:` + pyS + `*(.+)$`)
	managedByRE = regexp.MustCompile(`(?m)^managed_by:` + pyS + `*(.+)$`)
	fetchStatRE = regexp.MustCompile(`(?m)^fetch_status:` + pyS + `*(.+)$`)
)

// skippedFolderMarks mirrors the walk's skip list: SUBSTRING match on
// the walked path (the reference's `any(skip in root ...)` — a folder
// whose name merely CONTAINS a mark skips too). _review is NOT here:
// those are real notes.
var skippedFolderMarks = []string{
	"_moc", "_inbox", "attachments", ".obsidian", ".trash",
}

// VaultIndex is the in-memory index of all notes in the vault, keyed by
// normalized source URL. Rebuilt at the start of each processing run;
// updated incrementally as new notes are written. This is the GROUND
// TRUTH for dedup — if a URL is in the vault, it's been processed
// (regardless of what the SQLite cache says).
type VaultIndex struct {
	vaultPath string
	normalize func(string) string

	urlToPath  map[string]string // normalized_url -> note_path
	siteToPath map[string]string // site key (host minus www) -> note_path
	built      bool
}

// NewVaultIndex mirrors VaultIndex.__init__: normalizer nil ⇒
// links.NormalizeURL (the GitHub vault's behavior); the Websites vault
// passes links.NormalizeWebsiteURL (query params are part of a
// website's identity, unlike a GitHub repo's).
func NewVaultIndex(vaultPath string,
	normalizer func(string) string) *VaultIndex {
	if normalizer == nil {
		normalizer = links.NormalizeURL
	}
	return &VaultIndex{
		vaultPath:  vaultPath,
		normalize:  normalizer,
		urlToPath:  map[string]string{},
		siteToPath: map[string]string{},
	}
}

// Rebuild mirrors rebuild (vault_index.py:74-133): scan the vault and
// build the index, reading the head of each .md file (4000 chars — the
// frontmatter's site_links line and the ownership/status keys the site
// map reads live past the old 800-char window).
//
// The read is Python's open(fpath, 'r', encoding='utf-8') — TEXT mode
// (universal newlines) and STRICT (an undecodable note is skipped by
// the reference's except-pass, never replace-decoded). The walk is
// lexical (filepath.WalkDir) where Python's os.walk yields filesystem
// order — the documented divergence class (first-come-wins site-map
// order can differ; see doc.go).
func (v *VaultIndex) Rebuild(logSignal func(msg, level string)) {
	v.urlToPath = map[string]string{}
	v.siteToPath = map[string]string{}
	if v.vaultPath == "" {
		v.built = true
		return
	}
	if fi, err := os.Stat(v.vaultPath); err != nil || !fi.IsDir() {
		v.built = true
		return
	}

	count := 0
	_ = filepath.WalkDir(v.vaultPath, func(path string,
		d os.DirEntry, err error) error {
		if err != nil {
			return nil // the reference's walk skips unreadable entries
		}
		if d.IsDir() {
			// Skip non-note folders — but INCLUDE _review (real notes!)
			// and skip .trash (banned notes must stop counting).
			for _, skip := range skippedFolderMarks {
				if strings.Contains(path, skip) {
					return filepath.SkipDir
				}
			}
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}
		// The reference's per-file try/except-pass: an unreadable or
		// undecodable note is skipped, never fatal.
		func() {
			data, err := os.ReadFile(path)
			if err != nil {
				return
			}
			if !utf8.Valid(data) {
				return // strict decode would raise
			}
			content := helpers.PyNewlines(string(data))
			if len([]rune(content)) > 4000 {
				content = string([]rune(content)[:4000])
			}
			if m := sourceRE.FindStringSubmatch(content); m != nil {
				sourceURL := strings.TrimSpace(m[1])
				sourceURL = strings.Trim(sourceURL, `"'`)
				normalized := v.normalize(sourceURL)
				if normalized != "" {
					v.urlToPath[normalized] = path
					count++
					v.registerSiteNote(path, content, sourceURL,
						normalized)
				}
			}
			// ONE NOTE PER SITE: the consolidated links ride the site
			// note's frontmatter list; each indexes to the note that
			// now carries it.
			for _, m := range siteLinksRE.FindAllStringSubmatch(content,
				-1) {
				inner := strings.TrimSpace(m[1])
				if strings.HasPrefix(inner, "[") &&
					strings.HasSuffix(inner, "]") {
					inner = inner[1 : len(inner)-1]
				}
				for _, item := range strings.Split(inner, ",") {
					u := strings.Trim(strings.TrimSpace(item), `"`)
					u = strings.Trim(u, `'`)
					if u == "" {
						continue
					}
					nu := v.normalize(u)
					if nu == "" {
						continue
					}
					if _, ok := v.urlToPath[nu]; !ok {
						v.urlToPath[nu] = path
					}
				}
			}
		}()
		return nil
	})

	v.built = true
	if logSignal != nil {
		logSignal("📚 Vault index: "+strconv.Itoa(count)+
			" notes indexed", "info")
	}
}

// registerSiteNote mirrors _register_site_note (vault_index.py:135-158):
// register the note as the site's when it is a REAL one (the app's own
// — the managed_by line says so; outside _review; not a failed fetch).
// First-come-wins in the walk's order; the map is a probe, never a
// gate — an empty map simply leaves the law off.
func (v *VaultIndex) registerSiteNote(fpath, content, sourceURL,
	normalized string) {
	if strings.Contains(strings.ReplaceAll(fpath, "\\", "/"), "/_review/") {
		return
	}
	m := managedByRE.FindStringSubmatch(content)
	if m == nil || !strings.Contains(strings.ToLower(m[1]), "gitcurator") {
		return // a hand-written note is the owner's, never the anchor
	}
	if s := fetchStatRE.FindStringSubmatch(content); s != nil &&
		strings.ToLower(strings.Trim(strings.TrimSpace(s[1]), `"'`)) ==
			"failed" {
		return // a failed placeholder is not the site's note
	}
	key := links.SiteKeyOf(normalized)
	if key == "" {
		key = links.SiteKeyOf(sourceURL)
	}
	if key != "" {
		if _, ok := v.siteToPath[key]; !ok {
			v.siteToPath[key] = fpath
		}
	}
}

// HasURL mirrors has_url: is a URL (normalized) already in the vault?
func (v *VaultIndex) HasURL(url string) bool {
	_, ok := v.urlToPath[v.normalize(url)]
	return ok
}

// GetPath mirrors get_path: the note path for a URL, or "" if absent.
func (v *VaultIndex) GetPath(url string) string {
	return v.urlToPath[v.normalize(url)]
}

// SiteNoteFor mirrors site_note_for (v0.64.0 — ONE NOTE PER SITE's
// probe): the note that owns this link's SITE (its host), or "" when
// the vault holds no real note for the site.
func (v *VaultIndex) SiteNoteFor(url string) string {
	key := links.SiteKeyOf(v.normalize(url))
	if key == "" {
		key = links.SiteKeyOf(url)
	}
	return v.siteToPath[key]
}

// AddURL mirrors add_url: add a new URL→path mapping (called after
// writing a new note). The site map learns the new note too, by reading
// the note's own frontmatter — with errors='replace' (lenient, unlike
// rebuild's strict read).
func (v *VaultIndex) AddURL(url, path string) {
	v.urlToPath[v.normalize(url)] = path
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	content := helpers.PyNewlines(helpers.DecodeUTF8Replace(data))
	if len([]rune(content)) > 4000 {
		content = string([]rune(content)[:4000])
	}
	norm := v.normalize(url)
	v.registerSiteNote(path, content, norm, norm)
}

// Count mirrors the count property.
func (v *VaultIndex) Count() int {
	return len(v.urlToPath)
}

// MirrorIndex exports the index in the vault-mirror entries' shape:
// {url_normalized: {path: note_path}} as an OrderedMap. SORTED-URL
// order (Go maps are unordered — the sorted order keeps the pushed
// wire bytes deterministic; the reference iterates its dict in
// insertion order, a documented divergence recorded at
// workerclient.BuildVaultMirrorEntries). The category/title/
// first_processed/last_updated fields the entries builder reads stay
// absent — the builder's .get() defaults fill them, exactly like a
// caller that only knows the paths (T-62's Phase-5 mirror push).
func (v *VaultIndex) MirrorIndex() *storage.OrderedMap {
	out := storage.NewOrderedMap()
	if v == nil {
		return out
	}
	urls := make([]string, 0, len(v.urlToPath))
	for u := range v.urlToPath {
		urls = append(urls, u)
	}
	sort.Strings(urls)
	for _, u := range urls {
		info := storage.NewOrderedMap()
		info.Set("path", v.urlToPath[u])
		out.Set(u, info)
	}
	return out
}

// ---------------------------------------------------------------------------
// find_obsidian_vaults + _safe_moc_name
// ---------------------------------------------------------------------------

// FindObsidianVaults mirrors find_obsidian_vaults (vault_index.py:199-272):
// discover Obsidian vaults from (1) Obsidian's own vault registry
// (obsidian.json — %APPDATA% on Windows, ~/Library/Application Support
// on macOS, ~/.config on Linux) and (2) common filesystem paths (one
// level deep). Order preserved, de-duplicated by normpath.
func FindObsidianVaults() []string {
	var vaults []string

	// --- 1. Obsidian's own registry ---
	var candidates []string
	if appdata := os.Getenv("APPDATA"); appdata != "" {
		candidates = append(candidates,
			filepath.Join(appdata, "obsidian", "obsidian.json"))
	}
	if home := homeDir(); home != "" {
		candidates = append(candidates,
			filepath.Join(home, "Library", "Application Support",
				"obsidian", "obsidian.json"),
			filepath.Join(home, ".config", "obsidian", "obsidian.json"))
	}
	for _, cfgPath := range candidates {
		data, err := os.ReadFile(cfgPath)
		if err != nil {
			continue
		}
		var doc struct {
			Vaults map[string]struct {
				Path string `json:"path"`
			} `json:"vaults"`
		}
		if err := json.Unmarshal(data, &doc); err != nil {
			continue
		}
		// map iteration order is random in Go; the registry's values
		// are sorted by path for the deterministic Python-dict order
		// the reference's insertion produced.
		var paths []string
		for _, vault := range doc.Vaults {
			if vault.Path == "" {
				continue
			}
			paths = append(paths, vault.Path)
		}
		sortStrings(paths)
		for _, p := range paths {
			if fi, err := os.Stat(p); err == nil && fi.IsDir() {
				vaults = append(vaults, p)
			}
		}
	}

	// --- 2. Common path fallback (one level deep) ---
	home := homeDir()
	roots := []string{
		filepath.Join(home, "Documents"),
		filepath.Join(home, "OneDrive", "Documents"),
		filepath.Join(home, "Desktop"),
		filepath.Join(home, "Obsidian"),
		filepath.Join(home, "Vaults"),
		filepath.Join(home, "Documents", "Obsidian Vaults"),
		filepath.Join(home, "Documents", "Obsidian"),
		filepath.Join(home, "OneDrive", "Documents", "Obsidian Vaults"),
		filepath.Join(home, "OneDrive", "Documents", "Obsidian"),
		filepath.Join(home, "Desktop", "Obsidian Vaults"),
		filepath.Join(home, "Desktop", "Obsidian"),
	}
	for _, root := range roots {
		if fi, err := os.Stat(root); err != nil || !fi.IsDir() {
			continue
		}
		// root itself may be a vault
		if isDir(filepath.Join(root, ".obsidian")) {
			vaults = append(vaults, root)
		}
		// one level deep (os.listdir order — lexical in the port)
		entries, err := os.ReadDir(root)
		if err != nil {
			continue
		}
		for _, e := range entries {
			sub := filepath.Join(root, e.Name())
			if isDir(sub) && isDir(filepath.Join(sub, ".obsidian")) {
				vaults = append(vaults, sub)
			}
		}
	}

	// de-duplicate, preserve order
	seen := map[string]bool{}
	var unique []string
	for _, v := range vaults {
		norm := filepath.Clean(v)
		if !seen[norm] {
			seen[norm] = true
			unique = append(unique, v)
		}
	}
	return unique
}

// SafeMocName mirrors _safe_moc_name (vault_index.py:274-294): sanitize
// a category name for use as a MOC filename / wiki-link — '/' and '\'
// become '_', every Windows-illegal char is removed (<>:"|?* + control
// chars), whitespace collapses to single spaces, no trailing dots or
// spaces, empty falls back to "Uncategorized".
func SafeMocName(cat string) string {
	name := strings.ReplaceAll(strings.ReplaceAll(cat, "/", "_"),
		"\\", "_")
	name = illegalCharsRE.ReplaceAllString(name, "")
	name = whitespaceRE.ReplaceAllString(name, " ")
	name = strings.Trim(name, " .")
	if name == "" {
		return "Uncategorized"
	}
	return name
}

var (
	illegalCharsRE = regexp.MustCompile(`[<>:"|?*\x00-\x1f]`)
	whitespaceRE   = regexp.MustCompile(`\s+`)
)

func homeDir() string {
	h, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return h
}

func isDir(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.IsDir()
}

func sortStrings(s []string) {
	sort.Strings(s)
}
