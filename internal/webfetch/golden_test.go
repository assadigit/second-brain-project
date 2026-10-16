// golden_test.go — T-31: the generated-golden drivers.
//
//   - webfetch_pure.golden.json: the pure-logic corpus (constants, the
//     config contract, the header sets with ORDER, the verdict table,
//     the marker matrices, decompression, decode, penalties) pinned by
//     the reference, replayed by the port.
//   - webfetch_payloads.golden.json: THE WIRE CONTRACT — every
//     scenario's requests (method, path, headers) captured from the
//     unmodified reference against recording servers, replayed by the
//     port against the same fake-server shapes and compared after the
//     <ROOT>/<ROOT2>/<DOH>/<ARCHIVE>/<PROXY> normalization. The
//     Python runtime's own added headers (Host, Connection: close) are
//     filtered from the golden side — Go's net/http manages its own —
//     and header comparison is set-wise: Go's transport writes header
//     names in sorted order (the ORDER pin lives in the pure golden's
//     RequestHeaders pairs, the app's own construction order; the
//     fingerprint-order walls are the third door's jurisdiction).
//   - webfetch_behavior.golden.json: the fake-server behaviors (the
//     redirect truth, the phase-2 battery, DoH verdicts, the rescue
//     ladder, the penalties, the origin-down courtesy, the third
//     door's trigger matrix + verdicts + the autopip laws, the
//     both-doors alternation table, the preflight facts).
package webfetch

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// shared driver helpers
// ---------------------------------------------------------------------------

func loadGoldenFile(name string) (map[string]any, error) {
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		return nil, err
	}
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// goldenNormalize round-trips a value through encoding/json so it
// compares against the decoded golden (llmclient's driver law).
func goldenNormalize(v any) any {
	raw, err := json.Marshal(v)
	if err != nil {
		return v
	}
	var out any
	if err := json.Unmarshal(raw, &out); err != nil {
		return v
	}
	return out
}

func compareGolden(t *testing.T, path string, got, want any) {
	t.Helper()
	g := goldenNormalize(got)
	w := goldenNormalize(want)
	if !reflect.DeepEqual(g, w) {
		t.Fatalf("%s mismatch:\n got: %s\nwant: %s", path,
			mustJSON(g), mustJSON(w))
	}
}

func mustJSON(v any) string {
	data, _ := json.Marshal(v)
	return string(data)
}

// resToGolden — the generator's res_to_dict shape.
func resToGolden(r *FetchResult) map[string]any {
	if r == nil {
		return nil
	}
	return map[string]any{
		"url":           r.URL,
		"final_url":     r.FinalURL,
		"status":        r.Status,
		"reason":        r.Reason,
		"http_status":   r.HTTPStatus,
		"content_type":  r.ContentType,
		"charset":       r.Charset,
		"body_len":      len(r.Body),
		"category":      r.Category,
		"retry_after_s": r.RetryAfterS,
		"ok":            r.OK(),
		"text":          r.Text,
	}
}

// normalizeRoots — the <ROOT>-style placeholder replacement (the
// generator's normalize): roots maps port → tag.
func normalizeRoots(text string, roots map[int]string, extraPorts ...int) string {
	for port, tag := range roots {
		text = strings.ReplaceAll(text, "http://127.0.0.1:"+strconv.Itoa(port), "<"+tag+">")
		text = strings.ReplaceAll(text, "127.0.0.1:"+strconv.Itoa(port), "<"+tag+"-RAW>")
	}
	for _, port := range extraPorts {
		text = strings.ReplaceAll(text, "127.0.0.1:"+strconv.Itoa(port), "<DEADPORT>")
	}
	return text
}

func normResGolden(r *FetchResult, roots map[int]string, extraPorts ...int) map[string]any {
	d := resToGolden(r)
	d["url"] = normalizeRoots(d["url"].(string), roots, extraPorts...)
	d["final_url"] = normalizeRoots(d["final_url"].(string), roots, extraPorts...)
	d["reason"] = normalizeRoots(d["reason"].(string), roots, extraPorts...)
	d["text"] = normalizeRoots(d["text"].(string), roots, extraPorts...)
	return d
}

// goldenWireRequest — the Go side's captured request, in the golden's
// shape. The golden's Host/Connection pairs (the Python runtime's own)
// are filtered on the golden side by the comparison helper below.
func goldenWireRequest(rec wireRecord, roots map[int]string) map[string]any {
	pairs := []any{}
	for _, p := range rec.Headers {
		pairs = append(pairs, []any{p[0], normalizeRoots(p[1], roots)})
	}
	return map[string]any{
		"method":  rec.Method,
		"path":    rec.Path,
		"headers": pairs,
	}
}

// sortedHeaderPairs canonicalizes a header-pair list for set-wise
// comparison (sorts by name then value).
func sortedHeaderPairs(headers any) []any {
	pairs, _ := headers.([]any)
	out := make([]any, 0, len(pairs))
	for _, p := range pairs {
		out = append(out, p)
	}
	for i := 1; i < len(out); i++ {
		for j := i; j > 0; j-- {
			a, _ := out[j].([]any)
			b, _ := out[j-1].([]any)
			an, _ := a[0].(string)
			bn, _ := b[0].(string)
			av, _ := a[1].(string)
			bv, _ := b[1].(string)
			if bn > an || (bn == an && bv > av) {
				out[j], out[j-1] = out[j-1], out[j]
			} else {
				break
			}
		}
	}
	return out
}

// filterRuntimeHeaders drops the Python runtime's own added headers
// (Host, Connection: close) from a golden request's header pairs.
func filterRuntimeHeaders(headers any) []any {
	pairs, _ := headers.([]any)
	out := []any{}
	for _, p := range pairs {
		kv, ok := p.([]any)
		if !ok || len(kv) != 2 {
			continue
		}
		name, _ := kv[0].(string)
		if strings.EqualFold(name, "Host") || strings.EqualFold(name, "Connection") {
			continue
		}
		out = append(out, p)
	}
	return out
}

// impCallsGolden — the seam's calls in the Python fake's recorded
// shape: [url, {http,https} proxies dict | null, max_redirects].
func impCallsGolden(sessions []*fakeImpSession) []any {
	out := []any{}
	for _, s := range sessions {
		for _, c := range s.calls {
			var proxies any
			if c.ProxyURL != "" {
				proxies = map[string]any{"http": c.ProxyURL, "https": c.ProxyURL}
			}
			out = append(out, []any{c.URL, proxies, MaxRedirects})
		}
	}
	return out
}

