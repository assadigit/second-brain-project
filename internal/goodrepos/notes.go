// notes.go — the tiny YAML-frontmatter subset (_parse_frontmatter /
// _parse_scalar), the note parsers (_to_int / _owner_repo_from_url /
// _as_tags / _as_aliases / _parse_note), and scan_vault (the os.walk
// category walk).
package goodrepos

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/assadigit/second-brain-project/internal/helpers"
)

// fmKeyRe / fmItemRe / urlRe mirror _FM_KEY_RE / _FM_ITEM_RE / _URL_RE.
var (
	fmKeyRe  = regexp.MustCompile(`^([A-Za-z0-9_\-]+):\s*(.*)$`)
	fmItemRe = regexp.MustCompile(`^\s*-\s+(.*)$`)
	urlRe    = regexp.MustCompile(`github\.com/([A-Za-z0-9_.\-]+)/([A-Za-z0-9_.\-]+)`)
)

// pyTruthy mirrors Python's truthiness for the shapes frontmatter
// values take (None / "" / [] / {} / 0 / False are falsy).
func pyTruthy(v any) bool {
	switch t := v.(type) {
	case nil:
		return false
	case string:
		return t != ""
	case bool:
		return t
	case int:
		return t != 0
	case int64:
		return t != 0
	case float64:
		return t != 0
	case []any:
		return len(t) > 0
	case map[string]any:
		return len(t) > 0
	default:
		return true
	}
}

// pyStrOrEmpty mirrors str(v or ""): the falsy doors collapse to "".
func pyStrOrEmpty(v any) string {
	if !pyTruthy(v) {
		return ""
	}
	return helpers.PyStr(v)
}

// parseScalar mirrors _parse_scalar: frontmatter value → scalar/list.
// Handles quoted strings and inline "[a, b, c]" lists (parts keep
// their inner spaces — the reference's quirk). Never fails.
func parseScalar(raw string) any {
	s := strings.TrimSpace(raw)
	if len(s) >= 2 && strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
		inner := strings.TrimSpace(s[1 : len(s)-1])
		if inner == "" {
			return []any{}
		}
		var out []any
		for _, part := range strings.Split(inner, ",") {
			out = append(out, parseScalar(part))
		}
		return out
	}
	if len(s) >= 2 && s[0] == s[len(s)-1] && (s[0] == '\'' || s[0] == '"') {
		return strings.TrimSpace(s[1 : len(s)-1])
	}
	return s
}

// parseFrontmatter mirrors _parse_frontmatter: the '---' block —
// scalars, inline lists, block lists (aliases). Unknown shapes are
// ignored, never fatal. Key order is the reference's insertion order.
func parseFrontmatter(lines []string) *fmMap {
	m := newFMMap()
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return m
	}
	key := ""
	for _, ln := range lines[1:] {
		if strings.TrimSpace(ln) == "---" {
			break
		}
		if mm := fmKeyRe.FindStringSubmatch(ln); mm != nil {
			key = mm[1]
			m.set(key, parseScalar(mm[2]))
			continue
		}
		if mm := fmItemRe.FindStringSubmatch(ln); mm != nil && key != "" {
			value := parseScalar(mm[1])
			current, present := m.get(key)
			switch cur := current.(type) {
			case []any:
				m.set(key, append(cur, value))
			default:
				if !present || cur == "" {
					m.set(key, []any{value})
				} else {
					m.set(key, []any{cur, value})
				}
			}
		}
	}
	return m
}

// fmMap is the frontmatter dict: insertion-ordered, missing-key aware
// (the reference's fm.get returns None for absent keys — the "in
// (None, ”)" door needs the distinction).
type fmMap struct {
	keys   []string
	values map[string]any
}

func newFMMap() *fmMap { return &fmMap{values: map[string]any{}} }

func (m *fmMap) set(k string, v any) {
	if _, ok := m.values[k]; !ok {
		m.keys = append(m.keys, k)
	}
	m.values[k] = v
}

func (m *fmMap) get(k string) (any, bool) {
	v, ok := m.values[k]
	return v, ok
}

func (m *fmMap) value(k string) any { v, _ := m.get(k); return v }

// toInt mirrors _to_int: best-effort int ("140232", "1,402"); nil when
// unparsable. bool → nil (the isinstance check); ints pass through.
func toInt(value any) *int {
	switch v := value.(type) {
	case bool:
		return nil
	case int:
		n := v
		return &n
	case int64:
		n := int(v)
		return &n
	case float64:
		if v == float64(int(v)) {
			n := int(v)
			return &n
		}
		return nil
	case string:
		s := strings.ReplaceAll(strings.TrimSpace(v), ",", "")
		if intsRe.MatchString(s) {
			if n, err := strconv.Atoi(s); err == nil {
				return &n
			}
		}
		return nil
	default:
		return nil
	}
}

