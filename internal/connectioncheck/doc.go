// Package connectioncheck ports core/connection_check.py +
// core/net_guard.py + core/netctx.py (T-33, Phase 3): the Test
// Connection battery, the background sentinel trio, and the
// verify_ssl TLS-context switch.
//
// Spec: docs/port-blueprint/module-specs/spec-core-connection.md;
// ADR-0004 Am.3 (Option B — the queue rides the Worker HTTP API only).
//
// THE RE-SCOPE (ADR-0004 Am.3 / rider R-0 — the spec's banner):
// the Go app drops its direct Telegram (MTProto) leg entirely, so the
// battery runs FOUR sections — Vaults, LLM, GitHub, plus the Worker leg
// promoted to queue-reachability (the queue rides the Worker HTTP API
// only). check_telegram_local and telegram_live_result are NOT ported:
// no session file, no proxy probe, no login ever happens. The Telegram
// config keys keep being READ by the config layer for compatibility
// (ADR-0004 Am.3's compat law) but nothing here consumes them.
//
// THE PACKAGE LAWS (connection_check.py's own):
//   - NEVER raises — every check returns result values and survives
//     hostile configs, dead ports and missing folders (a crashed check
//     degrades to a "check crashed" error row, never a panic).
//   - Result shape: {name, level ok|warn|error|info, detail} — one log
//     line each.
//   - Loopback HTTP never rides the system proxy (the v0.15.1 rule);
//     api.github.com is NOT loopback, so it honors the environment
//     proxy exactly like the pipeline does.
//   - The Worker /health fetch deliberately BYPASSES the proxy
//     entirely (mirroring cloudflare_sync) — owners behind an
//     intercepting proxy would otherwise see a false ⚠️/❌.
//
// THE SENTINEL LAWS (net_guard.py's own):
//   - ANY HTTP answer (401/404 included) proves transport — only
//     transport failures count as offline.
//   - READ-ONLY — GETs only, no payload, no write verb anywhere.
//   - The warn-once offline law: the offline-launch modal fires
//     exactly once per session (offline_verdict is the pure logic).
//
// DOCUMENTED DIVERGENCES (each per the port law):
//   - Exception CLASS names in detail strings render Python's urllib
//     taxonomy where a mapping exists (URLError for transport errors,
//     TimeoutError for timeouts, ValueError for bad JSON,
//     UnicodeDecodeError for bad UTF-8, FileNotFoundError /
//     PermissionError for path errors); the fallback is Go's %T base
//     name. The pinned contract is the detail SHAPE, not the runtime
//     class text (the reference's own tests assert substrings only).
//   - Python's per-operation socket timeout collapses into
//     http.Client.Timeout (whole-request) — the T-31 precedent.
//   - proxy_reachable renders the REAL syscall errno (Linux
//     ECONNREFUSED = 111, matching the reference's connect_ex); DNS
//     and timeout failures render "Proxy check failed ({class})" with
//     Go's class names (Python: gaierror / timeout).
//   - _version_cmp filters ASCII digits only; Python's str.isdigit()
//     also admits Unicode digit codepoints (Arabic-Indic etc.) which
//     int() then parses — a pathological corner no test pins.
//   - The empty-Worker-section case (no worker URL configured) yields
//     a section with zero rows — summarize's empty-section law treats
//     it as ready, mirroring the reference's None-worker contract.
//   - The re-worded summarize headline ("...GitHub and the sync queue
//     are up") is this port's pinned decision per the spec's banner
//     (the Python headline's "Telegram" mention goes).
package connectioncheck
