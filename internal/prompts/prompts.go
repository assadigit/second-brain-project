// Package prompts ports app/gitcurator/core/prompts.py — the file-based
// prompt template loader with {{SLOT}} filling that REFUSES to run on
// unfilled slots, None/empty values, or values containing slot markers
// (the untrusted-input guard: a model answer or page text must never
// smuggle a new slot past validation).
//
// Spec: docs/port-blueprint/module-specs/spec-core-prompts.md;
// task T-13 (Phase 1). Parity is pinned by the generated golden in
// testdata/ (see prompts_test.go) and by the embedded prompt files in
// assets/prompts (sha256-pinned to the reference bytes).
//
// Two structural Go decisions (documented, behavior-preserving):
//
//   - The prompt files are EMBEDDED (assets/prompts, per the spec's Go
//     notes) instead of read from APP_DIR/prompts: the single-binary app
//     ships them inside the executable. The virtual root "prompts/"
//     plays PROMPTS_DIR's role in error messages. Reads still decode
//     plain UTF-8 with NO BOM stripping — the preserved Python quirk
//     (utf-8, not utf-8-sig; a BOM would corrupt slot 1) — and the tests
//     assert the shipped files carry no BOM.
//
//   - The filename grammar (os.path.basename(filename) != filename) is
//     pinned to NTPATH semantics: the reference app's production platform
//     is Windows, where os.path is ntpath ('\\' and '/' are separators, a
//     single-letter drive colon splits). The Go port is platform-stable
//     and replicates the production behavior; the Linux characterization
//     rig itself runs posixpath (the golden records both columns as
//     evidence — only the ntpath column is asserted).
package prompts

import (
	"fmt"
	"io/fs"
	"regexp"
	"strings"

	"github.com/assadigit/second-brain-project/assets"
	"github.com/assadigit/second-brain-project/internal/helpers"
)

// slotRE mirrors prompts.py _SLOT_RE: {{THIS_STYLE}} — an upper-case
// letter, then upper-case letters, digits and underscores. RE2-safe
// (doc 05 §3 item 8).
var slotRE = regexp.MustCompile(`\{\{([A-Z][A-Z0-9_]*)\}\}`)

// promptsRoot is the embedded FS's virtual directory — the stand-in for
// the reference's PROMPTS_DIR (constants.APP_DIR + "prompts") in error
// messages, matching the golden's dump-time normalization.
const promptsRoot = "prompts"

// PromptError mirrors prompts.py PromptError: a prompt template is
// missing or was left partially unfilled.
type PromptError struct {
	Msg string
}

// Error implements error.
func (e *PromptError) Error() string { return e.Msg }

// Slot is one named fill value. The Python API is keyword arguments
// (**values: str) — a dict whose iteration order is the CALL order, and
// the order the marker guard runs in; the slice preserves that order.
// Value nil mirrors Python None; non-string values are stringified with
// Python str() semantics (see pyStr).
type Slot struct {
	Name  string
	Value any
}

// SlotsIn mirrors slots_in: the ordered, de-duplicated slot names a
// template declares (first occurrence keeps its position). A nil/None
// template yields no slots (Python: `template or ”`).
func SlotsIn(template string) []string {
	var seen []string
	for _, name := range slotRE.FindAllStringSubmatch(template, -1) {
		if !contains(seen, name[1]) {
			seen = append(seen, name[1])
		}
	}
	return seen
}

func contains(list []string, s string) bool {
	for _, e := range list {
		if e == s {
			return true
		}
	}
	return false
}