var intsRe = regexp.MustCompile(`^-?\d+$`)

// ownerRepoFromURL mirrors _owner_repo_from_url:
// "https://github.com/owner/repo/issues/42" → "owner/repo".
func ownerRepoFromURL(url any) string {
	s, ok := url.(string)
	if !ok || s == "" {
		return ""
	}
	mm := urlRe.FindStringSubmatch(s)
	if mm == nil {
		return ""
	}
	owner, repo := mm[1], mm[2]
	if strings.HasSuffix(repo, ".git") && len(repo) > 4 {
		repo = repo[:len(repo)-4]
	}
	return owner + "/" + repo
}

// asTags mirrors _as_tags: tolerates the malformed scalar shape the
// LLM occasionally produces ("cp, python, server, tools]").
func asTags(value any) []string {
	var raw []string
	switch v := value.(type) {
	case []any:
		for _, t := range v {
			raw = append(raw, helpers.PyStr(t))
		}
	case string:
		s := strings.TrimSpace(helpers.StripChars(strings.TrimSpace(v), "[]"))
		s = strings.TrimSpace(s)
		if s != "" {
			for _, p := range strings.Split(s, ",") {
				raw = append(raw, strings.TrimSpace(p))
			}
		}
	}
	var out []string
	for _, r := range raw {
		t := strings.TrimSpace(r)
		t = strings.Trim(t, `"`)
		t = strings.Trim(t, `'`)
		t = strings.TrimSpace(t)
		if t != "" {
			out = append(out, t)
		}
	}
	return out
}

// asAliases mirrors _as_aliases.
func asAliases(value any) []string {
	switch v := value.(type) {
	case []any:
		var out []string
		for _, a := range v {
			if s := strings.TrimSpace(helpers.PyStr(a)); s != "" {
				out = append(out, s)
			}
		}
		return out
	case string:
		if s := strings.TrimSpace(v); s != "" {
			return []string{s}
		}
		return nil
	default:
		return nil
	}
}

// Note mirrors the _Note dataclass: one curated note, parsed just
// enough to feed the directory README. Stars is *int (nil = None).
type Note struct {
	Path          string // absolute path inside the vault
	RelPath       string // vault-relative, POSIX ("AI-Domain/…/x.md")
	Display       string // "owner/repo"
	URL           string // source url ("" when unknown)
	Title         string // "# <title>" from the body
	Tldr          string // "> **TL;DR:** …" (fallback: What is it?)
	Tags          []string
	Stars         *int
	Language      string
	Category      string // frontmatter category ("" when missing)
	DateProcessed string
	Credibility   string
}

// ParseNote mirrors _parse_note: read (text-mode — universal
// newlines translated BEFORE the parse, the T-41 law), frontmatter,
// body mining ("# title", "> **TL;DR:** …", the "## What is it?"
// fallback), then the display-name resolution (url → aliases →
// filename stem). Never errors — an unreadable note parses as empty.
func ParseNote(path, relPath string) *Note {
	data, err := os.ReadFile(path)
	raw := ""
	if err == nil {
		raw = helpers.PyNewlines(helpers.DecodeUTF8Replace(data))
	}
	lines := helpers.SplitLines(raw)

	fm := parseFrontmatter(lines)
	bodyStart := len(lines)
	if len(lines) > 0 && strings.TrimSpace(lines[0]) == "---" {
		for i := 1; i < len(lines); i++ {
			if strings.TrimSpace(lines[i]) == "---" {
				bodyStart = i + 1
				break
			}
		}
	}

	// body mining: "# title", "> **TL;DR:** …", "## What is it?" fallback
	title, tldr, fallback := "", "", ""
	mode := ""
	for _, ln := range lines[minInt(bodyStart, len(lines)):] {
		s := strings.TrimSpace(ln)
		if title == "" && strings.HasPrefix(s, "# ") && !strings.HasPrefix(s, "##") {
			title = strings.TrimSpace(s[2:])
		}
		if tldr == "" && strings.HasPrefix(s, ">") && strings.Contains(s, "**TL;DR:**") {
			tldr = strings.TrimSpace(strings.Trim(
				strings.TrimSpace(strings.SplitN(s, "**TL;DR:**", 2)[1]), `"`))
		}
		if strings.HasPrefix(s, "## ") {
			mode = helpers.PyLower(strings.TrimSpace(s[3:]))
			continue
		}
		if mode == "what is it?" && s != "" && fallback == "" {
			fallback = s
		}
	}

	source := fm.value("source")
	if !pyTruthy(source) {
		source = ""
	}
	aliases := asAliases(fm.value("aliases"))
	display := ownerRepoFromURL(source)
	url := ""
	if display != "" {
		url = "https://github.com/" + display
	} else {
		if s, ok := source.(string); ok && s != "" {
			url = s
		}
		display = ""
		for _, a := range aliases {
			if strings.Contains(a, "/") {
				display = a
				break
			}
		}
		if display == "" && len(aliases) > 0 {
			display = aliases[0]
		}
		if display == "" {
			stem := trimExt(filepath.Base(path))
			if stem == "" {
				stem = "unknown"
			}
			if i := strings.Index(stem, "_"); i >= 0 {
				stem = stem[:i]
			}
			display = stem
		}
	}

	tldrOut := tldr
	if tldrOut == "" {
		tldrOut = fallback
	}
	return &Note{
		Path:          path,
		RelPath:       relPath,
		Display:       display,
		URL:           url,
		Title:         title,
		Tldr:          tldrOut,
		Tags:          asTags(fm.value("tags")),
		Stars:         toInt(fm.value("stars")),
		Language:      strings.TrimSpace(pyStrOrEmpty(fm.value("primary_language"))),
		Category:      strings.TrimSpace(pyStrOrEmpty(fm.value("category"))),
		DateProcessed: strings.TrimSpace(pyStrOrEmpty(fm.value("date_processed"))),
		Credibility:   strings.TrimSpace(pyStrOrEmpty(fm.value("credibility_score"))),
	}
}

