// headers.go — the browser-grade request (v0.41.0): the Chrome-131
// header set, the bot-defense hint, the redirect target, the Referer
// law, Retry-After parsing, the HTTP error verdict table, the capped
// decompression, the PDF sniff, and the charset decode ladder.
package webfetch

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"golang.org/x/text/encoding/ianaindex"
)

// headerGetter — anything that can answer a header case-insensitively:
// http.Header (canonicalized), the tls-client response headers, and the
// test rig's map-backed getter all satisfy it (HTTPMessage.get
// semantics).
type headerGetter interface {
	Get(key string) string
}

// HeaderPair — one ordered header (the Chrome-131 set is built in
// Chrome's own order; the pure golden pins the pairs).
type HeaderPair struct {
	Key   string
	Value string
}

// RequestHeaders — Chrome 131's own navigation header set, in Chrome's
// own order. Chrome-only client hints (sec-ch-ua*) are dropped for a
// custom non-Chrome User-Agent so the request never contradicts itself.
func RequestHeaders(userAgent string) []HeaderPair {
	ua := userAgent
	if ua == "" {
		ua = UserAgent
	}
	headers := []HeaderPair{
		{"sec-ch-ua", `"Google Chrome";v="131", "Chromium";v="131", ` +
			`"Not_A Brand";v="24"`},
		{"sec-ch-ua-mobile", "?0"},
		{"sec-ch-ua-platform", `"Windows"`},
		{"Upgrade-Insecure-Requests", "1"},
		{"User-Agent", ua},
		{"Accept", "text/html,application/xhtml+xml,application/xml;q=0.9," +
			"image/avif,image/webp,image/apng,*/*;q=0.8," +
			"application/signed-exchange;v=b3;q=0.7"},
		{"Sec-Fetch-Site", "none"},
		{"Sec-Fetch-Mode", "navigate"},
		{"Sec-Fetch-User", "?1"},
		{"Sec-Fetch-Dest", "document"},
		{"Accept-Encoding", "gzip, deflate"},
		{"Accept-Language", "en-US,en;q=0.9"},
	}
	if !strings.Contains(ua, "Chrome/") {
		out := headers[:0]
		for _, p := range headers {
			if p.Key == "sec-ch-ua" || p.Key == "sec-ch-ua-mobile" ||
				p.Key == "sec-ch-ua-platform" {
				continue
			}
			out = append(out, p)
		}
		headers = out
	}
	return headers
}

// BotDefenseHint — a VISIBLE bot-defense marker on a 403/429 (""
// when nothing identifiable is on the wire). Names the wall in the
// failure reason.
func BotDefenseHint(headers headerGetter) string {
	if headers == nil {
		return ""
	}
	mitigated := strings.TrimSpace(headers.Get("cf-mitigated"))
	if mitigated != "" {
		return "bot defense (Cloudflare: " + mitigated + ")"
	}
	server := strings.ToLower(strings.TrimSpace(headers.Get("Server")))
	for _, b := range BotDefenseServers {
		if server == b {
			return "bot defense (server: " + server + ")"
		}
	}
	return ""
}

// RedirectTarget — the Location (or URI) header of a redirect answer,
// "" when the server sent none (spaces percent-encoded — the
// silent-give-up case urllib could not name).
func RedirectTarget(headers headerGetter) string {
	if headers == nil {
		return ""
	}
	for _, name := range []string{"Location", "URI"} {
		loc := strings.TrimSpace(headers.Get(name))
		if loc != "" {
			return strings.ReplaceAll(loc, " ", "%20")
		}
	}
	return ""
}

// HopHeaders — the header set for a redirect HOP: Chrome's navigation
// headers stay; Referer joins exactly the way a real browser sends it
// — the full URL for a same-host hop, the ORIGIN only cross-host
// (leaking the full URL cross-site is itself a bot tell). Referer is
// inserted right before Accept-Encoding (Chrome's order).
func HopHeaders(base []HeaderPair, prevURL, nextURL string) []HeaderPair {
	out := make([]HeaderPair, len(base))
	copy(out, base)
	prevHost := hostOf(prevURL)
	nextHost := hostOf(nextURL)
	referer := ""
	if prevHost != "" && prevHost == nextHost {
		referer = prevURL
	} else if u := schemeHostOf(prevURL); u != "" {
		referer = u + "/"
	}
	if referer == "" {
		return out
	}
	ordered := make([]HeaderPair, 0, len(out)+1)
	for _, p := range out {
		if p.Key == "Accept-Encoding" {
			ordered = append(ordered, HeaderPair{"Referer", referer})
		}
		ordered = append(ordered, p)
	}
	return ordered
}

// hostOf — the lowercased netloc (Python's urlparse().netloc).
func hostOf(raw string) string {
	if i := strings.Index(raw, "://"); i >= 0 {
		rest := raw[i+3:]
		if j := strings.IndexAny(rest, "/?#"); j >= 0 {
			rest = rest[:j]
		}
		return strings.ToLower(rest)
	}
	return ""
}

