# Module Spec — `core/connection_check.py` + `core/net_guard.py` + `core/netctx.py`

- **LOC**: 789 + 450 + 43 · **Go home**: `internal/connectioncheck` (battery + sentinel +
  ssl-context helpers) · **Status**: port (merged group)
- **Responsibility**: the four-section Test Connection battery (vaults, LLM, GitHub, Telegram +
  Worker version), the background sentinel trio (internet/AI/proxy) with the warn-once offline
  verdict, and the `verify_ssl` TLS context switch.

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
- `run_local_checks(config, on_section, on_result)`: sections in order; each check
  try/except → `"check crashed ({Type}: {e})"`; callback exceptions swallowed.
- `summarize(sections)`: headline `"ALL SYSTEMS READY — vaults, Telegram, LLM and GitHub are up"`
  or `"{ready}/{total} subsystems ready — N error(s) and M warning(s) (see the lines above)"`.

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
