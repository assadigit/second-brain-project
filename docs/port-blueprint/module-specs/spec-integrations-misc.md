# Module Spec — `integrations/{backfill_manager,error_reporter,banish_confirm,scan_confirm}.py`

- **LOC**: 364 + 288 + 131 + 153 · **Go home**: `internal/workerclient` (backfill.go,
  errors.go, confirm.go) · **Status**: port (merged group: the Worker-client satellites)
- **Responsibility**: the one-time bot-queue backfill; the local error outbox; the two Telegram
  confirmation round-trips (banish + scan).

## backfill_manager.py (`integrations/backfill_manager.py:`)

- `BackfillManager(sync, error_reporter)`; `run(all_links, chunk_size=50, progress_callback)`:
  dedupe by `url_normalized` before chunking; `POST /api/backfill` chunks; 3 retries per chunk
  (2 s × attempt); 0.5 s pause between chunks (rate-limit courtesy); `_verify(expected)` via
  `GET /api/verify` (`stats.total_ledger ≥ expected`); failures logged via the error reporter
  (`BACKFILL_CHUNK_FAILED` CRITICAL / `BACKFILl_VERIFY_MISMATCH` WARNING — typo is verbatim in
  source / `BACKFILL_COMPLETE` INFO); result `{total, inserted, skipped, chunks,
  duration_seconds, verified, error?}`; cancel flag per chunk.
- `extract_links_from_bot_messages(messages)` (280-350): static; builds
  `{url_normalized, url_original, url_type: github|non_github, github_owner, github_repo,
  first_seen_at}` using `CloudflareSync.normalize_url`.
- `is_cutover_complete()` reads the Worker verify flag.

## error_reporter.py (`integrations/error_reporter.py:`)

- SQLite `error_outbox.db` (short-lived connection per op; `threading.Lock`): table
  `error_outbox (id, severity, error_code, message, details JSON, occurred_at ISO-UTC, pushed,
  pushed_at, push_count)` + partial index on `pushed=0`.
- `log(severity, error_code, message, details)`: invalid severity → WARNING; also `print()`ed
  with an emoji. `get_unpushed(limit=100)`; `push_to_worker(sync)` → `POST /api/errors`
  `[{severity, error_code, message, details, occurred_at}]` (CRITICAL routes to an immediate
  Telegram DM server-side); rows marked pushed only on success; `cleanup(days=30)`;
  `prune_excess(max_count=1000)` (oldest DEBUG first); `get_stats()`.
- Severity constants CRITICAL/WARNING/INFO/DEBUG.

## banish_confirm.py (`integrations/banish_confirm.py:`)

- `make_telegram_confirm(config, log)` → `confirm(items, log) -> {'verdict': confirmed|
  declined|timeout|defer, 'report': callable|None}` or None (Worker not paired/enabled).
- Legs (all HMAC + 1010 UA): `POST /api/banish/propose {install_id, count, items:[{url,title,
  marker,door}], timeout_s}` → `{success, id}`; poll `GET /api/banish/status?id=` every
  **3.0 s** (Worker-unreachable None = still pending) until verdict or `DEFAULT_TIMEOUT_S 300`;
  `POST /api/banish/result {install_id, id, outcome: deleted|timeout, deleted: n}`.
- Verdict mapping: `superseded/timeout/done/stale` → 'timeout' (the safe defer — nothing
  deleted); `report(deleted)` only attached to a confirmed verdict, best-effort.

## scan_confirm.py (`integrations/scan_confirm.py:`)

The banish twin for the whole scan plan: `POST /api/scan/propose {install_id, deletions, moves,
new_folders, items:[{kind: delete|move, title, detail, marker}], summary ≤400 chars, timeout_s}`;
poll `/api/scan/status`; `POST /api/scan/result {install_id, id, outcome: applied|timeout,
applied, moved, folders}`. Move items speak the LANDABLE name (`from → to (as <name>)` — the
v0.82 bent-landing law). Same 3 s poll / 300 s window / safe-defer semantics.

## Tests

Loopback rig cases: propose→status→result sequences for each verdict; timeout safe-defer;
unreachable-poll tolerance; backfill chunking/retry/verify; error outbox lifecycle + push
marking; severity coercion.