// schemeHostOf — "scheme://netloc" ("" when either part is missing).
func schemeHostOf(raw string) string {
	i := strings.Index(raw, "://")
	if i < 0 {
		return ""
	}
	rest := raw[i+3:]
	if j := strings.IndexAny(rest, "/?#"); j >= 0 {
		rest = rest[:j]
	}
	if rest == "" {
		return ""
	}
	return raw[:i+3] + rest
}

// ParseRetryAfter — the site's Retry-After (seconds or an HTTP-date)
// as seconds from NOW; nil when absent or unparsable. The politeness
// law: a site that says "come back in N" gets believed.
func ParseRetryAfter(headers headerGetter) *float64 {
	if headers == nil {
		return nil
	}
	raw := strings.TrimSpace(headers.Get("Retry-After"))
	if raw == "" {
		return nil
	}
	if secs, err := parseIntFloat(raw); err == nil {
		if secs < 0 {
			secs = 0
		}
		return &secs
	}
	for _, layout := range []string{time.RFC1123, time.RFC1123Z, time.ANSIC} {
		if when, err := time.Parse(layout, raw); err == nil {
			if when.Location() != time.UTC {
				// HTTP-dates are GMT; time.Parse with RFC1123 already
				// yields UTC unless the zone name was numeric.
			}
			secs := when.Sub(time.Now().UTC()).Seconds()
			if secs < 0 {
				secs = 0
			}
			return &secs
		}
	}
	return nil
}

// parseIntFloat — Python's int(raw) semantics for Retry-After values:
// a plain (possibly signed) integer string.
func parseIntFloat(raw string) (float64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, errors.New("empty")
	}
	neg := false
	switch raw[0] {
	case '-':
		neg = true
		raw = raw[1:]
	case '+':
		raw = raw[1:]
	}
	if raw == "" {
		return 0, errors.New("no digits")
	}
	n := 0.0
	for _, c := range raw {
		if c < '0' || c > '9' {
			return 0, errors.New("not an int")
		}
		n = n*10 + float64(c-'0')
	}
	if neg {
		n = -n
	}
	return n, nil
}

// HTTPErrorVerdict — one place that names every HTTP error the
// owner's pile showed: (reason, category, retry_after_s). Codes
// outside the named families keep the honest plain "HTTP {code}".
func HTTPErrorVerdict(code int, headers headerGetter) (string, string, *float64) {
	switch {
	case code == 402:
		return "HTTP 402 — paywalled (payment required; the content " +
			"is behind a paywall, not gone)", "paywalled", nil
	case code == 401:
		return "HTTP 401 — authorization required (the page is " +
			"private; credentials the fetcher does not have)", "refused", nil
	case code == 405:
		return "HTTP 405 — method refused (the server does not allow " +
			"GET here — often an API-only endpoint)", "refused", nil
	}
	if origin, ok := CloudflareOriginStatuses[code]; ok {
		return fmt.Sprintf("HTTP %d — Cloudflare: the site's own server "+
				"%s (site-side, retry later — not the fetcher's fault)", code, origin),
			"retry_later", nil
	}
	switch code {
	case 404, 410:
		gone := "the page is gone"
		if code == 410 {
			gone = "the page was removed"
		}
		return fmt.Sprintf("HTTP %d — %s", code, gone), "dead", nil
	case 429:
		retryAfter := ParseRetryAfter(headers)
		hint := BotDefenseHint(headers)
		if hint != "" {
			return "HTTP 429 — " + hint, "blocked_bot", retryAfter
		}
		if retryAfter != nil {
			return fmt.Sprintf("HTTP 429 — rate limited (Retry-After: "+
					"%.0fs — the site asked for a pause)", *retryAfter),
				"retry_later", retryAfter
		}
		return "HTTP 429 — rate limited (no Retry-After given; the " +
			"domain limiter backs off anyway)", "retry_later", nil
	case 503:
		retryAfter := ParseRetryAfter(headers)
		if retryAfter != nil {
			return fmt.Sprintf("HTTP 503 — service unavailable (Retry-After: "+
				"%.0fs)", *retryAfter), "retry_later", retryAfter
		}
		return "HTTP 503 — service unavailable (the site is " +
			"temporarily down or maintaining)", "retry_later", nil
	case 403:
		hint := BotDefenseHint(headers)
		if hint != "" {
			return "HTTP 403 — " + hint, "blocked_bot", nil
		}
		return "HTTP 403 — forbidden (the server refused this client " +
			"without naming a wall)", "blocked_bot", nil
	}
	category := "refused"
	if code >= 500 || code == 408 {
		category = "retry_later"
	}
	return fmt.Sprintf("HTTP %d", code), category, nil
}