// FillTemplate mirrors fill_template(template, **values).
//
// Refusals (all *PromptError, byte-exact vs the reference):
//   - template nil (Python None)      → "no template given"
//   - missing/None/empty slot values  → "prompt slot(s) left unfilled: NAMES"
//     (names joined ", " in SlotsIn order)
//   - None value for a NON-slot key   → "prompt slot NAME got None"
//     (a None value for a DECLARED slot is caught by the unfilled check
//     first — values[name] in (None, ""))
//   - value matching the slot pattern → "value for slot NAME contains a
//     {{SLOT}} marker — refusing to fill (untrusted input)" (the guard
//     runs over EVERY provided value, including unused keys)
//   - leftover slots after filling    → the unfilled error again (a value
//     may legitimately carry a partial marker such as "{{C" that only
//     becomes a slot after substitution)
//
// The template parameter is `any` so the None case stays expressible;
// non-string non-nil templates fail like Python's TypeError would
// (before any PromptError).
func FillTemplate(template any, slots ...Slot) (string, error) {
	if template == nil {
		return "", &PromptError{Msg: "no template given"}
	}
	tmpl, ok := template.(string)
	if !ok {
		return "", fmt.Errorf("findall() argument 1 must be str, not %T", template)
	}

	// Python's values dict: call order, later duplicate wins.
	names := make([]string, 0, len(slots))
	values := make(map[string]any, len(slots))
	for _, s := range slots {
		if _, seen := values[s.Name]; !seen {
			names = append(names, s.Name)
		}
		values[s.Name] = s.Value
	}

	var missing []string
	for _, name := range SlotsIn(tmpl) {
		v, ok := values[name]
		if !ok || v == nil || v == "" {
			missing = append(missing, name)
		}
	}
	if missing != nil {
		return "", &PromptError{
			Msg: "prompt slot(s) left unfilled: " + strings.Join(missing, ", "),
		}
	}

	for _, name := range names {
		v := values[name]
		if v == nil {
			return "", &PromptError{Msg: "prompt slot " + name + " got None"}
		}
		text := PyStr(v)
		if slotRE.MatchString(text) {
			return "", &PromptError{
				Msg: "value for slot " + name + " contains a {{SLOT}} marker " +
					"— refusing to fill (untrusted input)",
			}
		}
		tmpl = strings.ReplaceAll(tmpl, "{{"+name+"}}", text)
	}

	if leftovers := slotRE.FindAllStringSubmatch(tmpl, -1); leftovers != nil {
		var names []string
		for _, m := range leftovers {
			names = append(names, m[1])
		}
		return "", &PromptError{
			Msg: "prompt slot(s) left unfilled: " + strings.Join(names, ", "),
		}
	}
	return tmpl, nil
}

// LoadPromptFile mirrors load_prompt_file: reject path separators (the
// ntpath basename grammar — see the package doc), read the embedded
// template (plain UTF-8, no BOM handling), then FillTemplate.
func LoadPromptFile(filename string, slots ...Slot) (string, error) {
	if filename == "" || ntpathBasename(filename) != filename {
		return "", &PromptError{
			Msg: "bad prompt filename: " + PyRepr(filename),
		}
	}
	path := promptsRoot + "/" + filename
	raw, err := fs.ReadFile(assets.PromptsFS, path)
	if err != nil {
		return "", &PromptError{
			Msg: "cannot read prompt file " + PyRepr(path) + ": " + PyOSErrorText(err, path),
		}
	}
	return FillTemplate(string(raw), slots...)
}

// LoadPrompt mirrors load_prompt: "w01_category" → w01_category.txt.
func LoadPrompt(stem string, slots ...Slot) (string, error) {
	return LoadPromptFile(stem+".txt", slots...)
}

// ---------------------------------------------------------------------------
// Python-shape helpers (pinned by the golden corpus).
// ---------------------------------------------------------------------------

// ntpathBasename reproduces ntpath.basename (the production platform's
// os.path): both '\\' and '/' are separators; a single-character prefix
// before ':' is a drive and splits off before the separator scan (drive
// letters are not validated — "1:foo" behaves like "C:foo", matching
// ntpath.splitdrive). UNC prefixes need no special casing here: they
// contain separators, so the basename result is identical either way.
func ntpathBasename(p string) string {
	if len(p) >= 2 && p[1] == ':' {
		p = p[2:]
	}
	i := len(p)
	for i > 0 && p[i-1] != '\\' && p[i-1] != '/' {
		i--
	}
	return p[i:]
}

// PyOSErrorText delegates to the canonical helpers home (the T-17
// promotion: the implementation moved; this alias keeps the prompts API
// and its historical call sites stable).
func PyOSErrorText(err error, path string) string {
	return helpers.PyOSErrorText(err, path)
}

// PyStr delegates to the canonical helpers home (the T-17 promotion;
// nil keeps Python's own word "None" there — callers with special None
// grammar still check nil BEFORE calling, which FillTemplate does).
func PyStr(v any) string { return helpers.PyStr(v) }

// PyRepr delegates to the canonical helpers home (the T-17 promotion).
func PyRepr(s string) string { return helpers.PyRepr(s) }
