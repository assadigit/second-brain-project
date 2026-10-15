// Package taxonomy ports app/gitcurator/core/taxonomy.py — parses the
// owner-editable Websites category file into the Taxonomy model and
// validates LLM answers against parsed names before any path is touched
// (untrusted-input safe: an unknown name never builds a path).
//
// Spec: docs/port-blueprint/module-specs/spec-core-taxonomy.md;
// task T-14 (Phase 1). Parity is pinned by the generated golden in
// testdata/ (see taxonomy_test.go): the reference's own fixture (copied
// into testdata), one synthetic mini-taxonomy per parser rule, every
// error message byte-exact, the BOM-sig read, Python str.splitlines()/
// strip() semantics, and the SHIPPED file's parse snapshot (embedded in
// assets/taxonomy, sha256-pinned).
//
// Port notes (documented structural decisions):
//   - The parser iterates str.splitlines() boundaries, not just '\n' —
//     pySplitLines reproduces Python's full boundary set (\r, \v, \f,
//     \x1c-\x1e, \x85, U+2028/2029); strip() uses Python's Unicode
//     whitespace set (pyStrip — U+00A0, \x1c, U+2028 … strip like Python,
//     unlike strings.TrimSpace).
//   - Regexes port with Go's RE2 `\s` (ASCII) semantics, the T-11
//     precedent: the taxonomy grammar's realistic corpus is ASCII
//     whitespace (the golden pins the actual behavior).
//   - folder_relpath rides os.path.join semantics: the platform separator
//     (os.PathSeparator) with NO path cleaning — matching os.path.join's
//     behavior per-platform, exactly like the reference.
//   - load_taxonomy reads utf-8-sig (BOM stripped — the opposite quirk of
//     prompts' plain utf-8 read, which is why the BOM'd file parses
//     cleanly here). Invalid UTF-8 raises the decode error OUTSIDE the
//     TaxonomyError contract, mirroring Python's UnicodeDecodeError
//     escaping `except OSError`.
package taxonomy

import (
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/assadigit/second-brain-project/internal/constants"
	"github.com/assadigit/second-brain-project/internal/prompts"
	"github.com/assadigit/second-brain-project/internal/storage"
)

// TaxonomyError mirrors taxonomy.py TaxonomyError: the taxonomy file
// could not be parsed unambiguously.
type TaxonomyError struct {
	Msg string
}

// Error implements error.
func (e *TaxonomyError) Error() string { return e.Msg }

// ---------------------------------------------------------------------------
// Line-level helpers (taxonomy.py:52-93).
// ---------------------------------------------------------------------------

// emojiRE mirrors _EMOJI_RE: strips emoji/symbol runs from headings —
// U+1F000-1FAFF, U+2600-27BF, U+2190-21FF, U+2B00-2BFF, U+FE0F, U+200D.
var emojiRE = regexp.MustCompile(
	`[\x{1F000}-\x{1FAFF}\x{2600}-\x{27BF}\x{2190}-\x{21FF}\x{2B00}-\x{2BFF}\x{FE0F}\x{200D}]`)

// trailingItalicRE mirrors _TRAILING_ITALIC_RE: a trailing "*(...)*" note.
var trailingItalicRE = regexp.MustCompile(`\s*\*\(.*\)\*\s*$`)

// bulletRE mirrors _BULLET_RE: a top-level "- " bullet.
var bulletRE = regexp.MustCompile(`^\s*-\s+(.*)$`)

// exampleTagsRE mirrors _EXAMPLE_TAGS_RE (case-insensitive).
var exampleTagsRE = regexp.MustCompile(`(?i)^\s*Example tags:\s*(.*)$`)

// noSubcategoriesRE mirrors _NO_SUBCATEGORIES_RE (case-insensitive,
// anywhere in the line).
var noSubcategoriesRE = regexp.MustCompile(`(?i)\(\s*no subcategories\s*\)`)

