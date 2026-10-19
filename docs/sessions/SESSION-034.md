# SESSION-034 — 2026-10-19 — Phase 6's composition layer: T-71 COMPLETE (`internal/app` is real — the typed bus + the ask-gates, the command surface, the lifecycle/locks, the DI root) — 1999 tests race-clean over 36 packages (+12 webui); the twenty-ninth full wipe repaired BEFORE work; one mid-session CONTEXT death survived without losing work (the continuation recipe proven again)

## Goal (from NEXT-SESSION.md at start)

SESSION-034 = **T-71 (L): `internal/app`** per doc 04 §2/§5 (read first — the
authority): the composition layer — the App struct, the typed event bus with
the ask-gates (300 s code / 300 s scan / 600 s LLM — the reference's identical
timeouts), the command surface (doc 04 §5's table), one-batch-at-a-time, the
single-instance lock with PID liveness healing, the startup passes with the
reference's delays, the shutdown grace windows. Acceptance: the command
contract tests + the lock/watchdog tests. Optional: T-72's first slice only
after a PROPER T-71; prefer closing cleanly at T-71's natural boundary.

## What was done (task IDs, files, commits — cite hashes)

- **The twenty-ninth full wipe repaired BEFORE work** (no commit — nothing
  tracked changed; the environment is rebuildable): both repos, the toolchain,
  the venv, and the module cache all gone at boot. Everything rebuilt from the
  committed recipes: go1.27.2 sha-verified, both repos re-cloned at the pinned
  revs with the ephemeral `GIT_CONFIG_COUNT` header form (no PAT persisted),
  the venv (Python 3.12.15), the wails CLI v2.16.0 (rebuilt 12:04, GOBIN into
  the toolchain tree). **The rootless webkit tree DEFERRED** — T-71 is pure Go;
  the committed recipe (webui/README.md + t-70-gate.md) rebuilds it when
  T-72's build legs need it. Verified at the close of this report's writing:
  the reference at `98f4fc71cfb8` byte-equal to PINNED-GITCURATOR-REV, go.mod
  unchanged (the authorized four).

- **T-71 slice (a) — commit `5453214`**: `events.go` — the typed fan-out Bus
  (LogLine/Progress/Status/BatchFinished/JobProgress/DiskFull/ModelChanged —
  the 8 queued Qt signals' transport; subscription order, per-subscriber panic
  isolation, the never-raise law) + the three ask-gates as channels with the
  reference's IDENTICAL timeouts (300 s code, 300 s scan, 600 s LLM), the
  verdict normalization + selection-shape laws, the byte-identical timeout
  lines, the headless fail-fast (headless.py:219-227) + THE NO-READER
  fail-fast (the v30 hang-bomb law's Go twin — a gate nobody can answer
  defers immediately). 20 tests.

- **T-71 slice (b) — commit `00f4cdb`**: `commands.go` + `lock.go` — the
  command surface over injected seams (Sync.FetchQueue/RunBatch/FastSync/
  DriftSeal, Scan.Run/ApplyPlan, Backup.SealVault/Resync/Now/Publish,
  Radar.Detect; nil seams refuse honestly, never improvise; every command
  BaseException-safe — recover + the honest crash line; BatchFinished EXACTLY
  once) + the OpLock (the TelegramLockManager port: owner-scoped, no nested
  re-acquire, describe/age/isStuck, the 35-minute threshold) + the one-batch
  law (second batch refused honestly, engine untouched, lock home after
  finish AND crash). 22 tests.

- **T-71 slice (c) — commit `eb74171`**: `lifecycle.go` +
  `process_posix.go`/`process_windows.go` — AcquireAppLock (gui/app.py:393-440
  verbatim shape: alive-PID refusal, stale-lock healing, the unknown-PID belt,
  the zombie-handle ladder, the last-resort overwrite, CONTINUE WITHOUT the
  lock — never exit on a lock failure), the watchdog (60 s cadence →
  force-release + the loud line), the startup passes at the reference's exact
  delays (sentinel 1.2 s + 60 s recurrence / llama 1.5 s / heal 2.0 s /
  backlog 2.5 s / radar 3.5 s; crash-isolated; ctx-cancelled), Shutdown's
  graces (batch 5 s, each worker 2 s, then the force-release), the PID
  liveness probe (signal-0 POSIX / image-name-verified tasklist Windows).
  15 tests.

- **T-71 slice (d) — commit `74ae634`**: `app.go` + `queue.go` + `scanjob.go`
  + `backupjob.go` — the DI root: `New(configPath, headless)` wires every
  production seam (the batch composition — tracker + vault index + note-state
  DB + cache, the runner bus-ward with LLMDecision → the 600 s gate, the
  Finish chain; the five startup-pass engines — sentinel/llama-autodetect/
  heal/backlog/quiet-radar with the engines' own gates), the
  `_bot_queue_job` classification half (the settlement AT THE QUEUE'S DOOR,
  the websites ladder, QueueReport's SettledList + WebsitesVaultMissing), the
  `_vault_scan_job` engine (the per-vault story with the reference's
  byte-shaped lines, BOTH confirm doors — the bus gate + the Worker scan
  trio's 3 s poll, the picking arithmetic, the merged verdict), the
  BackupWorker port (the zip's exclude laws + FIFO rotation + the arc-root
  law; the deferred-ledger rider's absence recorded — syncreview the
  placeholder seat). 10 production-seam tests. Same commit:
  **THE HERMETIC BANNER STUB** in `internal/cli/engine_test.go` — every
  engine test had run the REAL DownloadBanner (a live 500 KB network fetch
  whose goroutine could outlive the test via the 15 s join's escape hatch and
  record a late atomic-write into a LATER test's dry-run window) — SESSION-034's
  root cause of the golden's intermittent 13th entry; the runner's own doc
  always said "the tests inject the fake banner". And **doc 04 §2's
  LANDED-SHAPE NOTE** (below).

- **The mid-session CONTEXT death survived** (the continuation recipe proven
  again, at the chat level this time): slice (d)'s edits finished at 13:15,
  the session's context exhausted before verification; the resumed context
  diagnosed the tree from evidence (git status + the log + file mtimes:
  slices (a)–(c) committed and pushed, slice (d) complete but unverified),
  then RE-VERIFIED EVERYTHING FROM SCRATCH before committing — vet clean,
  gofmt clean, 67/67 internal/app race-clean, the full battery, the secret
  scan, the mode-flip check. A preserved draft is continued and audited,
  never blindly trusted.

## Gate evaluation (pass/fail/partial + EVIDENCE)

**PASS.** T-71's acceptance = the command contract tests + the lock/watchdog
tests.

- Command contract tests: 22 over fake deps (slice b — the one-batch law,
  the event wiring, the gate round-trips, the crash-safety law across the
  surface) + the production-seam tests (slice d — the wiring smoke, the batch
  smoke over the engine fakes, the scan gate round-trips: declined keeps
  everything / confirmed banishes to .trash, the backup zip inspection, the
  queue settlement at the door).
- Lock/watchdog tests: the OpLock laws (in b's 22) + lifecycle's 15 (the
  healing paths incl. alive-refusal via a reaped child, the watchdog's
  force-release + fresh-lock respect, the five passes' delay ORDER +
  recurrence + crash isolation + ctx cancellation, the grace windows).
- **67/67 `internal/app` tests race-clean** (`go test -count=1 -race
  ./internal/app/` — re-run in the resumed context post-recovery).
- **The full battery: 1999 PASS lines over 36 packages race-clean** (was
  1932/35 at SESSION-033; +67 = exactly internal/app; counted `-v` per
  package, arithmetic closed: 1932 + 67 = 1999) + **12 webui tests** (the
  nested module: deskbus + fixture) + 5 guibridge (inside the 1999).
- gofmt clean (`gofmt -l` empty over internal/app + internal/cli), go vet
  clean, the secret scan clean over every file committed, no mode-flips
  (`git diff --raw` all 100644→100644).
- T-72's first slice NOT started — **closed at T-71's natural boundary** per
  the mandate's own law (the same call SESSION-033 made for T-71; the
  theme-token CSS golden deserves a fresh session's full context).

## Spec updates made (file → what changed, why)

- `docs/port-blueprint/04-target-go-architecture.md` §2 — the LANDED-SHAPE
  NOTE: the §2 sketch was indicative; the landed `internal/app` differs where
  the port's own laws forced it, recorded in the landing commit: no
  `config.Store` type (the config package's `Load` returns the live
  `*storage.OrderedMap` — `App.ConfigPath` + a load door instead); no
  persistent `*store.DB` handle (the store law — every composition opens its
  stores per command and closes them); no Telegram field (ADR-0004 Am.3/
  Option B — the queue source is the Worker `/api/pending` poll wired as the
  `QueueFetch` seam); the LLM routing composed in the scan/batch adapters
  (`llmclient.ResolveTaskModel`); the batch's CLI-parity shell wiring (warmup
  family, inputs fetchers, websites phase, Phase-5 CLEAR) stays the CLI's own
  until T-72's binding lands it on the app surface.
- `docs/port-blueprint/12-migration-roadmap.md` — the SESSION-034 entry
  (T-71 COMPLETE, the four slices, the acceptance GREEN).
- `docs/port-blueprint/13-implementation-task-list.md` — the T-71 row
  COMPLETE (the full slice-by-slice record).
- `README.md` — the SESSION-034 status entry.

## Blockers & decisions needed

- None new. Carried: the T-72+ production frontend tooling question (vanilla
  vs vite) remains OPEN for the owner — **T-72's session should ask first**
  (the theme-token golden is toolchain-independent, but the log desk's
  production build is not); the go-directive acknowledgement (twenty-fifth
  session carrying it); x/net/html keep-or-swap.

## Owner actions requested

- The vanilla-vs-vite call before or early in T-72 (both work under wails
  v2; the PoC rides vanilla; an ADR records the choice either way).
- Nothing else urgent — no new blockers, no new dependencies (the authorized
  four unchanged; the wails require stays in the nested webui module).

## Handoff

- **SESSION-035 = T-72 (L): the frontend foundation** per doc 13's row: the
  theme-token CSS golden (spec-gui-theme-icons — the token table → the CSS
  var dump, golden-tested), icons, fonts, the layout shell, and the
  production log desk with coalescing (spec-gui-ui). Acceptance: the token
  golden + the log perf test. **The rootless webkit tree must be rebuilt
  first** (deferred this session — the committed recipe: webui/README.md +
  docs/poc-artifacts/t-70-gate.md). `guibridge/` RETIRES at T-72 (not
  before); `webui/` dissolves into the main module and binds `internal/app`
  directly per doc 04 §2's landed-shape note — the batch's CLI-parity shell
  wiring (warmup family, inputs fetchers, websites phase, Phase-5 CLEAR)
  lands on the app surface at that binding.
- The state card v35 (`docs/NEXT-SESSION.md`) carries the full read-first
  list + the SESSION-035 prompt.
- Push state: slices (a)–(d) + this close committed; push executed with the
  ephemeral PAT at session end (never stored, never written).
