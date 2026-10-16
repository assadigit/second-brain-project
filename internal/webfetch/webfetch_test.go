// webfetch_test.go — T-31's pure-logic batteries, ported FIRST from the
// Python characterization suites: test_webproxy.py's TestProxyFromConfig,
// test_ladder.py's TestHopHeaders/TestHttpVerdicts/TestCombinedCategory/
// TestRatePenalty, and the pure-golden driver (the corpus pinned by
// testdiff/goldens/dump_webfetch.py against the unmodified reference).
//
// The Python suites' release-bookkeeping classes (VERSION/CHANGELOG/CI/
// AGENTS/requirements contracts) are Python-repo bookkeeping — they stay
// out per the package-scope law (recorded in doc.go).
package webfetch

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// proxy_from_config — the config contract (test_webproxy.py)
// ---------------------------------------------------------------------------

func TestProxyFromConfig(t *testing.T) {
	cases := []struct {
		name string
		cfg  map[string]any
		want *Proxy
	}{
		{"missing_block", map[string]any{}, nil},
		{"disabled", map[string]any{"proxy": map[string]any{
			"enabled": false, "type": "socks5", "host": "127.0.0.1",
			"port": 10808}}, nil},
		{"use_for_web_false_opts_out", map[string]any{"proxy": map[string]any{
			"enabled": true, "type": "socks5", "host": "127.0.0.1",
			"port": 10808, "use_for_web": false}}, nil},
		{"use_for_web_MISSING_MEANS_ON", map[string]any{"proxy": map[string]any{
			"enabled": true, "type": "socks5", "host": "127.0.0.1",
			"port": 10808}},
			&Proxy{"socks5", "127.0.0.1", 10808, "", ""}},
		{"use_for_web_true", map[string]any{"proxy": map[string]any{
			"enabled": true, "use_for_web": true}},
			&Proxy{"socks5", "127.0.0.1", 10808, "", ""}},
		{"type_socks4", map[string]any{"proxy": map[string]any{
			"enabled": true, "type": "socks4"}},
			&Proxy{"socks4", "127.0.0.1", 10808, "", ""}},
		{"type_http", map[string]any{"proxy": map[string]any{
			"enabled": true, "type": "http"}},
			&Proxy{"http", "127.0.0.1", 10808, "", ""}},
		{"type_case_insensitive", map[string]any{"proxy": map[string]any{
			"enabled": true, "type": "  SOCKS5 "}},
			&Proxy{"socks5", "127.0.0.1", 10808, "", ""}},
		{"bad_type", map[string]any{"proxy": map[string]any{
			"enabled": true, "type": "mtproto"}}, nil},
		{"bad_port_string", map[string]any{"proxy": map[string]any{
			"enabled": true, "port": "nope"}}, nil},
		{"bad_port_zero", map[string]any{"proxy": map[string]any{
			"enabled": true, "port": 0}}, nil},
		{"bad_port_high", map[string]any{"proxy": map[string]any{
			"enabled": true, "port": 70000}}, nil},
		{"bad_port_negative", map[string]any{"proxy": map[string]any{
			"enabled": true, "port": -1}}, nil},
		{"port_empty_string_defaults", map[string]any{"proxy": map[string]any{
			"enabled": true, "port": ""}},
			&Proxy{"socks5", "127.0.0.1", 10808, "", ""}},
		{"port_string_number", map[string]any{"proxy": map[string]any{
			"enabled": true, "port": "3128"}},
			&Proxy{"socks5", "127.0.0.1", 3128, "", ""}},
		{"credentials_carried", map[string]any{"proxy": map[string]any{
			"enabled": true, "username": "u", "password": "p"}},
			&Proxy{"socks5", "127.0.0.1", 10808, "u", "p"}},
		{"host_defaults", map[string]any{"proxy": map[string]any{
			"enabled": true}},
			&Proxy{"socks5", "127.0.0.1", 10808, "", ""}},
		{"host_blank_rejected", map[string]any{"proxy": map[string]any{
			"enabled": true, "host": "   "}}, nil},
		{"not_a_dict_string", map[string]any{"proxy": "socks"}, nil},
		{"not_a_dict_list", map[string]any{"proxy": []any{1.0, 2.0}}, nil},
		{"enabled_missing", map[string]any{"proxy": map[string]any{
			"type": "socks5"}}, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ProxyFromConfig(tc.cfg)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("ProxyFromConfig(%v) = %+v, want %+v",
					tc.cfg, got, tc.want)
			}
		})
	}
	if ProxyFromConfig(nil) != nil {
		t.Fatal("ProxyFromConfig(nil) must be nil")
	}
}

