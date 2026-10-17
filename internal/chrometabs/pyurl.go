// pyurl.go — CPython 3.12's urlsplit, hand-rolled: Go's url.Parse
// REJECTS strings Python's urlparse accepts (a space in the host, a
// bare control character), and the fifth door's verdicts must not
// diverge on them. Only the pieces the module reads are parsed: the
// scheme (lowercased, when the run of chars is a valid scheme name)
// and the hostname (netloc after the last '@', brackets stripped,
// port split on the FIRST ':' — CPython _hostinfo's exact shapes).
//
// The netloc's bracket law rides with it: urlsplit RAISES on
// unbalanced brackets ("Invalid IPv6 URL") and, when both brackets
// exist, on a bracketed host that is not a valid IPv6 literal (or
// IPvFuture v-form) — and every caller catches that raise as "no
// parse" (page_is_real's except → proto=”), so an invalid netloc
// here reads as (scheme=”, hostname=”, hasNetloc=false).
package chrometabs

import (
	"net"
	"regexp"
	"strings"
)

// ipvFutureRe mirrors _check_bracketed_host's v-form:
// \Av[a-fA-F0-9]+\..+\Z.
var ipvFutureRe = regexp.MustCompile(`\Av[a-fA-F0-9]+\..+\z`)

// pyURLParts mirrors CPython urlsplit's lenient parsing. An invalid
// netloc (the bracket law) degrades to the callers' except-shape:
// everything empty.
func pyURLParts(rawURL string) (scheme, hostname string, hasNetloc bool) {
	u := rawURL
	// url.lstrip(_WHATWG_C0_CONTROL_OR_SPACE): leading C0 controls
	// and spaces go first…
	for len(u) > 0 && u[0] <= ' ' {
		u = u[1:]
	}
	// …then \t\r\n are removed everywhere (the WHATWG unsafe bytes).
	u = strings.Map(func(r rune) rune {
		switch r {
		case '\t', '\r', '\n':
			return -1
		}
		return r
	}, u)
	// the scheme: the first ':' with an ASCII-letter first char and a
	// run of [a-zA-Z0-9+-.] before it (url[:i].lower()).
	if i := strings.IndexByte(u, ':'); i > 0 && isASCIILetter(u[0]) {
		s := u[:i]
		ok := true
		for k := 0; k < len(s); k++ {
			if !isSchemeChar(s[k], k == 0) {
				ok = false
				break
			}
		}
		if ok {
			scheme = strings.ToLower(s)
			u = u[i+1:]
		}
	}
	if strings.HasPrefix(u, "//") {
		rest := u[2:]
		end := len(rest)
		for j := 0; j < len(rest); j++ {
			if rest[j] == '/' || rest[j] == '?' || rest[j] == '#' {
				end = j
				break
			}
		}
		netloc := rest[:end]
		hasNetloc = netloc != ""
		if !validPyNetloc(netloc) {
			// urlsplit raised — every caller's except reads "no parse"
			return "", "", false
		}
		// CPython _hostinfo: rpartition('@'), then the bracket or the
		// FIRST ':' partitions the port away; the hostname property
		// LOWERCASES (netloc keeps its case — only hostname folds).
		_, _, hostinfo := stringsLastPartition(netloc, '@')
		_, openBr, after := stringsPartition(hostinfo, '[')
		if openBr != "" {
			hostname, _, _ = stringsPartition(after, ']')
		} else {
			hostname, _, _ = stringsPartition(hostinfo, ':')
		}
		hostname = strings.ToLower(hostname)
	}
	return scheme, hostname, hasNetloc
}

func isASCIILetter(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// isSchemeChar answers whether c is in scheme_chars (the first char
// additionally must be an ASCII letter — url[0].isascii() and
// url[0].isalpha()).
func isSchemeChar(c byte, first bool) bool {
	if first {
		return isASCIILetter(c)
	}
	return isASCIILetter(c) || (c >= '0' && c <= '9') ||
		c == '+' || c == '-' || c == '.'
}

// validPyNetloc mirrors urlsplit's netloc laws: unbalanced brackets
// raise; when both brackets exist, _check_bracketed_netloc runs —
// nothing before the bracket, nothing between ']' and a ':' port,
// and the host (the bracketed literal, or the pre-colon text when
// the brackets live elsewhere) must be an IPv6 literal or an
// IPvFuture v-form (an IPv4 literal in brackets raises its own
// error). Best-effort equivalent on the IP parse: every shape
// CPython accepts as IPv6 (full/compressed hex, embedded dotted
// quad) parses in net.ParseIP with a ':' present; every shape it
// rejects (or accepts as IPv4) reads invalid here — the same
// no-parse outcome for the callers.
func validPyNetloc(netloc string) bool {
	hasOpen := strings.Contains(netloc, "[")
	hasClose := strings.Contains(netloc, "]")
	if hasOpen != hasClose {
		return false // "Invalid IPv6 URL" (unbalanced)
	}
	if !hasOpen {
		return true // no brackets — nothing more to check
	}
	_, _, hostinfo := stringsLastPartition(netloc, '@')
	before, openBr, bracketed := stringsPartition(hostinfo, '[')
	var host string
	if openBr != "" {
		if before != "" {
			return false // no data is allowed before a bracket
		}
		h, closeBr, port := stringsPartition(bracketed, ']')
		if closeBr != "" && port != "" && !strings.HasPrefix(port, ":") {
			return false // data after ']' before the port delimiter
		}
		host = h
	} else {
		// the brackets live before the '@' — the pre-colon text must
		// still survive _check_bracketed_host (an IPv literal or a
		// v-form; a normal hostname raises).
		host, _, _ = stringsPartition(hostinfo, ':')
	}
	// _check_bracketed_host
	if strings.HasPrefix(host, "v") {
		return ipvFutureRe.MatchString(host)
	}
	ip := net.ParseIP(host)
	return ip != nil && strings.Contains(host, ":")
}

// stringsPartition mirrors str.partition.
func stringsPartition(s string, sep byte) (before, sepStr, after string) {
	if i := strings.IndexByte(s, sep); i >= 0 {
		return s[:i], string(sep), s[i+1:]
	}
	return s, "", ""
}

// stringsLastPartition mirrors str.rpartition.
func stringsLastPartition(s string, sep byte) (before, sepStr, after string) {
	if i := strings.LastIndexByte(s, sep); i >= 0 {
		return s[:i], string(sep), s[i+1:]
	}
	return "", "", s
}