// ---------------------------------------------------------------------------
// The capped decompression (v0.41.0: the fetcher asks for gzip like a
// browser, so it must also decode it — and a "gzip bomb" must stop at
// the same byte budget as everything else).
// ---------------------------------------------------------------------------

// unsupportedEncodingError — Python's ValueError shape for encodings
// the reference does not decode itself.
type unsupportedEncodingError struct{ encoding string }

func (e *unsupportedEncodingError) Error() string {
	return fmt.Sprintf("unsupported Content-Encoding: '%s'", e.encoding)
}

func isUnsupportedEncoding(err error) bool {
	var ue *unsupportedEncodingError
	return errors.As(err, &ue)
}

// DecompressCapped decodes a gzip/deflate body capped at maxBytes
// DEcompressed bytes. Returns (data, truncated); an undecodable stream
// errors (the caller turns that into an honest failed/partial result).
func DecompressCapped(body []byte, encodingName string, maxBytes int) ([]byte, bool, error) {
	enc := strings.ToLower(strings.TrimSpace(encodingName))
	switch enc {
	case "gzip", "x-gzip":
		return decompressStream(body, maxBytes, func(r io.Reader) (io.Reader, error) {
			return gzip.NewReader(r)
		})
	case "deflate":
		// zlib-wrapped first, then the raw-deflate fallback (the same
		// wbits ladder the reference walks).
		data, truncated, err := decompressStream(body, maxBytes, func(r io.Reader) (io.Reader, error) {
			return zlib.NewReader(r)
		})
		if err == nil {
			return data, truncated, nil
		}
		return decompressStream(body, maxBytes, func(r io.Reader) (io.Reader, error) {
			return flate.NewReader(r), nil
		})
	}
	return nil, false, &unsupportedEncodingError{encoding: encodingName}
}

// decompressStream reads 64 KB chunks through the codec, absorbing at
// maxBytes, mirroring the reference's chunked law.
func decompressStream(body []byte, maxBytes int, wrap func(io.Reader) (io.Reader, error)) ([]byte, bool, error) {
	src, err := wrap(bytes.NewReader(body))
	if err != nil {
		return nil, false, err
	}
	var out []byte
	truncated := false
	buf := make([]byte, 64*1024)
	for !truncated {
		n, rerr := src.Read(buf)
		if n > 0 {
			if len(out)+n > maxBytes {
				out = append(out, buf[:maxBytes-len(out)]...)
				truncated = true
				break
			}
			out = append(out, buf[:n]...)
		}
		if rerr != nil {
			if rerr == io.EOF {
				break
			}
			return nil, false, rerr
		}
	}
	return out, truncated, nil
}

// LooksLikePDF — the PDF sniff: the content type, the octet-stream +
// .pdf URL, or the %PDF- magic.
func LooksLikePDF(contentType, url string, body []byte) bool {
	if strings.Contains(contentType, "application/pdf") {
		return true
	}
	if strings.Contains(contentType, "octet-stream") &&
		strings.HasSuffix(strings.ToLower(strings.TrimRight(url, "/")), ".pdf") {
		return true
	}
	return len(body) >= 5 && string(body[:5]) == "%PDF-"
}

// ---------------------------------------------------------------------------
// The charset decode ladder
// ---------------------------------------------------------------------------

// decodeCharset — one codec attempt (nil when the charset is unknown
// or the decode fails; the ladder falls through to UTF-8).
func decodeCharset(body []byte, charset string) *string {
	cs := strings.TrimSpace(charset)
	if cs == "" {
		return nil
	}
	if strings.EqualFold(cs, "utf-8") || strings.EqualFold(cs, "utf8") {
		// errors="replace": invalid sequences become U+FFFD (Go's
		// string() conversion preserves raw bytes — sanitize).
		s := strings.ToValidUTF8(string(body), "\uFFFD")
		return &s
	}
	enc, err := ianaindex.IANA.Encoding(cs)
	if err != nil || enc == nil {
		return nil
	}
	decoded, err := enc.NewDecoder().Bytes(body)
	if err != nil {
		return nil
	}
	s := string(decoded)
	return &s
}

// bytesReplaceInvalid — the utf-8 errors="replace" tail: invalid
// sequences become U+FFFD (Go's string() conversion already replaces
// invalid bytes with U+FFFD — the exact same observable shape).
func bytesReplaceInvalid(body []byte) []byte { return body }

// pyErrClass maps a decode error to the Python exception class name
// the reference's own reason line prints ("{encoding} decode failed:
// {class}" — ValueError for unsupported encodings, BadGzipFile for
// gzip stream errors). The house precedent: the class prefix is part
// of the user-visible shape.
func pyErrClass(err error) string {
	if isUnsupportedEncoding(err) {
		return "ValueError"
	}
	return "BadGzipFile"
}