func TestProxyFromConfigHostileNeverPanics(t *testing.T) {
	// The Boom config: every read raises (Python: a get() that throws).
	// Go's map read cannot throw; the equivalent hostile shape is a
	// config whose values are the wrong types everywhere — never a
	// panic, always nil.
	got := ProxyFromConfig(map[string]any{"proxy": map[string]any{
		"enabled": []any{"nested"}, "type": 1.5, "port": nil, "host": 0}})
	if got != nil {
		t.Fatalf("hostile config must yield nil, got %+v", got)
	}
}

func TestProxyLabel(t *testing.T) {
	if ProxyLabel(nil) != "none" {
		t.Fatal("nil proxy label")
	}
	if got := ProxyLabel(&Proxy{"socks5", "127.0.0.1", 10808, "", ""}); got != "SOCKS5 127.0.0.1:10808" {
		t.Fatalf("label = %q", got)
	}
	if got := ProxyLabel(&Proxy{"http", "p.example", 3128, "", ""}); got != "HTTP p.example:3128" {
		t.Fatalf("label = %q", got)
	}
}

// ---------------------------------------------------------------------------
// the request header set (v0.41.0's browser-grade law)
// ---------------------------------------------------------------------------

func TestRequestHeadersDefault(t *testing.T) {
	h := RequestHeaders(UserAgent)
	want := []HeaderPair{
		{"sec-ch-ua", `"Google Chrome";v="131", "Chromium";v="131", ` +
			`"Not_A Brand";v="24"`},
		{"sec-ch-ua-mobile", "?0"},
		{"sec-ch-ua-platform", `"Windows"`},
		{"Upgrade-Insecure-Requests", "1"},
		{"User-Agent", UserAgent},
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
	if !reflect.DeepEqual(h, want) {
		t.Fatalf("RequestHeaders default set/order mismatch:\n got: %#v\nwant: %#v", h, want)
	}
}

func TestRequestHeadersCustomDropsChromeHints(t *testing.T) {
	h := RequestHeaders("GitCurator-test/0.41")
	for _, p := range h {
		if strings.HasPrefix(strings.ToLower(p.Key), "sec-ch-ua") {
			t.Fatalf("custom non-Chrome UA must drop %q", p.Key)
		}
		if p.Key == "User-Agent" && p.Value != "GitCurator-test/0.41" {
			t.Fatalf("UA = %q", p.Value)
		}
	}
	found := false
	for _, p := range h {
		if p.Key == "Accept-Encoding" && p.Value == "gzip, deflate" {
			found = true
		}
	}
	if !found {
		t.Fatal("the rest of the browser set stays for a custom UA")
	}
}

func TestUserAgentConstant(t *testing.T) {
	// v0.41.0 — browser-grade Chrome; never the honest bot string.
	if !strings.Contains(UserAgent, "Chrome/131") || !strings.HasPrefix(UserAgent, "Mozilla/5.0") {
		t.Fatalf("USER_AGENT = %q", UserAgent)
	}
	if strings.Contains(UserAgent, "GitCurator") {
		t.Fatal("the UA must not name the app")
	}
}

// ---------------------------------------------------------------------------
// the Referer law (test_ladder.py's TestHopHeaders)
// ---------------------------------------------------------------------------

func TestHopHeadersSameHostFullURL(t *testing.T) {
	base := []HeaderPair{{"Accept", "text/html"}, {"Accept-Encoding", "gzip, deflate"}}
	h := HopHeaders(base, "https://example.org/a?q=1", "https://example.org/b")
	referer := ""
	acceptEncIdx := -1
	refererIdx := -1
	for i, p := range h {
		if p.Key == "Referer" {
			referer = p.Value
			refererIdx = i
		}
		if p.Key == "Accept-Encoding" {
			acceptEncIdx = i
		}
	}
	if referer != "https://example.org/a?q=1" {
		t.Fatalf("Referer = %q", referer)
	}
	// Chrome's order: Referer right before Accept-Encoding
	if refererIdx < 0 || acceptEncIdx < 0 || refererIdx > acceptEncIdx {
		t.Fatalf("Referer (%d) must precede Accept-Encoding (%d): %#v",
			refererIdx, acceptEncIdx, h)
	}
}

func TestHopHeadersCrossHostOriginOnly(t *testing.T) {
	base := []HeaderPair{{"Accept", "text/html"}, {"Accept-Encoding", "gzip, deflate"}}
	h := HopHeaders(base, "https://example.org/a?q=1", "https://other.example.net/b")
	referer := ""
	for _, p := range h {
		if p.Key == "Referer" {
			referer = p.Value
		}
	}
	if referer != "https://example.org/" {
		t.Fatalf("cross-host Referer = %q (origin only)", referer)
	}
}

func TestHopHeadersBadURLsNoReferer(t *testing.T) {
	base := []HeaderPair{{"Accept", "text/html"}}
	h := HopHeaders(base, "not a url", "")
	for _, p := range h {
		if p.Key == "Referer" {
			t.Fatal("bad URLs get no Referer")
		}
	}
}

// ---------------------------------------------------------------------------
// the HTTP verdicts (test_ladder.py's TestHttpVerdicts)
// ---------------------------------------------------------------------------

func TestHTTPErrorVerdicts(t *testing.T) {
	type verdict struct {
		reason     string
		category   string
		retryAfter *float64
	}
	cases := []struct {
		name string
		code int
		hdr  headerGetter
		want verdict
	}{
		{"402_paywalled", 402, hdr(),
			verdict{"HTTP 402 — paywalled (payment required; the content is behind a paywall, not gone)", "paywalled", nil}},
		{"401_private", 401, hdr(),
			verdict{"HTTP 401 — authorization required (the page is private; credentials the fetcher does not have)", "refused", nil}},
		{"405_method_refused", 405, hdr(),
			verdict{"HTTP 405 — method refused (the server does not allow GET here — often an API-only endpoint)", "refused", nil}},
		{"523_cloudflare_origin", 523, hdr(),
			verdict{"HTTP 523 — Cloudflare: the site's own server origin is unreachable (site-side, retry later — not the fetcher's fault)", "retry_later", nil}},
		{"429_retry_after", 429, hdr("Retry-After", "120"),
			verdict{"HTTP 429 — rate limited (Retry-After: 120s — the site asked for a pause)", "retry_later", floatPtr(120)}},
		{"429_no_retry_after", 429, hdr(),
			verdict{"HTTP 429 — rate limited (no Retry-After given; the domain limiter backs off anyway)", "retry_later", nil}},
		{"429_bot_defense_server", 429, hdr("Server", "cloudflare"),
			verdict{"HTTP 429 — bot defense (server: cloudflare)", "blocked_bot", nil}},
		{"429_bot_defense_cf_mitigated", 429, hdr("cf-mitigated", "challenge"),
			verdict{"HTTP 429 — bot defense (Cloudflare: challenge)", "blocked_bot", nil}},
		{"403_bot_defense_cf_mitigated", 403, hdr("cf-mitigated", "challenge"),
			verdict{"HTTP 403 — bot defense (Cloudflare: challenge)", "blocked_bot", nil}},
		{"403_without_marker", 403, hdr("Server", "nginx"),
			verdict{"HTTP 403 — forbidden (the server refused this client without naming a wall)", "blocked_bot", nil}},
		{"503_retry_after", 503, hdr("Retry-After", "30"),
			verdict{"HTTP 503 — service unavailable (Retry-After: 30s)", "retry_later", floatPtr(30)}},
		{"503_plain", 503, hdr(),
			verdict{"HTTP 503 — service unavailable (the site is temporarily down or maintaining)", "retry_later", nil}},
		{"404_dead", 404, hdr(),
			verdict{"HTTP 404 — the page is gone", "dead", nil}},
		{"410_dead", 410, hdr(),
			verdict{"HTTP 410 — the page was removed", "dead", nil}},
		{"500_retry_later", 500, hdr(), verdict{"HTTP 500", "retry_later", nil}},
		{"406_refused", 406, hdr(), verdict{"HTTP 406", "refused", nil}},
		{"408_retry_later", 408, hdr(), verdict{"HTTP 408", "retry_later", nil}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reason, category, ra := HTTPErrorVerdict(tc.code, tc.hdr)
			if reason != tc.want.reason {
				t.Fatalf("reason = %q\n want  %q", reason, tc.want.reason)
			}
			if category != tc.want.category {
				t.Fatalf("category = %q, want %q", category, tc.want.category)
			}
			if tc.want.retryAfter == nil {
				if ra != nil {
					t.Fatalf("retryAfter = %v, want nil", *ra)
				}
			} else if ra == nil || *ra != *tc.want.retryAfter {
				t.Fatalf("retryAfter = %v, want %v", ra, *tc.want.retryAfter)
			}
		})
	}
	// The Cloudflare origin family — every code names its truth.
	for code := 521; code <= 524; code++ {
		reason, category, _ := HTTPErrorVerdict(code, hdr())
		if category != "retry_later" || !strings.Contains(reason, "Cloudflare") ||
			!strings.Contains(reason, "site-side") {
			t.Fatalf("%d: reason=%q category=%q", code, reason, category)
		}
	}
}

