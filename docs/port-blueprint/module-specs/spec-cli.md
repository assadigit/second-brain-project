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
  listing + dry-run rehearsal (the live leg is T-62's batch engine);
  SESSION-029 landed T-62's first slice — THE RUN ENGINE'S DOORS: the
  preflights + the run-config card + the dry-run report writer + THE
  PHASE-5 CLEAR re-spec resolved on evidence and implemented.
  Remaining on the CLI books: T-62's remainder (the engine shell +
  the batch drivers + the SIGINT/heartbeat/exit law) + T-61's paint
  legs (the spinner under a real loop; the Windows console live test
  is Phase 7's clean-VM checklist).)
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
  **DATED CORRECTION (2026-10-21, SESSION-058 — reality wins)**: the live leg landed HERE.
  The T-62-era records had claimed it landed at SESSION-030 (the 6ccf184 commit message
  names internal/cli/hands.go; the commit's own stat never touched the file) — the honest
  boundary card and its pinning test stood until SESSION-058 landed the real leg. It
  mirrors cli.py 1204-1245 exactly: the preflight BEFORE the `__config_path__` stamp (the
  reference's own ordering quirk — cmd_auto stamps first, cmd_chrome_retry stamps after),
  the delivery through the chromeDeliverDoor seam (`ChromeDeliverFn` — the reference's
  module-function patch surface; the dry-run rehearsal routes through the same door), the
  delivered/redo cards, THE ROUTING LAW (the delivered pages + the redo set ride
  `NonGithub`, never the GitHub urls — the v0.57.0 fix), the batch through the
  runBatchDoor seam (`RunBatchFn`; nil → the real RunBatchVisual), the first-three-errors
  [:90]-runes composition, and the exit codes (0 = the batch's verdict; 1 = nothing
  delivered + nothing to redo). Proven three ways: the four chromeretry_test.go selections
  (the port of test_noteredo's TestCLIRouting + the post-batch ⚠️ warning), the
  differential golden's two new sections (`chrome_retry_live` + `chrome_retry_live_nothing`
  — the reference's own cli_main runs byte-compared, the golden at 110,503 bytes two-run
  byte-identical), and the mutation teeth (TestGoldenCLIMutationChromeLive).
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
  — **RE-SPEC RESOLVED (2026-10-18, SESSION-029, T-62 — the bounded open item of ADR-0004
  Am.3, closed on evidence, never improvised):** the msg-id cursor and the queue mark-read
  fall away with the Telethon fetch (the Worker queue is ledger-state-based, not
  message-id-based). What survives, on the evidence chain (ADR-0004 Am.3: "read-acknowledge
  is dropped; mirror-push is the read acknowledgement"; spec-worker-contract §8.4:
  "Mirror-push is the ack (idempotent by sync_id) — /api/vault_mirror/push remains the one
  and only way items leave the pending queue"):
  1. **The all-clear gate survives as the honest-completion law** — only a batch whose every
     GitHub link verified (`tracker.GetAllClear()`) may speak the Phase-5 CLEAR verdict;
     a not-verified batch speaks the ⏸️ verdict (the links stay PENDING, not acknowledged).
  2. **The queue-consumption mechanics are per-link by LEDGER STATE** — a verified link has
     a note → enters the vault index → the mirror push tells the Worker `in_vault` → it
     leaves `/api/pending`; an unverified link has no note → stays pending → the next
     `--auto` re-offers it. The anti-repeat law is preserved by the ledger itself, not by a
     message cursor (a strictly finer-grained honesty than the reference's all-or-nothing
     msg-id advance).
  3. **The mirror push rides the batch-complete moment** — the reference's own intended
     pattern: `CloudflareManager.on_batch_complete` (the `processing_finished` handler's
     end, per the integration snippet's law) schedules the debounced `_do_batch_push` after
     EVERY batch. REALITY-WINS RECORD: the pinned rev's MainWindow keeps `cf_manager = None`
     — the manager ships as a library plus the documented integration pattern, never
     live-wired, and `build_vault_mirror_entries` expects the `{url: {path, category,
     title}}` dict the app's VaultIndex class never was (`dict(instance)` raises TypeError
     inside the manager's own try/except). The port implements the pattern directly:
     `cli.PhaseFiveClear` + `cli.PushVaultMirrorAfterBatch` build the entries from the
     vault index's url→path map (`batch.VaultIndex.MirrorIndex()` — sorted-url order, the
     documented deterministic-wire divergence), the decommission events from the cache, the
     sha256 index hash over the sorted urls joined `|` (the manager's own recipe), and push
     through `workerclient.Sync.PushVaultMirror` (idempotent by sync_id). Best-effort and
     failure-isolated: a down Worker logs one line and never fails the batch; a disabled
     Worker is silent (not an event — the reference's own `is_enabled()` early-return).
  dry-run skips all seals and writes the report
  (`app/reports/dry-runs/dry_run_<stamp>.md`, 215-236) — **the report writer landed
  (SESSION-029): `cli.writeDryrunReport` over the ported `dryrun.RenderReportMarkdown`,
  the frozen-clock seam `dryrun.NowFn` added for the golden**; **SIGINT handler + 100 ms
  QTimer heartbeat so Ctrl+C works inside the event loop** (477-494) — the synchronous Go
  composition's twin (signal.Notify + the ticker driving `StatusLine.Tick`) lands with the
  engine slice (T-62 remainder); exit code = batch success flag (the engine's law).
- Preflights: `_preflight_github` (PyGithub `get_user().login`), `_preflight_proxy` (TCP 5 s),
  `_preflight_llm` (probe + interactive model picker).
  **WIRED (SESSION-029, T-62's doors)**: `cli.PreflightChecks` + the three preflights are
  ported with the model-picker bridge wired (`PromptModelMenu`/`SmartModelPick`/
  `ApplyModelChoice` — landed SESSION-028 — now driven by the preflight's flow: the
  not-pulled Ollama model opens the menu, the pick persists through the same save path).
  `run_config_card` (524-562) is ported (`cli.RunConfigCard`, the raw-comparison laws: the
  provider branches compare `llm_provider` RAW against "ollama"/"llamacpp" — no strip, no
  lower; `cloud_model`/proxy host+port render through `.get(key, default)` — a present
  null renders "None", an absent one the default; the token mask counts RUNES). The seams:
  `CLI.PreflightGithub` (the PyGithub leg; the default door rides the connectioncheck
  GitHub seam — the same GET /user the battery's check makes), `CLI.NetDial`
  (`socket.create_connection`; the class-name token rides the seam exactly like the worker
  check's OSError stand-in — the SHAPE is the contract), and the existing `CLI.OllamaModels`
  (an error's string is the class-name token for the unreachable line). The GitHub-error
  door asks "continue anyway? (y/N)" — only a TTY ever asks (the cron law).

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
honest T-62 boundary card for the live leg [superseded at SESSION-058:
the live leg is real — the dated correction above carries the record]),
and the dispatch updates
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

## SESSION-029 shipped record — T-62 slice 1: THE RUN ENGINE'S DOORS

Landed: `internal/cli/runconfig.go` (run_config_card — the boxed
run-configuration snapshot with the raw-comparison laws: the provider
branches compare RAW, `cloud_model`/proxy host+port render through
`.get(key, default)` so a present null renders "None", the token mask
counts runes), `internal/cli/preflight.go` (preflight_checks + the
three preflights — the PyGithub token check over the new
`CLI.PreflightGithub` seam whose default door rides the connectioncheck
GitHub seam, the TCP proxy reach over the new `CLI.NetDial` seam with
the class-name token, the provider branches with THE model-picker
bridge wired: the not-pulled Ollama model opens the landed
PromptModelMenu and the pick persists through ApplyModelChoice; the
llamacpp auto-fill + save; the cloud line; the GitHub-error
continue-anyway door with the TTY-only cron law), and
`internal/cli/phase5.go` (THE PHASE-5 CLEAR re-spec implemented — the
all-clear verdict speech + `PushVaultMirrorAfterBatch`, the
batch-complete mirror push: the entries from
`batch.VaultIndex.MirrorIndex()` (the new additive export, sorted-url
order), the decommission events from the cache, the sha256 index hash
over the sorted urls joined `|` — the manager's own recipe over
`workerclient.Sync.PushVaultMirror`; plus `writeDryrunReport` — the
report writer over the ported `dryrun.RenderReportMarkdown`).

Engine-package touches (additive, no behavior change):
`dryrun.NowFn` (the frozen-clock seam — the reference's generator
freezes the module's datetime; both stamp sites now ride the seam) and
`batch.VaultIndex.MirrorIndex()` (the url→path export in the mirror
entries' shape).

THE DIFFERENTIAL STONE GREW: dump_cli.py drives 10 more sections — the
four run-config cards, the four _preflight_github cases (the fake
PyGithub door swapped per case — ok / 401-rejected /
GatewayTimeout-unreachable / no-token), the four _preflight_proxy
cases (the fake socket.create_connection + the deterministic 42-ms
fake clock — disabled / reachable / string-port / TimeoutError), the
eight _preflight_llm cases (ollama ready / unreachable / empty / the
menu-pick WITH ITS SAVED CONFIG bytes / llamacpp ready / auto-fill WITH
ITS SAVED CONFIG / not-found / cloud), the preflight_checks
composition (the ok card + the github-rejected abort), and the
dry-run report (the reference's own record() over the frozen clock +
the written file's path + body) — goldencli.golden.json at 69,028
bytes, two-run byte-identical. The replay byte-compares every new
section (the saved configs included); the FIXED-LENGTH ROOTS law:
the run-config card's right-side padding bakes the path lengths into
the golden's bytes, so buildEnv now uses roots whose lengths equal the
generator's mkdtemp roots ("goreplay" = exactly the 8-rune suffix
width). TWO GENERATOR BUGS THE DIFFERENTIAL ITSELF CAUGHT: the
shallow-copy leak (apply_model_choice wrote through the shared nested
dicts — the recorded config must snapshot BEFORE the call) and the
ms-clock arithmetic (the fake answers must be 1000.0 → 1000.042 for
the reference's int((t1-t0)*1000) to render 42). The mutation teeth
grew with TestGoldenCLIMutationPreflights (the checks card, the LLM
saved config, the dry-run report body — each bites).

1877 tests race-clean over 31 packages (+11: the 10 preflight/phase5
unit tests + the new mutation test; the arithmetic closes exactly).

Remaining on T-62's books (SESSION-030+): the engine itself —
run_batch_visual's synchronous shell (the setup: cache + shadow,
note-state, dead/settled/dismissed, the vault index rebuild, the
LinkTracker intake, the repo-banish pass; the warmup family: the
Ollama warmup + autostart + the model_prompt_callback bridge, the
llamacpp re-probe, the cloud /models pre-flight; the website-phase
wiring over the ported Pipeline; the Finish chain; the seals + the
Phase-5 CLEAR + the dry-run path + the SIGINT/heartbeat + the exit
law) + the drivers (cmd_auto + cmd_sync's live beats over the Worker
/api/pending fetch + the _bot_queue_job classification half, the
wrong-vault guard, the >10 confirm, cmd_retry_failed, the import-file
mode, the chrome-retry live leg) + T-61's spinner under a real loop.

## SESSION-030 shipped record — T-62 slice 2: THE ENGINE + THE DRIVERS

Landed: `internal/cli/engine.go` (run_batch_visual as the synchronous
Go shell — the mode dispatch incl. the import grammar + the
review/master backlog modes, the setup family [the LinkTracker intake
with the banned/self drops + the per-platform _inbox tables over the
new `internal/platformintake` port, the cache open with the dry-run
shadow, note-state, the dead set, THE REPOS BANISHMENT pass, the
settled ledger, the vault index rebuild + SetVaultIndex, the
missing-note backfill, the §4.4 start-of-run for BOTH vaults, the
dismissed set, the batch snapshot], the warmup family [the Ollama
list/timeout + the detached autostart ladder with the platform split
(setsid / DETACHED), the warmup + the auto-switch/stand-in ladder over
the model_prompt_callback bridge, the llama.cpp re-probe + the model
resolution + save, the cloud /models pre-flight with the CloudLLMError
info-line law], the GitHub loop over the ported Runner (started via the
new `Start()` seam — the is_running __init__ law), the websites phase
per website_phase.py's exact wiring (the intake drops, the shadow
state, SettleTheLedger, the websites VaultIndex, the LLM router with
the per-task overrides, the Worker-mediated banish channel ported over
the propose/poll/result trio, SplitDismissedLinks, the bar re-base,
the backlog modes, the manifest marking, the Website Directory
rebuild, the _inbox prune, the done tally), the close + the Finish
chain, the finish seals in the reference's exact order, THE PHASE-5
CLEAR, the dry-run path, the SIGINT drain + the 100 ms heartbeat
ticker driving StatusLine.Tick (T-61's spinner under a real loop; the
StatusLine grew the paint mutex — Qt's signal-marshaling twin), and
the exit-code law); `internal/cli/fetchqueue.go` (the Option-B fetch —
the Worker /api/pending honest poll — + the _bot_queue_job
CLASSIFICATION half as its own coherent piece: the URL routing, the
vault filter with THE SETTLEMENT at the queue's own door, the
dead/decommissioned buckets, and the v0.24.1 websites ladder);
`internal/cli/autocmd.go` (cmd_auto with the wrong-vault guard +
the >10 confirm + the queue card + the dedup line, cmd_retry_failed,
the import-file mode); `internal/cli/syncbeats.go` (cmd_sync's live
beats 1b/2/3 — the battery's Worker leg IS the queue check, the
non-interactive wrong-vault guard answers the failure card); the
chrome-retry LIVE leg in hands.go (preflight → the real
DeliverPagesViaChrome → run_batch_visual with the pages as non_github);
and `internal/platformintake/` (gui/platform_intake.py's port: the
per-platform _inbox table writers + the pruner + the table-vault law,
the Clock seam).

Additive engine-package touches: Runner.Start/Position/SetTotal/
AdvancePosition/SetBatchFiles/SetWebsiteSummary;
websitepipeline.Pipeline.Taxonomy(); StatusLine's paint mutex +
EraseIfActive + the PromptModelMenu StatusPainter interface; the CLI's
engine doors (NewStatusLineFn / GHClientFactory / OllamaFactory /
AutostartServe / WorkerFetch / BannerDownload).

THE DIFFERENTIAL GREW: dump_cli.py drives four more sections —
auto_guard (the wrong-vault card over the monkeypatched Telethon fetch
+ scripted stdin "n"), auto_guard_yes (--yes never re-processes),
auto_confirm (the >10 TTY confirm with 11 links over an indexed vault),
and engine_dryrun (THE ENGINE'S OWN COMPOSITION: the reference's
run_batch_visual over a fake Qt event loop [pyqtSignal as callback
lists, QThread running inline, QCoreApplication.exec returning 0 — the
synchronous Go shell's exact twin], the fake PyGithub repo surface +
the fake ollama chat, one github link, dry-run — capturing the 29-line
structured log sequence, the BATCH DONE card, the exit code, and the
12-entry dry-run report body) → goldencli.golden.json at 86,570 bytes,
two-run byte-identical. The replay byte-compares every section; the
mutation teeth grew (TestGoldenCLIMutationDrivers).

REALITY-WINS FINDS the differential forced (fixed in-commit):
1. linktracker's verify report logs "GitHub processed: N" with level
   success when N > 0 (the reference's conditional — the port had
   always-info).
2. The engine's §4.4 start-of-run must pass the DB PATH even with no
   connection (the reference's run_start_check derives it when db is
   None; the dry-run detect branch needs it).
3. The batch's closing hygiene pass runs run_vault_hygiene over the
   GITHUB vault ONLY (the reference's _vault_hygiene_pass — NOT
   heal_machine_vaults, the startup heal's both-vaults door).
4. NewWebsiteStateDB("") is NOT the default path — the literal
   "cache.db" is (AppDir-anchored); "" opened a stray pragma-named
   file (caught by the classification tests).
5. The runner's is_running was never set by construction — the engine
   now calls the new Start() (the __init__ law).
6. The model-changed sync: the Go runner holds a plain-map copy of the
   config — the engine's onModelChanged mirrors the in-place write onto
   the shared OrderedMap before the save (the reference's
   ONE-shared-dict law).
7. The §4.4 classify CREATES the (empty) real cache.db even in a
   dry-run (the reference's classify_changes constructs the full
   NoteStateDB — the file's creation is not a recorded change; the
   test's law is "no rows leak", matching the reference).
8. The primary-language tag insert's case-sensitive membership quirk
   ("Go" not in [go, …] → 'go' inserted TWICE) is preserved on both
   sides; the golden's fake repos carry the full PyGithub surface
   (language/has_issues/license/has_wiki) so the reference's doc-score
   try-block runs to completion and both sides compute the same
   credibility.

Option-B divergences (asserted or documented, never silent): the
cmd_auto missing-settings check keeps the reference's five (the
wizard's compat law); the fetch-failure tip that pointed at --login is
dropped (the login door is an accepted loss); bot_queue_max_id is not
passed (mirror-push is the ack); the fetch's Telethon chatter fell
away (the Worker poll has no log lines — the spinner wraps a silent
call).

GO NOTES (the new house laws the engine golden taught):
- THE STATE-CHAIN LAW for isolated replays: the four SESSION-030
  sections are ONE state chain in the payload's own order (the
  settlement primes right before the drivers; each cmd_auto's
  classification writes feed the engine). A replay in a FRESH env
  (the mutation test) must reproduce the payload's own pre-state —
  the --reset-dead run (it clears the dead set; without it the
  engine's banish pass records a missing-note backfill the golden
  never saw) and the dry-run report writer's real <app>/reports dirs
  (without them the engine's makedirs rows say "create folder"
  instead of "folder exists") — or the engine's report body
  diverges. replayDriverScenarios carries the chain; the mutation
  test's preamble carries the pre-state.
- THE MUTATION-TEETH LAW: the drivers' mutation test must run the
  REAL replay against the mutated golden (a mutated guard card, a
  mutated report body, a mutated log sequence each FAIL the replay;
  the unmutated golden must still PASS — the harness guard) — a
  mutation test that only proves the pattern strings exist has no
  teeth (the SESSION-030 close caught exactly that and fixed it).
- THE DISPATCH-COMPLETION LAW (SESSION-031): a command is only LANDED
  when Run's dispatch routes the argv to it — unit tests that call the
  cmd* methods directly keep the suite green while the user-facing
  door stays stale (T-62's drivers landed in SESSION-030; T-63's smoke
  through Run caught `--auto` still answering the retired not-yet
  card). The acceptance for a CLI command drives `Run(argv)` — the
  smoke test is the completion proof, not decoration.

Remaining on T-62's books: nothing — the task is COMPLETE (the engine,
the drivers, the SIGINT/heartbeat/exit law, the chrome-retry live leg
all landed with the differential proof). T-61's remaining paint legs
and T-63 (launchers) follow.

## SESSION-031 shipped record — T-61's CLOSE-OUT AUDIT (the row closed with evidence)

The row demanded: ANSI paint, UTF-8 console hardening (65001), NO_COLOR,
spinner/bar — acceptance "Windows console test". The audit's findings:

- **The paint box is LANDED and TESTED** (SESSION-027, terminal.go): the
  C ANSI system with the NO_COLOR / non-TTY / --no-color law
  (TestEnableColorsDoor), the level styles (the dead-level-color quirk
  frozen per the reference), term_width's COLUMNS law
  (TestTermWidthColumnsLaw), mask/_mask_token, rule (TestRuleCap64),
  cli_print (the golden's cliprint_lines), banner (the golden's banner
  section — the braille/box glyph bytes), render_bar with banker's
  rounding (TestRenderBarBankersRounding + TestStatusLineBarWidth),
  StatusLine's erase/repaint grammar (TestStatusLineGrammar), strip-ANSI
  (TestStripANSITable), and the golden's paint_shapes section
  byte-compared with colors ON and OFF.
- **The 65001 hardening is LANDED** (platform_windows.go):
  SetConsoleOutputCP(65001) + SetConsoleCP(65001) via
  syscall.NewLazyDLL("kernel32.dll") — zero new dependencies — called by
  hardenStdio() at Run's start (the reference's _harden_stdio twin; the
  errors="replace" half is Go-native). COMPILE-LEVEL EVIDENCE
  (2026-10-19, this session): `GOOS=windows GOARCH=amd64 CGO_ENABLED=0
  go build ./...` and `GOOS=windows GOARCH=amd64 go vet ./...` both
  clean over the whole module.
- **The spinner under a real loop is LANDED** (SESSION-030): the
  engine's 100 ms heartbeat ticker driving StatusLine.Tick, with the
  StatusLine paint mutex (Qt's signal-marshaling twin) —
  engine_test.go's heartbeat proof ("the heartbeat never ticked"
  guard).
- **The exact honest remainder** (recorded, not this row's): the
  RUNTIME Windows-console legs — an actual console switching to 65001
  and rendering the braille spinner frames, the box-drawing banner and
  the VT sequences live — cannot run in the Linux sandbox; they are
  Phase 7's clean-machine checklist items (doc 10 §7 items 2 and 7;
  T-70's WebView2-less Windows VM), per the SESSION-031 kickoff's own
  scoping ("the clean-VM legs belong to Phase 7, not here").
- **ADR-0007 Am.2 records the re-instatement** (this session): Am.1
  had cut the owner-facing CLI polish; the owner's SESSION-027..031
  instruction line re-instated it (the paint is load-bearing golden
  surface; the launchers + the v0.92.0-go.1 tag are the owner's
  commanded books T-63/T-65). The ADR now says what the practice does.

## SESSION-031 shipped record — T-63: THE LAUNCHERS (cmd + the .bat set + the sh + the 1-INSTALL equivalent + the VERSION stamp + THE DISPATCH FLIP)

Landed:

- `cmd/gitcurator-cli/main.go` — the main package as the twin of
  main.py's CLI routing: the **--cli strip law ported verbatim**
  (main.py: `[a for a in sys.argv[1:] if a != "--cli"]` — the
  launchers keep the reference's own `--cli --auto --yes` run lines
  byte-close, and muscle memory keeps working; the strip is proven by
  main_test.go), the **--version stamp door** (doc 10 §5: the Go
  packaging addition — `gitcurator-cli <version>` to stdout, exit 0,
  wins before the CLI's parser; a documented divergence: the
  reference's argparse has no --version and answers the unknown-flag
  error), the `version` var stamped via `-ldflags "-X
  main.version=$(cat VERSION)"` (fallback "dev"), and the exit code
  passing straight through from `cli.Main`.
- `VERSION` (repo root): `0.92.0-go.1` — the milestone T-65's tag will
  carry; the build command shape is in doc 10 §2 + main.go's stamp
  comment (`GOFLAGS=-trimpath CGO_ENABLED=0 go build -o
  bin/GitCurator-cli.exe ./cmd/gitcurator-cli`).
- **THE LAUNCHER SET** in `cmd/gitcurator-cli/` (the owner's folder
  law — everything CLI with the CLI exe; T-65's mkzip flattens them to
  the zip root): `GitCurator-CLI.bat` (`--cli --auto --yes`),
  `GitCurator-SYNC.bat` (`--cli --sync %*`),
  `GitCurator-CLI-Setup.bat` (`--cli --init`),
  `Start-GitCurator-CLI.bat` (`--cli --auto --yes %*` — plus the
  owner's EXITCODE-card normalization the reference's v0.09 alias
  lacked), `gitcurator-cli.sh` (set -euo pipefail, the walk-up, the
  exe guard card, the same default + the flag pass-through),
  `1-INSTALL.bat` (the Go equivalent: `where git` + `git --version` +
  the next-steps card; NO registry writes, NO PATH mutations, nothing
  to install — the app is a native exe).
- **THE DISPATCH FLIP** (the session's reality-wins find): the T-62
  drivers had landed but Run's dispatch still routed `--auto`,
  `--sync`, `--retry-failed`, `--import-file` to not-yet cards — the
  shipped binary would have refused its flagship mode (the unit
  battery called the cmd* methods directly, so the suite stayed
  green). Flipped to the real drivers; the not-yet cards and the
  beat-1a `cmdSync` stub retired (the golden's sync_no_config /
  sync_bad_config sections byte-unchanged — cmdSyncLive opens with
  the same beat-1a check).

THE DIFFERENTIAL STONE GREW: dump_cli.py's payload now reads the
reference's OWN launcher bytes (the six files: the four cli-folder
.bats, gitcurator-cli.sh, the app-root 1-INSTALL.bat) and records
them raw — the full text with CRLF preserved inside the JSON strings
+ the byte-level facts (bytes/ascii_only/crlf/bare_lf) — goldencli
.golden.json at **101,777 bytes** (the growth line 48,698 → 69,028 →
86,570 → 101,777), two-run byte-identical. The Go replay
(TestGoldenCLILaunchers) reads the Go repo's own launcher files and
enforces the per-file CONTRACT: the SHARED lines byte-compared (the
skeleton: @echo off / setlocal / the title / `cd /d "%~dp0"` / :run /
`set EXITCODE=%ERRORLEVEL%` / the exit-card lines / pause / endlocal;
the .sh's usage block + set -euo pipefail + cd + the if/fi + exit 1), the
DIVERGENT pairs asserted BOTH ways (each present in its own side and
absent from the other — the walk-up's landmark main.py → VERSION, the
python invocation → the direct exe run, 1-INSTALL's Using/version
lines, the sh's exec lines), the ref-only python-world lines banned
from the Go side (the venv/python/py ladder, the pip installs, the
not-found cards), and the go-only exe-world lines (the git check, the
exe guard, Start's EXITCODE card). The reference's own laws are
checked on its recorded facts (every .bat ASCII + CRLF; the .sh
LF-only UTF-8); the Go side's laws are recomputed from the raw bytes.
The mutation teeth (TestGoldenCLIMutationLaunchers): a mutated
reference run line, a mutated shared card line, and a flipped
byte-level fact each FAIL the real comparison while the unmutated
golden still passes it (the harness guard) — three surfaces, verified
biting.

Tests: cmd/gitcurator-cli/launchers_test.go (the self-side byte laws:
ASCII+CRLF for every .bat with the final CRLF, LF-only + executable
for the .sh, the grammar skeleton with the walk-up landmark VERSION,
every python-discovery command line banned, chcp banned, the EXITCODE
echo + pause + endlocal tail, the 1-INSTALL git-only checks with the
no-mutation bans) + main_test.go (the version door + the stamp feeds
it + the strip law) + internal/cli TestBatchDispatchSmoke (the four
doors route to the REAL drivers — their own honest cards answer, never
a not-yet card). **1922 tests race-clean over 33 packages** (1911 at
SESSION-030's close; −1 retired +10 new +2 golden tests; the
arithmetic closes exactly). GOOS=windows build + vet clean; gofmt
clean; zero new dependencies.

Option-B divergences (asserted, never silent): the python-discovery
ladder is GONE by design (the exe runs direct — every discovery
command line is banned from the Go bytes and the reference's lines are
pinned as ref-only); the walk-up's landmark is VERSION (the Go app
root's marker — the exe embeds everything else); Setup's re-login
pointer is replaced by the re-run-the-wizard line (the --login door
answers the ADR-0004 Am.3 refusal card); Start's bare tail is
normalized to the EXITCODE card (the owner's NEXT-SESSION v31 line:
the whole set carries the echo + pause tail); main.py's --cli selector
is stripped by the exe's main (the two-exe law, ADR-0010); the
reference's GUI-side launchers (GitCurator.bat, GitCurator-DRY-RUN.bat
at the app root) are NOT in T-63's set — they ship with T-65's zip
(the DRY-RUN one runs the same `--cli --auto --yes --dry-run` the Go
CLI already supports; recorded for the packaging task).

Remaining on T-63's books: nothing — the task is COMPLETE (the
clean-VM smoke of the acceptance column is Phase 7's checklist, per
the same scoping as T-61's runtime legs). T-64 (the differential
corpus full run) and T-65 (the tag + zip) follow.