// compression helpers mirroring the generator's inputs (the
// DEcompressed properties are the pin; the compressed bytes may differ
// by runtime).
func gzCompress(t *testing.T, payload []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	if _, err := w.Write(payload); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func zlCompress(t *testing.T, payload []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)
	if _, err := w.Write(payload); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func rawDeflateBytes(t *testing.T, payload []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	w, _ := flate.NewWriter(&buf, -1)
	if _, err := w.Write(payload); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// ---------------------------------------------------------------------------
// THE PURE GOLDEN
// ---------------------------------------------------------------------------

func TestGoldenPure(t *testing.T) {
	data, err := loadGoldenFile("webfetch_pure.golden.json")
	if err != nil {
		t.Fatal(err)
	}
	corpus := data["corpus"].(map[string]any)
	check := func(path string, got any) {
		t.Helper()
		compareGolden(t, "pure."+path, got, corpus[path])
	}

	// constants
	check("constants", map[string]any{
		"USER_AGENT":                  UserAgent,
		"DEFAULT_PROXY_TYPE":          DefaultProxyType,
		"DEFAULT_PROXY_HOST":          DefaultProxyHost,
		"DEFAULT_PROXY_PORT":          DefaultProxyPort,
		"PROXY_PREFLIGHT_TIMEOUT_S":   ProxyPreflightTimeoutS,
		"DEFAULT_TIMEOUT_S":           DefaultTimeoutS,
		"DEFAULT_MAX_BYTES":           DefaultMaxBytes,
		"DEFAULT_DOMAIN_DELAY_S":      DefaultDomainDelayS,
		"MAX_REDIRECTS":               MaxRedirects,
		"_REDIRECT_STATUSES":          RedirectStatuses,
		"_CLOUDFLARE_ORIGIN_STATUSES": CloudflareOriginStatuses,
		"_TLS_HANDSHAKE_MARKERS":      TLSHandshakeMarkers,
		"_DNS_FAIL_MARKERS":           DNSFailMarkers,
		"_CATEGORY_PRECEDENCE":        CategoryPrecedence,
		"_NO_HEAL_CATEGORIES":         NoHealCategories,
		"_RATE_PENALTY_DEFAULT_S":     RatePenaltyDefaultS,
		"_RATE_PENALTY_CAP_S":         RatePenaltyCapS,
		"_DOH_ENDPOINT":               DoHEndpoint,
		"_DOH_TIMEOUT_S":              DoHTimeoutS,
		"_ARCHIVE_API":                ArchiveAPI,
		"_ARCHIVE_MAX_BYTES":          ArchiveMaxBytes,
		"_ORIGIN_ARCHIVE_TIMEOUT_S":   OriginArchiveTimeoutS,
		"ROUTE_REFUSAL_STATUSES":      RouteRefusalStatuses,
		"IMPERSONATE_TARGET":          ImpersonateTarget,
		"_IMP_INSTALL_HINT":           ImpInstallHint,
		"_CHALLENGE_WON_HINT":         ChallengeWonHint,
		"_AUTOPIP_TIMEOUT_S":          AutopipTimeoutS,
		"_BOT_DEFENSE_SERVERS":        BotDefenseServers,
		"_HTML_TYPES":                 HTMLTypes,
		"_PROXY_TYPES":                ProxyTypes,
	})

	// proxy_from_config — the whole config contract
	cfgs := map[string]map[string]any{
		"missing_block":                {},
		"none_config":                  nil,
		"not_a_dict_string":            {"proxy": "socks"},
		"not_a_dict_list":              {"proxy": []any{1.0, 2.0}},
		"disabled":                     {"proxy": m("enabled", false, "type", "socks5", "host", "127.0.0.1", "port", 10808)},
		"use_for_web_false":            {"proxy": m("enabled", true, "type", "socks5", "host", "127.0.0.1", "port", 10808, "use_for_web", false)},
		"use_for_web_missing_means_on": {"proxy": m("enabled", true, "type", "socks5", "host", "127.0.0.1", "port", 10808)},
		"use_for_web_true":             {"proxy": m("enabled", true, "use_for_web", true)},
		"type_socks5":                  {"proxy": m("enabled", true, "type", "socks5")},
		"type_socks4":                  {"proxy": m("enabled", true, "type", "socks4")},
		"type_http":                    {"proxy": m("enabled", true, "type", "http")},
		"type_case_insensitive":        {"proxy": m("enabled", true, "type", "  SOCKS5 ")},
		"bad_type":                     {"proxy": m("enabled", true, "type", "mtproto")},
		"bad_port_string":              {"proxy": m("enabled", true, "port", "nope")},
		"bad_port_zero":                {"proxy": m("enabled", true, "port", 0)},
		"bad_port_high":                {"proxy": m("enabled", true, "port", 70000)},
		"bad_port_negative":            {"proxy": m("enabled", true, "port", -1)},
		"port_empty_string_defaults":   {"proxy": m("enabled", true, "port", "")},
		"port_none_defaults":           {"proxy": m("enabled", true, "port", nil)},
		"port_string_number":           {"proxy": m("enabled", true, "port", "3128")},
		"credentials_carried":          {"proxy": m("enabled", true, "username", "u", "password", "p")},
		"host_defaults":                {"proxy": m("enabled", true)},
		"host_blank_rejected":          {"proxy": m("enabled", true, "host", "   ")},
		"enabled_missing":              {"proxy": m("type", "socks5")},
		"enabled_truthy_string":        {"proxy": m("enabled", "yes", "type", "http")},
	}
	gotCfg := map[string]any{}
	for name, cfg := range cfgs {
		gotCfg[name] = ProxyFromConfig(cfg)
	}
	check("proxy_from_config", gotCfg)

	check("proxy_label", map[string]any{
		"none":   ProxyLabel(nil),
		"socks5": ProxyLabel(&Proxy{"socks5", "127.0.0.1", 10808, "", ""}),
		"http":   ProxyLabel(&Proxy{"http", "p.example", 3128, "", ""}),
	})

	// the request header sets (ORDER included — pairs)
	pairs := func(h []HeaderPair) []any {
		out := []any{}
		for _, p := range h {
			out = append(out, []any{p.Key, p.Value})
		}
		return out
	}
	check("request_headers_default", pairs(RequestHeaders(UserAgent)))
	check("request_headers_custom_non_chrome", pairs(RequestHeaders("GitCurator-test/0.41")))
	check("request_headers_empty_ua_falls_back", pairs(RequestHeaders("")))
	check("request_headers_firefox_drops_hints", pairs(RequestHeaders(
		"Mozilla/5.0 (X11; Linux x86_64; rv:132.0) Gecko/20100101 Firefox/132.0")))

	hopGolden := func(h []HeaderPair) map[string]any {
		return map[string]any{"headers": pairs(h)}
	}
	base := []HeaderPair{{"Accept", "text/html"}, {"Accept-Encoding", "gzip, deflate"}}
	check("hop_headers", map[string]any{
		"same_host_full_url":     hopGolden(HopHeaders(base, "https://example.org/a?q=1", "https://example.org/b")),
		"cross_host_origin_only": hopGolden(HopHeaders(base, "https://example.org/a?q=1", "https://other.example.net/b")),
		"bad_urls_no_referer":    hopGolden(HopHeaders(base, "not a url", "")),
		"no_base_referer_absent": hopGolden(HopHeaders(nil, "https://x.test/a", "https://x.test/b")),
	})

	// the HTTP error verdict table
	hev := map[string]any{}
	for _, code := range []int{400, 401, 402, 403, 404, 405, 406, 408, 410, 418, 429, 451, 500, 502, 503, 504, 520, 521, 522, 523, 524, 599} {
		reason, cat, ra := HTTPErrorVerdict(code, hdr())
		hev[fmt.Sprintf("%d_plain", code)] = verdictTriple(reason, cat, ra)
	}
	for _, code := range []int{403, 429} {
		for name, h := range map[string]headerGetter{
			"server_cloudflare":      hdr("Server", "cloudflare"),
			"cf_mitigated_challenge": hdr("cf-mitigated", "challenge"),
			"server_nginx":           hdr("Server", "nginx"),
			"server_uppercase":       hdr("Server", "CloudFlare"),
		} {
			reason, cat, ra := HTTPErrorVerdict(code, h)
			hev[fmt.Sprintf("%d_%s", code, name)] = verdictTriple(reason, cat, ra)
		}
	}
	reason, cat, ra := HTTPErrorVerdict(429, hdr("Retry-After", "120"))
	hev["429_retry_after_120"] = verdictTriple(reason, cat, ra)
	reason, cat, ra = HTTPErrorVerdict(429, hdr("Retry-After", "0"))
	hev["429_retry_after_zero"] = verdictTriple(reason, cat, ra)
	reason, cat, ra = HTTPErrorVerdict(503, hdr("Retry-After", "30"))
	hev["503_retry_after_30"] = verdictTriple(reason, cat, ra)
	reason, cat, ra = HTTPErrorVerdict(503, hdr("Retry-After", "soon"))
	hev["503_retry_after_garbage"] = verdictTriple(reason, cat, ra)
	check("http_error_verdict", hev)

	// parse_retry_after — deterministic shapes + future-date booleans
	raFuture := ParseRetryAfter(hdr("Retry-After",
		time.Now().UTC().Add(90*time.Second).Format("Mon, 02 Jan 2006 15:04:05 GMT")))
	check("parse_retry_after", map[string]any{
		"seconds_7":            ParseRetryAfter(hdr("Retry-After", "7")),
		"seconds_0":            ParseRetryAfter(hdr("Retry-After", "0")),
		"negative_capped":      ParseRetryAfter(hdr("Retry-After", "-5")),
		"garbage":              ParseRetryAfter(hdr("Retry-After", "soon")),
		"empty":                ParseRetryAfter(hdr("Retry-After", "")),
		"missing":              ParseRetryAfter(hdr()),
		"none_headers":         ParseRetryAfter(nil),
		"past_date_zero":       ParseRetryAfter(hdr("Retry-After", "Tue, 07 Oct 2026 00:00:00 GMT")),
		"future_date_positive": raFuture != nil && *raFuture > 0.0,
		"future_date_lt_120":   raFuture != nil && *raFuture < 120.0,
	})

	check("bot_defense_hint", map[string]any{
		"cf_mitigated":       BotDefenseHint(hdr("cf-mitigated", "challenge")),
		"cf_mitigated_empty": BotDefenseHint(hdr("cf-mitigated", "  ")),
		"server_cloudflare":  BotDefenseHint(hdr("Server", "cloudflare")),
		"server_akamai":      BotDefenseHint(hdr("Server", "Akamai")),
		"server_nginx":       BotDefenseHint(hdr("Server", "nginx")),
		"none":               BotDefenseHint(hdr()),
		"none_headers":       BotDefenseHint(nil),
	})

	check("redirect_target", map[string]any{
		"location":             RedirectTarget(hdr("Location", "/next")),
		"uri_fallback":         RedirectTarget(hdr("URI", "/next-uri")),
		"location_with_spaces": RedirectTarget(hdr("Location", "/path with spaces/next")),
		"empty":                RedirectTarget(hdr("Location", "  ")),
		"missing":              RedirectTarget(hdr()),
		"none_headers":         RedirectTarget(nil),
	})

	variantsGolden := func(vs []Variant) []any {
		out := []any{}
		for _, v := range vs {
			out = append(out, []any{v.Label, v.URL})
		}
		return out
	}
	check("url_variants", map[string]any{
		"no_slash_gets_slash": variantsGolden(URLVariants("https://example.org/page")),
		"slash_dropped":       variantsGolden(URLVariants("https://example.org/page/")),
		"www_to_nonwww":       variantsGolden(URLVariants("https://www.example.org/page")),
		"bare_to_www":         variantsGolden(URLVariants("https://example.org/a?q=1#f")),
		"ip_no_www":           variantsGolden(URLVariants("https://192.0.2.10/page")),
		"localhost_no_www":    variantsGolden(URLVariants("http://localhost:8080/x")),
		"port_no_www":         variantsGolden(URLVariants("https://example.org:8443/page")),
		"userinfo_no_www":     variantsGolden(URLVariants("https://u@example.org/page")),
		"ftp_none":            variantsGolden(URLVariants("ftp://example.org/x")),
		"empty":               variantsGolden(URLVariants("")),
		"root_slash_only":     variantsGolden(URLVariants("https://example.org/")),
	})

	check("is_bare_hostname", map[string]any{
		"plain":     IsBareHostname("example.org"),
		"www":       IsBareHostname("www.example.org"),
		"empty":     IsBareHostname(""),
		"userinfo":  IsBareHostname("u@example.org"),
		"port":      IsBareHostname("example.org:80"),
		"localhost": IsBareHostname("localhost"),
		"no_dot":    IsBareHostname("intranet"),
		"ipv4":      IsBareHostname("192.0.2.10"),
		"ipv6":      IsBareHostname("::1"),
	})

	check("is_loopback_url", map[string]any{
		"localhost":            IsLoopbackURL("http://localhost/x"),
		"localhost_no_scheme":  IsLoopbackURL("localhost:8080"),
		"bare_localhost":       IsLoopbackURL("localhost"),
		"ipv4_127_0_0_1":       IsLoopbackURL("http://127.0.0.1/x"),
		"ipv4_127_range":       IsLoopbackURL("http://127.1.2.3/x"),
		"ipv6_loopback":        IsLoopbackURL("http://[::1]/x"),
		"ipv6_mapped_v4":       IsLoopbackURL("http://[::ffff:127.0.0.1]/x"),
		"private_not_loopback": IsLoopbackURL("http://10.0.0.1/x"),
		"public":               IsLoopbackURL("http://example.com/x"),
		"empty":                IsLoopbackURL(""),
		"garbage":              IsLoopbackURL("not a url at all"),
		"none":                 IsLoopbackURL(""),
	})

	// decompress_capped — the same inputs, the DEcompressed properties
	dec := func(data []byte, truncated bool, err error) map[string]any {
		if err != nil {
			return map[string]any{"error_class": pyErrClass(err)}
		}
		head := string(data)
		if len(head) > 16 {
			head = head[:16]
		}
		return map[string]any{"data_len": len(data), "truncated": truncated, "head": head}
	}
	gzOK := gzCompress(t, []byte(strings.Repeat("hello compressed world ", 4)))
	gzBomb := gzCompress(t, []byte(strings.Repeat("bomb ", 100_000)))
	zlOK := zlCompress(t, []byte(strings.Repeat("deflated content here ", 4)))
	rawPayload := rawDeflateBytes(t, []byte(strings.Repeat("raw deflate payload ", 4)))
	rawBomb := rawDeflateBytes(t, []byte(strings.Repeat("raw ", 100_000)))
	check("decompress_capped", map[string]any{
		"gzip_ok":              call3(gzOK, "gzip", 10_000, dec),
		"gzip_x_gzip":          call3(gzOK, "x-gzip", 10_000, dec),
		"gzip_bomb_capped":     call3(gzBomb, "gzip", 50_000, dec),
		"gzip_bomb_head":       call3(gzBomb, "gzip", 50_000, decHead5),
		"deflate_zlib":         call3(zlOK, "deflate", 10_000, dec),
		"deflate_raw_fallback": call3(rawPayload, "deflate", 10_000, dec),
		"deflate_raw_capped":   call3(rawBomb, "deflate", 50_000, dec),
		"unsupported_identity": call3([]byte("x"), "identity", 10, dec),
		"unsupported_brotli":   call3([]byte("x"), "br", 10, dec),
		"junk_gzip":            call3([]byte("not gzip at all"), "gzip", 10_000, dec),
		"empty_gzip":           call3(gzCompress(t, nil), "gzip", 10_000, dec),
	})

	check("looks_like_pdf", map[string]any{
		"content_type_pdf":                    LooksLikePDF("application/pdf", "http://x/y", nil),
		"octet_stream_pdf_url":                LooksLikePDF("application/octet-stream", "http://x/doc.pdf", nil),
		"octet_stream_pdf_url_trailing_slash": LooksLikePDF("application/octet-stream", "http://x/doc.pdf/", nil),
		"magic_bytes":                         LooksLikePDF("text/plain", "http://x/y", []byte("%PDF-1.7 rest")),
		"plain_no":                            LooksLikePDF("text/html", "http://x/y", []byte("<html>")),
		"octet_stream_non_pdf_url":            LooksLikePDF("application/octet-stream", "http://x/doc.bin", nil),
	})

	check("decode_best_effort", map[string]any{
		"utf8":                        DecodeBestEffort([]byte("héllo wörld"), ""),
		"charset_param":               DecodeBestEffort([]byte{0x44, 0x69, 0x73, 0x65, 0xF1, 0x6F}, "windows-1252"),
		"bad_charset_falls_back_utf8": DecodeBestEffort([]byte("plain ascii"), "not-a-charset"),
		"invalid_utf8_replaced":       DecodeBestEffort([]byte{0x63, 0x61, 0x66, 0xE9}, "utf-8"),
	})

	check("rate_limit_penalty", map[string]any{
		"ra_120":          RateLimitPenalty(res("HTTP 429", intPtr(429), "", floatPtr(120))),
		"ra_none_default": RateLimitPenalty(res("HTTP 429", intPtr(429), "", nil)),
		"ra_99999_capped": RateLimitPenalty(res("HTTP 429", intPtr(429), "", floatPtr(99999))),
		"ra_503":          RateLimitPenalty(res("HTTP 503", intPtr(503), "", floatPtr(30))),
		"not_rate_403":    RateLimitPenalty(res("HTTP 403", intPtr(403), "", nil)),
		"timeout":         RateLimitPenalty(res("timeout", nil, "", nil)),
		"none_result":     RateLimitPenalty(nil),
		"ok_result":       RateLimitPenalty(okRes("u", []byte("x"))),
	})

	check("combined_category", map[string]any{
		"dead_beats_blocked_bot":      CombinedCategory(res("HTTP 403", intPtr(403), "blocked_bot", nil), res("HTTP 404", intPtr(404), "dead", nil)),
		"paywalled_beats_timeout":     CombinedCategory(res("HTTP 402", intPtr(402), "paywalled", nil), res("timeout", nil, "retry_later", nil)),
		"redirect_broken_beats_retry": CombinedCategory(res("HTTP 307 — broken redirect", intPtr(307), "redirect_broken", nil), res("timeout", nil, "retry_later", nil)),
		"proxy_error_beats_nothing":   CombinedCategory(res("proxy down", nil, "proxy_error", nil), res("x", nil, "", nil)),
		"no_categories":               CombinedCategory(res("x", nil, "", nil), nil),
		"single":                      CombinedCategory(res("HTTP 429", intPtr(429), "retry_later", nil)),
	})

	// Python returns None where the port returns "" — the driver maps
	// the empty string to nil for the golden comparison.
	pu := func(s string) any {
		if s == "" {
			return nil
		}
		return s
	}
	check("proxy_url_for_curl", map[string]any{
		"socks5_becomes_socks5h": pu(ProxyURLForImpersonator(&Proxy{"socks5", "p.example.net", 10808, "", ""})),
		"socks4":                 pu(ProxyURLForImpersonator(&Proxy{"socks4", "h", 1, "", ""})),
		"http":                   pu(ProxyURLForImpersonator(&Proxy{"http", "h", 8080, "", ""})),
		"http_auth":              pu(ProxyURLForImpersonator(&Proxy{"http", "h", 8080, "u", "p"})),
		"auth_percent_encoded":   pu(ProxyURLForImpersonator(&Proxy{"socks5", "h", 1, "a b/c", "p@ss:word"})),
		"none":                   pu(ProxyURLForImpersonator(nil)),
		"bogus_type":             pu(ProxyURLForImpersonator(&Proxy{"bogus", "h", 1, "", ""})),
		"empty_type":             pu(ProxyURLForImpersonator(&Proxy{"", "h", 1, "", ""})),
	})

	check("is_tls_handshake_failure", map[string]any{
		"owner_sslv3_line":          IsTLSHandshakeFailure("proxy: connection: [SSL: SSLV3_ALERT_HANDSHAKE_FAILURE] sslv3 alert handshake failure (_ssl.c:1010) | direct: connection: [Errno 11001] getaddrinfo failed"),
		"alert_handshake_failure":   IsTLSHandshakeFailure("connection: alert handshake failure"),
		"handshake_timed_out":       IsTLSHandshakeFailure("connection: handshake operation timed out"),
		"unexpected_eof":            IsTLSHandshakeFailure("connection: [SSL: UNEXPECTED_EOF_WHILE_READING] EOF occurred in violation of protocol"),
		"wrong_version_number":      IsTLSHandshakeFailure("connection: [SSL: WRONG_VERSION_NUMBER] tls server"),
		"tlsv1_alert":               IsTLSHandshakeFailure("connection: [SSL: TLSV1_ALERT_PROTOCOL_VERSION]"),
		"case_insensitive":          IsTLSHandshakeFailure("connection: sslv3 alert something"),
		"cert_verify_failed_is_not": IsTLSHandshakeFailure("connection: [SSL: CERTIFICATE_VERIFY_FAILED] certificate verify failed"),
		"plain_timeout_is_not":      IsTLSHandshakeFailure("timeout"),
		"empty":                     IsTLSHandshakeFailure(""),
		"none":                      IsTLSHandshakeFailure(""),
	})

	check("is_dns_failure", map[string]any{
		"getaddrinfo":          IsDNSFailure("connection: [Errno 11001] getaddrinfo failed"),
		"no_such_host":         IsDNSFailure("connection: lookup gone.example.net: no such host"),
		"name_or_service":      IsDNSFailure("connection: name or service not known"),
		"temporary_failure":    IsDNSFailure("connection: temporary failure in name resolution"),
		"nodename":             IsDNSFailure("connection: nodename nor servname provided"),
		"wsahost":              IsDNSFailure("connection: WSAHOST_NOT_FOUND"),
		"plain_timeout_is_not": IsDNSFailure("timeout"),
		"refused_is_not":       IsDNSFailure("connection: refused"),
		"empty":                IsDNSFailure(""),
	})

	check("domain_of", map[string]any{
		"plain":   DomainOf("https://Example.COM/path?q=1"),
		"port":    DomainOf("http://127.0.0.1:9944/x"),
		"garbage": DomainOf("not a url"),
		"empty":   DomainOf(""),
	})
}

// m — a shorthand for building config maps.
func m(pairs ...any) map[string]any {
	out := map[string]any{}
	for i := 0; i+1 < len(pairs); i += 2 {
		out[pairs[i].(string)] = pairs[i+1]
	}
	return out
}

func verdictTriple(reason, category string, ra *float64) []any {
	return []any{reason, category, ra}
}

// call3 — DecompressCapped through the dec adapter.
func call3(body []byte, encoding string, maxBytes int,
	adapt func([]byte, bool, error) map[string]any) map[string]any {
	data, truncated, err := DecompressCapped(body, encoding, maxBytes)
	return adapt(data, truncated, err)
}

// decHead5 — the generator's gzip_bomb_head special case: the first
// five decompressed bytes as an (untruncated) head.
func decHead5(data []byte, _ bool, err error) map[string]any {
	if err != nil {
		return map[string]any{"error_class": pyErrClass(err)}
	}
	head := string(data)
	if len(head) > 5 {
		head = head[:5]
	}
	return map[string]any{"data_len": len(head), "truncated": false, "head": head}
}

// ---------------------------------------------------------------------------
// THE PAYLOAD (WIRE) GOLDEN
// ---------------------------------------------------------------------------

func TestGoldenPayloads(t *testing.T) {
	data, err := loadGoldenFile("webfetch_payloads.golden.json")
	if err != nil {
		t.Fatal(err)
	}
	scenarios := data["scenarios"].(map[string]any)

	// compareScenario: the Go side's requests (Go's net/http adds no
	// Host/Connection to r.Header) against the golden's pairs with the
	// Python runtime's own Host/Connection filtered out.
	compareScenario := func(tt *testing.T, name string, got map[string]any) {
		tt.Helper()
		want := scenarios[name].(map[string]any)
		gotReqs, _ := got["requests"].([]map[string]any)
		wantReqs, _ := want["requests"].([]any)
		if len(gotReqs) != len(wantReqs) {
			t.Fatalf("%s: %d requests, want %d", name, len(gotReqs), len(wantReqs))
		}
		for i := range gotReqs {
			g := gotReqs[i]
			w := wantReqs[i].(map[string]any)
			if g["method"] != w["method"] || g["path"] != w["path"] {
				tt.Fatalf("%s req %d: %v %v vs %v %v", name, i,
					g["method"], g["path"], w["method"], w["path"])
			}
			// Set-wise comparison: Go's transport writes header names
			// in sorted order (the documented divergence); the ORDER
			// pin lives in the pure golden's RequestHeaders pairs.
			compareGolden(tt, name+".headers",
				sortedHeaderPairs(g["headers"]),
				sortedHeaderPairs(filterRuntimeHeaders(w["headers"])))
		}
		for _, k := range []string{"result", "verdict", "snapshot", "timestamp"} {
			if _, ok := want[k]; ok {
				compareGolden(tt, name+"."+k, got[k], want[k])
			}
		}
	}

	htmlBody := []byte("<html><head><title>Practical Typography Guide</title></head><body>the article</body></html>")

	// 1. the plain fetch
	func(t *testing.T) {
		ws := newWireServer(t, map[string]wireRoute{
			"/article": {200, htmlBody, [][2]string{{"Content-Type", "text/html; charset=utf-8"}}},
		})
		roots := map[int]string{ws.port(): "ROOT"}
		ResetCookieJar()
		r := FetchURL(ws.root()+"/article", &Options{TimeoutS: 10})
		reqs := []map[string]any{}
		for _, rec := range ws.captured() {
			reqs = append(reqs, goldenWireRequest(rec, roots))
		}
		compareScenario(t, "plain_fetch", map[string]any{
			"requests": reqs,
			"result":   normResGolden(r, roots),
		})
	}(t)

	// 2. the redirect chain — hop cookies + the Referer law, LIVE
	func(t *testing.T) {
		ws := newWireServer(t, map[string]wireRoute{
			"/chain-a": {302, nil, [][2]string{{"Location", "/chain-b"}, {"Set-Cookie", "hopA=1; Path=/"}}},
			"/chain-b": {302, nil, [][2]string{{"Location", "/chain-c"}, {"Set-Cookie", "hopB=2; Path=/"}}},
			"/chain-c": {200, htmlBody, [][2]string{{"Content-Type", "text/html; charset=utf-8"}}},
		})
		roots := map[int]string{ws.port(): "ROOT"}
		ResetCookieJar()
		r := FetchURL(ws.root()+"/chain-a", &Options{TimeoutS: 10})
		reqs := []map[string]any{}
		for _, rec := range ws.captured() {
			reqs = append(reqs, goldenWireRequest(rec, roots))
		}
		compareScenario(t, "redirect_chain", map[string]any{
			"requests": reqs,
			"result":   normResGolden(r, roots),
		})
	}(t)

	// 3. the cross-host (different port = different netloc) redirect
	func(t *testing.T) {
		ws2 := newWireServer(t, map[string]wireRoute{
			"/x-b": {200, htmlBody, [][2]string{{"Content-Type", "text/html; charset=utf-8"}}},
		})
		ws1 := newWireServer(t, map[string]wireRoute{
			"/x-a": {302, nil, [][2]string{{"Location", ws2.root() + "/x-b"}}},
		})
		roots := map[int]string{ws1.port(): "ROOT", ws2.port(): "ROOT2"}
		ResetCookieJar()
		r := FetchURL(ws1.root()+"/x-a", &Options{TimeoutS: 10})
		reqs := []map[string]any{}
		for _, rec := range ws1.captured() {
			reqs = append(reqs, goldenWireRequest(rec, roots))
		}
		for _, rec := range ws2.captured() {
			reqs = append(reqs, goldenWireRequest(rec, roots))
		}
		compareScenario(t, "cross_host_redirect", map[string]any{
			"requests": reqs,
			"result":   normResGolden(r, roots),
		})
	}(t)

	// 4. the custom UA — the Chrome client hints are dropped
	func(t *testing.T) {
		ws := newWireServer(t, map[string]wireRoute{
			"/ua": {200, htmlBody, [][2]string{{"Content-Type", "text/html; charset=utf-8"}}},
		})
		roots := map[int]string{ws.port(): "ROOT"}
		ResetCookieJar()
		FetchURL(ws.root()+"/ua", &Options{TimeoutS: 10, UserAgent: "GitCurator-test/0.41"})
		reqs := []map[string]any{}
		for _, rec := range ws.captured() {
			reqs = append(reqs, goldenWireRequest(rec, roots))
		}
		compareScenario(t, "custom_ua_drops_hints", map[string]any{"requests": reqs})
	}(t)

	// 5. the DoH probe — patched endpoint, the exact request
	func(t *testing.T) {
		state := saveDoorState()
		defer state.restore()
		ws := newWireServer(t, map[string]wireRoute{
			"/dns-query": {200, jsonBytes(map[string]any{"Status": 0, "Answer": []any{}}),
				[][2]string{{"Content-Type", "application/dns-json"}}},
		})
		roots := map[int]string{ws.port(): "DOH"}
		realEP := DoHEndpoint
		DoHEndpoint = ws.root() + "/dns-query"
		ResetDoHCache()
		verdict := dohProbeOnceVar("gone.example.net", nil)
		DoHEndpoint = realEP
		reqs := []map[string]any{}
		for _, rec := range ws.captured() {
			reqs = append(reqs, goldenWireRequest(rec, roots))
		}
		compareScenario(t, "doh_probe_request", map[string]any{
			"requests": reqs,
			"verdict":  verdict,
		})
	}(t)

	// 6. the Wayback availability probe — patched API
	func(t *testing.T) {
		state := saveDoorState()
		defer state.restore()
		ws := newWireServer(t, map[string]wireRoute{
			"/wayback/available": {200, jsonBytes(map[string]any{
				"archived_snapshots": map[string]any{
					"closest": map[string]any{
						"available": true,
						"url":       "https://web.archive.org/web/20150601/https://past.example.org/a",
						"timestamp": "20150601"}}}),
				[][2]string{{"Content-Type", "application/json"}}},
		})
		roots := map[int]string{ws.port(): "ARCHIVE"}
		realAPI := ArchiveAPI
		ArchiveAPI = ws.root() + "/wayback/available"
		snap, ts := archiveLookupVar("https://past.example.org/a", nil, 15.0, time.Now())
		ArchiveAPI = realAPI
		reqs := []map[string]any{}
		for _, rec := range ws.captured() {
			reqs = append(reqs, goldenWireRequest(rec, roots))
		}
		compareScenario(t, "wayback_availability_request", map[string]any{
			"requests":  reqs,
			"snapshot":  snap,
			"timestamp": ts,
		})
	}(t)

	// 7. the http-type proxy — the absolute URI on the request line
	func(t *testing.T) {
		ws := newWireServer(t, map[string]wireRoute{
			"/any": {200, htmlBody, [][2]string{{"Content-Type", "text/html"}}},
		})
		roots := map[int]string{ws.port(): "PROXY"}
		ResetCookieJar()
		r := FetchURL("http://example.test/via-proxy", &Options{TimeoutS: 8,
			Proxy: &Proxy{Type: "http", Host: "127.0.0.1", Port: ws.port()}})
		reqs := []map[string]any{}
		for _, rec := range ws.captured() {
			reqs = append(reqs, goldenWireRequest(rec, roots))
		}
		compareScenario(t, "http_proxy_absolute_uri", map[string]any{
			"requests": reqs,
			"result":   normResGolden(r, roots),
		})
	}(t)
}

// ---------------------------------------------------------------------------
// THE BEHAVIOR GOLDEN
// ---------------------------------------------------------------------------

func TestGoldenBehavior(t *testing.T) {
	data, err := loadGoldenFile("webfetch_behavior.golden.json")
	if err != nil {
		t.Fatal(err)
	}
	probes := data["probes"].(map[string]any)
	probe := func(name string, fn func(t *testing.T) any) {
		t.Run(name, func(t *testing.T) {
			compareGolden(t, "behavior."+name, fn(t), probes[name])
		})
	}

	// --- the redirect truth -------------------------------------------
	probe("redirect_truth", func(t *testing.T) any {
		srv := newLadderServer(t)
		roots := map[int]string{srv.port(): "ROOT"}
		fetch := func(path string) *FetchResult {
			ResetCookieJar()
			return FetchURL(srv.root()+path, &Options{TimeoutS: 10})
		}
		return map[string]any{
			"hop_cookies_ride":                   normResGolden(fetch("/hop-cookie"), roots),
			"no_location_broken":                 normResGolden(fetch("/no-location"), roots),
			"loop_named":                         normResGolden(fetch("/loop"), roots),
			"hop_budget_holds":                   normResGolden(fetch("/chain/1"), roots),
			"moved_308_followed":                 normResGolden(fetch("/moved-308"), roots),
			"plain_fetch_full":                   normResGolden(fetch("/article"), roots),
			"cookie_target_gated_without_cookie": normResGolden(fetch("/hop-cookie-target"), roots),
		}
	})

	// --- the phase-2 battery --------------------------------------------
	p2 := func(t *testing.T) (*ladderServer, map[int]string, func(path string, opts *Options) *FetchResult) {
		srv := newPhase2Server(t)
		roots := map[int]string{srv.port(): "P2"}
		fetch := func(path string, opts *Options) *FetchResult {
			ResetCookieJar()
			if opts == nil {
				opts = &Options{TimeoutS: 10}
			}
			return FetchURL(srv.root()+path, opts)
		}
		return srv, roots, fetch
	}
	probe("phase2_full_fetch", func(t *testing.T) any {
		_, roots, fetch := p2(t)
		return normResGolden(fetch("/article", nil), roots)
	})
	probe("phase2_follows_redirect", func(t *testing.T) any {
		_, roots, fetch := p2(t)
		return normResGolden(fetch("/redirect", nil), roots)
	})
	probe("phase2_404_failed", func(t *testing.T) any {
		_, roots, fetch := p2(t)
		return normResGolden(fetch("/nope", nil), roots)
	})
	probe("phase2_timeout", func(t *testing.T) any {
		_, roots, fetch := p2(t)
		t0 := time.Now()
		r := fetch("/slow", &Options{TimeoutS: 1})
		d := normResGolden(r, roots)
		d["elapsed_lt_5"] = time.Since(t0) < 5*time.Second
		return d
	})
	probe("phase2_size_cap_partial", func(t *testing.T) any {
		_, roots, fetch := p2(t)
		return normResGolden(fetch("/huge", &Options{TimeoutS: 20, MaxBytes: 50_000}), roots)
	})
	probe("phase2_pdf_partial", func(t *testing.T) any {
		_, roots, fetch := p2(t)
		return normResGolden(fetch("/pdf", nil), roots)
	})
	probe("phase2_latin1_decodes", func(t *testing.T) any {
		_, roots, fetch := p2(t)
		return normResGolden(fetch("/latin1", nil), roots)
	})
	probe("phase2_redirect_loop_failed", func(t *testing.T) any {
		_, roots, fetch := p2(t)
		return normResGolden(fetch("/redirect-loop", nil), roots)
	})
	probe("phase2_bad_scheme", func(t *testing.T) any {
		return normResGolden(FetchURL("ftp://example.com/x", &Options{TimeoutS: 5}), nil)
	})
	probe("phase2_gzip_decoded", func(t *testing.T) any {
		_, roots, fetch := p2(t)
		return normResGolden(fetch("/gzip", nil), roots)
	})
	probe("phase2_deflate_decoded", func(t *testing.T) any {
		_, roots, fetch := p2(t)
		return normResGolden(fetch("/deflate", nil), roots)
	})
	probe("phase2_gzip_bomb_capped", func(t *testing.T) any {
		_, roots, fetch := p2(t)
		return normResGolden(fetch("/gzip-bomb", &Options{MaxBytes: 50_000}), roots)
	})
	probe("phase2_403_bot_defense", func(t *testing.T) any {
		_, roots, fetch := p2(t)
		return normResGolden(fetch("/blocked", nil), roots)
	})
	probe("phase2_rate_limiter_delay", func(t *testing.T) any {
		lim := NewDomainRateLimiter(0.4)
		lim.Wait("example.com")
		t0 := time.Now()
		lim.Wait("example.com")
		secondWait := time.Since(t0)
		t1 := time.Now()
		lim.Wait("other.com")
		otherWait := time.Since(t1)
		return map[string]any{
			"second_hit_waited": secondWait >= 350*time.Millisecond,
			"other_domain_free": otherWait < 200*time.Millisecond,
		}
	})

	// --- the DoH verdicts -------------------------------------------------
	probe("dns_verdict", func(t *testing.T) any {
		state := saveDoorState()
		defer state.restore()
		out := map[string]any{}
		arm := func(verdict string) *[][2]any {
			calls := &[][2]any{}
			dohProbeOnceVar = func(host string, proxy *Proxy) string {
				*calls = append(*calls, [2]any{host, proxy})
				return verdict
			}
			return calls
		}
		fetch := func(url string, opts *Options) *FetchResult {
			return FetchURL(url, opts)
		}
		calls := arm("nxdomain")
		fetchOnceVar = dnsFail
		out["nxdomain"] = map[string]any{
			"result":      normResGolden(fetch("http://gone.example.net/x", &Options{TimeoutS: 2}), nil),
			"probe_calls": *calls,
		}
		calls = arm("exists")
		out["exists"] = map[string]any{
			"result":      normResGolden(fetch("http://x.example.net/x", &Options{TimeoutS: 2}), nil),
			"probe_calls": *calls,
		}
		calls = arm("unknown")
		out["unknown"] = map[string]any{
			"result":      normResGolden(fetch("http://y.example.net/x", &Options{TimeoutS: 2}), nil),
			"probe_calls": *calls,
		}
		calls = arm("nxdomain")
		fetchOnceVar = dnsFail
		fetch("http://z.example.net/x", &Options{TimeoutS: 2})
		fetch("http://z.example.net/x", &Options{TimeoutS: 2})
		out["cached_per_host"] = map[string]any{"probe_call_count": len(*calls)}
		calls = arm("nxdomain")
		out["opt_out"] = map[string]any{
			"result":      normResGolden(fetch("http://w.example.net/x", &Options{TimeoutS: 2, NoDoHProbe: true}), nil),
			"probe_calls": *calls,
		}
		calls = arm("nxdomain")
		fetchOnceVar = func(url string, timeoutS float64, maxBytes int, ua string, p *Proxy, s time.Time) *FetchResult {
			return res("timeout", nil, "retry_later", nil)
		}
		out["non_dns_never_probes"] = map[string]any{
			"result":      normResGolden(fetch("http://v.example.net/x", &Options{TimeoutS: 2}), nil),
			"probe_calls": *calls,
		}
		return out
	})

	// --- the dead-page rescue ladder --------------------------------------
	probe("dead_page_rescue", func(t *testing.T) any {
		state := saveDoorState()
		defer state.restore()
		armAvailable(false)
		out := map[string]any{}
		dead := func(url string, timeoutS float64, maxBytes int, ua string, p *Proxy, s time.Time) *FetchResult {
			return res("HTTP 404 — the page is gone", intPtr(404), "dead", nil)
		}
		pages := map[string]*FetchResult{
			"https://moved.example.org/page":     res("HTTP 404 — the page is gone", intPtr(404), "dead", nil),
			"https://moved.example.org/page/":    res("HTTP 404 — the page is gone", intPtr(404), "dead", nil),
			"https://www.moved.example.org/page": okRes("https://www.moved.example.org/page", []byte("<html><body>the page</body></html>")),
		}
		archive := armArchive(t, "no-snapshot", "")
		fetchOnceVar = func(url string, timeoutS float64, maxBytes int, ua string, p *Proxy, s time.Time) *FetchResult {
			return pages[url]
		}
		out["www_variant_rescues"] = map[string]any{
			"result":        normResGolden(FetchURL("https://moved.example.org/page", &Options{TimeoutS: 2}), nil),
			"archive_calls": *archive,
		}
		state.restore()
		armAvailable(false)
		archive = armArchive(t, "no-snapshot", "")
		fetchOnceVar = dead
		out["nothing_rescued_dead"] = map[string]any{
			"result":        normResGolden(FetchURL("https://gone.example.org/dead", &Options{TimeoutS: 2}), nil),
			"archive_calls": *archive,
		}
		state.restore()
		armAvailable(false)
		archive = armArchive(t, "https://web.archive.org/web/20150601/https://past.example.org/a", "20150601")
		fetchOnceVar = func(url string, timeoutS float64, maxBytes int, ua string, p *Proxy, s time.Time) *FetchResult {
			if strings.Contains(url, "web.archive.org") {
				return okRes(url, []byte("<html><head><title>The Past</title></head></html>"))
			}
			return res("HTTP 410 — the page was removed", intPtr(410), "dead", nil)
		}
		out["archived_snapshot_partial"] = map[string]any{
			"result":        normResGolden(FetchURL("https://past.example.org/a", &Options{TimeoutS: 2}), nil),
			"archive_calls": *archive,
		}
		state.restore()
		armAvailable(false)
		archive = armArchive(t, "no-snapshot", "")
		fetchOnceVar = dead
		out["archive_opt_out"] = map[string]any{
			"result":        normResGolden(FetchURL("https://opt.example.org/x", &Options{TimeoutS: 2, NoArchive: true}), nil),
			"archive_calls": *archive,
		}
		state.restore()
		armAvailable(false)
		fetchOnceVar = dead
		out["loopback_never"] = normResGolden(FetchURL("http://127.0.0.1:9/x", &Options{TimeoutS: 2}), nil)
		return out
	})

	// --- the rate penalties -------------------------------------------------
	probe("rate_penalty", func(t *testing.T) any {
		state := saveDoorState()
		defer state.restore()
		armAvailable(false)
		lim := &fakeLimiter{}
		fetchOnceVar = func(url string, timeoutS float64, maxBytes int, ua string, p *Proxy, s time.Time) *FetchResult {
			return res("HTTP 429 — rate limited (Retry-After: 120s — the site asked for a pause)", intPtr(429), "retry_later", floatPtr(120))
		}
		first := map[string]any{
			"result":    normResGolden(FetchURL("http://limited.example.net/x", &Options{TimeoutS: 2, Proxy: testProxy, RateLimiter: lim}), nil),
			"penalties": goldenPenalties(lim),
		}
		lim = &fakeLimiter{}
		fetchOnceVar = func(url string, timeoutS float64, maxBytes int, ua string, p *Proxy, s time.Time) *FetchResult {
			return res("HTTP 429 — rate limited", intPtr(429), "retry_later", nil)
		}
		second := map[string]any{
			"result":    normResGolden(FetchURL("http://bare.example.net/x", &Options{TimeoutS: 2, RateLimiter: lim}), nil),
			"penalties": goldenPenalties(lim),
		}
		return map[string]any{
			"both_routes_pay_retry_after":  first,
			"bare_429_pays_capped_default": second,
		}
	})

	// --- the origin-down Wayback courtesy -------------------------------------
	probe("origin_down_courtesy", func(t *testing.T) any {
		state := saveDoorState()
		defer state.restore()
		out := map[string]any{}
		verdict52xGolden := func(legs ...*FetchResult) *FetchResult {
			result := &FetchResult{
				URL:    "https://origin.example.net/x",
				Status: "failed",
				Reason: "proxy: HTTP 403 — forbidden | direct: HTTP 523 — " +
					"Cloudflare: the site's own server origin is " +
					"unreachable (site-side, retry later — not the " +
					"fetcher's fault)",
				Category: "blocked_bot",
			}
			return finalVerdict(result, "", "https://origin.example.net/x", nil, nil,
				"origin.example.net", false, legs, true, time.Now())
		}
		snap := "https://web.archive.org/web/20230101120000/https://origin.example.net/x"
		archive := armArchive(t, snap, "20230101120000")
		r := verdict52xGolden(res("HTTP 403", intPtr(403), "", nil), res("HTTP 523", intPtr(523), "", nil))
		out["52x_names_snapshot"] = map[string]any{
			"reason": r.Reason, "category": r.Category, "archive_calls": *archive,
		}
		archive = armArchive(t, "no-snapshot", "")
		r = verdict52xGolden(res("HTTP 403", intPtr(403), "", nil), res("HTTP 523", intPtr(523), "", nil))
		out["no_snapshot_no_suffix"] = map[string]any{"reason": r.Reason, "archive_calls": *archive}
		archive = armArchive(t, "unreachable", "")
		r = verdict52xGolden(res("HTTP 403", intPtr(403), "", nil), res("HTTP 523", intPtr(523), "", nil))
		out["unreachable_no_suffix"] = map[string]any{"reason": r.Reason, "archive_calls": *archive}
		archive = armArchive(t, "no-snapshot", "")
		result := &FetchResult{URL: "https://origin.example.net/x", Status: "failed", Reason: "HTTP 523", Category: "retry_later"}
		finalVerdict(result, "", "https://origin.example.net/x", nil, nil,
			"origin.example.net", false, []*FetchResult{res("HTTP 523", intPtr(523), "", nil)}, false, time.Now())
		out["optout_never_probes"] = map[string]any{"archive_calls": *archive}
		archive = armArchive(t, "no-snapshot", "")
		r = verdict52xGolden(res("HTTP 403", intPtr(403), "", nil), res("HTTP 429", intPtr(429), "", nil))
		out["no_52x_never_probes"] = map[string]any{"reason": r.Reason, "archive_calls": *archive}
		archive = armArchive(t, "no-snapshot", "")
		url := "http://127.0.0.1:9/x"
		result = &FetchResult{URL: url, Status: "failed", Reason: "HTTP 523", Category: "retry_later"}
		finalVerdict(result, "", url, nil, nil, "127.0.0.1:9", false,
			[]*FetchResult{resAt("HTTP 523", intPtr(523), url)}, true, time.Now())
		out["loopback_never_probes"] = map[string]any{"archive_calls": *archive}
		archive = armArchive(t, "https://web.archive.org/web/2023/https://solo.example.org", "")
		url = "https://solo.example.org/y"
		result = &FetchResult{URL: url, Status: "failed", Reason: "HTTP 523 — Cloudflare", Category: "retry_later"}
		r = finalVerdict(result, "", url, nil, nil, "solo.example.org", false,
			[]*FetchResult{resAt("HTTP 523", intPtr(523), url)}, true, time.Now())
		out["single_52x_leg"] = map[string]any{"reason": r.Reason, "archive_calls": *archive}
		return out
	})

	// --- the third door's trigger matrix ----------------------------------------
	probe("third_door_trigger", func(t *testing.T) any {
		state := saveDoorState()
		defer state.restore()
		out := map[string]any{}
		wallURL := "http://walled.example.net/page"

		sessions := armFakeSession(&fakeImpResponse{StatusCode: 200})
		fetchOnceVar = func(url string, timeoutS float64, maxBytes int, ua string, p *Proxy, s time.Time) *FetchResult {
			if p != nil {
				return res("connection: [SSL: SSLV3_ALERT_HANDSHAKE_FAILURE] sslv3 alert handshake failure (_ssl.c:1010)", nil, "", nil)
			}
			return res("connection: [Errno 11001] getaddrinfo failed", nil, "", nil)
		}
		armDoH(t, "unknown")
		out["owners_tls_line_opens_door"] = map[string]any{
			"result":    normResGolden(FetchURL(wallURL, &Options{TimeoutS: 2, Proxy: testProxy}), nil),
			"imp_calls": impCallsGolden(*sessions),
		}
		state.restore()
		sessions = armFakeSession(&fakeImpResponse{StatusCode: 200})
		fetchOnceVar = func(url string, timeoutS float64, maxBytes int, ua string, p *Proxy, s time.Time) *FetchResult {
			return res("timeout", nil, "retry_later", nil)
		}
		out["plain_timeout_never_asks_chrome"] = map[string]any{
			"result":    normResGolden(FetchURL(wallURL, &Options{TimeoutS: 2, Proxy: testProxy}), nil),
			"imp_calls": impCallsGolden(*sessions),
		}
		state.restore()
		sessions = armFakeSession(&fakeImpResponse{StatusCode: 200})
		fetchOnceVar = func(url string, timeoutS float64, maxBytes int, ua string, p *Proxy, s time.Time) *FetchResult {
			return res("connection: [SSL: UNEXPECTED_EOF_WHILE_READING] EOF occurred in violation of protocol (_ssl.c:1010)", nil, "", nil)
		}
		out["single_door_tls_opens_direct"] = map[string]any{
			"result":    normResGolden(FetchURL(wallURL, &Options{TimeoutS: 2}), nil),
			"imp_calls": impCallsGolden(*sessions),
		}
		state.restore()
		sessions = armFakeSession(&fakeImpResponse{StatusCode: 200})
		fetchOnceVar = func(url string, timeoutS float64, maxBytes int, ua string, p *Proxy, s time.Time) *FetchResult {
			return res("connection: [SSL: CERTIFICATE_VERIFY_FAILED] certificate verify failed", nil, "", nil)
		}
		out["cert_failure_never_opens"] = map[string]any{
			"result":    normResGolden(FetchURL(wallURL, &Options{TimeoutS: 2}), nil),
			"imp_calls": impCallsGolden(*sessions),
		}
		refusal := map[string]any{}
		for _, code := range []int{403, 405, 429, 451} {
			state.restore()
			sessions = armFakeSession(&fakeImpResponse{StatusCode: 200})
			fetchOnceVar = func(url string, timeoutS float64, maxBytes int, ua string, p *Proxy, s time.Time) *FetchResult {
				return res("HTTP "+itoa(code), &code, "", nil)
			}
			r := FetchURL(wallURL, &Options{TimeoutS: 2})
			refusal[itoa(code)] = map[string]any{
				"ok":                      r.OK(),
				"reason_has_impersonated": strings.Contains(r.Reason, "via impersonated Chrome"),
			}
		}
		out["refusal_family_earns_door"] = refusal
		state.restore()
		sessions = armFakeSession(&fakeImpResponse{StatusCode: 200})
		fetchOnceVar = func(url string, timeoutS float64, maxBytes int, ua string, p *Proxy, s time.Time) *FetchResult {
			return res("HTTP 403", intPtr(403), "", nil)
		}
		out["opt_out_honored"] = map[string]any{
			"result":    normResGolden(FetchURL(wallURL, &Options{TimeoutS: 2, NoImpersonate: true}), nil),
			"imp_calls": impCallsGolden(*sessions),
		}
		state.restore()
		sessions = armFakeSession(&fakeImpResponse{StatusCode: 200})
		r := MaybeImpersonate("http://127.0.0.1:8080/x", 5, 2_000_000, []*Proxy{nil}, true,
			res(ownerLine, intPtr(403), "", nil), time.Now(), false)
		out["loopback_never_escalates"] = map[string]any{
			"result": normResGolden(r, nil), "imp_calls": impCallsGolden(*sessions),
		}
		state.restore()
		sessions = armFakeSession(&fakeImpResponse{StatusCode: 200})
		armAvailable(false)
		r = MaybeImpersonate(wallURL, 5, 2_000_000, []*Proxy{nil}, true,
			res(ownerLine, intPtr(403), "", nil), time.Now(), false)
		out["missing_library_hint"] = map[string]any{
			"result": normResGolden(r, nil), "imp_calls": impCallsGolden(*sessions),
		}
		return out
	})

	// --- the third door's verdicts ------------------------------------------------
	probe("third_door_verdicts", func(t *testing.T) any {
		state := saveDoorState()
		defer state.restore()
		out := map[string]any{}
		wallURL := "http://walled.example.net/page"
		ask := func(walled *FetchResult, routes []*Proxy) *FetchResult {
			return MaybeImpersonate(wallURL, 5, 2_000_000, routes, true, walled, time.Now(), false)
		}
		sessions := armFakeSession(&fakeImpResponse{StatusCode: 200,
			Body: []byte("<html><head><title>The Real Page</title></head><body>fix live</body></html>")})
		out["success_is_real_success"] = map[string]any{
			"result":    normResGolden(ask(res(ownerLine, intPtr(403), "", nil), []*Proxy{nil}), nil),
			"imp_calls": impCallsGolden(*sessions),
		}
		state.restore()
		challenged := &fakeImpResponse{StatusCode: 403,
			Header: hdr("Server", "cloudflare", "cf-mitigated", "challenge")}
		sessions = armFakeSession(challenged)
		out["challenge_survives_three_leg"] = map[string]any{
			"result":    normResGolden(ask(res(ownerLine, intPtr(403), "", nil), []*Proxy{nil}), nil),
			"imp_calls": impCallsGolden(*sessions),
		}
		state.restore()
		sessions = armFakeSession(challenged, challenged)
		out["two_routes_two_asks_proxy_first"] = map[string]any{
			"result":    normResGolden(ask(res(ownerLine, intPtr(403), "", nil), []*Proxy{testProxy, nil}), nil),
			"imp_calls": impCallsGolden(*sessions),
		}
		state.restore()
		sessions = armFakeSession(&fakeImpResponse{StatusCode: 403})
		out["plain_403_reports_cleanly"] = map[string]any{
			"result":    normResGolden(ask(res("proxy: HTTP 403 | direct: HTTP 403", intPtr(403), "", nil), []*Proxy{nil}), nil),
			"imp_calls": impCallsGolden(*sessions),
		}
		return out
	})

	// --- the autopip laws ------------------------------------------------------------
	probe("autopip_laws", func(t *testing.T) any {
		state := saveDoorState()
		defer state.restore()
		out := map[string]any{}
		wallURL := "http://walled.example.net/page"
		plain403 := func(url string, timeoutS float64, maxBytes int, ua string, p *Proxy, s time.Time) *FetchResult {
			return res("HTTP 403", intPtr(403), "", nil)
		}
		// installs_and_answers
		pipCalls := 0
		armAvailable(false)
		impMu.Lock()
		autopipTried = false
		autopipOK = false
		pipInstallImpersonator = func() bool {
			pipCalls++
			armAvailable(true)
			return true
		}
		impMu.Unlock()
		sessions := armFakeSession(&fakeImpResponse{StatusCode: 200})
		fetchOnceVar = plain403
		out["installs_and_answers"] = map[string]any{
			"result":    normResGolden(FetchURL(wallURL, &Options{TimeoutS: 2, ImpersonateAutopip: true}), nil),
			"pip_calls": pipCalls,
			"imp_calls": impCallsGolden(*sessions),
		}
		// off_keeps_hint
		state.restore()
		pipCalls = 0
		armAvailable(false)
		impMu.Lock()
		pipInstallImpersonator = func() bool {
			pipCalls++
			return true
		}
		impMu.Unlock()
		sessions = armFakeSession(&fakeImpResponse{StatusCode: 200})
		fetchOnceVar = plain403
		out["off_keeps_hint"] = map[string]any{
			"result":    normResGolden(FetchURL(wallURL, &Options{TimeoutS: 2}), nil),
			"pip_calls": pipCalls,
			"imp_calls": impCallsGolden(*sessions),
		}
		// failed_install_keeps_hint
		state.restore()
		pipCalls = 0
		armAvailable(false)
		impMu.Lock()
		autopipTried = false
		autopipOK = false
		pipInstallImpersonator = func() bool {
			pipCalls++
			return false
		}
		impMu.Unlock()
		sessions = armFakeSession(&fakeImpResponse{StatusCode: 200})
		fetchOnceVar = plain403
		out["failed_install_keeps_hint"] = map[string]any{
			"result":    normResGolden(FetchURL(wallURL, &Options{TimeoutS: 2, ImpersonateAutopip: true}), nil),
			"pip_calls": pipCalls,
		}
		// once_per_process
		state.restore()
		pipCalls = 0
		armAvailable(false)
		impMu.Lock()
		autopipTried = false
		autopipOK = false
		pipInstallImpersonator = func() bool {
			pipCalls++
			return false
		}
		impMu.Unlock()
		sessions = armFakeSession(&fakeImpResponse{StatusCode: 200})
		fetchOnceVar = plain403
		FetchURL(wallURL, &Options{TimeoutS: 2, ImpersonateAutopip: true})
		FetchURL(wallURL, &Options{TimeoutS: 2, ImpersonateAutopip: true})
		out["once_per_process"] = map[string]any{"pip_calls": pipCalls}
		// resource_truth_never_installs
		state.restore()
		pipCalls = 0
		armAvailable(false)
		impMu.Lock()
		pipInstallImpersonator = func() bool {
			pipCalls++
			return true
		}
		impMu.Unlock()
		sessions = armFakeSession(&fakeImpResponse{StatusCode: 200})
		fetchOnceVar = func(url string, timeoutS float64, maxBytes int, ua string, p *Proxy, s time.Time) *FetchResult {
			return res("HTTP 404", intPtr(404), "", nil)
		}
		armArchive(t, "no-snapshot", "")
		FetchURL(wallURL, &Options{TimeoutS: 2, ImpersonateAutopip: true})
		out["resource_truth_never_installs"] = map[string]any{"pip_calls": pipCalls}
		return out
	})

	// --- the both-doors alternation table ----------------------------------------------
	probe("both_doors", func(t *testing.T) any {
		state := saveDoorState()
		defer state.restore()
		out := map[string]any{}
		bothProxyLocal := &Proxy{Type: "socks5", Host: "127.0.0.1", Port: 1}
		routeCalls := func(calls *[]*Proxy) []any {
			out := []any{}
			for _, c := range *calls {
				out = append(out, c != nil)
			}
			return out
		}
		scripted := func(first string, firstStatus *int, secondURL string) *[]*Proxy {
			return scriptFetchOnce(func(url string, timeoutS float64, maxBytes int, ua string, p *Proxy, s time.Time) *FetchResult {
				if p != nil {
					return res(first, firstStatus, "", nil)
				}
				// a FRESH result per call (the shared-pointer trap: the
				// ladder mutates success reasons in place)
				return fullRoute(secondURL)
			})
		}
		cases := []struct {
			name        string
			first       string
			firstStatus *int
			secondURL   string
		}{
			{"403_falls_back_and_succeeds", "HTTP 403 — bot defense (server: cloudflare)", intPtr(403), "https://walled.test/a"},
			{"405_alternates", "HTTP 405", intPtr(405), "https://w.test/b"},
			{"429_alternates", "HTTP 429", intPtr(429), "https://w.test/c"},
			{"451_alternates", "HTTP 451", intPtr(451), "https://w.test/d"},
			{"timeout_still_alternates", "timeout", nil, "https://w.test/f"},
		}
		for _, tc := range cases {
			state.restore()
			armAvailable(false)
			calls := scripted(tc.first, tc.firstStatus, tc.secondURL)
			r := FetchURL("https://w.test/"+strings.SplitN(tc.name, "_", 2)[0], &Options{TimeoutS: 2, Proxy: bothProxyLocal})
			out[tc.name] = map[string]any{
				"result":      normResGolden(r, nil),
				"route_calls": routeCalls(calls),
			}
		}
		state.restore()
		armAvailable(false)
		calls := scriptFetchOnce(func(url string, timeoutS float64, maxBytes int, ua string, p *Proxy, s time.Time) *FetchResult {
			if p != nil {
				return res("HTTP 403 — bot defense (server: cloudflare)", intPtr(403), "", nil)
			}
			return res("HTTP 405", intPtr(405), "", nil)
		})
		out["walled_on_both_reports_both"] = map[string]any{
			"result":      normResGolden(FetchURL("https://w.test/e", &Options{TimeoutS: 2, Proxy: bothProxyLocal}), nil),
			"route_calls": routeCalls(calls),
		}
		state.restore()
		armAvailable(false)
		calls = scriptFetchOnce(func(url string, timeoutS float64, maxBytes int, ua string, p *Proxy, s time.Time) *FetchResult {
			if p != nil {
				return res("timeout", nil, "", nil)
			}
			return res("connection: timed out", nil, "", nil)
		})
		out["both_connection_dead_line"] = map[string]any{
			"result":      normResGolden(FetchURL("https://w.test/g", &Options{TimeoutS: 2, Proxy: bothProxyLocal}), nil),
			"route_calls": routeCalls(calls),
		}
		state.restore()
		armAvailable(false)
		calls = scriptFetchOnce(func(url string, timeoutS float64, maxBytes int, ua string, p *Proxy, s time.Time) *FetchResult {
			return fullRoute("https://w.test/h")
		})
		out["first_door_success_never_dials_second"] = map[string]any{
			"result":      normResGolden(FetchURL("https://w.test/h", &Options{TimeoutS: 2, Proxy: bothProxyLocal}), nil),
			"route_calls": routeCalls(calls),
		}
		state.restore()
		armAvailable(false)
		// (no archive stub: the real archive lookup runs and hits the
		// scripted 404 fetchOnce — the "unreachable" verdict, exactly
		// the generator's wiring)
		calls = scriptFetchOnce(func(url string, timeoutS float64, maxBytes int, ua string, p *Proxy, s time.Time) *FetchResult {
			return res("HTTP 404", intPtr(404), "", nil)
		})
		r := FetchURL("https://gone.test/404", &Options{TimeoutS: 2, Proxy: bothProxyLocal})
		allPrimary := true
		for _, c := range *calls {
			if c != bothProxyLocal {
				allPrimary = false
			}
		}
		out["404_truth_single_route_with_ladder"] = map[string]any{
			"result":      normResGolden(r, nil),
			"route_calls": routeCalls(calls),
			"all_primary": allPrimary,
		}
		for _, tc := range []struct {
			code   int
			reason string
		}{{401, "HTTP 401"}, {500, "HTTP 500"}} {
			state.restore()
			armAvailable(false)
			calls = scriptFetchOnce(func(url string, timeoutS float64, maxBytes int, ua string, p *Proxy, s time.Time) *FetchResult {
				return res(tc.reason, &tc.code, "", nil)
			})
			out[itoa(tc.code)+"_truth_single_attempt"] = map[string]any{
				"result":      normResGolden(FetchURL("https://gone.test/x", &Options{TimeoutS: 2, Proxy: bothProxyLocal}), nil),
				"route_calls": routeCalls(calls),
			}
		}
		state.restore()
		armAvailable(false)
		calls = scriptFetchOnce(func(url string, timeoutS float64, maxBytes int, ua string, p *Proxy, s time.Time) *FetchResult {
			return res("HTTP 403", intPtr(403), "", nil)
		})
		out["no_proxy_single_attempt"] = map[string]any{
			"result":      normResGolden(FetchURL("https://walled.test/x", &Options{TimeoutS: 2}), nil),
			"route_calls": routeCalls(calls),
		}
		state.restore()
		armAvailable(false)
		calls = scriptFetchOnce(func(url string, timeoutS float64, maxBytes int, ua string, p *Proxy, s time.Time) *FetchResult {
			return res("HTTP 403", intPtr(403), "", nil)
		})
		out["direct_optout_single_attempt"] = map[string]any{
			"result":      normResGolden(FetchURL("https://walled.test/y", &Options{TimeoutS: 2, Proxy: bothProxyLocal, NoDirectFallback: true}), nil),
			"route_calls": routeCalls(calls),
		}
		state.restore()
		armAvailable(false)
		calls = scriptFetchOnce(func(url string, timeoutS float64, maxBytes int, ua string, p *Proxy, s time.Time) *FetchResult {
			return res("HTTP 403", intPtr(403), "", nil)
		})
		out["loopback_never_alternates"] = map[string]any{
			"result":      normResGolden(FetchURL("http://127.0.0.1:9/z", &Options{TimeoutS: 2, Proxy: bothProxyLocal}), nil),
			"route_calls": routeCalls(calls),
		}
		out["bad_url"] = normResGolden(FetchURL("ftp://example.com/x", &Options{TimeoutS: 2}), nil)
		return out
	})

	// --- the preflight facts ------------------------------------------------------------
	probe("preflight", func(t *testing.T) any {
		port := alivePort(t)
		ok, why := WebProxyPreflight(&Proxy{Type: "http", Host: "127.0.0.1", Port: port}, 2)
		alive := []any{ok, why}
		deadPortNum := deadPort(t)
		ok, why = WebProxyPreflight(&Proxy{Type: "http", Host: "127.0.0.1", Port: deadPortNum}, 1)
		dead := map[string]any{
			"ok":                   ok,
			"why_has_host_port":    strings.Contains(why, "127.0.0.1:"+itoa(deadPortNum)),
			"why_has_unreachable":  strings.Contains(why, "unreachable"),
			"why_names_the_remedy": strings.Contains(why, "v2rayN"),
		}
		okNone, whyNone := WebProxyPreflight(nil, 2)
		none := []any{okNone, whyNone}
		return map[string]any{"alive": alive, "dead": dead, "none": none}
	})

	// --- the dead http proxy (fact shape) --------------------------------------------------
	probe("dead_http_proxy", func(t *testing.T) any {
		state := saveDoorState()
		defer state.restore()
		port := deadPort(t)
		r := FetchURL("http://example.test/x", &Options{TimeoutS: 4,
			Proxy: &Proxy{Type: "http", Host: "127.0.0.1", Port: port}})
		return map[string]any{
			"ok":               r.OK(),
			"status":           r.Status,
			"category":         r.Category,
			"reason_has_proxy": strings.Contains(strings.ToLower(r.Reason), "proxy"),
		}
	})
}

func goldenPenalties(lim *fakeLimiter) []any {
	out := []any{}
	for _, p := range lim.penalties {
		out = append(out, []any{p.domain, p.seconds})
	}
	return out
}
