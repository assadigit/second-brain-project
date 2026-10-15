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
	"errors"
	"fmt"
	"io/fs"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"encoding/json"

	"github.com/assadigit/second-brain-project/assets"
	"github.com/assadigit/second-brain-project/internal/storage"
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
		text := pyStr(v)
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

// PyOSErrorText formats a read failure the way str(OSError) does in
// Python — "[Errno N] Message: 'PATH'" — for the common syscall shapes
// both a disk read (os.ReadFile) and an embed-FS lookup produce:
// missing → ENOENT; permission → EACCES; a directory read ("is a
// directory" — os and embed spell it the same) → EISDIR. Anything else
// falls back to the Go error text, still inside the caller's wrapper.
func PyOSErrorText(err error, path string) string {
	var pathErr *fs.PathError
	if errors.As(err, &pathErr) {
		switch {
		case errors.Is(err, fs.ErrNotExist):
			return "[Errno 2] No such file or directory: " + PyRepr(path)
		case errors.Is(err, fs.ErrPermission):
			return "[Errno 13] Permission denied: " + PyRepr(path)
		case strings.Contains(err.Error(), "is a directory"):
			return "[Errno 21] Is a directory: " + PyRepr(path)
		}
	}
	return err.Error()
}

// pyStr mirrors Python str() for the value shapes the loader meets:
// string as-is; bool → "True"/"False"; ints in decimal; floats (and
// JSON number literals) via Python float repr — str(5) == "5" but
// str(5.0) == "5.0", 1e15 positional, 1e-05 scientific.
func pyStr(v any) string {
	switch vv := v.(type) {
	case string:
		return vv
	case bool:
		if vv {
			return "True"
		}
		return "False"
	case json.Number:
		lit := vv.String()
		if !strings.ContainsAny(lit, ".eE") {
			return lit // integral literal → Python int str()
		}
		f, err := vv.Float64()
		if err != nil {
			return lit
		}
		return storage.PyFloatRepr(f)
	case float64:
		return storage.PyFloatRepr(vv)
	case float32:
		return storage.PyFloatRepr(float64(vv))
	case int:
		return strconv.Itoa(vv)
	case int8:
		return strconv.FormatInt(int64(vv), 10)
	case int16:
		return strconv.FormatInt(int64(vv), 10)
	case int32:
		return strconv.FormatInt(int64(vv), 10)
	case int64:
		return strconv.FormatInt(vv, 10)
	case uint:
		return strconv.FormatUint(uint64(vv), 10)
	case uint8:
		return strconv.FormatUint(uint64(vv), 10)
	case uint16:
		return strconv.FormatUint(uint64(vv), 10)
	case uint32:
		return strconv.FormatUint(uint64(vv), 10)
	case uint64:
		return strconv.FormatUint(vv, 10)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// pyRepr mirrors Python repr() for strings: single quotes preferred,
// switching to double quotes only when the string contains ' and no ";
// \\ , \n, \r, \t escapes; \xNN for non-printable ASCII (including DEL);
// \uNNNN / \UNNNNNNNN for non-printable runes; printable Unicode literal
// (Go's unicode.IsPrint matches Python's printable set for these cases:
// both exclude the format/control/private-use categories and non-ASCII
// space separators).
//
// Exported as PyRepr: the {x!r} error-message shape is shared by other
// parity packages (taxonomy's refusal messages). T-17's helpers package
// will centralize it (same promotion law as dryrun's formatComma).
func PyRepr(s string) string {
	quote := byte('\'')
	if strings.ContainsRune(s, '\'') && !strings.ContainsRune(s, '"') {
		quote = '"'
	}
	var b strings.Builder
	b.WriteByte(quote)
	for _, r := range s {
		switch {
		case r == rune(quote):
			b.WriteByte('\\')
			b.WriteByte(quote)
		case r == '\\':
			b.WriteString(`\\`)
		case r == '\n':
			b.WriteString(`\n`)
		case r == '\r':
			b.WriteString(`\r`)
		case r == '\t':
			b.WriteString(`\t`)
		case r < 0x20 || r == 0x7f:
			fmt.Fprintf(&b, `\x%02x`, r)
		case r < 0x7f:
			b.WriteRune(r)
		case unicode.IsPrint(r):
			b.WriteRune(r)
		case r < 0x100:
			// non-printable Latin-1 (nbsp, soft hyphen): \xNN, like Python
			fmt.Fprintf(&b, `\x%02x`, r)
		case r <= 0xffff:
			fmt.Fprintf(&b, `\u%04x`, r)
		default:
			fmt.Fprintf(&b, `\U%08x`, r)
		}
	}
	b.WriteByte(quote)
	return b.String()
}
