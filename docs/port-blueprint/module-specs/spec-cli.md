# Module Spec — `gitcurator/cli/` (+ top-level shims + launchers)

> **RE-SCOPED (2026-10-15, ADR-0004 Am.3 — owner stamp "OPTION B CONFIRMED", rider R-0):**
> the Go app drops its direct Telegram (MTProto) leg; the bot queue rides the Cloudflare Worker
> HTTP API only. CLI ripples, all recorded below as dated notes at their contracts:
> (a) `--sync`'s CHECK and `--test-connection`'s live leg lose the Telegram part — the queue
> check rides the Worker only; (b) the Telethon-only modes (`--mark-read`, `--from-id/--to-id`,
> `--offset-start/--count`, `--single-id`, `--init`'s interactive Telegram login offer) are
> accepted losses (ADR-0004 Am.3) and do not function in the Go port — T-60 decides the honest
> flag surface; (c) the batch-end "mark the queue read" step falls away — mirror-push is the
> ack (spec-worker-contract §8.4). Everything describing the Telegram legs below is retained
> verbatim as the Python-reference record.

- **LOC**: cli.py 1,777 + cli_run.py 719 + cli_sync.py 381 + cli_settings.py 211 +
  cli_terminal.py 321 + shims 33 · **Go home**: `internal/cli` + `cmd/gitcurator-cli` ·
  **Status**: port in progress (Telegram legs re-scoped per the banner above;
  SESSION-027 landed T-60's first slice — THE CLI SHELL: the paint box, the
  settings halves, the sync door's pure halves, the honest flag surface +
  dispatch, the FULL --status card, --list-dead/--reset-dead; the wizard,
  --test-connection, --detect-llm, --hand-delivery, --chrome-retry and the
  batch engines follow).
- **THE HONEST FLAG SURFACE (T-60's decision, landed SESSION-027)**: the
  Telethon-only flags (`--login`, `--mark-read`, `--from-id/--to-id`,
  `--offset-start/--count`, `--single-id`) stay RECOGNIZED on the parser and
  answer explicit refusal cards citing ADR-0004 Am.3 (exit 1) — never a
  silent no-op, never an unknown-flag crash; they are dropped from the
  teaching surfaces (the friendly card and --help). Commands whose engines
  land later answer explicit in-development cards naming the task (the rest
  of T-60 / T-62), exit 1.
- **Responsibility**: the visual terminal client — full command surface, wizard, status/detect
  commands, `--auto` and manual batch modes — running the exact same pipeline via
  `QCoreApplication` (in Go: the same `internal/app` commands, no Qt anywhere).

## Command surface (`build_parser` cli.py:1634-1706)

`--init` · `--login` · `--auto` · `--sync` · `--list-dead` · `--reset-dead` ·
`--hand-delivery` · `--chrome-retry` · `--status` · `--test-connection` ·
`--detect-llm {ollama,llamacpp}` · `--retry-failed` · `--mark-read` ·
`--from-id/--to-id` · `--offset-start/--count` · `--import-file F` · `--single-id N` ·
`--vault PATH` · `--config PATH` · `--strikes N` · `--yes/-y` · `--dry-run` · `--no-color`.
Dispatch order (1718-1749): init → login → status → test-connection → detect-llm → list-dead →
reset-dead → hand-delivery → chrome-retry → mark-read → auto → sync → retry-failed → manual
batch; no flag → friendly help card.

## Key command contracts

- **`--init`** (108-299): wizard writes config.json (vault, Telegram api_id/hash/phone,
  bot_username, github_token, website/manual vault paths, LLM provider ollama|cloud|llamacpp +
  endpoint/model, proxy, 404 threshold 2-10); **runs the repo radar BEFORE saving** (160-185);
  offers interactive Telegram login at the end when stdin is a TTY
  — **DROPPED in the Go port (2026-10-15, ADR-0004 Am.3 / R-0):** no MTProto client ships, so
  no login is ever offered (accepted loss); the wizard still *collects* the Telegram keys for
  compatibility (the compat law) — the engine never uses them.
- **`--status`** (423-678): config summary rows (vault map ✓/◌/✗, repo slots, pipelines, masked
  secrets), sentinel probes, CacheDB/NoteStateDB/WebsiteStateDB stats. **REALITY CORRECTION
  (SESSION-027, found at implementation time)**: "never creates state" is true only for the
  note-state and websites blocks (they open the db file only when it already exists) — the
  CacheDB block calls `CacheDB()` with its default path UNCONDITIONALLY, and the reference's
  CacheDB constructor CREATES the file + tables when missing (the v0.09.4 APP_DIR-anchored
  default): a `--status` run on a fresh install does create APP_DIR/cache.db. The port
  preserves the behavior (the store's NewCacheDB over the same path).
- **`--test-connection`**: the battery + one LIVE Telegram probe (bot queue if bot configured,
  else account preview; non-interactive leg)
  — **re-scoped in the Go port (2026-10-15, ADR-0004 Am.3 / R-0):** the live Telegram probe is
  dropped; `--test-connection` runs the battery only, whose Worker leg (GET `/health`, proxy
  bypassed — spec-core-connection) is the queue-reachability check. The Telethon-only flags
  (`--mark-read`, `--from-id/--to-id`, `--offset-start/--count`, `--single-id` — the
  Saved-Messages/keyword/single-id/preview modes) are accepted losses and do not function;
  T-60 decides the honest flag surface (drop, or honest "unavailable" refusal), never a
  silent no-op.
- **`--auto`** (1252-1440): preflight → fetch with `min_id = last_processed_msg_id` →
  classification printout → **wrong-vault guard** (pending>0, in_vault=0, index=0, cache>0 →
  stop — 1374-1409) → optional >10 confirm (skipped by `--yes`) → `run_batch_visual`.
- **`--sync`** (cli_sync.py): CHECK (`sync_missing_essentials` pure + live battery with a
  non-interactive telegram leg) → THE BUTTON (Enter = sync, anything else quits; `--yes`/non-TTY
  never asks — the cron law) → SYNC (the `--auto` engine's two phases; wrong-vault guard
  non-interactive; GUI_INSTRUCTION card on every failure: never asks for a credential, never
  prompts a code, never opens the model menu)
  — **re-scoped in the Go port (2026-10-15, ADR-0004 Am.3 / R-0):** the queue check rides the
  Worker only — CHECK = `sync_missing_essentials` (pure) + the battery's Worker leg
  (non-interactive); no Telegram leg exists. The queue's data source is the Worker's
  `/api/pending` + the existing `get_pending_result` path (T-34 scope, unchanged).
- **`--detect-llm`**: probe → model menu (auto-pick with `--yes`) → set provider/model/URL →
  save. `--hand-delivery` / `--chrome-retry`: queue walled links / auto-scrape via a throwaway
  profile. `--list-dead`/`--reset-dead`: 404 quarantine. `--mark-read`.

## cli_run.py — the visual run engine

- `_LogShim`/`_code_prompt` (fetch-thread prompts under `_PRINT_LOCK`); `fetch_bot_queue`
  (the `_bot_queue_job` on a thread with law-split args); `_mark_bot_queue_read`
  — **Dated note (2026-10-15, ADR-0004 Am.3 / R-0):** in the Go port the fetch-thread code
  prompts fall away with the MTProto login (no code is ever prompted); the queue fetch becomes
  the Worker `/api/pending` call; `_mark_bot_queue_read` is dropped (read-acknowledge is an
  accepted loss — mirror-push is the ack, spec-worker-contract §8.4).
- `run_batch_visual` (239-504): wires the 8 worker signals (318-341); on finish: GitHub seal →
  GoodRepos publish → websites seal → manual seal (383-417) → **Phase-5 CLEAR** — only when
  every link verified (`tracker.get_all_clear()`) does it persist `last_processed_msg_id` to
  config.json and mark the queue read (425-469)
  — **Dated note (2026-10-15, ADR-0004 Am.3 / R-0):** the msg-id cursor and the queue mark-read
  step fall away with the Telethon fetch (the Worker queue is ledger-state-based, not
  message-id-based; mirror-push is the ack). The all-clear gate itself survives as the
  honest-completion law; its exact Phase-5 re-spec (what the all-clear now gates) is an open
  item for T-62 (Phase 5), bounded by ADR-0004 Am.3 — do not improvise it before then.
  dry-run skips all seals and writes the report
  (`app/reports/dry-runs/dry_run_<stamp>.md`, 215-236); **SIGINT handler + 100 ms QTimer
  heartbeat so Ctrl+C works inside the event loop** (477-494); exit code = batch success flag.
- Preflights: `_preflight_github` (PyGithub `get_user().login`), `_preflight_proxy` (TCP 5 s),
  `_preflight_llm` (probe + interactive model picker).

## cli_settings.py

`load_config` (user-config fill) / `save_config` (write + mirror) / `_mask_token` /
`_clean_config_for_save` (strips `__…__` keys) / `prompt_model_menu` (numbered, recommended
default, `q` abort, non-TTY auto-picks, 5 tries) / `apply_model_choice` (provider-aware).

## cli_terminal.py

- `C` ANSI class: `enable()` honors NO_COLOR / TTY / colorama `just_fix_windows_console()`
  fallback `os.system("")` VT trick; `paint`.
  **GO NOTE (the dead-level-color quirk, pinned by the golden)**: the reference builds
  `_LEVEL_STYLE` at MODULE IMPORT while C's codes are still the pre-enable empty strings —
  `enable()` mutates the class attributes but never the already-built dict, so the level
  COLORS are dead code in the reference (cli_print's icons and level text stay plain even
  with colors on; only the timestamp dims). The port freezes the same empty level colors.
- **`_harden_stdio` (runs at import)**: stdout/stderr → UTF-8 `errors="replace"`;
  Windows `SetConsoleOutputCP(65001)` + `SetConsoleCP(65001)` (the "CLI is not running"
  UnicodeEncodeError fix).
- `_gui_symbol` import hardening; `term_width`; `mask`; `rule`; `cli_print` (timestamped,
  thread-safe); `banner`; braille spinner frames `⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏`; `render_bar`;
  `StatusLine` (owns the bottom line, `\r` repaint, erase-then-log).

## Shims + launchers

- Top-level `cli_run.py`/`cli_settings.py`/`cli_terminal.py` = re-export shims (import-path
  law) — unnecessary in Go.
- `cli/__init__.py` re-exports the old surface.
- **Launchers** (`.bat`, ASCII-only, never `chcp`): `setlocal` → title → `cd /d "%~dp0"` →
  **walk-up** `if not exist "main.py" cd /d "%~dp0\..\.."` → venv-first python discovery
  (`.venv\Scripts\python.exe` → `python` → `py -3`, with the "Double-click 1-INSTALL.bat"
  failure card) → run (`GitCurator-CLI.bat` → `--cli --auto --yes`; `GitCurator-SYNC.bat` →
  `--cli --sync %*`; Setup → `--init`; Start → alias) → EXITCODE echo + `pause`.
  `gitcurator-cli.sh`: `set -euo pipefail`, walk-up, `python3`, default `--cli --auto --yes`.
- **Go equivalents**: launchers run `GitCurator-cli.exe` directly (no python discovery); keep
  the walk-up cd only if assets are expected relative to the exe (they are embedded — the
  walk-up law becomes unnecessary but harmless); ASCII/CRLF laws still apply.

## Tests

Flag-surface snapshot (help text); wizard flow (fake stdin); status card rendering;
`--sync` door beats (Enter/other/`--yes`/non-TTY); wrong-vault guard; Phase-5 gating of msg-id
persist + mark-read; dry-run report equality (differential case); Ctrl+C exit code; terminal
paint (ANSI/NO_COLOR); UTF-8 hardening on Windows.

---

## SESSION-027 shipped record — T-60's first slice: THE CLI SHELL

Landed (commit at the session close): `internal/cli` grew from the Phase-0
seed into the shell — `terminal.go` (the paint box: the C ANSI system with
the NO_COLOR/TTY/--no-color law, level styles, `term_width`'s COLUMNS law,
`mask`/`_mask_token`, `rule`, `cli_print` over the frozen-clock seam,
`banner`, `render_bar` with banker's rounding, `StatusLine`, `_strip_ansi`,
the Windows 65001 console hardening via syscall), `settings.go`
(`_config_path`/`load_config`/`save_config` over the T-12 config port, `_ask`
with the EOF law, `_mask_token`, `_clean_config_for_save`),
`syncdoor.go` (GUI_INSTRUCTION, `sync_failure_card`,
`sync_missing_essentials` with the isdir seam, `sync_confirm_prompt` with
the cron law) + cmd_sync's BEAT 1a wired, `cli.go` (the honest Option-B
flag surface + the dispatch order + the friendly card + `--help` + the
refusal/not-yet cards), `status.go` (the FULL card: 16 config rows, the
taxonomy row, the sentinel rows over the connectioncheck port's seams, the
Ollama/llama.cpp model checks, the CacheDB/NoteStateDB/WebsiteStateDB
stats), `dead.go` (--list-dead/--reset-dead).

The differential stone: `testdiff/goldens/dump_cli.py` (the generator —
drives the reference's own CLI shapes with the house seams: the gui.icons
stub unblocking the stdlib import chain, the gui.app stub exposing the REAL
CacheDB/dead_link_threshold, the net_guard stub, the fake ollama module,
APP_DIR patched before the cache_db import, the frozen clock, both temp
roots folded) → `internal/cli/testdata/goldencli.golden.json` (two-run
byte-identical) → `goldencli_test.go` (the replay: banner, paint shapes
with colors on AND off, cli_print lines, the essentials matrix, the button
beats, the failure card, the full status card over the rebuilt crafted env
(config + cache.db through the ported store's own methods + the pinned
quarantine rows), the no-config paths, list-dead, reset-dead; the mutation
test bites; the reference's no-flag help card kept as the pre-Option-B
RECORD, not byte-compared — the divergence is asserted, documented, and
pinned by the Go-side snapshots).

Remaining on T-60's books (SESSION-028+): the `--init` wizard (the repo
radar is ported; the llamacpp probe rides the seam), `--test-connection`
wiring (the battery port exists — `connectioncheck.RunLocalChecks` +
`Summarize`), `--detect-llm`, `--hand-delivery`, `--chrome-retry` (the
engines are ported; the CLI wiring follows), `smart_model_pick`/
`prompt_model_menu`/`apply_model_choice` (cli_settings's model-menu half —
the batch engine's model-picker bridge, T-62's books), then T-62 (--sync's
live beats + --auto + the dry-run report writer), T-63 (launchers), T-64
(the full differential corpus run), T-65 (the release tag).
