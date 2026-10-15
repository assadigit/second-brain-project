# Module Spec — `app/gitcurator/core/web_fetch.py`

- **LOC**: 1,888 · **Go home**: `internal/webfetch` · **Status**: port (ADR-0006)
- **Responsibility**: the polite fetcher + machine half of the five-door ladder: proxy/direct
  routes, redirect loop with cookies, caps, failure-category taxonomy, rate limiting, DoH
  verdicts, Wayback rescue, curl_cffi impersonation third door.

## Key constants (`web_fetch.py:`)

`USER_AGENT` Chrome 131 full string (136-138, config `web_user_agent` overrides) ·
`DEFAULT_TIMEOUT_S 20` · `DEFAULT_MAX_BYTES 2_000_000` · `DEFAULT_DOMAIN_DELAY_S 2.0` ·
`MAX_REDIRECTS 5` · `_REDIRECT_STATUSES (301,302,303,307,308)` ·
`ROUTE_REFUSAL_STATUSES (403,405,429,451)` · `_RATE_PENALTY_DEFAULT_S 60` / `_CAP 300` ·
`_DOH_ENDPOINT https://cloudflare-dns.com/dns-query` (5 s) ·
`_ARCHIVE_API https://archive.org/wayback/available` · `_NO_HEAL_CATEGORIES
('dead','paywalled','refused')` · category precedence `('paywalled','dead','refused',
'redirect_broken','blocked_bot','proxy_error','retry_later')`.

## Structures

- `FetchResult` (331-372): slots `(url, final_url, status full|partial|failed, reason,
  http_status, content_type, charset, body, text, elapsed_s, category, retry_after_s)`;
  `.ok = status != 'failed'`.
- `DomainRateLimiter` (375-423): `wait(domain)` sleeps ≤1 s slices to `last_hit+delay` or
  penalty; `penalize(domain, s)`; mutex-guarded map.
- Proxy: `proxy_from_config` (433-474) reads `proxy{enabled,type,host,port,use_for_web…}`;
  `use_for_web` missing = ON; `_socks_connect` uses **`rdns=True`** (DNS at proxy exit);
  HTTP proxy via `ProxyHandler` absolute-URI (same effect). Loopback → `ProxyHandler({})`.

## The fetch ladder (`fetch_url` 1559-1719)

1. hand-delivery door is checked **by the pipeline before calling** (`Door 0`).
2. Rate-limit wait.
3. **Door 1**: `_fetch_once` via proxy opener (if configured & preflight passed):
   - headers `_request_headers` (746-774): Chrome-131 navigation set in Chrome's order
     (sec-ch-ua trio only with the Chrome UA; `Accept-Encoding: gzip, deflate`).
   - `_open_following` (855-921): own redirect loop — Location/URI (spaces→`%20`), cookies
     absorbed incl. from error responses, Referer full-URL same-host / origin-only cross-host,
     loop detection, >5 → `'more than 5 redirects (a tracker chute…)'`, non-http target error,
     no-Location error.
   - body: 64 KB chunked reads to cap; compressed→capped-decode (`_decompress_capped`, gzip via
     GzipFile / deflate zlib then raw); truncated → `partial 'size cap ({n:,} bytes) reached'`;
     PDF → `partial 'pdf'`; success → `full`.
   - error taxonomy: `_http_error_verdict` (949-998) — 402 paywalled; 401/405 refused;
     404/410 dead; 429 blocked_bot (bot hint) or retry_later w/ Retry-After; 503 retry_later;
     403 blocked_bot; 521-524 retry_later + Cloudflare origin sentence; 408/≥500 retry_later;
     else refused. `_RedirectBrokenError` → `redirect_broken`; `URLError` → `connection: …`
     (`proxy_error` for WebProxyError else `retry_later`); socket.timeout → `'timeout'`.
4. HTTP error **not** in refusal set → dead-page rescue (`_dead_page_rescue` 1380-1428): URL
   variants (trailing slash, www/non-www — never scheme downgrades, never IPs) then Wayback
   availability + snapshot fetch → `partial` + `category='archived'` with the honest reason.
5. Connection failure or route-refusal → **Door 2 direct** `_fetch_once(proxy=None)`: success
   reason `"via direct fallback (proxy path failed: …)"`; both fail → combined
   `"proxy: {a} | direct: {b}"` + `_combined_category` precedence.
6. `_final_verdict` (1514-1556): DNS suffix via DoH (NXDOMAIN → dead + suffix
   "dns: NXDOMAIN even via DNS-over-HTTPS…"; exists → "the local resolver is lying…"); rate
   penalty; 521-524 → wayback courtesy note.
7. **Door 3 impersonation** (`_maybe_impersonate` 1457-1511): fires iff `http_status ∈
   {403,405,429,451}` or reason carries TLS-handshake markers; loopback never; one attempt per
   allowed route `[proxy, None]`; curl_cffi `Session(impersonate="chrome")` (one cached session
   so clearance cookies persist); success reason `"via impersonated Chrome (curl_cffi) — the
   stdlib doors were walled: {reason}"`; failure appends `" | chrome-impersonated: …"` + challenge
   hint `" — the challenge needs a live browser; the fetcher cannot solve it"`; missing lib →
   install hint or autopip (240 s `pip install`).

**Doors 4-5** live in hand_delivery/chrome_tabs (separate specs) and are orchestrated by the
pipeline, not fetch_url.

## Go mapping (ADR-0006)

- Routes as `*http.Transport` per (proxy|direct|loopback); custom `CheckRedirect`
  (`ErrUseLastResponse`) + manual loop; cookie jar shared; chunked capped reads;
  `DomainRateLimiter` identical semantics.
- Impersonation door via tls-client Chrome profile (PoC T-02); **no autopip** (always compiled
  in; config key read, no-op — documented).
- Wayback/DoH: plain GETs with the same endpoints/limits/UA.

## Tests

Fake-server suite: redirect chains (5-hop, loop, no-Location, non-http), cookie carry-over,
Referer policy, caps+partial, gzip/deflate, PDF, every HTTP status → category/reason table,
rate limiter + penalties, DoH verdicts (fake resolver), Wayback rescue (fake archive), TLS-marker
trigger matrix for the impersonation door (fake tls failure), loopback-never-proxied.
Port from `test_ladder`, `test_webproxy`, `test_thirddoor`.