func TestParseRetryAfter(t *testing.T) {
	if ra := ParseRetryAfter(hdr("Retry-After", "7")); ra == nil || *ra != 7 {
		t.Fatalf("seconds: %v", ra)
	}
	if ra := ParseRetryAfter(hdr("Retry-After", "0")); ra == nil || *ra != 0 {
		t.Fatalf("zero: %v", ra)
	}
	if ra := ParseRetryAfter(hdr("Retry-After", "-5")); ra == nil || *ra != 0 {
		t.Fatalf("negative capped: %v", ra)
	}
	// A future HTTP-date (90s out) parses into (60, 120).
	when := time.Now().UTC().Add(90 * time.Second).Format("Mon, 02 Jan 2006 15:04:05 GMT")
	ra := ParseRetryAfter(hdr("Retry-After", when))
	if ra == nil || *ra <= 60 || *ra >= 120 {
		t.Fatalf("future date: %v", ra)
	}
	// A past HTTP-date clamps to zero.
	if ra := ParseRetryAfter(hdr("Retry-After", "Tue, 07 Oct 2026 00:00:00 GMT")); ra == nil || *ra != 0 {
		t.Fatalf("past date: %v", ra)
	}
	// Garbage, empty, missing, nil headers.
	if ParseRetryAfter(hdr("Retry-After", "soon")) != nil {
		t.Fatal("garbage must be nil")
	}
	if ParseRetryAfter(hdr("Retry-After", "")) != nil {
		t.Fatal("empty must be nil")
	}
	if ParseRetryAfter(hdr()) != nil {
		t.Fatal("missing must be nil")
	}
	if ParseRetryAfter(nil) != nil {
		t.Fatal("nil headers must be nil")
	}
}