// dashSplitRE mirrors _DASH_SPLIT_RE: em/en dashes (with surrounding
// whitespace) separating a name from its description.
var dashSplitRE = regexp.MustCompile(`\s*[—–]\s*`)

// heading3RE / heading2RE mirror the `^###\s+` / `^##\s+` line matches.
var heading3RE = regexp.MustCompile(`^###\s+(.*)$`)
var heading2RE = regexp.MustCompile(`^##\s+(.*)$`)

// tagsRE mirrors the `#tag` hint extractor.
var tagsRE = regexp.MustCompile(`#([A-Za-z0-9\-_]+)`)

// leadingDashRE mirrors `^[—–-]\s*` (prose cleanup).
var leadingDashRE = regexp.MustCompile(`^[—–\-]\s*`)

// pureNoSubcategoriesRE mirrors re.fullmatch(r'no subcategories', …, I).
var pureNoSubcategoriesRE = regexp.MustCompile(`(?i)^no subcategories$`)

// stripCategoryHeading mirrors _strip_category_heading: trailing italic
// note first (it sits after the emoji), then the emoji, then strip.
func stripCategoryHeading(text string) string {
	text = trailingItalicRE.ReplaceAllString(text, "")
	text = emojiRE.ReplaceAllString(text, "")
	return pyStrip(text)
}

// stripSubcategory mirrors _strip_subcategory: bold markers stripped,
// everything from the em/en dash onward dropped. Python
// _DASH_SPLIT_RE.split(text, maxsplit=1)[0] — Go's Split(s, 2) is the
// maxsplit=1 equivalent.
func stripSubcategory(text string) string {
	text = pyStrip(text)
	text = strings.ReplaceAll(text, "**", "")
	parts := dashSplitRE.Split(text, 2)
	return pyStrip(parts[0])
}

// stripTagsLine mirrors _strip_tags_line: '#icons' → "icons" per tag.
func stripTagsLine(text string) []string {
	var out []string
	for _, m := range tagsRE.FindAllStringSubmatch(text, -1) {
		out = append(out, pyStrip(strings.TrimPrefix(m[1], "#")))
	}
	return out
}

// ---------------------------------------------------------------------------
// The model (taxonomy.py:100-206).
// ---------------------------------------------------------------------------

// Category is one parsed category: name, one-line definition,
// subcategories (ordered, deduped), subcategory definitions, tag hints.
type Category struct {
	Name                   string
	Definition             string
	Subcategories          []string
	SubcategoryDefinitions map[string]string
	TagHints               []string
}

// OneLine mirrors Category.one_line: "Name — definition" (em dash) when a
// definition exists, else just the name — the pass-1 classifier payload.
func (c *Category) OneLine() string {
	if c.Definition != "" {
		return c.Name + " — " + c.Definition
	}
	return c.Name
}

// Taxonomy is the parsed model: ordered categories + verbatim judgment
// rules. Duplicate category names are a construction error.
type Taxonomy struct {
	categories    []Category
	judgmentRules string
	sourcePath    string
	byName        map[string]int
}

// newTaxonomy mirrors Taxonomy.__init__ (incl. the duplicate check).
func newTaxonomy(categories []Category, judgmentRules, sourcePath string) (*Taxonomy, error) {
	byName := make(map[string]int, len(categories))
	for i, c := range categories {
		if _, dup := byName[c.Name]; dup {
			return nil, &TaxonomyError{Msg: "duplicate category name in taxonomy file"}
		}
		byName[c.Name] = i
	}
	return &Taxonomy{
		categories:    categories,
		judgmentRules: pyStrip(judgmentRules),
		sourcePath:    sourcePath,
		byName:        byName,
	}, nil
}

// Categories returns the ordered categories (the slice itself; treat as
// read-only — the Python attribute is the same shared list).
func (t *Taxonomy) Categories() []Category { return t.categories }

