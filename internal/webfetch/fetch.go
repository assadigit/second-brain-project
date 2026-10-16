// fetch.go — the fetch ladder: fetchOnce (one attempt: the browser-
// grade request, OUR redirect loop, the streamed size cap, the capped
// decompression, the PDF sniff, the failure taxonomy) and FetchURL
// (the whole ladder: rate-limit wait, the resource-truth branch with
// the dead-page rescue, the both-doors alternation, the third door,
// the final verdict).
package webfetch

import (
	"context"
	"errors"
	"io"
	"net/url"
	"strings"
	"time"
)

// fetchOnceFn — the seam signature (the Python suites patch
// _fetch_once with exactly this shape).
type fetchOnceFn func(url string, timeoutS float64, maxBytes int, userAgent string, proxy *Proxy, started time.Time) *FetchResult

// fetchOnceVar — the module-level seam (tests swap it; the ladder, the
// rescue ladder's variant asks, and the archive probes all ride it).
var fetchOnceVar fetchOnceFn = fetchOnce

// Options — FetchURL's knobs (nil = all defaults). The opt-out flags
// are inverted so the ZERO VALUE matches the reference's defaults
// (direct_fallback / impersonate_fallback / archive_fallback /
// doh_probe all default ON in Python).
type Options struct {
	TimeoutS           float64 // 0 → DefaultTimeoutS
	MaxBytes           int     // 0 → DefaultMaxBytes
	RateLimiter        RateLimiter
	UserAgent          string // "" → UserAgent
	Proxy              *Proxy // nil → no proxy
	NoDirectFallback   bool   // direct_fallback opt-out
	NoImpersonate      bool   // impersonate_fallback opt-out
	NoArchive          bool   // archive_fallback opt-out
	NoDoHProbe         bool   // doh_probe opt-out
	ImpersonateAutopip bool   // the no-op-by-default door-warmer (ADR-0006 §4)
}

// FetchURL fetches one URL politely. Never panics — every failure is a
// FetchResult(status "failed", reason, category).
//
// The ladder (v0.43–v0.47): the site's answer about the RESOURCE
// (404/401/5xx…) is the truth on any route — one attempt, with the
// dead-page rescue for 404/410; a route-aimed refusal
// (403/405/429/451) or a connection-class failure on the proxy route
// opens the DIRECT door; a wall on both stdlib doors asks the Chrome
// handshake (the third door); the final verdict pays DoH answers and
// rate penalties and names the Wayback courtesy.
func FetchURL(rawURL string, opts *Options) *FetchResult {
	if opts == nil {
		opts = &Options{}
	}
	timeoutS := opts.TimeoutS
	if timeoutS <= 0 {
		timeoutS = DefaultTimeoutS
	}
	maxBytes := opts.MaxBytes
	if maxBytes <= 0 {
		maxBytes = DefaultMaxBytes
	}
	userAgent := opts.UserAgent
	if userAgent == "" {
		userAgent = UserAgent
	}
	started := time.Now()
	domain := DomainOf(rawURL)
	if opts.RateLimiter != nil {
		opts.RateLimiter.Wait(domain)
	}

	if rawURL == "" || !strings.HasPrefix(strings.ToLower(rawURL), "http://") &&
		!strings.HasPrefix(strings.ToLower(rawURL), "https://") {
		return &FetchResult{
			URL: rawURL, FinalURL: rawURL, Status: "failed",
			Reason: "not an http(s) URL", Category: "bad_url",
			ElapsedS: elapsed(started),
		}
	}

	first := fetchOnceVar(rawURL, timeoutS, maxBytes, userAgent, opts.Proxy, started)
	if first.OK() {
		return first
	}
	if first.HTTPStatus != nil && !isRouteRefusal(*first.HTTPStatus) {
		// The site answered about the RESOURCE — the answer is the
		// truth on any route.
		if (*first.HTTPStatus == 404 || *first.HTTPStatus == 410) &&
			!IsLoopbackURL(rawURL) {
			rescue, tried, archive := deadPageRescue(
				rawURL, first, timeoutS, maxBytes, userAgent, opts.Proxy,
				opts.RateLimiter, !opts.NoArchive, started)
			if rescue != nil {
				return rescue
			}
			var extras []string
			if len(tried) > 0 {
				extras = append(extras, "variants tried: "+strings.Join(tried, ", "))
			}
			switch archive {
			case "no-snapshot":
				extras = append(extras, "no archived copy exists on the Wayback Machine")
			case "unreachable":
				extras = append(extras, "the Wayback Machine itself could not be asked")
			}
			if len(extras) > 0 {
				first.Reason = first.Reason + " (" + strings.Join(extras, "; ") + ")"
			}
			// the ladder found nothing — the dead verdict is final
			first.Category = "dead"
		}
		return finalVerdict(first, "", rawURL, opts.Proxy, opts.RateLimiter,
			domain, !opts.NoDoHProbe, []*FetchResult{first}, !opts.NoArchive, started)
	}
	if opts.Proxy == nil || opts.NoDirectFallback || IsLoopbackURL(rawURL) {
		// The single-door walled case still earns the third door; the
		// route list honors the direct opt-out.
		var routes []*Proxy
		if opts.Proxy != nil && opts.NoDirectFallback {
			routes = []*Proxy{opts.Proxy}
		} else {
			routes = []*Proxy{nil}
		}
		directReason := ""
		if opts.Proxy == nil {
			directReason = first.Reason
		}
		walled := maybeImpersonateDirect(rawURL, timeoutS, maxBytes, routes,
			!opts.NoImpersonate, first, started, opts.ImpersonateAutopip)
		return finalVerdict(walled, directReason, rawURL, opts.Proxy,
			opts.RateLimiter, domain, !opts.NoDoHProbe,
			[]*FetchResult{first}, !opts.NoArchive, started)
	}

	// The both-doors rule: the primary (proxied) route either never got
	// an answer or was REFUSED with a route-aimed status — the OTHER
	// door gets its one attempt: DIRECT.
	second := fetchOnceVar(rawURL, timeoutS, maxBytes, userAgent, nil, started)
	if second.OK() {
		second.Reason = "via direct fallback (proxy path failed: " +
			first.Reason + ")"
		return second
	}
	directReason := second.Reason
	second.Reason = "proxy: " + first.Reason + " | direct: " + directReason
	second.Category = CombinedCategory(first, second)
	// both stdlib doors walled: ask the Chrome handshake on the same
	// two routes (one polite attempt each).
	walled := maybeImpersonateDirect(rawURL, timeoutS, maxBytes,
		[]*Proxy{opts.Proxy, nil}, !opts.NoImpersonate, second, started,
		opts.ImpersonateAutopip)
	return finalVerdict(walled, directReason, rawURL, opts.Proxy,
		opts.RateLimiter, domain, !opts.NoDoHProbe,
		[]*FetchResult{first, second}, !opts.NoArchive, started)
}