// ---------------------------------------------------------------------------
// the URL variants + hostnames (the rescue ladder's rungs)
// ---------------------------------------------------------------------------

func TestURLVariants(t *testing.T) {
	cases := []struct {
		name string
		url  string
		want []Variant
	}{
		{"no_slash_gets_slash", "https://example.org/page", []Variant{
			{"trailing-slash", "https://example.org/page/"},
			{"www", "https://www.example.org/page"}}},
		{"slash_dropped", "https://example.org/page/", []Variant{
			{"no-trailing-slash", "https://example.org/page"},
			{"www", "https://www.example.org/page/"}}},
		{"www_to_nonwww", "https://www.example.org/page", []Variant{
			{"trailing-slash", "https://www.example.org/page/"},
			{"non-www", "https://example.org/page"}}},
		{"ip_no_www", "https://192.0.2.10/page", []Variant{
			{"trailing-slash", "https://192.0.2.10/page/"}}},
		{"localhost_no_www", "http://localhost:8080/x", []Variant{
			{"trailing-slash", "http://localhost:8080/x/"}}},
		{"port_no_www", "https://example.org:8443/page", []Variant{
			{"trailing-slash", "https://example.org:8443/page/"}}},
		{"userinfo_no_www", "https://u@example.org/page", []Variant{
			{"trailing-slash", "https://u@example.org/page/"}}},
		{"ftp_none", "ftp://example.org/x", nil},
		{"empty", "", nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := URLVariants(tc.url)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("URLVariants(%q) = %#v, want %#v", tc.url, got, tc.want)
			}
		})
	}
}

