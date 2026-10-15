# ADR-0006 — Network stack: stdlib net/http + SOCKS + Chrome-TLS impersonation

- **Status**: accepted for stdlib/SOCKS; provisional for the impersonation library
- **Date**: blueprint session (2026-10)

## Context

The Python app's outbound network behavior is unusually specific (all verified in investigation
report 1-c):

- **Web fetch** (`core/web_fetch.py`): stdlib urllib through a hand-rolled redirect loop (≤5 hops,
  cookie carry-over, Referer policy per hop), Chrome-131 navigation header set, per-domain rate
  limiting (2 s default + Retry-After penalties), 2 MB streamed cap, gzip/deflate capped decode,
  TLS **always verified** for websites; SOCKS5 with `rdns=True` (**DNS resolved at the proxy exit**),
  HTTP proxy via `ProxyHandler`; loopback never proxied.
- **Third door — curl_cffi** (`impersonate="chrome"`): Chrome's TLS fingerprint for sites that wall
  the stdlib client (403/405/429/451 or TLS-handshake markers). Optional dependency with a
  self-install (`pip install curl_cffi`, `web_fetch.py:276-308`) and one cached session.
- **LLM clients** (`core/llm_client.py`): urllib-based OpenAI-compatible + Anthropic Messages
  clients; `verify_ssl=False` default (netctx) behind config; loopback direct.
- **Worker client** (`cloud/cloudflare_sync.py`): proxy **bypassed** (`ProxyHandler({})`) for
  workers.dev; exact UA `GitCurator/<version> (+https://github.com/assadigit/GitCurator)`.
- **Chrome CDP fifth door** (`core/chrome_tabs.py`): raw-socket WebSocket to DevTools — no HTTP
  client involved.

## Problem

Reproduce this behavior in Go: proxy semantics (esp. DNS-at-exit), the redirect/cookie/UA ladder,
and a curl_cffi-equivalent Chrome TLS fingerprint.

## Decision

1. **Base stack: Go `net/http` with custom `http.Transport`** per route:
   - Proxy routes: `Transport.Proxy = http.ProxyURL(...)` for HTTP proxies. For SOCKS5, dial via
     `golang.org/x/net/proxy` SOCKS5 dialer with the **hostname passed unresolved** (that is the
     Go default when dialing through `proxy.SOCKS5` — the dialer sends the domain to the proxy),
     matching Python's `rdns=True`. Wrap into a custom `Transport.DialContext`.
   - Direct routes: `Transport.Proxy = nil` explicitly (Go's default honors `HTTP_PROXY`/`HTTPS_PROXY`
     env vars — **must be disabled** to match Python's explicit-opener semantics; see risks).
   - Loopback detection: one shared helper replicating `llm_client._is_loopback_url`
     (`core/llm_client.py:48-70`; note `chrome_tabs.is_loopback_url` additionally treats `.local`
     as loopback — keep both semantics where each was used).
2. **Redirect loop: implement manually** with `CheckRedirect: func(...) error { return http.ErrUseLastResponse }`
   and the same hop logic (≤5, cookies absorbed from error responses too, Referer full-URL same-host
   / origin-only cross-host, loop detection, `_redirect_target` Location/URI + space-encoding).
3. **Politeness:** replicate `DomainRateLimiter` (per-domain mutex map, 1 s sliced sleeps, penalty
   extension) and the byte cap via `io.LimitReader`-style chunked reads (64 KB chunks) with the same
   "partial at cap" semantics.
4. **Chrome-TLS impersonation (the curl_cffi replacement)**: use
   **`github.com/bogdanfinn/tls-client`** (wraps `refraction-networking/utls` + HTTP/2, offers
   Chrome client profiles) — *provisional*: verify license (Apache-2.0 at last check — confirm in
   implementation), API stability, and SOCKS support in the PoC.
   - Fallback: hand-roll with `refraction-networking/utls`
     (`tls.UClient(conn, cfg, tls.HelloChrome_Auto)`) + `http2.Transport`.
   - The degradation ladder, trigger statuses (403/405/429/451 + TLS markers), "challenge needs a
     live browser" hint strings, and one-session-cookie caching must match
     `spec-core-web-fetch.md`.
   - **No auto-install equivalent**: the "autopip" behavior is dropped (Go binaries can't self-
     install deps) — the capability is always present in the binary. Config keys
     `web_impersonate_fallback` / `web_impersonate_autopip` keep being read; `autopip` becomes a
     no-op (documented compatibility note).
5. **DoH + Wayback auxiliary doors**: plain `net/http` GETs with the same endpoints/limits
   (`cloudflare-dns.com/dns-query`, `archive.org/wayback/available`).
6. **User-Agent inventory** is pinned exactly (see `08-data-and-compatibility.md` §UA table).

## Consequences

- The `verify_ssl` split is preserved: websites always verify; LLM/cloud paths honor the config
  default (off) — implemented as a per-transport `TLSClientConfig`.
- Python's `_urlopen_direct` (empty ProxyHandler) maps to `Transport.Proxy = nil` + env-clearing.
- Go follows redirects by default — the custom `CheckRedirect` is mandatory, easy to miss; it is
  called out as a parity test case (golden redirect fixtures).

## Evidence

- Header set: `core/web_fetch.py:746-774`; redirect loop: `web_fetch.py:855-921`; rate limiter:
  `web_fetch.py:375-423`; impersonation trigger: `web_fetch.py:1457-1511`
- SOCKS rdns: `web_fetch.py:504-527`
- UA law: `cloud/cloudflare_sync.py:41-59`

## Verification required in implementation

- PoC: fetch a known-Cloudflare-protected URL via stdlib route (expect 403) and via tls-client
  Chrome profile (expect 200); verify SOCKS5 hostname passthrough (proxy log shows domain, not IP).
