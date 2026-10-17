// fnmatch.go — CPython 3.12's fnmatch (Lib/fnmatch.py), ported for
// _untrack_local_only's victim matching (the reference imports
// fnmatch inside the method; the belt lowercases both sides before
// every call, so os.path.normcase — identity on POSIX — is the exact
// behavior; the Windows normcase is never observed through the belt).
//
// Go-note convergences (both documented in the spec's Go notes):
//   - translate() emits `(?>.*?fixed)` atomic groups; Go's RE2 has no
//     atomic groups — `(?:.*?fixed)` accepts the SAME language for
//     whole-string boolean matching (the atomicity is a backtracking
//     optimization, not a semantic).
//   - translate() emits `(?!)` for an empty class (never match); RE2
//     has no lookaheads — the never-match flag short-circuits FnMatch
//     to false (the class matches no character, so neither does the
//     whole pattern).
//   - `(?s:…)\Z` becomes `(?s:…)\z` (RE2's end-of-text; Python's \Z).
package vaultseal

import (
	"regexp"
	"strings"
	"sync"
)

// pyReEscapeSpecials mirrors Python 3.12 re.escape's special set:
// ()[]{}?*+-|^$\.&~# plus the ASCII whitespace set (space, \t, \n,
// \r, \v, \f) — every other rune passes through raw.
func pyReEscapeSpecials(s string) string {
	var out strings.Builder
	for _, r := range s {
		if pyReEscapeSet(r) {
			out.WriteRune('\\')
		}
		out.WriteRune(r)
	}
	return out.String()
}

func pyReEscapeSet(r rune) bool {
	switch r {
	case '(', ')', '[', ']', '{', '}', '?', '*', '+', '-', '|', '^',
		'$', '\\', '.', '&', '~', '#', ' ', '\t', '\n', '\r', '\v', '\f':
		return true
	}
	return false
}

// pyEscapeSetOps mirrors re.sub(r'([&~|])', r'\\\1', stuff).
func pyEscapeSetOps(s string) string {
	var out strings.Builder
	for _, r := range s {
		if r == '&' || r == '~' || r == '|' {
			out.WriteRune('\\')
		}
		out.WriteRune(r)
	}
	return out.String()
}

// fmToken is one translate() piece: a STAR marker or a fixed regex
// fragment.
type fmToken struct {
	star bool
	text string
}

var (
	fmCacheMu sync.RWMutex
	fmCache   = map[string]*regexp.Regexp{}
	// fmNever holds the patterns whose translation carries the
	// never-match empty class.
	fmNever = map[string]bool{}
)

// FnMatch mirrors fnmatch.fnmatch(name, pat) with the POSIX normcase
// (identity): translate the pattern, match the WHOLE name.
func FnMatch(name, pat string) bool {
	fmCacheMu.RLock()
	never := fmNever[pat]
	re := fmCache[pat]
	fmCacheMu.RUnlock()
	if never {
		return false
	}
	if re == nil {
		expr, ok := fnmatchCompile(pat)
		if !ok {
			fmCacheMu.Lock()
			fmNever[pat] = true
			fmCacheMu.Unlock()
			return false
		}
		fmCacheMu.Lock()
		fmCache[pat] = expr
		fmCacheMu.Unlock()
		re = expr
	}
	return re.MatchString(name)
}

// fnmatchCompile translates + compiles one pattern; ok=false is the
// never-match (empty class) case.
func fnmatchCompile(pat string) (*regexp.Regexp, bool) {
	tokens, never := fnmatchTranslateTokens(pat)
	if never {
		return nil, false
	}
	// Deal with STARs: fixed pieces at the start, then
	// STAR fixed (STAR fixed …) — the final regex is anchored to the
	// whole string by the caller's \A + the emitted \z.
	var res strings.Builder
	inp := tokens
	i := 0
	for i < len(inp) && !inp[i].star {
		res.WriteString(inp[i].text)
		i++
	}
	for i < len(inp) {
		// inp[i] is STAR
		i++
		if i == len(inp) {
			res.WriteString(".*")
			break
		}
		var fixed strings.Builder
		for i < len(inp) && !inp[i].star {
			fixed.WriteString(inp[i].text)
			i++
		}
		fixedS := fixed.String()
		if i == len(inp) {
			res.WriteString(".*")
			res.WriteString(fixedS)
		} else {
			// Python's (?>.*?fixed) — RE2: the non-capturing twin.
			res.WriteString("(?:.*?" + fixedS + ")")
		}
	}
	return regexp.MustCompile(`\A(?s:` + res.String() + `)\z`), true
}