func TestIsBareHostname(t *testing.T) {
	yes := []string{"example.org", "www.example.org"}
	no := []string{"", "u@example.org", "example.org:80", "localhost",
		"intranet", "192.0.2.10", "::1"}
	for _, h := range yes {
		if !IsBareHostname(h) {
			t.Fatalf("%q must be bare", h)
		}
	}
	for _, h := range no {
		if IsBareHostname(h) {
			t.Fatalf("%q must not be bare", h)
		}
	}
}

func TestIsLoopbackURL(t *testing.T) {
	yes := []string{"http://localhost/x", "localhost:8080", "localhost",
		"http://127.0.0.1/x", "http://127.1.2.3/x", "http://[::1]/x",
		"http://[::ffff:127.0.0.1]/x"}
	no := []string{"http://10.0.0.1/x", "http://example.com/x", "",
		"not a url at all"}
	for _, u := range yes {
		if !IsLoopbackURL(u) {
			t.Fatalf("%q must be loopback", u)
		}
	}
	for _, u := range no {
		if IsLoopbackURL(u) {
			t.Fatalf("%q must not be loopback", u)
		}
	}
	if IsLoopbackURL("") {
		t.Fatal("empty is not loopback")
	}
}

// ---------------------------------------------------------------------------
// decompression, PDF, decode
// ---------------------------------------------------------------------------

func TestDecompressCapped(t *testing.T) {
	gzOK := gzipBytes(t, []byte("hello compressed world "+
		strings.Repeat("x", 40)))
	data, truncated, err := DecompressCapped(gzOK, "gzip", 10_000)
	if err != nil || truncated || string(data) != "hello compressed world "+strings.Repeat("x", 40) {
		t.Fatalf("gzip ok: %d bytes, truncated=%v, err=%v", len(data), truncated, err)
	}
	// x-gzip alias
	data, _, err = DecompressCapped(gzOK, "x-gzip", 10_000)
	if err != nil || len(data) == 0 {
		t.Fatalf("x-gzip: err=%v", err)
	}
	// the bomb: capped at maxBytes DEcompressed
	bomb := gzipBytes(t, []byte(strings.Repeat("bomb ", 100_000)))
	data, truncated, err = DecompressCapped(bomb, "gzip", 50_000)
	if err != nil || !truncated || len(data) != 50_000 || !strings.HasPrefix(string(data), "bomb ") {
		t.Fatalf("bomb: %d bytes, truncated=%v, err=%v", len(data), truncated, err)
	}
	// unsupported encodings error
	if _, _, err := DecompressCapped([]byte("x"), "identity", 10); err == nil {
		t.Fatal("identity must error")
	}
	if _, _, err := DecompressCapped([]byte("x"), "br", 10); err == nil {
		t.Fatal("br must error")
	}
	// junk gzip errors
	if _, _, err := DecompressCapped([]byte("not gzip at all"), "gzip", 10_000); err == nil {
		t.Fatal("junk gzip must error")
	}
}

func TestDecompressCappedDeflate(t *testing.T) {
	// zlib-wrapped deflate (the server-side default)
	zl := zlibBytes(t, []byte("deflated content here "+strings.Repeat("y", 40)))
	data, truncated, err := DecompressCapped(zl, "deflate", 10_000)
	if err != nil || truncated || len(data) == 0 {
		t.Fatalf("zlib deflate: %d, %v, %v", len(data), truncated, err)
	}
	// RAW deflate (no zlib header) — the wbits fallback
	raw := rawDeflate(t, []byte("raw deflate payload "+strings.Repeat("z", 40)))
	data, truncated, err = DecompressCapped(raw, "deflate", 10_000)
	if err != nil || truncated || len(data) == 0 {
		t.Fatalf("raw deflate: %d, %v, %v", len(data), truncated, err)
	}
	// raw deflate capped
	bigRaw := rawDeflate(t, []byte(strings.Repeat("raw ", 100_000)))
	data, truncated, err = DecompressCapped(bigRaw, "deflate", 50_000)
	if err != nil || !truncated || len(data) != 50_000 {
		t.Fatalf("raw capped: %d, %v, %v", len(data), truncated, err)
	}
}