func elapsed(started time.Time) float64 {
	return time.Since(started).Seconds()
}

// ---------------------------------------------------------------------------
// fetchOnce — one fetch attempt (the pre-v0.21.0 fetch_url body).
// ---------------------------------------------------------------------------

// fetchOnce never panics; started is the outer clock so elapsed
// covers fallbacks.
func fetchOnce(rawURL string, timeoutS float64, maxBytes int, userAgent string, p *Proxy, started time.Time) *FetchResult {
	headers := RequestHeaders(userAgent)

	// Public websites: certificates are verified (Go's default). A bad
	// cert is a failed fetch — not silently accepted.
	// Loopback → never ANY proxy; else the configured proxy.
	client, err := buildClient(p, IsLoopbackURL(rawURL), timeoutS)
	if err != nil {
		return &FetchResult{
			URL: rawURL, FinalURL: rawURL, Status: "failed",
			Reason: err.Error(), Category: "proxy_error",
			ElapsedS: elapsed(started),
		}
	}

	resp, finalURL, err := openFollowing(client, rawURL, headers)
	if err != nil {
		return fetchOnceError(rawURL, err, started)
	}
	defer func() { _ = resp.Body.Close() }()

	// Every HTTP error gets its reason and category from one place —
	// BEFORE any body read (urllib raises HTTPError out of the open;
	// the verdict is the truth, a broken error body must not turn into
	// a decode failure).
	if resp.StatusCode >= 400 {
		reason, category, retryAfter := HTTPErrorVerdict(resp.StatusCode, resp.Header)
		httpStatus := resp.StatusCode
		contentType := strings.ToLower(resp.Header.Get("Content-Type"))
		return &FetchResult{
			URL: rawURL, FinalURL: finalURL, Status: "failed",
			Reason: reason, HTTPStatus: &httpStatus, Category: category,
			RetryAfterS: retryAfter, ContentType: contentType,
			Charset:  charsetFromContentType(contentType),
			ElapsedS: elapsed(started),
		}
	}

	contentType := strings.ToLower(resp.Header.Get("Content-Type"))
	contentEncoding := strings.ToLower(strings.TrimSpace(resp.Header.Get("Content-Encoding")))
	charset := charsetFromContentType(contentType)
	httpStatus := resp.StatusCode

	// Stream with the size cap — never buffer a "huge page" whole.
	var chunks [][]byte
	total := 0
	truncated := false
	buf := make([]byte, 64*1024)
readLoop:
	for {
		n, rerr := resp.Body.Read(buf)
		if n > 0 {
			total += n
			if total > maxBytes {
				keep := n - (total - maxBytes)
				chunks = append(chunks, append([]byte(nil), buf[:keep]...))
				truncated = true
				break readLoop
			}
			chunks = append(chunks, append([]byte(nil), buf[:n]...))
		}
		if rerr != nil {
			if rerr != io.EOF && !isTimeoutErr(rerr) {
				return &FetchResult{
					URL: rawURL, FinalURL: finalURL, Status: "failed",
					Reason: rerr.Error(), Category: "retry_later",
					ElapsedS: elapsed(started),
				}
			}
			if isTimeoutErr(rerr) {
				return &FetchResult{
					URL: rawURL, FinalURL: finalURL, Status: "failed",
					Reason: "timeout", Category: "retry_later",
					ElapsedS: elapsed(started),
				}
			}
			break readLoop
		}
	}
	body := joinChunks(chunks)

	elapsedS := elapsed(started)

	// v0.41.0 — decode the compressed body; the cap bounds the
	// DEcompressed size too. A stream cut at the WIRE cap is
	// undecodable by construction — the size cap IS the honest reason.
	if contentEncoding != "" && contentEncoding != "identity" {
		if truncated {
			return &FetchResult{
				URL: rawURL, FinalURL: finalURL, Status: "partial",
				Reason:      "size cap (" + humanizeInt(maxBytes) + " bytes) reached",
				HTTPStatus:  &httpStatus,
				ContentType: contentType, Charset: charset,
				Body: nil, Text: "", ElapsedS: elapsedS,
			}
		}
		dec, decTruncated, derr := DecompressCapped(body, contentEncoding, maxBytes)
		if derr != nil {
			return &FetchResult{
				URL: rawURL, FinalURL: finalURL, Status: "failed",
				Reason:     contentEncoding + " decode failed: " + pyErrClass(derr),
				HTTPStatus: &httpStatus, ContentType: contentType,
				ElapsedS: elapsedS,
			}
		}
		body = dec
		truncated = truncated || decTruncated
	}

	if LooksLikePDF(contentType, rawURL, body) {
		return &FetchResult{
			URL: rawURL, FinalURL: finalURL, Status: "partial",
			Reason: "pdf", HTTPStatus: &httpStatus,
			ContentType: contentType, Charset: charset,
			ElapsedS: elapsedS,
		}
	}

	text := DecodeBestEffort(body, charset)

	if truncated {
		return &FetchResult{
			URL: rawURL, FinalURL: finalURL, Status: "partial",
			Reason:     "size cap (" + humanizeInt(maxBytes) + " bytes) reached",
			HTTPStatus: &httpStatus, ContentType: contentType,
			Charset: charset, Body: body, Text: text, ElapsedS: elapsedS,
		}
	}

	return &FetchResult{
		URL: rawURL, FinalURL: finalURL, Status: "full",
		Reason: "", HTTPStatus: &httpStatus, ContentType: contentType,
		Charset: charset, Body: body, Text: text, ElapsedS: elapsedS,
	}
}