// JudgmentRules mirrors judgment_rules: the verbatim rules passed to the
// classifier.
func (t *Taxonomy) JudgmentRules() string { return t.judgmentRules }

// SourcePath mirrors source_path.
func (t *Taxonomy) SourcePath() string { return t.sourcePath }

// CategoryNames mirrors category_names: the ordered names.
func (t *Taxonomy) CategoryNames() []string {
	names := make([]string, len(t.categories))
	for i, c := range t.categories {
		names[i] = c.Name
	}
	return names
}

// Category mirrors category(name): exact-match lookup after strip — model
// answers must match a parsed name; nil when unknown.
func (t *Taxonomy) Category(name string) *Category {
	if i, ok := t.byName[pyStrip(name)]; ok {
		return &t.categories[i]
	}
	return nil
}

// IsCategory mirrors is_category.
func (t *Taxonomy) IsCategory(name string) bool {
	_, ok := t.byName[pyStrip(name)]
	return ok
}

// SubcategoriesOf mirrors subcategories_of: a copy of the ordered list
// (nil-safe for unknown categories).
func (t *Taxonomy) SubcategoriesOf(categoryName string) []string {
	if cat := t.Category(categoryName); cat != nil {
		return append([]string(nil), cat.Subcategories...)
	}
	return nil
}

// IsSubcategoryOf mirrors is_subcategory_of: a stripped, non-empty name
// that is a member of the category's subcategory list.
func (t *Taxonomy) IsSubcategoryOf(categoryName, subcategory string) bool {
	sub := pyStrip(subcategory)
	return sub != "" && containsString(t.SubcategoriesOf(categoryName), sub)
}

func containsString(list []string, s string) bool {
	for _, e := range list {
		if e == s {
			return true
		}
	}
	return false
}

// CategoryOneLiner mirrors category_one_liner: the pass-1 payload —
// "- Name" / "- Name — definition" lines joined by newlines.
func (t *Taxonomy) CategoryOneLiner() string {
	lines := make([]string, len(t.categories))
	for i := range t.categories {
		lines[i] = "- " + t.categories[i].OneLine()
	}
	return strings.Join(lines, "\n")
}

// SubcategoryOneLiner mirrors subcategory_one_liner: the chosen category's
// pass-2 payload — "" for unknown, "- none (…)" when it has no
// subcategories, else "- Sub" / "- Sub — definition" lines.
func (t *Taxonomy) SubcategoryOneLiner(categoryName string) string {
	cat := t.Category(categoryName)
	if cat == nil {
		return ""
	}
	if len(cat.Subcategories) == 0 {
		return "- none (this category has no subcategories)"
	}
	lines := make([]string, 0, len(cat.Subcategories))
	for _, s := range cat.Subcategories {
		if d := cat.SubcategoryDefinitions[s]; d != "" {
			lines = append(lines, "- "+s+" — "+d)
		} else {
			lines = append(lines, "- "+s)
		}
	}
	return strings.Join(lines, "\n")
}

// TagHintsOf mirrors tag_hints_of: "#tag" hints joined by spaces.
func (t *Taxonomy) TagHintsOf(categoryName string) string {
	cat := t.Category(categoryName)
	if cat == nil {
		return ""
	}
	hints := make([]string, len(cat.TagHints))
	for i, h := range cat.TagHints {
		hints[i] = "#" + h
	}
	return strings.Join(hints, " ")
}

