package connectioncheck

// result.go — the battery's result rows and their two renderings
// (connection_check.py's levels/icons/render_line/short_line, v0.37.0
// compact one-line law included).

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/assadigit/second-brain-project/internal/helpers"
)

// Levels (connection_check.py's constants).
const (
	LevelOK    = "ok"
	LevelWarn  = "warn"
	LevelError = "error"
	LevelInfo  = "info"
)

// ICONS maps each level to its log icon; unknown levels render "•".
var ICONS = map[string]string{
	LevelOK:    "✅",
	LevelWarn:  "⚠️",
	LevelError: "❌",
	LevelInfo:  "ℹ️",
}

// GUILevels maps battery levels onto the GUI log vocabulary
// (MainWindow.log_message levels — identical for the CLI).
var GUILevels = map[string]string{
	LevelOK:    "success",
	LevelWarn:  "warning",
	LevelError: "error",
	LevelInfo:  "info",
}

// Result is one check row: {'name', 'level', 'detail'} — one log line.
type Result struct {
	Name   string `json:"name"`
	Level  string `json:"level"`
	Detail string `json:"detail"`
}

// result builds one row (the _result helper).
func result(name, level, detail string) Result {
	return Result{Name: name, Level: level, Detail: detail}
}

// RenderLine is the human line: "✅ Name — detail" (em dash).
func RenderLine(r Result) string {
	icon := "•"
	if ic, ok := ICONS[r.Level]; ok {
		icon = ic
	}
	name := r.Name
	if name == "" {
		name = "?"
	}
	return icon + " " + name + " — " + r.Detail
}

// The v0.37.0 compact one-line budget (the WHOLE visible line) and the
// per-detail floor.
const (
	shortLineBudget    = 72
	shortLineMinDetail = 36
)

// parenIsPathy reports whether a parenthesized group's content looks
// like a location, not a fact: a scheme (https://…), a drive letter
// (G:\…, G:/…), or an absolute path (/home/…, \\server\…). Groups
// like "(5000 req/h)" or "(classic, 'repo' scope)" stay.
func parenIsPathy(inner string) bool {
	s := trimPySpace(pyStr(pyOr(inner, "")))
	if s == "" {
		return false
	}
	rs := []rune(s)
	return strings.Contains(s, "://") ||
		strings.Contains(s, ":\\") ||
		strings.Contains(s, ":/") ||
		strings.HasPrefix(s, "/") ||
		strings.HasPrefix(s, "\\") ||
		(len(rs) > 3 && rs[1] == ':' && (rs[2] == '/' || rs[2] == '\\'))
}

// pyParenSub is Python's re.sub(r"\s*\(([^()]*)\)", …, text) with the
// UNICODE \s class: every non-nested paren group is a match, consuming
// the whitespace run immediately before it. dropPathy decides the
// replacement — "" for pathy groups; the whole match (whitespace
// included) kept otherwise — reproducing the leftmost-match scan.
func pyParenSub(text string, dropPathy bool) string {
	rs := []rune(text)
	var out strings.Builder
	pending := []rune{} // whitespace run preceding the next token
	for i := 0; i < len(rs); {
		r := rs[i]
		switch {
		case isPySpace(r):
			pending = append(pending, r)
			i++
		case r == '(':
			// The group closes at the first ')' with no '(' between.
			j := i + 1
			closed := -1
			for ; j < len(rs); j++ {
				if rs[j] == '(' {
					break
				}
				if rs[j] == ')' {
					closed = j
					break
				}
			}
			if closed < 0 {
				// No match at this '(' — the regex retries later
				// positions; emit (whitespace first) and continue.
				out.WriteString(string(pending))
				pending = pending[:0]
				out.WriteRune(r)
				i++
				continue
			}
			inner := string(rs[i+1 : closed])
			if dropPathy && parenIsPathy(inner) {
				// Replace the whole match (pending + group) with "".
				pending = pending[:0]
			} else {
				out.WriteString(string(pending))
				pending = pending[:0]
				out.WriteString(string(rs[i : closed+1]))
			}
			i = closed + 1
		default:
			out.WriteString(string(pending))
			pending = pending[:0]
			out.WriteRune(r)
			i++
		}
	}
	out.WriteString(string(pending))
	return out.String()
}

// isPySpace mirrors Python str.isspace() (ASCII whitespace plus
// \x1c-\x1f, U+0085, U+00A0 and the Unicode space category).
func isPySpace(r rune) bool {
	switch r {
	case ' ', '\t', '\n', '\r', '\v', '\f',
		0x1c, 0x1d, 0x1e, 0x1f, 0x85, 0xa0:
		return true
	}
	return unicode.IsSpace(r)
}

// shortDetail collapses a check detail into one short visible
// fragment of at most budget runes (path-like parentheticals dropped,
// whitespace collapsed, hard mid-ellipsis over budget).
func shortDetail(detail string, budget int) string {
	text := trimPySpace(pyStr(pyOr(detail, "")))
	if text != "" {
		stripped := trimPySpace(pyParenSub(text, true))
		if stripped != "" {
			text = stripped
		}
	}
	text = strings.Join(strings.FieldsFunc(text, isPySpace), " ")
	if utf8.RuneCountInString(text) <= budget {
		return text
	}
	keepHead := maxInt(12, int(float64(budget)*0.68))
	keepTail := maxInt(8, budget-keepHead-1)
	trs := []rune(text)
	return helpers.Rstrip(string(trs[:keepHead])) + "…" +
		helpers.Lstrip(string(trs[len(trs)-keepTail:]))
}

// ShortLine is the compact single-line form: "✅ Name — short detail".
// One line ALWAYS: path-like parentheticals removed, a hard
// mid-ellipsis budget on the WHOLE visible line. The full detail
// stays available through RenderLine (the dialog's row tooltip).
func ShortLine(r Result) string {
	icon := "•"
	if ic, ok := ICONS[r.Level]; ok {
		icon = ic
	}
	name := pyStr(pyOr(r.Name, ""))
	if name == "" {
		name = "?"
	}
	budget := maxInt(shortLineMinDetail,
		shortLineBudget-utf8.RuneCountInString(name)-3)
	detail := shortDetail(r.Detail, budget)
	if detail == "" {
		return icon + " " + name
	}
	return icon + " " + name + " — " + detail
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