// fetchOnceError — the connection-class verdicts: timeout, the proxy
// error family, and the generic connection shape.
func fetchOnceError(rawURL string, err error, started time.Time) *FetchResult {
	elapsedS := elapsed(started)
	if isTimeoutErr(err) {
		return &FetchResult{
			URL: rawURL, FinalURL: rawURL, Status: "failed",
			Reason: "timeout", Category: "retry_later", ElapsedS: elapsedS,
		}
	}
	var wpe *WebProxyError
	if errors.As(err, &wpe) {
		return &FetchResult{
			URL: rawURL, FinalURL: rawURL, Status: "failed",
			Reason: "connection: " + wpe.Error(), Category: "proxy_error",
			ElapsedS: elapsedS,
		}
	}
	var rbe *RedirectBrokenError
	if errors.As(err, &rbe) {
		code := rbe.Code
		return &FetchResult{
			URL: rawURL, FinalURL: rawURL, Status: "failed",
			Reason: rbe.Error(), HTTPStatus: &code,
			Category: "redirect_broken", ElapsedS: elapsedS,
		}
	}
	return &FetchResult{
		URL: rawURL, FinalURL: rawURL, Status: "failed",
		Reason: "connection: " + innerErrText(err), Category: "retry_later",
		ElapsedS: elapsedS,
	}
}

// isTimeoutErr — every timeout-class failure (the client's wall clock,
// context deadlines, net.Error timeouts, TLS handshake timeouts).
func isTimeoutErr(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "client.timeout") ||
		strings.Contains(msg, "timeout") && strings.Contains(msg, "exceeded") {
		return true
	}
	var ne interface{ Timeout() bool }
	if errors.As(err, &ne) {
		return ne.Timeout()
	}
	return false
}

// innerErrText — the url.Error's inner error (the method+URL prefix
// stays off the reason line, matching the reference's shape).
func innerErrText(err error) string {
	var ue *url.Error
	if errors.As(err, &ue) && ue.Err != nil {
		return ue.Err.Error()
	}
	return err.Error()
}

func joinChunks(chunks [][]byte) []byte {
	n := 0
	for _, c := range chunks {
		n += len(c)
	}
	out := make([]byte, 0, n)
	for _, c := range chunks {
		out = append(out, c...)
	}
	return out
}
func charsetFromContentType(contentType string) string {
	for _, part := range strings.Split(contentType, ";") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(strings.ToLower(part), "charset=") {
			cs := part[len("charset="):]
			cs = strings.TrimSpace(cs)
			cs = strings.Trim(cs, `"'`)
			return cs
		}
	}
	return ""
}