func TestLooksLikePDF(t *testing.T) {
	if !LooksLikePDF("application/pdf", "http://x/y", nil) {
		t.Fatal("content-type pdf")
	}
	if !LooksLikePDF("application/octet-stream", "http://x/doc.pdf", nil) {
		t.Fatal("octet-stream pdf url")
	}
	if !LooksLikePDF("application/octet-stream", "http://x/doc.pdf/", nil) {
		t.Fatal("octet-stream pdf url trailing slash")
	}
	if !LooksLikePDF("text/plain", "http://x/y", []byte("%PDF-1.7 rest")) {
		t.Fatal("magic bytes")
	}
	if LooksLikePDF("text/html", "http://x/y", []byte("<html>")) {
		t.Fatal("plain html is not pdf")
	}
	if LooksLikePDF("application/octet-stream", "http://x/doc.bin", nil) {
		t.Fatal("octet-stream non-pdf url")
	}
}

func TestDecodeBestEffort(t *testing.T) {
	if got := DecodeBestEffort([]byte("héllo wörld"), ""); got != "héllo wörld" {
		t.Fatalf("utf8: %q", got)
	}
	latin := []byte{0x44, 0x69, 0x73, 0x65, 0xF1, 0x6F} // "Diseño" in cp1252
	if got := DecodeBestEffort(latin, "windows-1252"); got != "Diseño" {
		t.Fatalf("windows-1252: %q", got)
	}
	if got := DecodeBestEffort([]byte("plain ascii"), "not-a-charset"); got != "plain ascii" {
		t.Fatalf("bad charset falls back to utf-8: %q", got)
	}
	if got := DecodeBestEffort([]byte{0x63, 0x61, 0x66, 0xE9}, "utf-8"); got != "caf\uFFFD" {
		t.Fatalf("invalid utf-8 replaced: %q", got)
	}
}

// ---------------------------------------------------------------------------
// the rate penalties + category combination (pure)
// ---------------------------------------------------------------------------

func TestRateLimitPenalty(t *testing.T) {
	if p := RateLimitPenalty(res("HTTP 429", intPtr(429), "retry_later", floatPtr(120))); p == nil || *p != 120 {
		t.Fatalf("ra 120: %v", p)
	}
	if p := RateLimitPenalty(res("HTTP 429", intPtr(429), "retry_later", nil)); p == nil || *p != RatePenaltyDefaultS {
		t.Fatalf("default: %v", p)
	}
	if p := RateLimitPenalty(res("HTTP 429", intPtr(429), "retry_later", floatPtr(99999))); p == nil || *p != 300 {
		t.Fatalf("cap: %v", p)
	}
	if p := RateLimitPenalty(res("HTTP 503", intPtr(503), "retry_later", floatPtr(30))); p == nil || *p != 30 {
		t.Fatalf("503 ra: %v", p)
	}
	if RateLimitPenalty(res("HTTP 403", intPtr(403), "", nil)) != nil {
		t.Fatal("403 pays nothing")
	}
	if RateLimitPenalty(res("timeout", nil, "retry_later", nil)) != nil {
		t.Fatal("timeout pays nothing")
	}
	if RateLimitPenalty(nil) != nil {
		t.Fatal("nil pays nothing")
	}
	if RateLimitPenalty(okRes("u", []byte("x"))) != nil {
		t.Fatal("ok pays nothing")
	}
}

func TestCombinedCategory(t *testing.T) {
	if got := CombinedCategory(res("HTTP 403", intPtr(403), "blocked_bot", nil),
		res("HTTP 404", intPtr(404), "dead", nil)); got != "dead" {
		t.Fatalf("dead beats wall: %q", got)
	}
	if got := CombinedCategory(res("HTTP 402", intPtr(402), "paywalled", nil),
		res("timeout", nil, "retry_later", nil)); got != "paywalled" {
		t.Fatalf("paywalled beats timeout: %q", got)
	}
	if got := CombinedCategory(res("HTTP 307 — broken redirect", intPtr(307), "redirect_broken", nil),
		res("timeout", nil, "retry_later", nil)); got != "redirect_broken" {
		t.Fatalf("redirect_broken beats retry: %q", got)
	}
	if got := CombinedCategory(res("x", nil, "", nil), nil); got != "" {
		t.Fatalf("no categories: %q", got)
	}
}

// ---------------------------------------------------------------------------
// the limiter laws (test_ladder.py's TestRatePenalty limiter legs)
// ---------------------------------------------------------------------------