// FolderRelpath mirrors folder_relpath: the vault-relative folder for a
// (validated) category/subcategory — safe_filename per component, joined
// with the platform separator exactly like os.path.join. An empty
// subcategory means the category folder itself. Unknown names are
// REFUSALS, never guesses.
func (t *Taxonomy) FolderRelpath(categoryName, subcategory string) (string, error) {
	cat := t.Category(categoryName)
	if cat == nil {
		return "", &TaxonomyError{
			Msg: "unknown category " + prompts.PyRepr(categoryName) +
				" — refusing to build a path",
		}
	}
	folder := storage.SafeFilename(cat.Name)
	if subcategory != "" {
		if !t.IsSubcategoryOf(categoryName, subcategory) {
			return "", &TaxonomyError{
				Msg: "unknown subcategory " + prompts.PyRepr(subcategory) +
					" for " + prompts.PyRepr(categoryName) +
					" — refusing to build a path",
			}
		}
		folder = osPathJoin(folder, storage.SafeFilename(subcategory))
	}
	return folder, nil
}

// osPathJoin mirrors os.path.join for the shapes folder_relpath builds
// (safe_filename components — no absolute parts, no separators inside):
// concatenate with the platform separator and DO NOT clean the result
// (an empty last component keeps its trailing separator, like Python).
func osPathJoin(elems ...string) string {
	var b strings.Builder
	for i, e := range elems {
		if i > 0 {
			b.WriteByte(os.PathSeparator)
		}
		b.WriteString(e)
	}
	return b.String()
}

// Summary mirrors summary(): "N categories, M subcategories from PATH".
func (t *Taxonomy) Summary() string {
	nSub := 0
	for _, c := range t.categories {
		nSub += len(c.Subcategories)
	}
	src := t.sourcePath
	if src == "" {
		src = "(inline)"
	}
	return strconv.Itoa(len(t.categories)) + " categories, " +
		strconv.Itoa(nSub) + " subcategories from " + src
}

// ---------------------------------------------------------------------------
// The parser (taxonomy.py:218-325).
// ---------------------------------------------------------------------------

// ignoredSections mirrors _IGNORED_SECTIONS (compared lower-cased).
var ignoredSections = map[string]bool{
	"how this works":  true,
	"not yet covered": true,
}

// rulesSection mirrors _RULES_SECTION (lower-cased form for comparison).
const rulesSection = "judgment-call rules"

