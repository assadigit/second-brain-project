# Module Spec — `gui/app.py` + `gui/headless.py` + `gui/log_bridge.py` + `gui/telegram_lazy.py`

- **LOC**: 474 + 282 + 84 + 63 · **Go home**: `cmd/gitcurator` (entry) +
  `internal/app/lifecycle.go` (lock, logging) + `internal/cli` (headless doors) ·
  **Status**: rehost/merge
- **Responsibility**: the facade entry point (re-export surface + `main()` + single-instance
  lock + rotating logs), the headless batch runner, the logging→Qt-signal bridge, and the lazy
  Telethon import shim.

## app.py (`gui/app.py:`)

- `__VERSION__` stamp printed to stderr at import, suppressed under `--cli` (31-38).
- **Re-export facade**: every historical import path kept working (174-227) — not needed in Go
  (fresh module system); the tombstone sections (262-380) document where code moved.
- `setup_logging(log_level, log_dir)` (234-255): console DEBUG + two RotatingFileHandlers
  (`logs/processing_<ts>.md` INFO, `.txt` DEBUG, 5 MB × 5 each). Appears unused by the GUI path —
  keep the file-logging behavior in Go via slog handlers.
- `main()` (387-470): `--headless` branch; **single-instance lock** `app.lock`
  (O_CREAT|O_EXCL) with PID liveness healing (`_is_process_running`, Windows zombie-handle
  retries), "Another instance is already running" + exit 1; HiDPI rounding policy PassThrough;
  `QApplication`; Fusion style; `MainWindow`; `show()`; `app.exec()`; finally: close+remove lock.
- Telethon deliberately NOT imported (cold-start saving, 155-163).

**Go**: `cmd/gitcurator/main.go` implements the same lock protocol (same filename, PID file,
stale healing) — cross-binary mutual exclusion with the Python app is a hard requirement
(ADR-0008); `slog` rotating file handlers with the same filenames.

## headless.py (`gui/headless.py:`)

- `run_headless(args)` (55-244): `QCoreApplication`; modes `--import-file`, `--single-id`
  (lazy Telethon), `--from-id/--to-id`, `--offset-start/--count`; worker started with
  `headless=True` (fail-fast doors: no code prompt, no disk-full spin, no LLM decision wait —
  219-227); on finish: GitHub seal → Websites seal (silent no-op when off) → Personal seal →
  GoodRepos publish → DONE/ERROR; model auto-switches persisted.
- Go: the CLI batch command with the same post-batch chain and fail-fast semantics.

## log_bridge.py

- `logging.Handler` → Qt signal (`log_signal.emit(msg, level)`), root logger INFO. Go: an slog
  handler publishing `LogLine` events on the bus.

## telegram_lazy.py

- Lazy Telethon import with failing stub. **Not ported** (Go imports are cheap; gotd is a normal
  dependency).

## Tests

Lock: acquire/stale-heal/zombie-pid/cross-binary (Go exe + Python exe) — the cross-binary case
is new and mandatory. Rotating log files created with the right names. Headless fail-fast
timings (no 5-min hangs).
