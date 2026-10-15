# Module Spec — `app/gitcurator/core/chrome_tabs.py`

- **LOC**: 2,583 · **Go home**: `internal/chrometabs` · **Status**: port
- **Responsibility**: the fifth door — the owner's real Chrome fetches walled pages. Hand-rolled
  RFC 6455 WebSocket + minimal CDP layer; Chrome launch/attach session ladder; per-site proxy
  route memory; wave-based tab engine with settle-watch and honest `page_is_real` verdicts;
  end-to-end delivery into the fourth door's folder.

## Constants (`chrome_tabs.py:`)

`CONFIG_GATE web_browser_retry` · `DEFAULT_PAGE_TIMEOUT_S 45` · `DEFAULT_WAVE 8` ·
profile knobs (`web_browser_profile_persist/dir`, `web_browser_attach`, `web_browser_attach_port`
9222, `web_browser_real_profile`, `web_browser_proxy` auto|off|on) ·
`CHROME_STARTUP_TIMEOUT_S 25` · `MAX_MESSAGE_BYTES 64 MB` · `_CHROME_FLAGS` (no-first-run,
no-default-browser-check, backgrounding-disablers, `--disable-blink-features=AutomationControlled`)
· `PAGE_SETTLE_S 2.0` · `_POLL_CADENCE_S 0.4` · `_ANTI_BOT_SCRIPT` (navigator.webdriver erasure)
· `_SOCKET_RECONNECTS 2` + `_RECONNECT_PAUSE_S 0.5` · `_ERR_CODE_RE ERR_[A-Z0-9_]{2,}` ·
crash markers ('aw, snap', "he's dead, jim", 'err_crashed') · challenge title strong/weak sets +
HTML markers (challenge-platform, cf-chl, turnstile, cdn-cgi/challenge, …) ·
`_CHALLENGE_TEXT_FLOOR 500` · `_MIN_REAL_HTML_BYTES 200`.

## Components

| Component | Lines | Contract |
|---|---|---|
| WebSocket client (`encode_client_frame`/`decode_frames`/`CDPSocket`) | 484-715 | masked client frames (os.urandom mask); unmasked server frames + fragmentation; ping→pong; close→`CDPClosedError`; 64 MB guard `a DevTools frame claimed {n} bytes…`; handshake `GET {path} HTTP/1.1` + `Sec-WebSocket-Key` base64(urandom(16)), requires `101`, 64 KB response cap; `call(method, params, timeout)` — id increment, `json.dumps(ensure_ascii=False)`, skip events/stale ids, `CDPError '{method}: {message}'`, deadline → `CDPTimeoutError '{method} timed out after {n}s'` |
| CDP usage | — | `Runtime.evaluate` (`returnByValue:true, awaitPromise:false`) with expressions `document.readyState + "\|" + location.href`, `document.documentElement.outerHTML`, `document.title`; `Page.enable` + `Page.addScriptToEvaluateOnNewDocument` (anti-bot); `Browser.close` (graceful) |
| HTTP endpoints (loopback) | 788-857 | `GET /json/version` (liveness, 3 s); **`PUT` `/json/new?{url}`** with GET fallback (Chrome 111+ law); `GET`/`PUT` `/json/close/{id}` best-effort |
| `ChromeSession` | 860-1266 | launch ladder (base flags → +container → +headless `--headless=new`); `_try_adopt` via `DevToolsActivePort` file + endpoint 200; `_launch_and_announce` reads stderr line `DevTools listening on ws://127.0.0.1:{port}/devtools/browser/{uuid}` within 25 s; stderr drain daemon thread; `alive()`; `close()` = graceful `Browser.close` (5 s calls, 8 s wait) then Windows `taskkill /PID {pid} /T` (+`/T /F`) / POSIX killpg SIGTERM→SIGKILL; cleanup erases only throwaway profiles |
| `AttachedChromeSession` | 1274-1328 | never launches/kills; closes only its own tabs |
| Real-profile rung | 1343-1601 | `%LOCALAPPDATA%\Google\Chrome\User Data` (win) / POSIX paths; junction `cmd /c mklink /J` at `<app>/chrome-real-link` (15 s) or symlink; dangling link re-point; never delete a real dir; `write_chrome_attach_bat` (ASCII CRLF .bat: taskkill chrome, wait 2 s, mklink, start with `--remote-debugging-port`, `--user-data-dir`, `--restore-last-session`) |
| Chrome discovery | 398-414, 1450-1478 | Windows program-files paths; POSIX `shutil.which` over chrome/chromium/msedge names; `chrome_is_running` via tasklist / pgrep |
| Proxy legs | 1609-1741 | `--proxy-server={type}://host:port` or `--no-proxy-server`; route memory `<APP_DIR>/chrome-proxy-memory.json` (`{domain: direct|proxy}`, atomic, indent 2, non-ASCII raw); `plan_proxy_legs` (auto = direct first, proxy ALWAYS planned as escalation landing pad); `remember_route` only for delivered pages |
| Wave engine | 1763-1820, 1832-2033 | waves of `wave` tabs opened first (parallel load), taken in turn; `should_continue` stop → `'stopped by the user'`; attach-first (`_probe_version_endpoint(9222)`) → `AttachedChromeSession`; else legs with escalation of direct-leg failures onto the proxy leg; per-URL `{'url','ok','html','title','error'}`; both-leg failures joined `' \| '` |
| `_take_live_page` | 2177-2242 | poll readyState+href every 0.4 s; must be `complete` at an http(s) URL and **stay** 2.0 s (href change restarts settle; challenge verdict restarts settle and keeps budget); deadline → honest reasons |
| `page_is_real(html, href, title)` | 2061-2113 | scheme http(s); no `ERR_*` code; no crash marker; no strong challenge title; weak-title words only when title ≤60 chars and ≤4 words; HTML markers only when visible text <500 chars; exact reason strings |
| `collect_failed_fetch_links` | 2342-2383 | end-of-run modal data (outcome review/failed; dedup by `normalize_website_url`; loopback dropped) |
| `deliver_pages_via_chrome` | 2391-2563 | reads config knobs; enqueues via hand_delivery (`door='auto'`); fetches; re-verifies every page before writing; writes `<folder>/<suggested_filename>` via atomic_write_bytes; returns `{'delivered','failed','urls','failed_links','folder'}` |

## Go notes

- WebSocket: use `gorilla/websocket` or `nhooyr.io/websocket` — but keep the exact handshake
  against the tab's `webSocketDebuggerUrl` and the same message-size guard semantics.
- Windows process control: `taskkill /T`, `mklink /J` via cmd — keep subprocess calls.
- Keep every honest error string (they surface in modals/logs).

## Tests

Frame codec round-trips (incl. fragmentation, ping, oversize guard); CDP call timeout/stale-id;
page_is_real grammar table (ERR codes, crash markers, strong/weak titles, HTML markers with
text floor); settle-watch state machine with a fake tab; route-memory plan/remember; deliver
re-verification (fake fetcher injecting a challenge page → discarded). Integration (tagged):
local Chromium end-to-end fetch of a data: URL.