// trimExt mirrors Path.stem (the final extension removed once).
func trimExt(name string) string {
	ext := filepath.Ext(name)
	if ext != "" {
		return name[:len(name)-len(ext)]
	}
	return name
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// CategoryNotes is one scanned category folder: rel is the
// vault-relative POSIX dir, Notes the parsed .md files in walk order.
type CategoryNotes struct {
	Rel   string
	Notes []*Note
}

// ScanVault mirrors scan_vault: walk the vault's category folders
// (os.walk topdown — sorted dirnames, IGNORED_DIRS pruned, the vault
// root not a category) and parse every curated note. The slice keeps
// the reference dict's insertion order (the walk's encounter order).
// Never errors.
func (s *GoodRepos) ScanVault() []*CategoryNotes {
	if s.vault == "" || !isDir(s.vault) {
		return nil
	}
	var found []*CategoryNotes
	walkVault(s.vault, func(dirpath string, dirnames, filenames []string) {
		relDir := toPyRelPath(dirpath, s.vault)
		if relDir == "." { // vault root: not a category folder
			return
		}
		var notes []*Note
		for _, fn := range filenames { // already sorted by the walk
			if !strings.HasSuffix(helpers.PyLower(fn), ".md") {
				continue
			}
			relPath := relDir + "/" + fn
			if hasDotDot(relPath) { // paranoid; the walk is safe
				continue
			}
			notes = append(notes, ParseNote(filepath.Join(dirpath, fn), relPath))
		}
		if len(notes) > 0 {
			found = append(found, &CategoryNotes{Rel: relDir, Notes: notes})
		}
	})
	return found
}

// walkVault is os.walk(topdown=True) — the T-45 law: a directory's
// own entry (its files) is yielded BEFORE its subdirs are walked,
// dirnames sorted+filtered IN PLACE (pruning them from the descent),
// filenames in ReadDir order. Never follows the walk into pruned
// dirs.
func walkVault(root string, visit func(dirpath string, dirnames, filenames []string)) {
	dirents, err := os.ReadDir(root)
	if err != nil {
		return // os.walk onerror: skipped silently (the reference default)
	}
	var dirnames []string
	var filenames []string
	for _, e := range dirents {
		if e.IsDir() {
			if !IgnoredDirs[e.Name()] {
				dirnames = append(dirnames, e.Name())
			}
		} else {
			filenames = append(filenames, e.Name())
		}
	}
	sortStrings(dirnames)
	sortStrings(filenames)
	visit(root, dirnames, filenames)
	for _, d := range dirnames {
		walkVault(filepath.Join(root, d), visit)
	}
}

// hasDotDot mirrors ".." in rel_path.split("/").
func hasDotDot(rel string) bool {
	for _, seg := range strings.Split(rel, "/") {
		if seg == ".." {
			return true
		}
	}
	return false
}

// toPyRelPath mirrors os.path.relpath(dirpath, vault) with os.sep →
// "/".
func toPyRelPath(dirpath, vault string) string {
	rel, err := filepath.Rel(vault, dirpath)
	if err != nil {
		return dirpath
	}
	return strings.ReplaceAll(rel, string(filepath.Separator), "/")
}

func isDir(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.IsDir()
}

func isFile(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && !fi.IsDir()
}

func sortStrings(s []string) { sort.Strings(s) }
