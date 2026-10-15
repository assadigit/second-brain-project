# 06 — Dependency Migration

> Python requirements from `app/requirements.txt` (9 packages) + stdlib features that matter + the
> GUI framework. External URLs were NOT consulted in this session (no network research performed);
> items needing verification are marked **[verify]** with the exact check to run in the
> implementation session.

## 1. The replacement matrix

| Python dependency | Where used (evidence) | Essential? | Go replacement | Semantics / risks | License note |
|---|---|---|---|---|---|
| **PyQt6** ≥6.5 | entire `gui/` (41 files) | essential | **Wails v2** (Go + WebView2) — ADR-0001; Fyne fallback | Behavior ported via commands/events; visual replication from token tables; WebView2 runtime dep | PyQt6 = GPL/commercial (fine, private app); Wails = MIT **[verify version pin]** |
| **Telethon** ≥1.34 | `integrations/telegram_fetch_worker.py`, `telethon_fetcher.py` | essential | **gotd/td** — ADR-0004 | MTProto user client; session format incompatible → one-time re-login; DC/proxy/retry semantics are gotd's | Telethon MIT; gotd **[verify — historically MIT]** |
| **PyGithub** ≥2.1 | `gui/processing_worker.py` (repo metadata, README) | essential | **google/go-github** (+ `golang.org/x/oauth2` static token source) | Same REST surface; header law ported (`Accept: application/vnd.github+json`, UA); 401-anonymous fallback logic is ours | both MIT |
| **ollama** (SDK) | `gui/worker/llm.py`, CLI detect | essential | **direct HTTP** to Ollama REST (`/api/chat`, `/api/tags`, `format:'json'`, `options.num_ctx/num_predict`) | The SDK is thin; payload shapes pinned in `spec-core-llm-client.md`; response dual-shape handling ported | — (drops a dep) |
| **colorama** ≥0.4.6 | `cli/cli_terminal.py:38-78` | incidental | manual ANSI + `golang.org/x/sys/windows` (`SetConsoleOutputCP(65001)`, VT enable); NO_COLOR law | Windows console UTF-8 hardening must run before any output (import-time in Python → `init()` in Go CLI) | MIT |
| **python-dotenv** ≥1.0 | optional env loading | incidental | **drop** (config.json + `GITCURATOR_USER_CONFIG` env are the real doors; no `.env` behavior observed in production paths) | — | — |
| **requests** ≥2.31 | listed in requirements; **no production import found in core/gui/cli** (investigations 1-a…1-f found urllib everywhere) | incidental | **drop** | If a hidden import surfaces, replace with `net/http` naturally | — |
| **PySocks** ≥1.7.1 | `core/web_fetch.py:485-527` (SOCKS5 rdns), Telethon proxy | essential | `golang.org/x/net/proxy` (SOCKS5 dialer; hostname passthrough = remote DNS) + `net/url` HTTP proxy | Must verify hostname-not-resolved-locally (proxy log shows domain); `socks5h://` semantics for curl-style URLs | BSD-3 |
| **curl_cffi** ≥0.7 (optional at runtime) | `core/web_fetch.py:247-325` impersonation door | essential capability | **bogdanfinn/tls-client** (utls wrapper) — provisional; fallback `refraction-networking/utls` + custom transport (ADR-0006) | Chrome TLS fingerprint quality **[verify]** against real walled sites; no auto-install (always compiled in) | **[verify license — believed Apache-2.0]** |
| *(stdlib)* `sqlite3` | all ledgers | essential | `modernc.org/sqlite` via `database/sql` (ADR-0003) | pure Go, no CGO; DSN pragmas **verified at T-20 (ADR-0003 Am.1)**: `?_pragma=busy_timeout(30000)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)`; the declared-type TIMESTAMP read conversion needs the CAST read law; pinned v1.36.1 (go-floor law, ADR-0002) | MIT |
| *(stdlib)* `html.parser` | `core/web_extract.py` | essential | `golang.org/x/net/html` parser walking | Conservative extraction rules must be re-pinned against the 5 HTML fixtures (title concat, meta precedence, invisible-depth) | BSD-3 |
| *(stdlib)* `ctypes` Win32 | user_config Known-Folder, `_is_process_running`, console CP | essential (Windows) | `golang.org/x/sys/windows` (`KnownFolderPath`, `OpenProcess`, `SetConsoleOutputCP`) | Same GUID `{FDD39AD0-238F-46AF-ADB4-6C85480369C7}` | BSD-3 |
| *(stdlib)* `http.cookiejar`, redirect loop | web_fetch | essential | `net/http/cookiejar` + custom `CheckRedirect` ladder | hop-Referer policy + cookie absorption from error responses must be hand-coded | — |
| *(stdlib)* `asyncio` (Telethon) | telegram worker | essential | gotd's own concurrency; Go native | — | — |
| *(stdlib)* `zipfile` (backups, zip tool) | backup_seal, build_zip | essential | `archive/zip` with explicit metadata (determinism) | compresslevel 9 + fixed timestamps for reproducibility | — |