// fnmatchTranslateTokens is the token-building half of CPython 3.12
// fnmatch.translate (the class grammar); never=true when an empty
// class `[]` appears (translate emits `(?!)` — never match).
func fnmatchTranslateTokens(pat string) ([]fmToken, bool) {
	runes := []rune(pat)
	var res []fmToken
	add := func(t fmToken) { res = append(res, t) }
	i, n := 0, len(runes)
	never := false
	for i < n {
		c := runes[i]
		i++
		switch c {
		case '*':
			// compress consecutive `*` into one
			if len(res) == 0 || !res[len(res)-1].star {
				add(fmToken{star: true})
			}
		case '?':
			add(fmToken{text: "."})
		case '[':
			j := i
			if j < n && runes[j] == '!' {
				j++
			}
			if j < n && runes[j] == ']' {
				j++
			}
			for j < n && runes[j] != ']' {
				j++
			}
			if j >= n {
				add(fmToken{text: `\[`})
			} else {
				stuff := string(runes[i:j])
				if !strings.Contains(stuff, "-") {
					stuff = strings.ReplaceAll(stuff, `\`, `\\`)
				} else {
					var chunks []string
					k := i + 1
					if runes[i] == '!' {
						k = i + 2
					}
					for {
						kk := indexRune(runes, '-', k, j)
						if kk < 0 {
							break
						}
						chunks = append(chunks, string(runes[i:kk]))
						i = kk + 1
						k = kk + 3
					}
					chunk := string(runes[i:j])
					if chunk != "" {
						chunks = append(chunks, chunk)
					} else if len(chunks) > 0 {
						chunks[len(chunks)-1] += "-"
					}
					// Remove empty ranges -- invalid in RE.
					for k := len(chunks) - 1; k > 0; k-- {
						if lastRune(chunks[k-1]) > firstRune(chunks[k]) {
							chunks[k-1] = dropLastRune(chunks[k-1]) + dropFirstRune(chunks[k])
							chunks = append(chunks[:k], chunks[k+1:]...)
						}
					}
					// Escape backslashes and hyphens for set difference
					// (--). Hyphens that create ranges shouldn't be escaped.
					var esc []string
					for _, s := range chunks {
						s = strings.ReplaceAll(s, `\`, `\\`)
						s = strings.ReplaceAll(s, "-", `\-`)
						esc = append(esc, s)
					}
					stuff = strings.Join(esc, "-")
				}
				// Escape set operations (&&, ~~ and ||).
				stuff = pyEscapeSetOps(stuff)
				i = j + 1
				if stuff == "" {
					// Empty range: never match.
					never = true
					add(fmToken{text: ""}) // the marker's text is never used
				} else if stuff == "!" {
					// Negated empty range: match any character.
					add(fmToken{text: "."})
				} else {
					stuffRunes := []rune(stuff)
					if stuffRunes[0] == '!' {
						stuff = "^" + string(stuffRunes[1:])
					} else if stuffRunes[0] == '^' || stuffRunes[0] == '[' {
						stuff = "\\" + stuff
					}
					add(fmToken{text: "[" + stuff + "]"})
				}
			}
		default:
			add(fmToken{text: pyReEscapeSpecials(string(c))})
		}
	}
	return res, never
}

// indexRune mirrors pat.find(ch, start, end) — the first ch in
// runes[start:end], or -1.
func indexRune(runes []rune, ch rune, start, end int) int {
	if start < 0 {
		start = 0
	}
	if end > len(runes) {
		end = len(runes)
	}
	for k := start; k < end; k++ {
		if runes[k] == ch {
			return k
		}
	}
	return -1
}

func lastRune(s string) rune {
	r := []rune(s)
	if len(r) == 0 {
		return 0
	}
	return r[len(r)-1]
}

func firstRune(s string) rune {
	r := []rune(s)
	if len(r) == 0 {
		return 0
	}
	return r[0]
}

func dropLastRune(s string) string {
	r := []rune(s)
	if len(r) == 0 {
		return ""
	}
	return string(r[:len(r)-1])
}

func dropFirstRune(s string) string {
	r := []rune(s)
	if len(r) == 0 {
		return ""
	}
	return string(r[1:])
}
