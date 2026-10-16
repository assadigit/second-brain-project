# Module Spec — `core/connection_check.py` + `core/net_guard.py` + `core/netctx.py`

> **RE-SCOPED (2026-10-15, ADR-0004 Am.3 — owner stamp "OPTION B CONFIRMED", rider R-0):** the
> Go app drops its direct Telegram (MTProto) leg entirely. The Test Connection battery therefore
> runs **three sections — vaults, LLM, GitHub — plus the Worker leg**: `check_telegram_local` is
> **dropped**, and the existing `check_worker_version` (GET `/health`, proxy bypassed, UA law)
> is promoted from an auxiliary version check to **the queue-reachability leg** (the queue rides
> the Worker HTTP API only). Everything below describing the Telegram section is retained
> verbatim as the Python-reference behavior record; the Go port (T-33) implements the re-scoped
> battery, pinning the new `summarize` wording (the Python headline's "Telegram" mention goes)
> in its result-shape tests. `net_guard` and `netctx` are unaffected.

- **LOC**: 789 + 450 + 43 · **Go home**: `internal/connectioncheck` (battery + sentinel +
  ssl-context helpers) · **Status**: port (merged group; battery re-scoped per the banner above)
- **Responsibility**: the Test Connection battery (vaults, LLM, GitHub, Worker
  version/reachability — Telegram dropped per the banner), the background sentinel trio
  (internet/AI/proxy) with the warn-once offline verdict, and the `verify_ssl` TLS context
  switch.

## connection_check.py (`connection_check.py:`)

- Result line = `{'name','level' ok|warn|error|info,'detail'}`; `render_line` = `"{icon} {name} —
  {detail}"` with icons ✅⚠️❌ℹ️; `short_line` budget 72 chars (drops pathy parentheticals,
  mid-ellipsizes).
- `_http_get_json(url, timeout 8)` UA `GitCurator-connection-check`; loopback direct.
- Sections: `check_vaults` (github must exist; websites parent-writability; manual owner-created
  + writability probe `.gc-write-test-` temp file) · `check_llm` (routes ollama/llamacpp/cloud;
  hint line appended when the active provider errors) · `check_github` (`GET /user` with
  Bearer + `application/vnd.github+json`; 401 → the fresh-token message with rate context; then
  `/repos/{login}/{repo}` per configured vault repo — 404 → the v0.83 never-creates info line;
  public repo → warn) · `check_telegram_local` (credentials with masked phone (first 5 + bullets),
  session file at `<APP_DIR>/session.session`, bot username, proxy TCP probe with latency ms,
  the "Telegram needs it in Iran" note) · `check_worker_version` (`GET {worker}/health` **proxy
  bypassed**, UA law; compare `EXPECTED_WORKER_VERSION`; older → warn with deploy pointer).
  - **Dated note (2026-10-15, ADR-0004 Am.3 / R-0):** in the Go port `check_telegram_local` is
    **dropped** (no MTProto client ships; the Telegram config keys keep being *read* for
    compatibility — the compat law — but no session file, proxy probe, or login ever happens),
    and `check_worker_version` is promoted to **the queue-reachability leg** (the queue rides
    the Worker HTTP API only), so it covers both the version gate and queue reachability.
- `run_local_checks(config, on_section, on_result)`: sections in order; each check
  try/except → `"check crashed ({Type}: {e})"`; callback exceptions swallowed.
- `summarize(sections)`: headline `"ALL SYSTEMS READY — vaults, Telegram, LLM and GitHub are up"`
  or `"{ready}/{total} subsystems ready — N error(s) and M warning(s) (see the lines above)"`.
  - **Dated note (2026-10-15, ADR-0004 Am.3 / R-0):** the Go port's `summarize` drops the
    "Telegram" mention (no Telegram section exists); the exact re-worded headline is pinned by
    T-33's result-shape tests when that task lands (Phase 3).

## net_guard.py (`net_guard.py:`)

- `_SENTINEL_UA "GitCurator-sentinel"`; endpoint ladder `[worker_url?, api.github.com,
  cloudflare-dns.com]`; `_http_transport_ok` — **any HTTP answer (incl. HTTPError) counts as
  ONLINE** (transport worked); drains ≤2048 bytes; returns `(ok, status, latency_ms, note)`.
- `probe_internet` (5 s): success detail `"{ep} answered HTTP {status} in {latency} ms"`; offline
  detail truncates to 300 chars.
- `ai_status` (6 s): ollama (`GET /api/tags` states on/warn/off) / llamacpp (probe+TypeError
  fallback) / cloud preflight → `{'state','kind','engine','detail','short'}`.
- `proxy_reachable` (2 s floor 0.5): `socket.connect_ex`; states disabled/connected/unreachable
  (`"— errno {n}"`).
- `sentinel_snapshot` → `{'ts','net','ai','proxy'}`; `offline_verdict(previous, online)` =
  warn-now / first-online / fell-offline / back-online / online / offline (the warn-once modal
  law); `sentinel_line` = `"Net: online ({host}) · AI: on — local (Ollama) · Proxy: connected"`.
- Runs OFF the GUI thread (the v0.85 raison d'être); 60 s cadence in the GUI.

## netctx.py (`netctx.py:`)

- `verify_ssl_enabled(config)` = `bool(config.get('verify_ssl', False))` (default **false** —
  the historical censored-network setting).
- `outbound_ssl_context(config)`: default context with `check_hostname=False,
  verify_mode=CERT_NONE` when off. Split law: the **websites fetcher always verifies**; LLM/cloud
  and these probes honor the switch.

## Go notes

- One package; the battery returns the same line objects; UA inventory asserted by tests.
- `proxy_reachable` = raw TCP dial with deadline; keep errno-style details.

## Tests

Port `test_connection` (73) + `test_netguard` (56) selections: result shapes, masked phone,
version compare (older/newer), summarize headlines, offline verdict transitions, any-answer-
is-online law, proxy states, verify_ssl split.

- **Dated note (2026-10-16, SESSION-014 / T-33 landed):** the selections actually
  ported are the CORE-module classes; with the Telegram section dropped (the banner
  above), its test classes fall away too — `TestTelegramLocal` (the masked-phone rows)
  and `TestTelegramLiveResult` are NOT ported (no session file, no proxy probe, no
  login ever happens in the Go app). The GUI-wiring classes (`TestGuiBattery`,
  `TestGuiWiring` in both files) belong to the GUI phase (Phase 5+); the CLI classes
  (`TestCli`) to the CLI phase — their source-contract asserts re-home there. The
  landed battery: 132 Go tests — the result shapes, the short_line differential
  (the reference's own `_short_detail` outputs, 22 cases byte-compared, generated at
  the pinned rev), the vaults matrix, the LLM routes, the GitHub ladder with the UA
  table from the wire, the Worker version ladder with the UA + proxy-bypass law, the
  re-scoped battery order (Vaults/LLM/GitHub/Worker) + the re-worded headline pinned
  exactly, the sentinel trio with the any-answer law, the offline-verdict transitions,
  and the verify_ssl split. The net_guard wiring classes' read-only/never-writes laws
  are pinned behaviorally (GET + no body from the wire) rather than by source grep.
