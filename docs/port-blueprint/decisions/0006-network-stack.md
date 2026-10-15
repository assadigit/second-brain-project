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

## Amendment 1 — PoC verdict (T-02, SESSION-001, 2026-10-15)

**Status: accepted — `github.com/bogdanfinn/tls-client` pinned at v1.16.0**, with two
reality corrections and one owner acknowledgement requested.

Evidence (full detail: `docs/poc-artifacts/t-02-tlsimp-poc.md`):

1. **Fingerprint substitution proven live** (tls.peet.ws echo): stdlib JA4
   `t13d1311_f57a46bbacb6_f50d94e863eb` over HTTP/1.1 vs Chrome_133 profile JA4
   `t13d1516h2_8daaf6152771_d8a2da3f94cd` with Chrome's H2 fingerprint — the exact layer
   Cloudflare inspects.
2. **Trigger matrix reproduced**: Go stdlib receives 403/429 on 14 real walled pages
   (Cloudflare/Akamai/DataDome/CloudFront) — the Python ladder's trigger statuses
   (403/405/429/451) are faithfully reproducible.
3. **Live 403→200 not demonstrable from the sandbox**: every live wall blocks its
   datacenter IP at the IP-reputation layer for BOTH transports (statuses recorded).
   The mechanism itself was proven hermetically against a locally controlled
   fingerprint gate (GREASE-led, ≥14-cipher ClientHello ⇒ Chrome): stdlib → **403**,
   Chrome profile → **200** (WALL_BYPASSED). Opportunistic live re-verification on the
   owner's residential connection is folded into T-31 (Phase 3 impersonation matrix).
4. **SOCKS5 verified** (`WithProxyUrl("socks5://…")`) through the hermetic relay: the
   proxy receives `CONNECT tls.peet.ws:443` with ATYP=DOMAIN — hostname unresolved,
   the `rdns=True` law holds.
5. Bonus: the Chrome profile negotiated HTTP/3 (QUIC) against reddit.com — recorded,
   unused by web_fetch.

**License reality check** (corrects the "Apache-2.0 at last check" assumption above):

- tls-client v1.16.0: **BSD-4-Clause** (Bogdan Finn, 2023) — advertising-acknowledgment
  clause; harmless for this private project, attribution required if ever published.
- fhttp v0.6.9: **no LICENSE file** (net/http fork; GitHub reports license: None) —
  private use fine, review before any public release.
- utls (both forks): Go BSD-3-Clause; quic-go-utls: MIT; websocket fork: Gorilla BSD;
  go-socks4 and hpkp: MIT.

**Owner acknowledgement requested** for the BSD-4-Clause + fhttp combination; the utls
hand-roll (Go BSD-3) stays the recorded fallback if the owner declines.

## Amendment 2 — Owner acknowledgement (2026-10-15)

The owner confirmed the SESSION-001 status report in full — the report that carried this
ADR's licence-reality correction and its acknowledgement request. Recorded decision:

- **Accepted**: `bogdanfinn/tls-client` v1.16.0 (**BSD-4-Clause**) stays the pinned
  impersonation library; fhttp v0.6.9 (no LICENSE file) is accepted for private use.
- The **utls hand-roll remains the recorded fallback** — the switch path stays documented
  and is taken immediately if the owner ever revokes this acknowledgement.
- The **review-before-any-public-release** precondition stands: the BSD-4-Clause
  attribution clause and the fhttp license review must happen before the repo or any
  artifact is ever made public (consistent with the private-until-launch law).