func TestLimiterWaitHonorsPaidPenalty(t *testing.T) {
	lim := NewDomainRateLimiter(0)
	lim.Penalize("slow.example.net", 0.35)
	t0 := time.Now()
	lim.Wait("slow.example.net")
	if time.Since(t0) < 250*time.Millisecond {
		t.Fatal("the penalty must be waited out")
	}
	t1 := time.Now()
	lim.Wait("fast.example.net")
	if time.Since(t1) >= 100*time.Millisecond {
		t.Fatal("other domains are unaffected")
	}
}

func TestPenaltyNeverPanicsOnJunk(t *testing.T) {
	lim := NewDomainRateLimiter(0)
	lim.Penalize("", 10)
	lim.Penalize("x", -5)
	lim.mu.Lock()
	n := len(lim.penaltyUntil)
	lim.mu.Unlock()
	if n != 0 {
		t.Fatalf("junk penalties recorded: %d", n)
	}
}

func TestLimiterEnforcesDelay(t *testing.T) {
	lim := NewDomainRateLimiter(0.4)
	lim.Wait("example.com")
	t0 := time.Now()
	lim.Wait("example.com")
	if time.Since(t0) < 350*time.Millisecond {
		t.Fatal("second hit must wait")
	}
	t1 := time.Now()
	lim.Wait("other.com")
	if time.Since(t1) >= 200*time.Millisecond {
		t.Fatal("different domain: no wait")
	}
}

// ---------------------------------------------------------------------------
// the impersonation proxy URL law (test_thirddoor.py's TestProxyUrlLaw)
// ---------------------------------------------------------------------------

func TestProxyURLForImpersonator(t *testing.T) {
	if got := ProxyURLForImpersonator(&Proxy{"socks5", "p.example.net", 10808, "", ""}); got != "socks5h://p.example.net:10808" {
		t.Fatalf("socks5 becomes socks5h: %q", got)
	}
	if got := ProxyURLForImpersonator(&Proxy{"socks4", "h", 1, "", ""}); got != "socks4://h:1" {
		t.Fatalf("socks4: %q", got)
	}
	if got := ProxyURLForImpersonator(&Proxy{"http", "h", 8080, "u", "p"}); got != "http://u:p@h:8080" {
		t.Fatalf("http auth: %q", got)
	}
	if got := ProxyURLForImpersonator(&Proxy{"socks5", "h", 1, "a b/c", "p@ss:word"}); got != "socks5h://a%20b%2Fc:p%40ss%3Aword@h:1" {
		t.Fatalf("auth percent-encoded: %q", got)
	}
	if ProxyURLForImpersonator(nil) != "" {
		t.Fatal("nil proxy")
	}
	if ProxyURLForImpersonator(&Proxy{"bogus", "h", 1, "", ""}) != "" {
		t.Fatal("bogus type")
	}
	if ProxyURLForImpersonator(&Proxy{"", "h", 1, "", ""}) != "" {
		t.Fatal("empty type")
	}
}

// ---------------------------------------------------------------------------
// the TLS / DNS marker laws
// ---------------------------------------------------------------------------

func TestIsTLSHandshakeFailure(t *testing.T) {
	yes := []string{
		"proxy: connection: [SSL: SSLV3_ALERT_HANDSHAKE_FAILURE] sslv3 alert handshake failure (_ssl.c:1010) | direct: connection: [Errno 11001] getaddrinfo failed",
		"connection: alert handshake failure",
		"connection: handshake operation timed out",
		"connection: [SSL: UNEXPECTED_EOF_WHILE_READING] EOF occurred in violation of protocol",
		"connection: [SSL: WRONG_VERSION_NUMBER] tls server",
		"connection: [SSL: TLSV1_ALERT_PROTOCOL_VERSION]",
		"connection: sslv3 alert something",
		// the Go dialects (documented in doc.go — the stdlib door's
		// runtime speaks its own handshake-refusal shapes)
		"Get \"x\": remote error: tls: handshake failure",
		"net/http: TLS handshake timeout",
		"tls: first record does not look like a TLS handshake",
		"http: server gave HTTP response to HTTPS client",
		"tls: wrong version number",
	}
	no := []string{
		"connection: [SSL: CERTIFICATE_VERIFY_FAILED] certificate verify failed",
		"timeout", "",
	}
	for _, s := range yes {
		if !IsTLSHandshakeFailure(s) {
			t.Fatalf("%q must be a TLS handshake failure", s)
		}
	}
	for _, s := range no {
		if IsTLSHandshakeFailure(s) {
			t.Fatalf("%q must not be a TLS handshake failure", s)
		}
	}
}

