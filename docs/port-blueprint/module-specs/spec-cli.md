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
  dispatch, the FULL --status card, --list-dead/--reset-dead; SESSION-028
  landed THE COMMAND WIRINGS: the `--init` wizard with the repo radar, the
  `--test-connection` battery wiring with the _section/_check_line renderers,
  `--detect-llm` with the model menu + set-and-save, cli_settings's
  model-menu half, `--hand-delivery` complete, and `--chrome-retry`'s
  listing + dry-run rehearsal (the live leg is T-62's batch engine).
  Remaining on T-60's books: only T-61's paint legs (the spinner under a
  real loop; the Windows console live test is Phase 7's clean-VM
  checklist).)
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
  **WIRED (SESSION-028)**: `cmdInit` ports the whole _ask sequence (the YOUR_ placeholder
  empty-start law, the isdir retry loop, the api-id digit law, the provider branches with the
  probe seam offering the detected model, the proxy block with the n-branch raw-keep law, the
  2–10 clamp), the radar wires (RadarPending/ApplyDetections/ReportLines run REAL over the
  `RadarAll` seam — the radar legs ride seams), save_config, and the saved box (the reference's
  `.ljust(64)`-on-the-painted-string quirk preserved). The closing card speaks the honest
  Option-B truth (no login line; `gitcurator-cli --auto`); the golden keeps the reference's
  non-interactive closing card as the pre-Option-B RECORD with the shared prefix (everything
  through the box) byte-compared and the divergence asserted.
  **GO NOTES**: the reference's `int()` edges (a non-decimal isdigit answer, a non-integer
  proxy port) crash the whole wizard — the port answers 0 / the or-door default instead
  (the happy paths are byte-identical); underscore integer literals and arbitrary-precision
  answers are likewise not mirrored (int64).
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
  **WIRED (SESSION-028)**: `cmdTestConnection` runs `connectioncheck.RunLocalChecks` +
  `Summarize` over the loaded config with cli_run.py's own `_section`/`_check_line`
  renderers ported into the paint box; the exit-code law (error → 1, warn → 0) holds.
  **GO NOTES (the Option-B divergences, all asserted by the golden — never silent):**
  the card's subtitle says "the sync queue (the Worker)" where the reference says
  "Telegram (live)"; the fourth section is titled "Worker" (its one row the /health
  verdict) where the reference's is "Telegram" (the local rows + the live line) — the
  reference's card is kept in the golden as the pre-Option-B RECORD with the shared
  sections [1/4]..[3/4] byte-compared; the verdict's ready/warning counts differ when
  the dropped telegram rows carried levels (the level + the exit code still match over
  the same crafted env); the battery's net doors ride the exported
  `connectioncheck.HTTPGetJSONFn` / `GitHubGetFn` seams (the reference's own monkeypatch
  surface — `_http_get_json` / `_github_get`), and the Worker /health verdict
  differential drives `check_worker_version`'s own injectable `http_get` over four
  crafted cases (match / older / unreachable / no-URL).
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
  **WIRED (SESSION-028)**: `--detect-llm` is complete — `runDetectJob` mirrors
  worker_jobs._quick_detect_job's composition synchronously over the CLI's probe seams
  (DetectOllama / ProbeLlama / DetectLlama; the reference rides a GUI thread + StatusLine
  spinner — the paint box is already ported, the log lines ride cli_print exactly as the
  drain loop printed them), the numbered menu (the only-one line, the ← current mark, the
  --yes auto-pick, the invalid-choice abort), set + save through `_clean_config_for_save`, and
  the SET card. cli_settings's model-menu half landed with it: `SmartModelPick` (the
  githubpipeline.PickBestModel twin — same heuristic as the pipeline, the recommendation and
  the headless auto-pick always agree), `PromptModelMenu` (the numbered picker with the
  recommendation as entry 1, the embedding tag override, q-decline, the five-try give-up,
  the non-TTY auto-pick), `ApplyModelChoice` (the provider-aware write+save). `--hand-delivery`
  is complete (the walled-rows listing over the StateDB adapter — the decomtable law keeps
  the concrete adapter at the call site —, the enqueue with queue.json + README.txt, the
  open-Chrome ladder, the folder card). `--chrome-retry` wires the listing (retry rows +
  the 🖐 hand rows + the redo set) and the FULL dry-run rehearsal; the live leg (preflight →
  the real Chrome delivery → `run_batch_visual` processing the delivered pages in the same
  run) waits for T-62's batch engine — an honest boundary card naming T-62, never a silent
  no-op (splitting the deliver from the process would diverge from the reference's one-run
  contract).
  **REALITY-WINS FIND (SESSION-028, fixed in-commit)**: the reference's `cmd_chrome_retry
  --dry-run` leg calls `deliver_pages_via_chrome` WITHOUT flipping the global dry-run switch
  (only the batch worker ever flips it) — the pinned rev's dry-run would launch the real
  Chrome while the card says "Chrome is never launched". The port flips the switch around
  the delivery (the card's own words + the deliver's rehearsal law are the contract) and
  restores + clears it right after; the golden's reference side stubs the deliver (the
  rehearsal's engine — its bytes are already pinned by T-44's chrometabs golden) so both
  sides answer the same card.

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

## SESSION-028 shipped record — T-60's remainder: THE COMMAND WIRINGS

Landed: `internal/cli/wizard.go` (cmd_init — the full _ask sequence with
the YOUR_ empty-start law + the isdir retry loop + the api-id digit law
+ the provider branches with the llamacpp probe seam + the proxy block
with the n-branch raw-keep law + the 2–10 clamp + the radar wires over
the RadarAll seam + save_config + the saved box with the .ljust(64)
painted-string quirk + the honest Option-B closing card),
`testconn.go` (cmd_test_connection — the battery over
connectioncheck.RunLocalChecks + Summarize with cli_run.py's
_section/_check_line renderers; the Worker leg IS the queue check; the
Option-B divergences asserted), `detectllm.go` (cmd_detect_llm — the
synchronous _quick_detect_job composition over the probe seams, the
numbered menu, set + save, the SET card), settings.go's model-menu half
(SmartModelPick/PromptModelMenu/ApplyModelChoice — the preflight's
model-picker bridge for T-62), `hands.go` (cmd_hand_delivery complete —
the walled listing over the StateDB adapter + the enqueue + the folder
card; cmd_chrome_retry — the listing + the full dry-run rehearsal + the
honest T-62 boundary card for the live leg), and the dispatch updates
(the not-yet cards retire for every implemented command).

Engine-package touches (additive, no behavior change):
connectioncheck's `_http_get_json`/`_github_get` mock surface exported
as HTTPGetJSONFn/GitHubGetFn (+ GitHubReply — the reference's own
monkeypatch surface, now injectable cross-package); githubpipeline's
`_is_embed_model` exported as IsEmbedModel; handdelivery's open-ladder
seams exported (FindChromeFn/SubprocessPopenFn/WebbrowserOpenFn).

The differential stone grew: dump_cli.py's payload now drives 9 more
sections — the battery card (the stubbed net doors + the live-telegram
stub; the RECORD) + the structured rows + the worker /health cases (the
reference's own http_get injection), the two detect cards with their
SAVED CONFIG bytes, the model-menu matrix (the REAL WorkerLlmMixin
picker behind the fake PyQt6/PyGithub guards — the reference's own
class, its import doors stubbed), the two wizard runs (scripted stdin +
the stubbed radar + the probe seams; the SAVED CONFIG bytes are the
wizard's own writes), the hand-delivery card + queue.json + README.txt
(frozen hd.time, the open ladder's no-browser answers), and the
chrome-retry dry-run card — 48,698 bytes, two-run byte-identical. The
replay byte-compares: the battery's shared sections [1/4]..[3/4] + the
worker rows + both detect cards + their saved configs + the model-menu
matrix + both wizard shared prefixes + both wizard saved configs + the
hand-delivery card + queue.json + README.txt + the chrome-retry card;
every Option-B divergence (the battery's subtitle/section-4/verdict
counts, the wizard's closing card) is asserted, never silent; the
mutation teeth grew with TestGoldenCLIMutationWirings (the wizard's
saved config, the enqueue's queue.json, the detect card — each bites).

Standing Go notes recorded above: the wizard's int() crash edges are
honestly hardened; the chrome-retry dry-run switch is the reality-wins
fix (the reference forgot to flip it); the detect job runs
synchronously (no GUI thread + spinner — the log lines are identical).

Remaining on the CLI books: T-61's paint legs (the spinner under a
real loop; the Windows console live test is Phase 7's clean-VM
checklist), then T-62 (the --sync live beats + the --auto engine + the
dry-run report writer + the SIGINT/exit-code law + the wrong-vault
guard + the chrome-retry live leg + the preflight with the
PromptModelMenu bridge), T-63 (launchers), T-64 (the full differential
corpus run), T-65 (the release tag).
