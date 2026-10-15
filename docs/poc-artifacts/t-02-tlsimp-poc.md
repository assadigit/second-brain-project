# T-02 — Chrome-TLS impersonation PoC (ADR-0006) — evidence

Recorded 2026-10-15, SESSION-001. Environment: Debian 13 sandbox, go1.27.2
linux/amd64, datacenter egress IP. Binary: `poc/tlsimp`.

## 1. Trigger matrix on real walled pages (stdlib transport)

`tlsimp -mode=probe` — Go stdlib with the Chrome navigation header set
(the same shape `core/web_fetch.py` sends). The Python ladder's trigger
statuses (403/405/429/451) reproduce faithfully:

| URL | stdlib status | wall |
|---|---|---|
| https://www.pexels.com | 403 | Cloudflare |
| https://www.yelp.com | 403 | DataDome |
| https://www.patreon.com | 403 | Cloudflare |
| https://www.fandom.com | 403 | Cloudflare |
| https://www.ebay.com | 403 | Akamai |
| https://www.stockx.com | 403 | Cloudflare |
| https://www.zillow.com | 403 | CloudFront |
| https://www.wayfair.com | **429** | Cloudflare |
| https://www.macys.com | 403 | Akamai |
| https://www.gamestop.com | 403 | Cloudflare |
| https://www.newegg.com | 403 | Cloudflare |
| https://www.homedepot.com | 403 | Akamai |
| https://www.reddit.com | 403 | (snooserv) |
| https://www.chegg.com | 403 | CloudFront |
| https://nowsecure.nl | 200 (UA-based wall only — passed stdlib) | Cloudflare |

## 2. Fingerprint substitution (live, tls.peet.ws echo)

`tlsimp -mode=fingerprint`:

| transport | JA4 | JA3 hash | HTTP |
|---|---|---|---|
| Go stdlib | `t13d1311_f57a46bbacb6_f50d94e863eb` | `20b27999…68fb` | HTTP/1.1 (h2 aborted) |
| tls-client `Chrome_133` | `t13d1516h2_8daaf6152771_d8a2da3f94cd` | `74e530e4…78c7` | h2 (Akamai h2 fp `52d84b11…ca86`) |

**FINGERPRINT_SUBSTITUTED** — the profile presents genuine Chrome
client-hello + HTTP/2 fingerprints at exactly the layer Cloudflare
inspects.

## 3. Live 403→200 limitation (honest) + hermetic wall proof

Every live wall above blocks the sandbox's datacenter IP at the
IP-reputation layer for BOTH transports (e.g. pexels: stdlib 403 /42 ms/,
Chrome_133 403 /35 ms/; same on patreon, fandom, stockx, wayfair 429/429,
macys, gamestop, newegg, homedepot, reddit, chegg). A live 403→200
demonstration therefore cannot be produced from this IP — that is an
environment property, not a library property.

The mechanism itself was proven against a locally controlled
fingerprint gate with Cloudflare's decision logic (GREASE-led,
≥14-cipher ClientHello ⇒ Chrome; Go never sends GREASE):

```
$ tlsimp -mode=localwall
[wall] fingerprint gate listening on https://127.0.0.1:8443 ...
[poc] stdlib   vs local fingerprint wall → HTTP 403
[poc] tls-client(Chrome_133) vs local fingerprint wall → HTTP 200
[poc] verdict: WALL_BYPASSED (stdlib=403, impersonation=200)
```

On the owner's residential connection (via their SOCKS proxy) the live
walls are expected to behave like the hermetic gate; opportunistic
re-verification lands with Phase 3 (T-31 impersonation matrix).

## 4. SOCKS5 hostname passthrough (rdns=True law)

`tlsimp -mode=fingerprint -socks=127.0.0.1:10808` (relay = `poc/telegram
-mode=socksrelay`) — same Chrome fingerprint through the proxy, and the
relay log records:

```
[relay] conn#1 CONNECT tls.peet.ws:443 (atyp=DOMAIN) — rdns passthrough evidence
```

The proxy receives the DOMAIN, unresolved — matching PySocks `rdns=True`.
(Verified equivalently for gotd in the T-01 artifact: DC addresses relay
with IPv6→IPv4 fallback.)

## 5. Bonus capability

The Chrome profile negotiated **HTTP/3 (QUIC)** against reddit.com
(`proto=HTTP/3.0`) — tls-client ships an h3 path when the profile allows
it. (GitCurator's web_fetch never uses h3; recorded for completeness.)

## 6. License check (reality vs the ADR's assumption)

| module | version | license (from module cache LICENSE files) |
|---|---|---|
| bogdanfinn/tls-client | v1.16.0 | **BSD-4-Clause** (NOT Apache-2.0 as ADR-0006 assumed) |
| bogdanfinn/fhttp | v0.6.9 | **no LICENSE file** (net/http fork; GitHub reports license: None) |
| bogdanfinn/utls | v1.7.8-barnius | BSD-3-Clause (Go Authors) |
| refraction-networking/utls | v1.8.2 (gotd dep) | BSD-3-Clause (Go Authors) |
| bogdanfinn/quic-go-utls | v1.0.10-utls | MIT |
| bogdanfinn/websocket | v1.5.6-barnius | BSD (Gorilla WebSocket Authors) |
| bdandy/go-socks4 | v1.2.3 | MIT |
| tam7t/hpkp | 2016 pin | MIT |

BSD-4-Clause carries the advertising-acknowledgment clause: harmless for
this private project, must be attributed if ever published. The fhttp
license gap is the same shape (private use fine, review before any
public release). utls hand-roll (Go BSD-3) remains the fallback.
**Owner acknowledgement requested** — see NEXT-SESSION.md.