func TestIsDNSFailure(t *testing.T) {
	yes := []string{
		"connection: [Errno 11001] getaddrinfo failed",
		"connection: lookup gone.example.net: no such host",
		"connection: name or service not known",
		"connection: temporary failure in name resolution",
		"connection: nodename nor servname provided",
		"connection: WSAHOST_NOT_FOUND",
		// the Go dialect (Go's resolver strings ride the Python markers)
		"dial tcp: lookup gone.example.net: no such host",
	}
	no := []string{"timeout", "connection: refused", ""}
	for _, s := range yes {
		if !IsDNSFailure(s) {
			t.Fatalf("%q must be a DNS failure", s)
		}
	}
	for _, s := range no {
		if IsDNSFailure(s) {
			t.Fatalf("%q must not be a DNS failure", s)
		}
	}
}

// ---------------------------------------------------------------------------
// the bot-defense hint + redirect target (pure)
// ---------------------------------------------------------------------------

func TestBotDefenseHint(t *testing.T) {
	if got := BotDefenseHint(hdr("cf-mitigated", "challenge")); got != "bot defense (Cloudflare: challenge)" {
		t.Fatalf("cf-mitigated: %q", got)
	}
	if got := BotDefenseHint(hdr("cf-mitigated", "  ")); got != "" {
		t.Fatalf("empty cf-mitigated: %q", got)
	}
	if got := BotDefenseHint(hdr("Server", "cloudflare")); got != "bot defense (server: cloudflare)" {
		t.Fatalf("server cloudflare: %q", got)
	}
	if got := BotDefenseHint(hdr("Server", "Akamai")); got != "bot defense (server: akamai)" {
		t.Fatalf("server case-folded: %q", got)
	}
	if got := BotDefenseHint(hdr("Server", "nginx")); got != "" {
		t.Fatalf("nginx: %q", got)
	}
	if got := BotDefenseHint(hdr()); got != "" {
		t.Fatalf("none: %q", got)
	}
	if got := BotDefenseHint(nil); got != "" {
		t.Fatalf("nil: %q", got)
	}
}

func TestRedirectTarget(t *testing.T) {
	if got := RedirectTarget(hdr("Location", "/next")); got != "/next" {
		t.Fatalf("location: %q", got)
	}
	if got := RedirectTarget(hdr("URI", "/next-uri")); got != "/next-uri" {
		t.Fatalf("uri fallback: %q", got)
	}
	if got := RedirectTarget(hdr("Location", "/path with spaces/next")); got != "/path%20with%20spaces/next" {
		t.Fatalf("spaces: %q", got)
	}
	if got := RedirectTarget(hdr("Location", "  ")); got != "" {
		t.Fatalf("blank: %q", got)
	}
	if got := RedirectTarget(hdr()); got != "" {
		t.Fatalf("missing: %q", got)
	}
	if got := RedirectTarget(nil); got != "" {
		t.Fatalf("nil: %q", got)
	}
}

func TestDomainOf(t *testing.T) {
	if got := DomainOf("https://Example.COM/path?q=1"); got != "example.com" {
		t.Fatalf("plain: %q", got)
	}
	if got := DomainOf("http://127.0.0.1:9944/x"); got != "127.0.0.1:9944" {
		t.Fatalf("port: %q", got)
	}
	if got := DomainOf("not a url"); got != "" {
		t.Fatalf("garbage: %q", got)
	}
}

// gzipBytes / zlibBytes / rawDeflate — the compression helpers for the
// decompression battery (the DEcompressed properties are the pin).
func gzipBytes(t *testing.T, payload []byte) []byte {
	t.Helper()
	return gzCompress(t, payload)
}

func zlibBytes(t *testing.T, payload []byte) []byte {
	t.Helper()
	return zlCompress(t, payload)
}

func rawDeflate(t *testing.T, payload []byte) []byte {
	t.Helper()
	return rawDeflateBytes(t, payload)
}