// ParseTaxonomy mirrors parse_taxonomy: the tolerant, line-oriented
// grammar (see the spec's §Parsing grammar). Raises TaxonomyError when
// the file has no categories at all.
func ParseTaxonomy(text, sourcePath string) (*Taxonomy, error) {
	var categories []Category
	var rulesLines []string
	section := "" // current ## section title (lower-cased); "" = none
	current := -1 // index of the open category; -1 = none
	sawDefinition := false

	for _, rawLine := range pySplitLines(text) {
		line := pyRstrip(rawLine)

		// blank lines keep state
		if pyStrip(line) == "" {
			continue
		}

		if m := heading3RE.FindStringSubmatch(line); m != nil {
			name := stripCategoryHeading(m[1])
			if name == "" {
				return nil, &TaxonomyError{
					Msg: "category heading without a name: " + prompts.PyRepr(line),
				}
			}
			categories = append(categories, Category{
				Name:                   name,
				SubcategoryDefinitions: map[string]string{},
			})
			current = len(categories) - 1
			sawDefinition = false
			continue
		}

		if m := heading2RE.FindStringSubmatch(line); m != nil {
			section = strings.ToLower(pyStrip(m[1]))
			current = -1 // a ## ends the current category block
			continue
		}

		if pyStrip(line) == "---" {
			continue
		}

		if section != "" && ignoredSections[section] {
			continue // ignored section, skip wholesale
		}

		// inside the judgment-rules section: bullets' bodies and
		// non-*( prose appended VERBATIM (stripped).
		if section != "" && section == rulesSection {
			if m := bulletRE.FindStringSubmatch(line); m != nil {
				rulesLines = append(rulesLines, pyStrip(m[1]))
			} else if pyStrip(line) != "" && !strings.HasPrefix(line, "*(") {
				rulesLines = append(rulesLines, pyStrip(line))
			}
			continue
		}

		// category body lines are only meaningful when a category is
		// open; stray top-level lines elsewhere are ignored.
		if current == -1 {
			continue
		}
		cat := &categories[current]

		if m := exampleTagsRE.FindStringSubmatch(line); m != nil {
			cat.TagHints = append(cat.TagHints, stripTagsLine(m[1])...)
			continue
		}

		if m := bulletRE.FindStringSubmatch(line); m != nil {
			body := pyStrip(m[1])
			subName := stripSubcategory(body)
			if subName == "" ||
				strings.HasPrefix(strings.ToLower(subName), "(empty for now") {
				continue
			}
			// A bullet repeats the "(no subcategories)" marker in some
			// categories — never a subcategory.
			if noSubcategoriesRE.MatchString(body) {
				continue
			}
			if !containsString(cat.Subcategories, subName) {
				cat.Subcategories = append(cat.Subcategories, subName)
				desc := dashSplitRE.Split(strings.ReplaceAll(body, "**", ""), 2)
				if len(desc) == 2 && pyStrip(desc[1]) != "" {
					cat.SubcategoryDefinitions[subName] = pyStrip(desc[1])
				}
			}
			continue
		}

		// Prose paragraph: the first one is the category's definition.
		prose := pyStrip(line)
		if strings.HasPrefix(prose, "*(") && strings.HasSuffix(prose, ")*") {
			// Slicing off the emphasis also removes the parentheses, so
			// compare the marker CONTENT directly (test-caught).
			inner := pyStrip(prose[2 : len(prose)-2])
			if pureNoSubcategoriesRE.MatchString(inner) {
				continue // pure marker — nothing to keep
			}
			inner = pyStrip(leadingDashRE.ReplaceAllString(inner, ""))
			// NOTE: the italic-note path checks the DEFINITION (not
			// saw_definition) and does NOT set it — a later prose line
			// may still win (preserved quirk, golden-pinned).
			if inner != "" && cat.Definition == "" {
				cat.Definition = inner
			}
			continue
		}
		body := noSubcategoriesRE.ReplaceAllString(prose, "")
		body = pyStrip(strings.ReplaceAll(body, "**", ""))
		body = strings.Trim(body, "*")
		body = pyStrip(body)
		body = pyStrip(leadingDashRE.ReplaceAllString(body, ""))
		if body != "" && !sawDefinition {
			cat.Definition = body
			sawDefinition = true
		}
	}

	if len(categories) == 0 {
		return nil, &TaxonomyError{
			Msg: "no '### <Category>' headings found — not a taxonomy file",
		}
	}

	return newTaxonomy(categories, strings.Join(rulesLines, "\n"), sourcePath)
}

// ---------------------------------------------------------------------------
// Loaders (taxonomy.py:328-342).
// ---------------------------------------------------------------------------

// LoadTaxonomy mirrors load_taxonomy: read + parse the taxonomy file,
// decoding UTF-8 with a BOM stripped (utf-8-sig — a BOM would otherwise
// corrupt the first heading; the OPPOSITE quirk of the prompts read).
// An unreadable file is a TaxonomyError; invalid UTF-8 is NOT (Python's
// UnicodeDecodeError escapes `except OSError` — the decode error is
// returned as a plain error).
func LoadTaxonomy(path string) (*Taxonomy, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, &TaxonomyError{
			Msg: "cannot read taxonomy file " + prompts.PyRepr(path) + ": " +
				prompts.PyOSErrorText(err, path),
		}
	}
	raw = stripBOM(raw)
	if !utf8.Valid(raw) {
		return nil, pyUTF8DecodeError(raw)
	}
	return ParseTaxonomy(string(raw), path)
}

// LoadTaxonomyFromConfig mirrors load_taxonomy_from_config: the effective
// taxonomy — the config override, else the bundled default.
func LoadTaxonomyFromConfig(config map[string]any) (*Taxonomy, error) {
	return LoadTaxonomy(constants.ResolveTaxonomyPath(config))
}
