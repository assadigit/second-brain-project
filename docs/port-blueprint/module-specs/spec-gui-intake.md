# Module Spec — `gui/platform_intake.py` + `gui/telegram_lock.py`

- **LOC**: 442 + 181 · **Go home**: `internal/batch/intake.go` + `internal/app/tglock.go` ·
  **Status**: port
- **Responsibility**: non-GitHub link intake into per-platform `_inbox/*.md` tables (idempotent,
  atomic, token-scrubbed, law-filtered) + the prune pass; the thread-safe single-slot Telegram
  lock with owner tracking.

## platform_intake.py (`gui/platform_intake.py:`)

- `_PLATFORM_DISPLAY` (9 platforms with emoji names) + `PLATFORM_INFO` → table filenames from
  `links.PLATFORM_TABLE_FILES`.
- `_inbox_table_vault(config)` (77-89): websites vault when set, else the GitHub vault.
- `write_inbox_links_by_platform` (119-345): two intake filters — banned domains NEVER collected
  (one 🚫 count line, no note/row), already-stored links not re-collected (🔗 line); rows carry
  scrubbed URLs; **new files only when ≥1 row survives** (an all-banned batch leaves no empty
  `_inbox/`); every write prunes legacy rows of both kinds; table grammar
  `| - | date | URL | domain | source | unreviewed | |` with header
  `> Auto-updated. DO NOT delete rows — only update the Status column. Edit Status to: ✅ reviewed / ❌ ignored`;
  atomic writes; dry-run routed through storage.
- `prune_inbox_tables` (348-442): removes banned + already-stored rows from existing tables.

## telegram_lock.py (`gui/telegram_lock.py:`)

- `TelegramLockManager`: `acquire(owner)` (re-acquire by the CURRENT owner refused too — the
  nested-acquire bug), `release(owner)` (owner=None unconditional; owner-scoped otherwise),
  `force_release(reason)`, `busy/owner/age_seconds/describe` (`'bot_check (running 84s)'`),
  `is_stuck(max_age)`, context manager `_Held` (releases only if acquired; never swallows).
- Watchdog threshold in the GUI: **35 min** (longer than the subprocess runner's 30-min hard
  cap); busy message with >300 s stuck hint.

## Go notes

- The lock ports as a mutex + owner/age fields with identical semantics (owner-scoped release,
  stuck detection, force-release reason logging).
- Intake tables: exact same grammar and filter laws.

## Tests

Platform classification → table file mapping; banned-never-collected; no-empty-file law; prune
matrix; token scrubbing in rows; lock: owner scope, re-acquire refusal, stuck detection, context
manager release law.