## 2. New Go-only dependencies (introduced by the target architecture)

| Dependency | Purpose | Notes |
|---|---|---|
| `github.com/wailsapp/wails/v2` | GUI shell | pin stable; v3 pre-release — do not ride main |
| `github.com/gotd/td` | Telegram MTProto | ADR-0004 PoC first |
| `github.com/google/go-github/vXX` | GitHub REST | pin the latest stable at implementation |
| `modernc.org/sqlite` | SQLite | ADR-0003 |
| `golang.org/x/net` | SOCKS dialing, html parsing | — |
| `golang.org/x/sys` | Windows specifics | — |
| `github.com/bogdanfinn/tls-client` *(provisional)* | Chrome TLS fingerprint | ADR-0006; fallback `refraction-networking/utls` |
| *(optional)* `github.com/rivo/tview` or none | CLI is plain stdout | prefer plain ANSI; no TUI dependency |

Policy: **no dependency without a mapping in this table**; the repo's own law is "minimal
dependencies … stop and ask" (SPEC.md §2.10) — the blueprint is that ask.

## 3. Semantics that differ and need tests

1. **SOCKS DNS**: Python `rdns=True` resolves at the proxy. Go's `proxy.SOCKS5` dialer passes the
   hostname by default **but** `net/http` may resolve before dialing if given a URL — implement via
   custom `DialContext` with the SOCKS dialer and never pre-resolve. Test: fake SOCKS server
   asserting it received a domain, not an IP.
2. **Redirects**: Go auto-follows; the ladder must disable that (`ErrUseLastResponse`) and
   reimplement hop cookies/Referer/loop detection (parity test with a fake redirect chain).
3. **TLS verification split**: websites verified; LLM/cloud honor `verify_ssl` (default false);
   DoH verified; CDP is plain ws. Per-transport configs, tested.
4. **Ollama wire format**: `format:'json'`, `options.num_ctx` — golden request capture against a
   fake Ollama (the Python test suite already fakes it; port those fakes).
5. **Anthropic**: `anthropic-version: 2023-06-01`, `x-api-key` (no Bearer), required
   `max_tokens` (4096 fallback) — payload goldens.
6. **HMAC client**: key is **sha256hex(secret) used as raw bytes** (the hex string's UTF-8 bytes) —
   an easy-to-get-wrong detail; loopback rig from `test_synclive` ported as Go test.
7. **UA law**: exact strings table (doc 08 §6). Any Go default UA (`Go-http-client/2.0`) is banned
   on Worker legs (Cloudflare 1010) — set explicitly on every client.
8. **git CLI**: same subcommands, `--porcelain` parsing, `%h%x00%s%x00%aI` format, token-URL
   redaction — integration tests with real git in temp dirs.
9. **WebView2 availability** — packaging concern, not library semantics (doc 10).

## 4. Licensing considerations

- Current stack is private/personal ("all rights reserved", README §License) — copyleft is not a
  blocker today, but if the project is ever distributed: PyQt6 (GPL/commercial) vs Wails (MIT) is
  an improvement; gotd/tls-client licenses must be confirmed **[verify]** before release beyond
  personal use.
- The Lucide icon pack (ISC) and Inter font (OFL) embed cleanly; keep the attribution notes from
  `gui/icons.py:11-15` in the Go assets.
